package handlers

import (
	"net/http"
	"strconv"
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
		"relay_port":    h.cfg.RelayPort,
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

// GetEventsOverTime returns event counts grouped by date for charting.
// STATS-API-001: GET /api/v1/stats/events-over-time
func (h *Handler) GetEventsOverTime(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse time_range query parameter first
	timeRange := r.URL.Query().Get("time_range")
	if timeRange == "" {
		timeRange = "7days"
	}

	if !h.db.IsRelayDBConnected() {
		respondJSON(w, http.StatusOK, map[string]interface{}{
			"data":       []interface{}{},
			"time_range": timeRange,
			"total":      0,
		})
		return
	}

	since, until := parseTimeRange(timeRange)

	data, err := h.db.GetEventsOverTime(ctx, since, until)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to get events over time", "STATS_FAILED")
		return
	}

	// Calculate total
	var total int64
	for _, d := range data {
		total += d.Count
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"data":       data,
		"time_range": timeRange,
		"total":      total,
	})
}

// GetEventsByKind returns event distribution by kind for charting.
// STATS-API-002: GET /api/v1/stats/events-by-kind
func (h *Handler) GetEventsByKind(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse time_range query parameter first
	timeRange := r.URL.Query().Get("time_range")
	if timeRange == "" {
		timeRange = "alltime"
	}

	if !h.db.IsRelayDBConnected() {
		respondJSON(w, http.StatusOK, map[string]interface{}{
			"kinds":      []interface{}{},
			"time_range": timeRange,
			"total":      0,
		})
		return
	}

	since, until := parseTimeRange(timeRange)

	kindCounts, err := h.db.GetEventsByKindInRange(ctx, since, until)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to get events by kind", "STATS_FAILED")
		return
	}

	// Calculate total
	var total int64
	for _, count := range kindCounts {
		total += count
	}

	// Build response with labels and percentages
	type kindInfo struct {
		Kind    int     `json:"kind"`
		Label   string  `json:"label"`
		Count   int64   `json:"count"`
		Percent float64 `json:"percent"`
	}

	var kinds []kindInfo
	for kind, count := range kindCounts {
		label := getKindLabel(kind)
		percent := 0.0
		if total > 0 {
			percent = float64(count) / float64(total) * 100
		}
		kinds = append(kinds, kindInfo{
			Kind:    kind,
			Label:   label,
			Count:   count,
			Percent: percent,
		})
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"kinds":      kinds,
		"time_range": timeRange,
		"total":      total,
	})
}

// GetTopAuthors returns the most active pubkeys by event count.
// STATS-API-003: GET /api/v1/stats/top-authors
func (h *Handler) GetTopAuthors(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse limit query parameter first
	limit := 10
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if parsed, err := strconv.Atoi(limitStr); err == nil && parsed > 0 {
			limit = parsed
			if limit > 100 {
				limit = 100
			}
		}
	}

	// Parse time_range query parameter
	timeRange := r.URL.Query().Get("time_range")
	if timeRange == "" {
		timeRange = "alltime"
	}

	if !h.db.IsRelayDBConnected() {
		respondJSON(w, http.StatusOK, map[string]interface{}{
			"authors":    []interface{}{},
			"time_range": timeRange,
			"limit":      limit,
		})
		return
	}

	since, until := parseTimeRange(timeRange)

	authors, err := h.db.GetTopAuthorsInRange(ctx, limit, since, until)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to get top authors", "STATS_FAILED")
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"authors":    authors,
		"time_range": timeRange,
		"limit":      limit,
	})
}

// parseTimeRange converts a time range string to since/until timestamps.
func parseTimeRange(rangeStr string) (since, until time.Time) {
	now := time.Now().UTC()
	until = now

	switch rangeStr {
	case "today":
		since = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	case "7days":
		since = now.AddDate(0, 0, -7)
	case "30days":
		since = now.AddDate(0, 0, -30)
	case "alltime":
		since = time.Time{} // Zero value - no filter
	default:
		since = now.AddDate(0, 0, -7) // Default to 7 days
	}

	return since, until
}

// getKindLabel returns a human-friendly label for a Nostr event kind.
func getKindLabel(kind int) string {
	switch kind {
	case 0:
		return "profiles"
	case 1:
		return "posts"
	case 3:
		return "follows"
	case 4, 14:
		return "dms"
	case 6:
		return "reposts"
	case 7:
		return "reactions"
	default:
		return "other"
	}
}
