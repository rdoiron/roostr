// Package config handles configuration management for Roostr.
// This includes reading/writing the relay config.toml and app settings.
package config

import (
	"os"
)

// Config holds the application configuration.
type Config struct {
	// Server settings
	Port string

	// Database paths
	RelayDBPath string
	AppDBPath   string

	// Relay settings
	ConfigPath  string
	RelayBinary string

	// Relay URLs (provided by platform)
	RelayURL   string // Local WebSocket URL (e.g., ws://umbrel.local:4848)
	TorAddress string // Tor .onion address (e.g., abc123...onion:4848)

	// UI settings
	StaticDir string // Directory containing built UI static files

	// Feature flags
	Debug bool
}

// Load reads configuration from environment variables with sensible defaults.
func Load() (*Config, error) {
	cfg := &Config{
		Port:        getEnv("PORT", "3001"),
		RelayDBPath: getEnv("RELAY_DB_PATH", "data/nostr.db"),
		AppDBPath:   getEnv("APP_DB_PATH", "data/roostr.db"),
		ConfigPath:  getEnv("CONFIG_PATH", "data/config.toml"),
		RelayBinary: getEnv("RELAY_BINARY", "/usr/bin/nostr-rs-relay"),
		RelayURL:    getEnv("RELAY_URL", ""),   // e.g., ws://umbrel.local:4848
		TorAddress:  getEnv("TOR_ADDRESS", ""), // e.g., abc123...onion:4848
		StaticDir:   getEnv("STATIC_DIR", ""),  // Directory with built UI files
		Debug:       getEnv("DEBUG", "") == "true",
	}

	return cfg, nil
}

// getEnv returns the value of an environment variable or a default value.
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
