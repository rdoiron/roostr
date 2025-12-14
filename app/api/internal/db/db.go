// Package db provides database access for Roostr.
// It manages two SQLite databases:
// - Relay DB (read-only): The nostr-rs-relay database
// - App DB (read-write): Roostr's own database for settings and metadata
package db

import (
	"context"
	"database/sql"
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"syscall"

	_ "github.com/mattn/go-sqlite3"
)

//go:embed schema.sql
var schema string

// DB holds database connections.
type DB struct {
	RelayDB *sql.DB // Read-only access to relay database
	AppDB   *sql.DB // Read-write access to app database

	relayPath string
	appPath   string
	mu        sync.RWMutex
}

// New creates a new DB instance and initializes connections.
func New(relayDBPath, appDBPath string) (*DB, error) {
	db := &DB{
		relayPath: relayDBPath,
		appPath:   appDBPath,
	}

	// Initialize app database (required)
	if err := db.initAppDB(); err != nil {
		return nil, fmt.Errorf("failed to initialize app database: %w", err)
	}

	// Try to connect to relay database (may not exist yet)
	if relayDBPath != "" {
		if err := db.connectRelayDB(); err != nil {
			// Log warning but don't fail - relay DB may not exist yet
			fmt.Printf("Warning: Could not connect to relay database: %v\n", err)
		}
	}

	return db, nil
}

// initAppDB initializes the application database with schema.
func (d *DB) initAppDB() error {
	// Ensure directory exists
	dir := filepath.Dir(d.appPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create database directory: %w", err)
	}

	// Open database with WAL mode for better concurrency
	dsn := fmt.Sprintf("%s?_journal_mode=WAL&_busy_timeout=5000&_foreign_keys=ON", d.appPath)
	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return fmt.Errorf("failed to open app database: %w", err)
	}

	// Set connection pool settings
	db.SetMaxOpenConns(1) // SQLite works best with single writer
	db.SetMaxIdleConns(1)

	// Test connection
	if err := db.Ping(); err != nil {
		db.Close()
		return fmt.Errorf("failed to ping app database: %w", err)
	}

	// Apply schema
	if _, err := db.Exec(schema); err != nil {
		db.Close()
		return fmt.Errorf("failed to apply schema: %w", err)
	}

	d.AppDB = db
	return nil
}

// connectRelayDB connects to the relay database in read-only mode.
func (d *DB) connectRelayDB() error {
	// Check if file exists
	if _, err := os.Stat(d.relayPath); os.IsNotExist(err) {
		return fmt.Errorf("relay database does not exist: %s", d.relayPath)
	}

	// Open in read-only mode
	dsn := fmt.Sprintf("file:%s?mode=ro&_busy_timeout=5000", d.relayPath)
	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return fmt.Errorf("failed to open relay database: %w", err)
	}

	// Set connection pool settings
	db.SetMaxOpenConns(3) // Allow multiple readers
	db.SetMaxIdleConns(1)

	// Test connection
	if err := db.Ping(); err != nil {
		db.Close()
		return fmt.Errorf("failed to ping relay database: %w", err)
	}

	d.RelayDB = db
	return nil
}

// Close closes all database connections.
func (d *DB) Close() error {
	d.mu.Lock()
	defer d.mu.Unlock()

	var errs []error

	if d.RelayDB != nil {
		if err := d.RelayDB.Close(); err != nil {
			errs = append(errs, fmt.Errorf("failed to close relay database: %w", err))
		}
		d.RelayDB = nil
	}

	if d.AppDB != nil {
		if err := d.AppDB.Close(); err != nil {
			errs = append(errs, fmt.Errorf("failed to close app database: %w", err))
		}
		d.AppDB = nil
	}

	if len(errs) > 0 {
		return errs[0]
	}
	return nil
}

// IsRelayDBConnected returns true if the relay database is connected.
func (d *DB) IsRelayDBConnected() bool {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.RelayDB != nil
}

// ReconnectRelayDB attempts to reconnect to the relay database.
// Useful when the relay creates its database after Roostr starts.
func (d *DB) ReconnectRelayDB() error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.RelayDB != nil {
		d.RelayDB.Close()
		d.RelayDB = nil
	}

	return d.connectRelayDB()
}

// Transaction executes a function within a database transaction on the app database.
func (d *DB) Transaction(ctx context.Context, fn func(tx *sql.Tx) error) error {
	tx, err := d.AppDB.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	if err := fn(tx); err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("failed to rollback: %v (original error: %w)", rbErr, err)
		}
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// GetSchemaVersion returns the current schema version.
func (d *DB) GetSchemaVersion() (int, error) {
	var version int
	err := d.AppDB.QueryRow("SELECT MAX(version) FROM schema_version").Scan(&version)
	if err != nil {
		return 0, err
	}
	return version, nil
}

// GetRelayDatabaseSize returns the size of the relay database file in bytes.
func (d *DB) GetRelayDatabaseSize() (int64, error) {
	if d.relayPath == "" {
		return 0, nil
	}

	info, err := os.Stat(d.relayPath)
	if err != nil {
		if os.IsNotExist(err) {
			return 0, nil
		}
		return 0, fmt.Errorf("failed to stat relay database: %w", err)
	}

	return info.Size(), nil
}

// GetAppDatabaseSize returns the size of the app database file in bytes.
func (d *DB) GetAppDatabaseSize() (int64, error) {
	if d.appPath == "" {
		return 0, nil
	}

	info, err := os.Stat(d.appPath)
	if err != nil {
		if os.IsNotExist(err) {
			return 0, nil
		}
		return 0, fmt.Errorf("failed to stat app database: %w", err)
	}

	return info.Size(), nil
}

// OpenRelayDBForWrite opens a temporary read-write connection to the relay database.
// The caller is responsible for closing the connection when done.
// This should only be used for maintenance operations like cleanup and vacuum.
func (d *DB) OpenRelayDBForWrite() (*sql.DB, error) {
	d.mu.RLock()
	relayPath := d.relayPath
	d.mu.RUnlock()

	if relayPath == "" {
		return nil, fmt.Errorf("relay database path not configured")
	}

	// Check if file exists
	if _, err := os.Stat(relayPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("relay database does not exist: %s", relayPath)
	}

	// Open in read-write mode with WAL and busy timeout
	dsn := fmt.Sprintf("file:%s?_journal_mode=WAL&_busy_timeout=10000", relayPath)
	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open relay database for write: %w", err)
	}

	// Use single connection for write operations
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)

	// Test connection
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping relay database: %w", err)
	}

	return db, nil
}

// GetRelayPath returns the path to the relay database file.
func (d *DB) GetRelayPath() string {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.relayPath
}

// GetAvailableDiskSpace returns the available disk space in bytes for the relay database path.
func (d *DB) GetAvailableDiskSpace() (int64, error) {
	d.mu.RLock()
	path := d.relayPath
	d.mu.RUnlock()

	if path == "" {
		return 0, fmt.Errorf("relay database path not configured")
	}

	// Get the directory of the database file
	dir := filepath.Dir(path)

	var stat syscall.Statfs_t
	if err := syscall.Statfs(dir, &stat); err != nil {
		return 0, fmt.Errorf("failed to get disk space: %w", err)
	}

	// Available space = available blocks * block size
	available := int64(stat.Bavail) * int64(stat.Bsize)
	return available, nil
}

// GetTotalDiskSpace returns the total disk space in bytes for the relay database path.
func (d *DB) GetTotalDiskSpace() (int64, error) {
	d.mu.RLock()
	path := d.relayPath
	d.mu.RUnlock()

	if path == "" {
		return 0, fmt.Errorf("relay database path not configured")
	}

	// Get the directory of the database file
	dir := filepath.Dir(path)

	var stat syscall.Statfs_t
	if err := syscall.Statfs(dir, &stat); err != nil {
		return 0, fmt.Errorf("failed to get disk space: %w", err)
	}

	// Total space = total blocks * block size
	total := int64(stat.Blocks) * int64(stat.Bsize)
	return total, nil
}
