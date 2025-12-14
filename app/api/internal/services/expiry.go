package services

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/roostr/roostr/app/api/internal/db"
	"github.com/roostr/roostr/app/api/internal/relay"
)

// ExpiryService handles automatic subscription expiry processing.
// It runs daily at midnight, finds expired paid users, marks them as expired,
// removes them from the whitelist, and syncs the relay config.
type ExpiryService struct {
	db        *db.DB
	configMgr *relay.ConfigManager
	relay     *relay.Relay
	stopCh    chan struct{}
	wg        sync.WaitGroup
	running   bool
	mu        sync.Mutex
}

// NewExpiryService creates a new expiry service.
func NewExpiryService(database *db.DB, configMgr *relay.ConfigManager, relayCtl *relay.Relay) *ExpiryService {
	return &ExpiryService{
		db:        database,
		configMgr: configMgr,
		relay:     relayCtl,
		stopCh:    make(chan struct{}),
	}
}

// Start begins the background expiry job.
// It runs at midnight daily to process expired subscriptions.
func (s *ExpiryService) Start() {
	s.mu.Lock()
	if s.running {
		s.mu.Unlock()
		return
	}
	s.running = true
	s.stopCh = make(chan struct{})
	s.mu.Unlock()

	s.wg.Add(1)
	go s.run()
}

// Stop gracefully stops the background expiry job.
func (s *ExpiryService) Stop() {
	s.mu.Lock()
	if !s.running {
		s.mu.Unlock()
		return
	}
	s.running = false
	close(s.stopCh)
	s.mu.Unlock()

	s.wg.Wait()
}

// run is the main loop for the expiry job.
func (s *ExpiryService) run() {
	defer s.wg.Done()

	log.Println("Expiry service started")

	// Calculate time until next midnight
	now := time.Now()
	nextMidnight := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, now.Location())
	timeUntilMidnight := nextMidnight.Sub(now)

	// Create a timer for the first run at midnight
	timer := time.NewTimer(timeUntilMidnight)
	defer timer.Stop()

	for {
		select {
		case <-s.stopCh:
			log.Println("Expiry service stopped")
			return
		case <-timer.C:
			// Run the expiry job
			s.processExpiredSubscriptions()

			// Reset timer for next day
			timer.Reset(24 * time.Hour)
		}
	}
}

// processExpiredSubscriptions finds and processes all expired paid users.
func (s *ExpiryService) processExpiredSubscriptions() {
	ctx := context.Background()

	log.Println("Starting expiry job")

	// Get expired users (active + expires_at < now)
	expired, err := s.db.GetExpiredPaidUsers(ctx)
	if err != nil {
		log.Printf("Failed to get expired paid users: %v", err)
		return
	}

	if len(expired) == 0 {
		log.Println("Expiry job: no expired subscriptions")
		return
	}

	// Process each expired user
	for _, user := range expired {
		log.Printf("Subscription expired: %s (tier: %s)", user.Npub, user.Tier)

		// Mark as expired
		if err := s.db.UpdatePaidUserStatus(ctx, user.Pubkey, "expired"); err != nil {
			log.Printf("Failed to update status for %s: %v", user.Pubkey, err)
			continue
		}

		// Remove from whitelist
		if err := s.db.RemoveWhitelistEntry(ctx, user.Pubkey); err != nil {
			log.Printf("Failed to remove %s from whitelist: %v", user.Pubkey, err)
		}

		// Audit log
		s.db.AddAuditLog(ctx, "subscription_expired", map[string]interface{}{
			"pubkey": user.Pubkey,
			"npub":   user.Npub,
			"tier":   user.Tier,
		}, "")
	}

	// Sync whitelist to config.toml and reload relay
	if err := s.syncWhitelist(ctx); err != nil {
		log.Printf("Warning: failed to sync whitelist: %v", err)
	}

	log.Printf("Expiry job completed: processed %d expired subscriptions", len(expired))
}

// syncWhitelist syncs the whitelist from DB to config.toml and reloads the relay.
func (s *ExpiryService) syncWhitelist(ctx context.Context) error {
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

// RunNow forces an immediate execution of the expiry job.
// This is useful for testing or manual triggers.
func (s *ExpiryService) RunNow() {
	go s.processExpiredSubscriptions()
}

// IsRunning returns whether the expiry service is currently running.
func (s *ExpiryService) IsRunning() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.running
}
