package handlers

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/roostr/roostr/app/api/internal/db"
	"github.com/roostr/roostr/app/api/internal/nostr"
	"github.com/roostr/roostr/app/api/internal/services"
)

// StartSyncRequest is the request body for starting a sync.
type StartSyncRequest struct {
	Pubkeys        []string `json:"pubkeys"`
	Relays         []string `json:"relays,omitempty"`
	EventKinds     []int    `json:"event_kinds,omitempty"`
	SinceTimestamp *int64   `json:"since_timestamp,omitempty"`
}

// StartSync initiates a sync job from public relays.
// POST /api/v1/sync/start
func (h *Handler) StartSync(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req StartSyncRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body", "INVALID_JSON")
		return
	}

	// Validate pubkeys
	if len(req.Pubkeys) == 0 {
		respondError(w, http.StatusBadRequest, "At least one pubkey is required", "MISSING_PUBKEYS")
		return
	}

	// Start sync via service
	syncReq := services.SyncRequest{
		Pubkeys:        req.Pubkeys,
		Relays:         req.Relays,
		EventKinds:     req.EventKinds,
		SinceTimestamp: req.SinceTimestamp,
	}

	jobID, err := h.services.Sync.StartSync(ctx, syncReq)
	if err != nil {
		if err.Error() == "a sync job is already running" {
			respondError(w, http.StatusConflict, err.Error(), "SYNC_ALREADY_RUNNING")
			return
		}
		respondError(w, http.StatusInternalServerError, "Failed to start sync: "+err.Error(), "SYNC_START_FAILED")
		return
	}

	// Return 202 Accepted with job ID
	w.Header().Set("Location", "/api/v1/sync/status")
	respondJSON(w, http.StatusAccepted, map[string]interface{}{
		"job_id":  jobID,
		"status":  "running",
		"message": "Sync job started",
	})
}

// GetSyncStatus returns the status of the current or specified sync job.
// GET /api/v1/sync/status
// GET /api/v1/sync/status?id=123
func (h *Handler) GetSyncStatus(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Check for specific job ID
	idStr := r.URL.Query().Get("id")

	var jobID int64
	if idStr != "" {
		var err error
		jobID, err = strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			respondError(w, http.StatusBadRequest, "Invalid job ID", "INVALID_ID")
			return
		}
	} else {
		// Get current running job
		jobID = h.services.Sync.GetCurrentJobID()
		if jobID == 0 {
			// No running job, check for most recent
			jobs, err := h.db.GetSyncJobs(ctx, "", 1, 0)
			if err != nil || len(jobs) == 0 {
				respondJSON(w, http.StatusOK, map[string]interface{}{
					"status":  "idle",
					"message": "No sync jobs found",
				})
				return
			}
			jobID = jobs[0].ID
		}
	}

	job, err := h.db.GetSyncJob(ctx, jobID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to get job status", "STATUS_FAILED")
		return
	}
	if job == nil {
		respondError(w, http.StatusNotFound, "Job not found", "JOB_NOT_FOUND")
		return
	}

	respondJSON(w, http.StatusOK, job)
}

// CancelSync cancels the currently running sync job.
// POST /api/v1/sync/cancel
func (h *Handler) CancelSync(w http.ResponseWriter, r *http.Request) {
	if err := h.services.Sync.CancelSync(); err != nil {
		respondError(w, http.StatusBadRequest, err.Error(), "CANCEL_FAILED")
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "Sync cancellation requested",
	})
}

// GetSyncHistory returns a list of past sync jobs.
// GET /api/v1/sync/history
// GET /api/v1/sync/history?limit=10&offset=0
func (h *Handler) GetSyncHistory(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse pagination params
	limit := 20
	offset := 0

	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			limit = parsed
			if limit > 100 {
				limit = 100
			}
		}
	}

	if o := r.URL.Query().Get("offset"); o != "" {
		if parsed, err := strconv.Atoi(o); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	jobs, err := h.db.GetSyncJobs(ctx, "", limit, offset)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to get sync history", "HISTORY_FAILED")
		return
	}

	// Ensure we return an empty array instead of null
	if jobs == nil {
		jobs = []db.SyncJob{}
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"jobs":   jobs,
		"limit":  limit,
		"offset": offset,
	})
}

// ============================================================================
// Sync Pubkeys
// ============================================================================

// GetSyncPubkeys returns all configured sync pubkeys.
// GET /api/v1/sync/pubkeys
func (h *Handler) GetSyncPubkeys(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	pubkeys, err := h.db.GetSyncPubkeys(ctx)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to get sync pubkeys", "DB_ERROR")
		return
	}

	if pubkeys == nil {
		pubkeys = []db.SyncPubkey{}
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"pubkeys": pubkeys,
	})
}

// AddSyncPubkeyRequest is the request body for adding a sync pubkey.
type AddSyncPubkeyRequest struct {
	Identifier string `json:"identifier"` // npub, hex, or NIP-05
}

// AddSyncPubkey adds a pubkey to the sync configuration.
// POST /api/v1/sync/pubkeys
func (h *Handler) AddSyncPubkey(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req AddSyncPubkeyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body", "INVALID_JSON")
		return
	}

	if req.Identifier == "" {
		respondError(w, http.StatusBadRequest, "Identifier is required", "MISSING_IDENTIFIER")
		return
	}

	// Resolve the identifier to hex pubkey and npub
	hexPubkey, npub, _, nip05Name, err := nostr.ResolveIdentity(ctx, req.Identifier)
	if err != nil {
		respondErrorWithDetails(w, http.StatusBadRequest, "Invalid identifier", "INVALID_IDENTIFIER", err.Error())
		return
	}

	// Check if already exists
	existing, err := h.db.GetSyncPubkeyByPubkey(ctx, hexPubkey)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Database error", "DB_ERROR")
		return
	}
	if existing != nil {
		respondError(w, http.StatusConflict, "Pubkey already in sync configuration", "DUPLICATE")
		return
	}

	// Determine nickname - use NIP-05 name if available, otherwise original input
	nickname := req.Identifier
	if nip05Name != "" {
		nickname = req.Identifier // Keep full NIP-05 identifier as nickname
	}

	entry := db.SyncPubkey{
		Pubkey:     hexPubkey,
		Npub:       npub,
		Nickname:   nickname,
		IsOperator: false,
	}

	if err := h.db.AddSyncPubkey(ctx, entry); err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to add sync pubkey", "DB_ERROR")
		return
	}

	// Fetch the entry to get the full object with timestamp
	added, _ := h.db.GetSyncPubkeyByPubkey(ctx, hexPubkey)
	if added == nil {
		added = &entry
	}

	respondJSON(w, http.StatusCreated, map[string]interface{}{
		"success": true,
		"pubkey":  added,
	})
}

// RemoveSyncPubkey removes a pubkey from sync configuration.
// DELETE /api/v1/sync/pubkeys/{pubkey}
func (h *Handler) RemoveSyncPubkey(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	pubkey := r.PathValue("pubkey")
	if pubkey == "" {
		respondError(w, http.StatusBadRequest, "Pubkey is required", "MISSING_PUBKEY")
		return
	}

	err := h.db.RemoveSyncPubkey(ctx, pubkey)
	if err != nil {
		if err.Error() == "cannot remove operator pubkey from sync configuration" {
			respondError(w, http.StatusForbidden, err.Error(), "OPERATOR_PROTECTED")
			return
		}
		if err.Error() == "pubkey not found in sync configuration" {
			respondError(w, http.StatusNotFound, err.Error(), "NOT_FOUND")
			return
		}
		respondError(w, http.StatusInternalServerError, "Failed to remove sync pubkey", "DB_ERROR")
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "Pubkey removed from sync configuration",
	})
}

// ============================================================================
// Sync Relays
// ============================================================================

// GetSyncRelays returns configured sync relays, or defaults if none configured.
// GET /api/v1/sync/relays
func (h *Handler) GetSyncRelays(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	relays, err := h.db.GetSyncRelays(ctx)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to get sync relays", "DB_ERROR")
		return
	}

	usingDefaults := false
	if len(relays) == 0 {
		// Return hardcoded defaults if none configured
		usingDefaults = true
		for _, url := range services.DefaultSyncRelays {
			relays = append(relays, db.SyncRelay{
				URL:       url,
				IsDefault: true,
			})
		}
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"relays":         relays,
		"using_defaults": usingDefaults,
	})
}

// AddSyncRelayRequest is the request body for adding a sync relay.
type AddSyncRelayRequest struct {
	URL string `json:"url"`
}

// AddSyncRelay adds a relay to the sync configuration.
// POST /api/v1/sync/relays
func (h *Handler) AddSyncRelay(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req AddSyncRelayRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body", "INVALID_JSON")
		return
	}

	if req.URL == "" {
		respondError(w, http.StatusBadRequest, "URL is required", "MISSING_URL")
		return
	}

	// Validate URL format
	if !isValidRelayURL(req.URL) {
		respondError(w, http.StatusBadRequest, "Invalid relay URL. Must start with wss:// or ws://", "INVALID_URL")
		return
	}

	// Check if it's a default relay
	isDefault := false
	for _, defaultRelay := range services.DefaultSyncRelays {
		if defaultRelay == req.URL {
			isDefault = true
			break
		}
	}

	entry := db.SyncRelay{
		URL:       req.URL,
		IsDefault: isDefault,
	}

	if err := h.db.AddSyncRelay(ctx, entry); err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to add sync relay", "DB_ERROR")
		return
	}

	respondJSON(w, http.StatusCreated, map[string]interface{}{
		"success": true,
		"relay":   entry,
	})
}

// RemoveSyncRelay removes a relay from sync configuration.
// DELETE /api/v1/sync/relays/{url}
func (h *Handler) RemoveSyncRelay(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	relayURL := r.PathValue("url")
	if relayURL == "" {
		respondError(w, http.StatusBadRequest, "URL is required", "MISSING_URL")
		return
	}

	if err := h.db.RemoveSyncRelay(ctx, relayURL); err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to remove sync relay", "DB_ERROR")
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "Relay removed from sync configuration",
	})
}

// ResetSyncRelays resets relays to defaults.
// POST /api/v1/sync/relays/reset
func (h *Handler) ResetSyncRelays(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Clear all existing relays
	if err := h.db.ResetSyncRelays(ctx); err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to reset relays", "DB_ERROR")
		return
	}

	// Insert defaults
	if err := h.db.InitSyncRelaysFromDefaults(ctx, services.DefaultSyncRelays); err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to initialize defaults", "DB_ERROR")
		return
	}

	// Return the new relay list
	relays, _ := h.db.GetSyncRelays(ctx)

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "Relays reset to defaults",
		"relays":  relays,
	})
}

// isValidRelayURL checks if a URL is a valid WebSocket relay URL
func isValidRelayURL(u string) bool {
	if !strings.HasPrefix(u, "wss://") && !strings.HasPrefix(u, "ws://") {
		return false
	}
	// Try to parse as URL
	_, err := url.Parse(u)
	return err == nil
}
