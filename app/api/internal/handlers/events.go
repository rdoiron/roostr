package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/roostr/roostr/app/api/internal/db"
)

// GetEvents returns a paginated list of events.
func (h *Handler) GetEvents(w http.ResponseWriter, r *http.Request) {
	if !h.db.IsRelayDBConnected() {
		respondError(w, http.StatusServiceUnavailable, "Relay database not connected", "RELAY_NOT_CONNECTED")
		return
	}

	// Parse query parameters
	query := r.URL.Query()

	filter := db.EventFilter{
		Limit:  parseIntParam(query.Get("limit"), 50),
		Offset: parseIntParam(query.Get("offset"), 0),
		Search: query.Get("search"),
	}

	// Parse kinds
	if kinds := query.Get("kinds"); kinds != "" {
		for _, k := range strings.Split(kinds, ",") {
			if kind, err := strconv.Atoi(strings.TrimSpace(k)); err == nil {
				filter.Kinds = append(filter.Kinds, kind)
			}
		}
	}

	// Parse authors
	if authors := query.Get("authors"); authors != "" {
		filter.Authors = strings.Split(authors, ",")
	}

	// Parse time range
	if since := query.Get("since"); since != "" {
		if ts, err := strconv.ParseInt(since, 10, 64); err == nil {
			filter.Since = time.Unix(ts, 0)
		}
	}
	if until := query.Get("until"); until != "" {
		if ts, err := strconv.ParseInt(until, 10, 64); err == nil {
			filter.Until = time.Unix(ts, 0)
		}
	}

	// Parse mentions filter (pubkey hex)
	if mentions := query.Get("mentions"); mentions != "" {
		filter.Mentions = mentions
	}

	events, err := h.db.GetEvents(r.Context(), filter)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to get events", "EVENTS_FETCH_FAILED")
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"events": events,
		"count":  len(events),
		"limit":  filter.Limit,
		"offset": filter.Offset,
	})
}

// GetEvent returns a single event by ID.
func (h *Handler) GetEvent(w http.ResponseWriter, r *http.Request) {
	if !h.db.IsRelayDBConnected() {
		respondError(w, http.StatusServiceUnavailable, "Relay database not connected", "RELAY_NOT_CONNECTED")
		return
	}

	id := r.PathValue("id")
	if id == "" {
		respondError(w, http.StatusBadRequest, "Event ID is required", "MISSING_ID")
		return
	}

	event, err := h.db.GetEvent(r.Context(), id)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to get event", "EVENT_FETCH_FAILED")
		return
	}

	if event == nil {
		respondError(w, http.StatusNotFound, "Event not found", "EVENT_NOT_FOUND")
		return
	}

	respondJSON(w, http.StatusOK, event)
}

// DeleteEvent queues an event for deletion.
func (h *Handler) DeleteEvent(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		respondError(w, http.StatusBadRequest, "Event ID is required", "MISSING_ID")
		return
	}

	// Verify event exists (good UX)
	if h.db.IsRelayDBConnected() {
		event, err := h.db.GetEvent(r.Context(), id)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "Failed to verify event", "EVENT_VERIFY_FAILED")
			return
		}
		if event == nil {
			respondError(w, http.StatusNotFound, "Event not found", "EVENT_NOT_FOUND")
			return
		}
	}

	// Parse optional reason from request body
	var reqBody struct {
		Reason string `json:"reason"`
	}
	json.NewDecoder(r.Body).Decode(&reqBody) // Ignore errors, reason is optional

	// Get operator pubkey for audit
	operatorPubkey, _ := h.db.GetAppState(r.Context(), "operator_pubkey")

	// Queue deletion request
	requestID, err := h.db.CreateDeletionRequest(r.Context(), id, operatorPubkey, reqBody.Reason)
	if err != nil {
		// Handle duplicate request
		if strings.Contains(err.Error(), "UNIQUE constraint") {
			respondError(w, http.StatusConflict, "Deletion already requested for this event", "DUPLICATE_REQUEST")
			return
		}
		respondError(w, http.StatusInternalServerError, "Failed to queue deletion", "DELETION_QUEUE_FAILED")
		return
	}

	// Return 202 Accepted
	respondJSON(w, http.StatusAccepted, map[string]interface{}{
		"message":    "Deletion request queued",
		"request_id": requestID,
		"event_id":   id,
		"status":     "pending",
	})
}

// GetRecentEvents returns the 10 most recent events for the dashboard.
func (h *Handler) GetRecentEvents(w http.ResponseWriter, r *http.Request) {
	if !h.db.IsRelayDBConnected() {
		respondJSON(w, http.StatusOK, map[string]interface{}{
			"events": []interface{}{},
		})
		return
	}

	events, err := h.db.GetRecentEvents(r.Context(), 10)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to get recent events", "EVENTS_FETCH_FAILED")
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"events": events,
	})
}

// parseIntParam parses an integer query parameter with a default value.
func parseIntParam(s string, defaultVal int) int {
	if s == "" {
		return defaultVal
	}
	if v, err := strconv.Atoi(s); err == nil {
		return v
	}
	return defaultVal
}
