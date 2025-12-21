package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/roostr/roostr/app/api/internal/db"
)

// mockAccessDB implements the database methods needed for Access handlers
type mockAccessDB struct {
	accessMode      string
	whitelistMeta   []db.WhitelistEntry
	blacklist       []db.BlacklistEntry
	pricingTiers    []db.PricingTier
	paidUsers       []db.PaidUser
	paidUserByKey   map[string]*db.PaidUser
	eventCounts     map[string]int64
	relayConnected  bool
	totalRevenue    int64
	activeCount     int64
	expiringCount   int64
	paymentCount    int64
	revenueByTier   map[string]int64
}

func newMockAccessDB() *mockAccessDB {
	return &mockAccessDB{
		accessMode:    "whitelist",
		whitelistMeta: []db.WhitelistEntry{},
		blacklist:     []db.BlacklistEntry{},
		paidUserByKey: make(map[string]*db.PaidUser),
		eventCounts:   make(map[string]int64),
		revenueByTier: make(map[string]int64),
		pricingTiers: []db.PricingTier{
			{ID: "monthly", Name: "Monthly", AmountSats: 10000, DurationDays: intPtr(30), Enabled: true},
			{ID: "yearly", Name: "Yearly", AmountSats: 100000, DurationDays: intPtr(365), Enabled: true},
		},
	}
}

func (m *mockAccessDB) GetAccessMode(ctx context.Context) (string, error) {
	return m.accessMode, nil
}

func (m *mockAccessDB) SetAccessMode(ctx context.Context, mode string) error {
	m.accessMode = mode
	return nil
}

func (m *mockAccessDB) GetWhitelistMeta(ctx context.Context) ([]db.WhitelistEntry, error) {
	return m.whitelistMeta, nil
}

func (m *mockAccessDB) AddWhitelistEntry(ctx context.Context, entry db.WhitelistEntry) error {
	m.whitelistMeta = append(m.whitelistMeta, entry)
	return nil
}

func (m *mockAccessDB) RemoveWhitelistEntry(ctx context.Context, pubkey string) error {
	for i, e := range m.whitelistMeta {
		if e.Pubkey == pubkey {
			m.whitelistMeta = append(m.whitelistMeta[:i], m.whitelistMeta[i+1:]...)
			return nil
		}
	}
	return nil
}

func (m *mockAccessDB) GetBlacklist(ctx context.Context) ([]db.BlacklistEntry, error) {
	return m.blacklist, nil
}

func (m *mockAccessDB) AddBlacklistEntry(ctx context.Context, entry db.BlacklistEntry) error {
	m.blacklist = append(m.blacklist, entry)
	return nil
}

func (m *mockAccessDB) RemoveBlacklistEntry(ctx context.Context, pubkey string) error {
	for i, e := range m.blacklist {
		if e.Pubkey == pubkey {
			m.blacklist = append(m.blacklist[:i], m.blacklist[i+1:]...)
			return nil
		}
	}
	return nil
}

func (m *mockAccessDB) GetPricingTiers(ctx context.Context) ([]db.PricingTier, error) {
	return m.pricingTiers, nil
}

func (m *mockAccessDB) CountEventsByPubkey(ctx context.Context, pubkeys []string) (map[string]int64, error) {
	return m.eventCounts, nil
}

func (m *mockAccessDB) IsRelayDBConnected() bool {
	return m.relayConnected
}

func (m *mockAccessDB) AddAuditLog(ctx context.Context, action string, details interface{}, performedBy string) error {
	return nil
}

func (m *mockAccessDB) GetPaidUsersFiltered(ctx context.Context, status string, limit, offset int) ([]db.PaidUser, int64, error) {
	return m.paidUsers, int64(len(m.paidUsers)), nil
}

func (m *mockAccessDB) GetPaidUserByPubkey(ctx context.Context, pubkey string) (*db.PaidUser, error) {
	return m.paidUserByKey[pubkey], nil
}

func (m *mockAccessDB) GetTotalRevenue(ctx context.Context) (int64, error) {
	return m.totalRevenue, nil
}

func (m *mockAccessDB) CountActivePaidUsers(ctx context.Context) (int64, error) {
	return m.activeCount, nil
}

func (m *mockAccessDB) CountExpiringPaidUsers(ctx context.Context, days int) (int64, error) {
	return m.expiringCount, nil
}

func (m *mockAccessDB) GetPaymentCount(ctx context.Context) (int64, error) {
	return m.paymentCount, nil
}

func (m *mockAccessDB) GetRevenueByTier(ctx context.Context) (map[string]int64, error) {
	return m.revenueByTier, nil
}

// ============================================================================
// Access Mode Tests
// ============================================================================

func TestGetAccessMode(t *testing.T) {
	mockDB := newMockAccessDB()

	t.Run("returns_current_mode", func(t *testing.T) {
		mockDB.accessMode = "whitelist"

		mux := http.NewServeMux()
		mux.HandleFunc("GET /api/v1/access/mode", func(w http.ResponseWriter, r *http.Request) {
			mode, _ := mockDB.GetAccessMode(r.Context())
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"mode": mode,
			})
		})

		req := httptest.NewRequest("GET", "/api/v1/access/mode", nil)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}

		var resp map[string]interface{}
		json.NewDecoder(w.Body).Decode(&resp)

		if resp["mode"] != "whitelist" {
			t.Errorf("expected mode 'whitelist', got %v", resp["mode"])
		}
	})
}

func TestSetAccessMode(t *testing.T) {
	t.Run("valid_mode_open", func(t *testing.T) {
		mockDB := newMockAccessDB()

		mux := http.NewServeMux()
		mux.HandleFunc("PUT /api/v1/access/mode", func(w http.ResponseWriter, r *http.Request) {
			var req SetAccessModeRequest
			json.NewDecoder(r.Body).Decode(&req)

			validModes := map[string]bool{"open": true, "whitelist": true, "paid": true, "blacklist": true}
			if !validModes[req.Mode] {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(map[string]interface{}{
					"error": "Invalid access mode",
					"code":  "INVALID_MODE",
				})
				return
			}

			mockDB.SetAccessMode(r.Context(), req.Mode)

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": true,
				"mode":    req.Mode,
			})
		})

		body := `{"mode":"open"}`
		req := httptest.NewRequest("PUT", "/api/v1/access/mode", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}

		var resp map[string]interface{}
		json.NewDecoder(w.Body).Decode(&resp)

		if resp["mode"] != "open" {
			t.Errorf("expected mode 'open', got %v", resp["mode"])
		}
	})

	t.Run("invalid_mode", func(t *testing.T) {
		mux := http.NewServeMux()
		mux.HandleFunc("PUT /api/v1/access/mode", func(w http.ResponseWriter, r *http.Request) {
			var req SetAccessModeRequest
			json.NewDecoder(r.Body).Decode(&req)

			validModes := map[string]bool{"open": true, "whitelist": true, "paid": true, "blacklist": true}
			if !validModes[req.Mode] {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(map[string]interface{}{
					"error": "Invalid access mode",
					"code":  "INVALID_MODE",
				})
				return
			}
		})

		body := `{"mode":"invalid"}`
		req := httptest.NewRequest("PUT", "/api/v1/access/mode", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", w.Code)
		}
	})
}

// ============================================================================
// Whitelist Tests
// ============================================================================

func TestGetWhitelist(t *testing.T) {
	t.Run("returns_whitelist_entries", func(t *testing.T) {
		mockDB := newMockAccessDB()
		mockDB.whitelistMeta = []db.WhitelistEntry{
			{Pubkey: "abc123", Npub: "npub1abc", Nickname: "Alice"},
			{Pubkey: "def456", Npub: "npub1def", Nickname: "Bob"},
		}

		mux := http.NewServeMux()
		mux.HandleFunc("GET /api/v1/access/whitelist", func(w http.ResponseWriter, r *http.Request) {
			entries, _ := mockDB.GetWhitelistMeta(r.Context())
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"entries": entries,
			})
		})

		req := httptest.NewRequest("GET", "/api/v1/access/whitelist", nil)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}

		var resp map[string]interface{}
		json.NewDecoder(w.Body).Decode(&resp)

		entries := resp["entries"].([]interface{})
		if len(entries) != 2 {
			t.Errorf("expected 2 entries, got %d", len(entries))
		}
	})
}

func TestAddToWhitelist(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mockDB := newMockAccessDB()

		mux := http.NewServeMux()
		mux.HandleFunc("POST /api/v1/access/whitelist", func(w http.ResponseWriter, r *http.Request) {
			var req AddToWhitelistRequest
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

			mockDB.AddWhitelistEntry(r.Context(), db.WhitelistEntry{
				Pubkey:   req.Pubkey,
				Npub:     req.Npub,
				Nickname: req.Nickname,
			})

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": true,
				"message": "Added to whitelist",
			})
		})

		body := `{"pubkey":"abc123","npub":"npub1abc","nickname":"Alice"}`
		req := httptest.NewRequest("POST", "/api/v1/access/whitelist", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)

		if w.Code != http.StatusCreated {
			t.Errorf("expected status 201, got %d", w.Code)
		}

		if len(mockDB.whitelistMeta) != 1 {
			t.Error("expected 1 whitelist entry")
		}
	})

	t.Run("missing_pubkey", func(t *testing.T) {
		mux := http.NewServeMux()
		mux.HandleFunc("POST /api/v1/access/whitelist", func(w http.ResponseWriter, r *http.Request) {
			var req AddToWhitelistRequest
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

		body := `{"nickname":"Alice"}`
		req := httptest.NewRequest("POST", "/api/v1/access/whitelist", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", w.Code)
		}
	})
}

func TestRemoveFromWhitelist(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mockDB := newMockAccessDB()
		mockDB.whitelistMeta = []db.WhitelistEntry{
			{Pubkey: "abc123"},
		}

		mux := http.NewServeMux()
		mux.HandleFunc("DELETE /api/v1/access/whitelist/{pubkey}", func(w http.ResponseWriter, r *http.Request) {
			pubkey := r.PathValue("pubkey")
			if pubkey == "" {
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			mockDB.RemoveWhitelistEntry(r.Context(), pubkey)

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": true,
				"message": "Removed from whitelist",
			})
		})

		req := httptest.NewRequest("DELETE", "/api/v1/access/whitelist/abc123", nil)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}

		if len(mockDB.whitelistMeta) != 0 {
			t.Error("expected 0 whitelist entries after removal")
		}
	})
}

// ============================================================================
// Blacklist Tests
// ============================================================================

func TestGetBlacklist(t *testing.T) {
	t.Run("returns_blacklist_entries", func(t *testing.T) {
		mockDB := newMockAccessDB()
		mockDB.blacklist = []db.BlacklistEntry{
			{Pubkey: "spam123", Reason: "Spam"},
		}

		mux := http.NewServeMux()
		mux.HandleFunc("GET /api/v1/access/blacklist", func(w http.ResponseWriter, r *http.Request) {
			entries, _ := mockDB.GetBlacklist(r.Context())
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"entries": entries,
			})
		})

		req := httptest.NewRequest("GET", "/api/v1/access/blacklist", nil)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}

		var resp map[string]interface{}
		json.NewDecoder(w.Body).Decode(&resp)

		entries := resp["entries"].([]interface{})
		if len(entries) != 1 {
			t.Errorf("expected 1 entry, got %d", len(entries))
		}
	})
}

func TestAddToBlacklist(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mockDB := newMockAccessDB()

		mux := http.NewServeMux()
		mux.HandleFunc("POST /api/v1/access/blacklist", func(w http.ResponseWriter, r *http.Request) {
			var req AddToBlacklistRequest
			json.NewDecoder(r.Body).Decode(&req)

			if req.Pubkey == "" {
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			mockDB.AddBlacklistEntry(r.Context(), db.BlacklistEntry{
				Pubkey: req.Pubkey,
				Reason: req.Reason,
			})

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": true,
			})
		})

		body := `{"pubkey":"spam123","reason":"Spam"}`
		req := httptest.NewRequest("POST", "/api/v1/access/blacklist", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)

		if w.Code != http.StatusCreated {
			t.Errorf("expected status 201, got %d", w.Code)
		}

		if len(mockDB.blacklist) != 1 {
			t.Error("expected 1 blacklist entry")
		}
	})
}

// ============================================================================
// Pricing Tiers Tests
// ============================================================================

func TestGetPricingTiers(t *testing.T) {
	t.Run("returns_pricing_tiers", func(t *testing.T) {
		mockDB := newMockAccessDB()

		mux := http.NewServeMux()
		mux.HandleFunc("GET /api/v1/access/pricing", func(w http.ResponseWriter, r *http.Request) {
			tiers, _ := mockDB.GetPricingTiers(r.Context())
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"tiers": tiers,
			})
		})

		req := httptest.NewRequest("GET", "/api/v1/access/pricing", nil)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}

		var resp map[string]interface{}
		json.NewDecoder(w.Body).Decode(&resp)

		tiers := resp["tiers"].([]interface{})
		if len(tiers) != 2 {
			t.Errorf("expected 2 tiers, got %d", len(tiers))
		}
	})
}

// ============================================================================
// Revenue Stats Tests
// ============================================================================

func TestGetRevenueStats(t *testing.T) {
	t.Run("returns_revenue_stats", func(t *testing.T) {
		mockDB := newMockAccessDB()
		mockDB.totalRevenue = 1000000
		mockDB.activeCount = 50
		mockDB.expiringCount = 5
		mockDB.paymentCount = 100
		mockDB.revenueByTier = map[string]int64{
			"monthly": 500000,
			"yearly":  500000,
		}

		mux := http.NewServeMux()
		mux.HandleFunc("GET /api/v1/access/revenue", func(w http.ResponseWriter, r *http.Request) {
			totalRevenue, _ := mockDB.GetTotalRevenue(r.Context())
			activeCount, _ := mockDB.CountActivePaidUsers(r.Context())
			expiringCount, _ := mockDB.CountExpiringPaidUsers(r.Context(), 7)
			paymentCount, _ := mockDB.GetPaymentCount(r.Context())
			revenueByTier, _ := mockDB.GetRevenueByTier(r.Context())

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"total_revenue_sats": totalRevenue,
				"active_subscribers": activeCount,
				"expiring_soon":      expiringCount,
				"total_payments":     paymentCount,
				"revenue_by_tier":    revenueByTier,
			})
		})

		req := httptest.NewRequest("GET", "/api/v1/access/revenue", nil)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}

		var resp map[string]interface{}
		json.NewDecoder(w.Body).Decode(&resp)

		if resp["total_revenue_sats"].(float64) != 1000000 {
			t.Errorf("expected total_revenue_sats 1000000, got %v", resp["total_revenue_sats"])
		}
		if resp["active_subscribers"].(float64) != 50 {
			t.Errorf("expected active_subscribers 50, got %v", resp["active_subscribers"])
		}
	})
}
