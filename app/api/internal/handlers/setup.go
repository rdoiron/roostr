package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/roostr/roostr/app/api/internal/db"
	"github.com/roostr/roostr/app/api/internal/nostr"
)

// GetSetupStatus returns whether initial setup has been completed.
// If completed, it also returns operator info and configuration.
func (h *Handler) GetSetupStatus(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	completed, err := h.db.IsSetupCompleted(ctx)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to check setup status", "SETUP_CHECK_FAILED")
		return
	}

	response := map[string]interface{}{
		"completed": completed,
	}

	// If setup is complete, include additional info
	if completed {
		if pubkey, err := h.db.GetOperatorPubkey(ctx); err == nil && pubkey != "" {
			response["operator_pubkey"] = pubkey
			// Generate npub from hex pubkey
			if npub, err := nostr.EncodeNpub(pubkey); err == nil {
				response["operator_npub"] = npub
			}
		}

		if mode, err := h.db.GetAccessMode(ctx); err == nil {
			response["access_mode"] = mode
		}
	}

	respondJSON(w, http.StatusOK, response)
}

// ValidateIdentity validates a pubkey or NIP-05 identifier.
// Query parameter: input (npub, hex pubkey, or NIP-05 identifier)
func (h *Handler) ValidateIdentity(w http.ResponseWriter, r *http.Request) {
	input := r.URL.Query().Get("input")
	if input == "" {
		respondError(w, http.StatusBadRequest, "Input parameter is required", "MISSING_INPUT")
		return
	}

	ctx := r.Context()

	// Try to resolve the identity
	hexPubkey, npub, source, nip05Name, err := nostr.ResolveIdentity(ctx, input)
	if err != nil {
		// Determine the appropriate error code
		errorCode := "INVALID_IDENTITY"
		errorMsg := "Invalid identity format"

		if errors.Is(err, nostr.ErrInvalidPubkey) || errors.Is(err, nostr.ErrInvalidNpub) {
			errorCode = "INVALID_PUBKEY"
			errorMsg = "Invalid pubkey: must be a valid npub or 64-character hex string"
		} else if errors.Is(err, nostr.ErrInvalidNIP05Format) {
			errorCode = "INVALID_NIP05"
			errorMsg = "Invalid NIP-05 identifier format"
		} else if errors.Is(err, nostr.ErrNIP05FetchFailed) {
			errorCode = "NIP05_FETCH_FAILED"
			errorMsg = "Could not fetch NIP-05 data from domain"
		} else if errors.Is(err, nostr.ErrNIP05NotFound) {
			errorCode = "NIP05_NOT_FOUND"
			errorMsg = "Name not found in NIP-05 response"
		} else if errors.Is(err, nostr.ErrNIP05InvalidPubkey) {
			errorCode = "NIP05_INVALID_PUBKEY"
			errorMsg = "NIP-05 response contains an invalid pubkey"
		}

		respondJSON(w, http.StatusOK, map[string]interface{}{
			"valid":   false,
			"error":   errorMsg,
			"code":    errorCode,
			"details": err.Error(),
		})
		return
	}

	// Build successful response
	response := map[string]interface{}{
		"valid":  true,
		"pubkey": hexPubkey,
		"npub":   npub,
		"source": source,
	}

	// Include NIP-05 name if resolved via NIP-05
	if source == "nip05" && nip05Name != "" {
		response["nip05_name"] = nip05Name
	}

	respondJSON(w, http.StatusOK, response)
}

// CompleteSetupRequest is the request body for completing setup.
type CompleteSetupRequest struct {
	// OperatorIdentity can be an npub, hex pubkey, or NIP-05 identifier
	OperatorIdentity string `json:"operator_identity"`
	// Legacy fields for backward compatibility
	OperatorPubkey string `json:"operator_pubkey,omitempty"`
	OperatorNpub   string `json:"operator_npub,omitempty"`
	// Relay configuration
	RelayName string `json:"relay_name"`
	RelayDesc string `json:"relay_description"`
	// Access mode: private, paid, public
	AccessMode string `json:"access_mode"`
}

// CompleteSetup marks the initial setup as complete.
func (h *Handler) CompleteSetup(w http.ResponseWriter, r *http.Request) {
	var req CompleteSetupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body", "INVALID_JSON")
		return
	}

	ctx := r.Context()

	// Check if already completed
	completed, _ := h.db.IsSetupCompleted(ctx)
	if completed {
		respondError(w, http.StatusConflict, "Setup already completed", "SETUP_ALREADY_DONE")
		return
	}

	// Resolve operator identity
	var hexPubkey, npub string
	var err error

	// Support both new operator_identity and legacy operator_pubkey fields
	if req.OperatorIdentity != "" {
		hexPubkey, npub, _, _, err = nostr.ResolveIdentity(ctx, req.OperatorIdentity)
		if err != nil {
			// Determine appropriate error response
			errorCode := "INVALID_IDENTITY"
			errorMsg := "Invalid operator identity"

			if errors.Is(err, nostr.ErrInvalidPubkey) || errors.Is(err, nostr.ErrInvalidNpub) {
				errorCode = "INVALID_PUBKEY"
				errorMsg = "Invalid pubkey: must be a valid npub or 64-character hex string"
			} else if errors.Is(err, nostr.ErrInvalidNIP05Format) {
				errorCode = "INVALID_NIP05"
				errorMsg = "Invalid NIP-05 identifier format"
			} else if errors.Is(err, nostr.ErrNIP05FetchFailed) {
				errorCode = "NIP05_FETCH_FAILED"
				errorMsg = "Could not fetch NIP-05 data from domain"
			} else if errors.Is(err, nostr.ErrNIP05NotFound) {
				errorCode = "NIP05_NOT_FOUND"
				errorMsg = "Name not found in NIP-05 response"
			}

			respondErrorWithDetails(w, http.StatusBadRequest, errorMsg, errorCode, err.Error())
			return
		}
	} else if req.OperatorPubkey != "" {
		// Legacy: use provided pubkey and npub directly
		hexPubkey, npub, err = nostr.ValidatePubkey(req.OperatorPubkey)
		if err != nil {
			respondError(w, http.StatusBadRequest, "Invalid operator pubkey", "INVALID_PUBKEY")
			return
		}
		// Use provided npub if hex pubkey was given
		if req.OperatorNpub != "" {
			npub = req.OperatorNpub
		}
	} else {
		respondError(w, http.StatusBadRequest, "Operator identity is required", "MISSING_IDENTITY")
		return
	}

	// Validate access mode
	accessMode := req.AccessMode
	if accessMode == "" {
		accessMode = "private"
	}
	if accessMode != "private" && accessMode != "paid" && accessMode != "public" {
		respondError(w, http.StatusBadRequest, "Invalid access mode. Must be: private, paid, or public", "INVALID_ACCESS_MODE")
		return
	}

	// Save operator pubkey
	if err := h.db.SetOperatorPubkey(ctx, hexPubkey); err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to save operator", "SAVE_FAILED")
		return
	}

	// Also store the npub for convenience
	if err := h.db.SetAppState(ctx, "operator_npub", npub); err != nil {
		// Non-fatal, continue
	}

	// Add operator to whitelist
	if err := h.db.AddWhitelistEntry(ctx, db.WhitelistEntry{
		Pubkey:     hexPubkey,
		Npub:       npub,
		Nickname:   "Operator",
		IsOperator: true,
	}); err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to whitelist operator", "WHITELIST_FAILED")
		return
	}

	// Set access mode
	if err := h.db.SetAccessMode(ctx, accessMode); err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to set access mode", "MODE_FAILED")
		return
	}

	// Update relay config with name, description, and operator contact
	if h.configMgr != nil {
		cfg, err := h.configMgr.Read()
		if err == nil {
			// Update info section with setup values
			if req.RelayName != "" {
				cfg.Info.Name = req.RelayName
			}
			if req.RelayDesc != "" {
				cfg.Info.Description = req.RelayDesc
			}
			// Set operator pubkey and contact
			cfg.Info.Pubkey = hexPubkey
			cfg.Info.Contact = npub

			// Write updated config
			if err := h.configMgr.Write(cfg); err != nil {
				// Log but don't fail setup - config can be updated later via settings
			}
		}
	}

	// Mark setup as complete
	if err := h.db.SetSetupCompleted(ctx); err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to complete setup", "COMPLETE_FAILED")
		return
	}

	// Log the action
	h.db.AddAuditLog(ctx, "setup_completed", map[string]string{
		"operator": hexPubkey,
		"mode":     accessMode,
	}, hexPubkey)

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"success":         true,
		"message":         "Setup completed successfully",
		"operator_pubkey": hexPubkey,
		"operator_npub":   npub,
		"access_mode":     accessMode,
	})
}
