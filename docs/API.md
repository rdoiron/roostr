# Roostr API Reference

Complete reference for the Roostr REST API.

## Overview

**Base URL:** `/api/v1/`

**Response Format:** JSON

**Error Response Format:**
```json
{
  "error": "Error message",
  "code": "ERROR_CODE",
  "details": "Additional information (optional)"
}
```

## Table of Contents

1. [Health & Status](#health--status)
2. [Setup](#setup)
3. [Dashboard & Statistics](#dashboard--statistics)
4. [Relay Control](#relay-control)
5. [Access Control](#access-control)
6. [Whitelist](#whitelist)
7. [Blacklist](#blacklist)
8. [Pricing & Paid Access](#pricing--paid-access)
9. [NIP-05 Resolution](#nip-05-resolution)
10. [Events](#events)
11. [Export](#export)
12. [Configuration](#configuration)
13. [Settings](#settings)
14. [Storage](#storage)
15. [Sync](#sync)
16. [Lightning](#lightning)
17. [Public Signup](#public-signup)
18. [Support](#support)

---

## Health & Status

### GET /health

Check API and relay database connection status.

**Response:**
```json
{
  "status": "ok",
  "relay_connected": true
}
```

---

## Setup

### GET /api/v1/setup/status

Get initial setup completion status.

**Response:**
```json
{
  "completed": true,
  "operator_pubkey": "hex pubkey",
  "operator_npub": "npub1...",
  "access_mode": "private"
}
```

### GET /api/v1/setup/validate-identity

Validate a pubkey or NIP-05 identifier.

**Query Parameters:**
| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `input` | string | Yes | npub, hex pubkey, or NIP-05 identifier |

**Response (valid):**
```json
{
  "valid": true,
  "pubkey": "hex pubkey",
  "npub": "npub1...",
  "source": "npub|hex|nip05",
  "nip05_name": "user@example.com"
}
```

**Response (invalid):**
```json
{
  "valid": false,
  "error": "Invalid pubkey format",
  "code": "INVALID_PUBKEY"
}
```

### POST /api/v1/setup/complete

Complete the initial setup wizard.

**Request Body:**
```json
{
  "operator_identity": "npub1... or user@example.com",
  "relay_name": "My Relay",
  "relay_description": "A private Nostr relay",
  "access_mode": "private"
}
```

**Response:**
```json
{
  "success": true,
  "message": "Setup completed successfully",
  "operator_pubkey": "hex pubkey",
  "operator_npub": "npub1...",
  "access_mode": "private"
}
```

---

## Dashboard & Statistics

### GET /api/v1/stats/summary

Get aggregate relay statistics for the dashboard.

**Response:**
```json
{
  "total_events": 12345,
  "events_today": 42,
  "storage_bytes": 52428800,
  "whitelisted_count": 5,
  "events_by_kind": {
    "posts": 8000,
    "follows": 100,
    "dms": 500,
    "reposts": 1000,
    "reactions": 2500,
    "other": 245
  },
  "uptime_seconds": 86400,
  "relay_status": "online"
}
```

### GET /api/v1/stats/stream

Server-Sent Events stream for real-time dashboard updates.

**Updates:** Every 2 seconds with keepalive every 15 seconds

**Events:**
- `connected` - Initial connection established
- `stats` - Dashboard statistics update

### GET /api/v1/stats/events-over-time

Get event counts grouped by time for charts.

**Query Parameters:**
| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `time_range` | string | `7days` | `today`, `7days`, `30days`, `alltime` |
| `timezone` | string | `UTC` | IANA timezone name |

**Response:**
```json
{
  "data": [
    {"date": "2025-12-21", "count": 150},
    {"date": "2025-12-22", "count": 200}
  ],
  "time_range": "7days",
  "total": 1050
}
```

### GET /api/v1/stats/events-by-kind

Get event distribution by kind.

**Query Parameters:**
| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `time_range` | string | `alltime` | `today`, `7days`, `30days`, `alltime` |

**Response:**
```json
{
  "kinds": [
    {"kind": 1, "label": "posts", "count": 8000, "percent": 65.0},
    {"kind": 7, "label": "reactions", "count": 2500, "percent": 20.3}
  ],
  "time_range": "alltime",
  "total": 12345
}
```

### GET /api/v1/stats/top-authors

Get most active pubkeys by event count.

**Query Parameters:**
| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `limit` | int | `10` | Max 100 |
| `time_range` | string | `alltime` | `today`, `7days`, `30days`, `alltime` |

**Response:**
```json
{
  "authors": [
    {"pubkey": "hex", "npub": "npub1...", "event_count": 500}
  ],
  "time_range": "alltime",
  "limit": 10
}
```

---

## Relay Control

### GET /api/v1/relay/status

Get relay process status with resource usage.

**Response:**
```json
{
  "status": "running",
  "pid": 1234,
  "memory_bytes": 52428800,
  "uptime_seconds": 86400,
  "database_connected": true,
  "api_uptime_seconds": 86500
}
```

### GET /api/v1/relay/urls

Get relay's WebSocket connection URLs.

**Response:**
```json
{
  "local": "ws://192.168.1.100:7000",
  "relay_port": 7000,
  "tor": "ws://abcd1234.onion",
  "tor_available": true
}
```

### POST /api/v1/relay/reload

Reload relay configuration via SIGHUP.

**Response:**
```json
{
  "success": true,
  "message": "Relay configuration reloaded"
}
```

### POST /api/v1/relay/restart

Restart the relay process (async).

**Response:**
```json
{
  "success": true,
  "message": "Relay restart initiated",
  "status": "restarting"
}
```

### GET /api/v1/relay/logs

Get recent relay log entries.

**Query Parameters:**
| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `limit` | int | `100` | Max 1000 |

**Response:**
```json
{
  "logs": [
    {
      "timestamp": "2025-12-22T14:32:01Z",
      "level": "INFO",
      "message": "New connection from 192.168.1.50"
    }
  ],
  "total_lines": 500
}
```

### GET /api/v1/relay/logs/stream

Server-Sent Events stream for real-time logs.

**Events:**
- `connected` - Initial connection
- `log` - New log entry

---

## Access Control

### GET /api/v1/access/mode

Get current access control mode.

**Response:**
```json
{
  "mode": "whitelist"
}
```

### PUT /api/v1/access/mode

Set access control mode.

**Request Body:**
```json
{
  "mode": "whitelist"
}
```

Valid modes: `open`, `whitelist`, `paid`, `blacklist`

**Response:**
```json
{
  "success": true,
  "mode": "whitelist"
}
```

---

## Whitelist

### GET /api/v1/access/whitelist

Get all whitelisted pubkeys.

**Response:**
```json
{
  "entries": [
    {
      "pubkey": "hex",
      "npub": "npub1...",
      "nickname": "Alice",
      "is_operator": true,
      "event_count": 1234
    }
  ]
}
```

### POST /api/v1/access/whitelist

Add a pubkey to the whitelist.

**Request Body:**
```json
{
  "pubkey": "hex pubkey",
  "npub": "npub1... (optional)",
  "nickname": "Alice (optional)"
}
```

**Response:**
```json
{
  "success": true,
  "message": "Added to whitelist"
}
```

### DELETE /api/v1/access/whitelist/{pubkey}

Remove a pubkey from the whitelist.

**Note:** Cannot remove the operator.

**Response:**
```json
{
  "success": true,
  "message": "Removed from whitelist"
}
```

### PATCH /api/v1/access/whitelist/{pubkey}

Update a whitelist entry.

**Request Body:**
```json
{
  "nickname": "New Nickname"
}
```

**Response:**
```json
{
  "success": true,
  "message": "Whitelist entry updated"
}
```

---

## Blacklist

### GET /api/v1/access/blacklist

Get all blacklisted pubkeys.

**Response:**
```json
{
  "entries": [
    {
      "pubkey": "hex",
      "npub": "npub1...",
      "reason": "Spam"
    }
  ]
}
```

### POST /api/v1/access/blacklist

Add a pubkey to the blacklist.

**Request Body:**
```json
{
  "pubkey": "hex pubkey",
  "npub": "npub1... (optional)",
  "reason": "Spam (optional)"
}
```

**Response:**
```json
{
  "success": true,
  "message": "Added to blacklist"
}
```

### DELETE /api/v1/access/blacklist/{pubkey}

Remove a pubkey from the blacklist.

**Response:**
```json
{
  "success": true,
  "message": "Removed from blacklist"
}
```

---

## Pricing & Paid Access

### GET /api/v1/access/pricing

Get all pricing tier configurations.

**Response:**
```json
{
  "tiers": [
    {
      "id": "monthly",
      "name": "Monthly",
      "amount_sats": 5000,
      "duration_days": 30,
      "enabled": true
    },
    {
      "id": "lifetime",
      "name": "Lifetime",
      "amount_sats": 50000,
      "duration_days": null,
      "enabled": true
    }
  ]
}
```

### PUT /api/v1/access/pricing

Update pricing tiers.

**Request Body:**
```json
{
  "tiers": [
    {
      "id": "monthly",
      "name": "Monthly",
      "amount_sats": 5000,
      "duration_days": 30,
      "enabled": true
    }
  ]
}
```

**Response:**
```json
{
  "success": true,
  "message": "Pricing updated"
}
```

### GET /api/v1/access/paid-users

Get paid users with pagination.

**Query Parameters:**
| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `status` | string | `all` | `active`, `expired`, `revoked`, `all` |
| `limit` | int | `50` | Max 100 |
| `offset` | int | `0` | Pagination offset |

**Response:**
```json
{
  "users": [
    {
      "pubkey": "hex",
      "npub": "npub1...",
      "tier": "monthly",
      "status": "active",
      "created_at": "2025-12-01T00:00:00Z",
      "expires_at": "2025-12-31T00:00:00Z",
      "event_count": 500
    }
  ],
  "total": 25,
  "limit": 50,
  "offset": 0
}
```

### DELETE /api/v1/access/paid-users/{pubkey}

Revoke access for a paid user.

**Response:**
```json
{
  "success": true,
  "message": "Access revoked"
}
```

### GET /api/v1/access/revenue

Get revenue summary statistics.

**Response:**
```json
{
  "total_revenue_sats": 250000,
  "active_subscribers": 15,
  "expiring_soon": 3,
  "total_payments": 25,
  "revenue_by_tier": {
    "monthly": 100000,
    "lifetime": 150000
  }
}
```

---

## NIP-05 Resolution

### GET /api/v1/nip05/{identifier}

Resolve a NIP-05 identifier to a pubkey.

**URL Parameters:**
| Parameter | Description |
|-----------|-------------|
| `identifier` | URL-encoded NIP-05 (e.g., `user%40example.com`) |

**Response:**
```json
{
  "name": "user",
  "domain": "example.com",
  "pubkey": "hex",
  "npub": "npub1...",
  "relays": ["wss://relay1.com", "wss://relay2.com"]
}
```

---

## Events

### GET /api/v1/events

Get paginated list of events with filtering.

**Query Parameters:**
| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `limit` | int | `50` | Max results per page |
| `offset` | int | `0` | Pagination offset |
| `search` | string | - | Search in content |
| `kinds` | string | - | Comma-separated kinds (e.g., `1,7`) |
| `authors` | string | - | Comma-separated hex pubkeys |
| `since` | int | - | Unix timestamp (events after) |
| `until` | int | - | Unix timestamp (events before) |
| `mentions` | string | - | Hex pubkey to find mentions of |

**Response:**
```json
{
  "events": [
    {
      "id": "hex event id",
      "pubkey": "hex",
      "created_at": 1703260800,
      "kind": 1,
      "content": "Hello Nostr!",
      "tags": [],
      "sig": "hex signature"
    }
  ],
  "count": 100,
  "limit": 50,
  "offset": 0
}
```

### GET /api/v1/events/{id}

Get a single event by ID.

**Response:** Full event object or 404 error.

### GET /api/v1/events/recent

Get 10 most recent events for dashboard.

**Response:**
```json
{
  "events": [...]
}
```

### DELETE /api/v1/events/{id}

Queue an event for deletion (NIP-09).

**Request Body:**
```json
{
  "reason": "Optional reason"
}
```

**Response (202 Accepted):**
```json
{
  "message": "Deletion request queued",
  "request_id": "uuid",
  "event_id": "hex",
  "status": "pending"
}
```

---

## Export

### GET /api/v1/events/export

Stream events as backup.

**Query Parameters:**
| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `format` | string | `ndjson` | `ndjson` or `json` |
| `kinds` | string | - | Comma-separated kinds |
| `since` | int | - | Unix timestamp |
| `until` | int | - | Unix timestamp |

**Response Headers:**
- `Content-Type`: `application/x-ndjson` or `application/json`
- `Content-Disposition`: `attachment; filename=events-YYYYMMDD.ndjson`
- `X-Total-Count`: Total events (if available)

**Response:** Streamed events in requested format.

### GET /api/v1/events/export/estimate

Get estimate of export size before downloading.

**Query Parameters:** Same as export endpoint.

**Response:**
```json
{
  "count": 12345,
  "estimated_bytes": 52428800
}
```

---

## Configuration

### GET /api/v1/config

Get relay configuration.

**Response:**
```json
{
  "info": {
    "name": "My Relay",
    "description": "A private Nostr relay",
    "contact": "admin@example.com",
    "relay_icon": "https://example.com/icon.png"
  },
  "limits": {
    "max_event_bytes": 65536,
    "max_ws_message_bytes": 131072,
    "messages_per_sec": 5,
    "max_subs_per_conn": 20,
    "min_pow_difficulty": 0
  },
  "authorization": {
    "nip42_auth": false,
    "event_kind_allowlist": []
  }
}
```

### PATCH /api/v1/config

Partial update of relay configuration.

**Request Body:** Same structure as GET, all fields optional.

**Response:**
```json
{
  "success": true,
  "message": "Configuration updated"
}
```

### POST /api/v1/config/reload

Signal relay to reload configuration file.

**Response:**
```json
{
  "success": true,
  "message": "Relay configuration reloaded"
}
```

---

## Settings

### GET /api/v1/settings/timezone

Get user's preferred timezone.

**Response:**
```json
{
  "timezone": "America/New_York"
}
```

### PUT /api/v1/settings/timezone

Set user's preferred timezone.

**Request Body:**
```json
{
  "timezone": "America/New_York"
}
```

Valid values: `UTC`, `auto`, or any IANA timezone name.

**Response:**
```json
{
  "timezone": "America/New_York"
}
```

---

## Storage

### GET /api/v1/storage/status

Get storage usage and health status.

**Response:**
```json
{
  "database_size": 52428800,
  "app_database_size": 1048576,
  "total_size": 53477376,
  "available_space": 10737418240,
  "total_space": 107374182400,
  "usage_percent": 10.5,
  "total_events": 12345,
  "oldest_event": "2024-01-01T00:00:00Z",
  "newest_event": "2025-12-22T14:00:00Z",
  "status": "healthy",
  "pending_deletions": 5
}
```

Status values: `healthy`, `warning`, `low`, `critical`

### GET /api/v1/storage/retention

Get retention policy settings.

**Response:**
```json
{
  "retention_days": 365,
  "exceptions": ["pubkey1", "pubkey2"],
  "honor_nip09": true,
  "last_run": "2025-12-22T00:00:00Z"
}
```

### PUT /api/v1/storage/retention

Update retention policy.

**Request Body:**
```json
{
  "retention_days": 365,
  "exceptions": ["pubkey1"],
  "honor_nip09": true
}
```

**Response:**
```json
{
  "success": true,
  "message": "Retention policy updated"
}
```

### POST /api/v1/storage/cleanup

Manual cleanup of events before a date.

**Request Body:**
```json
{
  "before_date": "2024-01-01T00:00:00Z"
}
```

**Response:**
```json
{
  "success": true,
  "deleted_count": 500,
  "space_freed": 5242880,
  "message": "Cleanup completed. Run VACUUM to fully reclaim disk space."
}
```

### GET /api/v1/storage/estimate

Estimate space freed by cleanup.

**Query Parameters:**
| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `before_date` | string | Yes | ISO 8601 timestamp |

**Response:**
```json
{
  "event_count": 500,
  "estimated_space": 5242880,
  "before_date": "2024-01-01T00:00:00Z"
}
```

### POST /api/v1/storage/vacuum

Run SQLite VACUUM on databases.

**Response:**
```json
{
  "success": true,
  "space_reclaimed": 5242880,
  "duration_ms": 1500
}
```

### GET /api/v1/storage/deletion-requests

Get NIP-09 deletion requests.

**Query Parameters:**
| Parameter | Type | Description |
|-----------|------|-------------|
| `status` | string | Filter by status |

**Response:**
```json
{
  "requests": [...],
  "total": 10
}
```

### POST /api/v1/storage/integrity-check

Run integrity check on databases.

**Response:**
```json
{
  "success": true,
  "app_db": {
    "ok": true,
    "result": "ok"
  },
  "relay_db": {
    "ok": true,
    "result": "ok"
  },
  "duration_ms": 500
}
```

---

## Sync

### POST /api/v1/sync/start

Start syncing events from public relays.

**Request Body:**
```json
{
  "pubkeys": ["hex1", "hex2"],
  "relays": ["wss://relay1.com"],
  "event_kinds": [1, 3, 6, 7],
  "since_timestamp": 1700000000
}
```

**Response (202 Accepted):**
```json
{
  "job_id": "abc123",
  "status": "running",
  "message": "Sync job started"
}
```

### GET /api/v1/sync/status

Get status of sync job.

**Query Parameters:**
| Parameter | Type | Description |
|-----------|------|-------------|
| `id` | int | Specific job ID (default: current running) |

**Response:**
```json
{
  "id": 1,
  "status": "running",
  "pubkeys": ["hex1", "hex2"],
  "relays": ["wss://relay1.com"],
  "events_synced": 1500,
  "started_at": "2025-12-22T14:00:00Z",
  "completed_at": null,
  "error": null
}
```

Status values: `running`, `completed`, `failed`, `cancelled`

### POST /api/v1/sync/cancel

Cancel the currently running sync job.

**Response:**
```json
{
  "success": true,
  "message": "Sync cancellation requested"
}
```

### GET /api/v1/sync/history

Get past sync jobs.

**Query Parameters:**
| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `limit` | int | `20` | Max 100 |
| `offset` | int | `0` | Pagination offset |

**Response:**
```json
{
  "jobs": [...],
  "limit": 20,
  "offset": 0
}
```

### GET /api/v1/sync/relays

Get default list of public relays for syncing.

**Response:**
```json
{
  "relays": [
    "wss://relay.damus.io",
    "wss://nos.lol",
    "wss://relay.nostr.band"
  ]
}
```

---

## Lightning

### GET /api/v1/lightning/status

Get Lightning node connection status.

**Response (connected):**
```json
{
  "configured": true,
  "enabled": true,
  "connected": true,
  "node_info": {
    "alias": "MyNode",
    "pubkey": "hex",
    "version": "0.17.0"
  },
  "balance": {
    "local": 1000000,
    "remote": 500000
  }
}
```

**Response (not connected):**
```json
{
  "configured": true,
  "enabled": true,
  "connected": false,
  "error": "Connection refused",
  "error_code": "CONNECTION_FAILED"
}
```

### PUT /api/v1/lightning/config

Save Lightning node configuration.

**Request Body:**
```json
{
  "host": "umbrel.local:8080",
  "macaroon_hex": "hex encoded macaroon",
  "tls_cert_path": "/path/to/cert",
  "enabled": true
}
```

**Response:**
```json
{
  "success": true,
  "message": "Lightning configuration saved"
}
```

### POST /api/v1/lightning/test

Test Lightning node connection.

**Request Body:** Same as PUT config.

**Response (success):**
```json
{
  "success": true,
  "node_info": {...},
  "message": "Connection successful"
}
```

**Response (failure):**
```json
{
  "success": false,
  "error": "Invalid macaroon",
  "error_code": "AUTH_FAILED",
  "message": "Connection failed"
}
```

---

## Public Signup

These endpoints are unauthenticated and used for the public signup flow.

### GET /public/relay-info

Get public relay info for signup page.

**Response:**
```json
{
  "paid_access_enabled": true,
  "lightning_configured": true,
  "name": "My Relay",
  "description": "A private Nostr relay",
  "tiers": [
    {
      "id": "monthly",
      "name": "Monthly",
      "amount_sats": 5000,
      "duration_days": 30
    }
  ]
}
```

### POST /public/create-invoice

Create Lightning invoice for signup.

**Request Body:**
```json
{
  "pubkey": "npub1... or hex",
  "tier_id": "monthly"
}
```

**Response (201 Created):**
```json
{
  "payment_hash": "hex",
  "payment_request": "lnbc...",
  "amount_sats": 5000,
  "tier_id": "monthly",
  "tier_name": "Monthly",
  "expires_at": "2025-12-22T15:00:00Z",
  "memo": "Relay access - Monthly"
}
```

### GET /public/invoice-status/{hash}

Check invoice payment status.

**URL Parameters:**
| Parameter | Description |
|-----------|-------------|
| `hash` | Payment hash from create-invoice |

**Response (pending):**
```json
{
  "status": "pending",
  "payment_hash": "hex",
  "expires_at": "2025-12-22T15:00:00Z"
}
```

**Response (paid):**
```json
{
  "status": "paid",
  "payment_hash": "hex",
  "paid_at": "2025-12-22T14:30:00Z",
  "message": "Access granted"
}
```

**Response (expired):**
```json
{
  "status": "expired",
  "payment_hash": "hex",
  "message": "Invoice expired"
}
```

---

## Support

### GET /api/v1/support/config

Get support and donation configuration.

**Response:**
```json
{
  "lightning_address": "ryand@getalby.com",
  "bitcoin_address": "[bitcoin-address]",
  "github_repo": "https://github.com/rdoiron/roostr",
  "developer_npub": "npub1...",
  "version": "0.1.0"
}
```

---

## Common Error Codes

| Code | Description |
|------|-------------|
| `INVALID_REQUEST` | Malformed request body |
| `INVALID_PUBKEY` | Invalid pubkey format |
| `NOT_FOUND` | Resource not found |
| `ALREADY_EXISTS` | Resource already exists |
| `UNAUTHORIZED` | Authentication required |
| `FORBIDDEN` | Permission denied |
| `INTERNAL_ERROR` | Server error |
| `RELAY_OFFLINE` | Relay is not running |
| `DATABASE_ERROR` | Database operation failed |
| `NIP05_FAILED` | NIP-05 resolution failed |
| `LIGHTNING_ERROR` | Lightning operation failed |

---

## HTTP Status Codes

| Code | Usage |
|------|-------|
| `200` | Success |
| `201` | Resource created |
| `202` | Request accepted (async operation) |
| `400` | Bad request |
| `404` | Not found |
| `409` | Conflict (duplicate) |
| `500` | Server error |
