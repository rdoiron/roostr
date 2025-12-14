package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/roostr/roostr/app/api/internal/nostr"
	"github.com/roostr/roostr/app/api/internal/services"
)

// GetRelayInfo returns public information about the relay for the signup page.
// GET /public/relay-info
func (h *Handler) GetRelayInfo(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Check if paid access is enabled
	accessMode, err := h.db.GetAccessMode(ctx)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to get access mode", "DB_ERROR")
		return
	}

	if accessMode != "paid" {
		respondJSON(w, http.StatusOK, map[string]interface{}{
			"paid_access_enabled": false,
			"message":             "Paid access is not enabled for this relay",
		})
		return
	}

	// Get relay name from config
	var relayName, relayDescription string
	if h.configMgr != nil {
		cfg, _ := h.configMgr.Read()
		if cfg != nil {
			relayName = cfg.Info.Name
			relayDescription = cfg.Info.Description
		}
	}

	// Get pricing tiers
	tiers, err := h.db.GetPricingTiers(ctx)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to get pricing tiers", "DB_ERROR")
		return
	}

	// Filter to only enabled tiers
	var enabledTiers []map[string]interface{}
	for _, t := range tiers {
		if t.Enabled {
			tier := map[string]interface{}{
				"id":          t.ID,
				"name":        t.Name,
				"amount_sats": t.AmountSats,
			}
			if t.DurationDays != nil {
				tier["duration_days"] = *t.DurationDays
			}
			enabledTiers = append(enabledTiers, tier)
		}
	}

	// Check if Lightning is configured
	lnConfigured := h.services.Lightning.IsConfigured()

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"paid_access_enabled":  true,
		"lightning_configured": lnConfigured,
		"name":                 relayName,
		"description":          relayDescription,
		"tiers":                enabledTiers,
	})
}

// CreateSignupInvoice creates a Lightning invoice for relay access signup.
// POST /public/create-invoice
func (h *Handler) CreateSignupInvoice(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Check if paid access is enabled
	accessMode, err := h.db.GetAccessMode(ctx)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to get access mode", "DB_ERROR")
		return
	}

	if accessMode != "paid" {
		respondError(w, http.StatusBadRequest, "Paid access is not enabled", "PAID_ACCESS_DISABLED")
		return
	}

	var req struct {
		Pubkey string `json:"pubkey"` // hex or npub
		TierID string `json:"tier_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body", "INVALID_JSON")
		return
	}

	if req.Pubkey == "" {
		respondError(w, http.StatusBadRequest, "Pubkey is required", "MISSING_PUBKEY")
		return
	}

	if req.TierID == "" {
		respondError(w, http.StatusBadRequest, "Tier ID is required", "MISSING_TIER")
		return
	}

	// Validate and convert pubkey format
	hexPubkey, npub, err := nostr.ValidatePubkey(req.Pubkey)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid pubkey format: "+err.Error(), "INVALID_PUBKEY")
		return
	}

	// Check if already whitelisted
	existing, _ := h.db.GetWhitelistEntryByPubkey(ctx, hexPubkey)
	if existing != nil {
		respondError(w, http.StatusConflict, "This pubkey already has access to the relay", "ALREADY_WHITELISTED")
		return
	}

	// Check if already a paid user with active status
	paidUser, _ := h.db.GetPaidUserByPubkey(ctx, hexPubkey)
	if paidUser != nil && paidUser.Status == "active" {
		respondError(w, http.StatusConflict, "This pubkey already has active paid access", "ALREADY_PAID")
		return
	}

	// Create the invoice
	invoice, err := h.services.Lightning.CreateAccessInvoice(ctx, services.AccessInvoiceRequest{
		Pubkey: hexPubkey,
		Npub:   npub,
		TierID: req.TierID,
	})
	if err != nil {
		if err == services.ErrLNDNotConfigured {
			respondError(w, http.StatusServiceUnavailable, "Lightning is not configured", "LN_NOT_CONFIGURED")
			return
		}
		respondError(w, http.StatusInternalServerError, "Failed to create invoice: "+err.Error(), "INVOICE_FAILED")
		return
	}

	respondJSON(w, http.StatusCreated, map[string]interface{}{
		"payment_hash":    invoice.PaymentHash,
		"payment_request": invoice.PaymentRequest,
		"amount_sats":     invoice.AmountSats,
		"tier_id":         invoice.TierID,
		"tier_name":       invoice.TierName,
		"expires_at":      invoice.ExpiresAt,
		"memo":            invoice.Memo,
	})
}

// GetInvoiceStatus checks the status of a signup invoice.
// GET /public/invoice-status/{hash}
func (h *Handler) GetInvoiceStatus(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	paymentHash := r.PathValue("hash")
	if paymentHash == "" {
		respondError(w, http.StatusBadRequest, "Payment hash is required", "MISSING_HASH")
		return
	}

	// Get the pending invoice from database
	pendingInvoice, err := h.db.GetPendingInvoice(ctx, paymentHash)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to get invoice", "DB_ERROR")
		return
	}

	if pendingInvoice == nil {
		respondError(w, http.StatusNotFound, "Invoice not found", "INVOICE_NOT_FOUND")
		return
	}

	// If already marked as paid, return that status
	if pendingInvoice.Status == "paid" {
		respondJSON(w, http.StatusOK, map[string]interface{}{
			"status":       "paid",
			"payment_hash": paymentHash,
			"paid_at":      pendingInvoice.PaidAt.Unix(),
		})
		return
	}

	// If expired locally, return that
	if pendingInvoice.Status == "expired" {
		respondJSON(w, http.StatusOK, map[string]interface{}{
			"status":       "expired",
			"payment_hash": paymentHash,
		})
		return
	}

	// Check with LND if it's been paid
	if h.services.Lightning.IsConfigured() {
		lndInvoice, err := h.services.Lightning.CheckInvoice(ctx, paymentHash)
		if err == nil && lndInvoice.Settled {
			// Invoice was paid! Process the payment (auto-whitelist)
			if err := h.services.InvoiceMonitor.ProcessPayment(ctx, paymentHash); err != nil {
				// Log the error but still return success to the client
				// The background service will retry if needed
				log.Printf("Warning: failed to process payment from status check: %v", err)
			}

			// Re-fetch to get the updated paid_at timestamp
			updatedInvoice, _ := h.db.GetPendingInvoice(ctx, paymentHash)
			var paidAt int64
			if updatedInvoice != nil && updatedInvoice.PaidAt != nil {
				paidAt = updatedInvoice.PaidAt.Unix()
			}

			respondJSON(w, http.StatusOK, map[string]interface{}{
				"status":       "paid",
				"payment_hash": paymentHash,
				"paid_at":      paidAt,
				"message":      "Payment confirmed! You now have access to the relay.",
			})
			return
		}
	}

	// Still pending
	respondJSON(w, http.StatusOK, map[string]interface{}{
		"status":       "pending",
		"payment_hash": paymentHash,
		"expires_at":   pendingInvoice.ExpiresAt.Unix(),
	})
}
