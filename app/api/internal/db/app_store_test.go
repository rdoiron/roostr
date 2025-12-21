package db

import (
	"context"
	"os"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// setupTestDB creates a temporary SQLite database for testing
func setupTestDB(t *testing.T) *DB {
	t.Helper()

	// Create temp file for app database
	tmpFile, err := os.CreateTemp("", "roostr-test-*.db")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	tmpFile.Close()
	t.Cleanup(func() { os.Remove(tmpFile.Name()) })

	// Initialize database (empty string for relay DB since we don't need it for app_store tests)
	database, err := New("", tmpFile.Name())
	if err != nil {
		t.Fatalf("failed to create database: %v", err)
	}
	t.Cleanup(func() { database.Close() })

	// Run migrations
	if err := database.Migrate(context.Background()); err != nil {
		t.Fatalf("failed to migrate database: %v", err)
	}

	return database
}

// ============================================================================
// App State Tests
// ============================================================================

func TestAppState(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()

	t.Run("GetAppState_returns_empty_for_missing_key", func(t *testing.T) {
		value, err := db.GetAppState(ctx, "nonexistent")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if value != "" {
			t.Errorf("expected empty string, got %q", value)
		}
	})

	t.Run("SetAppState_and_GetAppState", func(t *testing.T) {
		err := db.SetAppState(ctx, "test_key", "test_value")
		if err != nil {
			t.Fatalf("failed to set app state: %v", err)
		}

		value, err := db.GetAppState(ctx, "test_key")
		if err != nil {
			t.Fatalf("failed to get app state: %v", err)
		}
		if value != "test_value" {
			t.Errorf("expected 'test_value', got %q", value)
		}
	})

	t.Run("SetAppState_overwrites_existing", func(t *testing.T) {
		db.SetAppState(ctx, "overwrite_key", "first_value")
		db.SetAppState(ctx, "overwrite_key", "second_value")

		value, _ := db.GetAppState(ctx, "overwrite_key")
		if value != "second_value" {
			t.Errorf("expected 'second_value', got %q", value)
		}
	})

	t.Run("IsSetupCompleted_false_initially", func(t *testing.T) {
		completed, err := db.IsSetupCompleted(ctx)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if completed {
			t.Error("expected setup to not be completed initially")
		}
	})

	t.Run("SetSetupCompleted", func(t *testing.T) {
		err := db.SetSetupCompleted(ctx)
		if err != nil {
			t.Fatalf("failed to set setup completed: %v", err)
		}

		completed, _ := db.IsSetupCompleted(ctx)
		if !completed {
			t.Error("expected setup to be completed")
		}
	})

	t.Run("OperatorPubkey", func(t *testing.T) {
		pubkey, _ := db.GetOperatorPubkey(ctx)
		if pubkey != "" {
			t.Error("expected empty pubkey initially")
		}

		db.SetOperatorPubkey(ctx, "abc123hex")
		pubkey, _ = db.GetOperatorPubkey(ctx)
		if pubkey != "abc123hex" {
			t.Errorf("expected 'abc123hex', got %q", pubkey)
		}
	})

	t.Run("AccessMode_defaults_to_whitelist", func(t *testing.T) {
		mode, err := db.GetAccessMode(ctx)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if mode != "whitelist" {
			t.Errorf("expected default 'whitelist', got %q", mode)
		}
	})

	t.Run("AccessMode_set_and_get", func(t *testing.T) {
		db.SetAccessMode(ctx, "paid")
		mode, _ := db.GetAccessMode(ctx)
		if mode != "paid" {
			t.Errorf("expected 'paid', got %q", mode)
		}
	})

	t.Run("AccessMode_migrates_old_names", func(t *testing.T) {
		db.SetAppState(ctx, "access_mode", "private")
		mode, _ := db.GetAccessMode(ctx)
		if mode != "whitelist" {
			t.Errorf("expected 'private' to migrate to 'whitelist', got %q", mode)
		}

		db.SetAppState(ctx, "access_mode", "public")
		mode, _ = db.GetAccessMode(ctx)
		if mode != "open" {
			t.Errorf("expected 'public' to migrate to 'open', got %q", mode)
		}
	})

	t.Run("Timezone", func(t *testing.T) {
		tz, _ := db.GetTimezone(ctx)
		if tz != "auto" {
			t.Errorf("expected default 'auto', got %q", tz)
		}

		db.SetTimezone(ctx, "America/New_York")
		tz, _ = db.GetTimezone(ctx)
		if tz != "America/New_York" {
			t.Errorf("expected 'America/New_York', got %q", tz)
		}
	})
}

// ============================================================================
// Whitelist Tests
// ============================================================================

func TestWhitelist(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()

	t.Run("GetWhitelistCount_zero_initially", func(t *testing.T) {
		count, err := db.GetWhitelistCount(ctx)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if count != 0 {
			t.Errorf("expected 0, got %d", count)
		}
	})

	t.Run("AddWhitelistEntry", func(t *testing.T) {
		entry := WhitelistEntry{
			Pubkey:     "pubkey1",
			Npub:       "npub1test",
			Nickname:   "Test User",
			IsOperator: false,
			AddedBy:    "admin",
		}
		err := db.AddWhitelistEntry(ctx, entry)
		if err != nil {
			t.Fatalf("failed to add whitelist entry: %v", err)
		}

		count, _ := db.GetWhitelistCount(ctx)
		if count != 1 {
			t.Errorf("expected count 1, got %d", count)
		}
	})

	t.Run("GetWhitelistMeta", func(t *testing.T) {
		entries, err := db.GetWhitelistMeta(ctx)
		if err != nil {
			t.Fatalf("failed to get whitelist: %v", err)
		}
		if len(entries) == 0 {
			t.Fatal("expected at least one entry")
		}

		found := false
		for _, e := range entries {
			if e.Pubkey == "pubkey1" {
				found = true
				if e.Nickname != "Test User" {
					t.Errorf("expected nickname 'Test User', got %q", e.Nickname)
				}
			}
		}
		if !found {
			t.Error("expected to find pubkey1 in whitelist")
		}
	})

	t.Run("GetWhitelistEntryByPubkey", func(t *testing.T) {
		entry, err := db.GetWhitelistEntryByPubkey(ctx, "pubkey1")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if entry == nil {
			t.Fatal("expected entry to be found")
		}
		if entry.Nickname != "Test User" {
			t.Errorf("expected nickname 'Test User', got %q", entry.Nickname)
		}
	})

	t.Run("GetWhitelistEntryByPubkey_not_found", func(t *testing.T) {
		entry, err := db.GetWhitelistEntryByPubkey(ctx, "nonexistent")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if entry != nil {
			t.Error("expected nil for non-existent pubkey")
		}
	})

	t.Run("UpdateWhitelistNickname", func(t *testing.T) {
		err := db.UpdateWhitelistNickname(ctx, "pubkey1", "Updated Name")
		if err != nil {
			t.Fatalf("failed to update nickname: %v", err)
		}

		entry, _ := db.GetWhitelistEntryByPubkey(ctx, "pubkey1")
		if entry.Nickname != "Updated Name" {
			t.Errorf("expected 'Updated Name', got %q", entry.Nickname)
		}
	})

	t.Run("UpdateWhitelistNickname_not_found", func(t *testing.T) {
		err := db.UpdateWhitelistNickname(ctx, "nonexistent", "Name")
		if err == nil {
			t.Error("expected error for non-existent pubkey")
		}
	})

	t.Run("RemoveWhitelistEntry", func(t *testing.T) {
		err := db.RemoveWhitelistEntry(ctx, "pubkey1")
		if err != nil {
			t.Fatalf("failed to remove entry: %v", err)
		}

		entry, _ := db.GetWhitelistEntryByPubkey(ctx, "pubkey1")
		if entry != nil {
			t.Error("expected entry to be removed")
		}
	})

	t.Run("RemoveWhitelistEntry_cannot_remove_operator", func(t *testing.T) {
		// Add operator
		db.AddWhitelistEntry(ctx, WhitelistEntry{
			Pubkey:     "operator_pubkey",
			Npub:       "npub1operator",
			IsOperator: true,
		})

		err := db.RemoveWhitelistEntry(ctx, "operator_pubkey")
		if err == nil {
			t.Error("expected error when removing operator")
		}
	})

	t.Run("AddWhitelistEntry_upsert", func(t *testing.T) {
		// Add entry
		db.AddWhitelistEntry(ctx, WhitelistEntry{
			Pubkey:   "upsert_test",
			Npub:     "npub1upsert",
			Nickname: "First",
		})

		// Update with upsert (nickname should not change unless explicitly set)
		db.AddWhitelistEntry(ctx, WhitelistEntry{
			Pubkey: "upsert_test",
			Npub:   "npub1upsert_updated",
		})

		entry, _ := db.GetWhitelistEntryByPubkey(ctx, "upsert_test")
		if entry.Npub != "npub1upsert_updated" {
			t.Errorf("expected npub to be updated, got %q", entry.Npub)
		}
		// Nickname should be preserved via COALESCE
		if entry.Nickname != "First" {
			t.Errorf("expected nickname to be preserved, got %q", entry.Nickname)
		}
	})
}

// ============================================================================
// Blacklist Tests
// ============================================================================

func TestBlacklist(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()

	t.Run("GetBlacklist_empty_initially", func(t *testing.T) {
		entries, err := db.GetBlacklist(ctx)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(entries) != 0 {
			t.Errorf("expected empty blacklist, got %d entries", len(entries))
		}
	})

	t.Run("AddBlacklistEntry", func(t *testing.T) {
		entry := BlacklistEntry{
			Pubkey: "bad_actor",
			Npub:   "npub1bad",
			Reason: "Spam",
		}
		err := db.AddBlacklistEntry(ctx, entry)
		if err != nil {
			t.Fatalf("failed to add blacklist entry: %v", err)
		}

		entries, _ := db.GetBlacklist(ctx)
		if len(entries) != 1 {
			t.Fatalf("expected 1 entry, got %d", len(entries))
		}
		if entries[0].Reason != "Spam" {
			t.Errorf("expected reason 'Spam', got %q", entries[0].Reason)
		}
	})

	t.Run("AddBlacklistEntry_upsert_reason", func(t *testing.T) {
		db.AddBlacklistEntry(ctx, BlacklistEntry{
			Pubkey: "bad_actor",
			Npub:   "npub1bad",
			Reason: "Updated reason",
		})

		entries, _ := db.GetBlacklist(ctx)
		for _, e := range entries {
			if e.Pubkey == "bad_actor" && e.Reason != "Updated reason" {
				t.Errorf("expected reason to be updated, got %q", e.Reason)
			}
		}
	})

	t.Run("RemoveBlacklistEntry", func(t *testing.T) {
		err := db.RemoveBlacklistEntry(ctx, "bad_actor")
		if err != nil {
			t.Fatalf("failed to remove blacklist entry: %v", err)
		}

		entries, _ := db.GetBlacklist(ctx)
		for _, e := range entries {
			if e.Pubkey == "bad_actor" {
				t.Error("expected entry to be removed")
			}
		}
	})
}

// ============================================================================
// Paid Users Tests
// ============================================================================

func TestPaidUsers(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()

	t.Run("GetPaidUsers_empty_initially", func(t *testing.T) {
		users, err := db.GetPaidUsers(ctx)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(users) != 0 {
			t.Errorf("expected empty list, got %d users", len(users))
		}
	})

	t.Run("AddPaidUser_with_expiry", func(t *testing.T) {
		expiresAt := time.Now().Add(30 * 24 * time.Hour)
		user := PaidUser{
			Pubkey:     "paid_user1",
			Npub:       "npub1paid",
			Tier:       "monthly",
			AmountSats: 1000,
			Status:     "active",
			ExpiresAt:  &expiresAt,
		}
		err := db.AddPaidUser(ctx, user)
		if err != nil {
			t.Fatalf("failed to add paid user: %v", err)
		}

		retrieved, _ := db.GetPaidUserByPubkey(ctx, "paid_user1")
		if retrieved == nil {
			t.Fatal("expected user to be found")
		}
		if retrieved.Tier != "monthly" {
			t.Errorf("expected tier 'monthly', got %q", retrieved.Tier)
		}
		if retrieved.ExpiresAt == nil {
			t.Error("expected expires_at to be set")
		}
	})

	t.Run("AddPaidUser_lifetime_no_expiry", func(t *testing.T) {
		user := PaidUser{
			Pubkey:     "lifetime_user",
			Npub:       "npub1lifetime",
			Tier:       "lifetime",
			AmountSats: 10000,
			Status:     "active",
			ExpiresAt:  nil,
		}
		db.AddPaidUser(ctx, user)

		retrieved, _ := db.GetPaidUserByPubkey(ctx, "lifetime_user")
		if retrieved.ExpiresAt != nil {
			t.Error("expected nil expires_at for lifetime user")
		}
	})

	t.Run("GetPaidUserByPubkey_not_found", func(t *testing.T) {
		user, err := db.GetPaidUserByPubkey(ctx, "nonexistent")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if user != nil {
			t.Error("expected nil for non-existent user")
		}
	})

	t.Run("UpdatePaidUserStatus", func(t *testing.T) {
		db.UpdatePaidUserStatus(ctx, "paid_user1", "expired")
		user, _ := db.GetPaidUserByPubkey(ctx, "paid_user1")
		if user.Status != "expired" {
			t.Errorf("expected status 'expired', got %q", user.Status)
		}
	})

	t.Run("GetExpiredPaidUsers", func(t *testing.T) {
		// Add an expired user
		expiredTime := time.Now().Add(-24 * time.Hour)
		db.AddPaidUser(ctx, PaidUser{
			Pubkey:     "expired_user",
			Npub:       "npub1expired",
			Tier:       "monthly",
			AmountSats: 1000,
			Status:     "active",
			ExpiresAt:  &expiredTime,
		})

		expired, err := db.GetExpiredPaidUsers(ctx)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		found := false
		for _, u := range expired {
			if u.Pubkey == "expired_user" {
				found = true
			}
		}
		if !found {
			t.Error("expected to find expired user")
		}
	})

	t.Run("CountActivePaidUsers", func(t *testing.T) {
		count, err := db.CountActivePaidUsers(ctx)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		// Should have lifetime_user and expired_user with active status
		if count < 1 {
			t.Errorf("expected at least 1 active user, got %d", count)
		}
	})

	t.Run("GetPaidUsersFiltered", func(t *testing.T) {
		users, total, err := db.GetPaidUsersFiltered(ctx, "active", 10, 0)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if total < 1 {
			t.Errorf("expected at least 1 total, got %d", total)
		}
		if len(users) < 1 {
			t.Error("expected at least 1 user returned")
		}
	})
}

// ============================================================================
// Pricing Tiers Tests
// ============================================================================

func TestPricingTiers(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()

	t.Run("GetPricingTiers_returns_defaults", func(t *testing.T) {
		tiers, err := db.GetPricingTiers(ctx)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		// Default tiers should exist after migration
		if len(tiers) < 2 {
			t.Errorf("expected at least 2 default tiers, got %d", len(tiers))
		}
	})

	t.Run("UpdatePricingTier", func(t *testing.T) {
		tiers, _ := db.GetPricingTiers(ctx)
		if len(tiers) == 0 {
			t.Skip("no tiers to update")
		}

		tier := tiers[0]
		tier.AmountSats = 5000
		err := db.UpdatePricingTier(ctx, tier)
		if err != nil {
			t.Fatalf("failed to update tier: %v", err)
		}

		updated, _ := db.GetPricingTiers(ctx)
		for _, pt := range updated {
			if pt.ID == tier.ID && pt.AmountSats != 5000 {
				t.Errorf("expected amount 5000, got %d", pt.AmountSats)
			}
		}
	})
}

// ============================================================================
// Sync Jobs Tests
// ============================================================================

func TestSyncJobs(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()

	t.Run("CreateSyncJob", func(t *testing.T) {
		job := SyncJob{
			Pubkeys:    []string{"pubkey1", "pubkey2"},
			Relays:     []string{"wss://relay.example.com"},
			EventKinds: []int{1, 3},
		}
		id, err := db.CreateSyncJob(ctx, job)
		if err != nil {
			t.Fatalf("failed to create sync job: %v", err)
		}
		if id == 0 {
			t.Error("expected non-zero job ID")
		}
	})

	t.Run("GetSyncJob", func(t *testing.T) {
		job := SyncJob{
			Pubkeys: []string{"test_pubkey"},
			Relays:  []string{"wss://test.relay"},
		}
		id, _ := db.CreateSyncJob(ctx, job)

		retrieved, err := db.GetSyncJob(ctx, id)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if retrieved == nil {
			t.Fatal("expected job to be found")
		}
		if retrieved.Status != "running" {
			t.Errorf("expected status 'running', got %q", retrieved.Status)
		}
		if len(retrieved.Pubkeys) != 1 || retrieved.Pubkeys[0] != "test_pubkey" {
			t.Error("pubkeys not correctly stored/retrieved")
		}
	})

	t.Run("UpdateSyncJobProgress", func(t *testing.T) {
		job := SyncJob{Pubkeys: []string{"p"}, Relays: []string{"r"}}
		id, _ := db.CreateSyncJob(ctx, job)

		err := db.UpdateSyncJobProgress(ctx, id, 100, 90, 10)
		if err != nil {
			t.Fatalf("failed to update progress: %v", err)
		}

		retrieved, _ := db.GetSyncJob(ctx, id)
		if retrieved.EventsFetched != 100 {
			t.Errorf("expected 100 fetched, got %d", retrieved.EventsFetched)
		}
		if retrieved.EventsStored != 90 {
			t.Errorf("expected 90 stored, got %d", retrieved.EventsStored)
		}
	})

	t.Run("CompleteSyncJob", func(t *testing.T) {
		job := SyncJob{Pubkeys: []string{"p"}, Relays: []string{"r"}}
		id, _ := db.CreateSyncJob(ctx, job)

		err := db.CompleteSyncJob(ctx, id, "completed", "")
		if err != nil {
			t.Fatalf("failed to complete job: %v", err)
		}

		retrieved, _ := db.GetSyncJob(ctx, id)
		if retrieved.Status != "completed" {
			t.Errorf("expected status 'completed', got %q", retrieved.Status)
		}
		if retrieved.CompletedAt == nil {
			t.Error("expected completed_at to be set")
		}
	})

	t.Run("GetRunningSyncJob", func(t *testing.T) {
		// Complete all existing jobs first
		jobs, _ := db.GetSyncJobs(ctx, "running", 100, 0)
		for _, j := range jobs {
			db.CompleteSyncJob(ctx, j.ID, "completed", "")
		}

		// Create a new running job
		job := SyncJob{Pubkeys: []string{"running"}, Relays: []string{"r"}}
		id, _ := db.CreateSyncJob(ctx, job)

		running, err := db.GetRunningSyncJob(ctx)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if running == nil {
			t.Fatal("expected running job to be found")
		}
		if running.ID != id {
			t.Errorf("expected job ID %d, got %d", id, running.ID)
		}
	})

	t.Run("CancelSyncJob", func(t *testing.T) {
		job := SyncJob{Pubkeys: []string{"cancel"}, Relays: []string{"r"}}
		id, _ := db.CreateSyncJob(ctx, job)

		err := db.CancelSyncJob(ctx, id)
		if err != nil {
			t.Fatalf("failed to cancel job: %v", err)
		}

		retrieved, _ := db.GetSyncJob(ctx, id)
		if retrieved.Status != "cancelled" {
			t.Errorf("expected status 'cancelled', got %q", retrieved.Status)
		}
	})
}

// ============================================================================
// Deletion Requests Tests
// ============================================================================

func TestDeletionRequests(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()

	t.Run("CreateDeletionRequest", func(t *testing.T) {
		id, err := db.CreateDeletionRequest(ctx, "event123", "admin_pubkey", "Spam content")
		if err != nil {
			t.Fatalf("failed to create deletion request: %v", err)
		}
		if id == 0 {
			t.Error("expected non-zero request ID")
		}
	})

	t.Run("GetPendingDeletionRequests", func(t *testing.T) {
		requests, err := db.GetPendingDeletionRequests(ctx)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(requests) < 1 {
			t.Error("expected at least 1 pending request")
		}
	})

	t.Run("GetPendingDeletionCount", func(t *testing.T) {
		count, err := db.GetPendingDeletionCount(ctx)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if count < 1 {
			t.Errorf("expected at least 1, got %d", count)
		}
	})

	t.Run("UpdateDeletionRequestStatus", func(t *testing.T) {
		id, _ := db.CreateDeletionRequest(ctx, "event456", "admin", "Test")

		err := db.UpdateDeletionRequestStatus(ctx, id, "processed", 1)
		if err != nil {
			t.Fatalf("failed to update status: %v", err)
		}

		requests, _ := db.GetDeletionRequests(ctx, "processed")
		found := false
		for _, r := range requests {
			if r.ID == id {
				found = true
				if r.EventsDeleted != 1 {
					t.Errorf("expected 1 event deleted, got %d", r.EventsDeleted)
				}
			}
		}
		if !found {
			t.Error("expected to find processed request")
		}
	})
}

// ============================================================================
// Retention Policy Tests
// ============================================================================

func TestRetentionPolicy(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()

	t.Run("GetRetentionPolicy_defaults", func(t *testing.T) {
		policy, err := db.GetRetentionPolicy(ctx)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if policy == nil {
			t.Fatal("expected policy to not be nil")
		}
		if policy.RetentionDays != 0 {
			t.Errorf("expected default retention_days 0, got %d", policy.RetentionDays)
		}
		if !policy.HonorNIP09 {
			t.Error("expected honor_nip09 to default to true")
		}
	})

	t.Run("SetRetentionPolicy", func(t *testing.T) {
		policy := &RetentionPolicy{
			RetentionDays: 30,
			Exceptions:    []string{"kind:0", "kind:3"},
			HonorNIP09:    false,
		}
		err := db.SetRetentionPolicy(ctx, policy)
		if err != nil {
			t.Fatalf("failed to set retention policy: %v", err)
		}

		retrieved, _ := db.GetRetentionPolicy(ctx)
		if retrieved.RetentionDays != 30 {
			t.Errorf("expected 30 days, got %d", retrieved.RetentionDays)
		}
		if len(retrieved.Exceptions) != 2 {
			t.Errorf("expected 2 exceptions, got %d", len(retrieved.Exceptions))
		}
		if retrieved.HonorNIP09 {
			t.Error("expected honor_nip09 to be false")
		}
	})

	t.Run("SetLastRetentionRun", func(t *testing.T) {
		now := time.Now()
		err := db.SetLastRetentionRun(ctx, now)
		if err != nil {
			t.Fatalf("failed to set last retention run: %v", err)
		}

		policy, _ := db.GetRetentionPolicy(ctx)
		if policy.LastRun == nil {
			t.Error("expected last_run to be set")
		}
	})
}

// ============================================================================
// Pending Invoices Tests
// ============================================================================

func TestPendingInvoices(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()

	t.Run("CreatePendingInvoice", func(t *testing.T) {
		invoice := &PendingInvoice{
			PaymentHash:    "hash123",
			Pubkey:         "user_pubkey",
			Npub:           "npub1user",
			TierID:         "monthly",
			AmountSats:     1000,
			PaymentRequest: "lnbc...",
			Memo:           "Relay access",
			ExpiresAt:      time.Now().Add(time.Hour),
		}
		err := db.CreatePendingInvoice(ctx, invoice)
		if err != nil {
			t.Fatalf("failed to create pending invoice: %v", err)
		}
	})

	t.Run("GetPendingInvoice", func(t *testing.T) {
		invoice, err := db.GetPendingInvoice(ctx, "hash123")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if invoice == nil {
			t.Fatal("expected invoice to be found")
		}
		if invoice.AmountSats != 1000 {
			t.Errorf("expected 1000 sats, got %d", invoice.AmountSats)
		}
	})

	t.Run("GetPendingInvoice_not_found", func(t *testing.T) {
		invoice, err := db.GetPendingInvoice(ctx, "nonexistent")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if invoice != nil {
			t.Error("expected nil for non-existent invoice")
		}
	})

	t.Run("UpdatePendingInvoiceStatus", func(t *testing.T) {
		err := db.UpdatePendingInvoiceStatus(ctx, "hash123", "paid")
		if err != nil {
			t.Fatalf("failed to update status: %v", err)
		}

		invoice, _ := db.GetPendingInvoice(ctx, "hash123")
		if invoice.Status != "paid" {
			t.Errorf("expected status 'paid', got %q", invoice.Status)
		}
		if invoice.PaidAt == nil {
			t.Error("expected paid_at to be set")
		}
	})

	t.Run("GetPendingInvoicesByPubkey", func(t *testing.T) {
		// Create another invoice for same pubkey
		db.CreatePendingInvoice(ctx, &PendingInvoice{
			PaymentHash:    "hash456",
			Pubkey:         "user_pubkey",
			Npub:           "npub1user",
			TierID:         "yearly",
			AmountSats:     10000,
			PaymentRequest: "lnbc2...",
			ExpiresAt:      time.Now().Add(time.Hour),
		})

		invoices, err := db.GetPendingInvoicesByPubkey(ctx, "user_pubkey")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(invoices) < 2 {
			t.Errorf("expected at least 2 invoices, got %d", len(invoices))
		}
	})

	t.Run("GetPendingInvoicesAwaitingPayment", func(t *testing.T) {
		invoices, err := db.GetPendingInvoicesAwaitingPayment(ctx)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		// hash123 is paid, hash456 should be pending
		for _, inv := range invoices {
			if inv.Status != "pending" {
				t.Errorf("expected only pending invoices, got status %q", inv.Status)
			}
		}
	})
}

// ============================================================================
// Lightning Configuration Tests
// ============================================================================

func TestLightningConfig(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()

	t.Run("GetLightningConfig_default_row", func(t *testing.T) {
		// Schema inserts a default row with enabled=0
		cfg, err := db.GetLightningConfig(ctx)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if cfg == nil {
			t.Fatal("expected default config row from schema")
		}
		if cfg.Enabled {
			t.Error("expected Enabled to be false initially")
		}
		if cfg.NodeType != "" {
			t.Errorf("expected empty NodeType, got %s", cfg.NodeType)
		}
	})

	t.Run("SaveLightningConfig", func(t *testing.T) {
		cfg := &LightningConfig{
			NodeType: "lnd",
			Endpoint: "localhost:8080",
			Macaroon: "hexmacaroon",
			Enabled:  true,
		}
		err := db.SaveLightningConfig(ctx, cfg)
		if err != nil {
			t.Fatalf("failed to save config: %v", err)
		}

		retrieved, _ := db.GetLightningConfig(ctx)
		if retrieved == nil {
			t.Fatal("expected config to be found")
		}
		if retrieved.NodeType != "lnd" {
			t.Errorf("expected node_type 'lnd', got %q", retrieved.NodeType)
		}
		if !retrieved.Enabled {
			t.Error("expected enabled to be true")
		}
	})

	t.Run("SetLightningEnabled", func(t *testing.T) {
		err := db.SetLightningEnabled(ctx, false)
		if err != nil {
			t.Fatalf("failed to set enabled: %v", err)
		}

		cfg, _ := db.GetLightningConfig(ctx)
		if cfg.Enabled {
			t.Error("expected enabled to be false")
		}
	})

	t.Run("SetLightningVerified", func(t *testing.T) {
		err := db.SetLightningVerified(ctx)
		if err != nil {
			t.Fatalf("failed to set verified: %v", err)
		}

		cfg, _ := db.GetLightningConfig(ctx)
		if cfg.LastVerifiedAt == nil {
			t.Error("expected last_verified_at to be set")
		}
	})
}

// ============================================================================
// Audit Log Tests
// ============================================================================

func TestAuditLog(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()

	t.Run("AddAuditLog", func(t *testing.T) {
		err := db.AddAuditLog(ctx, "test_action", map[string]string{"key": "value"}, "admin")
		if err != nil {
			t.Fatalf("failed to add audit log: %v", err)
		}
	})

	t.Run("AddAuditLog_nil_details", func(t *testing.T) {
		err := db.AddAuditLog(ctx, "simple_action", nil, "")
		if err != nil {
			t.Fatalf("failed to add audit log with nil details: %v", err)
		}
	})
}

// ============================================================================
// Storage Management Tests
// ============================================================================

func TestStorageManagement(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()

	t.Run("LastVacuumRun", func(t *testing.T) {
		// Initially nil
		ts, err := db.GetLastVacuumRun(ctx)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if ts != nil {
			t.Error("expected nil initially")
		}

		// Set it
		now := time.Now()
		db.SetLastVacuumRun(ctx, now)

		ts, _ = db.GetLastVacuumRun(ctx)
		if ts == nil {
			t.Error("expected timestamp to be set")
		}
	})

	t.Run("LastIntegrityCheck", func(t *testing.T) {
		ts, _ := db.GetLastIntegrityCheck(ctx)
		if ts != nil {
			t.Error("expected nil initially")
		}

		now := time.Now()
		db.SetLastIntegrityCheck(ctx, now)

		ts, _ = db.GetLastIntegrityCheck(ctx)
		if ts == nil {
			t.Error("expected timestamp to be set")
		}
	})

	t.Run("RunAppVacuum", func(t *testing.T) {
		err := db.RunAppVacuum(ctx)
		if err != nil {
			t.Fatalf("failed to run vacuum: %v", err)
		}
	})

	t.Run("RunAppIntegrityCheck", func(t *testing.T) {
		ok, result, err := db.RunAppIntegrityCheck(ctx)
		if err != nil {
			t.Fatalf("failed to run integrity check: %v", err)
		}
		if !ok {
			t.Errorf("expected integrity check to pass, got result: %s", result)
		}
	})
}
