package handlers

import (
	"context"
	"encoding/json"
	"fmt"
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

	// Use hourly buckets for "today" view
	hourly := timeRange == "today"

	data, err := h.db.GetEventsOverTime(ctx, since, until, hourly)
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

// StreamDashboardStats streams dashboard statistics in real-time via Server-Sent Events (SSE).
// GET /api/v1/stats/stream
func (h *Handler) StreamDashboardStats(w http.ResponseWriter, r *http.Request) {
	// Set SSE headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no") // Disable nginx buffering

	// Ensure we can flush
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "SSE not supported", http.StatusInternalServerError)
		return
	}

	// Disable write timeout for this long-lived SSE connection
	rc := http.NewResponseController(w)
	rc.SetWriteDeadline(time.Time{}) // No deadline

	ctx := r.Context()

	// Send initial connection event
	fmt.Fprintf(w, "event: connected\ndata: {\"status\": \"connected\"}\n\n")
	flusher.Flush()

	// Ticker for pushing updates (every 2 seconds)
	updateTicker := time.NewTicker(2 * time.Second)
	defer updateTicker.Stop()

	// Keepalive ticker (every 15 seconds)
	keepaliveTicker := time.NewTicker(15 * time.Second)
	defer keepaliveTicker.Stop()

	// Send initial data immediately
	h.sendDashboardUpdate(w, flusher, ctx)

	for {
		select {
		case <-ctx.Done():
			return
		case <-updateTicker.C:
			h.sendDashboardUpdate(w, flusher, ctx)
		case <-keepaliveTicker.C:
			// Send keepalive comment
			fmt.Fprintf(w, ": keepalive\n\n")
			flusher.Flush()
		}
	}
}

// sendDashboardUpdate gathers and sends current dashboard stats.
func (h *Handler) sendDashboardUpdate(w http.ResponseWriter, flusher http.Flusher, ctx context.Context) {
	// Build stats data (same logic as GetStatsSummary)
	relayConnected := h.db.IsRelayDBConnected()
	relayStatus := "offline"
	if relayConnected {
		relayStatus = "online"
	}

	uptimeSeconds := int64(time.Since(h.startTime).Seconds())
	whitelistCount, _ := h.db.GetWhitelistCount(ctx)
	storageBytes, _ := h.db.GetRelayDatabaseSize()

	var stats map[string]interface{}
	var recentEvents interface{}
	var storage map[string]interface{}

	if !relayConnected {
		stats = map[string]interface{}{
			"total_events":      0,
			"events_today":      0,
			"storage_bytes":     storageBytes,
			"whitelisted_count": whitelistCount,
			"events_by_kind":    map[string]int64{},
			"uptime_seconds":    uptimeSeconds,
			"relay_status":      relayStatus,
		}
		recentEvents = []interface{}{}
	} else {
		// Get relay stats
		relayStats, err := h.db.GetRelayStats(ctx)
		if err != nil {
			// On error, send empty stats instead of returning
			stats = map[string]interface{}{
				"total_events":      0,
				"events_today":      0,
				"storage_bytes":     storageBytes,
				"whitelisted_count": whitelistCount,
				"events_by_kind":    map[string]int64{},
				"uptime_seconds":    uptimeSeconds,
				"relay_status":      relayStatus,
			}
			recentEvents = []interface{}{}
		} else {
			eventsToday, _ := h.db.GetEventsToday(ctx)

			// Build events_by_kind
			eventsByKind := map[string]int64{}
			var otherCount int64 = 0
			for kind, count := range relayStats.EventsByKind {
				switch kind {
				case 1:
					eventsByKind["posts"] = count
				case 3:
					eventsByKind["follows"] = count
				case 4, 14:
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

			stats = map[string]interface{}{
				"total_events":      relayStats.TotalEvents,
				"events_today":      eventsToday,
				"storage_bytes":     storageBytes,
				"whitelisted_count": whitelistCount,
				"events_by_kind":    eventsByKind,
				"uptime_seconds":    uptimeSeconds,
				"relay_status":      relayStatus,
			}

			// Get recent events
			events, err := h.db.GetRecentEvents(ctx, 10)
			if err != nil {
				recentEvents = []interface{}{}
			} else {
				recentEvents = events
			}
		}
	}

	// Get storage status
	relayDBSize, _ := h.db.GetRelayDatabaseSize()
	appDBSize, _ := h.db.GetAppDatabaseSize()
	totalSize := relayDBSize + appDBSize
	availableSpace, _ := h.db.GetAvailableDiskSpace()
	totalSpace, _ := h.db.GetTotalDiskSpace()

	var usagePercent float64
	if totalSpace > 0 {
		usedSpace := totalSpace - availableSpace
		usagePercent = float64(usedSpace) / float64(totalSpace) * 100
	}

	// Determine storage status
	storageStatus := "healthy"
	if usagePercent > 90 {
		storageStatus = "critical"
	} else if usagePercent > 80 {
		storageStatus = "low"
	} else if usagePercent > 70 {
		storageStatus = "warning"
	}

	storage = map[string]interface{}{
		"database_size":     relayDBSize,
		"app_database_size": appDBSize,
		"total_size":        totalSize,
		"available_space":   availableSpace,
		"total_space":       totalSpace,
		"usage_percent":     usagePercent,
		"status":            storageStatus,
	}

	// Build and send update
	update := map[string]interface{}{
		"stats":        stats,
		"recentEvents": recentEvents,
		"storage":      storage,
	}

	data, err := json.Marshal(update)
	if err != nil {
		return
	}

	fmt.Fprintf(w, "event: stats\ndata: %s\n\n", data)
	flusher.Flush()
}
