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
	ID        string     `json:"id"`
	Pubkey    string     `json:"pubkey"`
	CreatedAt time.Time  `json:"created_at"`
	Kind      int        `json:"kind"`
	Tags      [][]string `json:"tags"`
	Content   string     `json:"content"`
	Sig       string     `json:"sig"`
}

// ExportEvent represents a Nostr event for export with Unix timestamp.
// This matches the standard Nostr event format expected by other clients.
type ExportEvent struct {
	ID        string     `json:"id"`
	Pubkey    string     `json:"pubkey"`
	CreatedAt int64      `json:"created_at"` // Unix timestamp for Nostr compatibility
	Kind      int        `json:"kind"`
	Tags      [][]string `json:"tags"`
	Content   string     `json:"content"`
	Sig       string     `json:"sig"`
}

// EventFilter defines filters for querying events.
type EventFilter struct {
	IDs      []string  // Event IDs
	Authors  []string  // Pubkeys (hex)
	Kinds    []int     // Event kinds
	Since    time.Time // Events after this time
	Until    time.Time // Events before this time
	Limit    int       // Max results (default 50)
	Offset   int       // Pagination offset
	Search   string    // Content search (basic)
	Mentions string    // Filter events mentioning this pubkey (hex)
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
		SELECT event_hash, author, created_at, kind, content
		FROM event
		WHERE event_hash = ?
	`, idBytes)

	return scanEvent(row)
}

// GetEvents retrieves events matching the filter.
func (d *DB) GetEvents(ctx context.Context, filter EventFilter) ([]Event, error) {
	if d.RelayDB == nil {
		return nil, fmt.Errorf("relay database not connected")
	}

	// Build query - nostr-rs-relay uses event_hash for ID, author for pubkey,
	// and stores the full event JSON in content
	query := `SELECT event_hash, author, created_at, kind, content FROM event WHERE 1=1`
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
		query += fmt.Sprintf(" AND event_hash IN (%s)", strings.Join(placeholders, ","))
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
		query += fmt.Sprintf(" AND author IN (%s)", strings.Join(placeholders, ","))
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

	if filter.Mentions != "" {
		// Filter events that have a "p" tag mentioning this pubkey
		// In nostr-rs-relay, tags are stored within the content JSON
		query += " AND content LIKE ?"
		args = append(args, `%["p","`+filter.Mentions+`"%`)
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

	// Unique pubkeys (nostr-rs-relay uses 'author' column)
	err = d.RelayDB.QueryRowContext(ctx, "SELECT COUNT(DISTINCT author) FROM event").Scan(&stats.TotalPubkeys)
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

// GetEventsToday returns the count of events created today (since midnight in the given timezone).
func (d *DB) GetEventsToday(ctx context.Context, loc *time.Location) (int64, error) {
	if d.RelayDB == nil {
		return 0, fmt.Errorf("relay database not connected")
	}

	// Use UTC if no location provided
	if loc == nil {
		loc = time.UTC
	}

	// Get the start of today (midnight in user's timezone)
	now := time.Now().In(loc)
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc)

	var count int64
	err := d.RelayDB.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM event WHERE created_at >= ?",
		startOfDay.Unix(),
	).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count today's events: %w", err)
	}

	return count, nil
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
		err = d.RelayDB.QueryRowContext(ctx, "SELECT COUNT(*) FROM event WHERE author = ?", pubkeyBytes).Scan(&count)
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
		SELECT author, COUNT(*) as count
		FROM event
		GROUP BY author
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

// CountEventsBefore counts events created before the given timestamp.
func (d *DB) CountEventsBefore(ctx context.Context, before time.Time) (int64, error) {
	if d.RelayDB == nil {
		return 0, fmt.Errorf("relay database not connected")
	}

	var count int64
	err := d.RelayDB.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM event WHERE created_at < ?",
		before.Unix(),
	).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count events: %w", err)
	}

	return count, nil
}

// EstimateEventSize estimates the average size of an event in bytes.
// This is a rough estimate used for storage calculations.
func (d *DB) EstimateEventSize(ctx context.Context) (int64, error) {
	if d.RelayDB == nil {
		return 0, fmt.Errorf("relay database not connected")
	}

	// Get average content length plus overhead for other fields
	var avgContentLen sql.NullFloat64
	err := d.RelayDB.QueryRowContext(ctx,
		"SELECT AVG(LENGTH(content)) FROM event",
	).Scan(&avgContentLen)
	if err != nil {
		return 0, fmt.Errorf("failed to estimate event size: %w", err)
	}

	if !avgContentLen.Valid {
		// No events, use a default estimate
		return 500, nil
	}

	// Add overhead for id (32), pubkey (32), sig (64), timestamp (8), kind (4), tags (~200 avg)
	overhead := int64(340)
	return int64(avgContentLen.Float64) + overhead, nil
}

// Helper functions

// nostrEventJSON represents the full Nostr event as stored in nostr-rs-relay's content column
type nostrEventJSON struct {
	ID        string     `json:"id"`
	Pubkey    string     `json:"pubkey"`
	CreatedAt int64      `json:"created_at"`
	Kind      int        `json:"kind"`
	Tags      [][]string `json:"tags"`
	Content   string     `json:"content"`
	Sig       string     `json:"sig"`
}

func scanEvent(row *sql.Row) (*Event, error) {
	var idBytes, authorBytes []byte
	var createdAt int64
	var kind int
	var contentJSON string

	err := row.Scan(&idBytes, &authorBytes, &createdAt, &kind, &contentJSON)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to scan event: %w", err)
	}

	return parseEventFromDB(idBytes, authorBytes, createdAt, kind, contentJSON)
}

func scanEventRows(rows *sql.Rows) (*Event, error) {
	var idBytes, authorBytes []byte
	var createdAt int64
	var kind int
	var contentJSON string

	err := rows.Scan(&idBytes, &authorBytes, &createdAt, &kind, &contentJSON)
	if err != nil {
		return nil, fmt.Errorf("failed to scan event: %w", err)
	}

	return parseEventFromDB(idBytes, authorBytes, createdAt, kind, contentJSON)
}

// parseEventFromDB parses an event from nostr-rs-relay's database format.
// The content column contains the full serialized JSON event.
func parseEventFromDB(idBytes, authorBytes []byte, createdAt int64, kind int, contentJSON string) (*Event, error) {
	// Parse the full event from the content JSON
	var eventData nostrEventJSON
	if err := json.Unmarshal([]byte(contentJSON), &eventData); err != nil {
		// If parsing fails, return a minimal event with available data
		return &Event{
			ID:        hex.EncodeToString(idBytes),
			Pubkey:    hex.EncodeToString(authorBytes),
			CreatedAt: time.Unix(createdAt, 0),
			Kind:      kind,
			Tags:      [][]string{},
			Content:   contentJSON, // Return raw JSON as content
			Sig:       "",
		}, nil
	}

	return &Event{
		ID:        hex.EncodeToString(idBytes),
		Pubkey:    hex.EncodeToString(authorBytes),
		CreatedAt: time.Unix(createdAt, 0),
		Kind:      kind,
		Tags:      eventData.Tags,
		Content:   eventData.Content,
		Sig:       eventData.Sig,
	}, nil
}

// DateCount represents event count for a specific date.
type DateCount struct {
	Date  string `json:"date"`
	Count int64  `json:"count"`
}

// AuthorCount represents event count for a specific author.
type AuthorCount struct {
	Pubkey     string `json:"pubkey"`
	EventCount int64  `json:"event_count"`
}

// GetEventsOverTime returns event counts grouped by date within a time range.
// If hourly is true, groups by hour and returns all 24 hours for the day.
// The loc parameter specifies the timezone for formatting timestamps.
func (d *DB) GetEventsOverTime(ctx context.Context, since, until time.Time, hourly bool, loc *time.Location) ([]DateCount, error) {
	if d.RelayDB == nil {
		return nil, fmt.Errorf("relay database not connected")
	}

	// Default to UTC if no location provided
	if loc == nil {
		loc = time.UTC
	}

	// Calculate timezone offset in seconds
	// We use the 'since' time to get the offset (handles DST correctly for that date)
	_, offset := since.In(loc).Zone()

	var query string
	args := []interface{}{}

	if hourly {
		// Apply timezone offset to created_at before formatting
		query = `
			SELECT strftime('%Y-%m-%d %H:00', datetime(created_at + ?, 'unixepoch')) as date, COUNT(*) as count
			FROM event
			WHERE 1=1
		`
		args = append(args, offset)
	} else {
		// Apply timezone offset to created_at before formatting
		query = `
			SELECT DATE(datetime(created_at + ?, 'unixepoch')) as date, COUNT(*) as count
			FROM event
			WHERE 1=1
		`
		args = append(args, offset)
	}

	if !since.IsZero() {
		query += " AND created_at >= ?"
		args = append(args, since.Unix())
	}
	if !until.IsZero() {
		query += " AND created_at <= ?"
		args = append(args, until.Unix())
	}

	query += " GROUP BY date ORDER BY date"

	rows, err := d.RelayDB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query events over time: %w", err)
	}
	defer rows.Close()

	var results []DateCount
	for rows.Next() {
		var dc DateCount
		if err := rows.Scan(&dc.Date, &dc.Count); err != nil {
			return nil, err
		}
		results = append(results, dc)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	// For hourly data, fill in all 24 hours with zeros for missing hours
	// For daily data, fill in all days in the range with zeros for missing days
	if hourly && !since.IsZero() {
		results = fillAllHours(results, since, loc)
	} else if !hourly && !since.IsZero() && !until.IsZero() {
		results = fillAllDays(results, since, until, loc)
	}

	return results, nil
}

// fillAllHours ensures all 24 hours are represented in the results, filling missing hours with zeros.
func fillAllHours(data []DateCount, day time.Time, loc *time.Location) []DateCount {
	// Create a map of existing data
	existing := make(map[string]int64)
	for _, d := range data {
		existing[d.Date] = d.Count
	}

	// Generate all 24 hours for the day in the specified timezone
	dayInLoc := day.In(loc)
	dayStart := time.Date(dayInLoc.Year(), dayInLoc.Month(), dayInLoc.Day(), 0, 0, 0, 0, loc)
	var filled []DateCount
	for h := 0; h < 24; h++ {
		hourTime := dayStart.Add(time.Duration(h) * time.Hour)
		dateStr := hourTime.Format("2006-01-02 15:00")
		count := existing[dateStr]
		filled = append(filled, DateCount{Date: dateStr, Count: count})
	}

	return filled
}

// fillAllDays ensures all days in the range are represented, filling missing days with zeros.
func fillAllDays(data []DateCount, since, until time.Time, loc *time.Location) []DateCount {
	// Create a map of existing data
	existing := make(map[string]int64)
	for _, d := range data {
		existing[d.Date] = d.Count
	}

	// Start from the beginning of 'since' day in the specified timezone
	sinceInLoc := since.In(loc)
	untilInLoc := until.In(loc)
	dayStart := time.Date(sinceInLoc.Year(), sinceInLoc.Month(), sinceInLoc.Day(), 0, 0, 0, 0, loc)
	dayEnd := time.Date(untilInLoc.Year(), untilInLoc.Month(), untilInLoc.Day(), 0, 0, 0, 0, loc)
	var filled []DateCount

	for day := dayStart; !day.After(dayEnd); day = day.AddDate(0, 0, 1) {
		dateStr := day.Format("2006-01-02")
		count := existing[dateStr]
		filled = append(filled, DateCount{Date: dateStr, Count: count})
	}

	return filled
}

// GetEventsByKindInRange returns event counts by kind within a time range.
func (d *DB) GetEventsByKindInRange(ctx context.Context, since, until time.Time) (map[int]int64, error) {
	if d.RelayDB == nil {
		return nil, fmt.Errorf("relay database not connected")
	}

	query := `SELECT kind, COUNT(*) as count FROM event WHERE 1=1`
	args := []interface{}{}

	if !since.IsZero() {
		query += " AND created_at >= ?"
		args = append(args, since.Unix())
	}
	if !until.IsZero() {
		query += " AND created_at <= ?"
		args = append(args, until.Unix())
	}

	query += " GROUP BY kind ORDER BY count DESC"

	rows, err := d.RelayDB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query events by kind: %w", err)
	}
	defer rows.Close()

	results := make(map[int]int64)
	for rows.Next() {
		var kind int
		var count int64
		if err := rows.Scan(&kind, &count); err != nil {
			return nil, err
		}
		results[kind] = count
	}

	return results, rows.Err()
}

// GetTopAuthorsInRange returns the top authors by event count within a time range.
func (d *DB) GetTopAuthorsInRange(ctx context.Context, limit int, since, until time.Time) ([]AuthorCount, error) {
	if d.RelayDB == nil {
		return nil, fmt.Errorf("relay database not connected")
	}

	if limit <= 0 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}

	query := `SELECT author, COUNT(*) as count FROM event WHERE 1=1`
	args := []interface{}{}

	if !since.IsZero() {
		query += " AND created_at >= ?"
		args = append(args, since.Unix())
	}
	if !until.IsZero() {
		query += " AND created_at <= ?"
		args = append(args, until.Unix())
	}

	query += " GROUP BY author ORDER BY count DESC LIMIT ?"
	args = append(args, limit)

	rows, err := d.RelayDB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query top authors: %w", err)
	}
	defer rows.Close()

	var authors []AuthorCount
	for rows.Next() {
		var pubkeyBytes []byte
		var count int64
		if err := rows.Scan(&pubkeyBytes, &count); err != nil {
			return nil, err
		}
		authors = append(authors, AuthorCount{
			Pubkey:     hex.EncodeToString(pubkeyBytes),
			EventCount: count,
		})
	}

	return authors, rows.Err()
}

// CountEvents counts events matching the filter (for export progress tracking).
func (d *DB) CountEvents(ctx context.Context, filter EventFilter) (int64, error) {
	if d.RelayDB == nil {
		return 0, fmt.Errorf("relay database not connected")
	}

	// Build count query with same WHERE clauses as GetEvents
	query := `SELECT COUNT(*) FROM event WHERE 1=1`
	args := []interface{}{}

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

	var count int64
	err := d.RelayDB.QueryRowContext(ctx, query, args...).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count events: %w", err)
	}

	return count, nil
}

// StreamEvents streams events matching the filter to the callback function.
// Used for exports to avoid loading all events into memory.
func (d *DB) StreamEvents(ctx context.Context, filter EventFilter, callback func(ExportEvent) error) error {
	if d.RelayDB == nil {
		return fmt.Errorf("relay database not connected")
	}

	// Build query (same WHERE clauses as GetEvents, but no LIMIT for full export)
	// nostr-rs-relay uses event_hash for ID, author for pubkey, and stores full event JSON in content
	query := `SELECT event_hash, author, created_at, kind, content FROM event WHERE 1=1`
	args := []interface{}{}

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

	// Order by created_at for consistent export ordering
	query += " ORDER BY created_at ASC"

	rows, err := d.RelayDB.QueryContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to query events: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		// Check for context cancellation
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		var idBytes, authorBytes []byte
		var dbCreatedAt int64
		var kind int
		var contentJSON string

		err := rows.Scan(&idBytes, &authorBytes, &dbCreatedAt, &kind, &contentJSON)
		if err != nil {
			return fmt.Errorf("failed to scan event: %w", err)
		}

		// Parse the full event from the content JSON
		var eventData nostrEventJSON
		var event ExportEvent

		if err := json.Unmarshal([]byte(contentJSON), &eventData); err != nil {
			// If parsing fails, use available data
			event = ExportEvent{
				ID:        hex.EncodeToString(idBytes),
				Pubkey:    hex.EncodeToString(authorBytes),
				CreatedAt: dbCreatedAt,
				Kind:      kind,
				Tags:      [][]string{},
				Content:   contentJSON,
				Sig:       "",
			}
		} else {
			event = ExportEvent{
				ID:        hex.EncodeToString(idBytes),
				Pubkey:    hex.EncodeToString(authorBytes),
				CreatedAt: dbCreatedAt,
				Kind:      kind,
				Tags:      eventData.Tags,
				Content:   eventData.Content,
				Sig:       eventData.Sig,
			}
		}

		if err := callback(event); err != nil {
			return err
		}
	}

	return rows.Err()
}
