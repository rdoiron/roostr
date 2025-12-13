package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/roostr/roostr/app/api/internal/db"
)

// GetAccessMode returns the current access mode.
func (h *Handler) GetAccessMode(w http.ResponseWriter, r *http.Request) {
	mode, err := h.db.GetAccessMode(r.Context())
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to get access mode", "MODE_FETCH_FAILED")
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"mode": mode,
	})
}

// SetAccessModeRequest is the request body for setting access mode.
type SetAccessModeRequest struct {
	Mode string `json:"mode"`
}

// SetAccessMode updates the access mode.
func (h *Handler) SetAccessMode(w http.ResponseWriter, r *http.Request) {
	var req SetAccessModeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body", "INVALID_JSON")
		return
	}

	// Validate mode
	validModes := map[string]bool{"private": true, "paid": true, "public": true}
	if !validModes[req.Mode] {
		respondError(w, http.StatusBadRequest, "Invalid access mode", "INVALID_MODE")
		return
	}

	ctx := r.Context()
	if err := h.db.SetAccessMode(ctx, req.Mode); err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to set access mode", "MODE_SET_FAILED")
		return
	}

	// Log the action
	h.db.AddAuditLog(ctx, "access_mode_changed", map[string]string{"mode": req.Mode}, "")

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"mode":    req.Mode,
	})
}

// GetWhitelist returns all whitelisted pubkeys.
func (h *Handler) GetWhitelist(w http.ResponseWriter, r *http.Request) {
	entries, err := h.db.GetWhitelistMeta(r.Context())
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to get whitelist", "WHITELIST_FETCH_FAILED")
		return
	}

	// Get event counts for each pubkey if relay is connected
	if h.db.IsRelayDBConnected() {
		pubkeys := make([]string, len(entries))
		for i, e := range entries {
			pubkeys[i] = e.Pubkey
		}
		counts, _ := h.db.CountEventsByPubkey(r.Context(), pubkeys)

		// Build response with counts
		type entryWithCount struct {
			db.WhitelistEntry
			EventCount int64 `json:"event_count"`
		}

		result := make([]entryWithCount, len(entries))
		for i, e := range entries {
			result[i] = entryWithCount{
				WhitelistEntry: e,
				EventCount:     counts[e.Pubkey],
			}
		}

		respondJSON(w, http.StatusOK, map[string]interface{}{
			"entries": result,
		})
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"entries": entries,
	})
}

// AddToWhitelistRequest is the request body for adding to whitelist.
type AddToWhitelistRequest struct {
	Pubkey   string `json:"pubkey"`
	Npub     string `json:"npub"`
	Nickname string `json:"nickname,omitempty"`
}

// AddToWhitelist adds a pubkey to the whitelist.
func (h *Handler) AddToWhitelist(w http.ResponseWriter, r *http.Request) {
	var req AddToWhitelistRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body", "INVALID_JSON")
		return
	}

	if req.Pubkey == "" {
		respondError(w, http.StatusBadRequest, "Pubkey is required", "MISSING_PUBKEY")
		return
	}

	ctx := r.Context()
	entry := db.WhitelistEntry{
		Pubkey:   req.Pubkey,
		Npub:     req.Npub,
		Nickname: req.Nickname,
	}

	if err := h.db.AddWhitelistEntry(ctx, entry); err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to add to whitelist", "WHITELIST_ADD_FAILED")
		return
	}

	// Log the action
	h.db.AddAuditLog(ctx, "whitelist_add", map[string]string{
		"pubkey":   req.Pubkey,
		"nickname": req.Nickname,
	}, "")

	respondJSON(w, http.StatusCreated, map[string]interface{}{
		"success": true,
		"message": "Added to whitelist",
	})
}

// RemoveFromWhitelist removes a pubkey from the whitelist.
func (h *Handler) RemoveFromWhitelist(w http.ResponseWriter, r *http.Request) {
	pubkey := r.PathValue("pubkey")
	if pubkey == "" {
		respondError(w, http.StatusBadRequest, "Pubkey is required", "MISSING_PUBKEY")
		return
	}

	ctx := r.Context()
	if err := h.db.RemoveWhitelistEntry(ctx, pubkey); err != nil {
		if err.Error() == "cannot remove operator from whitelist" {
			respondError(w, http.StatusForbidden, err.Error(), "CANNOT_REMOVE_OPERATOR")
			return
		}
		respondError(w, http.StatusInternalServerError, "Failed to remove from whitelist", "WHITELIST_REMOVE_FAILED")
		return
	}

	// Log the action
	h.db.AddAuditLog(ctx, "whitelist_remove", map[string]string{"pubkey": pubkey}, "")

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "Removed from whitelist",
	})
}

// UpdateWhitelistEntryRequest is the request body for updating a whitelist entry.
type UpdateWhitelistEntryRequest struct {
	Nickname string `json:"nickname"`
}

// UpdateWhitelistEntry updates a whitelist entry (e.g., nickname).
func (h *Handler) UpdateWhitelistEntry(w http.ResponseWriter, r *http.Request) {
	pubkey := r.PathValue("pubkey")
	if pubkey == "" {
		respondError(w, http.StatusBadRequest, "Pubkey is required", "MISSING_PUBKEY")
		return
	}

	var req UpdateWhitelistEntryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body", "INVALID_JSON")
		return
	}

	ctx := r.Context()
	if err := h.db.UpdateWhitelistNickname(ctx, pubkey, req.Nickname); err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to update whitelist entry", "WHITELIST_UPDATE_FAILED")
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "Whitelist entry updated",
	})
}
