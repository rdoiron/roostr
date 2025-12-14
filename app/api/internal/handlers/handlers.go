// Package handlers contains HTTP handlers for the Roostr API.
// All handlers follow RESTful conventions and return JSON responses.
package handlers

import (
	"net/http"
	"time"

	"github.com/roostr/roostr/app/api/internal/config"
	"github.com/roostr/roostr/app/api/internal/db"
	"github.com/roostr/roostr/app/api/internal/relay"
)

// Handler holds dependencies for HTTP handlers.
type Handler struct {
	db        *db.DB
	cfg       *config.Config
	configMgr *relay.ConfigManager
	relay     *relay.Relay
	startTime time.Time // Server start time for uptime calculation
}

// New creates a new Handler instance with dependencies.
func New(database *db.DB, cfg *config.Config, configMgr *relay.ConfigManager, relayMgr *relay.Relay) *Handler {
	return &Handler{
		db:        database,
		cfg:       cfg,
		configMgr: configMgr,
		relay:     relayMgr,
		startTime: time.Now(),
	}
}

// RegisterRoutes registers all HTTP routes on the given mux.
func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	// Health check (both root and API paths for flexibility)
	mux.HandleFunc("GET /health", h.Health)
	mux.HandleFunc("GET /api/v1/health", h.Health)

	// Setup endpoints
	mux.HandleFunc("GET /api/v1/setup/status", h.GetSetupStatus)
	mux.HandleFunc("GET /api/v1/setup/validate-identity", h.ValidateIdentity)
	mux.HandleFunc("POST /api/v1/setup/complete", h.CompleteSetup)

	// Dashboard/Stats endpoints
	mux.HandleFunc("GET /api/v1/stats/summary", h.GetStatsSummary)
	mux.HandleFunc("GET /api/v1/relay/status", h.GetRelayStatus)
	mux.HandleFunc("GET /api/v1/relay/urls", h.GetRelayURLs)
	mux.HandleFunc("GET /api/v1/events/recent", h.GetRecentEvents)

	// Access control endpoints
	mux.HandleFunc("GET /api/v1/access/mode", h.GetAccessMode)
	mux.HandleFunc("PUT /api/v1/access/mode", h.SetAccessMode)

	// Whitelist endpoints
	mux.HandleFunc("GET /api/v1/access/whitelist", h.GetWhitelist)
	mux.HandleFunc("POST /api/v1/access/whitelist", h.AddToWhitelist)
	mux.HandleFunc("DELETE /api/v1/access/whitelist/{pubkey}", h.RemoveFromWhitelist)
	mux.HandleFunc("PATCH /api/v1/access/whitelist/{pubkey}", h.UpdateWhitelistEntry)

	// Blacklist endpoints
	mux.HandleFunc("GET /api/v1/access/blacklist", h.GetBlacklist)
	mux.HandleFunc("POST /api/v1/access/blacklist", h.AddToBlacklist)
	mux.HandleFunc("DELETE /api/v1/access/blacklist/{pubkey}", h.RemoveFromBlacklist)

	// NIP-05 resolution endpoint
	mux.HandleFunc("GET /api/v1/nip05/{identifier}", h.ResolveNIP05)

	// Event browser endpoints
	mux.HandleFunc("GET /api/v1/events", h.GetEvents)
	mux.HandleFunc("GET /api/v1/events/export", h.ExportEvents)
	mux.HandleFunc("GET /api/v1/events/{id}", h.GetEvent)
	mux.HandleFunc("DELETE /api/v1/events/{id}", h.DeleteEvent)

	// Configuration endpoints
	mux.HandleFunc("GET /api/v1/config", h.GetConfig)
	mux.HandleFunc("PATCH /api/v1/config", h.UpdateConfig)
	mux.HandleFunc("POST /api/v1/config/reload", h.ReloadConfig)

	// Storage management endpoints
	mux.HandleFunc("GET /api/v1/storage/status", h.GetStorageStatus)
	mux.HandleFunc("GET /api/v1/storage/retention", h.GetRetentionPolicy)
	mux.HandleFunc("PUT /api/v1/storage/retention", h.UpdateRetentionPolicy)
	mux.HandleFunc("POST /api/v1/storage/cleanup", h.ManualCleanup)
	mux.HandleFunc("POST /api/v1/storage/vacuum", h.RunVacuum)
	mux.HandleFunc("GET /api/v1/storage/deletion-requests", h.GetDeletionRequests)
	mux.HandleFunc("GET /api/v1/storage/estimate", h.GetStorageEstimate)
	mux.HandleFunc("POST /api/v1/storage/integrity-check", h.RunIntegrityCheck)

	// Serve static files for the UI (SPA fallback)
	mux.HandleFunc("/", h.ServeUI)
}

// ServeUI serves the Svelte SPA, falling back to index.html for client-side routing.
func (h *Handler) ServeUI(w http.ResponseWriter, r *http.Request) {
	// For API routes that weren't matched, return 404
	if len(r.URL.Path) > 4 && r.URL.Path[:4] == "/api" {
		http.NotFound(w, r)
		return
	}

	// Serve static files from the UI build directory
	// In production, this would serve from a static file server
	// For now, just return a placeholder
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(`<!DOCTYPE html>
<html>
<head><title>Roostr</title></head>
<body>
<h1>Roostr</h1>
<p>UI not built yet. Run <code>make build-ui</code> to build the frontend.</p>
<p>API is available at <a href="/api/v1/health">/api/v1/health</a></p>
</body>
</html>`))
}
