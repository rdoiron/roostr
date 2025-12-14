package handlers

import (
	"net/http"
	"time"
)

// GetStatsSummary returns aggregate statistics from the relay for the dashboard.
func (h *Handler) GetStatsSummary(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Determine relay status
	relayConnected := h.db.IsRelayDBConnected()
	relayStatus := "offline"
	if relayConnected {
		relayStatus = "online"
	}

	// Get uptime
	uptimeSeconds := int64(time.Since(h.startTime).Seconds())

	// Get whitelist count
	whitelistCount, _ := h.db.GetWhitelistCount(ctx)

	// Get storage size
	storageBytes, _ := h.db.GetRelayDatabaseSize()

	// If relay DB is not connected, return minimal data
	if !relayConnected {
		respondJSON(w, http.StatusOK, map[string]interface{}{
			"total_events":      0,
			"events_today":      0,
			"storage_bytes":     storageBytes,
			"whitelisted_count": whitelistCount,
			"events_by_kind":    map[string]int64{},
			"uptime_seconds":    uptimeSeconds,
			"relay_status":      relayStatus,
		})
		return
	}

	// Get relay stats
	stats, err := h.db.GetRelayStats(ctx)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to get stats", "STATS_FAILED")
		return
	}

	// Get events today
	eventsToday, _ := h.db.GetEventsToday(ctx)

	// Build events_by_kind with human-friendly labels per spec
	eventsByKind := map[string]int64{}
	var otherCount int64 = 0

	for kind, count := range stats.EventsByKind {
		switch kind {
		case 1:
			eventsByKind["posts"] = count
		case 3:
			eventsByKind["follows"] = count
		case 4, 14:
			// DMs (kind 4 and 14)
			eventsByKind["dms"] = eventsByKind["dms"] + count
		case 6:
			eventsByKind["reposts"] = count
		case 7:
			eventsByKind["reactions"] = count
		default:
			otherCount += count
		}
	}
	if otherCount > 0 {
		eventsByKind["other"] = otherCount
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"total_events":      stats.TotalEvents,
		"events_today":      eventsToday,
		"storage_bytes":     storageBytes,
		"whitelisted_count": whitelistCount,
		"events_by_kind":    eventsByKind,
		"uptime_seconds":    uptimeSeconds,
		"relay_status":      relayStatus,
	})
}

// GetRelayStatus returns the current status of the relay with detailed process information.
func (h *Handler) GetRelayStatus(w http.ResponseWriter, r *http.Request) {
	relayConnected := h.db.IsRelayDBConnected()
	apiUptimeSeconds := int64(time.Since(h.startTime).Seconds())

	// Determine relay status
	var status string
	var pid int
	var memoryBytes int64
	var relayUptimeSeconds int64

	if h.relay != nil {
		if h.relay.IsRestarting() {
			status = "restarting"
		} else if h.relay.IsRunning() {
			status = "running"
			pid = h.relay.GetPID()
			memoryBytes = h.relay.GetMemoryUsage()
			relayUptimeSeconds = h.relay.GetProcessUptime()
		} else {
			status = "stopped"
		}
	} else {
		// No relay manager - fall back to database connection check
		if relayConnected {
			status = "running"
		} else {
			status = "unknown"
		}
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"status":             status,
		"pid":                pid,
		"memory_bytes":       memoryBytes,
		"uptime_seconds":     relayUptimeSeconds,
		"database_connected": relayConnected,
		"api_uptime_seconds": apiUptimeSeconds,
	})
}

// GetRelayURLs returns the relay's local and Tor WebSocket URLs.
func (h *Handler) GetRelayURLs(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"local":         h.cfg.RelayURL,
		"tor":           "",
		"tor_available": false,
	}

	if h.cfg.TorAddress != "" {
		// Format Tor URL - TOR_ADDRESS includes the port
		response["tor"] = "ws://" + h.cfg.TorAddress
		response["tor_available"] = true
	}

	respondJSON(w, http.StatusOK, response)
}
