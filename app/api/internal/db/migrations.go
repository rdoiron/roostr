package db

import (
	"context"
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
	// Future migrations go here:
	// {
	// 	Version: 2,
	// 	Name:    "add_some_column",
	// 	Up:      `ALTER TABLE some_table ADD COLUMN new_column TEXT;`,
	// },
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
	return d.Transaction(ctx, func(tx *DB) error {
		// Execute migration SQL
		if _, err := d.AppDB.ExecContext(ctx, m.Up); err != nil {
			return err
		}

		// Record the migration
		if _, err := d.AppDB.ExecContext(ctx, "INSERT INTO schema_version (version) VALUES (?)", m.Version); err != nil {
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
