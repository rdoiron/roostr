package services

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/roostr/roostr/app/api/internal/db"
)

// RetentionService handles automatic event retention and cleanup.
type RetentionService struct {
	db              *db.DB
	deletionService *DeletionService
	interval        time.Duration
	stopCh          chan struct{}
	wg              sync.WaitGroup
	running         bool
	mu              sync.Mutex
}

// NewRetentionService creates a new retention service.
func NewRetentionService(database *db.DB, deletionService *DeletionService) *RetentionService {
	return &RetentionService{
		db:              database,
		deletionService: deletionService,
		interval:        24 * time.Hour, // Run daily
		stopCh:          make(chan struct{}),
	}
}

// Start begins the background retention job.
// It runs immediately on start and then at the configured interval.
func (s *RetentionService) Start() {
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

// Stop gracefully stops the background retention job.
func (s *RetentionService) Stop() {
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

// run is the main loop for the retention job.
func (s *RetentionService) run() {
	defer s.wg.Done()

	log.Println("Retention service started")

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
			log.Println("Retention service stopped")
			return
		case <-timer.C:
			// Run the retention job
			s.runRetention()

			// Reset timer for next day
			timer.Reset(24 * time.Hour)
		}
	}
}

// runRetention executes the retention policy.
func (s *RetentionService) runRetention() {
	ctx := context.Background()

	log.Println("Starting retention job")

	// Get retention policy
	policy, err := s.db.GetRetentionPolicy(ctx)
	if err != nil {
		log.Printf("Failed to get retention policy: %v", err)
		return
	}

	// Process NIP-09 deletion requests first
	if s.deletionService != nil && policy.HonorNIP09 {
		result, err := s.deletionService.ProcessPendingDeletions(ctx)
		if err != nil {
			log.Printf("Failed to process deletion requests: %v", err)
		} else if result.Processed > 0 {
			log.Printf("Processed %d deletion requests, deleted %d events", result.Processed, result.EventsDeleted)
		}
	}

	// Check if retention policy is enabled
	if policy.RetentionDays <= 0 {
		log.Println("Retention policy disabled (keep forever)")
		s.db.SetLastRetentionRun(ctx, time.Now())
		return
	}

	// Calculate cutoff date
	cutoff := time.Now().AddDate(0, 0, -int(policy.RetentionDays))

	// Get operator pubkey for exception handling
	operatorPubkey, _ := s.db.GetOperatorPubkey(ctx)

	// Open relay writer for deletion
	writer, err := s.db.NewRelayWriter()
	if err != nil {
		log.Printf("Failed to open relay database for writing: %v", err)
		return
	}
	defer writer.Close()

	// Delete old events
	deleted, err := writer.DeleteEventsBefore(ctx, cutoff, policy.Exceptions, operatorPubkey)
	if err != nil {
		log.Printf("Failed to delete old events: %v", err)
		return
	}

	// Update last run timestamp
	s.db.SetLastRetentionRun(ctx, time.Now())

	// Add audit log
	s.db.AddAuditLog(ctx, "retention_job_run", map[string]interface{}{
		"retention_days": policy.RetentionDays,
		"cutoff":         cutoff,
		"deleted":        deleted,
	}, "")

	log.Printf("Retention job completed: deleted %d events older than %v", deleted, cutoff)
}

// RunNow forces an immediate execution of the retention policy.
// This is useful for testing or manual triggers.
func (s *RetentionService) RunNow() {
	go s.runRetention()
}

// IsRunning returns whether the retention service is currently running.
func (s *RetentionService) IsRunning() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.running
}

// GetLastRun returns the timestamp of the last retention job run.
func (s *RetentionService) GetLastRun(ctx context.Context) (*time.Time, error) {
	policy, err := s.db.GetRetentionPolicy(ctx)
	if err != nil {
		return nil, err
	}
	return policy.LastRun, nil
}
