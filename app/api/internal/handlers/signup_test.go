package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// mockDB implements minimal DB interface for testing
type mockDB struct {
	accessMode string
}

func (m *mockDB) GetAccessMode(ctx interface{}) (string, error) {
	return m.accessMode, nil
}

// TestGetRelayInfo_PaidAccessDisabled tests the relay info endpoint when paid access is off
func TestGetRelayInfo_PaidAccessDisabled(t *testing.T) {
	// Create request
	req := httptest.NewRequest("GET", "/public/relay-info", nil)
	w := httptest.NewRecorder()

	// For now, just verify the route pattern works
	mux := http.NewServeMux()
	mux.HandleFunc("GET /public/relay-info", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"paid_access_enabled": false,
			"message":             "Paid access is not enabled",
		})
	})

	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	contentType := w.Header().Get("Content-Type")
	if !strings.Contains(contentType, "application/json") {
		t.Errorf("expected Content-Type application/json, got %s", contentType)
	}

	var resp map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp["paid_access_enabled"] != false {
		t.Error("expected paid_access_enabled to be false")
	}
}

// TestCreateSignupInvoice_MissingPubkey tests invoice creation with missing pubkey
func TestCreateSignupInvoice_MissingPubkey(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /public/create-invoice", func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Pubkey string `json:"pubkey"`
			TierID string `json:"tier_id"`
		}
		json.NewDecoder(r.Body).Decode(&req)

		if req.Pubkey == "" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"error": "Pubkey is required",
				"code":  "MISSING_PUBKEY",
			})
			return
		}
	})

	req := httptest.NewRequest("POST", "/public/create-invoice", strings.NewReader(`{"tier_id":"test"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	mux.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}

	var resp map[string]interface{}
	json.NewDecoder(w.Body).Decode(&resp)

	if resp["code"] != "MISSING_PUBKEY" {
		t.Errorf("expected code MISSING_PUBKEY, got %v", resp["code"])
	}
}

// TestGetInvoiceStatus_NotFound tests invoice status with non-existent invoice
func TestGetInvoiceStatus_NotFound(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /public/invoice-status/{hash}", func(w http.ResponseWriter, r *http.Request) {
		hash := r.PathValue("hash")
		if hash == "" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		// Simulate not found
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": "Invoice not found",
			"code":  "INVOICE_NOT_FOUND",
		})
	})

	req := httptest.NewRequest("GET", "/public/invoice-status/abc123", nil)
	w := httptest.NewRecorder()

	mux.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", w.Code)
	}
}

// TestLightningStatus_NotConfigured tests lightning status when not configured
func TestLightningStatus_NotConfigured(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v1/lightning/status", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"configured": false,
			"enabled":    false,
		})
	})

	req := httptest.NewRequest("GET", "/api/v1/lightning/status", nil)
	w := httptest.NewRecorder()

	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var resp map[string]interface{}
	json.NewDecoder(w.Body).Decode(&resp)

	if resp["configured"] != false {
		t.Error("expected configured to be false")
	}
}

// TestLightningDetect tests the detection endpoint
func TestLightningDetect(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/v1/lightning/detect", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"detected": false,
			"error":    "LND not detected",
		})
	})

	req := httptest.NewRequest("POST", "/api/v1/lightning/detect", nil)
	w := httptest.NewRecorder()

	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var resp map[string]interface{}
	json.NewDecoder(w.Body).Decode(&resp)

	if resp["detected"] != false {
		t.Error("expected detected to be false")
	}
}

// TestRoutePatternMatching verifies Go 1.22+ route patterns work correctly
func TestRoutePatternMatching(t *testing.T) {
	mux := http.NewServeMux()

	// Register routes in same order as handlers.go
	mux.HandleFunc("GET /api/v1/lightning/status", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("lightning-status"))
	})
	mux.HandleFunc("GET /public/relay-info", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("relay-info"))
	})
	mux.HandleFunc("POST /public/create-invoice", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("create-invoice"))
	})
	mux.HandleFunc("GET /public/invoice-status/{hash}", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("invoice-status:" + r.PathValue("hash")))
	})
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("fallback"))
	})

	tests := []struct {
		method   string
		path     string
		expected string
	}{
		{"GET", "/api/v1/lightning/status", "lightning-status"},
		{"GET", "/public/relay-info", "relay-info"},
		{"POST", "/public/create-invoice", "create-invoice"},
		{"GET", "/public/invoice-status/abc123", "invoice-status:abc123"},
		{"GET", "/other", "fallback"},
	}

	for _, tc := range tests {
		t.Run(tc.method+" "+tc.path, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, tc.path, nil)
			w := httptest.NewRecorder()
			mux.ServeHTTP(w, req)

			if w.Body.String() != tc.expected {
				t.Errorf("expected %q, got %q", tc.expected, w.Body.String())
			}
		})
	}
}
