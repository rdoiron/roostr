# Roostr

> Your Private Roost on Nostr

Roostr is a comprehensive web-based management interface for private Nostr relays, designed for self-hosters running [Umbrel](https://umbrel.com/) or [Start9](https://start9.com/).

## Features

- **Dashboard** - Real-time relay status, statistics, and activity feed
- **Access Control** - Whitelist/blacklist management with NIP-05 support
- **Event Browser** - Search, filter, and inspect stored events
- **Sync** - Import your history from public relays
- **Paid Access** - Monetize your relay with Lightning payments
- **Storage Management** - Retention policies, cleanup, NIP-09 deletion support
- **Export/Backup** - Download your events in standard formats

## Quick Start

```bash
# Clone the repository
git clone https://github.com/rdoiron/roostr.git
cd roostr

# Install dependencies
make deps

# Start development servers
make dev
```

Open `http://localhost:5173` and complete the setup wizard.

## Installation

### Umbrel

Coming soon to the Umbrel App Store. For manual installation, see Docker instructions below.

### Start9

Coming soon to the Start9 Marketplace.

### Docker

```bash
# Pull the image
docker pull rdoiron/roostr:0.1.0

# Run with Docker Compose (recommended)
cd platforms/umbrel
docker-compose up -d
```

### Manual / Development

#### Prerequisites

| Requirement | Version | Notes |
|-------------|---------|-------|
| Go | 1.22+ | [golang.org/dl](https://golang.org/dl/) |
| Node.js | 20+ | [nodejs.org](https://nodejs.org/) |
| SQLite3 headers | - | Required for CGO |

**Install SQLite3 headers:**
```bash
# Ubuntu/Debian
sudo apt install libsqlite3-dev

# macOS
brew install sqlite3

# Fedora
sudo dnf install sqlite-devel
```

**Setup:**
```bash
git clone https://github.com/rdoiron/roostr.git
cd roostr
make deps
make dev
```

## Configuration

Key environment variables:

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `3001` | API server port |
| `RELAY_DB_PATH` | `/data/nostr.db` | Path to relay SQLite database |
| `APP_DB_PATH` | `/data/roostr.db` | Path to app SQLite database |
| `CONFIG_PATH` | `/data/config.toml` | Path to relay config file |
| `RELAY_BINARY` | `/usr/bin/nostr-rs-relay` | Path to relay binary |

See [CLAUDE.md](./CLAUDE.md) for the complete configuration reference.

## Screenshots

*Coming soon*

## Tech Stack

| Component | Technology |
|-----------|------------|
| Frontend | Svelte 5 + SvelteKit + Tailwind CSS |
| Backend | Go 1.22+ |
| Relay | nostr-rs-relay |
| Database | SQLite |

## Development

```bash
make dev          # Run API and UI dev servers
make test         # Run all tests
make build        # Build for production
make lint         # Lint all code
```

See [CONTRIBUTING.md](./CONTRIBUTING.md) for the full development guide.

## API

The API is available at `/api/v1/`. Key endpoint categories:

- **Setup** - Initial configuration wizard
- **Stats** - Dashboard statistics and real-time updates
- **Access** - Whitelist, blacklist, and paid access management
- **Events** - Browse, search, and manage stored events
- **Sync** - Import events from public relays
- **Storage** - Retention policies and cleanup
- **Config** - Relay configuration

See [docs/API.md](./docs/API.md) for the complete API reference.

## Documentation

- [User Guide](./docs/USER-GUIDE.md) - End-user documentation
- [API Reference](./docs/API.md) - Complete API documentation
- [Contributing](./CONTRIBUTING.md) - Development setup and guidelines
- [Development Tasks](./docs/TASKS.md) - Task checklist and roadmap

## Support

If you find Roostr useful, consider supporting development:

Lightning: `ryand@getalby.com`
Bitcoin: `[bitcoin-address]`

## License

MIT License - see [LICENSE](./LICENSE) for details.

## Acknowledgments

- [nostr-rs-relay](https://github.com/scsibug/nostr-rs-relay) - The relay we wrap
- [Umbrel](https://umbrel.com/) - Home server platform
- [Start9](https://start9.com/) - Sovereign computing platform
- The Nostr community
