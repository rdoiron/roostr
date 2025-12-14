// Package relay provides control over the nostr-rs-relay process and configuration.
package relay

import (
	"bytes"
	"os"
	"sync"

	"github.com/BurntSushi/toml"
)

// Config represents the nostr-rs-relay configuration file structure.
// We only define the sections we need to modify; TOML will preserve others.
type Config struct {
	Info          InfoConfig          `toml:"info"`
	Database      DatabaseConfig      `toml:"database"`
	Network       NetworkConfig       `toml:"network"`
	Limits        LimitsConfig        `toml:"limits"`
	Authorization AuthorizationConfig `toml:"authorization"`
	Logging       LoggingConfig       `toml:"logging"`
}

// InfoConfig contains relay metadata.
type InfoConfig struct {
	RelayURL    string `toml:"relay_url"`
	Name        string `toml:"name"`
	Description string `toml:"description"`
	Pubkey      string `toml:"pubkey"`
	Contact     string `toml:"contact"`
	RelayIcon   string `toml:"relay_icon,omitempty"`
}

// DatabaseConfig contains database settings.
type DatabaseConfig struct {
	DataDirectory string `toml:"data_directory"`
}

// NetworkConfig contains network settings.
type NetworkConfig struct {
	Port    int    `toml:"port"`
	Address string `toml:"address"`
}

// LimitsConfig contains rate limiting settings.
type LimitsConfig struct {
	MessagesPerSec      int `toml:"messages_per_sec"`
	SubscriptionsPerMin int `toml:"subscriptions_per_min"`
	MaxEventBytes       int `toml:"max_event_bytes"`
	MaxWSMessageBytes   int `toml:"max_ws_message_bytes"`
	MaxSubsPerConn      int `toml:"max_subs_per_conn,omitempty"`
	MinPowDifficulty    int `toml:"min_pow_difficulty,omitempty"`
}

// AuthorizationConfig contains access control settings.
type AuthorizationConfig struct {
	NIP42Auth          bool     `toml:"nip42_auth"`
	PubkeyWhitelist    []string `toml:"pubkey_whitelist"`
	PubkeyBlacklist    []string `toml:"pubkey_blacklist,omitempty"`
	EventKindAllowlist []int    `toml:"event_kind_allowlist,omitempty"`
}

// LoggingConfig contains logging settings.
type LoggingConfig struct {
	FolderPath string `toml:"folder_path"`
	FilePrefix string `toml:"file_prefix"`
}

// ConfigManager handles reading and writing relay configuration.
type ConfigManager struct {
	path string
	mu   sync.RWMutex
}

// NewConfigManager creates a new ConfigManager for the given config file path.
func NewConfigManager(path string) *ConfigManager {
	return &ConfigManager{
		path: path,
	}
}

// Path returns the config file path.
func (cm *ConfigManager) Path() string {
	return cm.path
}

// Read reads the configuration from the TOML file.
func (cm *ConfigManager) Read() (*Config, error) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	var cfg Config
	if _, err := toml.DecodeFile(cm.path, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

// Write writes the configuration to the TOML file.
func (cm *ConfigManager) Write(cfg *Config) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	var buf bytes.Buffer
	buf.WriteString("# Roostr - nostr-rs-relay configuration\n")
	buf.WriteString("# This file is managed by Roostr. Manual edits may be overwritten.\n\n")

	encoder := toml.NewEncoder(&buf)
	if err := encoder.Encode(cfg); err != nil {
		return err
	}

	return os.WriteFile(cm.path, buf.Bytes(), 0644)
}

// UpdateWhitelist updates the pubkey whitelist in the config file.
// It reads the current config, updates the whitelist, and writes back.
func (cm *ConfigManager) UpdateWhitelist(pubkeys []string) error {
	cfg, err := cm.Read()
	if err != nil {
		return err
	}

	cfg.Authorization.PubkeyWhitelist = pubkeys
	return cm.Write(cfg)
}

// UpdateBlacklist updates the pubkey blacklist in the config file.
// It reads the current config, updates the blacklist, and writes back.
func (cm *ConfigManager) UpdateBlacklist(pubkeys []string) error {
	cfg, err := cm.Read()
	if err != nil {
		return err
	}

	cfg.Authorization.PubkeyBlacklist = pubkeys
	return cm.Write(cfg)
}

// GetWhitelist returns the current whitelist from the config file.
func (cm *ConfigManager) GetWhitelist() ([]string, error) {
	cfg, err := cm.Read()
	if err != nil {
		return nil, err
	}
	return cfg.Authorization.PubkeyWhitelist, nil
}

// GetBlacklist returns the current blacklist from the config file.
func (cm *ConfigManager) GetBlacklist() ([]string, error) {
	cfg, err := cm.Read()
	if err != nil {
		return nil, err
	}
	return cfg.Authorization.PubkeyBlacklist, nil
}
