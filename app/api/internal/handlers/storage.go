package handlers

import (
	"encoding/json"
	"net/http"
	"time"
)

// StorageStatusResponse represents the storage status response.
type StorageStatusResponse struct {
	DatabaseSize    int64      `json:"database_size"`
	AppDatabaseSize int64      `json:"app_database_size"`
	TotalSize       int64      `json:"total_size"`
	AvailableSpace  int64      `json:"available_space"`
	TotalSpace      int64      `json:"total_space"`
	UsagePercent    float64    `json:"usage_percent"`
	TotalEvents     int64      `json:"total_events"`
	OldestEvent     *time.Time `json:"oldest_event,omitempty"`
	NewestEvent     *time.Time `json:"newest_event,omitempty"`
	Status          string     `json:"status"`
	PendingDeletions int64     `json:"pending_deletions"`
}

// RetentionPolicyRequest represents a retention policy update request.
type RetentionPolicyRequest struct {
	RetentionDays int64    `json:"retention_days"`
	Exceptions    []string `json:"exceptions"`
	HonorNIP09    bool     `json:"honor_nip09"`
}

// CleanupRequest represents a manual cleanup request.
type CleanupRequest struct {
	BeforeDate      string   `json:"before_date"`      // ISO 8601 format
	ApplyExceptions bool     `json:"apply_exceptions"` // If true, use retention policy exceptions
	Exceptions      []string `json:"exceptions"`       // Optional explicit exceptions (overrides retention policy)
}

// GetStorageStatus returns the current storage usage and status.
// GET /api/v1/storage/status
func (h *Handler) GetStorageStatus(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get database sizes
	relayDBSize, err := h.db.GetRelayDatabaseSize()
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to get relay database size", "DB_SIZE_FAILED")
		return
	}

	appDBSize, err := h.db.GetAppDatabaseSize()
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to get app database size", "DB_SIZE_FAILED")
		return
	}

	totalSize := relayDBSize + appDBSize

	// Get available disk space
	availableSpace, err := h.db.GetAvailableDiskSpace()
	if err != nil {
		// Non-fatal error, set to 0
		availableSpace = 0
	}

	totalSpace, err := h.db.GetTotalDiskSpace()
	if err != nil {
		totalSpace = 0
	}

	// Calculate usage percent
	var usagePercent float64
	if totalSpace > 0 {
		usedSpace := totalSpace - availableSpace
		usagePercent = float64(usedSpace) / float64(totalSpace) * 100
	}

	// Get relay stats
	var totalEvents int64
	var oldestEvent, newestEvent *time.Time

	if h.db.IsRelayDBConnected() {
		stats, err := h.db.GetRelayStats(ctx)
		if err == nil {
			totalEvents = stats.TotalEvents
			if !stats.OldestEvent.IsZero() {
				oldestEvent = &stats.OldestEvent
			}
			if !stats.NewestEvent.IsZero() {
				newestEvent = &stats.NewestEvent
			}
		}
	}

	// Get pending deletion count
	pendingDeletions, err := h.db.GetPendingDeletionCount(ctx)
	if err != nil {
		pendingDeletions = 0
	}

	// Determine status based on usage
	status := "healthy"
	if usagePercent >= 95 {
		status = "critical"
	} else if usagePercent >= 90 {
		status = "low"
	} else if usagePercent >= 80 {
		status = "warning"
	}

	respondJSON(w, http.StatusOK, StorageStatusResponse{
		DatabaseSize:     relayDBSize,
		AppDatabaseSize:  appDBSize,
		TotalSize:        totalSize,
		AvailableSpace:   availableSpace,
		TotalSpace:       totalSpace,
		UsagePercent:     usagePercent,
		TotalEvents:      totalEvents,
		OldestEvent:      oldestEvent,
		NewestEvent:      newestEvent,
		Status:           status,
		PendingDeletions: pendingDeletions,
	})
}

// GetRetentionPolicy returns the current retention policy settings.
// GET /api/v1/storage/retention
func (h *Handler) GetRetentionPolicy(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	policy, err := h.db.GetRetentionPolicy(ctx)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to get retention policy", "RETENTION_GET_FAILED")
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"retention_days": policy.RetentionDays,
		"exceptions":     policy.Exceptions,
		"honor_nip09":    policy.HonorNIP09,
		"last_run":       policy.LastRun,
	})
}

// UpdateRetentionPolicy updates the retention policy settings.
// PUT /api/v1/storage/retention
func (h *Handler) UpdateRetentionPolicy(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req RetentionPolicyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body", "INVALID_JSON")
		return
	}

	// Validate retention days
	if req.RetentionDays < 0 {
		respondError(w, http.StatusBadRequest, "Retention days must be non-negative", "INVALID_RETENTION_DAYS")
		return
	}

	// Get current policy to preserve LastRun
	currentPolicy, err := h.db.GetRetentionPolicy(ctx)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to get current policy", "RETENTION_GET_FAILED")
		return
	}

	// Update policy
	newPolicy := currentPolicy
	newPolicy.RetentionDays = req.RetentionDays
	newPolicy.Exceptions = req.Exceptions
	newPolicy.HonorNIP09 = req.HonorNIP09

	if err := h.db.SetRetentionPolicy(ctx, newPolicy); err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to update retention policy", "RETENTION_SET_FAILED")
		return
	}

	// Add audit log
	h.db.AddAuditLog(ctx, "retention_policy_updated", map[string]interface{}{
		"retention_days": req.RetentionDays,
		"exceptions":     req.Exceptions,
		"honor_nip09":    req.HonorNIP09,
	}, "")

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"success":        true,
		"message":        "Retention policy updated",
		"retention_days": req.RetentionDays,
		"exceptions":     req.Exceptions,
		"honor_nip09":    req.HonorNIP09,
	})
}

// ManualCleanup performs a manual cleanup of events before a given date.
// POST /api/v1/storage/cleanup
func (h *Handler) ManualCleanup(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req CleanupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body", "INVALID_JSON")
		return
	}

	// Parse the before date
	beforeDate, err := time.Parse(time.RFC3339, req.BeforeDate)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid date format. Use ISO 8601 (YYYY-MM-DDTHH:MM:SSZ)", "INVALID_DATE")
		return
	}

	// Validate date is in the past
	if beforeDate.After(time.Now()) {
		respondError(w, http.StatusBadRequest, "Date must be in the past", "FUTURE_DATE")
		return
	}

	// Determine which exceptions to apply (if any)
	var exceptions []string
	var operatorPubkey string

	if req.ApplyExceptions {
		// Use explicit exceptions if provided, otherwise use retention policy
		if len(req.Exceptions) > 0 {
			exceptions = req.Exceptions
		} else {
			policy, err := h.db.GetRetentionPolicy(ctx)
			if err != nil {
				respondError(w, http.StatusInternalServerError, "Failed to get retention policy", "RETENTION_GET_FAILED")
				return
			}
			exceptions = policy.Exceptions
		}
		// Get operator pubkey for exception handling
		operatorPubkey, _ = h.db.GetOperatorPubkey(ctx)
	}
	// If ApplyExceptions is false, exceptions stays nil/empty - delete ALL events

	// Get size before cleanup
	sizeBefore, _ := h.db.GetRelayDatabaseSize()

	// Open relay writer for deletion
	writer, err := h.db.NewRelayWriter()
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to open database for writing", "DB_WRITE_FAILED")
		return
	}
	defer writer.Close()

	// Delete events
	deletedCount, err := writer.DeleteEventsBefore(ctx, beforeDate, exceptions, operatorPubkey)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to delete events", "DELETE_FAILED")
		return
	}

	// Get size after cleanup (before vacuum)
	sizeAfter, _ := h.db.GetRelayDatabaseSize()
	spaceFreed := sizeBefore - sizeAfter

	// Add audit log
	h.db.AddAuditLog(ctx, "manual_cleanup", map[string]interface{}{
		"before_date":       req.BeforeDate,
		"deleted_count":     deletedCount,
		"space_freed":       spaceFreed,
		"apply_exceptions":  req.ApplyExceptions,
		"exceptions_used":   exceptions,
	}, "")

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"success":       true,
		"deleted_count": deletedCount,
		"space_freed":   spaceFreed,
		"message":       "Cleanup completed. Run VACUUM to fully reclaim disk space.",
	})
}

// RunVacuum runs SQLite VACUUM on the databases to reclaim disk space.
// POST /api/v1/storage/vacuum
func (h *Handler) RunVacuum(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	startTime := time.Now()

	// Get sizes before vacuum
	relayDBSizeBefore, _ := h.db.GetRelayDatabaseSize()
	appDBSizeBefore, _ := h.db.GetAppDatabaseSize()

	// Vacuum app database
	if err := h.db.RunAppVacuum(ctx); err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to vacuum app database", "VACUUM_FAILED")
		return
	}

	// Vacuum relay database
	writer, err := h.db.NewRelayWriter()
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to open relay database for vacuum", "DB_WRITE_FAILED")
		return
	}
	defer writer.Close()

	if err := writer.RunVacuum(ctx); err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to vacuum relay database", "VACUUM_FAILED")
		return
	}

	// Get sizes after vacuum
	relayDBSizeAfter, _ := h.db.GetRelayDatabaseSize()
	appDBSizeAfter, _ := h.db.GetAppDatabaseSize()

	spaceReclaimed := (relayDBSizeBefore - relayDBSizeAfter) + (appDBSizeBefore - appDBSizeAfter)
	duration := time.Since(startTime)

	// Update last vacuum timestamp
	h.db.SetLastVacuumRun(ctx, time.Now())

	// Add audit log
	h.db.AddAuditLog(ctx, "vacuum_run", map[string]interface{}{
		"space_reclaimed": spaceReclaimed,
		"duration_ms":     duration.Milliseconds(),
	}, "")

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"success":         true,
		"space_reclaimed": spaceReclaimed,
		"duration_ms":     duration.Milliseconds(),
	})
}

// GetDeletionRequests returns the list of NIP-09 deletion requests.
// GET /api/v1/storage/deletion-requests
func (h *Handler) GetDeletionRequests(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get status filter from query params
	status := r.URL.Query().Get("status")

	requests, err := h.db.GetDeletionRequests(ctx, status)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to get deletion requests", "DELETE_REQUESTS_FAILED")
		return
	}

	// Get total count
	total := len(requests)

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"requests": requests,
		"total":    total,
	})
}

// GetStorageEstimate returns an estimate of space that would be freed by cleanup.
// GET /api/v1/storage/estimate?before_date=X&apply_exceptions=true
func (h *Handler) GetStorageEstimate(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse before_date from query params
	beforeDateStr := r.URL.Query().Get("before_date")
	if beforeDateStr == "" {
		respondError(w, http.StatusBadRequest, "before_date parameter is required", "MISSING_DATE")
		return
	}

	beforeDate, err := time.Parse(time.RFC3339, beforeDateStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid date format. Use ISO 8601 (YYYY-MM-DDTHH:MM:SSZ)", "INVALID_DATE")
		return
	}

	if !h.db.IsRelayDBConnected() {
		respondError(w, http.StatusServiceUnavailable, "Relay database not connected", "RELAY_NOT_CONNECTED")
		return
	}

	// Check if we should apply retention exceptions
	applyExceptions := r.URL.Query().Get("apply_exceptions") == "true"

	var eventCount int64
	var exceptions []string

	if applyExceptions {
		// Get retention policy exceptions
		policy, err := h.db.GetRetentionPolicy(ctx)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "Failed to get retention policy", "RETENTION_GET_FAILED")
			return
		}
		exceptions = policy.Exceptions

		// Get operator pubkey for exception handling
		operatorPubkey, _ := h.db.GetOperatorPubkey(ctx)

		// Count events excluding exceptions
		eventCount, err = h.db.CountEventsBeforeWithExceptions(ctx, beforeDate, exceptions, operatorPubkey)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "Failed to count events", "COUNT_FAILED")
			return
		}
	} else {
		// Count all events before date
		eventCount, err = h.db.CountEventsBefore(ctx, beforeDate)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "Failed to count events", "COUNT_FAILED")
			return
		}
	}

	// Estimate average event size
	avgSize, err := h.db.EstimateEventSize(ctx)
	if err != nil {
		avgSize = 500 // Default estimate
	}

	estimatedSpace := eventCount * avgSize

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"event_count":       eventCount,
		"estimated_space":   estimatedSpace,
		"before_date":       beforeDate,
		"apply_exceptions":  applyExceptions,
		"exceptions_applied": exceptions,
	})
}

// RunRetentionNow runs the retention policy immediately.
// POST /api/v1/storage/retention/run
func (h *Handler) RunRetentionNow(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if h.services == nil || h.services.Retention == nil {
		respondError(w, http.StatusServiceUnavailable, "Retention service not available", "SERVICE_UNAVAILABLE")
		return
	}

	result, err := h.services.Retention.RunNowSync(ctx)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to run retention: "+err.Error(), "RETENTION_FAILED")
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"success":                   true,
		"events_deleted":            result.EventsDeleted,
		"deletion_requests_processed": result.DeletionRequests,
		"deletion_events_deleted":   result.DeletionEventsDeleted,
		"retention_days":            result.RetentionDays,
		"cutoff":                    result.Cutoff,
		"disabled":                  result.Disabled,
	})
}

// RunIntegrityCheck runs an integrity check on the databases.
// POST /api/v1/storage/integrity-check
func (h *Handler) RunIntegrityCheck(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	startTime := time.Now()

	// Check app database
	appOK, appResult, err := h.db.RunAppIntegrityCheck(ctx)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to check app database integrity", "INTEGRITY_CHECK_FAILED")
		return
	}

	// Check relay database
	var relayOK bool
	var relayResult string

	writer, err := h.db.NewRelayWriter()
	if err != nil {
		relayOK = false
		relayResult = "Failed to open database"
	} else {
		defer writer.Close()
		relayOK, relayResult, err = writer.RunIntegrityCheck(ctx)
		if err != nil {
			relayOK = false
			relayResult = err.Error()
		}
	}

	duration := time.Since(startTime)

	// Update last integrity check timestamp
	h.db.SetLastIntegrityCheck(ctx, time.Now())

	// Add audit log
	h.db.AddAuditLog(ctx, "integrity_check", map[string]interface{}{
		"app_ok":       appOK,
		"relay_ok":     relayOK,
		"duration_ms":  duration.Milliseconds(),
	}, "")

	allOK := appOK && relayOK

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"success":     allOK,
		"app_db":      map[string]interface{}{"ok": appOK, "result": appResult},
		"relay_db":    map[string]interface{}{"ok": relayOK, "result": relayResult},
		"duration_ms": duration.Milliseconds(),
	})
}
