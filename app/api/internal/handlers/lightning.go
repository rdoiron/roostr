package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/roostr/roostr/app/api/internal/services"
)

// GetLightningStatus returns the current Lightning node connection status.
// GET /api/v1/lightning/status
func (h *Handler) GetLightningStatus(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Load config from database if not already loaded
	if err := h.services.Lightning.LoadConfig(ctx); err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to load Lightning config", "CONFIG_LOAD_FAILED")
		return
	}

	// Check if configured
	if !h.services.Lightning.IsConfigured() {
		respondJSON(w, http.StatusOK, map[string]interface{}{
			"configured": false,
			"enabled":    false,
			"message":    "Lightning node not configured",
		})
		return
	}

	// Get config from database to check enabled state
	dbCfg, err := h.db.GetLightningConfig(ctx)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to get config", "DB_ERROR")
		return
	}

	enabled := dbCfg != nil && dbCfg.Enabled

	// Try to get node info
	info, err := h.services.Lightning.GetInfo(ctx)
	if err != nil {
		// Return configured but not connected
		response := map[string]interface{}{
			"configured": true,
			"enabled":    enabled,
			"connected":  false,
			"error":      err.Error(),
		}

		if errors.Is(err, services.ErrLNDAuthFailed) {
			response["error_code"] = "AUTH_FAILED"
		} else if errors.Is(err, services.ErrLNDConnectionFailed) {
			response["error_code"] = "CONNECTION_FAILED"
		}

		respondJSON(w, http.StatusOK, response)
		return
	}

	// Get balance
	balance, _ := h.services.Lightning.GetBalance(ctx)

	// Update last verified timestamp
	h.db.SetLightningVerified(ctx)

	response := map[string]interface{}{
		"configured": true,
		"enabled":    enabled,
		"connected":  true,
		"node_info":  info,
	}

	if balance != nil {
		response["balance"] = balance
	}

	respondJSON(w, http.StatusOK, response)
}

// SaveLightningConfig saves the Lightning node configuration.
// PUT /api/v1/lightning/config
func (h *Handler) SaveLightningConfig(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req struct {
		Host        string `json:"host"`
		MacaroonHex string `json:"macaroon_hex"`
		TLSCertPath string `json:"tls_cert_path,omitempty"`
		Enabled     bool   `json:"enabled"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body", "INVALID_JSON")
		return
	}

	if req.Host == "" {
		respondError(w, http.StatusBadRequest, "Host is required", "MISSING_HOST")
		return
	}

	if req.MacaroonHex == "" {
		respondError(w, http.StatusBadRequest, "Macaroon is required", "MISSING_MACAROON")
		return
	}

	// Normalize host - strip protocol prefix if present
	host := strings.TrimPrefix(req.Host, "https://")
	host = strings.TrimPrefix(host, "http://")
	host = strings.TrimSuffix(host, "/")

	cfg := &services.LNDConfig{
		Host:        host,
		MacaroonHex: req.MacaroonHex,
		TLSCertPath: req.TLSCertPath,
	}

	// Save to database and update service
	if err := h.services.Lightning.SaveConfig(ctx, cfg, req.Enabled); err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to save config: "+err.Error(), "SAVE_FAILED")
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "Lightning configuration saved",
	})
}

// TestLightningConnection tests a Lightning node connection with provided config.
// POST /api/v1/lightning/test
func (h *Handler) TestLightningConnection(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req struct {
		Host        string `json:"host"`
		MacaroonHex string `json:"macaroon_hex"`
		TLSCertPath string `json:"tls_cert_path,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body", "INVALID_JSON")
		return
	}

	if req.Host == "" {
		respondError(w, http.StatusBadRequest, "Host is required", "MISSING_HOST")
		return
	}

	if req.MacaroonHex == "" {
		respondError(w, http.StatusBadRequest, "Macaroon is required", "MISSING_MACAROON")
		return
	}

	// Normalize host - strip protocol prefix if present
	host := strings.TrimPrefix(req.Host, "https://")
	host = strings.TrimPrefix(host, "http://")
	host = strings.TrimSuffix(host, "/")

	cfg := &services.LNDConfig{
		Host:        host,
		MacaroonHex: req.MacaroonHex,
		TLSCertPath: req.TLSCertPath,
	}

	// Test the connection
	info, err := h.services.Lightning.TestConnection(ctx, cfg)
	if err != nil {
		response := map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		}

		if errors.Is(err, services.ErrLNDAuthFailed) {
			response["error_code"] = "AUTH_FAILED"
			response["message"] = "Authentication failed. Please check your macaroon."
		} else if errors.Is(err, services.ErrLNDConnectionFailed) {
			response["error_code"] = "CONNECTION_FAILED"
			response["message"] = "Could not connect to the LND node. Please check the host address."
		} else {
			response["error_code"] = "UNKNOWN"
		}

		respondJSON(w, http.StatusOK, response)
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"success":   true,
		"node_info": info,
		"message":   "Successfully connected to LND node",
	})
}

