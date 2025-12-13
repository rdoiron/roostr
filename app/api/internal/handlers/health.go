package handlers

import (
	"net/http"
)

// Health returns the health status of the API.
func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	status := "ok"
	relayConnected := h.db.IsRelayDBConnected()

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"status":          status,
		"relay_connected": relayConnected,
	})
}
