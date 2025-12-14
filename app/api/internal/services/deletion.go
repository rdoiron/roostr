package services

import (
	"context"
	"log"

	"github.com/roostr/roostr/app/api/internal/db"
)

// DeletionService handles NIP-09 deletion request processing.
type DeletionService struct {
	db *db.DB
}

// NewDeletionService creates a new deletion service.
func NewDeletionService(database *db.DB) *DeletionService {
	return &DeletionService{
		db: database,
	}
}

// DeletionResult holds the result of processing deletion requests.
type DeletionResult struct {
	Processed     int   // Number of requests processed
	EventsDeleted int64 // Total events deleted
	Failed        int   // Number of failed requests
}

// ProcessPendingDeletions processes all pending NIP-09 deletion requests.
// It verifies that the deletion request author matches the target event author
// before deleting events.
func (s *DeletionService) ProcessPendingDeletions(ctx context.Context) (*DeletionResult, error) {
	// Check if NIP-09 deletions are honored
	policy, err := s.db.GetRetentionPolicy(ctx)
	if err != nil {
		return nil, err
	}

	if !policy.HonorNIP09 {
		log.Println("NIP-09 deletions not honored by policy, skipping")
		return &DeletionResult{}, nil
	}

	// Get pending deletion requests
	requests, err := s.db.GetPendingDeletionRequests(ctx)
	if err != nil {
		return nil, err
	}

	if len(requests) == 0 {
		return &DeletionResult{}, nil
	}

	// Open relay writer for deletions
	writer, err := s.db.NewRelayWriter()
	if err != nil {
		return nil, err
	}
	defer writer.Close()

	result := &DeletionResult{}

	for _, req := range requests {
		var eventsDeleted int64
		var validTargets []string

		// Verify each target event's author matches the request author
		for _, targetID := range req.TargetEventIDs {
			author, err := writer.GetEventAuthor(ctx, targetID)
			if err != nil {
				log.Printf("Failed to get author for event %s: %v", targetID, err)
				continue
			}

			if author == "" {
				// Event doesn't exist, consider it already deleted
				continue
			}

			if author == req.AuthorPubkey {
				// Author matches, add to valid targets
				validTargets = append(validTargets, targetID)
			} else {
				log.Printf("Rejecting deletion of event %s: requester %s is not author %s",
					targetID, req.AuthorPubkey[:16], author[:16])
			}
		}

		// Delete valid targets
		if len(validTargets) > 0 {
			deleted, err := writer.DeleteEventsByIDs(ctx, validTargets)
			if err != nil {
				log.Printf("Failed to delete events for request %d: %v", req.ID, err)
				// Mark as failed
				s.db.UpdateDeletionRequestStatus(ctx, req.ID, "failed", 0)
				result.Failed++
				continue
			}
			eventsDeleted = deleted
		}

		// Mark request as processed
		if err := s.db.UpdateDeletionRequestStatus(ctx, req.ID, "processed", eventsDeleted); err != nil {
			log.Printf("Failed to update deletion request %d status: %v", req.ID, err)
		}

		result.Processed++
		result.EventsDeleted += eventsDeleted
	}

	log.Printf("Processed %d deletion requests, deleted %d events, %d failed",
		result.Processed, result.EventsDeleted, result.Failed)

	return result, nil
}

// ProcessSingleRequest processes a specific deletion request by ID.
func (s *DeletionService) ProcessSingleRequest(ctx context.Context, requestID int64) error {
	requests, err := s.db.GetDeletionRequests(ctx, "pending")
	if err != nil {
		return err
	}

	var targetRequest *db.DeletionRequest
	for i := range requests {
		if requests[i].ID == requestID {
			targetRequest = &requests[i]
			break
		}
	}

	if targetRequest == nil {
		return nil // Request not found or not pending
	}

	// Process just this request
	writer, err := s.db.NewRelayWriter()
	if err != nil {
		return err
	}
	defer writer.Close()

	var validTargets []string
	for _, targetID := range targetRequest.TargetEventIDs {
		author, err := writer.GetEventAuthor(ctx, targetID)
		if err != nil || author == "" {
			continue
		}

		if author == targetRequest.AuthorPubkey {
			validTargets = append(validTargets, targetID)
		}
	}

	var eventsDeleted int64
	if len(validTargets) > 0 {
		eventsDeleted, err = writer.DeleteEventsByIDs(ctx, validTargets)
		if err != nil {
			s.db.UpdateDeletionRequestStatus(ctx, requestID, "failed", 0)
			return err
		}
	}

	return s.db.UpdateDeletionRequestStatus(ctx, requestID, "processed", eventsDeleted)
}
