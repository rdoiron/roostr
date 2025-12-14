package services

import (
	"encoding/hex"
	"os"
	"path/filepath"
	"strings"
)

// DetectLND attempts to automatically detect LND credentials.
// It checks environment variables first, then known paths for Umbrel and Start9.
func DetectLND() (*LNDConfig, error) {
	// 1. Check environment variables first (works on both platforms)
	if cfg := detectFromEnv(); cfg != nil {
		return cfg, nil
	}

	// 2. Try Umbrel paths
	if cfg := detectUmbrel(); cfg != nil {
		return cfg, nil
	}

	// 3. Try Start9 paths
	if cfg := detectStart9(); cfg != nil {
		return cfg, nil
	}

	return nil, ErrLNDNotDetected
}

// detectFromEnv checks for LND configuration in environment variables.
func detectFromEnv() *LNDConfig {
	host := os.Getenv("LND_REST_HOST")
	if host == "" {
		return nil
	}

	// Try pre-encoded macaroon first
	macaroonHex := os.Getenv("LND_MACAROON_HEX")
	if macaroonHex == "" {
		// Try reading from path
		macaroonPath := os.Getenv("LND_MACAROON_PATH")
		if macaroonPath != "" {
			if mac, err := readMacaroon(macaroonPath); err == nil {
				macaroonHex = mac
			}
		}
	}

	if macaroonHex == "" {
		return nil
	}

	return &LNDConfig{
		Host:        host,
		MacaroonHex: macaroonHex,
		TLSCertPath: os.Getenv("LND_CERT_PATH"),
	}
}

// detectUmbrel checks for LND on Umbrel.
func detectUmbrel() *LNDConfig {
	// Known Umbrel macaroon paths
	macaroonPaths := []string{
		"/umbrel/lnd/data/chain/bitcoin/mainnet/admin.macaroon",
		"/home/umbrel/umbrel/lnd/data/chain/bitcoin/mainnet/admin.macaroon",
		expandPath("~/umbrel/app-data/lightning/data/chain/bitcoin/mainnet/admin.macaroon"),
		// Umbrel 0.5+ paths
		"/home/umbrel/umbrel/app-data/lightning/data/chain/bitcoin/mainnet/admin.macaroon",
	}

	var macaroonHex string
	for _, path := range macaroonPaths {
		if path == "" {
			continue
		}
		if mac, err := readMacaroon(path); err == nil {
			macaroonHex = mac
			break
		}
	}

	if macaroonHex == "" {
		return nil
	}

	// Umbrel hosts - try common addresses
	hosts := []string{
		"umbrel.local:8080",
		"10.21.21.9:8080", // Umbrel's internal IP
		"localhost:8080",
	}

	// For now, return with the first host - connection test will validate
	return &LNDConfig{
		Host:        hosts[0],
		MacaroonHex: macaroonHex,
	}
}

// detectStart9 checks for LND on Start9/StartOS.
func detectStart9() *LNDConfig {
	// Known Start9 macaroon paths
	macaroonPaths := []string{
		"/mnt/embassy/lnd/admin.macaroon",
		"/embassy-data/package-data/lnd/data/chain/bitcoin/mainnet/admin.macaroon",
		// StartOS 0.3+ paths
		"/embassy-data/package-data/lnd/volumes/main/data/chain/bitcoin/mainnet/admin.macaroon",
	}

	var macaroonHex string
	for _, path := range macaroonPaths {
		if mac, err := readMacaroon(path); err == nil {
			macaroonHex = mac
			break
		}
	}

	if macaroonHex == "" {
		return nil
	}

	// Start9 hosts - check environment first, then try common addresses
	host := os.Getenv("LND_HOST")
	if host == "" {
		// Start9 service discovery
		hosts := []string{
			"lnd.embassy:8080",
			"localhost:8080",
		}
		host = hosts[0]
	}

	// Ensure port is included
	if !strings.Contains(host, ":") {
		host = host + ":8080"
	}

	return &LNDConfig{
		Host:        host,
		MacaroonHex: macaroonHex,
	}
}

// readMacaroon reads a macaroon file and returns it as hex-encoded string.
func readMacaroon(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(data), nil
}

// expandPath expands ~ to the user's home directory.
func expandPath(path string) string {
	if !strings.HasPrefix(path, "~/") {
		return path
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, path[2:])
}

// DetectionResult contains the result of LND detection.
type DetectionResult struct {
	Detected bool       `json:"detected"`
	Config   *LNDConfig `json:"config,omitempty"`
	Source   string     `json:"source,omitempty"` // "env", "umbrel", "start9"
	Error    string     `json:"error,omitempty"`
}

// DetectLNDWithSource attempts to detect LND and returns the source of detection.
func DetectLNDWithSource() *DetectionResult {
	// Check environment variables
	if cfg := detectFromEnv(); cfg != nil {
		return &DetectionResult{
			Detected: true,
			Config:   cfg,
			Source:   "env",
		}
	}

	// Check Umbrel
	if cfg := detectUmbrel(); cfg != nil {
		return &DetectionResult{
			Detected: true,
			Config:   cfg,
			Source:   "umbrel",
		}
	}

	// Check Start9
	if cfg := detectStart9(); cfg != nil {
		return &DetectionResult{
			Detected: true,
			Config:   cfg,
			Source:   "start9",
		}
	}

	return &DetectionResult{
		Detected: false,
		Error:    "LND not detected. Please configure manually.",
	}
}
