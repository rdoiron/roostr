package db

import (
	"context"
	"database/sql"
	"fmt"
)

// Migration represents a database migration.
type Migration struct {
	Version int
	Name    string
	Up      string // SQL to apply migration
}

// Migrations is the list of all migrations.
// Add new migrations to this list as the schema evolves.
var Migrations = []Migration{
	// Version 1 is the initial schema, applied via schema.sql
	{
		Version: 2,
		Name:    "add_pending_invoices",
		Up: `
-- Pending invoices for paid relay access
CREATE TABLE IF NOT EXISTS pending_invoices (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    payment_hash TEXT NOT NULL UNIQUE,    -- hex-encoded payment hash
    pubkey TEXT NOT NULL,                 -- hex format of the user's pubkey
    npub TEXT NOT NULL,                   -- bech32 format for display
    tier_id TEXT NOT NULL,                -- references pricing_tiers.id
    amount_sats INTEGER NOT NULL,         -- amount in satoshis
    payment_request TEXT NOT NULL,        -- BOLT11 invoice string
    memo TEXT,                            -- invoice memo/description
    status TEXT NOT NULL DEFAULT 'pending',  -- pending, paid, expired, cancelled
    created_at INTEGER NOT NULL DEFAULT (strftime('%s', 'now')),
    expires_at INTEGER NOT NULL,          -- when the invoice expires
    paid_at INTEGER,                      -- when payment was confirmed
    FOREIGN KEY (tier_id) REFERENCES pricing_tiers(id)
);

CREATE INDEX IF NOT EXISTS idx_pending_invoices_pubkey ON pending_invoices(pubkey);
CREATE INDEX IF NOT EXISTS idx_pending_invoices_status ON pending_invoices(status);
CREATE INDEX IF NOT EXISTS idx_pending_invoices_payment_hash ON pending_invoices(payment_hash);
CREATE INDEX IF NOT EXISTS idx_pending_invoices_expires ON pending_invoices(expires_at);
`,
	},
}

// Migrate runs all pending migrations.
func (d *DB) Migrate(ctx context.Context) error {
	currentVersion, err := d.GetSchemaVersion()
	if err != nil {
		return fmt.Errorf("failed to get schema version: %w", err)
	}

	for _, m := range Migrations {
		if m.Version <= currentVersion {
			continue
		}

		fmt.Printf("Applying migration %d: %s\n", m.Version, m.Name)

		if err := d.applyMigration(ctx, m); err != nil {
			return fmt.Errorf("failed to apply migration %d (%s): %w", m.Version, m.Name, err)
		}

		fmt.Printf("Migration %d applied successfully\n", m.Version)
	}

	return nil
}

func (d *DB) applyMigration(ctx context.Context, m Migration) error {
	return d.Transaction(ctx, func(tx *sql.Tx) error {
		// Execute migration SQL
		if _, err := tx.ExecContext(ctx, m.Up); err != nil {
			return err
		}

		// Record the migration
		if _, err := tx.ExecContext(ctx, "INSERT INTO schema_version (version) VALUES (?)", m.Version); err != nil {
			return err
		}

		return nil
	})
}

// Transaction wrapper that accepts *DB instead of *sql.Tx for simpler usage
func (d *DB) transactionForMigration(ctx context.Context, fn func() error) error {
	tx, err := d.AppDB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	if err := fn(); err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

// GetPendingMigrations returns migrations that haven't been applied yet.
func (d *DB) GetPendingMigrations() ([]Migration, error) {
	currentVersion, err := d.GetSchemaVersion()
	if err != nil {
		return nil, err
	}

	var pending []Migration
	for _, m := range Migrations {
		if m.Version > currentVersion {
			pending = append(pending, m)
		}
	}

	return pending, nil
}

// GetAppliedMigrations returns the list of applied migration versions.
func (d *DB) GetAppliedMigrations(ctx context.Context) ([]int, error) {
	rows, err := d.AppDB.QueryContext(ctx, "SELECT version FROM schema_version ORDER BY version")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var versions []int
	for rows.Next() {
		var v int
		if err := rows.Scan(&v); err != nil {
			return nil, err
		}
		versions = append(versions, v)
	}

	return versions, rows.Err()
}
