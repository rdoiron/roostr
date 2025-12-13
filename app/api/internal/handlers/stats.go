package handlers

import (
	"net/http"
)

// GetStatsSummary returns aggregate statistics from the relay.
func (h *Handler) GetStatsSummary(w http.ResponseWriter, r *http.Request) {
	if !h.db.IsRelayDBConnected() {
		respondJSON(w, http.StatusOK, map[string]interface{}{
			"relay_connected": false,
			"message":         "Relay database not connected",
		})
		return
	}

	stats, err := h.db.GetRelayStats(r.Context())
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to get stats", "STATS_FAILED")
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"relay_connected": true,
		"total_events":    stats.TotalEvents,
		"total_pubkeys":   stats.TotalPubkeys,
		"events_by_kind":  stats.EventsByKind,
		"oldest_event":    stats.OldestEvent,
		"newest_event":    stats.NewestEvent,
	})
}

// GetRelayStatus returns the current status of the relay.
func (h *Handler) GetRelayStatus(w http.ResponseWriter, r *http.Request) {
	relayConnected := h.db.IsRelayDBConnected()

	// TODO: Check actual relay process status
	status := "unknown"
	if relayConnected {
		status = "running"
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"status":           status,
		"database_connected": relayConnected,
	})
}
