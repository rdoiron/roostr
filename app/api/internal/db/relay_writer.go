package db

import (
	"context"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// RelayWriter provides write operations on the relay database.
// These operations require a temporary read-write connection.
type RelayWriter struct {
	db *sql.DB
}

// NewRelayWriter opens a temporary read-write connection to the relay database.
// The caller must call Close() when done.
func (d *DB) NewRelayWriter() (*RelayWriter, error) {
	db, err := d.OpenRelayDBForWrite()
	if err != nil {
		return nil, err
	}
	return &RelayWriter{db: db}, nil
}

// Close closes the relay writer connection.
func (w *RelayWriter) Close() error {
	if w.db != nil {
		return w.db.Close()
	}
	return nil
}

// DeleteEventsBefore deletes events created before the given timestamp.
// It respects the given exceptions (e.g., ["kind:0", "kind:3", "pubkey:abc123"]).
// Returns the number of deleted events.
func (w *RelayWriter) DeleteEventsBefore(ctx context.Context, before time.Time, exceptions []string, operatorPubkey string) (int64, error) {
	// Build the query with exceptions
	query := "DELETE FROM event WHERE created_at < ?"
	args := []interface{}{before.Unix()}

	// Parse exceptions
	var kindExceptions []int
	var pubkeyExceptions [][]byte

	for _, exc := range exceptions {
		if strings.HasPrefix(exc, "kind:") {
			var kind int
			fmt.Sscanf(exc, "kind:%d", &kind)
			kindExceptions = append(kindExceptions, kind)
		} else if strings.HasPrefix(exc, "pubkey:") {
			pubkey := strings.TrimPrefix(exc, "pubkey:")
			if pubkey == "operator" && operatorPubkey != "" {
				pubkeyBytes, err := hex.DecodeString(operatorPubkey)
				if err == nil {
					pubkeyExceptions = append(pubkeyExceptions, pubkeyBytes)
				}
			} else {
				pubkeyBytes, err := hex.DecodeString(pubkey)
				if err == nil {
					pubkeyExceptions = append(pubkeyExceptions, pubkeyBytes)
				}
			}
		}
	}

	// Add kind exceptions
	if len(kindExceptions) > 0 {
		placeholders := make([]string, len(kindExceptions))
		for i, kind := range kindExceptions {
			placeholders[i] = "?"
			args = append(args, kind)
		}
		query += fmt.Sprintf(" AND kind NOT IN (%s)", strings.Join(placeholders, ","))
	}

	// Add pubkey exceptions
	if len(pubkeyExceptions) > 0 {
		placeholders := make([]string, len(pubkeyExceptions))
		for i, pubkey := range pubkeyExceptions {
			placeholders[i] = "?"
			args = append(args, pubkey)
		}
		query += fmt.Sprintf(" AND pubkey NOT IN (%s)", strings.Join(placeholders, ","))
	}

	result, err := w.db.ExecContext(ctx, query, args...)
	if err != nil {
		return 0, fmt.Errorf("failed to delete events: %w", err)
	}

	count, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("failed to get affected rows: %w", err)
	}

	return count, nil
}

// DeleteEventsByIDs deletes specific events by their IDs.
// Returns the number of deleted events.
func (w *RelayWriter) DeleteEventsByIDs(ctx context.Context, ids []string) (int64, error) {
	if len(ids) == 0 {
		return 0, nil
	}

	// Convert hex IDs to bytes
	placeholders := make([]string, len(ids))
	args := make([]interface{}, len(ids))
	for i, id := range ids {
		idBytes, err := hex.DecodeString(id)
		if err != nil {
			return 0, fmt.Errorf("invalid event ID %s: %w", id, err)
		}
		placeholders[i] = "?"
		args[i] = idBytes
	}

	query := fmt.Sprintf("DELETE FROM event WHERE id IN (%s)", strings.Join(placeholders, ","))
	result, err := w.db.ExecContext(ctx, query, args...)
	if err != nil {
		return 0, fmt.Errorf("failed to delete events: %w", err)
	}

	count, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("failed to get affected rows: %w", err)
	}

	return count, nil
}

// GetEventAuthor returns the author pubkey of an event by ID.
// Returns empty string if not found.
func (w *RelayWriter) GetEventAuthor(ctx context.Context, eventID string) (string, error) {
	idBytes, err := hex.DecodeString(eventID)
	if err != nil {
		return "", fmt.Errorf("invalid event ID: %w", err)
	}

	var pubkeyBytes []byte
	err = w.db.QueryRowContext(ctx, "SELECT pubkey FROM event WHERE id = ?", idBytes).Scan(&pubkeyBytes)
	if err == sql.ErrNoRows {
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("failed to get event author: %w", err)
	}

	return hex.EncodeToString(pubkeyBytes), nil
}

// RunVacuum runs VACUUM on the relay database to reclaim space.
func (w *RelayWriter) RunVacuum(ctx context.Context) error {
	_, err := w.db.ExecContext(ctx, "VACUUM")
	if err != nil {
		return fmt.Errorf("failed to vacuum relay database: %w", err)
	}
	return nil
}

// RunIntegrityCheck runs an integrity check on the relay database.
// Returns true if the database is healthy, along with the result message.
func (w *RelayWriter) RunIntegrityCheck(ctx context.Context) (bool, string, error) {
	var result string
	err := w.db.QueryRowContext(ctx, "PRAGMA integrity_check").Scan(&result)
	if err != nil {
		return false, "", fmt.Errorf("failed to run integrity check: %w", err)
	}
	return result == "ok", result, nil
}

// GetDatabaseSizeBeforeAfterVacuum calculates the size difference after VACUUM.
// This is a helper that doesn't actually run VACUUM.
func (w *RelayWriter) GetPageInfo(ctx context.Context) (pageCount int64, freePages int64, pageSize int64, err error) {
	err = w.db.QueryRowContext(ctx, "PRAGMA page_count").Scan(&pageCount)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("failed to get page count: %w", err)
	}

	err = w.db.QueryRowContext(ctx, "PRAGMA freelist_count").Scan(&freePages)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("failed to get free page count: %w", err)
	}

	err = w.db.QueryRowContext(ctx, "PRAGMA page_size").Scan(&pageSize)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("failed to get page size: %w", err)
	}

	return pageCount, freePages, pageSize, nil
}

// InsertEvent inserts a Nostr event into the relay database.
// Uses INSERT OR IGNORE to handle duplicates gracefully.
// Returns true if the event was inserted (new), false if it already existed.
func (w *RelayWriter) InsertEvent(ctx context.Context, event *Event) (bool, error) {
	// Convert hex ID to bytes
	idBytes, err := hex.DecodeString(event.ID)
	if err != nil {
		return false, fmt.Errorf("invalid event ID: %w", err)
	}

	// Convert hex pubkey to bytes
	pubkeyBytes, err := hex.DecodeString(event.Pubkey)
	if err != nil {
		return false, fmt.Errorf("invalid pubkey: %w", err)
	}

	// Convert hex signature to bytes
	sigBytes, err := hex.DecodeString(event.Sig)
	if err != nil {
		return false, fmt.Errorf("invalid signature: %w", err)
	}

	// Serialize tags to JSON
	tagsJSON, err := json.Marshal(event.Tags)
	if err != nil {
		return false, fmt.Errorf("failed to serialize tags: %w", err)
	}

	// Insert with OR IGNORE for duplicate handling
	result, err := w.db.ExecContext(ctx, `
		INSERT OR IGNORE INTO event (id, pubkey, created_at, kind, tags, content, sig)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, idBytes, pubkeyBytes, event.CreatedAt.Unix(), event.Kind, string(tagsJSON), event.Content, sigBytes)
	if err != nil {
		return false, fmt.Errorf("failed to insert event: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("failed to get affected rows: %w", err)
	}

	return rows > 0, nil
}
