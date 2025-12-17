// Package relay provides control over the nostr-rs-relay process.
// This includes starting, stopping, reloading config, and monitoring status.
package relay

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
)

// LogEntry represents a parsed log line from the relay.
type LogEntry struct {
	Timestamp string `json:"timestamp"`
	Level     string `json:"level"`
	Message   string `json:"message"`
}

// LogBuffer is a thread-safe ring buffer for log entries with pub/sub support.
type LogBuffer struct {
	entries     []LogEntry
	mu          sync.RWMutex
	maxSize     int
	subscribers map[chan LogEntry]struct{}
	subMu       sync.RWMutex
}

// NewLogBuffer creates a new LogBuffer with the specified max size.
func NewLogBuffer(maxSize int) *LogBuffer {
	return &LogBuffer{
		entries:     make([]LogEntry, 0, maxSize),
		maxSize:     maxSize,
		subscribers: make(map[chan LogEntry]struct{}),
	}
}

// Add adds a log entry to the buffer and notifies subscribers.
func (lb *LogBuffer) Add(entry LogEntry) {
	lb.mu.Lock()
	if len(lb.entries) >= lb.maxSize {
		// Remove oldest entry
		lb.entries = lb.entries[1:]
	}
	lb.entries = append(lb.entries, entry)
	lb.mu.Unlock()

	// Notify subscribers (non-blocking)
	lb.subMu.RLock()
	for ch := range lb.subscribers {
		select {
		case ch <- entry:
		default:
			// Skip if subscriber is not ready
		}
	}
	lb.subMu.RUnlock()
}

// GetRecent returns the most recent log entries up to the limit.
func (lb *LogBuffer) GetRecent(limit int) []LogEntry {
	lb.mu.RLock()
	defer lb.mu.RUnlock()

	if limit <= 0 || limit > len(lb.entries) {
		limit = len(lb.entries)
	}

	// Return entries in reverse order (newest first)
	result := make([]LogEntry, limit)
	for i := 0; i < limit; i++ {
		result[i] = lb.entries[len(lb.entries)-1-i]
	}
	return result
}

// Subscribe returns a channel that receives new log entries.
func (lb *LogBuffer) Subscribe() chan LogEntry {
	ch := make(chan LogEntry, 100)
	lb.subMu.Lock()
	lb.subscribers[ch] = struct{}{}
	lb.subMu.Unlock()
	return ch
}

// Unsubscribe removes a subscriber channel.
func (lb *LogBuffer) Unsubscribe(ch chan LogEntry) {
	lb.subMu.Lock()
	delete(lb.subscribers, ch)
	lb.subMu.Unlock()
	close(ch)
}

// Relay manages the nostr-rs-relay process.
type Relay struct {
	BinaryPath string
	ConfigPath string
	cmd        *exec.Cmd
	logBuffer  *LogBuffer

	mu         sync.RWMutex
	restarting bool
}

// New creates a new Relay instance.
func New(binaryPath, configPath string) *Relay {
	return &Relay{
		BinaryPath: binaryPath,
		ConfigPath: configPath,
		logBuffer:  NewLogBuffer(1000), // Keep last 1000 log entries
	}
}

// IsRestarting returns true if a restart is currently in progress.
func (r *Relay) IsRestarting() bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.restarting
}

// setRestarting sets the restarting state.
func (r *Relay) setRestarting(val bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.restarting = val
}

// Start starts the relay process.
func (r *Relay) Start() error {
	if r.BinaryPath == "" {
		return fmt.Errorf("relay binary path not configured")
	}

	// Check if already running
	if r.IsRunning() {
		return nil // Already running
	}

	// Build command with config path if provided
	args := []string{}
	if r.ConfigPath != "" {
		args = append(args, "--config", r.ConfigPath)
	}

	r.cmd = exec.Command(r.BinaryPath, args...)

	// Capture stdout and stderr for log buffer
	stdoutPipe, err := r.cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdout pipe: %w", err)
	}
	stderrPipe, err := r.cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("failed to create stderr pipe: %w", err)
	}

	if err := r.cmd.Start(); err != nil {
		return fmt.Errorf("failed to start relay: %w", err)
	}

	// Start goroutines to capture output
	go r.captureOutput(stdoutPipe)
	go r.captureOutput(stderrPipe)

	return nil
}

// captureOutput reads from a pipe and stores log entries in the buffer.
func (r *Relay) captureOutput(pipe io.ReadCloser) {
	scanner := bufio.NewScanner(pipe)
	for scanner.Scan() {
		line := scanner.Text()
		// Also print to stdout for debugging
		fmt.Println(line)
		// Parse and store in buffer
		entry := parseLogLine(line)
		r.logBuffer.Add(entry)
	}
}

// logLineRegex matches nostr-rs-relay tracing format:
// 2024-01-15T10:30:00.123456Z  INFO nostr_rs_relay::server: message
var logLineRegex = regexp.MustCompile(`^(\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}(?:\.\d+)?Z?)\s+(\w+)\s+(.+)$`)

// parseLogLine parses a log line into a LogEntry.
func parseLogLine(line string) LogEntry {
	matches := logLineRegex.FindStringSubmatch(line)
	if matches != nil && len(matches) == 4 {
		return LogEntry{
			Timestamp: matches[1],
			Level:     matches[2],
			Message:   matches[3],
		}
	}
	// Fallback: treat entire line as message
	return LogEntry{
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Level:     "INFO",
		Message:   line,
	}
}

// Stop stops the relay process.
func (r *Relay) Stop() error {
	pid, err := r.findRelayPID()
	if err != nil {
		return err
	}

	if pid == 0 {
		return nil // Not running
	}

	process, err := os.FindProcess(pid)
	if err != nil {
		return fmt.Errorf("failed to find process %d: %w", pid, err)
	}

	// Send SIGTERM for graceful shutdown
	if err := process.Signal(syscall.SIGTERM); err != nil {
		return fmt.Errorf("failed to send SIGTERM to relay: %w", err)
	}

	// Wait for process to exit (with timeout)
	done := make(chan error, 1)
	go func() {
		_, err := process.Wait()
		done <- err
	}()

	select {
	case <-done:
		return nil
	case <-time.After(10 * time.Second):
		// Force kill if graceful shutdown takes too long
		process.Signal(syscall.SIGKILL)
		return nil
	}
}

// Restart performs an async restart of the relay process.
// It returns immediately and the restart happens in the background.
// Use IsRestarting() to check if restart is in progress.
func (r *Relay) Restart() error {
	if r.IsRestarting() {
		return fmt.Errorf("restart already in progress")
	}

	r.setRestarting(true)

	go func() {
		defer r.setRestarting(false)

		// Stop the relay
		if err := r.Stop(); err != nil {
			// Log error but continue - process may already be stopped
			fmt.Printf("Warning during relay stop: %v\n", err)
		}

		// Wait a moment for cleanup
		time.Sleep(500 * time.Millisecond)

		// Start the relay
		if err := r.Start(); err != nil {
			fmt.Printf("Error starting relay: %v\n", err)
		}
	}()

	return nil
}

// Reload sends SIGHUP to the relay process to reload its configuration.
// Returns nil if no relay process is found (graceful handling for dev environments).
func (r *Relay) Reload() error {
	pid, err := r.findRelayPID()
	if err != nil {
		// No relay running - this is okay in development
		return nil
	}

	if pid == 0 {
		return nil
	}

	// Send SIGHUP to trigger config reload
	process, err := os.FindProcess(pid)
	if err != nil {
		return fmt.Errorf("failed to find process %d: %w", pid, err)
	}

	if err := process.Signal(syscall.SIGHUP); err != nil {
		return fmt.Errorf("failed to send SIGHUP to relay: %w", err)
	}

	return nil
}

// findRelayPID finds the PID of the running nostr-rs-relay process.
// Returns 0 if no process is found.
func (r *Relay) findRelayPID() (int, error) {
	// Try pgrep first (works on Linux and macOS)
	cmd := exec.Command("pgrep", "-f", "nostr-rs-relay")
	var out bytes.Buffer
	cmd.Stdout = &out

	if err := cmd.Run(); err != nil {
		// pgrep returns exit code 1 if no process found - this is normal
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
			return 0, nil
		}
		// Try alternative method using ps
		return r.findRelayPIDWithPS()
	}

	// Parse the PID from output
	pidStr := strings.TrimSpace(out.String())
	if pidStr == "" {
		return 0, nil
	}

	// If multiple PIDs, take the first one
	pids := strings.Split(pidStr, "\n")
	if len(pids) == 0 {
		return 0, nil
	}

	pid, err := strconv.Atoi(strings.TrimSpace(pids[0]))
	if err != nil {
		return 0, fmt.Errorf("failed to parse PID: %w", err)
	}

	return pid, nil
}

// findRelayPIDWithPS is a fallback method using ps command.
func (r *Relay) findRelayPIDWithPS() (int, error) {
	cmd := exec.Command("ps", "aux")
	var out bytes.Buffer
	cmd.Stdout = &out

	if err := cmd.Run(); err != nil {
		return 0, nil // Can't find process, return 0
	}

	lines := strings.Split(out.String(), "\n")
	for _, line := range lines {
		if strings.Contains(line, "nostr-rs-relay") && !strings.Contains(line, "grep") {
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				pid, err := strconv.Atoi(fields[1])
				if err == nil {
					return pid, nil
				}
			}
		}
	}

	return 0, nil
}

// IsRunning checks if the relay process is currently running.
func (r *Relay) IsRunning() bool {
	pid, err := r.findRelayPID()
	if err != nil {
		return false
	}
	return pid > 0
}

// Status returns the current relay status.
func (r *Relay) Status() (bool, error) {
	return r.IsRunning(), nil
}

// GetPID returns the PID of the relay process, or 0 if not running.
func (r *Relay) GetPID() int {
	pid, err := r.findRelayPID()
	if err != nil {
		return 0
	}
	return pid
}

// GetMemoryUsage returns the memory usage of the relay process in bytes.
// Returns 0 if the process is not running or if memory info is unavailable.
func (r *Relay) GetMemoryUsage() int64 {
	pid := r.GetPID()
	if pid == 0 {
		return 0
	}

	// Read from /proc/{pid}/status on Linux
	statusPath := fmt.Sprintf("/proc/%d/status", pid)
	file, err := os.Open(statusPath)
	if err != nil {
		return 0
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "VmRSS:") {
			// Parse "VmRSS:     12345 kB"
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				kb, err := strconv.ParseInt(fields[1], 10, 64)
				if err == nil {
					return kb * 1024 // Convert KB to bytes
				}
			}
		}
	}

	return 0
}

// GetProcessUptime returns the uptime of the relay process in seconds.
// Returns 0 if the process is not running or if uptime info is unavailable.
func (r *Relay) GetProcessUptime() int64 {
	pid := r.GetPID()
	if pid == 0 {
		return 0
	}

	// Read process start time from /proc/{pid}/stat on Linux
	statPath := fmt.Sprintf("/proc/%d/stat", pid)
	data, err := os.ReadFile(statPath)
	if err != nil {
		return 0
	}

	// The stat file format has the start time as field 22 (1-indexed)
	// Format: pid (comm) state ppid pgrp session tty_nr tpgid flags minflt cminflt majflt cmajflt
	//         utime stime cutime cstime priority nice num_threads itrealvalue starttime ...
	// Find the closing paren of comm field first (comm can contain spaces)
	statStr := string(data)
	closeParenIdx := strings.LastIndex(statStr, ")")
	if closeParenIdx == -1 {
		return 0
	}

	// Fields after the comm field
	fieldsAfterComm := strings.Fields(statStr[closeParenIdx+1:])
	if len(fieldsAfterComm) < 20 {
		return 0
	}

	// starttime is field 22, which is index 19 in fieldsAfterComm (0-indexed, after pid and comm)
	startTimeTicks, err := strconv.ParseInt(fieldsAfterComm[19], 10, 64)
	if err != nil {
		return 0
	}

	// Get system boot time and clock ticks per second
	uptimeData, err := os.ReadFile("/proc/uptime")
	if err != nil {
		return 0
	}
	uptimeFields := strings.Fields(string(uptimeData))
	if len(uptimeFields) < 1 {
		return 0
	}
	systemUptime, err := strconv.ParseFloat(uptimeFields[0], 64)
	if err != nil {
		return 0
	}

	// Clock ticks per second (typically 100 on Linux)
	clockTicks := int64(100)

	// Calculate process uptime
	processStartSeconds := float64(startTimeTicks) / float64(clockTicks)
	processUptime := systemUptime - processStartSeconds

	if processUptime < 0 {
		return 0
	}

	return int64(processUptime)
}

// GetRecentLogs returns the most recent log entries.
func (r *Relay) GetRecentLogs(limit int) []LogEntry {
	if r.logBuffer == nil {
		return nil
	}
	return r.logBuffer.GetRecent(limit)
}

// SubscribeLogs returns a channel that receives new log entries.
func (r *Relay) SubscribeLogs() chan LogEntry {
	if r.logBuffer == nil {
		return nil
	}
	return r.logBuffer.Subscribe()
}

// UnsubscribeLogs removes a log subscriber.
func (r *Relay) UnsubscribeLogs(ch chan LogEntry) {
	if r.logBuffer == nil {
		return
	}
	r.logBuffer.Unsubscribe(ch)
}
