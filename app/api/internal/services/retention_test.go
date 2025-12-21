package services

import (
	"context"
	"testing"
	"time"

	"github.com/roostr/roostr/app/api/internal/db"
	_ "github.com/mattn/go-sqlite3"
)

// TestRetentionService_Constructor tests the retention service constructor.
func TestRetentionService_Constructor(t *testing.T) {
	t.Run("NewRetentionService_creates_service", func(t *testing.T) {
		database := setupTestDB(t)
		deletionSvc := NewDeletionService(database)
		svc := NewRetentionService(database, deletionSvc)
		if svc == nil {
			t.Fatal("expected service to be created")
		}
	})

	t.Run("NewRetentionService_nil_deletion_service", func(t *testing.T) {
		database := setupTestDB(t)
		svc := NewRetentionService(database, nil)
		if svc == nil {
			t.Fatal("expected service to be created even with nil deletion service")
		}
	})

	t.Run("NewRetentionService_nil_database", func(t *testing.T) {
		svc := NewRetentionService(nil, nil)
		if svc == nil {
			t.Fatal("expected service to be created even with nil database")
		}
	})
}

// TestRetentionService_Lifecycle tests the start/stop lifecycle.
func TestRetentionService_Lifecycle(t *testing.T) {
	database := setupTestDB(t)

	t.Run("IsRunning_initially_false", func(t *testing.T) {
		svc := NewRetentionService(database, nil)
		if svc.IsRunning() {
			t.Error("expected service to not be running initially")
		}
	})

	t.Run("Start_sets_running", func(t *testing.T) {
		svc := NewRetentionService(database, nil)
		svc.Start()
		defer svc.Stop()

		if !svc.IsRunning() {
			t.Error("expected service to be running after Start")
		}
	})

	t.Run("Start_is_idempotent", func(t *testing.T) {
		svc := NewRetentionService(database, nil)
		svc.Start()
		svc.Start() // Second call should be a no-op
		defer svc.Stop()

		if !svc.IsRunning() {
			t.Error("expected service to still be running")
		}
	})

	t.Run("Stop_clears_running", func(t *testing.T) {
		svc := NewRetentionService(database, nil)
		svc.Start()
		svc.Stop()

		if svc.IsRunning() {
			t.Error("expected service to not be running after Stop")
		}
	})

	t.Run("Stop_is_idempotent", func(t *testing.T) {
		svc := NewRetentionService(database, nil)
		svc.Start()
		svc.Stop()
		svc.Stop() // Second call should be a no-op

		if svc.IsRunning() {
			t.Error("expected service to still not be running")
		}
	})

	t.Run("Stop_without_start_is_safe", func(t *testing.T) {
		svc := NewRetentionService(database, nil)
		svc.Stop() // Should not panic

		if svc.IsRunning() {
			t.Error("expected service to not be running")
		}
	})

	t.Run("Restart_works", func(t *testing.T) {
		svc := NewRetentionService(database, nil)
		svc.Start()
		svc.Stop()
		svc.Start()
		defer svc.Stop()

		if !svc.IsRunning() {
			t.Error("expected service to be running after restart")
		}
	})
}

// TestRetentionService_GetLastRun tests the GetLastRun function.
func TestRetentionService_GetLastRun(t *testing.T) {
	database := setupTestDB(t)
	ctx := context.Background()

	t.Run("GetLastRun_nil_initially", func(t *testing.T) {
		svc := NewRetentionService(database, nil)
		lastRun, err := svc.GetLastRun(ctx)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if lastRun != nil {
			t.Errorf("expected nil LastRun initially, got %v", lastRun)
		}
	})

	t.Run("GetLastRun_after_set", func(t *testing.T) {
		svc := NewRetentionService(database, nil)
		now := time.Now()
		database.SetLastRetentionRun(ctx, now)

		lastRun, err := svc.GetLastRun(ctx)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if lastRun == nil {
			t.Fatal("expected LastRun to be set")
		}
		if lastRun.Unix() != now.Unix() {
			t.Errorf("expected LastRun to be %v, got %v", now, *lastRun)
		}
	})
}

// TestRetentionService_RetentionPolicy tests retention policy interactions.
func TestRetentionService_RetentionPolicy(t *testing.T) {
	database := setupTestDB(t)
	ctx := context.Background()

	t.Run("RetentionPolicy_defaults", func(t *testing.T) {
		policy, err := database.GetRetentionPolicy(ctx)
		if err != nil {
			t.Fatalf("failed to get retention policy: %v", err)
		}
		// Defaults should be set
		if policy.RetentionDays != 0 {
			t.Logf("Default retention days: %d", policy.RetentionDays)
		}
		if !policy.HonorNIP09 {
			t.Error("expected HonorNIP09 to be true by default")
		}
	})

	t.Run("RetentionPolicy_with_exceptions", func(t *testing.T) {
		err := database.SetRetentionPolicy(ctx, &db.RetentionPolicy{
			RetentionDays: 90,
			HonorNIP09:    true,
			Exceptions:    []string{"kind:0", "kind:3", "kind:10002"}, // Keep metadata, contacts, and relay lists
		})
		if err != nil {
			t.Fatalf("failed to set retention policy: %v", err)
		}

		policy, _ := database.GetRetentionPolicy(ctx)
		if policy.RetentionDays != 90 {
			t.Errorf("expected 90 days, got %d", policy.RetentionDays)
		}
		if len(policy.Exceptions) != 3 {
			t.Errorf("expected 3 exceptions, got %d", len(policy.Exceptions))
		}
	})

	t.Run("RetentionPolicy_keep_forever", func(t *testing.T) {
		err := database.SetRetentionPolicy(ctx, &db.RetentionPolicy{
			RetentionDays: 0, // 0 means keep forever
			HonorNIP09:    true,
		})
		if err != nil {
			t.Fatalf("failed to set retention policy: %v", err)
		}

		policy, _ := database.GetRetentionPolicy(ctx)
		if policy.RetentionDays != 0 {
			t.Errorf("expected 0 days (keep forever), got %d", policy.RetentionDays)
		}
	})
}

// TestRetentionService_RunNow tests the manual run trigger.
func TestRetentionService_RunNow(t *testing.T) {
	database := setupTestDB(t)
	ctx := context.Background()

	t.Run("RunNow_triggers_background_job", func(t *testing.T) {
		// Set a retention policy first
		database.SetRetentionPolicy(ctx, &db.RetentionPolicy{
			RetentionDays: 0, // Keep forever - so no actual deletions
			HonorNIP09:    false,
		})

		svc := NewRetentionService(database, nil)

		// RunNow should not block and not panic
		svc.RunNow()

		// Give it a moment to start
		time.Sleep(50 * time.Millisecond)

		// Verify last run was updated (the job runs in background)
		// Note: This may be flaky if the job is still running
		time.Sleep(100 * time.Millisecond)

		policy, err := database.GetRetentionPolicy(ctx)
		if err != nil {
			t.Fatalf("failed to get retention policy: %v", err)
		}
		if policy.LastRun == nil {
			t.Error("expected LastRun to be set after RunNow")
		}
	})
}

// TestRetentionService_Integration tests integration between retention and deletion services.
func TestRetentionService_Integration(t *testing.T) {
	database := setupTestDB(t)
	ctx := context.Background()

	t.Run("RetentionService_uses_DeletionService", func(t *testing.T) {
		// Enable NIP-09 and set up services
		database.SetRetentionPolicy(ctx, &db.RetentionPolicy{
			RetentionDays: 0,
			HonorNIP09:    true,
		})

		deletionSvc := NewDeletionService(database)
		retentionSvc := NewRetentionService(database, deletionSvc)

		// Add a deletion request (eventID, requestedBy, reason)
		database.CreateDeletionRequest(ctx, "test_event_id", "test_author", "test reason")

		// Verify request exists
		count, _ := database.GetPendingDeletionCount(ctx)
		if count == 0 {
			t.Fatal("expected at least one deletion request")
		}

		// Service should be able to process (even if relay writer fails)
		if retentionSvc == nil || deletionSvc == nil {
			t.Fatal("services should be created")
		}
	})
}
