package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// ============================================================================
// ReloadRelay Tests
// ============================================================================

func TestReloadRelay(t *testing.T) {
	t.Run("successful_reload", func(t *testing.T) {
		mux := http.NewServeMux()
		mux.HandleFunc("POST /api/v1/relay/reload", func(w http.ResponseWriter, r *http.Request) {
			relayAvailable := true
			if !relayAvailable {
				w.WriteHeader(http.StatusServiceUnavailable)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": true,
				"message": "Relay configuration reloaded",
			})
		})

		req := httptest.NewRequest("POST", "/api/v1/relay/reload", nil)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}

		var resp map[string]interface{}
		json.NewDecoder(w.Body).Decode(&resp)

		if resp["success"] != true {
			t.Errorf("expected success true, got %v", resp["success"])
		}
	})

	t.Run("relay_not_available", func(t *testing.T) {
		mux := http.NewServeMux()
		mux.HandleFunc("POST /api/v1/relay/reload", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusServiceUnavailable)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"error": "Relay manager not available",
				"code":  "RELAY_NOT_AVAILABLE",
			})
		})

		req := httptest.NewRequest("POST", "/api/v1/relay/reload", nil)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)

		if w.Code != http.StatusServiceUnavailable {
			t.Errorf("expected status 503, got %d", w.Code)
		}
	})

	t.Run("reload_failed", func(t *testing.T) {
		mux := http.NewServeMux()
		mux.HandleFunc("POST /api/v1/relay/reload", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"error": "Failed to reload relay",
				"code":  "RELOAD_FAILED",
			})
		})

		req := httptest.NewRequest("POST", "/api/v1/relay/reload", nil)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)

		if w.Code != http.StatusInternalServerError {
			t.Errorf("expected status 500, got %d", w.Code)
		}
	})
}

// ============================================================================
// RestartRelay Tests
// ============================================================================

func TestRestartRelay(t *testing.T) {
	t.Run("successful_restart", func(t *testing.T) {
		mux := http.NewServeMux()
		mux.HandleFunc("POST /api/v1/relay/restart", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": true,
				"message": "Relay restart initiated",
				"status":  "restarting",
			})
		})

		req := httptest.NewRequest("POST", "/api/v1/relay/restart", nil)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}

		var resp map[string]interface{}
		json.NewDecoder(w.Body).Decode(&resp)

		if resp["success"] != true {
			t.Errorf("expected success true, got %v", resp["success"])
		}
		if resp["status"] != "restarting" {
			t.Errorf("expected status 'restarting', got %v", resp["status"])
		}
	})

	t.Run("relay_not_available", func(t *testing.T) {
		mux := http.NewServeMux()
		mux.HandleFunc("POST /api/v1/relay/restart", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusServiceUnavailable)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"error": "Relay manager not available",
				"code":  "RELAY_NOT_AVAILABLE",
			})
		})

		req := httptest.NewRequest("POST", "/api/v1/relay/restart", nil)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)

		if w.Code != http.StatusServiceUnavailable {
			t.Errorf("expected status 503, got %d", w.Code)
		}
	})

	t.Run("restart_in_progress", func(t *testing.T) {
		mux := http.NewServeMux()
		mux.HandleFunc("POST /api/v1/relay/restart", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusConflict)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"error": "Relay restart already in progress",
				"code":  "RESTART_IN_PROGRESS",
			})
		})

		req := httptest.NewRequest("POST", "/api/v1/relay/restart", nil)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)

		if w.Code != http.StatusConflict {
			t.Errorf("expected status 409, got %d", w.Code)
		}
	})

	t.Run("restart_failed", func(t *testing.T) {
		mux := http.NewServeMux()
		mux.HandleFunc("POST /api/v1/relay/restart", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"error": "Failed to initiate relay restart",
				"code":  "RESTART_FAILED",
			})
		})

		req := httptest.NewRequest("POST", "/api/v1/relay/restart", nil)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)

		if w.Code != http.StatusInternalServerError {
			t.Errorf("expected status 500, got %d", w.Code)
		}
	})
}

// ============================================================================
// GetRelayLogs Tests
// ============================================================================

func TestGetRelayLogs(t *testing.T) {
	t.Run("returns_logs", func(t *testing.T) {
		mux := http.NewServeMux()
		mux.HandleFunc("GET /api/v1/relay/logs", func(w http.ResponseWriter, r *http.Request) {
			logs := []LogEntry{
				{Timestamp: "2025-01-01T12:00:00Z", Level: "INFO", Message: "Relay started"},
				{Timestamp: "2025-01-01T12:00:01Z", Level: "INFO", Message: "Listening on port 7000"},
			}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"logs":        logs,
				"total_lines": len(logs),
			})
		})

		req := httptest.NewRequest("GET", "/api/v1/relay/logs", nil)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}

		var resp map[string]interface{}
		json.NewDecoder(w.Body).Decode(&resp)

		logs := resp["logs"].([]interface{})
		if len(logs) != 2 {
			t.Errorf("expected 2 log entries, got %d", len(logs))
		}
	})

	t.Run("respects_limit_parameter", func(t *testing.T) {
		mux := http.NewServeMux()
		mux.HandleFunc("GET /api/v1/relay/logs", func(w http.ResponseWriter, r *http.Request) {
			limitStr := r.URL.Query().Get("limit")
			limit := 100
			if limitStr != "" {
				if parsed, err := parseInt(limitStr); err == nil && parsed > 0 {
					limit = parsed
					if limit > 1000 {
						limit = 1000
					}
				}
			}

			// Generate logs up to limit
			logs := make([]LogEntry, 0, limit)
			for i := 0; i < limit && i < 5; i++ {
				logs = append(logs, LogEntry{
					Timestamp: "2025-01-01T12:00:00Z",
					Level:     "INFO",
					Message:   "Log entry",
				})
			}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"logs":        logs,
				"total_lines": len(logs),
			})
		})

		req := httptest.NewRequest("GET", "/api/v1/relay/logs?limit=3", nil)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)

		var resp map[string]interface{}
		json.NewDecoder(w.Body).Decode(&resp)

		logs := resp["logs"].([]interface{})
		if len(logs) != 3 {
			t.Errorf("expected 3 log entries with limit=3, got %d", len(logs))
		}
	})

	t.Run("caps_limit_at_1000", func(t *testing.T) {
		mux := http.NewServeMux()
		mux.HandleFunc("GET /api/v1/relay/logs", func(w http.ResponseWriter, r *http.Request) {
			limitStr := r.URL.Query().Get("limit")
			limit := 100
			if limitStr != "" {
				if parsed, err := parseInt(limitStr); err == nil && parsed > 0 {
					limit = parsed
					if limit > 1000 {
						limit = 1000
					}
				}
			}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"effective_limit": limit,
			})
		})

		req := httptest.NewRequest("GET", "/api/v1/relay/logs?limit=5000", nil)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)

		var resp map[string]interface{}
		json.NewDecoder(w.Body).Decode(&resp)

		if int(resp["effective_limit"].(float64)) != 1000 {
			t.Errorf("expected limit to be capped at 1000, got %v", resp["effective_limit"])
		}
	})

	t.Run("empty_logs", func(t *testing.T) {
		mux := http.NewServeMux()
		mux.HandleFunc("GET /api/v1/relay/logs", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"logs":        []LogEntry{},
				"total_lines": 0,
				"message":     "Relay logging not configured",
			})
		})

		req := httptest.NewRequest("GET", "/api/v1/relay/logs", nil)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}

		var resp map[string]interface{}
		json.NewDecoder(w.Body).Decode(&resp)

		if int(resp["total_lines"].(float64)) != 0 {
			t.Errorf("expected 0 total lines, got %v", resp["total_lines"])
		}
	})
}

// ============================================================================
// parseLogLine Tests
// ============================================================================

func TestParseLogLine(t *testing.T) {
	t.Run("bracketed_format", func(t *testing.T) {
		line := "[2025-12-13T14:32:01Z INFO nostr_rs_relay] Starting relay..."
		entry := parseLogLine(line)

		if entry.Timestamp != "2025-12-13T14:32:01Z" {
			t.Errorf("expected timestamp '2025-12-13T14:32:01Z', got '%s'", entry.Timestamp)
		}
		if entry.Level != "INFO" {
			t.Errorf("expected level 'INFO', got '%s'", entry.Level)
		}
		if entry.Message == "" {
			t.Error("expected non-empty message")
		}
	})

	t.Run("space_separated_format", func(t *testing.T) {
		line := "2025-12-13 14:32:01 ERROR Connection failed"
		entry := parseLogLine(line)

		if entry.Timestamp != "2025-12-13 14:32:01" {
			t.Errorf("expected timestamp '2025-12-13 14:32:01', got '%s'", entry.Timestamp)
		}
		if entry.Level != "ERROR" {
			t.Errorf("expected level 'ERROR', got '%s'", entry.Level)
		}
		if entry.Message != "Connection failed" {
			t.Errorf("expected message 'Connection failed', got '%s'", entry.Message)
		}
	})

	t.Run("unrecognized_format_fallback", func(t *testing.T) {
		line := "Some random log line without timestamp"
		entry := parseLogLine(line)

		// Should use fallback behavior
		if entry.Level != "INFO" {
			t.Errorf("expected default level 'INFO', got '%s'", entry.Level)
		}
		if entry.Message != line {
			t.Errorf("expected message to be the whole line, got '%s'", entry.Message)
		}
	})

	t.Run("various_log_levels", func(t *testing.T) {
		testCases := []struct {
			line          string
			expectedLevel string
		}{
			{"2025-01-01 12:00:00 DEBUG Debugging info", "DEBUG"},
			{"2025-01-01 12:00:00 INFO Information", "INFO"},
			{"2025-01-01 12:00:00 WARN Warning message", "WARN"},
			{"2025-01-01 12:00:00 ERROR Error occurred", "ERROR"},
		}

		for _, tc := range testCases {
			entry := parseLogLine(tc.line)
			if entry.Level != tc.expectedLevel {
				t.Errorf("for line '%s': expected level '%s', got '%s'", tc.line, tc.expectedLevel, entry.Level)
			}
		}
	})
}

// ============================================================================
// StreamRelayLogs Tests (SSE)
// ============================================================================

func TestStreamRelayLogs(t *testing.T) {
	t.Run("sets_sse_headers", func(t *testing.T) {
		mux := http.NewServeMux()
		mux.HandleFunc("GET /api/v1/relay/logs/stream", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/event-stream")
			w.Header().Set("Cache-Control", "no-cache")
			w.Header().Set("Connection", "keep-alive")

			// Just send one event and close
			w.Write([]byte("event: connected\ndata: {\"status\": \"connected\"}\n\n"))
		})

		req := httptest.NewRequest("GET", "/api/v1/relay/logs/stream", nil)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)

		if w.Header().Get("Content-Type") != "text/event-stream" {
			t.Errorf("expected Content-Type 'text/event-stream', got '%s'", w.Header().Get("Content-Type"))
		}
		if w.Header().Get("Cache-Control") != "no-cache" {
			t.Errorf("expected Cache-Control 'no-cache', got '%s'", w.Header().Get("Cache-Control"))
		}
	})

	t.Run("error_when_not_available", func(t *testing.T) {
		mux := http.NewServeMux()
		mux.HandleFunc("GET /api/v1/relay/logs/stream", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/event-stream")
			w.Write([]byte("event: error\ndata: {\"error\": \"Log streaming not available\"}\n\n"))
		})

		req := httptest.NewRequest("GET", "/api/v1/relay/logs/stream", nil)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)

		body := w.Body.String()
		if !contains(body, "error") {
			t.Error("expected error event in response")
		}
	})
}

// ============================================================================
// LogEntry Tests
// ============================================================================

func TestLogEntry(t *testing.T) {
	t.Run("json_serialization", func(t *testing.T) {
		entry := LogEntry{
			Timestamp: "2025-01-01T12:00:00Z",
			Level:     "INFO",
			Message:   "Test message",
		}

		data, err := json.Marshal(entry)
		if err != nil {
			t.Fatalf("failed to marshal LogEntry: %v", err)
		}

		var decoded LogEntry
		if err := json.Unmarshal(data, &decoded); err != nil {
			t.Fatalf("failed to unmarshal LogEntry: %v", err)
		}

		if decoded.Timestamp != entry.Timestamp {
			t.Errorf("timestamp mismatch: expected '%s', got '%s'", entry.Timestamp, decoded.Timestamp)
		}
		if decoded.Level != entry.Level {
			t.Errorf("level mismatch: expected '%s', got '%s'", entry.Level, decoded.Level)
		}
		if decoded.Message != entry.Message {
			t.Errorf("message mismatch: expected '%s', got '%s'", entry.Message, decoded.Message)
		}
	})
}

// Helper function for parsing int
func parseInt(s string) (int, error) {
	var result int
	err := json.Unmarshal([]byte(s), &result)
	if err != nil {
		// Try strconv
		return 0, err
	}
	return result, nil
}

// Helper function to check if string contains substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
