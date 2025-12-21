package services

import (
	"context"
	"testing"
	"time"

	"github.com/roostr/roostr/app/api/internal/db"
	_ "github.com/mattn/go-sqlite3"
)

// TestDeletionService_Constructor tests the deletion service constructor.
func TestDeletionService_Constructor(t *testing.T) {
	t.Run("NewDeletionService_creates_service", func(t *testing.T) {
		database := setupTestDB(t)
		svc := NewDeletionService(database)
		if svc == nil {
			t.Fatal("expected service to be created")
		}
	})

	t.Run("NewDeletionService_nil_database", func(t *testing.T) {
		svc := NewDeletionService(nil)
		if svc == nil {
			t.Fatal("expected service to be created even with nil database")
		}
	})
}

// TestDeletionService_ProcessPendingDeletions tests the ProcessPendingDeletions function.
func TestDeletionService_ProcessPendingDeletions(t *testing.T) {
	database := setupTestDB(t)
	ctx := context.Background()

	t.Run("ProcessPendingDeletions_empty", func(t *testing.T) {
		svc := NewDeletionService(database)

		// Ensure NIP-09 is honored
		database.SetRetentionPolicy(ctx, &db.RetentionPolicy{
			HonorNIP09:    true,
			RetentionDays: 0,
		})

		result, err := svc.ProcessPendingDeletions(ctx)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.Processed != 0 {
			t.Errorf("expected 0 processed, got %d", result.Processed)
		}
		if result.EventsDeleted != 0 {
			t.Errorf("expected 0 deleted, got %d", result.EventsDeleted)
		}
	})

	t.Run("ProcessPendingDeletions_NIP09_disabled", func(t *testing.T) {
		svc := NewDeletionService(database)

		// Disable NIP-09
		database.SetRetentionPolicy(ctx, &db.RetentionPolicy{
			HonorNIP09:    false,
			RetentionDays: 0,
		})

		// Add a deletion request (eventID, requestedBy, reason)
		database.CreateDeletionRequest(ctx, "event_to_delete", "admin", "test deletion")

		result, err := svc.ProcessPendingDeletions(ctx)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		// Should skip processing because NIP-09 is disabled
		if result.Processed != 0 {
			t.Errorf("expected 0 processed when NIP-09 disabled, got %d", result.Processed)
		}
	})
}

// TestDeletionService_DeletionRequestLifecycle tests deletion request database operations.
func TestDeletionService_DeletionRequestLifecycle(t *testing.T) {
	database := setupTestDB(t)
	ctx := context.Background()

	t.Run("CreateDeletionRequest", func(t *testing.T) {
		// CreateDeletionRequest takes (ctx, eventID, requestedBy, reason)
		id, err := database.CreateDeletionRequest(ctx, "event1", "admin_pubkey", "cleanup old events")
		if err != nil {
			t.Fatalf("failed to create deletion request: %v", err)
		}
		if id <= 0 {
			t.Errorf("expected positive ID, got %d", id)
		}
	})

	t.Run("GetPendingDeletionRequests", func(t *testing.T) {
		// Add a new request
		_, err := database.CreateDeletionRequest(ctx, "eventA", "another_admin", "test deletion")
		if err != nil {
			t.Fatalf("failed to create deletion request: %v", err)
		}

		pending, err := database.GetPendingDeletionRequests(ctx)
		if err != nil {
			t.Fatalf("failed to get pending requests: %v", err)
		}
		if len(pending) == 0 {
			t.Error("expected at least one pending request")
		}
	})

	t.Run("GetPendingDeletionCount", func(t *testing.T) {
		count, err := database.GetPendingDeletionCount(ctx)
		if err != nil {
			t.Fatalf("failed to get pending count: %v", err)
		}
		if count <= 0 {
			t.Errorf("expected positive count, got %d", count)
		}
	})

	t.Run("UpdateDeletionRequestStatus", func(t *testing.T) {
		// Get a pending request
		pending, _ := database.GetPendingDeletionRequests(ctx)
		if len(pending) == 0 {
			t.Skip("no pending requests to update")
		}

		// Mark it as processed
		err := database.UpdateDeletionRequestStatus(ctx, pending[0].ID, "processed", 5)
		if err != nil {
			t.Fatalf("failed to update status: %v", err)
		}

		// Verify it's no longer pending
		stillPending, _ := database.GetPendingDeletionRequests(ctx)
		for _, p := range stillPending {
			if p.ID == pending[0].ID {
				t.Error("expected request to no longer be pending")
			}
		}
	})
}

// TestDeletionService_DeletionResult tests the DeletionResult struct.
func TestDeletionService_DeletionResult(t *testing.T) {
	t.Run("DeletionResult_defaults", func(t *testing.T) {
		result := &DeletionResult{}
		if result.Processed != 0 {
			t.Error("expected Processed to default to 0")
		}
		if result.EventsDeleted != 0 {
			t.Error("expected EventsDeleted to default to 0")
		}
		if result.Failed != 0 {
			t.Error("expected Failed to default to 0")
		}
	})
}

// TestDeletionService_RetentionPolicyIntegration tests the interaction with retention policy.
func TestDeletionService_RetentionPolicyIntegration(t *testing.T) {
	database := setupTestDB(t)
	ctx := context.Background()

	t.Run("GetRetentionPolicy_defaults", func(t *testing.T) {
		policy, err := database.GetRetentionPolicy(ctx)
		if err != nil {
			t.Fatalf("failed to get retention policy: %v", err)
		}
		// Default should honor NIP-09
		if !policy.HonorNIP09 {
			t.Error("expected default policy to honor NIP-09")
		}
	})

	t.Run("SetRetentionPolicy_honors_NIP09", func(t *testing.T) {
		err := database.SetRetentionPolicy(ctx, &db.RetentionPolicy{
			HonorNIP09:    true,
			RetentionDays: 30,
			Exceptions:    []string{"kind:0", "kind:3"}, // Keep metadata and contact lists
		})
		if err != nil {
			t.Fatalf("failed to set retention policy: %v", err)
		}

		policy, _ := database.GetRetentionPolicy(ctx)
		if !policy.HonorNIP09 {
			t.Error("expected HonorNIP09 to be true")
		}
		if policy.RetentionDays != 30 {
			t.Errorf("expected 30 days retention, got %d", policy.RetentionDays)
		}
	})

	t.Run("SetRetentionPolicy_disables_NIP09", func(t *testing.T) {
		err := database.SetRetentionPolicy(ctx, &db.RetentionPolicy{
			HonorNIP09:    false,
			RetentionDays: 0,
		})
		if err != nil {
			t.Fatalf("failed to set retention policy: %v", err)
		}

		policy, _ := database.GetRetentionPolicy(ctx)
		if policy.HonorNIP09 {
			t.Error("expected HonorNIP09 to be false")
		}
	})

	t.Run("SetLastRetentionRun", func(t *testing.T) {
		now := time.Now()
		err := database.SetLastRetentionRun(ctx, now)
		if err != nil {
			t.Fatalf("failed to set last retention run: %v", err)
		}

		policy, _ := database.GetRetentionPolicy(ctx)
		if policy.LastRun == nil {
			t.Fatal("expected LastRun to be set")
		}
		// Check within a second
		if policy.LastRun.Unix() != now.Unix() {
			t.Errorf("expected LastRun to be %v, got %v", now, policy.LastRun)
		}
	})
}
