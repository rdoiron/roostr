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
