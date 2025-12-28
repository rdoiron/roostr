// Package handlers contains HTTP handlers for the Roostr API.
// All handlers follow RESTful conventions and return JSON responses.
package handlers

import (
	"net/http"
	"os"
	"time"

	"github.com/roostr/roostr/app/api/internal/config"
	"github.com/roostr/roostr/app/api/internal/db"
	"github.com/roostr/roostr/app/api/internal/relay"
	"github.com/roostr/roostr/app/api/internal/services"
)

// Handler holds dependencies for HTTP handlers.
type Handler struct {
	db        *db.DB
	cfg       *config.Config
	configMgr *relay.ConfigManager
	relay     *relay.Relay
	services  *services.Services
	startTime time.Time // Server start time for uptime calculation
}

// New creates a new Handler instance with dependencies.
func New(database *db.DB, cfg *config.Config, configMgr *relay.ConfigManager, relayMgr *relay.Relay, svc *services.Services) *Handler {
	return &Handler{
		db:        database,
		cfg:       cfg,
		configMgr: configMgr,
		relay:     relayMgr,
		services:  svc,
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
	mux.HandleFunc("GET /api/v1/stats/stream", h.StreamDashboardStats)
	mux.HandleFunc("GET /api/v1/stats/events-over-time", h.GetEventsOverTime)
	mux.HandleFunc("GET /api/v1/stats/events-by-kind", h.GetEventsByKind)
	mux.HandleFunc("GET /api/v1/stats/top-authors", h.GetTopAuthors)
	mux.HandleFunc("GET /api/v1/relay/status", h.GetRelayStatus)
	mux.HandleFunc("GET /api/v1/relay/urls", h.GetRelayURLs)
	mux.HandleFunc("GET /api/v1/events/recent", h.GetRecentEvents)

	// Relay control endpoints
	mux.HandleFunc("POST /api/v1/relay/reload", h.ReloadRelay)
	mux.HandleFunc("POST /api/v1/relay/restart", h.RestartRelay)
	mux.HandleFunc("GET /api/v1/relay/logs", h.GetRelayLogs)
	mux.HandleFunc("GET /api/v1/relay/logs/stream", h.StreamRelayLogs)

	// Access control endpoints
	mux.HandleFunc("GET /api/v1/access/mode", h.GetAccessMode)
	mux.HandleFunc("PUT /api/v1/access/mode", h.SetAccessMode)

	// Whitelist endpoints
	mux.HandleFunc("GET /api/v1/access/whitelist", h.GetWhitelist)
	mux.HandleFunc("POST /api/v1/access/whitelist", h.AddToWhitelist)
	mux.HandleFunc("POST /api/v1/access/whitelist/bulk", h.BulkAddToWhitelist)
	mux.HandleFunc("DELETE /api/v1/access/whitelist/{pubkey}", h.RemoveFromWhitelist)
	mux.HandleFunc("PATCH /api/v1/access/whitelist/{pubkey}", h.UpdateWhitelistEntry)

	// Blacklist endpoints
	mux.HandleFunc("GET /api/v1/access/blacklist", h.GetBlacklist)
	mux.HandleFunc("POST /api/v1/access/blacklist", h.AddToBlacklist)
	mux.HandleFunc("DELETE /api/v1/access/blacklist/{pubkey}", h.RemoveFromBlacklist)

	// Paid access endpoints
	mux.HandleFunc("GET /api/v1/access/pricing", h.GetPricingTiers)
	mux.HandleFunc("PUT /api/v1/access/pricing", h.UpdatePricingTiers)
	mux.HandleFunc("GET /api/v1/access/paid-users", h.GetPaidUsers)
	mux.HandleFunc("DELETE /api/v1/access/paid-users/{pubkey}", h.RevokePaidUserAccess)
	mux.HandleFunc("GET /api/v1/access/revenue", h.GetRevenueStats)

	// NIP-05 resolution endpoint
	mux.HandleFunc("GET /api/v1/nip05/{identifier}", h.ResolveNIP05)

	// Event browser endpoints
	mux.HandleFunc("GET /api/v1/events", h.GetEvents)
	mux.HandleFunc("GET /api/v1/events/export", h.ExportEvents)
	mux.HandleFunc("GET /api/v1/events/export/estimate", h.GetExportEstimate)
	mux.HandleFunc("POST /api/v1/events/import", h.ImportEvents)
	mux.HandleFunc("GET /api/v1/events/{id}", h.GetEvent)
	mux.HandleFunc("DELETE /api/v1/events/{id}", h.DeleteEvent)

	// Configuration endpoints
	mux.HandleFunc("GET /api/v1/config", h.GetConfig)
	mux.HandleFunc("PATCH /api/v1/config", h.UpdateConfig)
	mux.HandleFunc("POST /api/v1/config/reload", h.ReloadConfig)

	// Settings endpoints
	mux.HandleFunc("GET /api/v1/settings/timezone", h.GetTimezone)
	mux.HandleFunc("PUT /api/v1/settings/timezone", h.SetTimezone)

	// Storage management endpoints
	mux.HandleFunc("GET /api/v1/storage/status", h.GetStorageStatus)
	mux.HandleFunc("GET /api/v1/storage/retention", h.GetRetentionPolicy)
	mux.HandleFunc("PUT /api/v1/storage/retention", h.UpdateRetentionPolicy)
	mux.HandleFunc("POST /api/v1/storage/cleanup", h.ManualCleanup)
	mux.HandleFunc("POST /api/v1/storage/vacuum", h.RunVacuum)
	mux.HandleFunc("GET /api/v1/storage/deletion-requests", h.GetDeletionRequests)
	mux.HandleFunc("GET /api/v1/storage/estimate", h.GetStorageEstimate)
	mux.HandleFunc("POST /api/v1/storage/integrity-check", h.RunIntegrityCheck)

	// Sync endpoints
	mux.HandleFunc("POST /api/v1/sync/start", h.StartSync)
	mux.HandleFunc("GET /api/v1/sync/status", h.GetSyncStatus)
	mux.HandleFunc("POST /api/v1/sync/cancel", h.CancelSync)
	mux.HandleFunc("GET /api/v1/sync/history", h.GetSyncHistory)
	// Sync pubkeys configuration
	mux.HandleFunc("GET /api/v1/sync/pubkeys", h.GetSyncPubkeys)
	mux.HandleFunc("POST /api/v1/sync/pubkeys", h.AddSyncPubkey)
	mux.HandleFunc("DELETE /api/v1/sync/pubkeys/{pubkey}", h.RemoveSyncPubkey)
	// Sync relays configuration
	mux.HandleFunc("GET /api/v1/sync/relays", h.GetSyncRelays)
	mux.HandleFunc("POST /api/v1/sync/relays", h.AddSyncRelay)
	mux.HandleFunc("DELETE /api/v1/sync/relays/{url}", h.RemoveSyncRelay)
	mux.HandleFunc("POST /api/v1/sync/relays/reset", h.ResetSyncRelays)

	// Support endpoints
	mux.HandleFunc("GET /api/v1/support/config", h.GetSupportConfig)

	// Lightning endpoints
	mux.HandleFunc("GET /api/v1/lightning/status", h.GetLightningStatus)
	mux.HandleFunc("PUT /api/v1/lightning/config", h.SaveLightningConfig)
	mux.HandleFunc("POST /api/v1/lightning/test", h.TestLightningConnection)

	// Public signup endpoints (no auth required)
	mux.HandleFunc("GET /public/relay-info", h.GetRelayInfo)
	mux.HandleFunc("POST /public/create-invoice", h.CreateSignupInvoice)
	mux.HandleFunc("GET /public/invoice-status/{hash}", h.GetInvoiceStatus)

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

	// For public API routes that weren't matched, return 404
	if len(r.URL.Path) > 7 && r.URL.Path[:7] == "/public" {
		http.NotFound(w, r)
		return
	}

	// If static directory is not configured, show placeholder
	if h.cfg.StaticDir == "" {
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
		return
	}

	// Serve static files from the configured directory
	h.serveStaticFile(w, r)
}

// serveStaticFile serves a static file or falls back to index.html for SPA routing.
func (h *Handler) serveStaticFile(w http.ResponseWriter, r *http.Request) {
	// Clean the path to prevent directory traversal
	path := r.URL.Path
	if path == "/" {
		path = "/index.html"
	}

	// Try to serve the requested file
	filePath := h.cfg.StaticDir + path

	// Check if the file exists
	info, err := os.Stat(filePath)
	if err != nil || info.IsDir() {
		// File doesn't exist or is a directory - serve index.html for SPA routing
		filePath = h.cfg.StaticDir + "/index.html"
	}

	// Serve the file
	http.ServeFile(w, r, filePath)
}
