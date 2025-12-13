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

// GetAccessMode returns the current access mode (private, paid, public).
func (d *DB) GetAccessMode(ctx context.Context) (string, error) {
	mode, err := d.GetAppState(ctx, "access_mode")
	if err != nil {
		return "", err
	}
	if mode == "" {
		return "private", nil
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
