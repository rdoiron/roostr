package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
)

// UpdateConfigRequest represents a partial config update request.
type UpdateConfigRequest struct {
	Info          *InfoUpdate          `json:"info,omitempty"`
	Limits        *LimitsUpdate        `json:"limits,omitempty"`
	Authorization *AuthorizationUpdate `json:"authorization,omitempty"`
}

// InfoUpdate contains optional info field updates.
type InfoUpdate struct {
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
	Contact     *string `json:"contact,omitempty"`
	RelayIcon   *string `json:"relay_icon,omitempty"`
}

// LimitsUpdate contains optional limits field updates.
type LimitsUpdate struct {
	MaxEventBytes     *int `json:"max_event_bytes,omitempty"`
	MaxWSMessageBytes *int `json:"max_ws_message_bytes,omitempty"`
	MessagesPerSec    *int `json:"messages_per_sec,omitempty"`
	MaxSubsPerConn    *int `json:"max_subs_per_conn,omitempty"`
	MinPowDifficulty  *int `json:"min_pow_difficulty,omitempty"`
}

// AuthorizationUpdate contains optional authorization field updates.
type AuthorizationUpdate struct {
	NIP42Auth          *bool  `json:"nip42_auth,omitempty"`
	EventKindAllowlist *[]int `json:"event_kind_allowlist,omitempty"`
}

// GetConfig returns the current relay configuration.
// GET /api/v1/config
func (h *Handler) GetConfig(w http.ResponseWriter, r *http.Request) {
	if h.configMgr == nil {
		respondError(w, http.StatusServiceUnavailable, "Config manager not available", "CONFIG_NOT_AVAILABLE")
		return
	}

	cfg, err := h.configMgr.Read()
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to read config", "CONFIG_READ_FAILED")
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"info": map[string]interface{}{
			"name":        cfg.Info.Name,
			"description": cfg.Info.Description,
			"contact":     cfg.Info.Contact,
			"relay_icon":  cfg.Info.RelayIcon,
		},
		"limits": map[string]interface{}{
			"max_event_bytes":      cfg.Limits.MaxEventBytes,
			"max_ws_message_bytes": cfg.Limits.MaxWSMessageBytes,
			"messages_per_sec":     cfg.Limits.MessagesPerSec,
			"max_subs_per_conn":    cfg.Limits.MaxSubsPerConn,
			"min_pow_difficulty":   cfg.Limits.MinPowDifficulty,
		},
		"authorization": map[string]interface{}{
			"nip42_auth":           cfg.Authorization.NIP42Auth,
			"event_kind_allowlist": cfg.Authorization.EventKindAllowlist,
		},
	})
}

// UpdateConfig updates the relay configuration with the provided fields.
// PATCH /api/v1/config
func (h *Handler) UpdateConfig(w http.ResponseWriter, r *http.Request) {
	if h.configMgr == nil {
		respondError(w, http.StatusServiceUnavailable, "Config manager not available", "CONFIG_NOT_AVAILABLE")
		return
	}

	var req UpdateConfigRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body", "INVALID_JSON")
		return
	}

	// Validate the update request
	if err := validateConfigUpdate(&req); err != nil {
		respondError(w, http.StatusBadRequest, err.Error(), "VALIDATION_ERROR")
		return
	}

	// Read current config
	cfg, err := h.configMgr.Read()
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to read config", "CONFIG_READ_FAILED")
		return
	}

	// Merge updates into config
	if req.Info != nil {
		if req.Info.Name != nil {
			cfg.Info.Name = *req.Info.Name
		}
		if req.Info.Description != nil {
			cfg.Info.Description = *req.Info.Description
		}
		if req.Info.Contact != nil {
			cfg.Info.Contact = *req.Info.Contact
		}
		if req.Info.RelayIcon != nil {
			cfg.Info.RelayIcon = *req.Info.RelayIcon
		}
	}

	if req.Limits != nil {
		if req.Limits.MaxEventBytes != nil {
			cfg.Limits.MaxEventBytes = *req.Limits.MaxEventBytes
		}
		if req.Limits.MaxWSMessageBytes != nil {
			cfg.Limits.MaxWSMessageBytes = *req.Limits.MaxWSMessageBytes
		}
		if req.Limits.MessagesPerSec != nil {
			cfg.Limits.MessagesPerSec = *req.Limits.MessagesPerSec
		}
		if req.Limits.MaxSubsPerConn != nil {
			cfg.Limits.MaxSubsPerConn = *req.Limits.MaxSubsPerConn
		}
		if req.Limits.MinPowDifficulty != nil {
			cfg.Limits.MinPowDifficulty = *req.Limits.MinPowDifficulty
		}
	}

	if req.Authorization != nil {
		if req.Authorization.NIP42Auth != nil {
			cfg.Authorization.NIP42Auth = *req.Authorization.NIP42Auth
		}
		if req.Authorization.EventKindAllowlist != nil {
			cfg.Authorization.EventKindAllowlist = *req.Authorization.EventKindAllowlist
		}
	}

	// Write updated config
	if err := h.configMgr.Write(cfg); err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to write config", "CONFIG_WRITE_FAILED")
		return
	}

	// Reload relay to apply changes
	if h.relay != nil {
		if err := h.relay.Reload(); err != nil {
			log.Printf("Warning: failed to reload relay: %v", err)
		}
	}

	// Add audit log
	ctx := r.Context()
	h.db.AddAuditLog(ctx, "config_updated", map[string]string{
		"updated_sections": getUpdatedSections(&req),
	}, "")

	// Return success with updated config
	respondJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "Configuration updated",
		"config": map[string]interface{}{
			"info": map[string]interface{}{
				"name":        cfg.Info.Name,
				"description": cfg.Info.Description,
				"contact":     cfg.Info.Contact,
				"relay_icon":  cfg.Info.RelayIcon,
			},
			"limits": map[string]interface{}{
				"max_event_bytes":      cfg.Limits.MaxEventBytes,
				"max_ws_message_bytes": cfg.Limits.MaxWSMessageBytes,
				"messages_per_sec":     cfg.Limits.MessagesPerSec,
				"max_subs_per_conn":    cfg.Limits.MaxSubsPerConn,
				"min_pow_difficulty":   cfg.Limits.MinPowDifficulty,
			},
			"authorization": map[string]interface{}{
				"nip42_auth":           cfg.Authorization.NIP42Auth,
				"event_kind_allowlist": cfg.Authorization.EventKindAllowlist,
			},
		},
	})
}

// ReloadConfig signals the relay to reload its configuration.
// POST /api/v1/config/reload
func (h *Handler) ReloadConfig(w http.ResponseWriter, r *http.Request) {
	if h.relay == nil {
		respondError(w, http.StatusServiceUnavailable, "Relay manager not available", "RELAY_NOT_AVAILABLE")
		return
	}

	if err := h.relay.Reload(); err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to reload relay", "RELOAD_FAILED")
		return
	}

	ctx := r.Context()
	h.db.AddAuditLog(ctx, "config_reloaded", nil, "")

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "Relay configuration reloaded",
	})
}

// validateConfigUpdate validates the config update request.
func validateConfigUpdate(req *UpdateConfigRequest) error {
	if req.Info != nil {
		// Name: max 64 chars
		if req.Info.Name != nil && len(*req.Info.Name) > 64 {
			return fmt.Errorf("name must be 64 characters or less")
		}
		// Description: max 500 chars
		if req.Info.Description != nil && len(*req.Info.Description) > 500 {
			return fmt.Errorf("description must be 500 characters or less")
		}
		// RelayIcon: valid URL if provided
		if req.Info.RelayIcon != nil && *req.Info.RelayIcon != "" {
			if _, err := url.Parse(*req.Info.RelayIcon); err != nil {
				return fmt.Errorf("relay_icon must be a valid URL")
			}
		}
	}

	if req.Limits != nil {
		// MaxEventBytes: 1KB - 16MB
		if req.Limits.MaxEventBytes != nil {
			if *req.Limits.MaxEventBytes < 1024 || *req.Limits.MaxEventBytes > 16*1024*1024 {
				return fmt.Errorf("max_event_bytes must be between 1024 and 16777216")
			}
		}
		// MaxWSMessageBytes: 1KB - 16MB
		if req.Limits.MaxWSMessageBytes != nil {
			if *req.Limits.MaxWSMessageBytes < 1024 || *req.Limits.MaxWSMessageBytes > 16*1024*1024 {
				return fmt.Errorf("max_ws_message_bytes must be between 1024 and 16777216")
			}
		}
		// MessagesPerSec: 1-100
		if req.Limits.MessagesPerSec != nil {
			if *req.Limits.MessagesPerSec < 1 || *req.Limits.MessagesPerSec > 100 {
				return fmt.Errorf("messages_per_sec must be between 1 and 100")
			}
		}
		// MaxSubsPerConn: 1-100
		if req.Limits.MaxSubsPerConn != nil {
			if *req.Limits.MaxSubsPerConn < 1 || *req.Limits.MaxSubsPerConn > 100 {
				return fmt.Errorf("max_subs_per_conn must be between 1 and 100")
			}
		}
		// MinPowDifficulty: 0-32
		if req.Limits.MinPowDifficulty != nil {
			if *req.Limits.MinPowDifficulty < 0 || *req.Limits.MinPowDifficulty > 32 {
				return fmt.Errorf("min_pow_difficulty must be between 0 and 32")
			}
		}
	}

	if req.Authorization != nil {
		// EventKindAllowlist: all values must be non-negative
		if req.Authorization.EventKindAllowlist != nil {
			for _, kind := range *req.Authorization.EventKindAllowlist {
				if kind < 0 {
					return fmt.Errorf("event kinds must be non-negative integers")
				}
			}
		}
	}

	return nil
}

// getUpdatedSections returns a comma-separated list of updated sections for audit logging.
func getUpdatedSections(req *UpdateConfigRequest) string {
	var sections []string
	if req.Info != nil {
		sections = append(sections, "info")
	}
	if req.Limits != nil {
		sections = append(sections, "limits")
	}
	if req.Authorization != nil {
		sections = append(sections, "authorization")
	}
	if len(sections) == 0 {
		return "none"
	}
	result := sections[0]
	for i := 1; i < len(sections); i++ {
		result += "," + sections[i]
	}
	return result
}
