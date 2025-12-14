package db

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"
)

// ============================================================================
// App State
// ============================================================================

// GetAppState retrieves a value from the app_state table.
func (d *DB) GetAppState(ctx context.Context, key string) (string, error) {
	var value string
	err := d.AppDB.QueryRowContext(ctx, "SELECT value FROM app_state WHERE key = ?", key).Scan(&value)
	if err == sql.ErrNoRows {
		return "", nil
	}
	return value, err
}

// SetAppState sets a value in the app_state table.
func (d *DB) SetAppState(ctx context.Context, key, value string) error {
	_, err := d.AppDB.ExecContext(ctx, `
		INSERT INTO app_state (key, value, updated_at) VALUES (?, ?, strftime('%s', 'now'))
		ON CONFLICT(key) DO UPDATE SET value = excluded.value, updated_at = excluded.updated_at
	`, key, value)
	return err
}

// IsSetupCompleted checks if the initial setup has been completed.
func (d *DB) IsSetupCompleted(ctx context.Context) (bool, error) {
	value, err := d.GetAppState(ctx, "setup_completed")
	if err != nil {
		return false, err
	}
	return value == "true", nil
}

// SetSetupCompleted marks setup as completed.
func (d *DB) SetSetupCompleted(ctx context.Context) error {
	return d.SetAppState(ctx, "setup_completed", "true")
}

// GetOperatorPubkey returns the relay operator's pubkey.
func (d *DB) GetOperatorPubkey(ctx context.Context) (string, error) {
	return d.GetAppState(ctx, "operator_pubkey")
}

// SetOperatorPubkey sets the relay operator's pubkey.
func (d *DB) SetOperatorPubkey(ctx context.Context, pubkey string) error {
	return d.SetAppState(ctx, "operator_pubkey", pubkey)
}

// GetAccessMode returns the current access mode (open, whitelist, paid, blacklist).
func (d *DB) GetAccessMode(ctx context.Context) (string, error) {
	mode, err := d.GetAppState(ctx, "access_mode")
	if err != nil {
		return "", err
	}
	if mode == "" {
		return "whitelist", nil
	}
	// Migrate old mode names to new ones
	switch mode {
	case "private":
		return "whitelist", nil
	case "public":
		return "open", nil
	}
	return mode, nil
}

// SetAccessMode sets the access mode.
func (d *DB) SetAccessMode(ctx context.Context, mode string) error {
	return d.SetAppState(ctx, "access_mode", mode)
}

// ============================================================================
// Whitelist Metadata
// ============================================================================

// GetWhitelistCount returns the number of entries in the whitelist.
func (d *DB) GetWhitelistCount(ctx context.Context) (int64, error) {
	var count int64
	err := d.AppDB.QueryRowContext(ctx, "SELECT COUNT(*) FROM whitelist_meta").Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

// WhitelistEntry represents a whitelisted pubkey with metadata.
type WhitelistEntry struct {
	Pubkey     string    `json:"pubkey"`
	Npub       string    `json:"npub"`
	Nickname   string    `json:"nickname,omitempty"`
	IsOperator bool      `json:"is_operator"`
	AddedAt    time.Time `json:"added_at"`
	AddedBy    string    `json:"added_by,omitempty"`
}

// GetWhitelistMeta retrieves all whitelist metadata.
func (d *DB) GetWhitelistMeta(ctx context.Context) ([]WhitelistEntry, error) {
	rows, err := d.AppDB.QueryContext(ctx, `
		SELECT pubkey, npub, nickname, is_operator, added_at, added_by
		FROM whitelist_meta
		ORDER BY is_operator DESC, added_at ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []WhitelistEntry
	for rows.Next() {
		var e WhitelistEntry
		var nickname, addedBy sql.NullString
		var addedAt int64

		err := rows.Scan(&e.Pubkey, &e.Npub, &nickname, &e.IsOperator, &addedAt, &addedBy)
		if err != nil {
			return nil, err
		}

		e.Nickname = nickname.String
		e.AddedBy = addedBy.String
		e.AddedAt = time.Unix(addedAt, 0)
		entries = append(entries, e)
	}

	return entries, rows.Err()
}

// GetWhitelistEntryByPubkey retrieves a single whitelist entry.
func (d *DB) GetWhitelistEntryByPubkey(ctx context.Context, pubkey string) (*WhitelistEntry, error) {
	var e WhitelistEntry
	var nickname, addedBy sql.NullString
	var addedAt int64

	err := d.AppDB.QueryRowContext(ctx, `
		SELECT pubkey, npub, nickname, is_operator, added_at, added_by
		FROM whitelist_meta WHERE pubkey = ?
	`, pubkey).Scan(&e.Pubkey, &e.Npub, &nickname, &e.IsOperator, &addedAt, &addedBy)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	e.Nickname = nickname.String
	e.AddedBy = addedBy.String
	e.AddedAt = time.Unix(addedAt, 0)
	return &e, nil
}

// AddWhitelistEntry adds or updates a whitelist entry.
func (d *DB) AddWhitelistEntry(ctx context.Context, entry WhitelistEntry) error {
	_, err := d.AppDB.ExecContext(ctx, `
		INSERT INTO whitelist_meta (pubkey, npub, nickname, is_operator, added_at, added_by)
		VALUES (?, ?, ?, ?, strftime('%s', 'now'), ?)
		ON CONFLICT(pubkey) DO UPDATE SET
			npub = excluded.npub,
			nickname = COALESCE(excluded.nickname, whitelist_meta.nickname)
	`, entry.Pubkey, entry.Npub, nullString(entry.Nickname), entry.IsOperator, nullString(entry.AddedBy))
	return err
}

// UpdateWhitelistNickname updates the nickname for a whitelist entry.
func (d *DB) UpdateWhitelistNickname(ctx context.Context, pubkey, nickname string) error {
	result, err := d.AppDB.ExecContext(ctx, `
		UPDATE whitelist_meta SET nickname = ? WHERE pubkey = ?
	`, nullString(nickname), pubkey)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("pubkey not found in whitelist")
	}
	return nil
}

// RemoveWhitelistEntry removes a whitelist entry.
func (d *DB) RemoveWhitelistEntry(ctx context.Context, pubkey string) error {
	// Prevent removing operator
	var isOperator bool
	err := d.AppDB.QueryRowContext(ctx, "SELECT is_operator FROM whitelist_meta WHERE pubkey = ?", pubkey).Scan(&isOperator)
	if err == sql.ErrNoRows {
		return fmt.Errorf("pubkey not found in whitelist")
	}
	if err != nil {
		return err
	}
	if isOperator {
		return fmt.Errorf("cannot remove operator from whitelist")
	}

	_, err = d.AppDB.ExecContext(ctx, "DELETE FROM whitelist_meta WHERE pubkey = ?", pubkey)
	return err
}

// ============================================================================
// Blacklist
// ============================================================================

// BlacklistEntry represents a blacklisted pubkey.
type BlacklistEntry struct {
	Pubkey  string    `json:"pubkey"`
	Npub    string    `json:"npub"`
	Reason  string    `json:"reason,omitempty"`
	AddedAt time.Time `json:"added_at"`
}

// GetBlacklist retrieves all blacklist entries.
func (d *DB) GetBlacklist(ctx context.Context) ([]BlacklistEntry, error) {
	rows, err := d.AppDB.QueryContext(ctx, `
		SELECT pubkey, npub, reason, added_at FROM blacklist ORDER BY added_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []BlacklistEntry
	for rows.Next() {
		var e BlacklistEntry
		var reason sql.NullString
		var addedAt int64

		err := rows.Scan(&e.Pubkey, &e.Npub, &reason, &addedAt)
		if err != nil {
			return nil, err
		}

		e.Reason = reason.String
		e.AddedAt = time.Unix(addedAt, 0)
		entries = append(entries, e)
	}

	return entries, rows.Err()
}

// AddBlacklistEntry adds a pubkey to the blacklist.
func (d *DB) AddBlacklistEntry(ctx context.Context, entry BlacklistEntry) error {
	_, err := d.AppDB.ExecContext(ctx, `
		INSERT INTO blacklist (pubkey, npub, reason, added_at)
		VALUES (?, ?, ?, strftime('%s', 'now'))
		ON CONFLICT(pubkey) DO UPDATE SET reason = excluded.reason
	`, entry.Pubkey, entry.Npub, nullString(entry.Reason))
	return err
}

// RemoveBlacklistEntry removes a pubkey from the blacklist.
func (d *DB) RemoveBlacklistEntry(ctx context.Context, pubkey string) error {
	_, err := d.AppDB.ExecContext(ctx, "DELETE FROM blacklist WHERE pubkey = ?", pubkey)
	return err
}

// ============================================================================
// Paid Users
// ============================================================================

// PaidUser represents a user with paid access.
type PaidUser struct {
	ID            int64     `json:"id"`
	Pubkey        string    `json:"pubkey"`
	Npub          string    `json:"npub"`
	Tier          string    `json:"tier"`
	AmountSats    int64     `json:"amount_sats"`
	Status        string    `json:"status"`
	CreatedAt     time.Time `json:"created_at"`
	ExpiresAt     *time.Time `json:"expires_at,omitempty"`
	LastPaymentAt time.Time `json:"last_payment_at"`
}

// GetPaidUsers retrieves all paid users.
func (d *DB) GetPaidUsers(ctx context.Context) ([]PaidUser, error) {
	rows, err := d.AppDB.QueryContext(ctx, `
		SELECT id, pubkey, npub, tier, amount_sats, status, created_at, expires_at, last_payment_at
		FROM paid_users ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []PaidUser
	for rows.Next() {
		var u PaidUser
		var createdAt, expiresAt, lastPaymentAt sql.NullInt64

		err := rows.Scan(&u.ID, &u.Pubkey, &u.Npub, &u.Tier, &u.AmountSats, &u.Status, &createdAt, &expiresAt, &lastPaymentAt)
		if err != nil {
			return nil, err
		}

		if createdAt.Valid {
			u.CreatedAt = time.Unix(createdAt.Int64, 0)
		}
		if expiresAt.Valid {
			t := time.Unix(expiresAt.Int64, 0)
			u.ExpiresAt = &t
		}
		if lastPaymentAt.Valid {
			u.LastPaymentAt = time.Unix(lastPaymentAt.Int64, 0)
		}
		users = append(users, u)
	}

	return users, rows.Err()
}

// GetPaidUserByPubkey retrieves a paid user by pubkey.
func (d *DB) GetPaidUserByPubkey(ctx context.Context, pubkey string) (*PaidUser, error) {
	var u PaidUser
	var createdAt, expiresAt, lastPaymentAt sql.NullInt64

	err := d.AppDB.QueryRowContext(ctx, `
		SELECT id, pubkey, npub, tier, amount_sats, status, created_at, expires_at, last_payment_at
		FROM paid_users WHERE pubkey = ?
	`, pubkey).Scan(&u.ID, &u.Pubkey, &u.Npub, &u.Tier, &u.AmountSats, &u.Status, &createdAt, &expiresAt, &lastPaymentAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	if createdAt.Valid {
		u.CreatedAt = time.Unix(createdAt.Int64, 0)
	}
	if expiresAt.Valid {
		t := time.Unix(expiresAt.Int64, 0)
		u.ExpiresAt = &t
	}
	if lastPaymentAt.Valid {
		u.LastPaymentAt = time.Unix(lastPaymentAt.Int64, 0)
	}

	return &u, nil
}

// AddPaidUser adds a new paid user.
func (d *DB) AddPaidUser(ctx context.Context, user PaidUser) error {
	var expiresAt interface{}
	if user.ExpiresAt != nil {
		expiresAt = user.ExpiresAt.Unix()
	}

	_, err := d.AppDB.ExecContext(ctx, `
		INSERT INTO paid_users (pubkey, npub, tier, amount_sats, status, created_at, expires_at, last_payment_at)
		VALUES (?, ?, ?, ?, ?, strftime('%s', 'now'), ?, strftime('%s', 'now'))
	`, user.Pubkey, user.Npub, user.Tier, user.AmountSats, user.Status, expiresAt)
	return err
}

// UpdatePaidUserStatus updates a paid user's status.
func (d *DB) UpdatePaidUserStatus(ctx context.Context, pubkey, status string) error {
	_, err := d.AppDB.ExecContext(ctx, `
		UPDATE paid_users SET status = ? WHERE pubkey = ?
	`, status, pubkey)
	return err
}

// GetExpiredPaidUsers returns paid users whose access has expired.
func (d *DB) GetExpiredPaidUsers(ctx context.Context) ([]PaidUser, error) {
	rows, err := d.AppDB.QueryContext(ctx, `
		SELECT id, pubkey, npub, tier, amount_sats, status, created_at, expires_at, last_payment_at
		FROM paid_users
		WHERE status = 'active' AND expires_at IS NOT NULL AND expires_at < strftime('%s', 'now')
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []PaidUser
	for rows.Next() {
		var u PaidUser
		var createdAt, expiresAt, lastPaymentAt sql.NullInt64

		err := rows.Scan(&u.ID, &u.Pubkey, &u.Npub, &u.Tier, &u.AmountSats, &u.Status, &createdAt, &expiresAt, &lastPaymentAt)
		if err != nil {
			return nil, err
		}

		if createdAt.Valid {
			u.CreatedAt = time.Unix(createdAt.Int64, 0)
		}
		if expiresAt.Valid {
			t := time.Unix(expiresAt.Int64, 0)
			u.ExpiresAt = &t
		}
		if lastPaymentAt.Valid {
			u.LastPaymentAt = time.Unix(lastPaymentAt.Int64, 0)
		}
		users = append(users, u)
	}

	return users, rows.Err()
}

// ============================================================================
// Pricing Tiers
// ============================================================================

// PricingTier represents a pricing tier for paid access.
type PricingTier struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	AmountSats   int64  `json:"amount_sats"`
	DurationDays *int   `json:"duration_days,omitempty"`
	Enabled      bool   `json:"enabled"`
	SortOrder    int    `json:"sort_order"`
}

// GetPricingTiers retrieves all pricing tiers.
func (d *DB) GetPricingTiers(ctx context.Context) ([]PricingTier, error) {
	rows, err := d.AppDB.QueryContext(ctx, `
		SELECT id, name, amount_sats, duration_days, enabled, sort_order
		FROM pricing_tiers ORDER BY sort_order
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tiers []PricingTier
	for rows.Next() {
		var t PricingTier
		var durationDays sql.NullInt64

		err := rows.Scan(&t.ID, &t.Name, &t.AmountSats, &durationDays, &t.Enabled, &t.SortOrder)
		if err != nil {
			return nil, err
		}

		if durationDays.Valid {
			days := int(durationDays.Int64)
			t.DurationDays = &days
		}
		tiers = append(tiers, t)
	}

	return tiers, rows.Err()
}

// UpdatePricingTier updates a pricing tier.
func (d *DB) UpdatePricingTier(ctx context.Context, tier PricingTier) error {
	var durationDays interface{}
	if tier.DurationDays != nil {
		durationDays = *tier.DurationDays
	}

	_, err := d.AppDB.ExecContext(ctx, `
		UPDATE pricing_tiers
		SET name = ?, amount_sats = ?, duration_days = ?, enabled = ?, sort_order = ?
		WHERE id = ?
	`, tier.Name, tier.AmountSats, durationDays, tier.Enabled, tier.SortOrder, tier.ID)
	return err
}

// ============================================================================
// Sync Jobs
// ============================================================================

// SyncJob represents a sync job from public relays.
type SyncJob struct {
	ID            int64     `json:"id"`
	Status        string    `json:"status"`
	Pubkeys       []string  `json:"pubkeys"`
	Relays        []string  `json:"relays"`
	EventKinds    []int     `json:"event_kinds,omitempty"`
	SinceTimestamp *time.Time `json:"since_timestamp,omitempty"`
	StartedAt     time.Time `json:"started_at"`
	CompletedAt   *time.Time `json:"completed_at,omitempty"`
	EventsFetched int64     `json:"events_fetched"`
	EventsStored  int64     `json:"events_stored"`
	EventsSkipped int64     `json:"events_skipped"`
	ErrorMessage  string    `json:"error_message,omitempty"`
}

// CreateSyncJob creates a new sync job.
func (d *DB) CreateSyncJob(ctx context.Context, job SyncJob) (int64, error) {
	pubkeysJSON, _ := json.Marshal(job.Pubkeys)
	relaysJSON, _ := json.Marshal(job.Relays)
	var kindsJSON []byte
	if len(job.EventKinds) > 0 {
		kindsJSON, _ = json.Marshal(job.EventKinds)
	}
	var sinceTimestamp interface{}
	if job.SinceTimestamp != nil {
		sinceTimestamp = job.SinceTimestamp.Unix()
	}

	result, err := d.AppDB.ExecContext(ctx, `
		INSERT INTO sync_jobs (status, pubkeys, relays, event_kinds, since_timestamp)
		VALUES ('running', ?, ?, ?, ?)
	`, string(pubkeysJSON), string(relaysJSON), nullString(string(kindsJSON)), sinceTimestamp)
	if err != nil {
		return 0, err
	}

	return result.LastInsertId()
}

// UpdateSyncJobProgress updates the progress of a sync job.
func (d *DB) UpdateSyncJobProgress(ctx context.Context, id int64, fetched, stored, skipped int64) error {
	_, err := d.AppDB.ExecContext(ctx, `
		UPDATE sync_jobs
		SET events_fetched = ?, events_stored = ?, events_skipped = ?
		WHERE id = ?
	`, fetched, stored, skipped, id)
	return err
}

// CompleteSyncJob marks a sync job as completed.
func (d *DB) CompleteSyncJob(ctx context.Context, id int64, status string, errorMsg string) error {
	_, err := d.AppDB.ExecContext(ctx, `
		UPDATE sync_jobs
		SET status = ?, completed_at = strftime('%s', 'now'), error_message = ?
		WHERE id = ?
	`, status, nullString(errorMsg), id)
	return err
}

// GetSyncJob retrieves a sync job by ID.
func (d *DB) GetSyncJob(ctx context.Context, id int64) (*SyncJob, error) {
	var job SyncJob
	var pubkeysJSON, relaysJSON string
	var kindsJSON sql.NullString
	var startedAt, completedAt, sinceTimestamp sql.NullInt64
	var errorMsg sql.NullString

	err := d.AppDB.QueryRowContext(ctx, `
		SELECT id, status, pubkeys, relays, event_kinds, since_timestamp, started_at, completed_at,
		       events_fetched, events_stored, events_skipped, error_message
		FROM sync_jobs WHERE id = ?
	`, id).Scan(&job.ID, &job.Status, &pubkeysJSON, &relaysJSON, &kindsJSON, &sinceTimestamp,
		&startedAt, &completedAt, &job.EventsFetched, &job.EventsStored, &job.EventsSkipped, &errorMsg)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	json.Unmarshal([]byte(pubkeysJSON), &job.Pubkeys)
	json.Unmarshal([]byte(relaysJSON), &job.Relays)
	if kindsJSON.Valid {
		json.Unmarshal([]byte(kindsJSON.String), &job.EventKinds)
	}
	if sinceTimestamp.Valid {
		t := time.Unix(sinceTimestamp.Int64, 0)
		job.SinceTimestamp = &t
	}
	if startedAt.Valid {
		job.StartedAt = time.Unix(startedAt.Int64, 0)
	}
	if completedAt.Valid {
		t := time.Unix(completedAt.Int64, 0)
		job.CompletedAt = &t
	}
	job.ErrorMessage = errorMsg.String

	return &job, nil
}

// ============================================================================
// Deletion Requests
// ============================================================================

// DeletionRequest represents a queued event deletion request.
type DeletionRequest struct {
	ID             int64      `json:"id"`
	EventID        string     `json:"event_id"`
	AuthorPubkey   string     `json:"author_pubkey"`
	TargetEventIDs []string   `json:"target_event_ids"`
	Reason         string     `json:"reason,omitempty"`
	Status         string     `json:"status"`
	ReceivedAt     time.Time  `json:"received_at"`
	ProcessedAt    *time.Time `json:"processed_at,omitempty"`
	EventsDeleted  int64      `json:"events_deleted"`
}

// CreateDeletionRequest queues an event for deletion.
// Returns the request ID.
func (d *DB) CreateDeletionRequest(ctx context.Context, eventID, requestedBy, reason string) (int64, error) {
	// Use a unique identifier for admin-initiated deletions
	adminRequestID := fmt.Sprintf("admin-%d", time.Now().UnixNano())

	// Store as JSON array for compatibility with NIP-09 format
	targetIDs, _ := json.Marshal([]string{eventID})

	result, err := d.AppDB.ExecContext(ctx, `
		INSERT INTO deletion_requests (event_id, author_pubkey, target_event_ids, reason, status)
		VALUES (?, ?, ?, ?, 'pending')
	`, adminRequestID, requestedBy, string(targetIDs), nullString(reason))

	if err != nil {
		return 0, fmt.Errorf("failed to create deletion request: %w", err)
	}

	return result.LastInsertId()
}

// GetPendingDeletionRequests retrieves all pending deletion requests.
func (d *DB) GetPendingDeletionRequests(ctx context.Context) ([]DeletionRequest, error) {
	return d.GetDeletionRequests(ctx, "pending")
}

// GetDeletionRequests retrieves deletion requests by status.
// Pass empty string for status to get all requests.
func (d *DB) GetDeletionRequests(ctx context.Context, status string) ([]DeletionRequest, error) {
	var query string
	var args []interface{}

	if status == "" {
		query = `
			SELECT id, event_id, author_pubkey, target_event_ids, reason, status, received_at, processed_at, events_deleted
			FROM deletion_requests
			ORDER BY received_at DESC
		`
	} else {
		query = `
			SELECT id, event_id, author_pubkey, target_event_ids, reason, status, received_at, processed_at, events_deleted
			FROM deletion_requests
			WHERE status = ?
			ORDER BY received_at DESC
		`
		args = append(args, status)
	}

	rows, err := d.AppDB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var requests []DeletionRequest
	for rows.Next() {
		var r DeletionRequest
		var targetIDsJSON string
		var reason sql.NullString
		var receivedAt int64
		var processedAt sql.NullInt64

		err := rows.Scan(&r.ID, &r.EventID, &r.AuthorPubkey, &targetIDsJSON, &reason, &r.Status, &receivedAt, &processedAt, &r.EventsDeleted)
		if err != nil {
			return nil, err
		}

		json.Unmarshal([]byte(targetIDsJSON), &r.TargetEventIDs)
		r.Reason = reason.String
		r.ReceivedAt = time.Unix(receivedAt, 0)
		if processedAt.Valid {
			t := time.Unix(processedAt.Int64, 0)
			r.ProcessedAt = &t
		}
		requests = append(requests, r)
	}

	return requests, rows.Err()
}

// GetPendingDeletionCount returns the count of pending deletion requests.
func (d *DB) GetPendingDeletionCount(ctx context.Context) (int64, error) {
	var count int64
	err := d.AppDB.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM deletion_requests WHERE status = 'pending'
	`).Scan(&count)
	return count, err
}

// UpdateDeletionRequestStatus updates the status of a deletion request.
func (d *DB) UpdateDeletionRequestStatus(ctx context.Context, id int64, status string, eventsDeleted int64) error {
	_, err := d.AppDB.ExecContext(ctx, `
		UPDATE deletion_requests
		SET status = ?, processed_at = strftime('%s', 'now'), events_deleted = ?
		WHERE id = ?
	`, status, eventsDeleted, id)
	return err
}

// ============================================================================
// Storage Management
// ============================================================================

// RetentionPolicy represents the storage retention settings.
type RetentionPolicy struct {
	RetentionDays int64    `json:"retention_days"` // 0 = keep forever
	Exceptions    []string `json:"exceptions"`     // e.g., ["kind:0", "kind:3", "pubkey:operator"]
	HonorNIP09    bool     `json:"honor_nip09"`
	LastRun       *time.Time `json:"last_run,omitempty"`
}

// GetRetentionPolicy retrieves the current retention policy settings.
func (d *DB) GetRetentionPolicy(ctx context.Context) (*RetentionPolicy, error) {
	policy := &RetentionPolicy{}

	// Get retention_days
	daysStr, err := d.GetAppState(ctx, "retention_days")
	if err != nil {
		return nil, fmt.Errorf("failed to get retention_days: %w", err)
	}
	if daysStr != "" {
		fmt.Sscanf(daysStr, "%d", &policy.RetentionDays)
	}

	// Get retention_exceptions
	exceptionsStr, err := d.GetAppState(ctx, "retention_exceptions")
	if err != nil {
		return nil, fmt.Errorf("failed to get retention_exceptions: %w", err)
	}
	if exceptionsStr != "" {
		json.Unmarshal([]byte(exceptionsStr), &policy.Exceptions)
	}

	// Get honor_nip09
	honorStr, err := d.GetAppState(ctx, "honor_nip09")
	if err != nil {
		return nil, fmt.Errorf("failed to get honor_nip09: %w", err)
	}
	policy.HonorNIP09 = honorStr != "false"

	// Get last_retention_run
	lastRunStr, err := d.GetAppState(ctx, "last_retention_run")
	if err != nil {
		return nil, fmt.Errorf("failed to get last_retention_run: %w", err)
	}
	if lastRunStr != "" && lastRunStr != "0" {
		var ts int64
		fmt.Sscanf(lastRunStr, "%d", &ts)
		if ts > 0 {
			t := time.Unix(ts, 0)
			policy.LastRun = &t
		}
	}

	return policy, nil
}

// SetRetentionPolicy saves the retention policy settings.
func (d *DB) SetRetentionPolicy(ctx context.Context, policy *RetentionPolicy) error {
	// Set retention_days
	if err := d.SetAppState(ctx, "retention_days", fmt.Sprintf("%d", policy.RetentionDays)); err != nil {
		return fmt.Errorf("failed to set retention_days: %w", err)
	}

	// Set retention_exceptions
	exceptionsJSON, _ := json.Marshal(policy.Exceptions)
	if err := d.SetAppState(ctx, "retention_exceptions", string(exceptionsJSON)); err != nil {
		return fmt.Errorf("failed to set retention_exceptions: %w", err)
	}

	// Set honor_nip09
	honorStr := "true"
	if !policy.HonorNIP09 {
		honorStr = "false"
	}
	if err := d.SetAppState(ctx, "honor_nip09", honorStr); err != nil {
		return fmt.Errorf("failed to set honor_nip09: %w", err)
	}

	return nil
}

// SetLastRetentionRun updates the timestamp of the last retention job run.
func (d *DB) SetLastRetentionRun(ctx context.Context, t time.Time) error {
	return d.SetAppState(ctx, "last_retention_run", fmt.Sprintf("%d", t.Unix()))
}

// GetLastVacuumRun returns the timestamp of the last VACUUM operation.
func (d *DB) GetLastVacuumRun(ctx context.Context) (*time.Time, error) {
	ts, err := d.GetAppState(ctx, "last_vacuum_run")
	if err != nil {
		return nil, err
	}
	if ts == "" || ts == "0" {
		return nil, nil
	}
	var unix int64
	fmt.Sscanf(ts, "%d", &unix)
	if unix > 0 {
		t := time.Unix(unix, 0)
		return &t, nil
	}
	return nil, nil
}

// SetLastVacuumRun updates the timestamp of the last VACUUM operation.
func (d *DB) SetLastVacuumRun(ctx context.Context, t time.Time) error {
	return d.SetAppState(ctx, "last_vacuum_run", fmt.Sprintf("%d", t.Unix()))
}

// GetLastIntegrityCheck returns the timestamp of the last integrity check.
func (d *DB) GetLastIntegrityCheck(ctx context.Context) (*time.Time, error) {
	ts, err := d.GetAppState(ctx, "last_integrity_check")
	if err != nil {
		return nil, err
	}
	if ts == "" || ts == "0" {
		return nil, nil
	}
	var unix int64
	fmt.Sscanf(ts, "%d", &unix)
	if unix > 0 {
		t := time.Unix(unix, 0)
		return &t, nil
	}
	return nil, nil
}

// SetLastIntegrityCheck updates the timestamp of the last integrity check.
func (d *DB) SetLastIntegrityCheck(ctx context.Context, t time.Time) error {
	return d.SetAppState(ctx, "last_integrity_check", fmt.Sprintf("%d", t.Unix()))
}

// RunAppVacuum runs VACUUM on the app database to reclaim space.
func (d *DB) RunAppVacuum(ctx context.Context) error {
	_, err := d.AppDB.ExecContext(ctx, "VACUUM")
	if err != nil {
		return fmt.Errorf("failed to vacuum app database: %w", err)
	}
	return nil
}

// RunAppIntegrityCheck runs an integrity check on the app database.
func (d *DB) RunAppIntegrityCheck(ctx context.Context) (bool, string, error) {
	var result string
	err := d.AppDB.QueryRowContext(ctx, "PRAGMA integrity_check").Scan(&result)
	if err != nil {
		return false, "", fmt.Errorf("failed to run integrity check: %w", err)
	}
	return result == "ok", result, nil
}

// ============================================================================
// Audit Log
// ============================================================================

// AddAuditLog adds an entry to the audit log.
func (d *DB) AddAuditLog(ctx context.Context, action string, details interface{}, performedBy string) error {
	var detailsJSON []byte
	if details != nil {
		detailsJSON, _ = json.Marshal(details)
	}

	_, err := d.AppDB.ExecContext(ctx, `
		INSERT INTO audit_log (action, details, performed_by) VALUES (?, ?, ?)
	`, action, nullString(string(detailsJSON)), nullString(performedBy))
	return err
}

// ============================================================================
// Helpers
// ============================================================================

func nullString(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}
