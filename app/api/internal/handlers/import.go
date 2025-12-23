package handlers

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/roostr/roostr/app/api/internal/db"
	"github.com/roostr/roostr/app/api/internal/nostr"
)

// ImportEventsRequest defines options for the import operation.
type ImportEventsRequest struct {
	VerifySignatures bool `json:"verify_signatures"` // Default: true
	SkipDuplicates   bool `json:"skip_duplicates"`   // Default: true
	StopOnError      bool `json:"stop_on_error"`     // Default: false
}

// ImportEventsResponse contains the results of an import operation.
type ImportEventsResponse struct {
	Total      int      `json:"total"`       // Total events in file
	Processed  int      `json:"processed"`   // Events attempted
	Added      int      `json:"added"`       // Successfully inserted (new)
	Duplicates int      `json:"duplicates"`  // Already existed
	Errors     int      `json:"errors"`      // Failed to insert
	ErrorList  []string `json:"error_list"`  // Error messages (limited to first 100)
}

// ImportEvents handles POST /api/v1/events/import
// Accepts NDJSON or JSON array format file uploads.
// Compatible with exports from strfry, nosdump, nostrudel, and other Nostr tools.
func (h *Handler) ImportEvents(w http.ResponseWriter, r *http.Request) {
	if !h.db.IsRelayDBConnected() {
		respondError(w, http.StatusServiceUnavailable, "Relay database not connected", "RELAY_NOT_CONNECTED")
		return
	}

	// Parse multipart form (max 500MB file)
	if err := r.ParseMultipartForm(500 << 20); err != nil {
		respondError(w, http.StatusBadRequest, "Failed to parse form data", "INVALID_FORM")
		return
	}

	// Get the uploaded file
	file, header, err := r.FormFile("file")
	if err != nil {
		respondError(w, http.StatusBadRequest, "No file provided", "MISSING_FILE")
		return
	}
	defer file.Close()

	log.Printf("Importing events from file: %s (%d bytes)", header.Filename, header.Size)

	// Parse options from form data
	options := ImportEventsRequest{
		VerifySignatures: true,  // Default: verify
		SkipDuplicates:   true,  // Default: skip
		StopOnError:      false, // Default: continue
	}

	if verifyStr := r.FormValue("verify_signatures"); verifyStr != "" {
		options.VerifySignatures = verifyStr == "true"
	}
	if skipStr := r.FormValue("skip_duplicates"); skipStr != "" {
		options.SkipDuplicates = skipStr == "true"
	}
	if stopStr := r.FormValue("stop_on_error"); stopStr != "" {
		options.StopOnError = stopStr == "true"
	}

	// Read the entire file into memory
	// For very large files, this could be streamed, but keeping it simple for now
	data, err := io.ReadAll(file)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to read file", "READ_ERROR")
		return
	}

	// Detect format by finding first non-whitespace character
	format := detectFormat(data)
	log.Printf("Detected format: %s", format)

	// Parse events based on format
	var events []*nostr.SyncEvent
	if format == "json" {
		events, err = parseJSONArray(data)
	} else {
		events, err = parseNDJSON(data)
	}

	if err != nil {
		respondError(w, http.StatusBadRequest, fmt.Sprintf("Failed to parse events: %v", err), "PARSE_ERROR")
		return
	}

	// Create a relay writer for inserting events
	writer, err := h.db.NewRelayWriter()
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to open database for writing", "DB_WRITE_ERROR")
		return
	}
	defer writer.Close()

	// Import events
	response := h.importEvents(r.Context(), writer, events, options)

	log.Printf("Import complete: %d total, %d added, %d duplicates, %d errors",
		response.Total, response.Added, response.Duplicates, response.Errors)

	respondJSON(w, http.StatusOK, response)
}

// detectFormat determines if the data is NDJSON or JSON array.
func detectFormat(data []byte) string {
	// Find first non-whitespace character
	for _, b := range data {
		if b == ' ' || b == '\t' || b == '\n' || b == '\r' {
			continue
		}
		if b == '[' {
			return "json"
		}
		if b == '{' {
			return "ndjson"
		}
		break
	}
	return "ndjson" // Default
}

// parseJSONArray parses events from a JSON array.
func parseJSONArray(data []byte) ([]*nostr.SyncEvent, error) {
	var events []*nostr.SyncEvent
	if err := json.Unmarshal(data, &events); err != nil {
		return nil, fmt.Errorf("invalid JSON array: %w", err)
	}
	return events, nil
}

// parseNDJSON parses events from newline-delimited JSON.
func parseNDJSON(data []byte) ([]*nostr.SyncEvent, error) {
	var events []*nostr.SyncEvent
	scanner := bufio.NewScanner(bytes.NewReader(data))

	// Increase buffer size for large events
	const maxScanTokenSize = 1024 * 1024 // 1MB per line
	buf := make([]byte, maxScanTokenSize)
	scanner.Buffer(buf, maxScanTokenSize)

	lineNum := 0
	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue // Skip empty lines
		}

		var event nostr.SyncEvent
		if err := json.Unmarshal([]byte(line), &event); err != nil {
			return nil, fmt.Errorf("line %d: invalid JSON: %w", lineNum, err)
		}
		events = append(events, &event)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading file: %w", err)
	}

	return events, nil
}

// importEvents processes and inserts events into the database.
func (h *Handler) importEvents(ctx context.Context, writer *db.RelayWriter, events []*nostr.SyncEvent, options ImportEventsRequest) ImportEventsResponse {
	response := ImportEventsResponse{
		Total:     len(events),
		ErrorList: make([]string, 0),
	}

	for i, event := range events {
		response.Processed++

		// Verify event if requested
		if options.VerifySignatures {
			if err := event.Verify(); err != nil {
				errMsg := fmt.Sprintf("Event %d: verification failed: %v", i+1, err)
				response.Errors++
				if len(response.ErrorList) < 100 {
					response.ErrorList = append(response.ErrorList, errMsg)
				}
				if options.StopOnError {
					break
				}
				continue
			}
		}

		// Convert to DB format
		dbEvent := &db.Event{
			ID:        event.ID,
			Pubkey:    event.Pubkey,
			CreatedAt: time.Unix(event.CreatedAt, 0),
			Kind:      event.Kind,
			Tags:      event.Tags,
			Content:   event.Content,
			Sig:       event.Sig,
		}

		// Insert event
		inserted, err := writer.InsertEvent(ctx, dbEvent)
		if err != nil {
			errMsg := fmt.Sprintf("Event %d: insert failed: %v", i+1, err)
			response.Errors++
			if len(response.ErrorList) < 100 {
				response.ErrorList = append(response.ErrorList, errMsg)
			}
			if options.StopOnError {
				break
			}
			continue
		}

		if inserted {
			response.Added++
		} else {
			response.Duplicates++
		}
	}

	return response
}
