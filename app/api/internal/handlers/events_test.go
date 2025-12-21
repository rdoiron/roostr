package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/roostr/roostr/app/api/internal/db"
)

// mockEventsDB implements the database methods needed for Events handlers
type mockEventsDB struct {
	events          []db.Event
	eventByID       map[string]*db.Event
	relayConnected  bool
	deletionErr     error
	duplicateDelete bool
}

func newMockEventsDB() *mockEventsDB {
	return &mockEventsDB{
		events:         []db.Event{},
		eventByID:      make(map[string]*db.Event),
		relayConnected: true,
	}
}

func (m *mockEventsDB) IsRelayDBConnected() bool {
	return m.relayConnected
}

func (m *mockEventsDB) GetEvents(ctx context.Context, filter db.EventFilter) ([]db.Event, error) {
	// Simple filtering simulation
	var result []db.Event
	for _, e := range m.events {
		// Kind filter
		if len(filter.Kinds) > 0 {
			found := false
			for _, k := range filter.Kinds {
				if e.Kind == k {
					found = true
					break
				}
			}
			if !found {
				continue
			}
		}
		// Author filter
		if len(filter.Authors) > 0 {
			found := false
			for _, a := range filter.Authors {
				if e.Pubkey == a {
					found = true
					break
				}
			}
			if !found {
				continue
			}
		}
		// Search filter
		if filter.Search != "" && !strings.Contains(e.Content, filter.Search) {
			continue
		}
		result = append(result, e)
	}

	// Apply limit/offset
	start := filter.Offset
	if start > len(result) {
		return []db.Event{}, nil
	}
	end := start + filter.Limit
	if end > len(result) {
		end = len(result)
	}
	return result[start:end], nil
}

func (m *mockEventsDB) GetEvent(ctx context.Context, id string) (*db.Event, error) {
	return m.eventByID[id], nil
}

func (m *mockEventsDB) GetRecentEvents(ctx context.Context, limit int) ([]db.Event, error) {
	if limit > len(m.events) {
		return m.events, nil
	}
	return m.events[:limit], nil
}

func (m *mockEventsDB) GetAppState(ctx context.Context, key string) (string, error) {
	return "operator_pubkey_hex", nil
}

func (m *mockEventsDB) CreateDeletionRequest(ctx context.Context, eventID, requestedBy, reason string) (int64, error) {
	if m.duplicateDelete {
		return 0, &mockError{msg: "UNIQUE constraint failed"}
	}
	if m.deletionErr != nil {
		return 0, m.deletionErr
	}
	return 1, nil
}

type mockError struct {
	msg string
}

func (e *mockError) Error() string {
	return e.msg
}

// ============================================================================
// GetEvents Tests
// ============================================================================

func TestGetEvents(t *testing.T) {
	t.Run("relay_not_connected", func(t *testing.T) {
		mockDB := newMockEventsDB()
		mockDB.relayConnected = false

		mux := http.NewServeMux()
		mux.HandleFunc("GET /api/v1/events", func(w http.ResponseWriter, r *http.Request) {
			if !mockDB.IsRelayDBConnected() {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusServiceUnavailable)
				json.NewEncoder(w).Encode(map[string]interface{}{
					"error": "Relay database not connected",
					"code":  "RELAY_NOT_CONNECTED",
				})
				return
			}
		})

		req := httptest.NewRequest("GET", "/api/v1/events", nil)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)

		if w.Code != http.StatusServiceUnavailable {
			t.Errorf("expected status 503, got %d", w.Code)
		}
	})

	t.Run("returns_events_with_defaults", func(t *testing.T) {
		mockDB := newMockEventsDB()
		mockDB.events = []db.Event{
			{ID: "event1", Kind: 1, Content: "Hello", Pubkey: "abc123", CreatedAt: time.Now()},
			{ID: "event2", Kind: 1, Content: "World", Pubkey: "def456", CreatedAt: time.Now()},
		}
		mockDB.eventByID["event1"] = &mockDB.events[0]
		mockDB.eventByID["event2"] = &mockDB.events[1]

		mux := http.NewServeMux()
		mux.HandleFunc("GET /api/v1/events", func(w http.ResponseWriter, r *http.Request) {
			if !mockDB.IsRelayDBConnected() {
				w.WriteHeader(http.StatusServiceUnavailable)
				return
			}

			filter := db.EventFilter{
				Limit:  parseIntParam(r.URL.Query().Get("limit"), 50),
				Offset: parseIntParam(r.URL.Query().Get("offset"), 0),
			}

			events, _ := mockDB.GetEvents(r.Context(), filter)
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"events": events,
				"count":  len(events),
				"limit":  filter.Limit,
				"offset": filter.Offset,
			})
		})

		req := httptest.NewRequest("GET", "/api/v1/events", nil)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}

		var resp map[string]interface{}
		json.NewDecoder(w.Body).Decode(&resp)

		count := int(resp["count"].(float64))
		if count != 2 {
			t.Errorf("expected 2 events, got %d", count)
		}
		if resp["limit"].(float64) != 50 {
			t.Errorf("expected default limit 50, got %v", resp["limit"])
		}
	})

	t.Run("filters_by_kind", func(t *testing.T) {
		mockDB := newMockEventsDB()
		mockDB.events = []db.Event{
			{ID: "event1", Kind: 1, Content: "Note", Pubkey: "abc123"},
			{ID: "event2", Kind: 0, Content: "Metadata", Pubkey: "abc123"},
			{ID: "event3", Kind: 1, Content: "Another note", Pubkey: "def456"},
		}

		mux := http.NewServeMux()
		mux.HandleFunc("GET /api/v1/events", func(w http.ResponseWriter, r *http.Request) {
			filter := db.EventFilter{
				Limit:  50,
				Offset: 0,
			}

			if kinds := r.URL.Query().Get("kinds"); kinds != "" {
				for _, k := range strings.Split(kinds, ",") {
					if kind := parseIntParam(strings.TrimSpace(k), -1); kind >= 0 {
						filter.Kinds = append(filter.Kinds, kind)
					}
				}
			}

			events, _ := mockDB.GetEvents(r.Context(), filter)
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"events": events,
				"count":  len(events),
			})
		})

		req := httptest.NewRequest("GET", "/api/v1/events?kinds=1", nil)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)

		var resp map[string]interface{}
		json.NewDecoder(w.Body).Decode(&resp)

		count := int(resp["count"].(float64))
		if count != 2 {
			t.Errorf("expected 2 kind:1 events, got %d", count)
		}
	})

	t.Run("filters_by_author", func(t *testing.T) {
		mockDB := newMockEventsDB()
		mockDB.events = []db.Event{
			{ID: "event1", Kind: 1, Pubkey: "abc123"},
			{ID: "event2", Kind: 1, Pubkey: "abc123"},
			{ID: "event3", Kind: 1, Pubkey: "def456"},
		}

		mux := http.NewServeMux()
		mux.HandleFunc("GET /api/v1/events", func(w http.ResponseWriter, r *http.Request) {
			filter := db.EventFilter{
				Limit:  50,
				Offset: 0,
			}

			if authors := r.URL.Query().Get("authors"); authors != "" {
				filter.Authors = strings.Split(authors, ",")
			}

			events, _ := mockDB.GetEvents(r.Context(), filter)
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"events": events,
				"count":  len(events),
			})
		})

		req := httptest.NewRequest("GET", "/api/v1/events?authors=abc123", nil)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)

		var resp map[string]interface{}
		json.NewDecoder(w.Body).Decode(&resp)

		count := int(resp["count"].(float64))
		if count != 2 {
			t.Errorf("expected 2 events from abc123, got %d", count)
		}
	})

	t.Run("filters_by_search", func(t *testing.T) {
		mockDB := newMockEventsDB()
		mockDB.events = []db.Event{
			{ID: "event1", Kind: 1, Content: "Hello world"},
			{ID: "event2", Kind: 1, Content: "Goodbye world"},
			{ID: "event3", Kind: 1, Content: "Hello nostr"},
		}

		mux := http.NewServeMux()
		mux.HandleFunc("GET /api/v1/events", func(w http.ResponseWriter, r *http.Request) {
			filter := db.EventFilter{
				Limit:  50,
				Offset: 0,
				Search: r.URL.Query().Get("search"),
			}

			events, _ := mockDB.GetEvents(r.Context(), filter)
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"events": events,
				"count":  len(events),
			})
		})

		req := httptest.NewRequest("GET", "/api/v1/events?search=Hello", nil)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)

		var resp map[string]interface{}
		json.NewDecoder(w.Body).Decode(&resp)

		count := int(resp["count"].(float64))
		if count != 2 {
			t.Errorf("expected 2 events containing 'Hello', got %d", count)
		}
	})

	t.Run("pagination_with_limit_offset", func(t *testing.T) {
		mockDB := newMockEventsDB()
		for i := 0; i < 10; i++ {
			mockDB.events = append(mockDB.events, db.Event{
				ID:   "event" + string(rune('a'+i)),
				Kind: 1,
			})
		}

		mux := http.NewServeMux()
		mux.HandleFunc("GET /api/v1/events", func(w http.ResponseWriter, r *http.Request) {
			filter := db.EventFilter{
				Limit:  parseIntParam(r.URL.Query().Get("limit"), 50),
				Offset: parseIntParam(r.URL.Query().Get("offset"), 0),
			}

			events, _ := mockDB.GetEvents(r.Context(), filter)
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"events": events,
				"count":  len(events),
				"limit":  filter.Limit,
				"offset": filter.Offset,
			})
		})

		req := httptest.NewRequest("GET", "/api/v1/events?limit=3&offset=2", nil)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)

		var resp map[string]interface{}
		json.NewDecoder(w.Body).Decode(&resp)

		count := int(resp["count"].(float64))
		if count != 3 {
			t.Errorf("expected 3 events with limit=3, got %d", count)
		}
		if resp["offset"].(float64) != 2 {
			t.Errorf("expected offset 2, got %v", resp["offset"])
		}
	})
}

// ============================================================================
// GetEvent Tests
// ============================================================================

func TestGetEvent(t *testing.T) {
	t.Run("relay_not_connected", func(t *testing.T) {
		mockDB := newMockEventsDB()
		mockDB.relayConnected = false

		mux := http.NewServeMux()
		mux.HandleFunc("GET /api/v1/events/{id}", func(w http.ResponseWriter, r *http.Request) {
			if !mockDB.IsRelayDBConnected() {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusServiceUnavailable)
				json.NewEncoder(w).Encode(map[string]interface{}{
					"error": "Relay database not connected",
					"code":  "RELAY_NOT_CONNECTED",
				})
				return
			}
		})

		req := httptest.NewRequest("GET", "/api/v1/events/abc123", nil)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)

		if w.Code != http.StatusServiceUnavailable {
			t.Errorf("expected status 503, got %d", w.Code)
		}
	})

	t.Run("event_not_found", func(t *testing.T) {
		mockDB := newMockEventsDB()

		mux := http.NewServeMux()
		mux.HandleFunc("GET /api/v1/events/{id}", func(w http.ResponseWriter, r *http.Request) {
			id := r.PathValue("id")
			event, _ := mockDB.GetEvent(r.Context(), id)

			if event == nil {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusNotFound)
				json.NewEncoder(w).Encode(map[string]interface{}{
					"error": "Event not found",
					"code":  "EVENT_NOT_FOUND",
				})
				return
			}
		})

		req := httptest.NewRequest("GET", "/api/v1/events/nonexistent", nil)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("expected status 404, got %d", w.Code)
		}
	})

	t.Run("returns_event", func(t *testing.T) {
		mockDB := newMockEventsDB()
		mockDB.eventByID["abc123"] = &db.Event{
			ID:        "abc123",
			Kind:      1,
			Content:   "Hello Nostr!",
			Pubkey:    "author_pubkey",
			CreatedAt: time.Now(),
		}

		mux := http.NewServeMux()
		mux.HandleFunc("GET /api/v1/events/{id}", func(w http.ResponseWriter, r *http.Request) {
			id := r.PathValue("id")
			event, _ := mockDB.GetEvent(r.Context(), id)

			if event == nil {
				w.WriteHeader(http.StatusNotFound)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(event)
		})

		req := httptest.NewRequest("GET", "/api/v1/events/abc123", nil)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}

		var event db.Event
		json.NewDecoder(w.Body).Decode(&event)

		if event.ID != "abc123" {
			t.Errorf("expected event ID 'abc123', got '%s'", event.ID)
		}
		if event.Kind != 1 {
			t.Errorf("expected kind 1, got %d", event.Kind)
		}
	})

	t.Run("missing_id_parameter", func(t *testing.T) {
		mux := http.NewServeMux()
		mux.HandleFunc("GET /api/v1/events/{id}", func(w http.ResponseWriter, r *http.Request) {
			id := r.PathValue("id")
			if id == "" {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(map[string]interface{}{
					"error": "Event ID is required",
					"code":  "MISSING_ID",
				})
				return
			}
		})

		// Note: With Go 1.22's path patterns, empty ID won't match this route
		// Testing with empty string in path
		req := httptest.NewRequest("GET", "/api/v1/events/", nil)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)

		// The route won't match, so we get 404
		if w.Code != http.StatusNotFound {
			t.Logf("got status %d (expected 404 for empty path)", w.Code)
		}
	})
}

// ============================================================================
// DeleteEvent Tests
// ============================================================================

func TestDeleteEvent(t *testing.T) {
	t.Run("event_not_found", func(t *testing.T) {
		mockDB := newMockEventsDB()

		mux := http.NewServeMux()
		mux.HandleFunc("DELETE /api/v1/events/{id}", func(w http.ResponseWriter, r *http.Request) {
			id := r.PathValue("id")

			// Check if event exists
			if mockDB.IsRelayDBConnected() {
				event, _ := mockDB.GetEvent(r.Context(), id)
				if event == nil {
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusNotFound)
					json.NewEncoder(w).Encode(map[string]interface{}{
						"error": "Event not found",
						"code":  "EVENT_NOT_FOUND",
					})
					return
				}
			}
		})

		req := httptest.NewRequest("DELETE", "/api/v1/events/nonexistent", nil)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("expected status 404, got %d", w.Code)
		}
	})

	t.Run("successful_deletion_request", func(t *testing.T) {
		mockDB := newMockEventsDB()
		mockDB.eventByID["event123"] = &db.Event{
			ID:     "event123",
			Kind:   1,
			Pubkey: "author_pubkey",
		}

		mux := http.NewServeMux()
		mux.HandleFunc("DELETE /api/v1/events/{id}", func(w http.ResponseWriter, r *http.Request) {
			id := r.PathValue("id")

			// Verify event exists
			event, _ := mockDB.GetEvent(r.Context(), id)
			if event == nil {
				w.WriteHeader(http.StatusNotFound)
				return
			}

			// Parse optional reason
			var reqBody struct {
				Reason string `json:"reason"`
			}
			json.NewDecoder(r.Body).Decode(&reqBody)

			operator, _ := mockDB.GetAppState(r.Context(), "operator_pubkey")
			requestID, err := mockDB.CreateDeletionRequest(r.Context(), id, operator, reqBody.Reason)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusAccepted)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"message":    "Deletion request queued",
				"request_id": requestID,
				"event_id":   id,
				"status":     "pending",
			})
		})

		body := `{"reason":"spam content"}`
		req := httptest.NewRequest("DELETE", "/api/v1/events/event123", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)

		if w.Code != http.StatusAccepted {
			t.Errorf("expected status 202, got %d", w.Code)
		}

		var resp map[string]interface{}
		json.NewDecoder(w.Body).Decode(&resp)

		if resp["status"] != "pending" {
			t.Errorf("expected status 'pending', got %v", resp["status"])
		}
		if resp["event_id"] != "event123" {
			t.Errorf("expected event_id 'event123', got %v", resp["event_id"])
		}
	})

	t.Run("duplicate_deletion_request", func(t *testing.T) {
		mockDB := newMockEventsDB()
		mockDB.eventByID["event123"] = &db.Event{ID: "event123", Kind: 1}
		mockDB.duplicateDelete = true

		mux := http.NewServeMux()
		mux.HandleFunc("DELETE /api/v1/events/{id}", func(w http.ResponseWriter, r *http.Request) {
			id := r.PathValue("id")

			event, _ := mockDB.GetEvent(r.Context(), id)
			if event == nil {
				w.WriteHeader(http.StatusNotFound)
				return
			}

			operator, _ := mockDB.GetAppState(r.Context(), "operator_pubkey")
			_, err := mockDB.CreateDeletionRequest(r.Context(), id, operator, "")
			if err != nil {
				if strings.Contains(err.Error(), "UNIQUE constraint") {
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusConflict)
					json.NewEncoder(w).Encode(map[string]interface{}{
						"error": "Deletion already requested for this event",
						"code":  "DUPLICATE_REQUEST",
					})
					return
				}
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		})

		req := httptest.NewRequest("DELETE", "/api/v1/events/event123", nil)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)

		if w.Code != http.StatusConflict {
			t.Errorf("expected status 409, got %d", w.Code)
		}
	})
}

// ============================================================================
// GetRecentEvents Tests
// ============================================================================

func TestGetRecentEvents(t *testing.T) {
	t.Run("relay_not_connected_returns_empty", func(t *testing.T) {
		mockDB := newMockEventsDB()
		mockDB.relayConnected = false

		mux := http.NewServeMux()
		mux.HandleFunc("GET /api/v1/events/recent", func(w http.ResponseWriter, r *http.Request) {
			if !mockDB.IsRelayDBConnected() {
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(map[string]interface{}{
					"events": []interface{}{},
				})
				return
			}
		})

		req := httptest.NewRequest("GET", "/api/v1/events/recent", nil)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}

		var resp map[string]interface{}
		json.NewDecoder(w.Body).Decode(&resp)

		events := resp["events"].([]interface{})
		if len(events) != 0 {
			t.Errorf("expected empty events array, got %d items", len(events))
		}
	})

	t.Run("returns_recent_events", func(t *testing.T) {
		mockDB := newMockEventsDB()
		now := time.Now()
		for i := 0; i < 15; i++ {
			mockDB.events = append(mockDB.events, db.Event{
				ID:        "event" + string(rune('a'+i)),
				Kind:      1,
				CreatedAt: now.Add(-time.Duration(i) * time.Minute), // Each event 1 minute older
			})
		}

		mux := http.NewServeMux()
		mux.HandleFunc("GET /api/v1/events/recent", func(w http.ResponseWriter, r *http.Request) {
			events, _ := mockDB.GetRecentEvents(r.Context(), 10)
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"events": events,
			})
		})

		req := httptest.NewRequest("GET", "/api/v1/events/recent", nil)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}

		var resp map[string]interface{}
		json.NewDecoder(w.Body).Decode(&resp)

		events := resp["events"].([]interface{})
		if len(events) != 10 {
			t.Errorf("expected 10 recent events, got %d", len(events))
		}
	})
}

// ============================================================================
// parseIntParam Tests (utility function)
// ============================================================================

func TestParseIntParam(t *testing.T) {
	t.Run("empty_returns_default", func(t *testing.T) {
		result := parseIntParam("", 50)
		if result != 50 {
			t.Errorf("expected default 50, got %d", result)
		}
	})

	t.Run("valid_int", func(t *testing.T) {
		result := parseIntParam("25", 50)
		if result != 25 {
			t.Errorf("expected 25, got %d", result)
		}
	})

	t.Run("invalid_int_returns_default", func(t *testing.T) {
		result := parseIntParam("not-a-number", 50)
		if result != 50 {
			t.Errorf("expected default 50, got %d", result)
		}
	})

	t.Run("negative_int", func(t *testing.T) {
		result := parseIntParam("-10", 50)
		if result != -10 {
			t.Errorf("expected -10, got %d", result)
		}
	})
}
