package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/roostr/roostr/app/api/internal/db"
)

// ExportEvents handles GET /api/v1/events/export
// Streams events as NDJSON or JSON for backup/migration.
func (h *Handler) ExportEvents(w http.ResponseWriter, r *http.Request) {
	if !h.db.IsRelayDBConnected() {
		respondError(w, http.StatusServiceUnavailable, "Relay database not connected", "RELAY_NOT_CONNECTED")
		return
	}

	// Parse query parameters
	query := r.URL.Query()

	// Format: ndjson (default) or json
	format := query.Get("format")
	if format == "" {
		format = "ndjson"
	}
	if format != "ndjson" && format != "json" {
		respondError(w, http.StatusBadRequest, "Invalid format. Must be 'ndjson' or 'json'", "INVALID_FORMAT")
		return
	}

	// Build filter (no Limit/Offset for full export)
	filter := db.EventFilter{}

	// Parse kinds
	if kinds := query.Get("kinds"); kinds != "" {
		for _, k := range strings.Split(kinds, ",") {
			if kind, err := strconv.Atoi(strings.TrimSpace(k)); err == nil {
				filter.Kinds = append(filter.Kinds, kind)
			}
		}
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

	// Count total events for progress tracking
	count, err := h.db.CountEvents(r.Context(), filter)
	if err != nil {
		log.Printf("Warning: failed to count events for export: %v", err)
		// Continue without count header
	}

	// Parse timezone for filename date
	timezone := query.Get("timezone")
	loc := time.UTC
	if timezone != "" && timezone != "UTC" {
		if parsed, err := time.LoadLocation(timezone); err == nil {
			loc = parsed
		}
	}

	// Generate filename with current date in user's timezone
	filename := fmt.Sprintf("nostr-backup-%s", time.Now().In(loc).Format("2006-01-02"))
	if format == "ndjson" {
		filename += ".ndjson"
	} else {
		filename += ".json"
	}

	// Set response headers
	if format == "ndjson" {
		w.Header().Set("Content-Type", "application/x-ndjson")
	} else {
		w.Header().Set("Content-Type", "application/json")
	}
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))
	if count > 0 {
		w.Header().Set("X-Total-Count", strconv.FormatInt(count, 10))
	}

	// Get flusher for streaming
	flusher, ok := w.(http.Flusher)
	if !ok {
		respondError(w, http.StatusInternalServerError, "Streaming not supported", "STREAMING_NOT_SUPPORTED")
		return
	}

	// Write response based on format
	if format == "ndjson" {
		h.streamNDJSON(w, r, filter, flusher)
	} else {
		h.streamJSON(w, r, filter, flusher)
	}
}

// streamNDJSON writes events as newline-delimited JSON.
func (h *Handler) streamNDJSON(w http.ResponseWriter, r *http.Request, filter db.EventFilter, flusher http.Flusher) {
	eventCount := 0

	err := h.db.StreamEvents(r.Context(), filter, func(event db.ExportEvent) error {
		// Encode event to JSON
		data, err := json.Marshal(event)
		if err != nil {
			return fmt.Errorf("failed to marshal event: %w", err)
		}

		// Write JSON line
		if _, err := w.Write(data); err != nil {
			return err
		}
		if _, err := w.Write([]byte("\n")); err != nil {
			return err
		}

		eventCount++
		// Flush every 100 events for responsive streaming
		if eventCount%100 == 0 {
			flusher.Flush()
		}

		return nil
	})

	if err != nil {
		// Can't send error response after headers are written
		// Log it and stop
		log.Printf("Export stream error: %v", err)
	}

	// Final flush
	flusher.Flush()
}

// GetExportEstimate handles GET /api/v1/events/export/estimate
// Returns the estimated event count and byte size for an export with the given filters.
func (h *Handler) GetExportEstimate(w http.ResponseWriter, r *http.Request) {
	if !h.db.IsRelayDBConnected() {
		respondError(w, http.StatusServiceUnavailable, "Relay database not connected", "RELAY_NOT_CONNECTED")
		return
	}

	// Parse query parameters
	query := r.URL.Query()

	// Build filter
	filter := db.EventFilter{}

	// Parse kinds
	if kinds := query.Get("kinds"); kinds != "" {
		for _, k := range strings.Split(kinds, ",") {
			if kind, err := strconv.Atoi(strings.TrimSpace(k)); err == nil {
				filter.Kinds = append(filter.Kinds, kind)
			}
		}
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

	// Count events
	count, err := h.db.CountEvents(r.Context(), filter)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to count events", "COUNT_FAILED")
		return
	}

	// Estimate size: ~500 bytes per event (average)
	estimatedBytes := count * 500

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"count":           count,
		"estimated_bytes": estimatedBytes,
	})
}

// streamJSON writes events as a JSON array.
func (h *Handler) streamJSON(w http.ResponseWriter, r *http.Request, filter db.EventFilter, flusher http.Flusher) {
	// Write opening bracket
	if _, err := w.Write([]byte("[\n")); err != nil {
		log.Printf("Export stream error: %v", err)
		return
	}

	first := true
	eventCount := 0

	err := h.db.StreamEvents(r.Context(), filter, func(event db.ExportEvent) error {
		// Encode event to JSON
		data, err := json.Marshal(event)
		if err != nil {
			return fmt.Errorf("failed to marshal event: %w", err)
		}

		// Write comma separator for all but first event
		if !first {
			if _, err := w.Write([]byte(",\n")); err != nil {
				return err
			}
		}
		first = false

		// Write JSON event
		if _, err := w.Write(data); err != nil {
			return err
		}

		eventCount++
		// Flush every 100 events
		if eventCount%100 == 0 {
			flusher.Flush()
		}

		return nil
	})

	if err != nil {
		log.Printf("Export stream error: %v", err)
	}

	// Write closing bracket
	if _, err := w.Write([]byte("\n]")); err != nil {
		log.Printf("Export stream error: %v", err)
	}

	// Final flush
	flusher.Flush()
}
