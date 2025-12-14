package handlers

import (
	"bufio"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

// LogEntry represents a parsed relay log entry.
type LogEntry struct {
	Timestamp string `json:"timestamp"`
	Level     string `json:"level"`
	Message   string `json:"message"`
}

// ReloadRelay signals the relay to reload its configuration.
// POST /api/v1/relay/reload
func (h *Handler) ReloadRelay(w http.ResponseWriter, r *http.Request) {
	if h.relay == nil {
		respondError(w, http.StatusServiceUnavailable, "Relay manager not available", "RELAY_NOT_AVAILABLE")
		return
	}

	if err := h.relay.Reload(); err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to reload relay", "RELOAD_FAILED")
		return
	}

	ctx := r.Context()
	h.db.AddAuditLog(ctx, "relay_reloaded", nil, "")

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "Relay configuration reloaded",
	})
}

// RestartRelay initiates an async restart of the relay process.
// POST /api/v1/relay/restart
func (h *Handler) RestartRelay(w http.ResponseWriter, r *http.Request) {
	if h.relay == nil {
		respondError(w, http.StatusServiceUnavailable, "Relay manager not available", "RELAY_NOT_AVAILABLE")
		return
	}

	if h.relay.IsRestarting() {
		respondError(w, http.StatusConflict, "Relay restart already in progress", "RESTART_IN_PROGRESS")
		return
	}

	if err := h.relay.Restart(); err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to initiate relay restart", "RESTART_FAILED")
		return
	}

	ctx := r.Context()
	h.db.AddAuditLog(ctx, "relay_restart_initiated", nil, "")

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "Relay restart initiated",
		"status":  "restarting",
	})
}

// GetRelayLogs returns recent log entries from the relay log file.
// GET /api/v1/relay/logs
func (h *Handler) GetRelayLogs(w http.ResponseWriter, r *http.Request) {
	// Parse limit parameter (default 100, max 1000)
	limit := 100
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if parsed, err := strconv.Atoi(limitStr); err == nil && parsed > 0 {
			limit = parsed
			if limit > 1000 {
				limit = 1000
			}
		}
	}

	// Get log file path from config
	logPath, err := h.getRelayLogPath()
	if err != nil {
		respondError(w, http.StatusServiceUnavailable, "Could not determine log file path", "LOG_PATH_ERROR")
		return
	}

	if logPath == "" {
		respondJSON(w, http.StatusOK, map[string]interface{}{
			"logs":        []LogEntry{},
			"total_lines": 0,
			"message":     "Relay logging not configured",
		})
		return
	}

	// Read and parse log entries
	entries, err := readLogFile(logPath, limit)
	if err != nil {
		// Log file may not exist yet or be inaccessible
		respondJSON(w, http.StatusOK, map[string]interface{}{
			"logs":        []LogEntry{},
			"total_lines": 0,
			"message":     "Log file not available: " + err.Error(),
		})
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"logs":        entries,
		"total_lines": len(entries),
	})
}

// getRelayLogPath determines the path to the relay's log file.
func (h *Handler) getRelayLogPath() (string, error) {
	if h.configMgr == nil {
		return "", nil
	}

	cfg, err := h.configMgr.Read()
	if err != nil {
		return "", err
	}

	if cfg.Logging.FolderPath == "" {
		return "", nil
	}

	// Find the most recent log file in the folder
	return findMostRecentLogFile(cfg.Logging.FolderPath, cfg.Logging.FilePrefix)
}

// findMostRecentLogFile finds the most recent log file matching the prefix.
func findMostRecentLogFile(folder, prefix string) (string, error) {
	entries, err := os.ReadDir(folder)
	if err != nil {
		return "", err
	}

	var logFiles []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		// Match files starting with prefix or ending with .log
		if (prefix != "" && strings.HasPrefix(name, prefix)) || strings.HasSuffix(name, ".log") {
			logFiles = append(logFiles, filepath.Join(folder, name))
		}
	}

	if len(logFiles) == 0 {
		return "", nil
	}

	// Sort by modification time (newest first)
	sort.Slice(logFiles, func(i, j int) bool {
		infoI, _ := os.Stat(logFiles[i])
		infoJ, _ := os.Stat(logFiles[j])
		if infoI == nil || infoJ == nil {
			return false
		}
		return infoI.ModTime().After(infoJ.ModTime())
	})

	return logFiles[0], nil
}

// readLogFile reads the last N lines from a log file and parses them.
func readLogFile(path string, limit int) ([]LogEntry, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// Read all lines into memory (for simplicity with small log files)
	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	// Get last N lines
	startIdx := len(lines) - limit
	if startIdx < 0 {
		startIdx = 0
	}
	recentLines := lines[startIdx:]

	// Parse lines into LogEntry structs
	entries := make([]LogEntry, 0, len(recentLines))
	for _, line := range recentLines {
		entry := parseLogLine(line)
		entries = append(entries, entry)
	}

	// Reverse so newest is first
	for i, j := 0, len(entries)-1; i < j; i, j = i+1, j-1 {
		entries[i], entries[j] = entries[j], entries[i]
	}

	return entries, nil
}

// logLineRegex matches common log formats like:
// [2025-12-13T14:32:01Z INFO nostr_rs_relay] Message here
// 2025-12-13 14:32:01 INFO Message here
var logLineRegex = regexp.MustCompile(`^\[?(\d{4}-\d{2}-\d{2}[T ]\d{2}:\d{2}:\d{2}[Z]?)\]?\s*(\w+)\s*(.*)$`)

// parseLogLine parses a log line into a LogEntry struct.
func parseLogLine(line string) LogEntry {
	matches := logLineRegex.FindStringSubmatch(line)
	if len(matches) >= 4 {
		return LogEntry{
			Timestamp: matches[1],
			Level:     strings.ToUpper(matches[2]),
			Message:   strings.TrimSpace(matches[3]),
		}
	}

	// Fallback: return the whole line as the message
	return LogEntry{
		Timestamp: "",
		Level:     "INFO",
		Message:   line,
	}
}

// StreamRelayLogs streams relay logs in real-time via Server-Sent Events (SSE).
// GET /api/v1/relay/logs/stream
func (h *Handler) StreamRelayLogs(w http.ResponseWriter, r *http.Request) {
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

	// Get log file path
	logPath, err := h.getRelayLogPath()
	if err != nil || logPath == "" {
		fmt.Fprintf(w, "event: error\ndata: {\"error\": \"Log file not configured\"}\n\n")
		flusher.Flush()
		return
	}

	// Open log file
	file, err := os.Open(logPath)
	if err != nil {
		fmt.Fprintf(w, "event: error\ndata: {\"error\": \"Cannot open log file\"}\n\n")
		flusher.Flush()
		return
	}
	defer file.Close()

	// Seek to end of file
	file.Seek(0, 2)

	// Track position
	lastPos, _ := file.Seek(0, 1)

	// Send initial connection event
	fmt.Fprintf(w, "event: connected\ndata: {\"status\": \"connected\"}\n\n")
	flusher.Flush()

	// Ticker for checking new log entries
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	// Use request context to detect client disconnect
	ctx := r.Context()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			// Check for new content
			info, err := file.Stat()
			if err != nil {
				continue
			}

			currentSize := info.Size()
			if currentSize > lastPos {
				// Read new content
				file.Seek(lastPos, 0)
				scanner := bufio.NewScanner(file)
				for scanner.Scan() {
					line := scanner.Text()
					entry := parseLogLine(line)
					// Send as SSE event
					fmt.Fprintf(w, "event: log\ndata: {\"timestamp\":%q,\"level\":%q,\"message\":%q}\n\n",
						entry.Timestamp, entry.Level, entry.Message)
					flusher.Flush()
				}
				lastPos, _ = file.Seek(0, 1)
			} else if currentSize < lastPos {
				// File was truncated/rotated - seek to beginning
				lastPos = 0
				file.Seek(0, 0)
			}

			// Send keepalive comment every tick to detect disconnects
			fmt.Fprintf(w, ": keepalive\n\n")
			flusher.Flush()
		}
	}
}
