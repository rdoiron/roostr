package services

import (
	"context"
	"testing"
	"time"

	"github.com/roostr/roostr/app/api/internal/db"
	_ "github.com/mattn/go-sqlite3"
)

// TestSyncService_Constructor tests the sync service constructor.
func TestSyncService_Constructor(t *testing.T) {
	t.Run("NewSyncService_creates_service", func(t *testing.T) {
		database := setupTestDB(t)
		svc := NewSyncService(database)
		if svc == nil {
			t.Fatal("expected service to be created")
		}
	})

	t.Run("NewSyncService_nil_database", func(t *testing.T) {
		svc := NewSyncService(nil)
		if svc == nil {
			t.Fatal("expected service to be created even with nil database")
		}
	})
}

// TestSyncService_IsRunning tests the IsRunning function.
func TestSyncService_IsRunning(t *testing.T) {
	database := setupTestDB(t)

	t.Run("IsRunning_initially_false", func(t *testing.T) {
		svc := NewSyncService(database)
		if svc.IsRunning() {
			t.Error("expected sync service to not be running initially")
		}
	})

	t.Run("GetCurrentJobID_initially_zero", func(t *testing.T) {
		svc := NewSyncService(database)
		if svc.GetCurrentJobID() != 0 {
			t.Error("expected current job ID to be 0 initially")
		}
	})
}

// TestSyncService_StartSync_Validation tests input validation for StartSync.
func TestSyncService_StartSync_Validation(t *testing.T) {
	database := setupTestDB(t)
	ctx := context.Background()

	t.Run("StartSync_requires_pubkeys", func(t *testing.T) {
		svc := NewSyncService(database)
		_, err := svc.StartSync(ctx, SyncRequest{
			Pubkeys: []string{}, // Empty pubkeys
			Relays:  DefaultSyncRelays,
		})
		if err == nil {
			t.Error("expected error for empty pubkeys")
		}
	})

	t.Run("StartSync_uses_default_relays", func(t *testing.T) {
		// Can't fully test this without network, but we can verify the structure
		req := SyncRequest{
			Pubkeys: []string{"1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"},
			Relays:  nil, // Should use defaults
		}
		if len(req.Relays) == 0 && len(DefaultSyncRelays) == 0 {
			t.Error("expected DefaultSyncRelays to be populated")
		}
	})
}

// TestSyncService_DefaultSyncRelays tests the default relays.
func TestSyncService_DefaultSyncRelays(t *testing.T) {
	t.Run("DefaultSyncRelays_is_populated", func(t *testing.T) {
		if len(DefaultSyncRelays) == 0 {
			t.Error("expected DefaultSyncRelays to have entries")
		}
	})

	t.Run("DefaultSyncRelays_all_wss", func(t *testing.T) {
		for _, relay := range DefaultSyncRelays {
			if len(relay) < 6 || relay[:6] != "wss://" {
				t.Errorf("expected relay to start with wss://, got %s", relay)
			}
		}
	})
}

// TestSyncService_CancelSync tests the CancelSync function.
func TestSyncService_CancelSync(t *testing.T) {
	database := setupTestDB(t)

	t.Run("CancelSync_when_not_running", func(t *testing.T) {
		svc := NewSyncService(database)
		err := svc.CancelSync()
		if err == nil {
			t.Error("expected error when cancelling without running job")
		}
	})
}

// TestSyncService_SyncJobDatabase tests database operations for sync jobs.
func TestSyncService_SyncJobDatabase(t *testing.T) {
	database := setupTestDB(t)
	ctx := context.Background()

	t.Run("CreateSyncJob", func(t *testing.T) {
		job := db.SyncJob{
			Pubkeys: []string{"abc123", "def456"},
			Relays:  []string{"wss://relay.example.com"},
		}

		id, err := database.CreateSyncJob(ctx, job)
		if err != nil {
			t.Fatalf("failed to create sync job: %v", err)
		}
		if id <= 0 {
			t.Errorf("expected positive job ID, got %d", id)
		}
	})

	t.Run("GetSyncJob", func(t *testing.T) {
		// Create a job first
		job := db.SyncJob{
			Pubkeys:    []string{"ghi789"},
			Relays:     []string{"wss://relay.test.com"},
			EventKinds: []int{1, 4},
		}
		id, _ := database.CreateSyncJob(ctx, job)

		// Retrieve it
		retrieved, err := database.GetSyncJob(ctx, id)
		if err != nil {
			t.Fatalf("failed to get sync job: %v", err)
		}
		if retrieved == nil {
			t.Fatal("expected job to be found")
		}
		if retrieved.Status != "running" {
			t.Errorf("expected status 'running', got '%s'", retrieved.Status)
		}
		if len(retrieved.Pubkeys) != 1 || retrieved.Pubkeys[0] != "ghi789" {
			t.Error("pubkeys mismatch")
		}
	})

	t.Run("GetSyncJob_not_found", func(t *testing.T) {
		job, err := database.GetSyncJob(ctx, 99999)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if job != nil {
			t.Error("expected nil for non-existent job")
		}
	})

	t.Run("UpdateSyncJobProgress", func(t *testing.T) {
		job := db.SyncJob{
			Pubkeys: []string{"progress_test"},
			Relays:  []string{"wss://test.com"},
		}
		id, _ := database.CreateSyncJob(ctx, job)

		// Update progress
		err := database.UpdateSyncJobProgress(ctx, id, 100, 50, 10)
		if err != nil {
			t.Fatalf("failed to update progress: %v", err)
		}

		// Verify
		updated, _ := database.GetSyncJob(ctx, id)
		if updated.EventsFetched != 100 {
			t.Errorf("expected 100 fetched, got %d", updated.EventsFetched)
		}
		if updated.EventsStored != 50 {
			t.Errorf("expected 50 stored, got %d", updated.EventsStored)
		}
		if updated.EventsSkipped != 10 {
			t.Errorf("expected 10 skipped, got %d", updated.EventsSkipped)
		}
	})

	t.Run("CompleteSyncJob_success", func(t *testing.T) {
		job := db.SyncJob{
			Pubkeys: []string{"complete_test"},
			Relays:  []string{"wss://test.com"},
		}
		id, _ := database.CreateSyncJob(ctx, job)

		// Complete successfully
		err := database.CompleteSyncJob(ctx, id, "completed", "")
		if err != nil {
			t.Fatalf("failed to complete job: %v", err)
		}

		// Verify
		completed, _ := database.GetSyncJob(ctx, id)
		if completed.Status != "completed" {
			t.Errorf("expected status 'completed', got '%s'", completed.Status)
		}
		if completed.CompletedAt == nil {
			t.Error("expected CompletedAt to be set")
		}
	})

	t.Run("CompleteSyncJob_failed", func(t *testing.T) {
		job := db.SyncJob{
			Pubkeys: []string{"failed_test"},
			Relays:  []string{"wss://test.com"},
		}
		id, _ := database.CreateSyncJob(ctx, job)

		// Complete with failure
		err := database.CompleteSyncJob(ctx, id, "failed", "connection timeout")
		if err != nil {
			t.Fatalf("failed to complete job: %v", err)
		}

		// Verify
		failed, _ := database.GetSyncJob(ctx, id)
		if failed.Status != "failed" {
			t.Errorf("expected status 'failed', got '%s'", failed.Status)
		}
		if failed.ErrorMessage != "connection timeout" {
			t.Errorf("expected error message, got '%s'", failed.ErrorMessage)
		}
	})

	t.Run("CompleteSyncJob_cancelled", func(t *testing.T) {
		job := db.SyncJob{
			Pubkeys: []string{"cancel_test"},
			Relays:  []string{"wss://test.com"},
		}
		id, _ := database.CreateSyncJob(ctx, job)

		// Complete with cancellation
		err := database.CompleteSyncJob(ctx, id, "cancelled", "")
		if err != nil {
			t.Fatalf("failed to complete job: %v", err)
		}

		// Verify
		cancelled, _ := database.GetSyncJob(ctx, id)
		if cancelled.Status != "cancelled" {
			t.Errorf("expected status 'cancelled', got '%s'", cancelled.Status)
		}
	})
}

// TestSyncService_SyncJobHistory tests sync job history retrieval.
func TestSyncService_SyncJobHistory(t *testing.T) {
	database := setupTestDB(t)
	ctx := context.Background()

	t.Run("GetSyncJobs_history", func(t *testing.T) {
		// Create multiple completed jobs
		for i := 0; i < 3; i++ {
			job := db.SyncJob{
				Pubkeys: []string{"history_test"},
				Relays:  []string{"wss://test.com"},
			}
			id, _ := database.CreateSyncJob(ctx, job)
			database.CompleteSyncJob(ctx, id, "completed", "")
		}

		// Get all jobs (empty status = all)
		jobs, err := database.GetSyncJobs(ctx, "", 10, 0)
		if err != nil {
			t.Fatalf("failed to get sync jobs: %v", err)
		}
		if len(jobs) < 3 {
			t.Errorf("expected at least 3 jobs, got %d", len(jobs))
		}
	})

	t.Run("GetSyncJobs_limited", func(t *testing.T) {
		jobs, err := database.GetSyncJobs(ctx, "", 2, 0)
		if err != nil {
			t.Fatalf("failed to get sync jobs: %v", err)
		}
		if len(jobs) > 2 {
			t.Errorf("expected max 2 jobs, got %d", len(jobs))
		}
	})
}

// TestSyncService_RunningSyncJob tests the running job detection.
func TestSyncService_RunningSyncJob(t *testing.T) {
	database := setupTestDB(t)
	ctx := context.Background()

	t.Run("GetRunningSyncJob_none", func(t *testing.T) {
		// Complete all existing jobs first
		jobs, _ := database.GetSyncJobs(ctx, "running", 100, 0)
		for _, j := range jobs {
			database.CompleteSyncJob(ctx, j.ID, "completed", "")
		}

		job, err := database.GetRunningSyncJob(ctx)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if job != nil {
			t.Error("expected no running job")
		}
	})

	t.Run("GetRunningSyncJob_exists", func(t *testing.T) {
		// Create a running job
		newJob := db.SyncJob{
			Pubkeys: []string{"running_test"},
			Relays:  []string{"wss://test.com"},
		}
		_, err := database.CreateSyncJob(ctx, newJob)
		if err != nil {
			t.Fatalf("failed to create job: %v", err)
		}

		job, err := database.GetRunningSyncJob(ctx)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if job == nil {
			t.Error("expected to find running job")
		}
	})
}

// TestSyncService_SyncJobWithTimestamp tests sync jobs with timestamp filters.
func TestSyncService_SyncJobWithTimestamp(t *testing.T) {
	database := setupTestDB(t)
	ctx := context.Background()

	t.Run("CreateSyncJob_with_since", func(t *testing.T) {
		since := time.Now().Add(-24 * time.Hour)
		job := db.SyncJob{
			Pubkeys:        []string{"timestamp_test"},
			Relays:         []string{"wss://test.com"},
			SinceTimestamp: &since,
		}

		id, err := database.CreateSyncJob(ctx, job)
		if err != nil {
			t.Fatalf("failed to create job: %v", err)
		}

		retrieved, _ := database.GetSyncJob(ctx, id)
		if retrieved.SinceTimestamp == nil {
			t.Error("expected SinceTimestamp to be set")
		}
	})
}

// TestSyncRequest tests the SyncRequest struct.
func TestSyncRequest(t *testing.T) {
	t.Run("SyncRequest_defaults", func(t *testing.T) {
		req := SyncRequest{}
		if len(req.Pubkeys) != 0 {
			t.Error("expected empty Pubkeys by default")
		}
		if len(req.Relays) != 0 {
			t.Error("expected empty Relays by default")
		}
		if len(req.EventKinds) != 0 {
			t.Error("expected empty EventKinds by default")
		}
		if req.SinceTimestamp != nil {
			t.Error("expected nil SinceTimestamp by default")
		}
	})

	t.Run("SyncRequest_with_values", func(t *testing.T) {
		since := int64(1234567890)
		req := SyncRequest{
			Pubkeys:        []string{"abc", "def"},
			Relays:         []string{"wss://relay1.com", "wss://relay2.com"},
			EventKinds:     []int{1, 4, 30023},
			SinceTimestamp: &since,
		}

		if len(req.Pubkeys) != 2 {
			t.Error("expected 2 pubkeys")
		}
		if len(req.Relays) != 2 {
			t.Error("expected 2 relays")
		}
		if len(req.EventKinds) != 3 {
			t.Error("expected 3 event kinds")
		}
		if *req.SinceTimestamp != 1234567890 {
			t.Error("expected since timestamp to match")
		}
	})
}
