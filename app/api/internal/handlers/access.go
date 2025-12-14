package handlers

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"net/url"

	"github.com/roostr/roostr/app/api/internal/db"
	"github.com/roostr/roostr/app/api/internal/nostr"
)

// ============================================================================
// Access Mode
// ============================================================================

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

// SetAccessMode updates the access mode and syncs to config.toml.
func (h *Handler) SetAccessMode(w http.ResponseWriter, r *http.Request) {
	var req SetAccessModeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body", "INVALID_JSON")
		return
	}

	// Validate mode
	// Modes: open (anyone), whitelist (only allowed), paid (whitelist + paid), blacklist (block specific)
	validModes := map[string]bool{"open": true, "whitelist": true, "paid": true, "blacklist": true}
	if !validModes[req.Mode] {
		respondError(w, http.StatusBadRequest, "Invalid access mode. Must be: open, whitelist, paid, or blacklist", "INVALID_MODE")
		return
	}

	ctx := r.Context()
	if err := h.db.SetAccessMode(ctx, req.Mode); err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to set access mode", "MODE_SET_FAILED")
		return
	}

	// Sync to config.toml based on mode
	if err := h.syncConfigFromDB(ctx); err != nil {
		log.Printf("Warning: failed to sync config.toml: %v", err)
	}

	// Log the action
	h.db.AddAuditLog(ctx, "access_mode_changed", map[string]string{"mode": req.Mode}, "")

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"mode":    req.Mode,
	})
}

// ============================================================================
// Whitelist
// ============================================================================

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

// AddToWhitelist adds a pubkey to the whitelist and syncs to config.toml.
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

	// Sync to config.toml and reload relay
	if err := h.syncConfigFromDB(ctx); err != nil {
		log.Printf("Warning: failed to sync config.toml: %v", err)
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

// RemoveFromWhitelist removes a pubkey from the whitelist and syncs to config.toml.
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

	// Sync to config.toml and reload relay
	if err := h.syncConfigFromDB(ctx); err != nil {
		log.Printf("Warning: failed to sync config.toml: %v", err)
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

// ============================================================================
// Blacklist
// ============================================================================

// GetBlacklist returns all blacklisted pubkeys.
func (h *Handler) GetBlacklist(w http.ResponseWriter, r *http.Request) {
	entries, err := h.db.GetBlacklist(r.Context())
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to get blacklist", "BLACKLIST_FETCH_FAILED")
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"entries": entries,
	})
}

// AddToBlacklistRequest is the request body for adding to blacklist.
type AddToBlacklistRequest struct {
	Pubkey string `json:"pubkey"`
	Npub   string `json:"npub"`
	Reason string `json:"reason,omitempty"`
}

// AddToBlacklist adds a pubkey to the blacklist and syncs to config.toml.
func (h *Handler) AddToBlacklist(w http.ResponseWriter, r *http.Request) {
	var req AddToBlacklistRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body", "INVALID_JSON")
		return
	}

	if req.Pubkey == "" {
		respondError(w, http.StatusBadRequest, "Pubkey is required", "MISSING_PUBKEY")
		return
	}

	ctx := r.Context()
	entry := db.BlacklistEntry{
		Pubkey: req.Pubkey,
		Npub:   req.Npub,
		Reason: req.Reason,
	}

	if err := h.db.AddBlacklistEntry(ctx, entry); err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to add to blacklist", "BLACKLIST_ADD_FAILED")
		return
	}

	// Sync to config.toml and reload relay
	if err := h.syncConfigFromDB(ctx); err != nil {
		log.Printf("Warning: failed to sync config.toml: %v", err)
	}

	// Log the action
	h.db.AddAuditLog(ctx, "blacklist_add", map[string]string{
		"pubkey": req.Pubkey,
		"reason": req.Reason,
	}, "")

	respondJSON(w, http.StatusCreated, map[string]interface{}{
		"success": true,
		"message": "Added to blacklist",
	})
}

// RemoveFromBlacklist removes a pubkey from the blacklist and syncs to config.toml.
func (h *Handler) RemoveFromBlacklist(w http.ResponseWriter, r *http.Request) {
	pubkey := r.PathValue("pubkey")
	if pubkey == "" {
		respondError(w, http.StatusBadRequest, "Pubkey is required", "MISSING_PUBKEY")
		return
	}

	ctx := r.Context()
	if err := h.db.RemoveBlacklistEntry(ctx, pubkey); err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to remove from blacklist", "BLACKLIST_REMOVE_FAILED")
		return
	}

	// Sync to config.toml and reload relay
	if err := h.syncConfigFromDB(ctx); err != nil {
		log.Printf("Warning: failed to sync config.toml: %v", err)
	}

	// Log the action
	h.db.AddAuditLog(ctx, "blacklist_remove", map[string]string{"pubkey": pubkey}, "")

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "Removed from blacklist",
	})
}

// ============================================================================
// NIP-05 Resolution
// ============================================================================

// ResolveNIP05 resolves a NIP-05 identifier to a pubkey.
func (h *Handler) ResolveNIP05(w http.ResponseWriter, r *http.Request) {
	identifier := r.PathValue("identifier")
	if identifier == "" {
		respondError(w, http.StatusBadRequest, "Identifier is required", "MISSING_IDENTIFIER")
		return
	}

	// URL decode the identifier (handles @ -> %40)
	decoded, err := url.QueryUnescape(identifier)
	if err != nil {
		decoded = identifier
	}

	result, err := nostr.ResolveNIP05(r.Context(), decoded)
	if err != nil {
		switch err {
		case nostr.ErrInvalidNIP05Format:
			respondError(w, http.StatusBadRequest, "Invalid NIP-05 identifier format", "INVALID_NIP05_FORMAT")
		case nostr.ErrNIP05NotFound:
			respondError(w, http.StatusNotFound, "Name not found at domain", "NIP05_NOT_FOUND")
		case nostr.ErrNIP05InvalidPubkey:
			respondError(w, http.StatusBadGateway, "Domain returned invalid pubkey", "NIP05_INVALID_PUBKEY")
		default:
			respondErrorWithDetails(w, http.StatusBadGateway, "Failed to resolve NIP-05", "NIP05_FETCH_FAILED", err.Error())
		}
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"name":   result.Name,
		"domain": result.Domain,
		"pubkey": result.Pubkey,
		"npub":   result.Npub,
		"relays": result.Relays,
	})
}

// ============================================================================
// Config Sync Helper
// ============================================================================

// syncConfigFromDB reads the whitelist/blacklist from DB and writes to config.toml.
// It also reloads the relay if it's running.
func (h *Handler) syncConfigFromDB(_ interface{}) error {
	if h.configMgr == nil {
		return nil // No config manager, skip sync
	}

	// Use a background context for the sync operations
	ctx := context.Background()

	// Get whitelist pubkeys from DB
	entries, err := h.db.GetWhitelistMeta(ctx)
	if err != nil {
		return err
	}

	// Extract hex pubkeys for config.toml
	whitelist := make([]string, len(entries))
	for i, e := range entries {
		whitelist[i] = e.Pubkey
	}

	// Update config.toml whitelist
	if err := h.configMgr.UpdateWhitelist(whitelist); err != nil {
		return err
	}

	// Get blacklist pubkeys from DB
	blacklistEntries, err := h.db.GetBlacklist(ctx)
	if err != nil {
		return err
	}

	// Extract hex pubkeys for config.toml
	blacklist := make([]string, len(blacklistEntries))
	for i, e := range blacklistEntries {
		blacklist[i] = e.Pubkey
	}

	// Update config.toml blacklist
	if err := h.configMgr.UpdateBlacklist(blacklist); err != nil {
		return err
	}

	// Reload relay to pick up config changes
	if h.relay != nil {
		if err := h.relay.Reload(); err != nil {
			log.Printf("Warning: failed to reload relay: %v", err)
		}
	}

	return nil
}
