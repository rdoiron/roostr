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

// Event represents a Nostr event from the relay database.
type Event struct {
	ID        string    `json:"id"`
	Pubkey    string    `json:"pubkey"`
	CreatedAt time.Time `json:"created_at"`
	Kind      int       `json:"kind"`
	Tags      [][]string `json:"tags"`
	Content   string    `json:"content"`
	Sig       string    `json:"sig"`
}

// EventFilter defines filters for querying events.
type EventFilter struct {
	IDs     []string  // Event IDs
	Authors []string  // Pubkeys (hex)
	Kinds   []int     // Event kinds
	Since   time.Time // Events after this time
	Until   time.Time // Events before this time
	Limit   int       // Max results (default 50)
	Offset  int       // Pagination offset
	Search  string    // Content search (basic)
}

// RelayStats holds aggregate statistics from the relay database.
type RelayStats struct {
	TotalEvents   int64            `json:"total_events"`
	TotalPubkeys  int64            `json:"total_pubkeys"`
	EventsByKind  map[int]int64    `json:"events_by_kind"`
	DatabaseSize  int64            `json:"database_size_bytes"`
	OldestEvent   time.Time        `json:"oldest_event"`
	NewestEvent   time.Time        `json:"newest_event"`
}

// GetEvent retrieves a single event by ID.
func (d *DB) GetEvent(ctx context.Context, id string) (*Event, error) {
	if d.RelayDB == nil {
		return nil, fmt.Errorf("relay database not connected")
	}

	idBytes, err := hex.DecodeString(id)
	if err != nil {
		return nil, fmt.Errorf("invalid event ID: %w", err)
	}

	row := d.RelayDB.QueryRowContext(ctx, `
		SELECT id, pubkey, created_at, kind, tags, content, sig
		FROM event
		WHERE id = ?
	`, idBytes)

	return scanEvent(row)
}

// GetEvents retrieves events matching the filter.
func (d *DB) GetEvents(ctx context.Context, filter EventFilter) ([]Event, error) {
	if d.RelayDB == nil {
		return nil, fmt.Errorf("relay database not connected")
	}

	// Build query
	query := `SELECT id, pubkey, created_at, kind, tags, content, sig FROM event WHERE 1=1`
	args := []interface{}{}

	if len(filter.IDs) > 0 {
		placeholders := make([]string, len(filter.IDs))
		for i, id := range filter.IDs {
			idBytes, err := hex.DecodeString(id)
			if err != nil {
				return nil, fmt.Errorf("invalid event ID: %w", err)
			}
			placeholders[i] = "?"
			args = append(args, idBytes)
		}
		query += fmt.Sprintf(" AND id IN (%s)", strings.Join(placeholders, ","))
	}

	if len(filter.Authors) > 0 {
		placeholders := make([]string, len(filter.Authors))
		for i, pubkey := range filter.Authors {
			pubkeyBytes, err := hex.DecodeString(pubkey)
			if err != nil {
				return nil, fmt.Errorf("invalid pubkey: %w", err)
			}
			placeholders[i] = "?"
			args = append(args, pubkeyBytes)
		}
		query += fmt.Sprintf(" AND pubkey IN (%s)", strings.Join(placeholders, ","))
	}

	if len(filter.Kinds) > 0 {
		placeholders := make([]string, len(filter.Kinds))
		for i, kind := range filter.Kinds {
			placeholders[i] = "?"
			args = append(args, kind)
		}
		query += fmt.Sprintf(" AND kind IN (%s)", strings.Join(placeholders, ","))
	}

	if !filter.Since.IsZero() {
		query += " AND created_at >= ?"
		args = append(args, filter.Since.Unix())
	}

	if !filter.Until.IsZero() {
		query += " AND created_at <= ?"
		args = append(args, filter.Until.Unix())
	}

	if filter.Search != "" {
		query += " AND content LIKE ?"
		args = append(args, "%"+filter.Search+"%")
	}

	// Order and pagination
	query += " ORDER BY created_at DESC"

	limit := filter.Limit
	if limit <= 0 {
		limit = 50
	}
	if limit > 1000 {
		limit = 1000
	}
	query += fmt.Sprintf(" LIMIT %d", limit)

	if filter.Offset > 0 {
		query += fmt.Sprintf(" OFFSET %d", filter.Offset)
	}

	rows, err := d.RelayDB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query events: %w", err)
	}
	defer rows.Close()

	var events []Event
	for rows.Next() {
		event, err := scanEventRows(rows)
		if err != nil {
			return nil, err
		}
		events = append(events, *event)
	}

	return events, rows.Err()
}

// GetRecentEvents retrieves the most recent events.
func (d *DB) GetRecentEvents(ctx context.Context, limit int) ([]Event, error) {
	return d.GetEvents(ctx, EventFilter{Limit: limit})
}

// GetRelayStats retrieves aggregate statistics from the relay database.
func (d *DB) GetRelayStats(ctx context.Context) (*RelayStats, error) {
	if d.RelayDB == nil {
		return nil, fmt.Errorf("relay database not connected")
	}

	stats := &RelayStats{
		EventsByKind: make(map[int]int64),
	}

	// Total events
	err := d.RelayDB.QueryRowContext(ctx, "SELECT COUNT(*) FROM event").Scan(&stats.TotalEvents)
	if err != nil {
		return nil, fmt.Errorf("failed to count events: %w", err)
	}

	// Unique pubkeys
	err = d.RelayDB.QueryRowContext(ctx, "SELECT COUNT(DISTINCT pubkey) FROM event").Scan(&stats.TotalPubkeys)
	if err != nil {
		return nil, fmt.Errorf("failed to count pubkeys: %w", err)
	}

	// Events by kind
	rows, err := d.RelayDB.QueryContext(ctx, "SELECT kind, COUNT(*) FROM event GROUP BY kind ORDER BY COUNT(*) DESC")
	if err != nil {
		return nil, fmt.Errorf("failed to count events by kind: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var kind int
		var count int64
		if err := rows.Scan(&kind, &count); err != nil {
			return nil, err
		}
		stats.EventsByKind[kind] = count
	}

	// Oldest and newest event timestamps
	var oldest, newest sql.NullInt64
	err = d.RelayDB.QueryRowContext(ctx, "SELECT MIN(created_at), MAX(created_at) FROM event").Scan(&oldest, &newest)
	if err != nil {
		return nil, fmt.Errorf("failed to get event time range: %w", err)
	}
	if oldest.Valid {
		stats.OldestEvent = time.Unix(oldest.Int64, 0)
	}
	if newest.Valid {
		stats.NewestEvent = time.Unix(newest.Int64, 0)
	}

	return stats, nil
}

// CountEventsByPubkey counts events for each pubkey.
func (d *DB) CountEventsByPubkey(ctx context.Context, pubkeys []string) (map[string]int64, error) {
	if d.RelayDB == nil {
		return nil, fmt.Errorf("relay database not connected")
	}

	result := make(map[string]int64)

	for _, pubkey := range pubkeys {
		pubkeyBytes, err := hex.DecodeString(pubkey)
		if err != nil {
			continue
		}

		var count int64
		err = d.RelayDB.QueryRowContext(ctx, "SELECT COUNT(*) FROM event WHERE pubkey = ?", pubkeyBytes).Scan(&count)
		if err != nil {
			continue
		}
		result[pubkey] = count
	}

	return result, nil
}

// GetTopAuthors returns the pubkeys with the most events.
func (d *DB) GetTopAuthors(ctx context.Context, limit int) ([]struct {
	Pubkey     string `json:"pubkey"`
	EventCount int64  `json:"event_count"`
}, error) {
	if d.RelayDB == nil {
		return nil, fmt.Errorf("relay database not connected")
	}

	if limit <= 0 {
		limit = 10
	}

	rows, err := d.RelayDB.QueryContext(ctx, `
		SELECT pubkey, COUNT(*) as count
		FROM event
		GROUP BY pubkey
		ORDER BY count DESC
		LIMIT ?
	`, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query top authors: %w", err)
	}
	defer rows.Close()

	var authors []struct {
		Pubkey     string `json:"pubkey"`
		EventCount int64  `json:"event_count"`
	}

	for rows.Next() {
		var pubkeyBytes []byte
		var count int64
		if err := rows.Scan(&pubkeyBytes, &count); err != nil {
			return nil, err
		}
		authors = append(authors, struct {
			Pubkey     string `json:"pubkey"`
			EventCount int64  `json:"event_count"`
		}{
			Pubkey:     hex.EncodeToString(pubkeyBytes),
			EventCount: count,
		})
	}

	return authors, rows.Err()
}

// Helper functions

func scanEvent(row *sql.Row) (*Event, error) {
	var idBytes, pubkeyBytes, sigBytes []byte
	var createdAt int64
	var kind int
	var tagsJSON, content string

	err := row.Scan(&idBytes, &pubkeyBytes, &createdAt, &kind, &tagsJSON, &content, &sigBytes)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to scan event: %w", err)
	}

	return parseEvent(idBytes, pubkeyBytes, createdAt, kind, tagsJSON, content, sigBytes)
}

func scanEventRows(rows *sql.Rows) (*Event, error) {
	var idBytes, pubkeyBytes, sigBytes []byte
	var createdAt int64
	var kind int
	var tagsJSON, content string

	err := rows.Scan(&idBytes, &pubkeyBytes, &createdAt, &kind, &tagsJSON, &content, &sigBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to scan event: %w", err)
	}

	return parseEvent(idBytes, pubkeyBytes, createdAt, kind, tagsJSON, content, sigBytes)
}

func parseEvent(idBytes, pubkeyBytes []byte, createdAt int64, kind int, tagsJSON, content string, sigBytes []byte) (*Event, error) {
	var tags [][]string
	if tagsJSON != "" {
		if err := json.Unmarshal([]byte(tagsJSON), &tags); err != nil {
			// If JSON parsing fails, use empty tags
			tags = [][]string{}
		}
	}

	return &Event{
		ID:        hex.EncodeToString(idBytes),
		Pubkey:    hex.EncodeToString(pubkeyBytes),
		CreatedAt: time.Unix(createdAt, 0),
		Kind:      kind,
		Tags:      tags,
		Content:   content,
		Sig:       hex.EncodeToString(sigBytes),
	}, nil
}
