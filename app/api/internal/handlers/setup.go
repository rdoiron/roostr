package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/roostr/roostr/app/api/internal/db"
)

// GetSetupStatus returns whether initial setup has been completed.
func (h *Handler) GetSetupStatus(w http.ResponseWriter, r *http.Request) {
	completed, err := h.db.IsSetupCompleted(r.Context())
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to check setup status", "SETUP_CHECK_FAILED")
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"completed": completed,
	})
}

// CompleteSetupRequest is the request body for completing setup.
type CompleteSetupRequest struct {
	OperatorPubkey string `json:"operator_pubkey"`
	OperatorNpub   string `json:"operator_npub"`
	RelayName      string `json:"relay_name"`
	RelayDesc      string `json:"relay_description"`
	AccessMode     string `json:"access_mode"`
}

// CompleteSetup marks the initial setup as complete.
func (h *Handler) CompleteSetup(w http.ResponseWriter, r *http.Request) {
	var req CompleteSetupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body", "INVALID_JSON")
		return
	}

	// Validate required fields
	if req.OperatorPubkey == "" {
		respondError(w, http.StatusBadRequest, "Operator pubkey is required", "MISSING_PUBKEY")
		return
	}

	ctx := r.Context()

	// Check if already completed
	completed, _ := h.db.IsSetupCompleted(ctx)
	if completed {
		respondError(w, http.StatusConflict, "Setup already completed", "SETUP_ALREADY_DONE")
		return
	}

	// Save operator pubkey
	if err := h.db.SetOperatorPubkey(ctx, req.OperatorPubkey); err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to save operator", "SAVE_FAILED")
		return
	}

	// Add operator to whitelist
	if err := h.db.AddWhitelistEntry(ctx, db.WhitelistEntry{
		Pubkey:     req.OperatorPubkey,
		Npub:       req.OperatorNpub,
		Nickname:   "Operator",
		IsOperator: true,
	}); err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to whitelist operator", "WHITELIST_FAILED")
		return
	}

	// Set access mode
	accessMode := req.AccessMode
	if accessMode == "" {
		accessMode = "private"
	}
	if err := h.db.SetAccessMode(ctx, accessMode); err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to set access mode", "MODE_FAILED")
		return
	}

	// Mark setup as complete
	if err := h.db.SetSetupCompleted(ctx); err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to complete setup", "COMPLETE_FAILED")
		return
	}

	// Log the action
	h.db.AddAuditLog(ctx, "setup_completed", map[string]string{
		"operator": req.OperatorPubkey,
		"mode":     accessMode,
	}, req.OperatorPubkey)

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "Setup completed successfully",
	})
}
