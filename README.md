# 🐓 Roostr

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

## Installation

### Umbrel

Coming soon to the Umbrel App Store.

### Start9

Coming soon to the Start9 Marketplace.

### Manual / Development

```bash
# Clone the repo
git clone https://github.com/yourusername/roostr.git
cd roostr

# Start development servers
make dev
```

## Screenshots

*Coming soon*

## Tech Stack

- **Frontend**: Svelte 5 + SvelteKit + Tailwind CSS
- **Backend**: Go 1.22+ (with SQLite and TOML libraries)
- **Relay**: nostr-rs-relay
- **Database**: SQLite

## Development

See [CLAUDE.md](./CLAUDE.md) for development conventions and [docs/TASKS.md](./docs/TASKS.md) for the development roadmap.

```bash
# Run development servers
make dev

# Run tests
make test

# Build for production
make build
```

## Documentation

- [Full Specification](./docs/SPECIFICATION.md)
- [Development Tasks](./docs/TASKS.md)

## Support

If you find Roostr useful, consider supporting development:

⚡ Lightning: `[your-lightning-address]`  
₿ Bitcoin: `[your-bitcoin-address]`

## License

MIT License - see [LICENSE](./LICENSE) for details.

## Acknowledgments

- [nostr-rs-relay](https://github.com/scsibug/nostr-rs-relay) - The relay we wrap
- [Umbrel](https://umbrel.com/) - Home server platform
- [Start9](https://start9.com/) - Sovereign computing platform
- The Nostr community 💜
