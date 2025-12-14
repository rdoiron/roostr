package services

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/roostr/roostr/app/api/internal/db"
	"github.com/roostr/roostr/app/api/internal/relay"
)

// InvoiceMonitorService monitors pending invoices and processes payments.
// It uses both polling and LND WebSocket subscription for payment detection.
type InvoiceMonitorService struct {
	db        *db.DB
	lightning *LightningService
	configMgr *relay.ConfigManager
	relay     *relay.Relay
	interval  time.Duration
	stopCh    chan struct{}
	wg        sync.WaitGroup
	running   bool
	mu        sync.Mutex
}

// NewInvoiceMonitorService creates a new InvoiceMonitorService.
func NewInvoiceMonitorService(
	database *db.DB,
	lightning *LightningService,
	configMgr *relay.ConfigManager,
	relayCtl *relay.Relay,
) *InvoiceMonitorService {
	return &InvoiceMonitorService{
		db:        database,
		lightning: lightning,
		configMgr: configMgr,
		relay:     relayCtl,
		interval:  10 * time.Second,
		stopCh:    make(chan struct{}),
	}
}

// Start begins the background invoice monitoring.
func (s *InvoiceMonitorService) Start() {
	s.mu.Lock()
	if s.running {
		s.mu.Unlock()
		return
	}
	s.running = true
	s.stopCh = make(chan struct{})
	s.mu.Unlock()

	// Start polling goroutine
	s.wg.Add(1)
	go s.runPoller()

	// Start subscription goroutine
	s.wg.Add(1)
	go s.runSubscription()

	log.Println("Invoice monitor service started")
}

// Stop gracefully stops the invoice monitoring.
func (s *InvoiceMonitorService) Stop() {
	s.mu.Lock()
	if !s.running {
		s.mu.Unlock()
		return
	}
	s.running = false
	close(s.stopCh)
	s.mu.Unlock()

	s.wg.Wait()
	log.Println("Invoice monitor service stopped")
}

// IsRunning returns whether the service is currently running.
func (s *InvoiceMonitorService) IsRunning() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.running
}

// runPoller polls pending invoices at regular intervals.
func (s *InvoiceMonitorService) runPoller() {
	defer s.wg.Done()

	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	// Run immediately on start
	s.checkPendingInvoices()

	for {
		select {
		case <-s.stopCh:
			return
		case <-ticker.C:
			s.checkPendingInvoices()
		}
	}
}

// runSubscription subscribes to LND invoice updates via streaming API.
func (s *InvoiceMonitorService) runSubscription() {
	defer s.wg.Done()

	for {
		select {
		case <-s.stopCh:
			return
		default:
			if !s.lightning.IsConfigured() {
				// LND not configured, wait and retry
				time.Sleep(30 * time.Second)
				continue
			}

			err := s.subscribeInvoices()
			if err != nil {
				log.Printf("Invoice subscription error: %v, reconnecting in 5s", err)
				select {
				case <-s.stopCh:
					return
				case <-time.After(5 * time.Second):
					// Continue to retry
				}
			}
		}
	}
}

// subscribeInvoices connects to LND's invoice subscription stream.
func (s *InvoiceMonitorService) subscribeInvoices() error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Monitor for stop signal
	go func() {
		<-s.stopCh
		cancel()
	}()

	return s.lightning.SubscribeInvoices(ctx, func(paymentHash string, settled bool) {
		if settled {
			log.Printf("Invoice subscription: received settled invoice %s", paymentHash)
			if err := s.ProcessPayment(context.Background(), paymentHash); err != nil {
				log.Printf("Failed to process payment from subscription: %v", err)
			}
		}
	})
}

// checkPendingInvoices checks all pending invoices with LND.
func (s *InvoiceMonitorService) checkPendingInvoices() {
	if !s.lightning.IsConfigured() {
		return
	}

	ctx := context.Background()

	// Get all pending invoices that haven't expired
	invoices, err := s.db.GetPendingInvoicesAwaitingPayment(ctx)
	if err != nil {
		log.Printf("Failed to get pending invoices: %v", err)
		return
	}

	for _, invoice := range invoices {
		// Check with LND if this invoice has been paid
		lndInvoice, err := s.lightning.CheckInvoice(ctx, invoice.PaymentHash)
		if err != nil {
			// Log but continue checking other invoices
			log.Printf("Failed to check invoice %s: %v", invoice.PaymentHash, err)
			continue
		}

		if lndInvoice.Settled {
			log.Printf("Invoice poller: detected settled invoice %s", invoice.PaymentHash)
			if err := s.ProcessPayment(ctx, invoice.PaymentHash); err != nil {
				log.Printf("Failed to process payment: %v", err)
			}
		}
	}
}

// ProcessPayment handles a confirmed payment by auto-whitelisting the user.
// This method is idempotent - safe to call multiple times for the same payment.
func (s *InvoiceMonitorService) ProcessPayment(ctx context.Context, paymentHash string) error {
	// 1. Get the pending invoice
	pending, err := s.db.GetPendingInvoice(ctx, paymentHash)
	if err != nil {
		return err
	}
	if pending == nil {
		// Invoice not found in our database - ignore (might be unrelated invoice)
		return nil
	}
	if pending.Status == "paid" {
		// Already processed - idempotent
		return nil
	}

	log.Printf("Processing payment for pubkey %s (tier: %s, amount: %d sats)",
		pending.Pubkey, pending.TierID, pending.AmountSats)

	// 2. Get the pricing tier for expiry calculation
	tier, err := s.getPricingTier(ctx, pending.TierID)
	if err != nil {
		return err
	}
	if tier == nil {
		log.Printf("Warning: tier %s not found, using tier name from invoice", pending.TierID)
	}

	// 3. Calculate expiry date
	var expiresAt *time.Time
	if tier != nil && tier.DurationDays != nil {
		t := time.Now().AddDate(0, 0, *tier.DurationDays)
		expiresAt = &t
	}

	// 4. Add to whitelist
	whitelistEntry := db.WhitelistEntry{
		Pubkey:  pending.Pubkey,
		Npub:    pending.Npub,
		AddedBy: "payment:" + pending.TierID,
	}
	if err := s.db.AddWhitelistEntry(ctx, whitelistEntry); err != nil {
		log.Printf("Warning: failed to add whitelist entry: %v", err)
		// Continue - user might already be whitelisted
	}

	// 5. Create or update paid user record
	tierName := pending.TierID
	if tier != nil {
		tierName = tier.Name
	}
	paidUser := db.PaidUser{
		Pubkey:     pending.Pubkey,
		Npub:       pending.Npub,
		Tier:       tierName,
		AmountSats: pending.AmountSats,
		Status:     "active",
		ExpiresAt:  expiresAt,
	}
	if err := s.db.AddPaidUser(ctx, paidUser); err != nil {
		log.Printf("Warning: failed to add paid user: %v", err)
		// Continue - might be a renewal
	}

	// 6. Mark invoice as paid
	if err := s.db.UpdatePendingInvoiceStatus(ctx, paymentHash, "paid"); err != nil {
		log.Printf("Warning: failed to update invoice status: %v", err)
	}

	// 7. Add payment history
	if err := s.db.AddPaymentHistory(ctx, pending.Pubkey, paymentHash, pending.TierID, pending.AmountSats, pending.PaymentRequest); err != nil {
		log.Printf("Warning: failed to add payment history: %v", err)
	}

	// 8. Sync config.toml and reload relay
	if err := s.syncWhitelist(ctx); err != nil {
		log.Printf("Warning: failed to sync whitelist: %v", err)
	}

	// 9. Audit log
	s.db.AddAuditLog(ctx, "payment_confirmed", map[string]interface{}{
		"pubkey":       pending.Pubkey,
		"tier":         pending.TierID,
		"amount_sats":  pending.AmountSats,
		"payment_hash": paymentHash,
	}, "")

	log.Printf("Successfully processed payment for %s", pending.Pubkey)
	return nil
}

// getPricingTier retrieves a pricing tier by ID.
func (s *InvoiceMonitorService) getPricingTier(ctx context.Context, tierID string) (*db.PricingTier, error) {
	tiers, err := s.db.GetPricingTiers(ctx)
	if err != nil {
		return nil, err
	}
	for _, t := range tiers {
		if t.ID == tierID {
			return &t, nil
		}
	}
	return nil, nil
}

// syncWhitelist syncs the whitelist from DB to config.toml and reloads the relay.
func (s *InvoiceMonitorService) syncWhitelist(ctx context.Context) error {
	if s.configMgr == nil {
		return nil
	}

	// Get whitelist entries from DB
	entries, err := s.db.GetWhitelistMeta(ctx)
	if err != nil {
		return err
	}

	// Extract hex pubkeys for config.toml
	whitelist := make([]string, len(entries))
	for i, e := range entries {
		whitelist[i] = e.Pubkey
	}

	// Update config.toml whitelist
	if err := s.configMgr.UpdateWhitelist(whitelist); err != nil {
		return err
	}

	// Reload relay to pick up config changes
	if s.relay != nil {
		if err := s.relay.Reload(); err != nil {
			log.Printf("Warning: failed to reload relay: %v", err)
		}
	}

	return nil
}
