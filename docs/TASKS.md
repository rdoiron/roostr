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
- [ ] UI-001: Create app layout component (nav, header)
- [ ] UI-002: Set up routing structure
- [ ] UI-003: Create API client module (/lib/api/)
- [ ] UI-004: Add loading and error state components
- [ ] UI-005: Set up Tailwind theme (colors, fonts)

---

## Phase 2: Setup Wizard

### Setup API
- [ ] SETUP-API-001: GET /api/v1/setup/status endpoint
- [ ] SETUP-API-002: POST /api/v1/setup/complete endpoint
- [ ] SETUP-API-003: Implement setup state persistence
- [ ] SETUP-API-004: Add operator pubkey validation
- [ ] SETUP-API-005: Add NIP-05 resolution for setup

### Setup UI
- [ ] SETUP-UI-001: Create setup wizard container/flow
- [ ] SETUP-UI-002: Step 1 - Welcome screen
- [ ] SETUP-UI-003: Step 2 - Identity (npub input)
- [ ] SETUP-UI-004: Step 3 - Relay name/description
- [ ] SETUP-UI-005: Step 4 - Access mode selection
- [ ] SETUP-UI-006: Step 5 - Add others (optional)
- [ ] SETUP-UI-007: Step 6 - Complete screen with URLs
- [ ] SETUP-UI-008: Add setup redirect middleware

---

## Phase 3: Dashboard

### Dashboard API
- [ ] DASH-API-001: GET /api/v1/stats/summary endpoint
- [ ] DASH-API-002: GET /api/v1/relay/urls endpoint
- [ ] DASH-API-003: GET /api/v1/relay/status endpoint
- [ ] DASH-API-004: GET /api/v1/events/recent endpoint (last 10)
- [ ] DASH-API-005: Implement Tor URL detection

### Dashboard UI
- [ ] DASH-UI-001: Create dashboard page layout
- [ ] DASH-UI-002: Relay status card (online/offline, uptime)
- [ ] DASH-UI-003: URL display with copy buttons
- [ ] DASH-UI-004: QR code generation for URLs
- [ ] DASH-UI-005: Stats cards (total events, storage, pubkeys)
- [ ] DASH-UI-006: Event type breakdown cards
- [ ] DASH-UI-007: Recent activity feed (clickable)
- [ ] DASH-UI-008: Quick action buttons
- [ ] DASH-UI-009: Auto-refresh for live stats

---

## Phase 4: Access Control

### Access Control API
- [ ] ACCESS-API-001: GET /api/v1/access/mode endpoint
- [ ] ACCESS-API-002: PUT /api/v1/access/mode endpoint
- [ ] ACCESS-API-003: GET /api/v1/access/whitelist endpoint
- [ ] ACCESS-API-004: POST /api/v1/access/whitelist endpoint
- [ ] ACCESS-API-005: DELETE /api/v1/access/whitelist/:pubkey endpoint
- [ ] ACCESS-API-006: Implement config.toml reader/writer
- [ ] ACCESS-API-007: Add relay config reload (SIGHUP)
- [ ] ACCESS-API-008: GET /api/v1/nip05/:identifier resolution
- [ ] ACCESS-API-009: Blacklist CRUD endpoints

### Access Control UI
- [ ] ACCESS-UI-001: Create access control page
- [ ] ACCESS-UI-002: Access mode selector (radio buttons)
- [ ] ACCESS-UI-003: Whitelist display with pubkey cards
- [ ] ACCESS-UI-004: Add pubkey modal (npub/NIP-05 input)
- [ ] ACCESS-UI-005: Pubkey validation and NIP-05 lookup
- [ ] ACCESS-UI-006: Edit nickname functionality
- [ ] ACCESS-UI-007: Remove pubkey with confirmation
- [ ] ACCESS-UI-008: Bulk import/export buttons
- [ ] ACCESS-UI-009: Event count per pubkey display

---

## Phase 5: Event Browser

### Event Browser API
- [ ] EVENTS-API-001: GET /api/v1/events endpoint (paginated)
- [ ] EVENTS-API-002: Add filtering (kind, author, date range)
- [ ] EVENTS-API-003: GET /api/v1/events/:id endpoint
- [ ] EVENTS-API-004: DELETE /api/v1/events/:id endpoint
- [ ] EVENTS-API-005: Add search by content (basic)

### Event Browser UI
- [ ] EVENTS-UI-001: Create event browser page
- [ ] EVENTS-UI-002: Filter controls (kind, author, date)
- [ ] EVENTS-UI-003: Event list with pagination
- [ ] EVENTS-UI-004: Event card component (kind-specific rendering)
- [ ] EVENTS-UI-005: Event detail modal (raw JSON view)
- [ ] EVENTS-UI-006: Delete event with confirmation
- [ ] EVENTS-UI-007: Deep link support (?id=xxx)
- [ ] EVENTS-UI-008: "Mentions me" filter

---

## Phase 6: Configuration

### Configuration API
- [ ] CONFIG-API-001: GET /api/v1/config endpoint
- [ ] CONFIG-API-002: PATCH /api/v1/config endpoint
- [ ] CONFIG-API-003: POST /api/v1/config/reload endpoint
- [ ] CONFIG-API-004: Validate config before saving

### Configuration UI
- [ ] CONFIG-UI-001: Create configuration page
- [ ] CONFIG-UI-002: Relay identity section (name, description, contact)
- [ ] CONFIG-UI-003: Limits section (rate limits, max sizes)
- [ ] CONFIG-UI-004: Event policies section (accepted kinds, PoW)
- [ ] CONFIG-UI-005: Save with validation feedback
- [ ] CONFIG-UI-006: Reset to defaults option

---

## Phase 7: Storage Management

### Storage API
- [ ] STORAGE-API-001: GET /api/v1/storage/status endpoint
- [ ] STORAGE-API-002: GET /api/v1/storage/retention endpoint
- [ ] STORAGE-API-003: PUT /api/v1/storage/retention endpoint
- [ ] STORAGE-API-004: POST /api/v1/storage/cleanup endpoint
- [ ] STORAGE-API-005: POST /api/v1/storage/vacuum endpoint
- [ ] STORAGE-API-006: GET /api/v1/storage/deletion-requests endpoint
- [ ] STORAGE-API-007: Implement NIP-09 deletion request processing
- [ ] STORAGE-API-008: Background retention job

### Storage UI
- [ ] STORAGE-UI-001: Add storage card to dashboard
- [ ] STORAGE-UI-002: Create storage management page
- [ ] STORAGE-UI-003: Usage visualization (progress bar)
- [ ] STORAGE-UI-004: Retention policy settings
- [ ] STORAGE-UI-005: Manual cleanup interface
- [ ] STORAGE-UI-006: NIP-09 deletion request list
- [ ] STORAGE-UI-007: Database maintenance buttons
- [ ] STORAGE-UI-008: Storage alerts/warnings

---

## Phase 8: Export & Backup

### Export API
- [ ] EXPORT-API-001: GET /api/v1/events/export endpoint (streaming)
- [ ] EXPORT-API-002: Add format parameter (ndjson, json)
- [ ] EXPORT-API-003: Add filters (kinds, date range)
- [ ] EXPORT-API-004: Add progress tracking for large exports

### Export UI
- [ ] EXPORT-UI-001: Create export page/modal
- [ ] EXPORT-UI-002: Event type selection
- [ ] EXPORT-UI-003: Date range picker
- [ ] EXPORT-UI-004: Format selection
- [ ] EXPORT-UI-005: Size estimation
- [ ] EXPORT-UI-006: Download progress indicator

---

## Phase 9: Sync from Public Relays

### Sync API
- [ ] SYNC-API-001: POST /api/v1/sync/start endpoint
- [ ] SYNC-API-002: GET /api/v1/sync/status endpoint
- [ ] SYNC-API-003: POST /api/v1/sync/cancel endpoint
- [ ] SYNC-API-004: Implement Nostr client for fetching
- [ ] SYNC-API-005: Background sync job with progress
- [ ] SYNC-API-006: Duplicate detection and skip
- [ ] SYNC-API-007: GET /api/v1/sync/history endpoint

### Sync UI
- [ ] SYNC-UI-001: Add sync button to dashboard
- [ ] SYNC-UI-002: Create sync configuration modal
- [ ] SYNC-UI-003: Pubkey selection (from whitelist)
- [ ] SYNC-UI-004: Relay selection (defaults + custom)
- [ ] SYNC-UI-005: Event type selection
- [ ] SYNC-UI-006: Sync progress display
- [ ] SYNC-UI-007: Sync complete summary
- [ ] SYNC-UI-008: Background sync indicator

---

## Phase 10: Support & Donations

### Support UI
- [ ] SUPPORT-UI-001: Create support page
- [ ] SUPPORT-UI-002: Lightning address display with QR
- [ ] SUPPORT-UI-003: Bitcoin address display with QR
- [ ] SUPPORT-UI-004: WebLN integration (tip button)
- [ ] SUPPORT-UI-005: Help links section
- [ ] SUPPORT-UI-006: About section with version

---

## Phase 11: Paid Relay Access

### Lightning Integration
- [ ] LN-001: LND connection manager
- [ ] LN-002: Auto-detect Umbrel LND
- [ ] LN-003: Invoice generation
- [ ] LN-004: Payment verification
- [ ] LN-005: Invoice subscription/polling

### Paid Access API
- [ ] PAID-API-001: GET /api/v1/access/pricing endpoint
- [ ] PAID-API-002: PUT /api/v1/access/pricing endpoint
- [ ] PAID-API-003: GET /api/v1/access/paid-users endpoint
- [ ] PAID-API-004: DELETE /api/v1/access/paid-users/:pubkey endpoint
- [ ] PAID-API-005: GET /api/v1/access/revenue endpoint
- [ ] PAID-API-006: GET /api/v1/lightning/status endpoint
- [ ] PAID-API-007: PUT /api/v1/lightning/config endpoint

### Public Signup API
- [ ] SIGNUP-API-001: GET /public/relay-info endpoint
- [ ] SIGNUP-API-002: POST /public/create-invoice endpoint
- [ ] SIGNUP-API-003: GET /public/invoice-status/:hash endpoint
- [ ] SIGNUP-API-004: Auto-whitelist on payment

### Paid Access UI
- [ ] PAID-UI-001: Pricing configuration section
- [ ] PAID-UI-002: Lightning node connection UI
- [ ] PAID-UI-003: Paid users list
- [ ] PAID-UI-004: Revenue summary
- [ ] PAID-UI-005: Public signup page
- [ ] PAID-UI-006: Plan selection cards
- [ ] PAID-UI-007: Invoice display with QR
- [ ] PAID-UI-008: Payment confirmation screen

### Subscription Management
- [ ] SUB-001: Expiry tracking
- [ ] SUB-002: Background expiry job
- [ ] SUB-003: Expiry warning display

---

## Phase 12: Relay Controls

### Relay Control API
- [ ] RELAY-API-001: GET /api/v1/relay/status (detailed)
- [ ] RELAY-API-002: POST /api/v1/relay/reload endpoint
- [ ] RELAY-API-003: POST /api/v1/relay/restart endpoint
- [ ] RELAY-API-004: GET /api/v1/relay/logs endpoint
- [ ] RELAY-API-005: WebSocket /api/v1/relay/logs/stream

### Relay Control UI
- [ ] RELAY-UI-001: Relay control section in settings
- [ ] RELAY-UI-002: Status display (PID, memory, uptime)
- [ ] RELAY-UI-003: Reload/restart buttons
- [ ] RELAY-UI-004: Log viewer
- [ ] RELAY-UI-005: Real-time log streaming

---

## Phase 13: Statistics & Charts

### Statistics API
- [ ] STATS-API-001: GET /api/v1/stats/events-over-time endpoint
- [ ] STATS-API-002: GET /api/v1/stats/events-by-kind endpoint
- [ ] STATS-API-003: GET /api/v1/stats/top-authors endpoint

### Statistics UI
- [ ] STATS-UI-001: Create statistics page
- [ ] STATS-UI-002: Time range selector
- [ ] STATS-UI-003: Events over time chart
- [ ] STATS-UI-004: Events by kind chart
- [ ] STATS-UI-005: Top authors list

---

## Phase 14: Platform Packaging

### Umbrel Package
- [ ] UMBREL-001: Create docker-compose.yml
- [ ] UMBREL-002: Create umbrel-app.yml manifest
- [ ] UMBREL-003: Create exports.sh
- [ ] UMBREL-004: Add app icon
- [ ] UMBREL-005: Test on Umbrel
- [ ] UMBREL-006: Submit to Umbrel App Store

### Start9 Package
- [ ] STARTOS-001: Create Dockerfile
- [ ] STARTOS-002: Create manifest.yaml
- [ ] STARTOS-003: Create instructions.md
- [ ] STARTOS-004: Create health check script
- [ ] STARTOS-005: Create config scripts
- [ ] STARTOS-006: Build .s9pk package
- [ ] STARTOS-007: Test on StartOS
- [ ] STARTOS-008: Submit to Start9 Marketplace

---

## Phase 15: Polish & Launch

### Testing
- [ ] TEST-001: API endpoint tests (Go)
- [ ] TEST-002: Database query tests
- [ ] TEST-003: UI component tests
- [ ] TEST-004: End-to-end tests
- [ ] TEST-005: Manual testing checklist

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
| 1. Foundation | 17 | 0 | 0% |
| 2. Setup Wizard | 13 | 0 | 0% |
| 3. Dashboard | 14 | 0 | 0% |
| 4. Access Control | 18 | 0 | 0% |
| 5. Event Browser | 13 | 0 | 0% |
| 6. Configuration | 10 | 0 | 0% |
| 7. Storage | 16 | 0 | 0% |
| 8. Export | 10 | 0 | 0% |
| 9. Sync | 15 | 0 | 0% |
| 10. Support | 6 | 0 | 0% |
| 11. Paid Access | 24 | 0 | 0% |
| 12. Relay Controls | 10 | 0 | 0% |
| 13. Statistics | 8 | 0 | 0% |
| 14. Packaging | 14 | 0 | 0% |
| 15. Polish | 14 | 0 | 0% |
| **TOTAL** | **202** | **0** | **0%** |

---

## Current Focus

**Next Task:** SETUP-001: Initialize git repo with .gitignore

See SPECIFICATION.md for full details on any feature.
