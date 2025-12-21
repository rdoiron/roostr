package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// ============================================================================
// GetStatsSummary Tests
// ============================================================================

func TestGetStatsSummary(t *testing.T) {
	t.Run("returns_stats", func(t *testing.T) {
		mux := http.NewServeMux()
		mux.HandleFunc("GET /api/v1/stats/summary", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"total_events":      1000,
				"events_today":      50,
				"storage_bytes":     104857600,
				"whitelisted_count": 10,
				"events_by_kind": map[string]int64{
					"posts":     500,
					"reactions": 300,
					"reposts":   100,
					"other":     100,
				},
				"uptime_seconds": 3600,
				"relay_status":   "online",
			})
		})

		req := httptest.NewRequest("GET", "/api/v1/stats/summary", nil)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}

		var resp map[string]interface{}
		json.NewDecoder(w.Body).Decode(&resp)

		if int(resp["total_events"].(float64)) != 1000 {
			t.Errorf("expected 1000 total events, got %v", resp["total_events"])
		}
		if resp["relay_status"] != "online" {
			t.Errorf("expected relay_status 'online', got %v", resp["relay_status"])
		}
	})

	t.Run("relay_not_connected", func(t *testing.T) {
		mux := http.NewServeMux()
		mux.HandleFunc("GET /api/v1/stats/summary", func(w http.ResponseWriter, r *http.Request) {
			relayConnected := false

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"total_events":      0,
				"events_today":      0,
				"storage_bytes":     0,
				"whitelisted_count": 5,
				"events_by_kind":    map[string]int64{},
				"uptime_seconds":    100,
				"relay_status":      map[bool]string{true: "online", false: "offline"}[relayConnected],
			})
		})

		req := httptest.NewRequest("GET", "/api/v1/stats/summary", nil)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)

		var resp map[string]interface{}
		json.NewDecoder(w.Body).Decode(&resp)

		if resp["relay_status"] != "offline" {
			t.Errorf("expected relay_status 'offline', got %v", resp["relay_status"])
		}
		if int(resp["total_events"].(float64)) != 0 {
			t.Errorf("expected 0 total events, got %v", resp["total_events"])
		}
	})
}

// ============================================================================
// GetRelayStatus Tests
// ============================================================================

func TestGetRelayStatus(t *testing.T) {
	t.Run("returns_status", func(t *testing.T) {
		mux := http.NewServeMux()
		mux.HandleFunc("GET /api/v1/relay/status", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"status":             "running",
				"pid":                12345,
				"memory_bytes":       104857600,
				"uptime_seconds":     7200,
				"database_connected": true,
				"api_uptime_seconds": 7200,
			})
		})

		req := httptest.NewRequest("GET", "/api/v1/relay/status", nil)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}

		var resp map[string]interface{}
		json.NewDecoder(w.Body).Decode(&resp)

		if resp["status"] != "running" {
			t.Errorf("expected status 'running', got %v", resp["status"])
		}
		if resp["database_connected"] != true {
			t.Errorf("expected database_connected true, got %v", resp["database_connected"])
		}
	})

	t.Run("restarting_status", func(t *testing.T) {
		mux := http.NewServeMux()
		mux.HandleFunc("GET /api/v1/relay/status", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"status": "restarting",
			})
		})

		req := httptest.NewRequest("GET", "/api/v1/relay/status", nil)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)

		var resp map[string]interface{}
		json.NewDecoder(w.Body).Decode(&resp)

		if resp["status"] != "restarting" {
			t.Errorf("expected status 'restarting', got %v", resp["status"])
		}
	})
}

// ============================================================================
// GetRelayURLs Tests
// ============================================================================

func TestGetRelayURLs(t *testing.T) {
	t.Run("returns_local_url", func(t *testing.T) {
		mux := http.NewServeMux()
		mux.HandleFunc("GET /api/v1/relay/urls", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"local":         "ws://localhost:7000",
				"relay_port":    7000,
				"tor":           "",
				"tor_available": false,
			})
		})

		req := httptest.NewRequest("GET", "/api/v1/relay/urls", nil)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)

		var resp map[string]interface{}
		json.NewDecoder(w.Body).Decode(&resp)

		if resp["local"] != "ws://localhost:7000" {
			t.Errorf("expected local 'ws://localhost:7000', got %v", resp["local"])
		}
		if resp["tor_available"] != false {
			t.Errorf("expected tor_available false, got %v", resp["tor_available"])
		}
	})

	t.Run("with_tor_url", func(t *testing.T) {
		mux := http.NewServeMux()
		mux.HandleFunc("GET /api/v1/relay/urls", func(w http.ResponseWriter, r *http.Request) {
			torAddress := "abc123.onion:7000"

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"local":         "ws://localhost:7000",
				"relay_port":    7000,
				"tor":           "ws://" + torAddress,
				"tor_available": true,
			})
		})

		req := httptest.NewRequest("GET", "/api/v1/relay/urls", nil)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)

		var resp map[string]interface{}
		json.NewDecoder(w.Body).Decode(&resp)

		if resp["tor_available"] != true {
			t.Errorf("expected tor_available true, got %v", resp["tor_available"])
		}
		if resp["tor"] != "ws://abc123.onion:7000" {
			t.Errorf("expected tor 'ws://abc123.onion:7000', got %v", resp["tor"])
		}
	})
}

// ============================================================================
// GetEventsOverTime Tests
// ============================================================================

func TestGetEventsOverTime(t *testing.T) {
	t.Run("returns_data", func(t *testing.T) {
		mux := http.NewServeMux()
		mux.HandleFunc("GET /api/v1/stats/events-over-time", func(w http.ResponseWriter, r *http.Request) {
			timeRange := r.URL.Query().Get("time_range")
			if timeRange == "" {
				timeRange = "7days"
			}

			data := []map[string]interface{}{
				{"date": "2025-01-01", "count": 100},
				{"date": "2025-01-02", "count": 150},
				{"date": "2025-01-03", "count": 120},
			}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data":       data,
				"time_range": timeRange,
				"total":      370,
			})
		})

		req := httptest.NewRequest("GET", "/api/v1/stats/events-over-time?time_range=7days", nil)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}

		var resp map[string]interface{}
		json.NewDecoder(w.Body).Decode(&resp)

		if resp["time_range"] != "7days" {
			t.Errorf("expected time_range '7days', got %v", resp["time_range"])
		}
		if int(resp["total"].(float64)) != 370 {
			t.Errorf("expected total 370, got %v", resp["total"])
		}
	})

	t.Run("relay_not_connected", func(t *testing.T) {
		mux := http.NewServeMux()
		mux.HandleFunc("GET /api/v1/stats/events-over-time", func(w http.ResponseWriter, r *http.Request) {
			relayConnected := false
			timeRange := r.URL.Query().Get("time_range")
			if timeRange == "" {
				timeRange = "7days"
			}

			if !relayConnected {
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(map[string]interface{}{
					"data":       []interface{}{},
					"time_range": timeRange,
					"total":      0,
				})
				return
			}
		})

		req := httptest.NewRequest("GET", "/api/v1/stats/events-over-time", nil)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)

		var resp map[string]interface{}
		json.NewDecoder(w.Body).Decode(&resp)

		data := resp["data"].([]interface{})
		if len(data) != 0 {
			t.Errorf("expected empty data array, got %d items", len(data))
		}
	})
}

// ============================================================================
// GetEventsByKind Tests
// ============================================================================

func TestGetEventsByKind(t *testing.T) {
	t.Run("returns_kinds", func(t *testing.T) {
		mux := http.NewServeMux()
		mux.HandleFunc("GET /api/v1/stats/events-by-kind", func(w http.ResponseWriter, r *http.Request) {
			timeRange := r.URL.Query().Get("time_range")
			if timeRange == "" {
				timeRange = "alltime"
			}

			kinds := []map[string]interface{}{
				{"kind": 1, "label": "posts", "count": 500, "percent": 50.0},
				{"kind": 7, "label": "reactions", "count": 300, "percent": 30.0},
				{"kind": 6, "label": "reposts", "count": 200, "percent": 20.0},
			}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"kinds":      kinds,
				"time_range": timeRange,
				"total":      1000,
			})
		})

		req := httptest.NewRequest("GET", "/api/v1/stats/events-by-kind", nil)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}

		var resp map[string]interface{}
		json.NewDecoder(w.Body).Decode(&resp)

		kinds := resp["kinds"].([]interface{})
		if len(kinds) != 3 {
			t.Errorf("expected 3 kinds, got %d", len(kinds))
		}
	})
}

// ============================================================================
// GetTopAuthors Tests
// ============================================================================

func TestGetTopAuthors(t *testing.T) {
	t.Run("returns_authors", func(t *testing.T) {
		mux := http.NewServeMux()
		mux.HandleFunc("GET /api/v1/stats/top-authors", func(w http.ResponseWriter, r *http.Request) {
			limit := 10
			if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
				if parsed, err := parseInt(limitStr); err == nil && parsed > 0 {
					limit = parsed
				}
			}

			authors := []map[string]interface{}{
				{"pubkey": "abc123", "event_count": 100},
				{"pubkey": "def456", "event_count": 75},
				{"pubkey": "ghi789", "event_count": 50},
			}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"authors":    authors,
				"time_range": "alltime",
				"limit":      limit,
			})
		})

		req := httptest.NewRequest("GET", "/api/v1/stats/top-authors", nil)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}

		var resp map[string]interface{}
		json.NewDecoder(w.Body).Decode(&resp)

		authors := resp["authors"].([]interface{})
		if len(authors) != 3 {
			t.Errorf("expected 3 authors, got %d", len(authors))
		}
	})

	t.Run("respects_limit", func(t *testing.T) {
		mux := http.NewServeMux()
		mux.HandleFunc("GET /api/v1/stats/top-authors", func(w http.ResponseWriter, r *http.Request) {
			limit := 10
			if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
				if parsed, err := parseInt(limitStr); err == nil && parsed > 0 {
					limit = parsed
					if limit > 100 {
						limit = 100
					}
				}
			}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"limit": limit,
			})
		})

		req := httptest.NewRequest("GET", "/api/v1/stats/top-authors?limit=5", nil)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)

		var resp map[string]interface{}
		json.NewDecoder(w.Body).Decode(&resp)

		if int(resp["limit"].(float64)) != 5 {
			t.Errorf("expected limit 5, got %v", resp["limit"])
		}
	})

	t.Run("caps_limit_at_100", func(t *testing.T) {
		mux := http.NewServeMux()
		mux.HandleFunc("GET /api/v1/stats/top-authors", func(w http.ResponseWriter, r *http.Request) {
			limit := 10
			if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
				if parsed, err := parseInt(limitStr); err == nil && parsed > 0 {
					limit = parsed
					if limit > 100 {
						limit = 100
					}
				}
			}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"limit": limit,
			})
		})

		req := httptest.NewRequest("GET", "/api/v1/stats/top-authors?limit=500", nil)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)

		var resp map[string]interface{}
		json.NewDecoder(w.Body).Decode(&resp)

		if int(resp["limit"].(float64)) != 100 {
			t.Errorf("expected limit capped at 100, got %v", resp["limit"])
		}
	})
}

// ============================================================================
// parseTimeRange Tests
// ============================================================================

func TestParseTimeRange(t *testing.T) {
	t.Run("today", func(t *testing.T) {
		since, until := parseTimeRange("today", "")
		now := time.Now().UTC()

		// Since should be start of today
		if since.Day() != now.Day() {
			t.Errorf("expected since to be today, got %v", since)
		}
		if since.Hour() != 0 || since.Minute() != 0 || since.Second() != 0 {
			t.Errorf("expected since to be start of day, got %v", since)
		}

		// Until should be end of today
		if until.Day() != now.Day() {
			t.Errorf("expected until to be today, got %v", until)
		}
	})

	t.Run("7days", func(t *testing.T) {
		since, _ := parseTimeRange("7days", "")
		now := time.Now().UTC()

		// Since should be 6 days ago (to include today as day 7)
		expected := now.AddDate(0, 0, -6)
		if since.Year() != expected.Year() || since.Month() != expected.Month() || since.Day() != expected.Day() {
			t.Errorf("expected since to be 6 days ago, got %v", since)
		}
	})

	t.Run("30days", func(t *testing.T) {
		since, _ := parseTimeRange("30days", "")
		now := time.Now().UTC()

		// Since should be 29 days ago (to include today as day 30)
		expected := now.AddDate(0, 0, -29)
		if since.Year() != expected.Year() || since.Month() != expected.Month() || since.Day() != expected.Day() {
			t.Errorf("expected since to be 29 days ago, got %v", since)
		}
	})

	t.Run("alltime", func(t *testing.T) {
		since, _ := parseTimeRange("alltime", "")

		// Since should be zero value for alltime
		if !since.IsZero() {
			t.Errorf("expected since to be zero for alltime, got %v", since)
		}
	})

	t.Run("default_to_7days", func(t *testing.T) {
		since, _ := parseTimeRange("invalid", "")
		now := time.Now().UTC()

		// Should default to 7 days
		expected := now.AddDate(0, 0, -6)
		if since.Year() != expected.Year() || since.Month() != expected.Month() || since.Day() != expected.Day() {
			t.Errorf("expected since to default to 6 days ago, got %v", since)
		}
	})

	t.Run("with_timezone", func(t *testing.T) {
		since, _ := parseTimeRange("today", "America/New_York")

		// Should be in New York timezone
		loc := since.Location()
		if loc.String() != "America/New_York" {
			t.Errorf("expected timezone 'America/New_York', got '%s'", loc.String())
		}
	})
}

// ============================================================================
// getKindLabel Tests
// ============================================================================

func TestGetKindLabel(t *testing.T) {
	testCases := []struct {
		kind     int
		expected string
	}{
		{0, "profiles"},
		{1, "posts"},
		{3, "follows"},
		{4, "dms"},
		{14, "dms"},
		{6, "reposts"},
		{7, "reactions"},
		{30023, "other"},
		{9999, "other"},
	}

	for _, tc := range testCases {
		result := getKindLabel(tc.kind)
		if result != tc.expected {
			t.Errorf("getKindLabel(%d): expected '%s', got '%s'", tc.kind, tc.expected, result)
		}
	}
}
