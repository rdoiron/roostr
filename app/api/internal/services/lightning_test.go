package services

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// TestLN001_LightningServiceBasics tests basic LightningService functionality (LN-001)
func TestLN001_LightningServiceBasics(t *testing.T) {
	t.Run("NewLightningService", func(t *testing.T) {
		svc := NewLightningService(nil)
		if svc == nil {
			t.Fatal("expected service to be created")
		}
		if svc.client == nil {
			t.Error("expected HTTP client to be initialized")
		}
	})

	t.Run("IsConfigured_false_when_no_config", func(t *testing.T) {
		svc := NewLightningService(nil)
		if svc.IsConfigured() {
			t.Error("expected IsConfigured to be false with no config")
		}
	})

	t.Run("Configure_and_IsConfigured", func(t *testing.T) {
		svc := NewLightningService(nil)
		cfg := &LNDConfig{
			Host:        "localhost:8080",
			MacaroonHex: "0201036c6e6402f801030a",
		}
		svc.Configure(cfg)

		if !svc.IsConfigured() {
			t.Error("expected IsConfigured to be true after Configure")
		}
	})

	t.Run("IsConfigured_false_with_empty_host", func(t *testing.T) {
		svc := NewLightningService(nil)
		cfg := &LNDConfig{
			Host:        "",
			MacaroonHex: "0201036c6e6402f801030a",
		}
		svc.Configure(cfg)

		if svc.IsConfigured() {
			t.Error("expected IsConfigured to be false with empty host")
		}
	})

	t.Run("IsConfigured_false_with_empty_macaroon", func(t *testing.T) {
		svc := NewLightningService(nil)
		cfg := &LNDConfig{
			Host:        "localhost:8080",
			MacaroonHex: "",
		}
		svc.Configure(cfg)

		if svc.IsConfigured() {
			t.Error("expected IsConfigured to be false with empty macaroon")
		}
	})

	t.Run("GetConfig_returns_copy", func(t *testing.T) {
		svc := NewLightningService(nil)
		cfg := &LNDConfig{
			Host:        "localhost:8080",
			MacaroonHex: "abc123",
			TLSCertPath: "/path/to/cert",
		}
		svc.Configure(cfg)

		got := svc.GetConfig()
		if got == nil {
			t.Fatal("expected GetConfig to return config")
		}
		if got.Host != cfg.Host {
			t.Errorf("expected Host %s, got %s", cfg.Host, got.Host)
		}
		if got.MacaroonHex != cfg.MacaroonHex {
			t.Errorf("expected MacaroonHex %s, got %s", cfg.MacaroonHex, got.MacaroonHex)
		}
		if got.TLSCertPath != cfg.TLSCertPath {
			t.Errorf("expected TLSCertPath %s, got %s", cfg.TLSCertPath, got.TLSCertPath)
		}
	})

	t.Run("GetConfig_returns_nil_when_not_configured", func(t *testing.T) {
		svc := NewLightningService(nil)
		got := svc.GetConfig()
		if got != nil {
			t.Error("expected GetConfig to return nil when not configured")
		}
	})
}

// TestLN001_LNDRESTClient tests LND REST API client functionality (LN-001)
func TestLN001_LNDRESTClient(t *testing.T) {
	t.Run("GetInfo_returns_error_when_not_configured", func(t *testing.T) {
		svc := NewLightningService(nil)
		_, err := svc.GetInfo(context.Background())
		if err != ErrLNDNotConfigured {
			t.Errorf("expected ErrLNDNotConfigured, got %v", err)
		}
	})

	t.Run("GetBalance_returns_error_when_not_configured", func(t *testing.T) {
		svc := NewLightningService(nil)
		_, err := svc.GetBalance(context.Background())
		if err != ErrLNDNotConfigured {
			t.Errorf("expected ErrLNDNotConfigured, got %v", err)
		}
	})

	t.Run("CreateInvoice_returns_error_when_not_configured", func(t *testing.T) {
		svc := NewLightningService(nil)
		_, err := svc.CreateInvoice(context.Background(), 1000, "test", 900)
		if err != ErrLNDNotConfigured {
			t.Errorf("expected ErrLNDNotConfigured, got %v", err)
		}
	})

	t.Run("CheckInvoice_returns_error_when_not_configured", func(t *testing.T) {
		svc := NewLightningService(nil)
		_, err := svc.CheckInvoice(context.Background(), "abc123")
		if err != ErrLNDNotConfigured {
			t.Errorf("expected ErrLNDNotConfigured, got %v", err)
		}
	})

	t.Run("TestConnection_returns_error_with_nil_config", func(t *testing.T) {
		svc := NewLightningService(nil)
		_, err := svc.TestConnection(context.Background(), nil)
		if err != ErrLNDNotConfigured {
			t.Errorf("expected ErrLNDNotConfigured, got %v", err)
		}
	})

	t.Run("TestConnection_returns_error_with_empty_config", func(t *testing.T) {
		svc := NewLightningService(nil)
		_, err := svc.TestConnection(context.Background(), &LNDConfig{})
		if err != ErrLNDNotConfigured {
			t.Errorf("expected ErrLNDNotConfigured, got %v", err)
		}
	})
}

// TestLN001_MockLNDServer tests LND REST API client with a mock server (LN-001)
func TestLN001_MockLNDServer(t *testing.T) {
	t.Run("GetInfo_success", func(t *testing.T) {
		server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/v1/getinfo" {
				t.Errorf("unexpected path: %s", r.URL.Path)
			}
			if r.Header.Get("Grpc-Metadata-macaroon") != "testmacaroon" {
				t.Error("missing or incorrect macaroon header")
			}
			json.NewEncoder(w).Encode(map[string]interface{}{
				"alias":               "TestNode",
				"identity_pubkey":     "02abc123",
				"version":             "0.17.0",
				"synced_to_chain":     true,
				"synced_to_graph":     true,
				"num_active_channels": 5,
				"num_peers":           10,
				"block_height":        800000,
			})
		}))
		defer server.Close()

		svc := &LightningService{
			client: server.Client(),
		}

		host := strings.TrimPrefix(server.URL, "https://")
		cfg := &LNDConfig{
			Host:        host,
			MacaroonHex: "testmacaroon",
		}

		info, err := svc.TestConnection(context.Background(), cfg)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if info.Alias != "TestNode" {
			t.Errorf("expected alias TestNode, got %s", info.Alias)
		}
		if info.Pubkey != "02abc123" {
			t.Errorf("expected pubkey 02abc123, got %s", info.Pubkey)
		}
		if !info.SyncedToChain {
			t.Error("expected SyncedToChain to be true")
		}
		if info.NumActiveChannels != 5 {
			t.Errorf("expected 5 channels, got %d", info.NumActiveChannels)
		}
	})

	t.Run("GetInfo_unauthorized", func(t *testing.T) {
		server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusUnauthorized)
		}))
		defer server.Close()

		svc := &LightningService{
			client: server.Client(),
		}

		host := strings.TrimPrefix(server.URL, "https://")
		cfg := &LNDConfig{
			Host:        host,
			MacaroonHex: "badmacaroon",
		}

		_, err := svc.TestConnection(context.Background(), cfg)
		if err != ErrLNDAuthFailed {
			t.Errorf("expected ErrLNDAuthFailed, got %v", err)
		}
	})

	t.Run("GetBalance_success", func(t *testing.T) {
		server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/v1/balance/channels" {
				t.Errorf("unexpected path: %s", r.URL.Path)
			}
			json.NewEncoder(w).Encode(map[string]interface{}{
				"local_balance":  "1000000",
				"remote_balance": "500000",
			})
		}))
		defer server.Close()

		svc := &LightningService{
			client: server.Client(),
			config: &LNDConfig{
				Host:        strings.TrimPrefix(server.URL, "https://"),
				MacaroonHex: "testmacaroon",
			},
		}

		balance, err := svc.GetBalance(context.Background())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if balance.LocalBalance != 1000000 {
			t.Errorf("expected local balance 1000000, got %d", balance.LocalBalance)
		}
		if balance.RemoteBalance != 500000 {
			t.Errorf("expected remote balance 500000, got %d", balance.RemoteBalance)
		}
		if balance.TotalBalance != 1500000 {
			t.Errorf("expected total balance 1500000, got %d", balance.TotalBalance)
		}
	})

	t.Run("CreateInvoice_success", func(t *testing.T) {
		// Create a test payment hash (32 bytes)
		paymentHash := make([]byte, 32)
		for i := range paymentHash {
			paymentHash[i] = byte(i)
		}

		server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/v1/invoices" {
				t.Errorf("unexpected path: %s", r.URL.Path)
			}
			if r.Method != "POST" {
				t.Errorf("expected POST, got %s", r.Method)
			}

			var req map[string]interface{}
			json.NewDecoder(r.Body).Decode(&req)

			if req["value"].(float64) != 1000 {
				t.Errorf("expected value 1000, got %v", req["value"])
			}
			if req["memo"].(string) != "Test invoice" {
				t.Errorf("expected memo 'Test invoice', got %v", req["memo"])
			}

			// LND returns r_hash as base64
			json.NewEncoder(w).Encode(map[string]interface{}{
				"r_hash":          "AAECAwQFBgcICQoLDA0ODxAREhMUFRYXGBkaGxwdHh8=", // base64 of sequential bytes
				"payment_request": "lnbc10u1p0test",
			})
		}))
		defer server.Close()

		svc := &LightningService{
			client: server.Client(),
			config: &LNDConfig{
				Host:        strings.TrimPrefix(server.URL, "https://"),
				MacaroonHex: "testmacaroon",
			},
		}

		invoice, err := svc.CreateInvoice(context.Background(), 1000, "Test invoice", 900)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if invoice.PaymentRequest != "lnbc10u1p0test" {
			t.Errorf("expected payment request lnbc10u1p0test, got %s", invoice.PaymentRequest)
		}
		expectedHash := hex.EncodeToString(paymentHash)
		if invoice.PaymentHash != expectedHash {
			t.Errorf("expected payment hash %s, got %s", expectedHash, invoice.PaymentHash)
		}
		if invoice.AmountSats != 1000 {
			t.Errorf("expected amount 1000, got %d", invoice.AmountSats)
		}
		if invoice.Memo != "Test invoice" {
			t.Errorf("expected memo 'Test invoice', got %s", invoice.Memo)
		}
		if invoice.Settled {
			t.Error("expected invoice to not be settled")
		}

		// Check expiry is roughly 15 minutes from now
		expectedExpiry := time.Now().Add(900 * time.Second)
		if invoice.ExpiresAt.Before(expectedExpiry.Add(-10*time.Second)) || invoice.ExpiresAt.After(expectedExpiry.Add(10*time.Second)) {
			t.Errorf("expiry time not in expected range")
		}
	})
}


// TestLN003_CreateAccessInvoice tests access invoice creation logic (LN-003)
func TestLN003_CreateAccessInvoice(t *testing.T) {
	t.Run("CreateAccessInvoice_returns_error_when_not_configured", func(t *testing.T) {
		svc := NewLightningService(nil)
		_, err := svc.CreateAccessInvoice(context.Background(), AccessInvoiceRequest{
			Pubkey: "abc123",
			Npub:   "npub1test",
			TierID: "tier1",
		})
		if err != ErrLNDNotConfigured {
			t.Errorf("expected ErrLNDNotConfigured, got %v", err)
		}
	})
}

// TestLN001_Types tests type definitions (LN-001)
func TestLN001_Types(t *testing.T) {
	t.Run("LNDConfig_JSON", func(t *testing.T) {
		cfg := &LNDConfig{
			Host:        "localhost:8080",
			MacaroonHex: "abc123",
			TLSCertPath: "/path/to/cert",
		}

		data, err := json.Marshal(cfg)
		if err != nil {
			t.Fatalf("failed to marshal: %v", err)
		}

		var decoded LNDConfig
		if err := json.Unmarshal(data, &decoded); err != nil {
			t.Fatalf("failed to unmarshal: %v", err)
		}

		if decoded.Host != cfg.Host {
			t.Errorf("expected host %s, got %s", cfg.Host, decoded.Host)
		}
	})

	t.Run("NodeInfo_JSON", func(t *testing.T) {
		info := &NodeInfo{
			Alias:             "TestNode",
			Pubkey:            "02abc",
			Version:           "0.17.0",
			SyncedToChain:     true,
			NumActiveChannels: 5,
		}

		data, err := json.Marshal(info)
		if err != nil {
			t.Fatalf("failed to marshal: %v", err)
		}

		var decoded NodeInfo
		if err := json.Unmarshal(data, &decoded); err != nil {
			t.Fatalf("failed to unmarshal: %v", err)
		}

		if decoded.Alias != info.Alias {
			t.Errorf("expected alias %s, got %s", info.Alias, decoded.Alias)
		}
	})

	t.Run("Invoice_JSON", func(t *testing.T) {
		now := time.Now()
		invoice := &Invoice{
			PaymentRequest: "lnbc10u1p0test",
			PaymentHash:    "abc123",
			AmountSats:     1000,
			ExpiresAt:      now,
			Memo:           "Test",
			Settled:        true,
			SettledAt:      &now,
		}

		data, err := json.Marshal(invoice)
		if err != nil {
			t.Fatalf("failed to marshal: %v", err)
		}

		var decoded Invoice
		if err := json.Unmarshal(data, &decoded); err != nil {
			t.Fatalf("failed to unmarshal: %v", err)
		}

		if decoded.PaymentRequest != invoice.PaymentRequest {
			t.Errorf("expected payment request %s, got %s", invoice.PaymentRequest, decoded.PaymentRequest)
		}
		if decoded.Settled != invoice.Settled {
			t.Errorf("expected settled %v, got %v", invoice.Settled, decoded.Settled)
		}
	})
}

// TestLN001_Errors tests error definitions (LN-001)
func TestLN001_Errors(t *testing.T) {
	tests := []struct {
		err     error
		message string
	}{
		{ErrLNDNotConfigured, "LND is not configured"},
		{ErrLNDConnectionFailed, "failed to connect to LND"},
		{ErrLNDAuthFailed, "LND authentication failed"},
		{ErrLNDNotSynced, "LND node is not synced to chain"},
	}

	for _, tc := range tests {
		t.Run(tc.message, func(t *testing.T) {
			if tc.err.Error() != tc.message {
				t.Errorf("expected error message %q, got %q", tc.message, tc.err.Error())
			}
		})
	}
}
