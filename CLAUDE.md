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
| Backend | Go 1.22+ |
| Relay | nostr-rs-relay (Rust, external process) |
| Database | SQLite (relay DB read-only, app DB read-write) |
| Packaging | Docker Compose (Umbrel), .s9pk (Start9) |

### Go Dependencies

| Package | Purpose |
|---------|---------|
| `github.com/mattn/go-sqlite3` | SQLite database driver |
| `github.com/BurntSushi/toml` | TOML config file parsing for relay config.toml |
| `github.com/btcsuite/btcd/btcec/v2` | BIP-340 Schnorr signatures for Nostr event verification |

## Project Structure

```
roostr/
├── CLAUDE.md              # You're reading it
├── README.md              # User-facing docs
├── Makefile               # Build commands
├── docs/
│   ├── USER-GUIDE.md      # End-user documentation
│   ├── API.md             # Complete API reference
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

1. `docs/TASKS.md` - Task checklist with current progress
2. `docs/USER-GUIDE.md` - Feature documentation
3. `docs/API.md` - API endpoint reference
4. This file for conventions and structure

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

## Docker

**Image:** `rdoiron/roostr`
**Current version:** `0.1.0`

Always use the versioned tag that matches `platforms/umbrel/umbrel-app.yml`. Do NOT use or create a `:latest` tag.

```bash
# Build and push (from project root)
docker build -f platforms/umbrel/Dockerfile -t rdoiron/roostr:0.1.0 .
docker push rdoiron/roostr:0.1.0
```

When releasing a new version:
1. Update version in `platforms/umbrel/umbrel-app.yml`
2. Update version in `platforms/umbrel/docker-compose.yml`
3. Update the "Current version" in this section
4. Build and push with the new tag

## StartOS Packaging

**Target:** StartOS v0.3.5.x (stable) using s9pk v1 format

Key learnings:
- **Container must run as root** - StartOS mounts `/data` with root ownership. Do NOT use `USER` directive in Dockerfile.
- **SDK is hard to install locally** - Use Docker-based build via `make pack` in `platforms/startos/`
- **s9pk format requires SDK** - Cannot manually create tar archives; must use `start-sdk pack`

Build locally:
```bash
cd platforms/startos
make x86    # Build x86_64 image (~5 min)
make arm    # Build ARM64 image (~60 min via QEMU)
make pack   # Create roostr.s9pk using Docker-based SDK
```

For StartOS v0.4.x (alpha), convert v1 to v2:
```bash
start-cli s9pk convert roostr.s9pk
```

See `platforms/startos/README-BUILD.md` for full documentation.

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

**Relay Database (nostr-rs-relay)**
- Owned by nostr-rs-relay
- We read from it; write only for sync imports
- Main table: `event` with nostr-rs-relay schema:
  - `id` INTEGER PRIMARY KEY (auto-generated)
  - `event_hash` BLOB NOT NULL (32-byte event ID)
  - `first_seen` INTEGER NOT NULL (Unix timestamp when received)
  - `created_at` INTEGER (event creation timestamp)
  - `author` BLOB NOT NULL (32-byte pubkey)
  - `delegated_by` BLOB (optional)
  - `kind` INTEGER
  - `hidden` INTEGER
  - `content` TEXT NOT NULL (full serialized event JSON)
  - UNIQUE constraint on `event_hash`
- Note: nostr-rs-relay stores the complete event JSON in `content`, not just the content field

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
