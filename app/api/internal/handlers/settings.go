package handlers

import (
	"encoding/json"
	"net/http"
	"time"
)

// GetTimezone returns the user's preferred timezone.
// GET /api/v1/settings/timezone
func (h *Handler) GetTimezone(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	tz, err := h.db.GetTimezone(ctx)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to get timezone", "TIMEZONE_FAILED")
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"timezone": tz})
}

// SetTimezone sets the user's preferred timezone.
// PUT /api/v1/settings/timezone
func (h *Handler) SetTimezone(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req struct {
		Timezone string `json:"timezone"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body", "INVALID_REQUEST")
		return
	}

	// Validate timezone
	if req.Timezone != "auto" && req.Timezone != "" {
		if _, err := time.LoadLocation(req.Timezone); err != nil {
			respondError(w, http.StatusBadRequest, "Invalid timezone", "INVALID_TIMEZONE")
			return
		}
	}

	// Default to "auto" if empty
	if req.Timezone == "" {
		req.Timezone = "auto"
	}

	if err := h.db.SetTimezone(ctx, req.Timezone); err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to save timezone", "TIMEZONE_FAILED")
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"timezone": req.Timezone})
}
