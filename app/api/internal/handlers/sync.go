package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/roostr/roostr/app/api/internal/db"
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

// GetDefaultRelays returns the default list of public relays for syncing.
// GET /api/v1/sync/relays
func (h *Handler) GetDefaultRelays(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, http.StatusOK, map[string]interface{}{
		"relays": services.DefaultSyncRelays,
	})
}
