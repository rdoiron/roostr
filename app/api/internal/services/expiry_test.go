package services

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/roostr/roostr/app/api/internal/db"
	_ "github.com/mattn/go-sqlite3"
)

// setupTestDB creates an in-memory SQLite database for testing
func setupTestDB(t *testing.T) *db.DB {
	t.Helper()

	// Create temp file for app database
	tmpFile, err := os.CreateTemp("", "roostr-test-*.db")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	tmpFile.Close()
	t.Cleanup(func() { os.Remove(tmpFile.Name()) })

	// Initialize database
	database, err := db.New("", tmpFile.Name())
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

// TestSUB001_ExpiryTracking tests that expiry tracking is correctly set up (SUB-001)
func TestSUB001_ExpiryTracking(t *testing.T) {
	database := setupTestDB(t)
	ctx := context.Background()

	t.Run("AddPaidUser_with_expiry", func(t *testing.T) {
		expiresAt := time.Now().Add(30 * 24 * time.Hour)
		user := db.PaidUser{
			Pubkey:     "abc123",
			Npub:       "npub1test",
			Tier:       "monthly",
			AmountSats: 1000,
			Status:     "active",
			ExpiresAt:  &expiresAt,
		}

		if err := database.AddPaidUser(ctx, user); err != nil {
			t.Fatalf("failed to add paid user: %v", err)
		}

		// Verify user was added
		retrieved, err := database.GetPaidUserByPubkey(ctx, "abc123")
		if err != nil {
			t.Fatalf("failed to get paid user: %v", err)
		}
		if retrieved == nil {
			t.Fatal("expected user to be found")
		}
		if retrieved.ExpiresAt == nil {
			t.Fatal("expected expires_at to be set")
		}
		if retrieved.Status != "active" {
			t.Errorf("expected status 'active', got '%s'", retrieved.Status)
		}
	})

	t.Run("AddPaidUser_lifetime_no_expiry", func(t *testing.T) {
		user := db.PaidUser{
			Pubkey:     "def456",
			Npub:       "npub1lifetime",
			Tier:       "lifetime",
			AmountSats: 10000,
			Status:     "active",
			ExpiresAt:  nil,
		}

		if err := database.AddPaidUser(ctx, user); err != nil {
			t.Fatalf("failed to add paid user: %v", err)
		}

		// Verify user was added with nil expires_at
		retrieved, err := database.GetPaidUserByPubkey(ctx, "def456")
		if err != nil {
			t.Fatalf("failed to get paid user: %v", err)
		}
		if retrieved == nil {
			t.Fatal("expected user to be found")
		}
		if retrieved.ExpiresAt != nil {
			t.Error("expected expires_at to be nil for lifetime tier")
		}
	})

	t.Run("GetExpiredPaidUsers_finds_expired", func(t *testing.T) {
		// Add an expired user
		expiredTime := time.Now().Add(-24 * time.Hour) // Expired yesterday
		expiredUser := db.PaidUser{
			Pubkey:     "expired123",
			Npub:       "npub1expired",
			Tier:       "monthly",
			AmountSats: 1000,
			Status:     "active",
			ExpiresAt:  &expiredTime,
		}
		if err := database.AddPaidUser(ctx, expiredUser); err != nil {
			t.Fatalf("failed to add expired user: %v", err)
		}

		// Get expired users
		expired, err := database.GetExpiredPaidUsers(ctx)
		if err != nil {
			t.Fatalf("failed to get expired users: %v", err)
		}

		// Should find the expired user
		found := false
		for _, u := range expired {
			if u.Pubkey == "expired123" {
				found = true
				break
			}
		}
		if !found {
			t.Error("expected to find expired user")
		}
	})

	t.Run("CountExpiringPaidUsers_counts_correctly", func(t *testing.T) {
		// Add a user expiring in 3 days
		expiringTime := time.Now().Add(3 * 24 * time.Hour)
		expiringUser := db.PaidUser{
			Pubkey:     "expiring123",
			Npub:       "npub1expiring",
			Tier:       "monthly",
			AmountSats: 1000,
			Status:     "active",
			ExpiresAt:  &expiringTime,
		}
		if err := database.AddPaidUser(ctx, expiringUser); err != nil {
			t.Fatalf("failed to add expiring user: %v", err)
		}

		// Count users expiring within 7 days
		count, err := database.CountExpiringPaidUsers(ctx, 7)
		if err != nil {
			t.Fatalf("failed to count expiring users: %v", err)
		}

		if count < 1 {
			t.Errorf("expected at least 1 user expiring within 7 days, got %d", count)
		}
	})
}

// TestSUB002_BackgroundExpiryJob tests the background expiry job (SUB-002)
func TestSUB002_BackgroundExpiryJob(t *testing.T) {
	t.Run("ExpiryService_lifecycle", func(t *testing.T) {
		database := setupTestDB(t)
		svc := NewExpiryService(database, nil, nil)

		// Initially not running
		if svc.IsRunning() {
			t.Error("expected service to not be running initially")
		}

		// Start service
		svc.Start()
		if !svc.IsRunning() {
			t.Error("expected service to be running after Start")
		}

		// Start again should be idempotent
		svc.Start()
		if !svc.IsRunning() {
			t.Error("expected service to still be running")
		}

		// Stop service
		svc.Stop()
		if svc.IsRunning() {
			t.Error("expected service to not be running after Stop")
		}

		// Stop again should be idempotent
		svc.Stop()
		if svc.IsRunning() {
			t.Error("expected service to still not be running")
		}
	})

	t.Run("ExpiryService_processes_expired_users", func(t *testing.T) {
		database := setupTestDB(t)
		ctx := context.Background()

		// Add an expired user to the database
		expiredTime := time.Now().Add(-24 * time.Hour)
		expiredUser := db.PaidUser{
			Pubkey:     "toexpire123",
			Npub:       "npub1toexpire",
			Tier:       "monthly",
			AmountSats: 1000,
			Status:     "active",
			ExpiresAt:  &expiredTime,
		}
		if err := database.AddPaidUser(ctx, expiredUser); err != nil {
			t.Fatalf("failed to add expired user: %v", err)
		}

		// Add to whitelist
		if err := database.AddWhitelistEntry(ctx, db.WhitelistEntry{
			Pubkey:  "toexpire123",
			Npub:    "npub1toexpire",
			AddedBy: "payment:monthly",
		}); err != nil {
			t.Fatalf("failed to add whitelist entry: %v", err)
		}

		// Create service and run expiry manually
		svc := NewExpiryService(database, nil, nil)
		svc.RunNow()

		// Wait a bit for the goroutine to process
		time.Sleep(100 * time.Millisecond)

		// Verify user status was updated to expired
		user, err := database.GetPaidUserByPubkey(ctx, "toexpire123")
		if err != nil {
			t.Fatalf("failed to get user: %v", err)
		}
		if user == nil {
			t.Fatal("expected user to exist")
		}
		if user.Status != "expired" {
			t.Errorf("expected status 'expired', got '%s'", user.Status)
		}

		// Verify user was removed from whitelist
		entries, err := database.GetWhitelistMeta(ctx)
		if err != nil {
			t.Fatalf("failed to get whitelist: %v", err)
		}
		for _, e := range entries {
			if e.Pubkey == "toexpire123" {
				t.Error("expected user to be removed from whitelist")
			}
		}
	})

	t.Run("ExpiryService_ignores_lifetime_users", func(t *testing.T) {
		database := setupTestDB(t)
		ctx := context.Background()

		// Add a lifetime user (no expiry)
		lifetimeUser := db.PaidUser{
			Pubkey:     "lifetime789",
			Npub:       "npub1lifetime",
			Tier:       "lifetime",
			AmountSats: 10000,
			Status:     "active",
			ExpiresAt:  nil,
		}
		if err := database.AddPaidUser(ctx, lifetimeUser); err != nil {
			t.Fatalf("failed to add lifetime user: %v", err)
		}

		// Add to whitelist
		if err := database.AddWhitelistEntry(ctx, db.WhitelistEntry{
			Pubkey:  "lifetime789",
			Npub:    "npub1lifetime",
			AddedBy: "payment:lifetime",
		}); err != nil {
			t.Fatalf("failed to add whitelist entry: %v", err)
		}

		// Run expiry job
		svc := NewExpiryService(database, nil, nil)
		svc.RunNow()
		time.Sleep(100 * time.Millisecond)

		// Verify lifetime user is still active
		user, err := database.GetPaidUserByPubkey(ctx, "lifetime789")
		if err != nil {
			t.Fatalf("failed to get user: %v", err)
		}
		if user.Status != "active" {
			t.Errorf("expected lifetime user to remain 'active', got '%s'", user.Status)
		}

		// Verify still in whitelist
		entries, err := database.GetWhitelistMeta(ctx)
		if err != nil {
			t.Fatalf("failed to get whitelist: %v", err)
		}
		found := false
		for _, e := range entries {
			if e.Pubkey == "lifetime789" {
				found = true
				break
			}
		}
		if !found {
			t.Error("expected lifetime user to remain in whitelist")
		}
	})

	t.Run("ExpiryService_ignores_already_expired", func(t *testing.T) {
		database := setupTestDB(t)
		ctx := context.Background()

		// Add an already-expired user (status already 'expired')
		expiredTime := time.Now().Add(-48 * time.Hour)
		alreadyExpiredUser := db.PaidUser{
			Pubkey:     "alreadyexpired",
			Npub:       "npub1already",
			Tier:       "monthly",
			AmountSats: 1000,
			Status:     "expired", // Already marked as expired
			ExpiresAt:  &expiredTime,
		}
		if err := database.AddPaidUser(ctx, alreadyExpiredUser); err != nil {
			t.Fatalf("failed to add already expired user: %v", err)
		}

		// Get expired users - should not include this one
		expired, err := database.GetExpiredPaidUsers(ctx)
		if err != nil {
			t.Fatalf("failed to get expired users: %v", err)
		}

		for _, u := range expired {
			if u.Pubkey == "alreadyexpired" {
				t.Error("expected already-expired user to NOT be returned by GetExpiredPaidUsers")
			}
		}
	})
}

// TestSUB003_ExpiryWarningDisplay tests expiry warning related functionality (SUB-003)
func TestSUB003_ExpiryWarningDisplay(t *testing.T) {
	database := setupTestDB(t)
	ctx := context.Background()

	t.Run("CountExpiringPaidUsers_within_7_days", func(t *testing.T) {
		// Add users with various expiry times
		users := []struct {
			pubkey    string
			daysUntil int
		}{
			{"warn1", 3},  // Should be counted (3 days)
			{"warn2", 7},  // Should be counted (exactly 7 days)
			{"warn3", 10}, // Should NOT be counted (10 days)
		}

		for _, u := range users {
			expiryTime := time.Now().Add(time.Duration(u.daysUntil) * 24 * time.Hour)
			user := db.PaidUser{
				Pubkey:     u.pubkey,
				Npub:       "npub1" + u.pubkey,
				Tier:       "monthly",
				AmountSats: 1000,
				Status:     "active",
				ExpiresAt:  &expiryTime,
			}
			if err := database.AddPaidUser(ctx, user); err != nil {
				t.Fatalf("failed to add user %s: %v", u.pubkey, err)
			}
		}

		// Count users expiring within 7 days
		count, err := database.CountExpiringPaidUsers(ctx, 7)
		if err != nil {
			t.Fatalf("failed to count expiring users: %v", err)
		}

		// Should count warn1 and warn2, but not warn3
		// Note: there might be users from other tests, so check at least 2
		if count < 2 {
			t.Errorf("expected at least 2 users expiring within 7 days, got %d", count)
		}
	})

	t.Run("ExpiresAt_returned_in_user_data", func(t *testing.T) {
		// Get all paid users and verify expires_at is populated
		users, _, err := database.GetPaidUsersFiltered(ctx, "active", 100, 0)
		if err != nil {
			t.Fatalf("failed to get paid users: %v", err)
		}

		for _, u := range users {
			if u.Tier != "lifetime" && u.ExpiresAt == nil {
				t.Errorf("expected non-lifetime user %s to have expires_at set", u.Pubkey)
			}
		}
	})
}

// TestExpiryService_Constructor tests the constructor (SUB-002)
func TestExpiryService_Constructor(t *testing.T) {
	t.Run("NewExpiryService_creates_service", func(t *testing.T) {
		svc := NewExpiryService(nil, nil, nil)
		if svc == nil {
			t.Fatal("expected service to be created")
		}
		if svc.stopCh == nil {
			t.Error("expected stopCh to be initialized")
		}
	})
}
