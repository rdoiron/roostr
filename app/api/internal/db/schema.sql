-- Roostr App Database Schema
-- This is the application database (read-write), separate from the relay database (read-only).

-- ============================================================================
-- App State & Settings
-- ============================================================================

-- Key-value store for app settings
CREATE TABLE IF NOT EXISTS app_state (
    key TEXT PRIMARY KEY,
    value TEXT NOT NULL,
    updated_at INTEGER NOT NULL DEFAULT (strftime('%s', 'now'))
);

-- Default settings
INSERT OR IGNORE INTO app_state (key, value) VALUES
    ('setup_completed', 'false'),
    ('operator_pubkey', ''),
    ('access_mode', 'whitelist');  -- open, whitelist, paid, blacklist

-- ============================================================================
-- Whitelist Metadata
-- ============================================================================

-- Nicknames and metadata for whitelisted pubkeys
-- The actual whitelist is stored in config.toml, this is supplementary data
CREATE TABLE IF NOT EXISTS whitelist_meta (
    pubkey TEXT PRIMARY KEY,           -- hex format
    npub TEXT NOT NULL,                 -- bech32 format for display
    nickname TEXT,                      -- user-assigned nickname
    is_operator INTEGER NOT NULL DEFAULT 0,  -- 1 if this is the relay operator
    added_at INTEGER NOT NULL DEFAULT (strftime('%s', 'now')),
    added_by TEXT                       -- pubkey of who added them (null if self-added)
);

CREATE INDEX IF NOT EXISTS idx_whitelist_meta_added ON whitelist_meta(added_at);

-- ============================================================================
-- Blacklist
-- ============================================================================

-- Blacklisted pubkeys with reason
CREATE TABLE IF NOT EXISTS blacklist (
    pubkey TEXT PRIMARY KEY,           -- hex format
    npub TEXT NOT NULL,                 -- bech32 format
    reason TEXT,                        -- optional reason for blacklisting
    added_at INTEGER NOT NULL DEFAULT (strftime('%s', 'now'))
);

-- ============================================================================
-- Paid Access
-- ============================================================================

-- Users who have paid for relay access
CREATE TABLE IF NOT EXISTS paid_users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    pubkey TEXT NOT NULL UNIQUE,        -- hex format
    npub TEXT NOT NULL,                 -- bech32 format
    tier TEXT NOT NULL,                 -- 'monthly', 'yearly', 'lifetime', etc.
    amount_sats INTEGER NOT NULL,       -- amount paid in satoshis
    status TEXT NOT NULL DEFAULT 'active',  -- active, expired, cancelled
    created_at INTEGER NOT NULL DEFAULT (strftime('%s', 'now')),
    expires_at INTEGER,                 -- NULL for lifetime
    last_payment_at INTEGER NOT NULL DEFAULT (strftime('%s', 'now'))
);

CREATE INDEX IF NOT EXISTS idx_paid_users_status ON paid_users(status);
CREATE INDEX IF NOT EXISTS idx_paid_users_expires ON paid_users(expires_at);
CREATE INDEX IF NOT EXISTS idx_paid_users_pubkey ON paid_users(pubkey);

-- Payment history for paid users
CREATE TABLE IF NOT EXISTS payment_history (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    pubkey TEXT NOT NULL,               -- hex format
    payment_hash TEXT NOT NULL UNIQUE,  -- Lightning payment hash
    tier TEXT NOT NULL,
    amount_sats INTEGER NOT NULL,
    paid_at INTEGER NOT NULL DEFAULT (strftime('%s', 'now')),
    invoice TEXT,                       -- the Lightning invoice (optional, for records)
    FOREIGN KEY (pubkey) REFERENCES paid_users(pubkey) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_payment_history_pubkey ON payment_history(pubkey);
CREATE INDEX IF NOT EXISTS idx_payment_history_paid ON payment_history(paid_at);

-- Pricing configuration
CREATE TABLE IF NOT EXISTS pricing_tiers (
    id TEXT PRIMARY KEY,                -- 'monthly', 'yearly', 'lifetime'
    name TEXT NOT NULL,                 -- Display name
    amount_sats INTEGER NOT NULL,       -- Price in satoshis
    duration_days INTEGER,              -- NULL for lifetime
    enabled INTEGER NOT NULL DEFAULT 1,
    sort_order INTEGER NOT NULL DEFAULT 0
);

-- Default pricing tiers
INSERT OR IGNORE INTO pricing_tiers (id, name, amount_sats, duration_days, sort_order) VALUES
    ('monthly', 'Monthly', 5000, 30, 1),
    ('yearly', 'Yearly', 50000, 365, 2),
    ('lifetime', 'Lifetime', 100000, NULL, 3);

-- ============================================================================
-- Deletion Requests (NIP-09)
-- ============================================================================

-- Track NIP-09 deletion request events
CREATE TABLE IF NOT EXISTS deletion_requests (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    event_id TEXT NOT NULL UNIQUE,      -- The kind 5 deletion event ID
    author_pubkey TEXT NOT NULL,        -- Who requested the deletion
    target_event_ids TEXT NOT NULL,     -- JSON array of event IDs to delete
    reason TEXT,                        -- Optional deletion reason from 'content'
    status TEXT NOT NULL DEFAULT 'pending',  -- pending, processed, failed
    received_at INTEGER NOT NULL DEFAULT (strftime('%s', 'now')),
    processed_at INTEGER,
    events_deleted INTEGER DEFAULT 0    -- Count of actually deleted events
);

CREATE INDEX IF NOT EXISTS idx_deletion_requests_status ON deletion_requests(status);
CREATE INDEX IF NOT EXISTS idx_deletion_requests_author ON deletion_requests(author_pubkey);

-- ============================================================================
-- Sync Jobs
-- ============================================================================

-- History of sync operations from public relays
CREATE TABLE IF NOT EXISTS sync_jobs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    status TEXT NOT NULL DEFAULT 'running',  -- running, completed, failed, cancelled
    pubkeys TEXT NOT NULL,              -- JSON array of pubkeys being synced
    relays TEXT NOT NULL,               -- JSON array of source relay URLs
    event_kinds TEXT,                   -- JSON array of event kinds to sync (NULL = all)
    since_timestamp INTEGER,            -- Only sync events after this time
    started_at INTEGER NOT NULL DEFAULT (strftime('%s', 'now')),
    completed_at INTEGER,
    events_fetched INTEGER DEFAULT 0,
    events_stored INTEGER DEFAULT 0,    -- New events actually stored
    events_skipped INTEGER DEFAULT 0,   -- Duplicates skipped
    error_message TEXT
);

CREATE INDEX IF NOT EXISTS idx_sync_jobs_status ON sync_jobs(status);
CREATE INDEX IF NOT EXISTS idx_sync_jobs_started ON sync_jobs(started_at);

-- ============================================================================
-- Lightning Node Configuration
-- ============================================================================

-- Lightning node connection settings
CREATE TABLE IF NOT EXISTS lightning_config (
    id INTEGER PRIMARY KEY CHECK (id = 1),  -- Singleton table
    node_type TEXT,                     -- 'lnd', 'cln', 'lnbits', NULL if not configured
    endpoint TEXT,                      -- REST endpoint URL
    macaroon TEXT,                      -- Hex-encoded macaroon (encrypted at rest ideally)
    cert TEXT,                          -- TLS certificate (if needed)
    enabled INTEGER NOT NULL DEFAULT 0,
    last_verified_at INTEGER,           -- Last successful connection test
    updated_at INTEGER NOT NULL DEFAULT (strftime('%s', 'now'))
);

-- Initialize singleton row
INSERT OR IGNORE INTO lightning_config (id, enabled) VALUES (1, 0);

-- ============================================================================
-- Audit Log (Optional)
-- ============================================================================

-- Track important actions for debugging/audit
CREATE TABLE IF NOT EXISTS audit_log (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    action TEXT NOT NULL,               -- 'whitelist_add', 'whitelist_remove', 'config_change', etc.
    details TEXT,                       -- JSON with action-specific details
    performed_by TEXT,                  -- pubkey of who performed the action (if applicable)
    created_at INTEGER NOT NULL DEFAULT (strftime('%s', 'now'))
);

CREATE INDEX IF NOT EXISTS idx_audit_log_action ON audit_log(action);
CREATE INDEX IF NOT EXISTS idx_audit_log_created ON audit_log(created_at);

-- ============================================================================
-- Schema Version
-- ============================================================================

-- Track schema migrations
CREATE TABLE IF NOT EXISTS schema_version (
    version INTEGER PRIMARY KEY,
    applied_at INTEGER NOT NULL DEFAULT (strftime('%s', 'now'))
);

-- Mark initial schema version
INSERT OR IGNORE INTO schema_version (version) VALUES (1);
