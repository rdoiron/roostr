package services

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/roostr/roostr/app/api/internal/db"
)

// Lightning errors
var (
	ErrLNDNotConfigured    = errors.New("LND is not configured")
	ErrLNDConnectionFailed = errors.New("failed to connect to LND")
	ErrLNDAuthFailed       = errors.New("LND authentication failed")
	ErrLNDNotSynced        = errors.New("LND node is not synced to chain")
	ErrLNDNotDetected      = errors.New("LND node not detected")
)

// LNDConfig holds the configuration for connecting to an LND node.
type LNDConfig struct {
	Host        string `json:"host"`         // e.g., "umbrel.local:8080"
	MacaroonHex string `json:"macaroon_hex"` // admin.macaroon as hex
	TLSCertPath string `json:"tls_cert_path,omitempty"`
}

// NodeInfo contains information about the Lightning node.
type NodeInfo struct {
	Alias           string `json:"alias"`
	Pubkey          string `json:"pubkey"`
	Version         string `json:"version"`
	SyncedToChain   bool   `json:"synced_to_chain"`
	SyncedToGraph   bool   `json:"synced_to_graph"`
	NumActiveChannels int  `json:"num_active_channels"`
	NumPeers        int    `json:"num_peers"`
	BlockHeight     int64  `json:"block_height"`
}

// ChannelBalance contains the node's channel balance information.
type ChannelBalance struct {
	LocalBalance  int64 `json:"local_balance"`
	RemoteBalance int64 `json:"remote_balance"`
	TotalBalance  int64 `json:"total_balance"`
}

// Invoice represents a Lightning invoice.
type Invoice struct {
	PaymentRequest string    `json:"payment_request"` // BOLT11 invoice string
	PaymentHash    string    `json:"payment_hash"`    // hex-encoded
	AmountSats     int64     `json:"amount_sats"`
	ExpiresAt      time.Time `json:"expires_at"`
	Memo           string    `json:"memo"`
	Settled        bool      `json:"settled"`
	SettledAt      *time.Time `json:"settled_at,omitempty"`
}

// LightningService handles Lightning Network operations via LND.
type LightningService struct {
	db     *db.DB
	mu     sync.RWMutex
	client *http.Client
	config *LNDConfig
}

// NewLightningService creates a new Lightning service.
func NewLightningService(database *db.DB) *LightningService {
	// Create HTTP client with TLS config that skips verification
	// (common for local Umbrel/Start9 setups)
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}

	return &LightningService{
		db: database,
		client: &http.Client{
			Transport: tr,
			Timeout:   30 * time.Second,
		},
	}
}

// Configure sets the LND configuration.
func (s *LightningService) Configure(cfg *LNDConfig) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.config = cfg
}

// GetConfig returns the current LND configuration.
func (s *LightningService) GetConfig() *LNDConfig {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.config == nil {
		return nil
	}
	// Return a copy
	return &LNDConfig{
		Host:        s.config.Host,
		MacaroonHex: s.config.MacaroonHex,
		TLSCertPath: s.config.TLSCertPath,
	}
}

// IsConfigured returns true if LND is configured.
func (s *LightningService) IsConfigured() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.config != nil && s.config.Host != "" && s.config.MacaroonHex != ""
}

// LoadConfig loads the configuration from the database.
func (s *LightningService) LoadConfig(ctx context.Context) error {
	cfg, err := s.db.GetLightningConfig(ctx)
	if err != nil {
		return fmt.Errorf("failed to load lightning config: %w", err)
	}
	if cfg == nil || cfg.Endpoint == "" {
		return nil // No config stored yet
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	s.config = &LNDConfig{
		Host:        cfg.Endpoint,
		MacaroonHex: cfg.Macaroon,
		TLSCertPath: cfg.Cert,
	}
	return nil
}

// SaveConfig saves the configuration to the database.
func (s *LightningService) SaveConfig(ctx context.Context, cfg *LNDConfig, enabled bool) error {
	dbCfg := &db.LightningConfig{
		NodeType: "lnd",
		Endpoint: cfg.Host,
		Macaroon: cfg.MacaroonHex,
		Cert:     cfg.TLSCertPath,
		Enabled:  enabled,
	}

	if err := s.db.SaveLightningConfig(ctx, dbCfg); err != nil {
		return fmt.Errorf("failed to save lightning config: %w", err)
	}

	s.mu.Lock()
	s.config = cfg
	s.mu.Unlock()

	return nil
}

// TestConnection tests the connection to LND with the provided config.
// Returns node info if successful.
func (s *LightningService) TestConnection(ctx context.Context, cfg *LNDConfig) (*NodeInfo, error) {
	if cfg == nil || cfg.Host == "" || cfg.MacaroonHex == "" {
		return nil, ErrLNDNotConfigured
	}

	// Use provided config for this test
	info, err := s.getInfoWithConfig(ctx, cfg)
	if err != nil {
		return nil, err
	}

	return info, nil
}

// GetInfo returns information about the connected LND node.
func (s *LightningService) GetInfo(ctx context.Context) (*NodeInfo, error) {
	cfg := s.GetConfig()
	if cfg == nil {
		return nil, ErrLNDNotConfigured
	}
	return s.getInfoWithConfig(ctx, cfg)
}

func (s *LightningService) getInfoWithConfig(ctx context.Context, cfg *LNDConfig) (*NodeInfo, error) {
	resp, err := s.doRequest(ctx, cfg, "GET", "/v1/getinfo", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return nil, ErrLNDAuthFailed
	}
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("%w: %s", ErrLNDConnectionFailed, string(body))
	}

	var result struct {
		Alias             string `json:"alias"`
		IdentityPubkey    string `json:"identity_pubkey"`
		Version           string `json:"version"`
		SyncedToChain     bool   `json:"synced_to_chain"`
		SyncedToGraph     bool   `json:"synced_to_graph"`
		NumActiveChannels int    `json:"num_active_channels"`
		NumPeers          int    `json:"num_peers"`
		BlockHeight       int64  `json:"block_height"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode LND response: %w", err)
	}

	return &NodeInfo{
		Alias:             result.Alias,
		Pubkey:            result.IdentityPubkey,
		Version:           result.Version,
		SyncedToChain:     result.SyncedToChain,
		SyncedToGraph:     result.SyncedToGraph,
		NumActiveChannels: result.NumActiveChannels,
		NumPeers:          result.NumPeers,
		BlockHeight:       result.BlockHeight,
	}, nil
}

// GetBalance returns the channel balance of the LND node.
func (s *LightningService) GetBalance(ctx context.Context) (*ChannelBalance, error) {
	cfg := s.GetConfig()
	if cfg == nil {
		return nil, ErrLNDNotConfigured
	}

	resp, err := s.doRequest(ctx, cfg, "GET", "/v1/balance/channels", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("%w: %s", ErrLNDConnectionFailed, string(body))
	}

	var result struct {
		LocalBalance  string `json:"local_balance"`
		RemoteBalance string `json:"remote_balance"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode balance response: %w", err)
	}

	var localBal, remoteBal int64
	fmt.Sscanf(result.LocalBalance, "%d", &localBal)
	fmt.Sscanf(result.RemoteBalance, "%d", &remoteBal)

	return &ChannelBalance{
		LocalBalance:  localBal,
		RemoteBalance: remoteBal,
		TotalBalance:  localBal + remoteBal,
	}, nil
}

// CreateInvoice generates a Lightning invoice.
func (s *LightningService) CreateInvoice(ctx context.Context, amountSats int64, memo string, expirySecs int64) (*Invoice, error) {
	cfg := s.GetConfig()
	if cfg == nil {
		return nil, ErrLNDNotConfigured
	}

	if expirySecs <= 0 {
		expirySecs = 900 // Default 15 minutes
	}

	reqBody := map[string]interface{}{
		"value":  amountSats,
		"memo":   memo,
		"expiry": expirySecs,
	}

	resp, err := s.doRequest(ctx, cfg, "POST", "/v1/invoices", reqBody)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to create invoice: %s", string(body))
	}

	var result struct {
		RHash          string `json:"r_hash"` // base64 encoded
		PaymentRequest string `json:"payment_request"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode invoice response: %w", err)
	}

	// Convert r_hash from base64 to hex
	rHashBytes, err := base64.StdEncoding.DecodeString(result.RHash)
	if err != nil {
		return nil, fmt.Errorf("failed to decode payment hash: %w", err)
	}
	paymentHash := hex.EncodeToString(rHashBytes)

	return &Invoice{
		PaymentRequest: result.PaymentRequest,
		PaymentHash:    paymentHash,
		AmountSats:     amountSats,
		ExpiresAt:      time.Now().Add(time.Duration(expirySecs) * time.Second),
		Memo:           memo,
		Settled:        false,
	}, nil
}

// CheckInvoice checks the status of an invoice by payment hash.
func (s *LightningService) CheckInvoice(ctx context.Context, paymentHash string) (*Invoice, error) {
	cfg := s.GetConfig()
	if cfg == nil {
		return nil, ErrLNDNotConfigured
	}

	// LND expects the payment hash as URL-safe base64
	hashBytes, err := hex.DecodeString(paymentHash)
	if err != nil {
		return nil, fmt.Errorf("invalid payment hash: %w", err)
	}
	hashB64 := base64.URLEncoding.EncodeToString(hashBytes)

	resp, err := s.doRequest(ctx, cfg, "GET", "/v1/invoice/"+hashB64, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("invoice not found")
	}
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to check invoice: %s", string(body))
	}

	var result struct {
		Memo           string `json:"memo"`
		Value          string `json:"value"`
		Settled        bool   `json:"settled"`
		SettleDate     string `json:"settle_date"`
		PaymentRequest string `json:"payment_request"`
		Expiry         string `json:"expiry"`
		CreationDate   string `json:"creation_date"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode invoice response: %w", err)
	}

	var amountSats int64
	fmt.Sscanf(result.Value, "%d", &amountSats)

	var expirySecs, creationDate int64
	fmt.Sscanf(result.Expiry, "%d", &expirySecs)
	fmt.Sscanf(result.CreationDate, "%d", &creationDate)

	invoice := &Invoice{
		PaymentRequest: result.PaymentRequest,
		PaymentHash:    paymentHash,
		AmountSats:     amountSats,
		ExpiresAt:      time.Unix(creationDate+expirySecs, 0),
		Memo:           result.Memo,
		Settled:        result.Settled,
	}

	if result.Settled && result.SettleDate != "" && result.SettleDate != "0" {
		var settleTs int64
		fmt.Sscanf(result.SettleDate, "%d", &settleTs)
		if settleTs > 0 {
			t := time.Unix(settleTs, 0)
			invoice.SettledAt = &t
		}
	}

	return invoice, nil
}

// AccessInvoiceRequest contains the parameters for creating an access invoice.
type AccessInvoiceRequest struct {
	Pubkey string // hex pubkey
	Npub   string // bech32 npub
	TierID string // pricing tier ID
}

// AccessInvoice represents an invoice for relay access.
type AccessInvoice struct {
	PaymentHash    string `json:"payment_hash"`
	PaymentRequest string `json:"payment_request"`
	AmountSats     int64  `json:"amount_sats"`
	TierID         string `json:"tier_id"`
	TierName       string `json:"tier_name"`
	ExpiresAt      int64  `json:"expires_at"` // Unix timestamp
	Memo           string `json:"memo"`
}

// CreateAccessInvoice creates an invoice for paid relay access.
// It creates an invoice via LND and stores it in the database for tracking.
func (s *LightningService) CreateAccessInvoice(ctx context.Context, req AccessInvoiceRequest) (*AccessInvoice, error) {
	if !s.IsConfigured() {
		return nil, ErrLNDNotConfigured
	}

	// Get the pricing tier
	tier, err := s.getPricingTier(ctx, req.TierID)
	if err != nil {
		return nil, fmt.Errorf("failed to get pricing tier: %w", err)
	}
	if tier == nil {
		return nil, fmt.Errorf("pricing tier not found: %s", req.TierID)
	}
	if !tier.Enabled {
		return nil, fmt.Errorf("pricing tier is disabled: %s", req.TierID)
	}

	// Create memo for the invoice
	shortPubkey := req.Pubkey
	if len(shortPubkey) > 12 {
		shortPubkey = shortPubkey[:6] + "..." + shortPubkey[len(shortPubkey)-6:]
	}
	memo := fmt.Sprintf("Roostr %s access for %s", tier.Name, shortPubkey)

	// Create invoice via LND (15 minute expiry)
	expirySecs := int64(900)
	invoice, err := s.CreateInvoice(ctx, tier.AmountSats, memo, expirySecs)
	if err != nil {
		return nil, fmt.Errorf("failed to create LND invoice: %w", err)
	}

	// Store in database for tracking
	pendingInvoice := &db.PendingInvoice{
		PaymentHash:    invoice.PaymentHash,
		Pubkey:         req.Pubkey,
		Npub:           req.Npub,
		TierID:         req.TierID,
		AmountSats:     tier.AmountSats,
		PaymentRequest: invoice.PaymentRequest,
		Memo:           memo,
		ExpiresAt:      invoice.ExpiresAt,
	}

	if err := s.db.CreatePendingInvoice(ctx, pendingInvoice); err != nil {
		return nil, fmt.Errorf("failed to store pending invoice: %w", err)
	}

	return &AccessInvoice{
		PaymentHash:    invoice.PaymentHash,
		PaymentRequest: invoice.PaymentRequest,
		AmountSats:     tier.AmountSats,
		TierID:         req.TierID,
		TierName:       tier.Name,
		ExpiresAt:      invoice.ExpiresAt.Unix(),
		Memo:           memo,
	}, nil
}

// getPricingTier retrieves a pricing tier by ID.
func (s *LightningService) getPricingTier(ctx context.Context, tierID string) (*db.PricingTier, error) {
	tiers, err := s.db.GetPricingTiers(ctx)
	if err != nil {
		return nil, err
	}
	for _, t := range tiers {
		if t.ID == tierID {
			return &t, nil
		}
	}
	return nil, nil
}

// GetPendingInvoice retrieves a pending invoice by payment hash.
func (s *LightningService) GetPendingInvoice(ctx context.Context, paymentHash string) (*db.PendingInvoice, error) {
	return s.db.GetPendingInvoice(ctx, paymentHash)
}

// doRequest performs an HTTP request to the LND REST API.
func (s *LightningService) doRequest(ctx context.Context, cfg *LNDConfig, method, path string, body interface{}) (*http.Response, error) {
	url := fmt.Sprintf("https://%s%s", cfg.Host, path)

	var reqBody io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		reqBody = bytes.NewReader(jsonBody)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set the macaroon header
	req.Header.Set("Grpc-Metadata-macaroon", cfg.MacaroonHex)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrLNDConnectionFailed, err)
	}

	return resp, nil
}
