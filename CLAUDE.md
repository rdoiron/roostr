# Roostr

> "Your Private Roost on Nostr"

Private Nostr relay management app for Umbrel and Start9 platforms.

## What This Project Is

Roostr is a web-based admin interface for managing a private Nostr relay. It wraps `nostr-rs-relay` and provides:

- Dashboard with relay status and statistics
- Whitelist/blacklist management for access control
- Event browser to explore stored content
- Sync from public relays to import history
- Paid relay access via Lightning Network
- Export/backup functionality
- Storage management with NIP-09 deletion support

Target platforms: Umbrel App Store and Start9 StartOS.

## Tech Stack

| Component | Technology |
|-----------|------------|
| Frontend | Svelte 5 + SvelteKit + Tailwind CSS |
| Backend | Go 1.21+ (standard library preferred) |
| Relay | nostr-rs-relay (Rust, external process) |
| Database | SQLite (relay DB read-only, app DB read-write) |
| Packaging | Docker Compose (Umbrel), .s9pk (Start9) |

## Project Structure

```
roostr/
├── CLAUDE.md              # You're reading it
├── README.md              # User-facing docs
├── Makefile               # Build commands
├── docs/
│   ├── SPECIFICATION.md   # Full product spec (READ THIS)
│   └── TASKS.md           # Development task checklist
├── app/
│   ├── api/               # Go backend
│   │   ├── cmd/           # Entry point
│   │   ├── internal/
│   │   │   ├── handlers/  # HTTP handlers
│   │   │   ├── services/  # Business logic
│   │   │   ├── db/        # Database access
│   │   │   ├── relay/     # Relay process control
│   │   │   └── config/    # Configuration management
│   │   └── go.mod
│   └── ui/                # Svelte frontend
│       ├── src/
│       │   ├── routes/    # SvelteKit pages
│       │   └── lib/       # Components, stores, utils
│       ├── static/
│       └── package.json
└── platforms/
    ├── umbrel/            # Umbrel packaging
    │   ├── docker-compose.yml
    │   └── umbrel-app.yml
    └── startos/           # Start9 packaging
        ├── Dockerfile
        ├── manifest.yaml
        └── scripts/
```

## Coding Conventions

### Go Backend

- Use standard library where possible (net/http, database/sql, encoding/json)
- Minimal dependencies: only add what's truly needed
- Handler files should be focused (<200 lines)
- Return structured JSON errors: `{"error": "message", "code": "ERROR_CODE"}`
- Use context for cancellation and timeouts
- Table-driven tests preferred

### Svelte Frontend

- Svelte 5 runes syntax ($state, $derived, $effect)
- Small, focused components (<150 lines)
- Tailwind for styling, no custom CSS unless necessary
- API calls through `/lib/api/` client module
- Stores for shared state in `/lib/stores/`

### API Design

- RESTful endpoints under `/api/v1/`
- Consistent response format
- Pagination: `?limit=50&offset=0`
- Filtering: query params (e.g., `?kind=1&author=npub...`)
- Use appropriate HTTP methods and status codes

### File Naming

- Go: `snake_case.go`
- Svelte: `PascalCase.svelte` for components, `+page.svelte` for routes
- General: lowercase with hyphens for directories

## Key Files to Read

When starting work, read these for context:

1. `docs/SPECIFICATION.md` - Full product specification
2. `docs/TASKS.md` - Task checklist with current progress
3. This file for conventions and structure

## Common Commands

```bash
# Development
make dev          # Run API and UI dev servers
make api          # Run just the API
make ui           # Run just the UI

# Building
make build        # Build everything
make build-api    # Build Go binary
make build-ui     # Build Svelte app

# Testing
make test         # Run all tests
make test-api     # Test Go code
make test-ui      # Test Svelte code

# Linting
make lint         # Lint everything

# Database
make db-reset     # Reset app database
make db-migrate   # Run migrations

# Packaging
make package-umbrel   # Build Umbrel package
make package-startos  # Build Start9 package
```

## Environment Variables

```bash
# API
PORT=3001                    # API server port (default: 3001)
RELAY_DB_PATH=/data/nostr.db # Path to relay's SQLite DB
APP_DB_PATH=/data/roostr.db  # Path to app's SQLite DB
CONFIG_PATH=/data/config.toml # Path to relay config
RELAY_BINARY=/usr/bin/nostr-rs-relay

# UI
PUBLIC_API_URL=http://localhost:3001/api/v1
```

## Database Notes

**Relay Database (read-only)**
- Owned by nostr-rs-relay
- We only SELECT from it
- Main table: `event` (id, pubkey, created_at, kind, tags, content, sig)

**App Database (read-write)**
- Owned by Roostr
- Stores: app_state, whitelist_meta, paid_users, deletion_requests, etc.

## Important Context

- The relay process (nostr-rs-relay) runs separately
- We control it via config.toml modifications and SIGHUP/restart
- NIP-42 authentication is handled by the relay, we configure it
- Tor URLs are provided by the platform (Umbrel/Start9), we detect them
- Lightning integration talks to user's own LND/CLN node

## Current Development Phase

Check `docs/TASKS.md` for current progress and next tasks.

## Git Commits

- Do not include Claude Code branding or co-author lines in commits
- Use descriptive commit messages following existing style

## Local Development Notes

- Port 8080 is unavailable on this machine; API defaults to port 3001
- Vite dev server proxies `/api` requests to `localhost:3001`
