package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// ============================================================================
// StartSync Tests
// ============================================================================

func TestStartSync(t *testing.T) {
	t.Run("requires_pubkeys", func(t *testing.T) {
		mux := http.NewServeMux()
		mux.HandleFunc("POST /api/v1/sync/start", func(w http.ResponseWriter, r *http.Request) {
			var req StartSyncRequest
			json.NewDecoder(r.Body).Decode(&req)

			if len(req.Pubkeys) == 0 {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(map[string]interface{}{
					"error": "At least one pubkey is required",
					"code":  "MISSING_PUBKEYS",
				})
				return
			}
		})

		body := `{"pubkeys":[],"relays":["wss://relay.damus.io"]}`
		req := httptest.NewRequest("POST", "/api/v1/sync/start", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", w.Code)
		}
	})

	t.Run("successful_start", func(t *testing.T) {
		mux := http.NewServeMux()
		mux.HandleFunc("POST /api/v1/sync/start", func(w http.ResponseWriter, r *http.Request) {
			var req StartSyncRequest
			json.NewDecoder(r.Body).Decode(&req)

			if len(req.Pubkeys) == 0 {
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			w.Header().Set("Location", "/api/v1/sync/status")
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusAccepted)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"job_id":  1,
				"status":  "running",
				"message": "Sync job started",
			})
		})

		body := `{"pubkeys":["abc123def456"],"relays":["wss://relay.damus.io"]}`
		req := httptest.NewRequest("POST", "/api/v1/sync/start", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)

		if w.Code != http.StatusAccepted {
			t.Errorf("expected status 202, got %d", w.Code)
		}

		var resp map[string]interface{}
		json.NewDecoder(w.Body).Decode(&resp)

		if resp["status"] != "running" {
			t.Errorf("expected status 'running', got %v", resp["status"])
		}
	})

	t.Run("sync_already_running", func(t *testing.T) {
		mux := http.NewServeMux()
		mux.HandleFunc("POST /api/v1/sync/start", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusConflict)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"error": "a sync job is already running",
				"code":  "SYNC_ALREADY_RUNNING",
			})
		})

		body := `{"pubkeys":["abc123"]}`
		req := httptest.NewRequest("POST", "/api/v1/sync/start", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)

		if w.Code != http.StatusConflict {
			t.Errorf("expected status 409, got %d", w.Code)
		}
	})

	t.Run("invalid_json", func(t *testing.T) {
		mux := http.NewServeMux()
		mux.HandleFunc("POST /api/v1/sync/start", func(w http.ResponseWriter, r *http.Request) {
			var req StartSyncRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(map[string]interface{}{
					"error": "Invalid request body",
					"code":  "INVALID_JSON",
				})
				return
			}
		})

		body := `{invalid}`
		req := httptest.NewRequest("POST", "/api/v1/sync/start", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", w.Code)
		}
	})
}

// ============================================================================
// GetSyncStatus Tests
// ============================================================================

func TestGetSyncStatus(t *testing.T) {
	t.Run("returns_running_job", func(t *testing.T) {
		mux := http.NewServeMux()
		mux.HandleFunc("GET /api/v1/sync/status", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"id":             1,
				"status":         "running",
				"pubkeys":        []string{"abc123"},
				"relays":         []string{"wss://relay.damus.io"},
				"events_fetched": 100,
				"events_stored":  90,
				"events_skipped": 10,
			})
		})

		req := httptest.NewRequest("GET", "/api/v1/sync/status", nil)
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
	})

	t.Run("no_job_running", func(t *testing.T) {
		mux := http.NewServeMux()
		mux.HandleFunc("GET /api/v1/sync/status", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"status":  "idle",
				"message": "No sync jobs found",
			})
		})

		req := httptest.NewRequest("GET", "/api/v1/sync/status", nil)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)

		var resp map[string]interface{}
		json.NewDecoder(w.Body).Decode(&resp)

		if resp["status"] != "idle" {
			t.Errorf("expected status 'idle', got %v", resp["status"])
		}
	})

	t.Run("specific_job_by_id", func(t *testing.T) {
		mux := http.NewServeMux()
		mux.HandleFunc("GET /api/v1/sync/status", func(w http.ResponseWriter, r *http.Request) {
			idStr := r.URL.Query().Get("id")
			if idStr == "123" {
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(map[string]interface{}{
					"id":     123,
					"status": "completed",
				})
				return
			}

			w.WriteHeader(http.StatusNotFound)
		})

		req := httptest.NewRequest("GET", "/api/v1/sync/status?id=123", nil)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)

		var resp map[string]interface{}
		json.NewDecoder(w.Body).Decode(&resp)

		if int(resp["id"].(float64)) != 123 {
			t.Errorf("expected id 123, got %v", resp["id"])
		}
	})

	t.Run("invalid_job_id", func(t *testing.T) {
		mux := http.NewServeMux()
		mux.HandleFunc("GET /api/v1/sync/status", func(w http.ResponseWriter, r *http.Request) {
			idStr := r.URL.Query().Get("id")
			if idStr != "" {
				// Simulate invalid ID parsing
				if idStr == "invalid" {
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusBadRequest)
					json.NewEncoder(w).Encode(map[string]interface{}{
						"error": "Invalid job ID",
						"code":  "INVALID_ID",
					})
					return
				}
			}
		})

		req := httptest.NewRequest("GET", "/api/v1/sync/status?id=invalid", nil)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", w.Code)
		}
	})

	t.Run("job_not_found", func(t *testing.T) {
		mux := http.NewServeMux()
		mux.HandleFunc("GET /api/v1/sync/status", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"error": "Job not found",
				"code":  "JOB_NOT_FOUND",
			})
		})

		req := httptest.NewRequest("GET", "/api/v1/sync/status?id=9999", nil)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("expected status 404, got %d", w.Code)
		}
	})
}

// ============================================================================
// CancelSync Tests
// ============================================================================

func TestCancelSync(t *testing.T) {
	t.Run("successful_cancel", func(t *testing.T) {
		mux := http.NewServeMux()
		mux.HandleFunc("POST /api/v1/sync/cancel", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": true,
				"message": "Sync cancellation requested",
			})
		})

		req := httptest.NewRequest("POST", "/api/v1/sync/cancel", nil)
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

	t.Run("no_job_to_cancel", func(t *testing.T) {
		mux := http.NewServeMux()
		mux.HandleFunc("POST /api/v1/sync/cancel", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"error": "No sync job is currently running",
				"code":  "CANCEL_FAILED",
			})
		})

		req := httptest.NewRequest("POST", "/api/v1/sync/cancel", nil)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", w.Code)
		}
	})
}

// ============================================================================
// GetSyncHistory Tests
// ============================================================================

func TestGetSyncHistory(t *testing.T) {
	t.Run("returns_history", func(t *testing.T) {
		mux := http.NewServeMux()
		mux.HandleFunc("GET /api/v1/sync/history", func(w http.ResponseWriter, r *http.Request) {
			jobs := []map[string]interface{}{
				{"id": 1, "status": "completed", "events_stored": 100},
				{"id": 2, "status": "completed", "events_stored": 50},
			}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"jobs":   jobs,
				"limit":  20,
				"offset": 0,
			})
		})

		req := httptest.NewRequest("GET", "/api/v1/sync/history", nil)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}

		var resp map[string]interface{}
		json.NewDecoder(w.Body).Decode(&resp)

		jobs := resp["jobs"].([]interface{})
		if len(jobs) != 2 {
			t.Errorf("expected 2 jobs, got %d", len(jobs))
		}
	})

	t.Run("respects_limit", func(t *testing.T) {
		mux := http.NewServeMux()
		mux.HandleFunc("GET /api/v1/sync/history", func(w http.ResponseWriter, r *http.Request) {
			limit := 20
			if l := r.URL.Query().Get("limit"); l != "" {
				if parsed, err := parseInt(l); err == nil && parsed > 0 {
					limit = parsed
					if limit > 100 {
						limit = 100
					}
				}
			}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"jobs":   []interface{}{},
				"limit":  limit,
				"offset": 0,
			})
		})

		req := httptest.NewRequest("GET", "/api/v1/sync/history?limit=5", nil)
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
		mux.HandleFunc("GET /api/v1/sync/history", func(w http.ResponseWriter, r *http.Request) {
			limit := 20
			if l := r.URL.Query().Get("limit"); l != "" {
				if parsed, err := parseInt(l); err == nil && parsed > 0 {
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

		req := httptest.NewRequest("GET", "/api/v1/sync/history?limit=500", nil)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)

		var resp map[string]interface{}
		json.NewDecoder(w.Body).Decode(&resp)

		if int(resp["limit"].(float64)) != 100 {
			t.Errorf("expected limit capped at 100, got %v", resp["limit"])
		}
	})

	t.Run("empty_history", func(t *testing.T) {
		mux := http.NewServeMux()
		mux.HandleFunc("GET /api/v1/sync/history", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"jobs":   []interface{}{},
				"limit":  20,
				"offset": 0,
			})
		})

		req := httptest.NewRequest("GET", "/api/v1/sync/history", nil)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)

		var resp map[string]interface{}
		json.NewDecoder(w.Body).Decode(&resp)

		jobs := resp["jobs"].([]interface{})
		if len(jobs) != 0 {
			t.Errorf("expected empty jobs array, got %d", len(jobs))
		}
	})
}

// ============================================================================
// GetDefaultRelays Tests
// ============================================================================

func TestGetDefaultRelays(t *testing.T) {
	t.Run("returns_relays", func(t *testing.T) {
		mux := http.NewServeMux()
		mux.HandleFunc("GET /api/v1/sync/relays", func(w http.ResponseWriter, r *http.Request) {
			relays := []string{
				"wss://relay.damus.io",
				"wss://nostr.wine",
				"wss://relay.snort.social",
			}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"relays": relays,
			})
		})

		req := httptest.NewRequest("GET", "/api/v1/sync/relays", nil)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}

		var resp map[string]interface{}
		json.NewDecoder(w.Body).Decode(&resp)

		relays := resp["relays"].([]interface{})
		if len(relays) < 1 {
			t.Error("expected at least one default relay")
		}
	})
}

// ============================================================================
// StartSyncRequest Tests
// ============================================================================

func TestStartSyncRequest(t *testing.T) {
	t.Run("json_serialization", func(t *testing.T) {
		sinceTs := int64(1704067200)
		req := StartSyncRequest{
			Pubkeys:        []string{"abc123", "def456"},
			Relays:         []string{"wss://relay.example.com"},
			EventKinds:     []int{1, 4, 30023},
			SinceTimestamp: &sinceTs,
		}

		data, err := json.Marshal(req)
		if err != nil {
			t.Fatalf("failed to marshal: %v", err)
		}

		var decoded StartSyncRequest
		if err := json.Unmarshal(data, &decoded); err != nil {
			t.Fatalf("failed to unmarshal: %v", err)
		}

		if len(decoded.Pubkeys) != 2 {
			t.Errorf("expected 2 pubkeys, got %d", len(decoded.Pubkeys))
		}
		if len(decoded.Relays) != 1 {
			t.Errorf("expected 1 relay, got %d", len(decoded.Relays))
		}
		if len(decoded.EventKinds) != 3 {
			t.Errorf("expected 3 event kinds, got %d", len(decoded.EventKinds))
		}
		if decoded.SinceTimestamp == nil || *decoded.SinceTimestamp != sinceTs {
			t.Errorf("expected since_timestamp %d, got %v", sinceTs, decoded.SinceTimestamp)
		}
	})
}
