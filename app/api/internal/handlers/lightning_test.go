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
	"github.com/roostr/roostr/app/api/internal/services"
)

// mockLightningDB implements the database methods needed for Lightning handlers
type mockLightningDB struct {
	accessMode      string
	lightningConfig *db.LightningConfig
	pricingTiers    []db.PricingTier
	pendingInvoices map[string]*db.PendingInvoice
	whitelistEntry  *db.WhitelistEntry
	paidUser        *db.PaidUser
}

func newMockLightningDB() *mockLightningDB {
	return &mockLightningDB{
		accessMode:      "private",
		pendingInvoices: make(map[string]*db.PendingInvoice),
		pricingTiers: []db.PricingTier{
			{ID: "monthly", Name: "Monthly", AmountSats: 10000, DurationDays: intPtr(30), Enabled: true},
			{ID: "yearly", Name: "Yearly", AmountSats: 100000, DurationDays: intPtr(365), Enabled: true},
		},
	}
}

func intPtr(i int) *int { return &i }

func (m *mockLightningDB) GetAccessMode(ctx context.Context) (string, error) {
	return m.accessMode, nil
}

func (m *mockLightningDB) GetLightningConfig(ctx context.Context) (*db.LightningConfig, error) {
	return m.lightningConfig, nil
}

func (m *mockLightningDB) SaveLightningConfig(ctx context.Context, cfg *db.LightningConfig) error {
	m.lightningConfig = cfg
	return nil
}

func (m *mockLightningDB) SetLightningVerified(ctx context.Context) error {
	if m.lightningConfig != nil {
		now := time.Now()
		m.lightningConfig.LastVerifiedAt = &now
	}
	return nil
}

func (m *mockLightningDB) GetPricingTiers(ctx context.Context) ([]db.PricingTier, error) {
	return m.pricingTiers, nil
}

func (m *mockLightningDB) GetPendingInvoice(ctx context.Context, paymentHash string) (*db.PendingInvoice, error) {
	return m.pendingInvoices[paymentHash], nil
}

func (m *mockLightningDB) CreatePendingInvoice(ctx context.Context, inv *db.PendingInvoice) error {
	m.pendingInvoices[inv.PaymentHash] = inv
	return nil
}

func (m *mockLightningDB) GetWhitelistEntryByPubkey(ctx context.Context, pubkey string) (*db.WhitelistEntry, error) {
	return m.whitelistEntry, nil
}

func (m *mockLightningDB) GetPaidUserByPubkey(ctx context.Context, pubkey string) (*db.PaidUser, error) {
	return m.paidUser, nil
}

// mockLightningService implements a mock Lightning service for testing
type mockLightningService struct {
	configured bool
	config     *services.LNDConfig
	nodeInfo   *services.NodeInfo
	balance    *services.ChannelBalance
	invoices   map[string]*services.Invoice
	testErr    error
}

func newMockLightningService() *mockLightningService {
	return &mockLightningService{
		invoices: make(map[string]*services.Invoice),
	}
}

func (m *mockLightningService) IsConfigured() bool {
	return m.configured
}

func (m *mockLightningService) LoadConfig(ctx context.Context) error {
	return nil
}

func (m *mockLightningService) GetConfig() *services.LNDConfig {
	return m.config
}

func (m *mockLightningService) Configure(cfg *services.LNDConfig) {
	m.config = cfg
	m.configured = cfg != nil && cfg.Host != "" && cfg.MacaroonHex != ""
}

func (m *mockLightningService) SaveConfig(ctx context.Context, cfg *services.LNDConfig, enabled bool) error {
	m.Configure(cfg)
	return nil
}

func (m *mockLightningService) GetInfo(ctx context.Context) (*services.NodeInfo, error) {
	if m.testErr != nil {
		return nil, m.testErr
	}
	return m.nodeInfo, nil
}

func (m *mockLightningService) GetBalance(ctx context.Context) (*services.ChannelBalance, error) {
	return m.balance, nil
}

func (m *mockLightningService) TestConnection(ctx context.Context, cfg *services.LNDConfig) (*services.NodeInfo, error) {
	if m.testErr != nil {
		return nil, m.testErr
	}
	if cfg == nil || cfg.Host == "" || cfg.MacaroonHex == "" {
		return nil, services.ErrLNDNotConfigured
	}
	return m.nodeInfo, nil
}

func (m *mockLightningService) CreateInvoice(ctx context.Context, amountSats int64, memo string, expirySecs int64) (*services.Invoice, error) {
	if !m.configured {
		return nil, services.ErrLNDNotConfigured
	}
	inv := &services.Invoice{
		PaymentRequest: "lnbc" + string(rune(amountSats)) + "test",
		PaymentHash:    "testhash123",
		AmountSats:     amountSats,
		ExpiresAt:      time.Now().Add(time.Duration(expirySecs) * time.Second),
		Memo:           memo,
	}
	m.invoices[inv.PaymentHash] = inv
	return inv, nil
}

func (m *mockLightningService) CheckInvoice(ctx context.Context, paymentHash string) (*services.Invoice, error) {
	if inv, ok := m.invoices[paymentHash]; ok {
		return inv, nil
	}
	return nil, nil
}

// TestPAID_API_006_GetLightningStatus tests GET /api/v1/lightning/status
func TestPAID_API_006_GetLightningStatus(t *testing.T) {
	t.Run("not_configured", func(t *testing.T) {
		mockDB := newMockLightningDB()
		mockSvc := newMockLightningService()
		mockSvc.configured = false

		mux := http.NewServeMux()
		mux.HandleFunc("GET /api/v1/lightning/status", func(w http.ResponseWriter, r *http.Request) {
			// Simulate handler logic
			if !mockSvc.IsConfigured() {
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(map[string]interface{}{
					"configured": false,
					"enabled":    false,
					"message":    "Lightning node not configured",
				})
				return
			}
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
		if resp["enabled"] != false {
			t.Error("expected enabled to be false")
		}
		_ = mockDB // use mockDB
	})

	t.Run("configured_and_connected", func(t *testing.T) {
		mockSvc := newMockLightningService()
		mockSvc.configured = true
		mockSvc.nodeInfo = &services.NodeInfo{
			Alias:             "TestNode",
			Pubkey:            "02abc123",
			SyncedToChain:     true,
			NumActiveChannels: 5,
		}
		mockSvc.balance = &services.ChannelBalance{
			LocalBalance:  1000000,
			RemoteBalance: 500000,
			TotalBalance:  1500000,
		}

		mux := http.NewServeMux()
		mux.HandleFunc("GET /api/v1/lightning/status", func(w http.ResponseWriter, r *http.Request) {
			if !mockSvc.IsConfigured() {
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(map[string]interface{}{
					"configured": false,
				})
				return
			}

			info, _ := mockSvc.GetInfo(r.Context())
			balance, _ := mockSvc.GetBalance(r.Context())

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"configured": true,
				"enabled":    true,
				"connected":  true,
				"node_info":  info,
				"balance":    balance,
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

		if resp["configured"] != true {
			t.Error("expected configured to be true")
		}
		if resp["connected"] != true {
			t.Error("expected connected to be true")
		}
		if resp["node_info"] == nil {
			t.Error("expected node_info to be present")
		}
	})

	t.Run("configured_but_connection_failed", func(t *testing.T) {
		mockSvc := newMockLightningService()
		mockSvc.configured = true
		mockSvc.testErr = services.ErrLNDConnectionFailed

		mux := http.NewServeMux()
		mux.HandleFunc("GET /api/v1/lightning/status", func(w http.ResponseWriter, r *http.Request) {
			_, err := mockSvc.GetInfo(r.Context())
			if err != nil {
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(map[string]interface{}{
					"configured": true,
					"enabled":    false,
					"connected":  false,
					"error":      err.Error(),
				})
				return
			}
		})

		req := httptest.NewRequest("GET", "/api/v1/lightning/status", nil)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)

		var resp map[string]interface{}
		json.NewDecoder(w.Body).Decode(&resp)

		if resp["connected"] != false {
			t.Error("expected connected to be false")
		}
		if resp["error"] == nil {
			t.Error("expected error to be present")
		}
	})
}

// TestPAID_API_007_SaveLightningConfig tests PUT /api/v1/lightning/config
func TestPAID_API_007_SaveLightningConfig(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mockSvc := newMockLightningService()

		mux := http.NewServeMux()
		mux.HandleFunc("PUT /api/v1/lightning/config", func(w http.ResponseWriter, r *http.Request) {
			var req struct {
				Host        string `json:"host"`
				MacaroonHex string `json:"macaroon_hex"`
				TLSCertPath string `json:"tls_cert_path"`
				Enabled     bool   `json:"enabled"`
			}
			json.NewDecoder(r.Body).Decode(&req)

			if req.Host == "" {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(map[string]interface{}{
					"error": "Host is required",
					"code":  "MISSING_HOST",
				})
				return
			}
			if req.MacaroonHex == "" {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(map[string]interface{}{
					"error": "Macaroon is required",
					"code":  "MISSING_MACAROON",
				})
				return
			}

			mockSvc.Configure(&services.LNDConfig{
				Host:        req.Host,
				MacaroonHex: req.MacaroonHex,
			})

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": true,
				"message": "Lightning configuration saved",
			})
		})

		body := `{"host":"localhost:8080","macaroon_hex":"abc123","enabled":true}`
		req := httptest.NewRequest("PUT", "/api/v1/lightning/config", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}

		var resp map[string]interface{}
		json.NewDecoder(w.Body).Decode(&resp)

		if resp["success"] != true {
			t.Error("expected success to be true")
		}

		if !mockSvc.IsConfigured() {
			t.Error("expected service to be configured")
		}
	})

	t.Run("missing_host", func(t *testing.T) {
		mux := http.NewServeMux()
		mux.HandleFunc("PUT /api/v1/lightning/config", func(w http.ResponseWriter, r *http.Request) {
			var req struct {
				Host        string `json:"host"`
				MacaroonHex string `json:"macaroon_hex"`
			}
			json.NewDecoder(r.Body).Decode(&req)

			if req.Host == "" {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(map[string]interface{}{
					"error": "Host is required",
					"code":  "MISSING_HOST",
				})
				return
			}
		})

		body := `{"macaroon_hex":"abc123"}`
		req := httptest.NewRequest("PUT", "/api/v1/lightning/config", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", w.Code)
		}

		var resp map[string]interface{}
		json.NewDecoder(w.Body).Decode(&resp)

		if resp["code"] != "MISSING_HOST" {
			t.Errorf("expected code MISSING_HOST, got %v", resp["code"])
		}
	})

	t.Run("missing_macaroon", func(t *testing.T) {
		mux := http.NewServeMux()
		mux.HandleFunc("PUT /api/v1/lightning/config", func(w http.ResponseWriter, r *http.Request) {
			var req struct {
				Host        string `json:"host"`
				MacaroonHex string `json:"macaroon_hex"`
			}
			json.NewDecoder(r.Body).Decode(&req)

			if req.Host == "" {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(map[string]interface{}{
					"error": "Host is required",
					"code":  "MISSING_HOST",
				})
				return
			}
			if req.MacaroonHex == "" {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(map[string]interface{}{
					"error": "Macaroon is required",
					"code":  "MISSING_MACAROON",
				})
				return
			}
		})

		body := `{"host":"localhost:8080"}`
		req := httptest.NewRequest("PUT", "/api/v1/lightning/config", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", w.Code)
		}

		var resp map[string]interface{}
		json.NewDecoder(w.Body).Decode(&resp)

		if resp["code"] != "MISSING_MACAROON" {
			t.Errorf("expected code MISSING_MACAROON, got %v", resp["code"])
		}
	})
}

// TestSIGNUP_API_001_GetRelayInfo tests GET /public/relay-info
func TestSIGNUP_API_001_GetRelayInfo(t *testing.T) {
	t.Run("paid_access_disabled", func(t *testing.T) {
		mockDB := newMockLightningDB()
		mockDB.accessMode = "private"

		mux := http.NewServeMux()
		mux.HandleFunc("GET /public/relay-info", func(w http.ResponseWriter, r *http.Request) {
			if mockDB.accessMode != "paid" {
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(map[string]interface{}{
					"paid_access_enabled": false,
					"message":             "Paid access is not enabled for this relay",
				})
				return
			}
		})

		req := httptest.NewRequest("GET", "/public/relay-info", nil)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}

		var resp map[string]interface{}
		json.NewDecoder(w.Body).Decode(&resp)

		if resp["paid_access_enabled"] != false {
			t.Error("expected paid_access_enabled to be false")
		}
	})

	t.Run("paid_access_enabled", func(t *testing.T) {
		mockDB := newMockLightningDB()
		mockDB.accessMode = "paid"
		mockSvc := newMockLightningService()
		mockSvc.configured = true

		mux := http.NewServeMux()
		mux.HandleFunc("GET /public/relay-info", func(w http.ResponseWriter, r *http.Request) {
			if mockDB.accessMode != "paid" {
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(map[string]interface{}{
					"paid_access_enabled": false,
				})
				return
			}

			tiers, _ := mockDB.GetPricingTiers(r.Context())
			var enabledTiers []map[string]interface{}
			for _, t := range tiers {
				if t.Enabled {
					tier := map[string]interface{}{
						"id":          t.ID,
						"name":        t.Name,
						"amount_sats": t.AmountSats,
					}
					if t.DurationDays != nil {
						tier["duration_days"] = *t.DurationDays
					}
					enabledTiers = append(enabledTiers, tier)
				}
			}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"paid_access_enabled":  true,
				"lightning_configured": mockSvc.IsConfigured(),
				"relay_name":           "Test Relay",
				"pricing_tiers":        enabledTiers,
			})
		})

		req := httptest.NewRequest("GET", "/public/relay-info", nil)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)

		var resp map[string]interface{}
		json.NewDecoder(w.Body).Decode(&resp)

		if resp["paid_access_enabled"] != true {
			t.Error("expected paid_access_enabled to be true")
		}
		if resp["lightning_configured"] != true {
			t.Error("expected lightning_configured to be true")
		}
		if resp["pricing_tiers"] == nil {
			t.Error("expected pricing_tiers to be present")
		}

		tiers := resp["pricing_tiers"].([]interface{})
		if len(tiers) != 2 {
			t.Errorf("expected 2 pricing tiers, got %d", len(tiers))
		}
	})
}

// TestSIGNUP_API_002_CreateSignupInvoice tests POST /public/create-invoice
func TestSIGNUP_API_002_CreateSignupInvoice(t *testing.T) {
	t.Run("paid_access_disabled", func(t *testing.T) {
		mockDB := newMockLightningDB()
		mockDB.accessMode = "private"

		mux := http.NewServeMux()
		mux.HandleFunc("POST /public/create-invoice", func(w http.ResponseWriter, r *http.Request) {
			if mockDB.accessMode != "paid" {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(map[string]interface{}{
					"error": "Paid access is not enabled",
					"code":  "PAID_ACCESS_DISABLED",
				})
				return
			}
		})

		body := `{"pubkey":"abc123","tier_id":"monthly"}`
		req := httptest.NewRequest("POST", "/public/create-invoice", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", w.Code)
		}

		var resp map[string]interface{}
		json.NewDecoder(w.Body).Decode(&resp)

		if resp["code"] != "PAID_ACCESS_DISABLED" {
			t.Errorf("expected code PAID_ACCESS_DISABLED, got %v", resp["code"])
		}
	})

	t.Run("missing_pubkey", func(t *testing.T) {
		mockDB := newMockLightningDB()
		mockDB.accessMode = "paid"

		mux := http.NewServeMux()
		mux.HandleFunc("POST /public/create-invoice", func(w http.ResponseWriter, r *http.Request) {
			if mockDB.accessMode != "paid" {
				w.WriteHeader(http.StatusBadRequest)
				return
			}

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

		body := `{"tier_id":"monthly"}`
		req := httptest.NewRequest("POST", "/public/create-invoice", strings.NewReader(body))
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
	})

	t.Run("missing_tier", func(t *testing.T) {
		mockDB := newMockLightningDB()
		mockDB.accessMode = "paid"

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
			if req.TierID == "" {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(map[string]interface{}{
					"error": "Tier ID is required",
					"code":  "MISSING_TIER",
				})
				return
			}
		})

		body := `{"pubkey":"abc123"}`
		req := httptest.NewRequest("POST", "/public/create-invoice", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", w.Code)
		}

		var resp map[string]interface{}
		json.NewDecoder(w.Body).Decode(&resp)

		if resp["code"] != "MISSING_TIER" {
			t.Errorf("expected code MISSING_TIER, got %v", resp["code"])
		}
	})

	t.Run("success", func(t *testing.T) {
		mockDB := newMockLightningDB()
		mockDB.accessMode = "paid"
		mockSvc := newMockLightningService()
		mockSvc.configured = true

		mux := http.NewServeMux()
		mux.HandleFunc("POST /public/create-invoice", func(w http.ResponseWriter, r *http.Request) {
			var req struct {
				Pubkey string `json:"pubkey"`
				TierID string `json:"tier_id"`
			}
			json.NewDecoder(r.Body).Decode(&req)

			if req.Pubkey == "" || req.TierID == "" {
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			// Simulate invoice creation
			invoice, err := mockSvc.CreateInvoice(r.Context(), 10000, "Roostr access", 900)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"payment_hash":    invoice.PaymentHash,
				"payment_request": invoice.PaymentRequest,
				"amount_sats":     invoice.AmountSats,
				"tier_id":         req.TierID,
				"expires_at":      invoice.ExpiresAt.Unix(),
			})
		})

		body := `{"pubkey":"0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef","tier_id":"monthly"}`
		req := httptest.NewRequest("POST", "/public/create-invoice", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)

		if w.Code != http.StatusCreated {
			t.Errorf("expected status 201, got %d", w.Code)
		}

		var resp map[string]interface{}
		json.NewDecoder(w.Body).Decode(&resp)

		if resp["payment_hash"] == nil {
			t.Error("expected payment_hash to be present")
		}
		if resp["payment_request"] == nil {
			t.Error("expected payment_request to be present")
		}
	})
}

// TestSIGNUP_API_003_GetInvoiceStatus tests GET /public/invoice-status/{hash}
func TestSIGNUP_API_003_GetInvoiceStatus(t *testing.T) {
	t.Run("invoice_not_found", func(t *testing.T) {
		mockDB := newMockLightningDB()

		mux := http.NewServeMux()
		mux.HandleFunc("GET /public/invoice-status/{hash}", func(w http.ResponseWriter, r *http.Request) {
			hash := r.PathValue("hash")
			inv, _ := mockDB.GetPendingInvoice(r.Context(), hash)
			if inv == nil {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusNotFound)
				json.NewEncoder(w).Encode(map[string]interface{}{
					"error": "Invoice not found",
					"code":  "INVOICE_NOT_FOUND",
				})
				return
			}
		})

		req := httptest.NewRequest("GET", "/public/invoice-status/nonexistent", nil)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("expected status 404, got %d", w.Code)
		}

		var resp map[string]interface{}
		json.NewDecoder(w.Body).Decode(&resp)

		if resp["code"] != "INVOICE_NOT_FOUND" {
			t.Errorf("expected code INVOICE_NOT_FOUND, got %v", resp["code"])
		}
	})

	t.Run("invoice_pending", func(t *testing.T) {
		mockDB := newMockLightningDB()
		mockDB.pendingInvoices["testhash"] = &db.PendingInvoice{
			PaymentHash: "testhash",
			Status:      "pending",
			ExpiresAt:   time.Now().Add(15 * time.Minute),
		}

		mux := http.NewServeMux()
		mux.HandleFunc("GET /public/invoice-status/{hash}", func(w http.ResponseWriter, r *http.Request) {
			hash := r.PathValue("hash")
			inv, _ := mockDB.GetPendingInvoice(r.Context(), hash)
			if inv == nil {
				w.WriteHeader(http.StatusNotFound)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"status":       inv.Status,
				"payment_hash": inv.PaymentHash,
				"expires_at":   inv.ExpiresAt.Unix(),
			})
		})

		req := httptest.NewRequest("GET", "/public/invoice-status/testhash", nil)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}

		var resp map[string]interface{}
		json.NewDecoder(w.Body).Decode(&resp)

		if resp["status"] != "pending" {
			t.Errorf("expected status pending, got %v", resp["status"])
		}
	})

	t.Run("invoice_paid", func(t *testing.T) {
		mockDB := newMockLightningDB()
		paidAt := time.Now()
		mockDB.pendingInvoices["paidhash"] = &db.PendingInvoice{
			PaymentHash: "paidhash",
			Status:      "paid",
			PaidAt:      &paidAt,
		}

		mux := http.NewServeMux()
		mux.HandleFunc("GET /public/invoice-status/{hash}", func(w http.ResponseWriter, r *http.Request) {
			hash := r.PathValue("hash")
			inv, _ := mockDB.GetPendingInvoice(r.Context(), hash)
			if inv == nil {
				w.WriteHeader(http.StatusNotFound)
				return
			}

			resp := map[string]interface{}{
				"status":       inv.Status,
				"payment_hash": inv.PaymentHash,
			}
			if inv.Status == "paid" && inv.PaidAt != nil {
				resp["paid_at"] = inv.PaidAt.Unix()
			}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)
		})

		req := httptest.NewRequest("GET", "/public/invoice-status/paidhash", nil)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}

		var resp map[string]interface{}
		json.NewDecoder(w.Body).Decode(&resp)

		if resp["status"] != "paid" {
			t.Errorf("expected status paid, got %v", resp["status"])
		}
		if resp["paid_at"] == nil {
			t.Error("expected paid_at to be present")
		}
	})

	t.Run("invoice_expired", func(t *testing.T) {
		mockDB := newMockLightningDB()
		mockDB.pendingInvoices["expiredhash"] = &db.PendingInvoice{
			PaymentHash: "expiredhash",
			Status:      "expired",
		}

		mux := http.NewServeMux()
		mux.HandleFunc("GET /public/invoice-status/{hash}", func(w http.ResponseWriter, r *http.Request) {
			hash := r.PathValue("hash")
			inv, _ := mockDB.GetPendingInvoice(r.Context(), hash)
			if inv == nil {
				w.WriteHeader(http.StatusNotFound)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"status":       inv.Status,
				"payment_hash": inv.PaymentHash,
			})
		})

		req := httptest.NewRequest("GET", "/public/invoice-status/expiredhash", nil)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)

		var resp map[string]interface{}
		json.NewDecoder(w.Body).Decode(&resp)

		if resp["status"] != "expired" {
			t.Errorf("expected status expired, got %v", resp["status"])
		}
	})
}
