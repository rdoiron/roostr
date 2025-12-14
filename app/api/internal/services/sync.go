package services

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/roostr/roostr/app/api/internal/db"
	"github.com/roostr/roostr/app/api/internal/nostr"
)

// DefaultSyncRelays are the default public relays to sync from.
var DefaultSyncRelays = []string{
	"wss://relay.damus.io",
	"wss://relay.nostr.band",
	"wss://nos.lol",
	"wss://relay.primal.net",
	"wss://nostr.wine",
	"wss://relay.snort.social",
}

// SyncService handles syncing events from public relays.
type SyncService struct {
	db       *db.DB
	mu       sync.Mutex
	cancelFn context.CancelFunc
	jobID    int64
	running  bool
}

// NewSyncService creates a new sync service.
func NewSyncService(database *db.DB) *SyncService {
	return &SyncService{db: database}
}

// SyncRequest contains parameters for starting a sync job.
type SyncRequest struct {
	Pubkeys        []string `json:"pubkeys"`
	Relays         []string `json:"relays"`
	EventKinds     []int    `json:"event_kinds,omitempty"`
	SinceTimestamp *int64   `json:"since_timestamp,omitempty"`
}

// StartSync begins a new sync job.
// Returns the job ID and an error if a sync is already in progress.
func (s *SyncService) StartSync(ctx context.Context, req SyncRequest) (int64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.running {
		return 0, fmt.Errorf("a sync job is already running")
	}

	// Validate request
	if len(req.Pubkeys) == 0 {
		return 0, fmt.Errorf("at least one pubkey is required")
	}
	if len(req.Relays) == 0 {
		req.Relays = DefaultSyncRelays
	}

	// Create job record
	job := db.SyncJob{
		Pubkeys:    req.Pubkeys,
		Relays:     req.Relays,
		EventKinds: req.EventKinds,
	}
	if req.SinceTimestamp != nil {
		t := time.Unix(*req.SinceTimestamp, 0)
		job.SinceTimestamp = &t
	}

	jobID, err := s.db.CreateSyncJob(ctx, job)
	if err != nil {
		return 0, fmt.Errorf("failed to create sync job: %w", err)
	}

	// Create cancellable context for the job
	jobCtx, cancel := context.WithCancel(context.Background())
	s.cancelFn = cancel
	s.jobID = jobID
	s.running = true

	// Start background sync
	go s.runSync(jobCtx, jobID, req)

	return jobID, nil
}

// runSync is the background goroutine that performs the actual sync.
func (s *SyncService) runSync(ctx context.Context, jobID int64, req SyncRequest) {
	defer func() {
		s.mu.Lock()
		s.running = false
		s.cancelFn = nil
		s.jobID = 0
		s.mu.Unlock()
	}()

	var totalFetched, totalStored, totalSkipped int64
	var lastError string
	finalStatus := "completed"

	// Open relay writer for insertions
	writer, err := s.db.NewRelayWriter()
	if err != nil {
		log.Printf("Sync job %d: failed to open relay writer: %v", jobID, err)
		s.db.CompleteSyncJob(ctx, jobID, "failed", fmt.Sprintf("failed to open relay writer: %v", err))
		return
	}
	defer writer.Close()

	// Progress update helper
	updateProgress := func() {
		s.db.UpdateSyncJobProgress(ctx, jobID, totalFetched, totalStored, totalSkipped)
	}

	// For each relay
	for _, relayURL := range req.Relays {
		// Check cancellation
		select {
		case <-ctx.Done():
			finalStatus = "cancelled"
			goto done
		default:
		}

		log.Printf("Sync job %d: connecting to %s", jobID, relayURL)

		// Connect to relay
		client := nostr.NewClient(relayURL)
		if err := client.Connect(ctx); err != nil {
			log.Printf("Sync job %d: failed to connect to %s: %v", jobID, relayURL, err)
			lastError = fmt.Sprintf("failed to connect to %s: %v", relayURL, err)
			continue
		}

		// For each pubkey
		for _, pubkey := range req.Pubkeys {
			// Check cancellation
			select {
			case <-ctx.Done():
				client.Close()
				finalStatus = "cancelled"
				goto done
			default:
			}

			log.Printf("Sync job %d: syncing pubkey %s from %s", jobID, pubkey[:16], relayURL)

			// Build filter
			filter := nostr.Filter{
				Authors: []string{pubkey},
			}
			if len(req.EventKinds) > 0 {
				filter.Kinds = req.EventKinds
			}
			if req.SinceTimestamp != nil {
				filter.Since = req.SinceTimestamp
			}

			// Subscribe and receive events
			err := client.Subscribe(ctx, filter, func(event *nostr.SyncEvent) error {
				totalFetched++

				// Verify event signature
				if err := event.Verify(); err != nil {
					log.Printf("Sync job %d: skipping invalid event %s: %v", jobID, event.ID[:16], err)
					totalSkipped++
					return nil
				}

				// Convert to db.Event
				dbEvent := &db.Event{
					ID:        event.ID,
					Pubkey:    event.Pubkey,
					CreatedAt: time.Unix(event.CreatedAt, 0),
					Kind:      event.Kind,
					Tags:      event.Tags,
					Content:   event.Content,
					Sig:       event.Sig,
				}

				// Insert event
				inserted, err := writer.InsertEvent(ctx, dbEvent)
				if err != nil {
					log.Printf("Sync job %d: failed to insert event %s: %v", jobID, event.ID[:16], err)
					return nil // Continue despite errors
				}

				if inserted {
					totalStored++
				} else {
					totalSkipped++
				}

				// Periodic progress update (every 100 events)
				if totalFetched%100 == 0 {
					updateProgress()
				}

				return nil
			})

			if err != nil {
				if ctx.Err() != nil {
					// Context was cancelled
					client.Close()
					finalStatus = "cancelled"
					goto done
				}
				log.Printf("Sync job %d: error syncing %s from %s: %v", jobID, pubkey[:16], relayURL, err)
			}
		}

		client.Close()
	}

done:
	// Final progress update
	updateProgress()

	// Complete the job
	if lastError != "" && finalStatus == "completed" && totalStored == 0 && totalFetched == 0 {
		finalStatus = "failed"
	}
	s.db.CompleteSyncJob(ctx, jobID, finalStatus, lastError)

	log.Printf("Sync job %d %s: fetched=%d, stored=%d, skipped=%d",
		jobID, finalStatus, totalFetched, totalStored, totalSkipped)
}

// CancelSync cancels the currently running sync job.
func (s *SyncService) CancelSync() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.running || s.cancelFn == nil {
		return fmt.Errorf("no sync job is running")
	}

	s.cancelFn()
	return nil
}

// GetCurrentJobID returns the ID of the currently running job, or 0 if none.
func (s *SyncService) GetCurrentJobID() int64 {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.jobID
}

// IsRunning returns whether a sync is currently in progress.
func (s *SyncService) IsRunning() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.running
}
