package db

import (
	"context"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"os"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// setupTestRelayDB creates a test DB with a relay database for testing relay reader functions.
func setupTestRelayDB(t *testing.T) *DB {
	t.Helper()

	// Create temp file for relay database
	relayFile, err := os.CreateTemp("", "relay_test_*.db")
	if err != nil {
		t.Fatalf("failed to create temp relay db file: %v", err)
	}
	relayPath := relayFile.Name()
	relayFile.Close()
	t.Cleanup(func() { os.Remove(relayPath) })

	// Create temp file for app database (needed for DB struct)
	appFile, err := os.CreateTemp("", "app_test_*.db")
	if err != nil {
		t.Fatalf("failed to create temp app db file: %v", err)
	}
	appPath := appFile.Name()
	appFile.Close()
	t.Cleanup(func() { os.Remove(appPath) })

	// Open relay database directly
	relayDB, err := sql.Open("sqlite3", relayPath)
	if err != nil {
		t.Fatalf("failed to open relay db: %v", err)
	}
	t.Cleanup(func() { relayDB.Close() })

	// Create relay event table matching nostr-rs-relay schema
	_, err = relayDB.Exec(`
		CREATE TABLE event (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			event_hash BLOB NOT NULL UNIQUE,
			first_seen INTEGER NOT NULL,
			created_at INTEGER,
			author BLOB NOT NULL,
			delegated_by BLOB,
			kind INTEGER,
			hidden INTEGER,
			content TEXT NOT NULL
		);
		CREATE INDEX idx_event_author ON event(author);
		CREATE INDEX idx_event_kind ON event(kind);
		CREATE INDEX idx_event_created_at ON event(created_at);
	`)
	if err != nil {
		t.Fatalf("failed to create relay schema: %v", err)
	}

	// Use New() for app database to get proper initialization
	database, err := New("", appPath)
	if err != nil {
		t.Fatalf("failed to create database: %v", err)
	}
	t.Cleanup(func() { database.Close() })

	// Run migrations on app DB
	if err := database.Migrate(context.Background()); err != nil {
		t.Fatalf("failed to migrate database: %v", err)
	}

	// Attach the relay database
	database.RelayDB = relayDB

	return database
}

// insertTestEvent inserts a test event into the relay database.
func insertTestEvent(t *testing.T, db *sql.DB, id, pubkey string, kind int, createdAt time.Time, content string) {
	t.Helper()

	idBytes, _ := hex.DecodeString(id)
	pubkeyBytes, _ := hex.DecodeString(pubkey)

	// Create the full event JSON as nostr-rs-relay stores it
	eventJSON := map[string]interface{}{
		"id":         id,
		"pubkey":     pubkey,
		"created_at": createdAt.Unix(),
		"kind":       kind,
		"tags":       [][]string{},
		"content":    content,
		"sig":        "0000000000000000000000000000000000000000000000000000000000000000",
	}
	contentBytes, _ := json.Marshal(eventJSON)

	_, err := db.Exec(`
		INSERT INTO event (event_hash, first_seen, created_at, author, kind, hidden, content)
		VALUES (?, ?, ?, ?, ?, 0, ?)
	`, idBytes, time.Now().Unix(), createdAt.Unix(), pubkeyBytes, kind, string(contentBytes))
	if err != nil {
		t.Fatalf("failed to insert test event: %v", err)
	}
}

// insertTestEventWithTags inserts a test event with tags into the relay database.
func insertTestEventWithTags(t *testing.T, db *sql.DB, id, pubkey string, kind int, createdAt time.Time, content string, tags [][]string) {
	t.Helper()

	idBytes, _ := hex.DecodeString(id)
	pubkeyBytes, _ := hex.DecodeString(pubkey)

	// Create the full event JSON as nostr-rs-relay stores it
	eventJSON := map[string]interface{}{
		"id":         id,
		"pubkey":     pubkey,
		"created_at": createdAt.Unix(),
		"kind":       kind,
		"tags":       tags,
		"content":    content,
		"sig":        "0000000000000000000000000000000000000000000000000000000000000000",
	}
	contentBytes, _ := json.Marshal(eventJSON)

	_, err := db.Exec(`
		INSERT INTO event (event_hash, first_seen, created_at, author, kind, hidden, content)
		VALUES (?, ?, ?, ?, ?, 0, ?)
	`, idBytes, time.Now().Unix(), createdAt.Unix(), pubkeyBytes, kind, string(contentBytes))
	if err != nil {
		t.Fatalf("failed to insert test event with tags: %v", err)
	}
}

// Test pubkeys and IDs (valid 64-char hex strings)
const (
	testPubkey1 = "1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"
	testPubkey2 = "fedcba0987654321fedcba0987654321fedcba0987654321fedcba0987654321"
	testPubkey3 = "abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890"
	testEventID1 = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	testEventID2 = "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
	testEventID3 = "cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc"
	testEventID4 = "dddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddd"
	testEventID5 = "eeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee"
)

// ============================================================================
// GetEvent Tests
// ============================================================================

func TestGetEvent(t *testing.T) {
	db := setupTestRelayDB(t)
	ctx := context.Background()

	// Insert a test event
	now := time.Now().Truncate(time.Second)
	insertTestEvent(t, db.RelayDB, testEventID1, testPubkey1, 1, now, "Hello world")

	t.Run("GetEvent_existing", func(t *testing.T) {
		event, err := db.GetEvent(ctx, testEventID1)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if event == nil {
			t.Fatal("expected event, got nil")
		}
		if event.ID != testEventID1 {
			t.Errorf("expected ID %s, got %s", testEventID1, event.ID)
		}
		if event.Pubkey != testPubkey1 {
			t.Errorf("expected pubkey %s, got %s", testPubkey1, event.Pubkey)
		}
		if event.Kind != 1 {
			t.Errorf("expected kind 1, got %d", event.Kind)
		}
		if event.Content != "Hello world" {
			t.Errorf("expected content 'Hello world', got %s", event.Content)
		}
	})

	t.Run("GetEvent_not_found", func(t *testing.T) {
		event, err := db.GetEvent(ctx, testEventID2)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if event != nil {
			t.Errorf("expected nil, got event")
		}
	})

	t.Run("GetEvent_invalid_id", func(t *testing.T) {
		_, err := db.GetEvent(ctx, "not-hex")
		if err == nil {
			t.Error("expected error for invalid hex ID")
		}
	})
}

// ============================================================================
// GetEvents (Filtering) Tests
// ============================================================================

func TestGetEvents(t *testing.T) {
	db := setupTestRelayDB(t)
	ctx := context.Background()

	// Insert test events
	now := time.Now().Truncate(time.Second)
	insertTestEvent(t, db.RelayDB, testEventID1, testPubkey1, 1, now, "First note")
	insertTestEvent(t, db.RelayDB, testEventID2, testPubkey1, 1, now.Add(-time.Hour), "Second note")
	insertTestEvent(t, db.RelayDB, testEventID3, testPubkey2, 0, now.Add(-2*time.Hour), "Metadata event")
	insertTestEvent(t, db.RelayDB, testEventID4, testPubkey2, 3, now.Add(-3*time.Hour), "Contact list")
	insertTestEvent(t, db.RelayDB, testEventID5, testPubkey1, 4, now.Add(-4*time.Hour), "Encrypted message")

	t.Run("GetEvents_no_filter", func(t *testing.T) {
		events, err := db.GetEvents(ctx, EventFilter{})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(events) != 5 {
			t.Errorf("expected 5 events, got %d", len(events))
		}
	})

	t.Run("GetEvents_filter_by_kind", func(t *testing.T) {
		events, err := db.GetEvents(ctx, EventFilter{Kinds: []int{1}})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(events) != 2 {
			t.Errorf("expected 2 events of kind 1, got %d", len(events))
		}
	})

	t.Run("GetEvents_filter_by_author", func(t *testing.T) {
		events, err := db.GetEvents(ctx, EventFilter{Authors: []string{testPubkey1}})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(events) != 3 {
			t.Errorf("expected 3 events from pubkey1, got %d", len(events))
		}
	})

	t.Run("GetEvents_filter_by_multiple_kinds", func(t *testing.T) {
		events, err := db.GetEvents(ctx, EventFilter{Kinds: []int{0, 3}})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(events) != 2 {
			t.Errorf("expected 2 events of kinds 0 or 3, got %d", len(events))
		}
	})

	t.Run("GetEvents_filter_by_id", func(t *testing.T) {
		events, err := db.GetEvents(ctx, EventFilter{IDs: []string{testEventID1, testEventID3}})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(events) != 2 {
			t.Errorf("expected 2 specific events, got %d", len(events))
		}
	})

	t.Run("GetEvents_filter_since", func(t *testing.T) {
		events, err := db.GetEvents(ctx, EventFilter{Since: now.Add(-90 * time.Minute)})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(events) != 2 {
			t.Errorf("expected 2 events after 90 min ago, got %d", len(events))
		}
	})

	t.Run("GetEvents_filter_until", func(t *testing.T) {
		events, err := db.GetEvents(ctx, EventFilter{Until: now.Add(-90 * time.Minute)})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(events) != 3 {
			t.Errorf("expected 3 events before 90 min ago, got %d", len(events))
		}
	})

	t.Run("GetEvents_filter_search", func(t *testing.T) {
		events, err := db.GetEvents(ctx, EventFilter{Search: "note"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(events) != 2 {
			t.Errorf("expected 2 events containing 'note', got %d", len(events))
		}
	})

	t.Run("GetEvents_limit", func(t *testing.T) {
		events, err := db.GetEvents(ctx, EventFilter{Limit: 2})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(events) != 2 {
			t.Errorf("expected 2 events (limit), got %d", len(events))
		}
	})

	t.Run("GetEvents_offset", func(t *testing.T) {
		events, err := db.GetEvents(ctx, EventFilter{Limit: 2, Offset: 2})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(events) != 2 {
			t.Errorf("expected 2 events (offset), got %d", len(events))
		}
	})

	t.Run("GetEvents_ordered_by_created_at_desc", func(t *testing.T) {
		events, err := db.GetEvents(ctx, EventFilter{})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		// First event should be the newest
		if events[0].ID != testEventID1 {
			t.Errorf("expected newest event first, got %s", events[0].ID)
		}
	})
}

// ============================================================================
// GetEvents with Mentions Tests
// ============================================================================

func TestGetEventsWithMentions(t *testing.T) {
	db := setupTestRelayDB(t)
	ctx := context.Background()

	now := time.Now().Truncate(time.Second)

	// Insert event with p tag mentioning testPubkey2
	tags := [][]string{{"p", testPubkey2}}
	insertTestEventWithTags(t, db.RelayDB, testEventID1, testPubkey1, 1, now, "Mentioning someone", tags)

	// Insert event without mentions
	insertTestEvent(t, db.RelayDB, testEventID2, testPubkey1, 1, now, "No mentions")

	t.Run("GetEvents_filter_mentions", func(t *testing.T) {
		events, err := db.GetEvents(ctx, EventFilter{Mentions: testPubkey2})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(events) != 1 {
			t.Errorf("expected 1 event mentioning pubkey2, got %d", len(events))
		}
		if len(events) > 0 && events[0].ID != testEventID1 {
			t.Errorf("expected event %s, got %s", testEventID1, events[0].ID)
		}
	})
}

// ============================================================================
// GetRecentEvents Tests
// ============================================================================

func TestGetRecentEvents(t *testing.T) {
	db := setupTestRelayDB(t)
	ctx := context.Background()

	now := time.Now().Truncate(time.Second)
	insertTestEvent(t, db.RelayDB, testEventID1, testPubkey1, 1, now, "Event 1")
	insertTestEvent(t, db.RelayDB, testEventID2, testPubkey1, 1, now.Add(-time.Hour), "Event 2")
	insertTestEvent(t, db.RelayDB, testEventID3, testPubkey1, 1, now.Add(-2*time.Hour), "Event 3")

	t.Run("GetRecentEvents", func(t *testing.T) {
		events, err := db.GetRecentEvents(ctx, 2)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(events) != 2 {
			t.Errorf("expected 2 events, got %d", len(events))
		}
		// Should be ordered newest first
		if len(events) > 0 && events[0].ID != testEventID1 {
			t.Errorf("expected newest event first")
		}
	})
}

// ============================================================================
// GetRelayStats Tests
// ============================================================================

func TestGetRelayStats(t *testing.T) {
	db := setupTestRelayDB(t)
	ctx := context.Background()

	t.Run("GetRelayStats_empty", func(t *testing.T) {
		stats, err := db.GetRelayStats(ctx)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if stats.TotalEvents != 0 {
			t.Errorf("expected 0 events, got %d", stats.TotalEvents)
		}
		if stats.TotalPubkeys != 0 {
			t.Errorf("expected 0 pubkeys, got %d", stats.TotalPubkeys)
		}
	})

	// Insert test events
	now := time.Now().Truncate(time.Second)
	insertTestEvent(t, db.RelayDB, testEventID1, testPubkey1, 1, now, "Event 1")
	insertTestEvent(t, db.RelayDB, testEventID2, testPubkey1, 1, now.Add(-time.Hour), "Event 2")
	insertTestEvent(t, db.RelayDB, testEventID3, testPubkey2, 0, now.Add(-2*time.Hour), "Event 3")
	insertTestEvent(t, db.RelayDB, testEventID4, testPubkey2, 3, now.Add(-3*time.Hour), "Event 4")

	t.Run("GetRelayStats_with_data", func(t *testing.T) {
		stats, err := db.GetRelayStats(ctx)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if stats.TotalEvents != 4 {
			t.Errorf("expected 4 events, got %d", stats.TotalEvents)
		}
		if stats.TotalPubkeys != 2 {
			t.Errorf("expected 2 unique pubkeys, got %d", stats.TotalPubkeys)
		}
		if stats.EventsByKind[1] != 2 {
			t.Errorf("expected 2 events of kind 1, got %d", stats.EventsByKind[1])
		}
		if stats.EventsByKind[0] != 1 {
			t.Errorf("expected 1 event of kind 0, got %d", stats.EventsByKind[0])
		}
		if stats.OldestEvent.IsZero() {
			t.Error("expected oldest event timestamp")
		}
		if stats.NewestEvent.IsZero() {
			t.Error("expected newest event timestamp")
		}
	})
}

// ============================================================================
// GetEventsToday Tests
// ============================================================================

func TestGetEventsToday(t *testing.T) {
	db := setupTestRelayDB(t)
	ctx := context.Background()

	now := time.Now().UTC()
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)

	// Insert event from today
	insertTestEvent(t, db.RelayDB, testEventID1, testPubkey1, 1, startOfDay.Add(time.Hour), "Today's event")
	// Insert event from yesterday
	insertTestEvent(t, db.RelayDB, testEventID2, testPubkey1, 1, startOfDay.Add(-time.Hour), "Yesterday's event")

	t.Run("GetEventsToday", func(t *testing.T) {
		count, err := db.GetEventsToday(ctx, time.UTC)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if count != 1 {
			t.Errorf("expected 1 event today, got %d", count)
		}
	})
}

// ============================================================================
// CountEventsByPubkey Tests
// ============================================================================

func TestCountEventsByPubkey(t *testing.T) {
	db := setupTestRelayDB(t)
	ctx := context.Background()

	now := time.Now().Truncate(time.Second)
	insertTestEvent(t, db.RelayDB, testEventID1, testPubkey1, 1, now, "Event 1")
	insertTestEvent(t, db.RelayDB, testEventID2, testPubkey1, 1, now.Add(-time.Hour), "Event 2")
	insertTestEvent(t, db.RelayDB, testEventID3, testPubkey2, 1, now.Add(-2*time.Hour), "Event 3")

	t.Run("CountEventsByPubkey", func(t *testing.T) {
		counts, err := db.CountEventsByPubkey(ctx, []string{testPubkey1, testPubkey2, testPubkey3})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if counts[testPubkey1] != 2 {
			t.Errorf("expected 2 events for pubkey1, got %d", counts[testPubkey1])
		}
		if counts[testPubkey2] != 1 {
			t.Errorf("expected 1 event for pubkey2, got %d", counts[testPubkey2])
		}
		if counts[testPubkey3] != 0 {
			t.Errorf("expected 0 events for pubkey3, got %d", counts[testPubkey3])
		}
	})
}

// ============================================================================
// GetTopAuthors Tests
// ============================================================================

func TestGetTopAuthors(t *testing.T) {
	db := setupTestRelayDB(t)
	ctx := context.Background()

	now := time.Now().Truncate(time.Second)
	// pubkey1 has 3 events, pubkey2 has 2, pubkey3 has 1
	insertTestEvent(t, db.RelayDB, testEventID1, testPubkey1, 1, now, "Event 1")
	insertTestEvent(t, db.RelayDB, testEventID2, testPubkey1, 1, now.Add(-time.Hour), "Event 2")
	insertTestEvent(t, db.RelayDB, testEventID3, testPubkey1, 1, now.Add(-2*time.Hour), "Event 3")
	insertTestEvent(t, db.RelayDB, testEventID4, testPubkey2, 1, now.Add(-3*time.Hour), "Event 4")
	insertTestEvent(t, db.RelayDB, testEventID5, testPubkey2, 1, now.Add(-4*time.Hour), "Event 5")
	insertTestEvent(t, db.RelayDB, "ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff", testPubkey3, 1, now.Add(-5*time.Hour), "Event 6")

	t.Run("GetTopAuthors_default_limit", func(t *testing.T) {
		authors, err := db.GetTopAuthors(ctx, 0)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(authors) != 3 {
			t.Errorf("expected 3 authors, got %d", len(authors))
		}
		// First should be pubkey1 with 3 events
		if len(authors) > 0 && authors[0].Pubkey != testPubkey1 {
			t.Errorf("expected pubkey1 first, got %s", authors[0].Pubkey)
		}
		if len(authors) > 0 && authors[0].EventCount != 3 {
			t.Errorf("expected 3 events for top author, got %d", authors[0].EventCount)
		}
	})

	t.Run("GetTopAuthors_limited", func(t *testing.T) {
		authors, err := db.GetTopAuthors(ctx, 2)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(authors) != 2 {
			t.Errorf("expected 2 authors (limit), got %d", len(authors))
		}
	})
}

// ============================================================================
// CountEventsBefore Tests
// ============================================================================

func TestCountEventsBefore(t *testing.T) {
	db := setupTestRelayDB(t)
	ctx := context.Background()

	now := time.Now().Truncate(time.Second)
	insertTestEvent(t, db.RelayDB, testEventID1, testPubkey1, 1, now, "Event 1")
	insertTestEvent(t, db.RelayDB, testEventID2, testPubkey1, 1, now.Add(-2*time.Hour), "Event 2")
	insertTestEvent(t, db.RelayDB, testEventID3, testPubkey1, 1, now.Add(-4*time.Hour), "Event 3")

	t.Run("CountEventsBefore", func(t *testing.T) {
		count, err := db.CountEventsBefore(ctx, now.Add(-time.Hour))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if count != 2 {
			t.Errorf("expected 2 events before 1 hour ago, got %d", count)
		}
	})
}

// ============================================================================
// EstimateEventSize Tests
// ============================================================================

func TestEstimateEventSize(t *testing.T) {
	db := setupTestRelayDB(t)
	ctx := context.Background()

	t.Run("EstimateEventSize_empty", func(t *testing.T) {
		size, err := db.EstimateEventSize(ctx)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		// Default estimate for empty database
		if size != 500 {
			t.Errorf("expected default estimate 500, got %d", size)
		}
	})

	now := time.Now()
	insertTestEvent(t, db.RelayDB, testEventID1, testPubkey1, 1, now, "Hello world")

	t.Run("EstimateEventSize_with_data", func(t *testing.T) {
		size, err := db.EstimateEventSize(ctx)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		// Should be greater than the overhead (340)
		if size < 340 {
			t.Errorf("expected size > 340, got %d", size)
		}
	})
}

// ============================================================================
// GetEventsOverTime Tests
// ============================================================================

func TestGetEventsOverTime(t *testing.T) {
	db := setupTestRelayDB(t)
	ctx := context.Background()

	now := time.Now().UTC().Truncate(time.Hour)
	yesterday := now.AddDate(0, 0, -1)
	twoDaysAgo := now.AddDate(0, 0, -2)

	insertTestEvent(t, db.RelayDB, testEventID1, testPubkey1, 1, now.Add(-time.Hour), "Event 1")
	insertTestEvent(t, db.RelayDB, testEventID2, testPubkey1, 1, yesterday, "Event 2")
	insertTestEvent(t, db.RelayDB, testEventID3, testPubkey1, 1, yesterday.Add(-time.Hour), "Event 3")
	insertTestEvent(t, db.RelayDB, testEventID4, testPubkey1, 1, twoDaysAgo, "Event 4")

	t.Run("GetEventsOverTime_daily", func(t *testing.T) {
		results, err := db.GetEventsOverTime(ctx, twoDaysAgo, now, false, time.UTC)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(results) < 2 {
			t.Errorf("expected at least 2 date counts, got %d", len(results))
		}
	})

	t.Run("GetEventsOverTime_hourly", func(t *testing.T) {
		results, err := db.GetEventsOverTime(ctx, now.Add(-24*time.Hour), now, true, time.UTC)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		// Should have 24 hours filled in
		if len(results) != 24 {
			t.Errorf("expected 24 hour counts, got %d", len(results))
		}
	})
}

// ============================================================================
// GetEventsByKindInRange Tests
// ============================================================================

func TestGetEventsByKindInRange(t *testing.T) {
	db := setupTestRelayDB(t)
	ctx := context.Background()

	now := time.Now().Truncate(time.Second)
	yesterday := now.AddDate(0, 0, -1)

	insertTestEvent(t, db.RelayDB, testEventID1, testPubkey1, 1, now, "Note 1")
	insertTestEvent(t, db.RelayDB, testEventID2, testPubkey1, 1, now.Add(-time.Hour), "Note 2")
	insertTestEvent(t, db.RelayDB, testEventID3, testPubkey1, 0, now.Add(-2*time.Hour), "Metadata")
	insertTestEvent(t, db.RelayDB, testEventID4, testPubkey1, 3, yesterday, "Contact list")

	t.Run("GetEventsByKindInRange_all", func(t *testing.T) {
		results, err := db.GetEventsByKindInRange(ctx, time.Time{}, time.Time{})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if results[1] != 2 {
			t.Errorf("expected 2 events of kind 1, got %d", results[1])
		}
		if results[0] != 1 {
			t.Errorf("expected 1 event of kind 0, got %d", results[0])
		}
	})

	t.Run("GetEventsByKindInRange_filtered", func(t *testing.T) {
		results, err := db.GetEventsByKindInRange(ctx, now.Add(-3*time.Hour), now)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		// Should not include the contact list from yesterday
		if results[3] != 0 {
			t.Errorf("expected 0 events of kind 3 in range, got %d", results[3])
		}
	})
}

// ============================================================================
// GetTopAuthorsInRange Tests
// ============================================================================

func TestGetTopAuthorsInRange(t *testing.T) {
	db := setupTestRelayDB(t)
	ctx := context.Background()

	now := time.Now().Truncate(time.Second)
	yesterday := now.AddDate(0, 0, -1)

	// Today: pubkey1 has 2 events, pubkey2 has 1
	insertTestEvent(t, db.RelayDB, testEventID1, testPubkey1, 1, now, "Event 1")
	insertTestEvent(t, db.RelayDB, testEventID2, testPubkey1, 1, now.Add(-time.Hour), "Event 2")
	insertTestEvent(t, db.RelayDB, testEventID3, testPubkey2, 1, now.Add(-2*time.Hour), "Event 3")
	// Yesterday: pubkey2 has 2 events (making total 3)
	insertTestEvent(t, db.RelayDB, testEventID4, testPubkey2, 1, yesterday, "Event 4")
	insertTestEvent(t, db.RelayDB, testEventID5, testPubkey2, 1, yesterday.Add(-time.Hour), "Event 5")

	t.Run("GetTopAuthorsInRange_today_only", func(t *testing.T) {
		authors, err := db.GetTopAuthorsInRange(ctx, 10, now.Add(-3*time.Hour), now)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(authors) != 2 {
			t.Errorf("expected 2 authors, got %d", len(authors))
		}
		// pubkey1 should be first with 2 events today
		if len(authors) > 0 && authors[0].Pubkey != testPubkey1 {
			t.Errorf("expected pubkey1 first for today's range, got %s", authors[0].Pubkey)
		}
	})
}

// ============================================================================
// CountEvents Tests
// ============================================================================

func TestCountEvents(t *testing.T) {
	db := setupTestRelayDB(t)
	ctx := context.Background()

	now := time.Now().Truncate(time.Second)
	insertTestEvent(t, db.RelayDB, testEventID1, testPubkey1, 1, now, "Note")
	insertTestEvent(t, db.RelayDB, testEventID2, testPubkey1, 0, now.Add(-time.Hour), "Metadata")
	insertTestEvent(t, db.RelayDB, testEventID3, testPubkey1, 3, now.Add(-2*time.Hour), "Contacts")

	t.Run("CountEvents_all", func(t *testing.T) {
		count, err := db.CountEvents(ctx, EventFilter{})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if count != 3 {
			t.Errorf("expected 3 events, got %d", count)
		}
	})

	t.Run("CountEvents_by_kind", func(t *testing.T) {
		count, err := db.CountEvents(ctx, EventFilter{Kinds: []int{1}})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if count != 1 {
			t.Errorf("expected 1 event of kind 1, got %d", count)
		}
	})

	t.Run("CountEvents_since", func(t *testing.T) {
		count, err := db.CountEvents(ctx, EventFilter{Since: now.Add(-90 * time.Minute)})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if count != 2 {
			t.Errorf("expected 2 events since 90 min ago, got %d", count)
		}
	})
}

// ============================================================================
// StreamEvents Tests
// ============================================================================

func TestStreamEvents(t *testing.T) {
	db := setupTestRelayDB(t)
	ctx := context.Background()

	now := time.Now().Truncate(time.Second)
	insertTestEvent(t, db.RelayDB, testEventID1, testPubkey1, 1, now, "Note 1")
	insertTestEvent(t, db.RelayDB, testEventID2, testPubkey1, 1, now.Add(-time.Hour), "Note 2")
	insertTestEvent(t, db.RelayDB, testEventID3, testPubkey1, 0, now.Add(-2*time.Hour), "Metadata")

	t.Run("StreamEvents_all", func(t *testing.T) {
		var events []ExportEvent
		err := db.StreamEvents(ctx, EventFilter{}, func(e ExportEvent) error {
			events = append(events, e)
			return nil
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(events) != 3 {
			t.Errorf("expected 3 events, got %d", len(events))
		}
	})

	t.Run("StreamEvents_filtered_by_kind", func(t *testing.T) {
		var events []ExportEvent
		err := db.StreamEvents(ctx, EventFilter{Kinds: []int{1}}, func(e ExportEvent) error {
			events = append(events, e)
			return nil
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(events) != 2 {
			t.Errorf("expected 2 events of kind 1, got %d", len(events))
		}
	})

	t.Run("StreamEvents_ordered_asc", func(t *testing.T) {
		var events []ExportEvent
		err := db.StreamEvents(ctx, EventFilter{}, func(e ExportEvent) error {
			events = append(events, e)
			return nil
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		// Should be ordered oldest first (ASC)
		if len(events) > 0 && events[0].ID != testEventID3 {
			t.Errorf("expected oldest event first in stream, got %s", events[0].ID)
		}
	})

	t.Run("StreamEvents_callback_error", func(t *testing.T) {
		callbackErr := &CallbackError{}
		err := db.StreamEvents(ctx, EventFilter{}, func(e ExportEvent) error {
			return callbackErr
		})
		if err != callbackErr {
			t.Errorf("expected callback error to be returned")
		}
	})
}

// CallbackError is a test error type for StreamEvents callback testing.
type CallbackError struct{}

func (e *CallbackError) Error() string { return "callback error" }

// ============================================================================
// Nil RelayDB Tests
// ============================================================================

func TestRelayReaderNilRelayDB(t *testing.T) {
	// Create DB without relay database
	db := setupTestDB(t)
	ctx := context.Background()

	t.Run("GetEvent_nil_relay", func(t *testing.T) {
		_, err := db.GetEvent(ctx, testEventID1)
		if err == nil {
			t.Error("expected error for nil relay database")
		}
	})

	t.Run("GetEvents_nil_relay", func(t *testing.T) {
		_, err := db.GetEvents(ctx, EventFilter{})
		if err == nil {
			t.Error("expected error for nil relay database")
		}
	})

	t.Run("GetRelayStats_nil_relay", func(t *testing.T) {
		_, err := db.GetRelayStats(ctx)
		if err == nil {
			t.Error("expected error for nil relay database")
		}
	})

	t.Run("CountEvents_nil_relay", func(t *testing.T) {
		_, err := db.CountEvents(ctx, EventFilter{})
		if err == nil {
			t.Error("expected error for nil relay database")
		}
	})

	t.Run("StreamEvents_nil_relay", func(t *testing.T) {
		err := db.StreamEvents(ctx, EventFilter{}, func(e ExportEvent) error { return nil })
		if err == nil {
			t.Error("expected error for nil relay database")
		}
	})
}
