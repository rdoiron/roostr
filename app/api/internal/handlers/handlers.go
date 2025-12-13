// Package handlers contains HTTP handlers for the Roostr API.
// All handlers follow RESTful conventions and return JSON responses.
package handlers

import (
	"net/http"

	"github.com/roostr/roostr/app/api/internal/config"
	"github.com/roostr/roostr/app/api/internal/db"
)

// Handler holds dependencies for HTTP handlers.
type Handler struct {
	db  *db.DB
	cfg *config.Config
}

// New creates a new Handler instance with dependencies.
func New(database *db.DB, cfg *config.Config) *Handler {
	return &Handler{
		db:  database,
		cfg: cfg,
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

	// Access control endpoints
	mux.HandleFunc("GET /api/v1/access/mode", h.GetAccessMode)
	mux.HandleFunc("PUT /api/v1/access/mode", h.SetAccessMode)
	mux.HandleFunc("GET /api/v1/access/whitelist", h.GetWhitelist)
	mux.HandleFunc("POST /api/v1/access/whitelist", h.AddToWhitelist)
	mux.HandleFunc("DELETE /api/v1/access/whitelist/{pubkey}", h.RemoveFromWhitelist)
	mux.HandleFunc("PATCH /api/v1/access/whitelist/{pubkey}", h.UpdateWhitelistEntry)

	// Event browser endpoints
	mux.HandleFunc("GET /api/v1/events", h.GetEvents)
	mux.HandleFunc("GET /api/v1/events/{id}", h.GetEvent)
	mux.HandleFunc("DELETE /api/v1/events/{id}", h.DeleteEvent)

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
