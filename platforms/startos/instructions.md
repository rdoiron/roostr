# Roostr Instructions

Roostr is your private roost on Nostr - a web-based admin interface for managing your personal Nostr relay.

## Features

- **Dashboard**: View relay status, statistics, and recent activity
- **Access Control**: Manage who can read and write to your relay via whitelist
- **Event Browser**: Explore and search events stored on your relay
- **Sync**: Import your history from public relays
- **Paid Access**: Optionally accept Lightning payments for relay access
- **Storage Management**: Monitor disk usage and configure retention policies
- **Export/Backup**: Export your events for backup or migration

## Initial Setup

1. Open Roostr from your StartOS dashboard
2. Complete the setup wizard:
   - Enter your Nostr public key (npub) or NIP-05 identifier
   - Configure your relay name and description
   - Choose your access mode (private, whitelist, or public)
   - Optionally add other pubkeys to your whitelist
3. You'll receive your relay URLs (Tor and LAN) to use with Nostr clients

## Connecting Nostr Clients

After setup, configure your Nostr client to connect to your relay:

### Tor URL (Recommended for Privacy)
Use the `.onion` WebSocket URL provided in the dashboard. Most Nostr clients support Tor connections.

### LAN URL
For local network access, use the LAN URL shown in the dashboard. This works for clients on the same network as your Start9 server.

## Access Modes

- **Private**: Only you (the operator) can read and write
- **Whitelist**: Only approved pubkeys can access (you control the list)
- **Public**: Anyone can read; only whitelisted users can write

## Syncing Your History

To import your existing Nostr content:

1. Go to the Sync page from the dashboard
2. Select which pubkeys to sync (defaults to your whitelist)
3. Choose which public relays to fetch from
4. Select event types to import
5. Start the sync and monitor progress

## Lightning Integration (Optional)

If you want to accept payments for relay access:

1. Ensure you have a Lightning node running on StartOS (LND recommended)
2. Configure the Lightning connection in Roostr settings
3. Set your pricing tiers
4. Share your public signup page with potential users

## Backup and Restore

Roostr data is automatically included in StartOS backups. This includes:

- Your relay database (all stored events)
- Application settings
- Whitelist and access configuration
- Paid user records (if applicable)

To restore, use the standard StartOS restore function.

## Troubleshooting

### Relay not connecting
- Check the relay status in the dashboard
- View relay logs from the Settings page
- Try reloading or restarting the relay

### Events not syncing
- Ensure the source relays are reachable
- Check that the pubkeys are correct (hex format internally)
- Some relays may rate-limit requests; try fewer concurrent syncs

### Web interface not loading
- Clear your browser cache
- Try a different browser
- Check StartOS logs for errors

## Support

- **Issues**: https://github.com/roostr/roostr/issues
- **Nostr**: Follow project updates on Nostr
