# Roostr Development Tasks

This document tracks all development tasks. Check off items as they're completed.

---

## Phase 1: Foundation

### Project Setup
- [x] SETUP-001: Initialize git repo with .gitignore
- [x] SETUP-002: Create Makefile with all targets
- [x] SETUP-003: Initialize Go module (app/api/go.mod)
- [x] SETUP-004: Initialize Svelte project (app/ui)
- [x] SETUP-005: Configure Tailwind CSS
- [x] SETUP-006: Create basic Docker Compose for development
- [x] SETUP-007: Set up folder structure with placeholder files

### Database Layer
- [x] DB-001: Create app database schema (app/api/internal/db/schema.sql)
- [x] DB-002: Implement database connection manager
- [x] DB-003: Implement relay DB reader (read-only queries)
- [x] DB-004: Implement app DB manager (read-write)
- [x] DB-005: Add database migration support

### Core API Infrastructure
- [x] API-001: Create HTTP server with router
- [x] API-002: Add middleware (logging, CORS, error handling)
- [x] API-003: Implement health check endpoint (GET /health)
- [x] API-004: Add structured error responses
- [x] API-005: Create response helper utilities

### Core UI Infrastructure
- [x] UI-001: Create app layout component (nav, header)
- [x] UI-002: Set up routing structure
- [x] UI-003: Create API client module (/lib/api/)
- [x] UI-004: Add loading and error state components
- [x] UI-005: Set up Tailwind theme (colors, fonts)

---

## Phase 2: Setup Wizard

### Setup API
- [x] SETUP-API-001: GET /api/v1/setup/status endpoint
- [x] SETUP-API-002: POST /api/v1/setup/complete endpoint
- [x] SETUP-API-003: Implement setup state persistence
- [x] SETUP-API-004: Add operator pubkey validation
- [x] SETUP-API-005: Add NIP-05 resolution for setup

### Setup UI
- [x] SETUP-UI-001: Create setup wizard container/flow
- [x] SETUP-UI-002: Step 1 - Welcome screen
- [x] SETUP-UI-003: Step 2 - Identity (npub input)
- [x] SETUP-UI-004: Step 3 - Relay name/description
- [x] SETUP-UI-005: Step 4 - Access mode selection
- [x] SETUP-UI-006: Step 5 - Add others (optional)
- [x] SETUP-UI-007: Step 6 - Complete screen with URLs
- [x] SETUP-UI-008: Add setup redirect middleware

---

## Phase 3: Dashboard

### Dashboard API
- [x] DASH-API-001: GET /api/v1/stats/summary endpoint
- [x] DASH-API-002: GET /api/v1/relay/urls endpoint
- [x] DASH-API-003: GET /api/v1/relay/status endpoint
- [x] DASH-API-004: GET /api/v1/events/recent endpoint (last 10)
- [x] DASH-API-005: Implement Tor URL detection

### Dashboard UI
- [x] DASH-UI-001: Create dashboard page layout
- [x] DASH-UI-002: Relay status card (online/offline, uptime)
- [x] DASH-UI-003: URL display with copy buttons
- [x] DASH-UI-004: QR code generation for URLs
- [x] DASH-UI-005: Stats cards (total events, storage, pubkeys)
- [x] DASH-UI-006: Event type breakdown cards
- [x] DASH-UI-007: Recent activity feed (clickable)
- [x] DASH-UI-008: Quick action buttons
- [x] DASH-UI-009: Auto-refresh for live stats

---

## Phase 4: Access Control

### Access Control API
- [x] ACCESS-API-001: GET /api/v1/access/mode endpoint
- [x] ACCESS-API-002: PUT /api/v1/access/mode endpoint
- [x] ACCESS-API-003: GET /api/v1/access/whitelist endpoint
- [x] ACCESS-API-004: POST /api/v1/access/whitelist endpoint
- [x] ACCESS-API-005: DELETE /api/v1/access/whitelist/:pubkey endpoint
- [x] ACCESS-API-006: Implement config.toml reader/writer
- [x] ACCESS-API-007: Add relay config reload (SIGHUP)
- [x] ACCESS-API-008: GET /api/v1/nip05/:identifier resolution
- [x] ACCESS-API-009: Blacklist CRUD endpoints

### Access Control UI
- [x] ACCESS-UI-001: Create access control page
- [x] ACCESS-UI-002: Access mode selector (radio buttons)
- [x] ACCESS-UI-003: Whitelist display with pubkey cards
- [x] ACCESS-UI-004: Add pubkey modal (npub/NIP-05 input)
- [x] ACCESS-UI-005: Pubkey validation and NIP-05 lookup
- [x] ACCESS-UI-006: Edit nickname functionality
- [x] ACCESS-UI-007: Remove pubkey with confirmation
- [x] ACCESS-UI-008: Bulk import/export buttons
- [x] ACCESS-UI-009: Event count per pubkey display

---

## Phase 5: Event Browser

### Event Browser API
- [x] EVENTS-API-001: GET /api/v1/events endpoint (paginated)
- [x] EVENTS-API-002: Add filtering (kind, author, date range)
- [x] EVENTS-API-003: GET /api/v1/events/:id endpoint
- [x] EVENTS-API-004: DELETE /api/v1/events/:id endpoint
- [x] EVENTS-API-005: Add search by content (basic)

### Event Browser UI
- [x] EVENTS-UI-001: Create event browser page
- [x] EVENTS-UI-002: Filter controls (kind, author, date)
- [x] EVENTS-UI-003: Event list with pagination
- [x] EVENTS-UI-004: Event card component (kind-specific rendering)
- [x] EVENTS-UI-005: Event detail modal (raw JSON view)
- [x] EVENTS-UI-006: Delete event with confirmation
- [x] EVENTS-UI-007: Deep link support (?id=xxx)
- [x] EVENTS-UI-008: "Mentions me" filter

---

## Phase 6: Configuration

### Configuration API
- [x] CONFIG-API-001: GET /api/v1/config endpoint
- [x] CONFIG-API-002: PATCH /api/v1/config endpoint
- [x] CONFIG-API-003: POST /api/v1/config/reload endpoint
- [x] CONFIG-API-004: Validate config before saving

### Configuration UI
- [x] CONFIG-UI-001: Create configuration page
- [x] CONFIG-UI-002: Relay identity section (name, description, contact)
- [x] CONFIG-UI-003: Limits section (rate limits, max sizes)
- [x] CONFIG-UI-004: Event policies section (accepted kinds, PoW)
- [x] CONFIG-UI-005: Save with validation feedback
- [x] CONFIG-UI-006: Reset to defaults option

---

## Phase 7: Storage Management

### Storage API
- [x] STORAGE-API-001: GET /api/v1/storage/status endpoint
- [x] STORAGE-API-002: GET /api/v1/storage/retention endpoint
- [x] STORAGE-API-003: PUT /api/v1/storage/retention endpoint
- [x] STORAGE-API-004: POST /api/v1/storage/cleanup endpoint
- [x] STORAGE-API-005: POST /api/v1/storage/vacuum endpoint
- [x] STORAGE-API-006: GET /api/v1/storage/deletion-requests endpoint
- [x] STORAGE-API-007: Implement NIP-09 deletion request processing
- [x] STORAGE-API-008: Background retention job

### Storage UI
- [x] STORAGE-UI-001: Add storage card to dashboard
- [x] STORAGE-UI-002: Create storage management page
- [x] STORAGE-UI-003: Usage visualization (progress bar)
- [x] STORAGE-UI-004: Retention policy settings
- [x] STORAGE-UI-005: Manual cleanup interface
- [x] STORAGE-UI-006: NIP-09 deletion request list
- [x] STORAGE-UI-007: Database maintenance buttons
- [x] STORAGE-UI-008: Storage alerts/warnings

---

## Phase 8: Export & Backup

### Export API
- [x] EXPORT-API-001: GET /api/v1/events/export endpoint (streaming)
- [x] EXPORT-API-002: Add format parameter (ndjson, json)
- [x] EXPORT-API-003: Add filters (kinds, date range)
- [x] EXPORT-API-004: Add progress tracking for large exports

### Export UI
- [x] EXPORT-UI-001: Create export page/modal
- [x] EXPORT-UI-002: Event type selection
- [x] EXPORT-UI-003: Date range picker
- [x] EXPORT-UI-004: Format selection
- [x] EXPORT-UI-005: Size estimation
- [x] EXPORT-UI-006: Download progress indicator

---

## Phase 9: Sync from Public Relays

### Sync API
- [x] SYNC-API-001: POST /api/v1/sync/start endpoint
- [x] SYNC-API-002: GET /api/v1/sync/status endpoint
- [x] SYNC-API-003: POST /api/v1/sync/cancel endpoint
- [x] SYNC-API-004: Implement Nostr client for fetching
- [x] SYNC-API-005: Background sync job with progress
- [x] SYNC-API-006: Duplicate detection and skip
- [x] SYNC-API-007: GET /api/v1/sync/history endpoint

### Sync UI
- [x] SYNC-UI-001: Add sync button to dashboard
- [x] SYNC-UI-002: Create sync configuration modal
- [x] SYNC-UI-003: Pubkey selection (from whitelist)
- [x] SYNC-UI-004: Relay selection (defaults + custom)
- [x] SYNC-UI-005: Event type selection
- [x] SYNC-UI-006: Sync progress display
- [x] SYNC-UI-007: Sync complete summary
- [x] SYNC-UI-008: Background sync indicator

---

## Phase 10: Support & Donations

### Support UI
- [x] SUPPORT-UI-001: Create support page
- [x] SUPPORT-UI-002: Lightning address display with QR
- [x] SUPPORT-UI-003: Bitcoin address display with QR
- [x] SUPPORT-UI-004: WebLN integration (tip button)
- [x] SUPPORT-UI-005: Help links section
- [x] SUPPORT-UI-006: About section with version

---

## Phase 11: Paid Relay Access

### Lightning Integration
- [x] LN-001: LND connection manager
- [x] LN-002: Auto-detect Umbrel LND
- [x] LN-003: Invoice generation
- [x] LN-004: Payment verification
- [x] LN-005: Invoice subscription/polling

### Paid Access API
- [x] PAID-API-001: GET /api/v1/access/pricing endpoint
- [x] PAID-API-002: PUT /api/v1/access/pricing endpoint
- [x] PAID-API-003: GET /api/v1/access/paid-users endpoint
- [x] PAID-API-004: DELETE /api/v1/access/paid-users/:pubkey endpoint
- [x] PAID-API-005: GET /api/v1/access/revenue endpoint
- [x] PAID-API-006: GET /api/v1/lightning/status endpoint
- [x] PAID-API-007: PUT /api/v1/lightning/config endpoint

### Public Signup API
- [x] SIGNUP-API-001: GET /public/relay-info endpoint
- [x] SIGNUP-API-002: POST /public/create-invoice endpoint
- [x] SIGNUP-API-003: GET /public/invoice-status/:hash endpoint
- [x] SIGNUP-API-004: Auto-whitelist on payment

### Paid Access UI
- [x] PAID-UI-001: Pricing configuration section
- [x] PAID-UI-002: Lightning node connection UI
- [x] PAID-UI-003: Paid users list
- [x] PAID-UI-004: Revenue summary
- [x] PAID-UI-005: Public signup page
- [x] PAID-UI-006: Plan selection cards
- [x] PAID-UI-007: Invoice display with QR
- [x] PAID-UI-008: Payment confirmation screen

### Subscription Management
- [x] SUB-001: Expiry tracking
- [x] SUB-002: Background expiry job
- [x] SUB-003: Expiry warning display

---

## Phase 12: Relay Controls

### Relay Control API
- [x] RELAY-API-001: GET /api/v1/relay/status (detailed)
- [x] RELAY-API-002: POST /api/v1/relay/reload endpoint
- [x] RELAY-API-003: POST /api/v1/relay/restart endpoint
- [x] RELAY-API-004: GET /api/v1/relay/logs endpoint
- [x] RELAY-API-005: SSE /api/v1/relay/logs/stream (uses SSE, no external deps)

### Relay Control UI
- [x] RELAY-UI-001: Relay control section in settings
- [x] RELAY-UI-002: Status display (PID, memory, uptime)
- [x] RELAY-UI-003: Reload/restart buttons
- [x] RELAY-UI-004: Log viewer
- [x] RELAY-UI-005: Real-time log streaming

---

## Phase 13: Statistics & Charts

### Statistics API
- [x] STATS-API-001: GET /api/v1/stats/events-over-time endpoint
- [x] STATS-API-002: GET /api/v1/stats/events-by-kind endpoint
- [x] STATS-API-003: GET /api/v1/stats/top-authors endpoint

### Statistics UI
- [x] STATS-UI-001: Create statistics page
- [x] STATS-UI-002: Time range selector
- [x] STATS-UI-003: Events over time chart
- [x] STATS-UI-004: Events by kind chart
- [x] STATS-UI-005: Top authors list

---

## Phase 14: Platform Packaging

### Umbrel Package
- [x] UMBREL-001: Create docker-compose.yml
- [x] UMBREL-002: Create umbrel-app.yml manifest
- [x] UMBREL-003: Create exports.sh
- [x] UMBREL-004: Add app icon
- [x] UMBREL-005: Test on Umbrel
- [ ] UMBREL-006: Submit to Umbrel App Store

### Start9 Package
- [x] STARTOS-001: Create Dockerfile
- [x] STARTOS-002: Create manifest.yaml
- [x] STARTOS-003: Create instructions.md
- [x] STARTOS-004: Create health check script
- [x] STARTOS-005: Create config scripts
- [x] STARTOS-006: Build .s9pk package
- [ ] STARTOS-007: Test on StartOS
- [ ] STARTOS-008: Submit to Start9 Marketplace

---

## Phase 15: Polish & Launch

### Testing
- [x] TEST-001: API endpoint tests (Go)
- [x] TEST-002: Database query tests
- [x] TEST-003: UI component tests
- [ ] TEST-004: End-to-end tests
- [x] TEST-005: Manual testing checklist

### Documentation
- [ ] DOCS-001: README with installation
- [ ] DOCS-002: User guide
- [ ] DOCS-003: API documentation
- [ ] DOCS-004: Contributing guide
- [ ] DOCS-005: Screenshots for app stores

### Launch
- [ ] LAUNCH-001: Create GitHub releases
- [ ] LAUNCH-002: Announce on Nostr
- [ ] LAUNCH-003: Create demo video
- [ ] LAUNCH-004: Gather initial feedback

---

## Progress Summary

| Phase | Tasks | Completed | Progress |
|-------|-------|-----------|----------|
| 1. Foundation | 17 | 17 | 100% |
| 2. Setup Wizard | 13 | 13 | 100% |
| 3. Dashboard | 14 | 14 | 100% |
| 4. Access Control | 18 | 18 | 100% |
| 5. Event Browser | 13 | 13 | 100% |
| 6. Configuration | 10 | 10 | 100% |
| 7. Storage | 16 | 16 | 100% |
| 8. Export | 10 | 10 | 100% |
| 9. Sync | 15 | 15 | 100% |
| 10. Support | 6 | 6 | 100% |
| 11. Paid Access | 27 | 27 | 100% |
| 12. Relay Controls | 10 | 10 | 100% |
| 13. Statistics | 8 | 8 | 100% |
| 14. Packaging | 14 | 11 | 79% |
| 15. Polish | 14 | 4 | 29% |
| **TOTAL** | **205** | **192** | **94%** |

---

## Current Focus

**Completed:** Full test coverage with 605 tests total.

- Go backend: 478 tests (handlers, services, database layer)
- UI: 127 tests (utilities, API client, Svelte components)

**Next Tasks:**

- TEST-004: End-to-end tests (Playwright)
- STARTOS-007: Test on StartOS
- DOCS-001 through DOCS-005: Documentation

See SPECIFICATION.md for full details on any feature.
