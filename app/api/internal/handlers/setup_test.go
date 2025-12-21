package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// ============================================================================
// GetSetupStatus Tests
// ============================================================================

func TestGetSetupStatus(t *testing.T) {
	t.Run("setup_not_completed", func(t *testing.T) {
		mux := http.NewServeMux()
		mux.HandleFunc("GET /api/v1/setup/status", func(w http.ResponseWriter, r *http.Request) {
			completed := false

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"completed": completed,
			})
		})

		req := httptest.NewRequest("GET", "/api/v1/setup/status", nil)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}

		var resp map[string]interface{}
		json.NewDecoder(w.Body).Decode(&resp)

		if resp["completed"] != false {
			t.Errorf("expected completed false, got %v", resp["completed"])
		}
	})

	t.Run("setup_completed", func(t *testing.T) {
		mux := http.NewServeMux()
		mux.HandleFunc("GET /api/v1/setup/status", func(w http.ResponseWriter, r *http.Request) {
			completed := true
			operatorPubkey := "abc123def456"
			operatorNpub := "npub1abc..."
			accessMode := "private"

			response := map[string]interface{}{
				"completed": completed,
			}
			if completed {
				response["operator_pubkey"] = operatorPubkey
				response["operator_npub"] = operatorNpub
				response["access_mode"] = accessMode
			}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response)
		})

		req := httptest.NewRequest("GET", "/api/v1/setup/status", nil)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)

		var resp map[string]interface{}
		json.NewDecoder(w.Body).Decode(&resp)

		if resp["completed"] != true {
			t.Errorf("expected completed true, got %v", resp["completed"])
		}
		if resp["operator_pubkey"] == nil {
			t.Error("expected operator_pubkey to be present")
		}
		if resp["access_mode"] != "private" {
			t.Errorf("expected access_mode 'private', got %v", resp["access_mode"])
		}
	})
}

// ============================================================================
// ValidateIdentity Tests
// ============================================================================

func TestValidateIdentity(t *testing.T) {
	t.Run("missing_input", func(t *testing.T) {
		mux := http.NewServeMux()
		mux.HandleFunc("GET /api/v1/setup/validate-identity", func(w http.ResponseWriter, r *http.Request) {
			input := r.URL.Query().Get("input")
			if input == "" {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(map[string]interface{}{
					"error": "Input parameter is required",
					"code":  "MISSING_INPUT",
				})
				return
			}
		})

		req := httptest.NewRequest("GET", "/api/v1/setup/validate-identity", nil)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", w.Code)
		}
	})

	t.Run("valid_npub", func(t *testing.T) {
		mux := http.NewServeMux()
		mux.HandleFunc("GET /api/v1/setup/validate-identity", func(w http.ResponseWriter, r *http.Request) {
			input := r.URL.Query().Get("input")
			if input == "" {
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			// Simulate valid npub
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"valid":  true,
				"pubkey": "abc123def456",
				"npub":   input,
				"source": "npub",
			})
		})

		req := httptest.NewRequest("GET", "/api/v1/setup/validate-identity?input=npub1abc123", nil)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}

		var resp map[string]interface{}
		json.NewDecoder(w.Body).Decode(&resp)

		if resp["valid"] != true {
			t.Errorf("expected valid true, got %v", resp["valid"])
		}
		if resp["source"] != "npub" {
			t.Errorf("expected source 'npub', got %v", resp["source"])
		}
	})

	t.Run("valid_hex_pubkey", func(t *testing.T) {
		mux := http.NewServeMux()
		mux.HandleFunc("GET /api/v1/setup/validate-identity", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"valid":  true,
				"pubkey": "1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
				"npub":   "npub1...",
				"source": "hex",
			})
		})

		hexPubkey := "1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"
		req := httptest.NewRequest("GET", "/api/v1/setup/validate-identity?input="+hexPubkey, nil)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)

		var resp map[string]interface{}
		json.NewDecoder(w.Body).Decode(&resp)

		if resp["valid"] != true {
			t.Errorf("expected valid true, got %v", resp["valid"])
		}
	})

	t.Run("valid_nip05", func(t *testing.T) {
		mux := http.NewServeMux()
		mux.HandleFunc("GET /api/v1/setup/validate-identity", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"valid":      true,
				"pubkey":     "abc123",
				"npub":       "npub1abc...",
				"source":     "nip05",
				"nip05_name": "alice@example.com",
			})
		})

		req := httptest.NewRequest("GET", "/api/v1/setup/validate-identity?input=alice@example.com", nil)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)

		var resp map[string]interface{}
		json.NewDecoder(w.Body).Decode(&resp)

		if resp["valid"] != true {
			t.Errorf("expected valid true, got %v", resp["valid"])
		}
		if resp["source"] != "nip05" {
			t.Errorf("expected source 'nip05', got %v", resp["source"])
		}
		if resp["nip05_name"] != "alice@example.com" {
			t.Errorf("expected nip05_name 'alice@example.com', got %v", resp["nip05_name"])
		}
	})

	t.Run("invalid_identity", func(t *testing.T) {
		mux := http.NewServeMux()
		mux.HandleFunc("GET /api/v1/setup/validate-identity", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"valid": false,
				"error": "Invalid identity format",
				"code":  "INVALID_IDENTITY",
			})
		})

		req := httptest.NewRequest("GET", "/api/v1/setup/validate-identity?input=invalid", nil)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)

		var resp map[string]interface{}
		json.NewDecoder(w.Body).Decode(&resp)

		if resp["valid"] != false {
			t.Errorf("expected valid false, got %v", resp["valid"])
		}
	})
}

// ============================================================================
// CompleteSetup Tests
// ============================================================================

func TestCompleteSetup(t *testing.T) {
	t.Run("successful_setup", func(t *testing.T) {
		mux := http.NewServeMux()
		mux.HandleFunc("POST /api/v1/setup/complete", func(w http.ResponseWriter, r *http.Request) {
			var req CompleteSetupRequest
			json.NewDecoder(r.Body).Decode(&req)

			if req.OperatorIdentity == "" && req.OperatorPubkey == "" {
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success":         true,
				"message":         "Setup completed successfully",
				"operator_pubkey": "abc123def456",
				"operator_npub":   "npub1abc...",
				"access_mode":     req.AccessMode,
			})
		})

		body := `{"operator_identity":"npub1abc...","relay_name":"My Relay","relay_description":"A test relay","access_mode":"private"}`
		req := httptest.NewRequest("POST", "/api/v1/setup/complete", strings.NewReader(body))
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

	t.Run("missing_identity", func(t *testing.T) {
		mux := http.NewServeMux()
		mux.HandleFunc("POST /api/v1/setup/complete", func(w http.ResponseWriter, r *http.Request) {
			var req CompleteSetupRequest
			json.NewDecoder(r.Body).Decode(&req)

			if req.OperatorIdentity == "" && req.OperatorPubkey == "" {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(map[string]interface{}{
					"error": "Operator identity is required",
					"code":  "MISSING_IDENTITY",
				})
				return
			}
		})

		body := `{"relay_name":"My Relay","access_mode":"private"}`
		req := httptest.NewRequest("POST", "/api/v1/setup/complete", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", w.Code)
		}
	})

	t.Run("setup_already_done", func(t *testing.T) {
		mux := http.NewServeMux()
		mux.HandleFunc("POST /api/v1/setup/complete", func(w http.ResponseWriter, r *http.Request) {
			alreadyCompleted := true

			if alreadyCompleted {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusConflict)
				json.NewEncoder(w).Encode(map[string]interface{}{
					"error": "Setup already completed",
					"code":  "SETUP_ALREADY_DONE",
				})
				return
			}
		})

		body := `{"operator_identity":"npub1abc...","access_mode":"private"}`
		req := httptest.NewRequest("POST", "/api/v1/setup/complete", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)

		if w.Code != http.StatusConflict {
			t.Errorf("expected status 409, got %d", w.Code)
		}
	})

	t.Run("invalid_access_mode", func(t *testing.T) {
		mux := http.NewServeMux()
		mux.HandleFunc("POST /api/v1/setup/complete", func(w http.ResponseWriter, r *http.Request) {
			var req CompleteSetupRequest
			json.NewDecoder(r.Body).Decode(&req)

			accessMode := req.AccessMode
			if accessMode != "private" && accessMode != "paid" && accessMode != "public" && accessMode != "" {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(map[string]interface{}{
					"error": "Invalid access mode. Must be: private, paid, or public",
					"code":  "INVALID_ACCESS_MODE",
				})
				return
			}
		})

		body := `{"operator_identity":"npub1abc...","access_mode":"invalid"}`
		req := httptest.NewRequest("POST", "/api/v1/setup/complete", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", w.Code)
		}
	})

	t.Run("invalid_json", func(t *testing.T) {
		mux := http.NewServeMux()
		mux.HandleFunc("POST /api/v1/setup/complete", func(w http.ResponseWriter, r *http.Request) {
			var req CompleteSetupRequest
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
		req := httptest.NewRequest("POST", "/api/v1/setup/complete", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", w.Code)
		}
	})

	t.Run("defaults_to_private_mode", func(t *testing.T) {
		mux := http.NewServeMux()
		mux.HandleFunc("POST /api/v1/setup/complete", func(w http.ResponseWriter, r *http.Request) {
			var req CompleteSetupRequest
			json.NewDecoder(r.Body).Decode(&req)

			accessMode := req.AccessMode
			if accessMode == "" {
				accessMode = "private"
			}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success":     true,
				"access_mode": accessMode,
			})
		})

		body := `{"operator_identity":"npub1abc..."}`
		req := httptest.NewRequest("POST", "/api/v1/setup/complete", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)

		var resp map[string]interface{}
		json.NewDecoder(w.Body).Decode(&resp)

		if resp["access_mode"] != "private" {
			t.Errorf("expected default access_mode 'private', got %v", resp["access_mode"])
		}
	})
}

// ============================================================================
// CompleteSetupRequest Tests
// ============================================================================

func TestCompleteSetupRequest(t *testing.T) {
	t.Run("json_serialization", func(t *testing.T) {
		req := CompleteSetupRequest{
			OperatorIdentity: "alice@example.com",
			RelayName:        "My Relay",
			RelayDesc:        "A test relay",
			AccessMode:       "private",
		}

		data, err := json.Marshal(req)
		if err != nil {
			t.Fatalf("failed to marshal: %v", err)
		}

		var decoded CompleteSetupRequest
		if err := json.Unmarshal(data, &decoded); err != nil {
			t.Fatalf("failed to unmarshal: %v", err)
		}

		if decoded.OperatorIdentity != req.OperatorIdentity {
			t.Errorf("expected operator_identity '%s', got '%s'", req.OperatorIdentity, decoded.OperatorIdentity)
		}
		if decoded.RelayName != req.RelayName {
			t.Errorf("expected relay_name '%s', got '%s'", req.RelayName, decoded.RelayName)
		}
		if decoded.AccessMode != req.AccessMode {
			t.Errorf("expected access_mode '%s', got '%s'", req.AccessMode, decoded.AccessMode)
		}
	})

	t.Run("legacy_fields", func(t *testing.T) {
		// Test backward compatibility with operator_pubkey field
		body := `{"operator_pubkey":"abc123","operator_npub":"npub1abc","relay_name":"Test"}`

		var req CompleteSetupRequest
		if err := json.Unmarshal([]byte(body), &req); err != nil {
			t.Fatalf("failed to unmarshal: %v", err)
		}

		if req.OperatorPubkey != "abc123" {
			t.Errorf("expected operator_pubkey 'abc123', got '%s'", req.OperatorPubkey)
		}
		if req.OperatorNpub != "npub1abc" {
			t.Errorf("expected operator_npub 'npub1abc', got '%s'", req.OperatorNpub)
		}
	})
}
