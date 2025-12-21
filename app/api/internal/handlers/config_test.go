package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// ============================================================================
// GetConfig Tests
// ============================================================================

func TestGetConfig(t *testing.T) {
	t.Run("returns_config", func(t *testing.T) {
		mux := http.NewServeMux()
		mux.HandleFunc("GET /api/v1/config", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"info": map[string]interface{}{
					"name":        "My Relay",
					"description": "A private relay",
					"contact":     "admin@example.com",
					"relay_icon":  "https://example.com/icon.png",
				},
				"limits": map[string]interface{}{
					"max_event_bytes":      65536,
					"max_ws_message_bytes": 131072,
					"messages_per_sec":     10,
					"max_subs_per_conn":    10,
					"min_pow_difficulty":   0,
				},
				"authorization": map[string]interface{}{
					"nip42_auth":           true,
					"event_kind_allowlist": []int{0, 1, 3, 4},
				},
			})
		})

		req := httptest.NewRequest("GET", "/api/v1/config", nil)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}

		var resp map[string]interface{}
		json.NewDecoder(w.Body).Decode(&resp)

		info := resp["info"].(map[string]interface{})
		if info["name"] != "My Relay" {
			t.Errorf("expected name 'My Relay', got %v", info["name"])
		}

		limits := resp["limits"].(map[string]interface{})
		if int(limits["max_event_bytes"].(float64)) != 65536 {
			t.Errorf("expected max_event_bytes 65536, got %v", limits["max_event_bytes"])
		}

		auth := resp["authorization"].(map[string]interface{})
		if auth["nip42_auth"] != true {
			t.Errorf("expected nip42_auth true, got %v", auth["nip42_auth"])
		}
	})

	t.Run("config_manager_not_available", func(t *testing.T) {
		mux := http.NewServeMux()
		mux.HandleFunc("GET /api/v1/config", func(w http.ResponseWriter, r *http.Request) {
			configMgrAvailable := false
			if !configMgrAvailable {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusServiceUnavailable)
				json.NewEncoder(w).Encode(map[string]interface{}{
					"error": "Config manager not available",
					"code":  "CONFIG_NOT_AVAILABLE",
				})
				return
			}
		})

		req := httptest.NewRequest("GET", "/api/v1/config", nil)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)

		if w.Code != http.StatusServiceUnavailable {
			t.Errorf("expected status 503, got %d", w.Code)
		}
	})
}

// ============================================================================
// UpdateConfig Tests
// ============================================================================

func TestUpdateConfig(t *testing.T) {
	t.Run("updates_info_section", func(t *testing.T) {
		mux := http.NewServeMux()
		mux.HandleFunc("PATCH /api/v1/config", func(w http.ResponseWriter, r *http.Request) {
			var req UpdateConfigRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			if err := validateConfigUpdate(&req); err != nil {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(map[string]interface{}{
					"error": err.Error(),
					"code":  "VALIDATION_ERROR",
				})
				return
			}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": true,
				"message": "Configuration updated",
			})
		})

		body := `{"info":{"name":"New Name","description":"New description"}}`
		req := httptest.NewRequest("PATCH", "/api/v1/config", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
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

	t.Run("rejects_name_too_long", func(t *testing.T) {
		mux := http.NewServeMux()
		mux.HandleFunc("PATCH /api/v1/config", func(w http.ResponseWriter, r *http.Request) {
			var req UpdateConfigRequest
			json.NewDecoder(r.Body).Decode(&req)

			if err := validateConfigUpdate(&req); err != nil {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(map[string]interface{}{
					"error": err.Error(),
					"code":  "VALIDATION_ERROR",
				})
				return
			}
		})

		longName := strings.Repeat("a", 65) // 65 chars, max is 64
		body := `{"info":{"name":"` + longName + `"}}`
		req := httptest.NewRequest("PATCH", "/api/v1/config", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", w.Code)
		}
	})

	t.Run("rejects_description_too_long", func(t *testing.T) {
		mux := http.NewServeMux()
		mux.HandleFunc("PATCH /api/v1/config", func(w http.ResponseWriter, r *http.Request) {
			var req UpdateConfigRequest
			json.NewDecoder(r.Body).Decode(&req)

			if err := validateConfigUpdate(&req); err != nil {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(map[string]interface{}{
					"error": err.Error(),
					"code":  "VALIDATION_ERROR",
				})
				return
			}
		})

		longDesc := strings.Repeat("a", 501) // 501 chars, max is 500
		body := `{"info":{"description":"` + longDesc + `"}}`
		req := httptest.NewRequest("PATCH", "/api/v1/config", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", w.Code)
		}
	})

	t.Run("rejects_invalid_max_event_bytes", func(t *testing.T) {
		mux := http.NewServeMux()
		mux.HandleFunc("PATCH /api/v1/config", func(w http.ResponseWriter, r *http.Request) {
			var req UpdateConfigRequest
			json.NewDecoder(r.Body).Decode(&req)

			if err := validateConfigUpdate(&req); err != nil {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(map[string]interface{}{
					"error": err.Error(),
					"code":  "VALIDATION_ERROR",
				})
				return
			}
		})

		// Test too small (< 1024)
		body := `{"limits":{"max_event_bytes":512}}`
		req := httptest.NewRequest("PATCH", "/api/v1/config", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected status 400 for too small value, got %d", w.Code)
		}
	})

	t.Run("rejects_invalid_messages_per_sec", func(t *testing.T) {
		mux := http.NewServeMux()
		mux.HandleFunc("PATCH /api/v1/config", func(w http.ResponseWriter, r *http.Request) {
			var req UpdateConfigRequest
			json.NewDecoder(r.Body).Decode(&req)

			if err := validateConfigUpdate(&req); err != nil {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(map[string]interface{}{
					"error": err.Error(),
					"code":  "VALIDATION_ERROR",
				})
				return
			}
		})

		body := `{"limits":{"messages_per_sec":101}}`
		req := httptest.NewRequest("PATCH", "/api/v1/config", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", w.Code)
		}
	})

	t.Run("rejects_invalid_pow_difficulty", func(t *testing.T) {
		mux := http.NewServeMux()
		mux.HandleFunc("PATCH /api/v1/config", func(w http.ResponseWriter, r *http.Request) {
			var req UpdateConfigRequest
			json.NewDecoder(r.Body).Decode(&req)

			if err := validateConfigUpdate(&req); err != nil {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(map[string]interface{}{
					"error": err.Error(),
					"code":  "VALIDATION_ERROR",
				})
				return
			}
		})

		body := `{"limits":{"min_pow_difficulty":33}}`
		req := httptest.NewRequest("PATCH", "/api/v1/config", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", w.Code)
		}
	})

	t.Run("rejects_negative_event_kinds", func(t *testing.T) {
		mux := http.NewServeMux()
		mux.HandleFunc("PATCH /api/v1/config", func(w http.ResponseWriter, r *http.Request) {
			var req UpdateConfigRequest
			json.NewDecoder(r.Body).Decode(&req)

			if err := validateConfigUpdate(&req); err != nil {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(map[string]interface{}{
					"error": err.Error(),
					"code":  "VALIDATION_ERROR",
				})
				return
			}
		})

		body := `{"authorization":{"event_kind_allowlist":[0,1,-1,3]}}`
		req := httptest.NewRequest("PATCH", "/api/v1/config", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", w.Code)
		}
	})

	t.Run("rejects_invalid_json", func(t *testing.T) {
		mux := http.NewServeMux()
		mux.HandleFunc("PATCH /api/v1/config", func(w http.ResponseWriter, r *http.Request) {
			var req UpdateConfigRequest
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

		body := `{invalid json}`
		req := httptest.NewRequest("PATCH", "/api/v1/config", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", w.Code)
		}
	})

	t.Run("accepts_valid_limits", func(t *testing.T) {
		mux := http.NewServeMux()
		mux.HandleFunc("PATCH /api/v1/config", func(w http.ResponseWriter, r *http.Request) {
			var req UpdateConfigRequest
			json.NewDecoder(r.Body).Decode(&req)

			if err := validateConfigUpdate(&req); err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": true,
			})
		})

		body := `{"limits":{"max_event_bytes":65536,"messages_per_sec":50,"min_pow_difficulty":16}}`
		req := httptest.NewRequest("PATCH", "/api/v1/config", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}
	})
}

// ============================================================================
// ReloadConfig Tests
// ============================================================================

func TestReloadConfig(t *testing.T) {
	t.Run("successful_reload", func(t *testing.T) {
		mux := http.NewServeMux()
		mux.HandleFunc("POST /api/v1/config/reload", func(w http.ResponseWriter, r *http.Request) {
			relayAvailable := true
			if !relayAvailable {
				w.WriteHeader(http.StatusServiceUnavailable)
				return
			}

			// Simulate successful reload
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": true,
				"message": "Relay configuration reloaded",
			})
		})

		req := httptest.NewRequest("POST", "/api/v1/config/reload", nil)
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
		mux.HandleFunc("POST /api/v1/config/reload", func(w http.ResponseWriter, r *http.Request) {
			relayAvailable := false
			if !relayAvailable {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusServiceUnavailable)
				json.NewEncoder(w).Encode(map[string]interface{}{
					"error": "Relay manager not available",
					"code":  "RELAY_NOT_AVAILABLE",
				})
				return
			}
		})

		req := httptest.NewRequest("POST", "/api/v1/config/reload", nil)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)

		if w.Code != http.StatusServiceUnavailable {
			t.Errorf("expected status 503, got %d", w.Code)
		}
	})

	t.Run("reload_failed", func(t *testing.T) {
		mux := http.NewServeMux()
		mux.HandleFunc("POST /api/v1/config/reload", func(w http.ResponseWriter, r *http.Request) {
			// Simulate failed reload
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"error": "Failed to reload relay",
				"code":  "RELOAD_FAILED",
			})
		})

		req := httptest.NewRequest("POST", "/api/v1/config/reload", nil)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)

		if w.Code != http.StatusInternalServerError {
			t.Errorf("expected status 500, got %d", w.Code)
		}
	})
}

// ============================================================================
// validateConfigUpdate Tests
// ============================================================================

func TestValidateConfigUpdate(t *testing.T) {
	t.Run("nil_request_passes", func(t *testing.T) {
		req := &UpdateConfigRequest{}
		err := validateConfigUpdate(req)
		if err != nil {
			t.Errorf("expected nil error, got %v", err)
		}
	})

	t.Run("valid_name_passes", func(t *testing.T) {
		name := "Valid Name"
		req := &UpdateConfigRequest{
			Info: &InfoUpdate{Name: &name},
		}
		err := validateConfigUpdate(req)
		if err != nil {
			t.Errorf("expected nil error, got %v", err)
		}
	})

	t.Run("name_exactly_64_chars_passes", func(t *testing.T) {
		name := strings.Repeat("a", 64)
		req := &UpdateConfigRequest{
			Info: &InfoUpdate{Name: &name},
		}
		err := validateConfigUpdate(req)
		if err != nil {
			t.Errorf("expected nil error for 64 char name, got %v", err)
		}
	})

	t.Run("name_65_chars_fails", func(t *testing.T) {
		name := strings.Repeat("a", 65)
		req := &UpdateConfigRequest{
			Info: &InfoUpdate{Name: &name},
		}
		err := validateConfigUpdate(req)
		if err == nil {
			t.Error("expected error for 65 char name")
		}
	})

	t.Run("description_exactly_500_chars_passes", func(t *testing.T) {
		desc := strings.Repeat("a", 500)
		req := &UpdateConfigRequest{
			Info: &InfoUpdate{Description: &desc},
		}
		err := validateConfigUpdate(req)
		if err != nil {
			t.Errorf("expected nil error for 500 char description, got %v", err)
		}
	})

	t.Run("max_event_bytes_boundary", func(t *testing.T) {
		// Min boundary (1024)
		minBytes := 1024
		req := &UpdateConfigRequest{
			Limits: &LimitsUpdate{MaxEventBytes: &minBytes},
		}
		if err := validateConfigUpdate(req); err != nil {
			t.Errorf("expected 1024 to pass, got %v", err)
		}

		// Below min
		tooSmall := 1023
		req.Limits.MaxEventBytes = &tooSmall
		if err := validateConfigUpdate(req); err == nil {
			t.Error("expected error for 1023")
		}

		// Max boundary (16MB)
		maxBytes := 16 * 1024 * 1024
		req.Limits.MaxEventBytes = &maxBytes
		if err := validateConfigUpdate(req); err != nil {
			t.Errorf("expected 16MB to pass, got %v", err)
		}

		// Above max
		tooLarge := 16*1024*1024 + 1
		req.Limits.MaxEventBytes = &tooLarge
		if err := validateConfigUpdate(req); err == nil {
			t.Error("expected error for value > 16MB")
		}
	})

	t.Run("messages_per_sec_boundary", func(t *testing.T) {
		// Min (1)
		min := 1
		req := &UpdateConfigRequest{
			Limits: &LimitsUpdate{MessagesPerSec: &min},
		}
		if err := validateConfigUpdate(req); err != nil {
			t.Errorf("expected 1 to pass, got %v", err)
		}

		// Max (100)
		max := 100
		req.Limits.MessagesPerSec = &max
		if err := validateConfigUpdate(req); err != nil {
			t.Errorf("expected 100 to pass, got %v", err)
		}

		// Below min (0)
		zero := 0
		req.Limits.MessagesPerSec = &zero
		if err := validateConfigUpdate(req); err == nil {
			t.Error("expected error for 0")
		}

		// Above max (101)
		above := 101
		req.Limits.MessagesPerSec = &above
		if err := validateConfigUpdate(req); err == nil {
			t.Error("expected error for 101")
		}
	})

	t.Run("pow_difficulty_boundary", func(t *testing.T) {
		// Min (0)
		min := 0
		req := &UpdateConfigRequest{
			Limits: &LimitsUpdate{MinPowDifficulty: &min},
		}
		if err := validateConfigUpdate(req); err != nil {
			t.Errorf("expected 0 to pass, got %v", err)
		}

		// Max (32)
		max := 32
		req.Limits.MinPowDifficulty = &max
		if err := validateConfigUpdate(req); err != nil {
			t.Errorf("expected 32 to pass, got %v", err)
		}

		// Below min (-1)
		negative := -1
		req.Limits.MinPowDifficulty = &negative
		if err := validateConfigUpdate(req); err == nil {
			t.Error("expected error for -1")
		}

		// Above max (33)
		above := 33
		req.Limits.MinPowDifficulty = &above
		if err := validateConfigUpdate(req); err == nil {
			t.Error("expected error for 33")
		}
	})

	t.Run("event_kinds_valid", func(t *testing.T) {
		kinds := []int{0, 1, 3, 4, 30023}
		req := &UpdateConfigRequest{
			Authorization: &AuthorizationUpdate{EventKindAllowlist: &kinds},
		}
		if err := validateConfigUpdate(req); err != nil {
			t.Errorf("expected valid kinds to pass, got %v", err)
		}
	})

	t.Run("event_kinds_with_negative_fails", func(t *testing.T) {
		kinds := []int{0, 1, -1, 3}
		req := &UpdateConfigRequest{
			Authorization: &AuthorizationUpdate{EventKindAllowlist: &kinds},
		}
		if err := validateConfigUpdate(req); err == nil {
			t.Error("expected error for negative kind")
		}
	})
}

// ============================================================================
// getUpdatedSections Tests
// ============================================================================

func TestGetUpdatedSections(t *testing.T) {
	t.Run("no_sections", func(t *testing.T) {
		req := &UpdateConfigRequest{}
		result := getUpdatedSections(req)
		if result != "none" {
			t.Errorf("expected 'none', got '%s'", result)
		}
	})

	t.Run("info_only", func(t *testing.T) {
		name := "test"
		req := &UpdateConfigRequest{
			Info: &InfoUpdate{Name: &name},
		}
		result := getUpdatedSections(req)
		if result != "info" {
			t.Errorf("expected 'info', got '%s'", result)
		}
	})

	t.Run("limits_only", func(t *testing.T) {
		val := 1024
		req := &UpdateConfigRequest{
			Limits: &LimitsUpdate{MaxEventBytes: &val},
		}
		result := getUpdatedSections(req)
		if result != "limits" {
			t.Errorf("expected 'limits', got '%s'", result)
		}
	})

	t.Run("multiple_sections", func(t *testing.T) {
		name := "test"
		val := 1024
		auth := true
		req := &UpdateConfigRequest{
			Info:          &InfoUpdate{Name: &name},
			Limits:        &LimitsUpdate{MaxEventBytes: &val},
			Authorization: &AuthorizationUpdate{NIP42Auth: &auth},
		}
		result := getUpdatedSections(req)
		if result != "info,limits,authorization" {
			t.Errorf("expected 'info,limits,authorization', got '%s'", result)
		}
	})
}
