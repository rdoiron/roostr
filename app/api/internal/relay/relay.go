// Package relay provides control over the nostr-rs-relay process.
// This includes starting, stopping, reloading config, and monitoring status.
package relay

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
)

// Relay manages the nostr-rs-relay process.
type Relay struct {
	BinaryPath string
	ConfigPath string
	cmd        *exec.Cmd
}

// New creates a new Relay instance.
func New(binaryPath, configPath string) *Relay {
	return &Relay{
		BinaryPath: binaryPath,
		ConfigPath: configPath,
	}
}

// Start starts the relay process.
func (r *Relay) Start() error {
	// TODO: Implement relay start
	return nil
}

// Stop stops the relay process.
func (r *Relay) Stop() error {
	// TODO: Implement relay stop
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
