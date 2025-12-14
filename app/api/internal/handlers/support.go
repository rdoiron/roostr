package handlers

import (
	"net/http"
	"os"
)

// SupportConfigResponse contains support and donation configuration.
type SupportConfigResponse struct {
	LightningAddress string `json:"lightning_address"`
	BitcoinAddress   string `json:"bitcoin_address"`
	GithubRepo       string `json:"github_repo"`
	DeveloperNpub    string `json:"developer_npub"`
	Version          string `json:"version"`
}

// GetSupportConfig returns the support and donation configuration.
// GET /api/v1/support/config
func (h *Handler) GetSupportConfig(w http.ResponseWriter, r *http.Request) {
	config := SupportConfigResponse{
		LightningAddress: getEnvOrDefault("DONATION_LIGHTNING_ADDRESS", "donate@example.com"),
		BitcoinAddress:   getEnvOrDefault("DONATION_BITCOIN_ADDRESS", "bc1qexample..."),
		GithubRepo:       getEnvOrDefault("GITHUB_REPO", "https://github.com/roostr/roostr"),
		DeveloperNpub:    getEnvOrDefault("DEVELOPER_NPUB", "npub1..."),
		Version:          "0.1.0",
	}

	respondJSON(w, http.StatusOK, config)
}

// getEnvOrDefault returns the environment variable value or a default.
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
