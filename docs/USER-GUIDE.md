# Roostr User Guide

This guide covers everything you need to know to use Roostr effectively.

## Table of Contents

1. [Getting Started](#getting-started)
2. [Dashboard](#dashboard)
3. [Access Control](#access-control)
4. [Event Browser](#event-browser)
5. [Sync from Public Relays](#sync-from-public-relays)
6. [Paid Relay Access](#paid-relay-access)
7. [Storage Management](#storage-management)
8. [Export & Backup](#export--backup)
9. [Configuration](#configuration)
10. [Troubleshooting](#troubleshooting)

---

## Getting Started

### First-Run Setup Wizard

When you first access Roostr, you'll be guided through a setup wizard:

**Step 1: Welcome**
Introduction to your private Nostr relay.

**Step 2: Set Your Identity**
Enter your Nostr identity in one of these formats:
- **npub**: Your public key in bech32 format (e.g., `npub1abc...xyz`)
- **Hex**: Your 64-character hex public key
- **NIP-05**: Your verified identifier (e.g., `you@example.com`)

This identity becomes the relay operator and is always whitelisted.

**Step 3: Name Your Relay**
Give your relay a name and description. These appear in Nostr clients that support relay metadata.

**Step 4: Choose Access Mode**
- **Private**: Only whitelisted pubkeys can connect (most restrictive)
- **Paid**: Users pay via Lightning to gain access
- **Public**: Anyone can connect (least restrictive)

**Step 5: Add Others (Optional)**
Invite friends or family by adding their pubkeys to the whitelist.

**Step 6: Complete**
View your relay's connection URLs (local and Tor if available).

---

## Dashboard

The dashboard provides an at-a-glance view of your relay's status.

### Relay Status

Shows whether your relay is **online** or **offline**, along with:
- Process ID (PID)
- Memory usage
- Uptime

### Connection URLs

Your relay has two WebSocket URLs for connecting Nostr clients:

- **Local URL**: `ws://your-server-ip:7000` - Use on your local network
- **Tor URL**: `ws://xxxx.onion` - Access from anywhere (if Tor is enabled)

Click the copy button to copy URLs, or scan the QR code with a mobile client.

### Statistics

- **Total Events**: Number of events stored in the relay database
- **Events Today**: Events received in the last 24 hours
- **Storage Used**: Database size on disk
- **Whitelisted Users**: Number of pubkeys in your whitelist

### Event Type Breakdown

Visual breakdown of stored events by type:
- Posts (kind 1)
- Follows (kind 3)
- Reposts (kind 6)
- Reactions (kind 7)
- DMs (kind 4)
- Other kinds

### Recent Activity

Live feed of the most recent events received by your relay. Click any event to view details.

### Quick Actions

- **Sync**: Import events from public relays
- **Export**: Download your events as backup
- **Reload Config**: Apply configuration changes without restart

---

## Access Control

Control who can read from and write to your relay.

### Access Modes

| Mode | Who Can Connect | Best For |
|------|-----------------|----------|
| **Private** | Only whitelisted pubkeys | Personal/family use |
| **Paid** | Whitelisted + paid subscribers | Monetizing your relay |
| **Public** | Anyone | Community relays |

### Whitelist

The whitelist contains pubkeys that can always connect, regardless of access mode.

**Adding Users:**
1. Click "Add User"
2. Enter their npub, hex pubkey, or NIP-05 identifier
3. Optionally add a nickname for easy reference
4. Click "Add"

**NIP-05 Lookup:**
Enter an identifier like `alice@example.com` and Roostr will resolve it to the correct pubkey.

**Bulk Import/Export:**
- Export your whitelist as JSON for backup
- Import pubkeys from a JSON file

**Managing Users:**
- Edit nicknames by clicking the edit icon
- Remove users by clicking the delete icon (operator cannot be removed)
- View event count for each user

### Blacklist

Block specific pubkeys from connecting (only applies in Public mode).

---

## Event Browser

Browse, search, and manage events stored in your relay.

### Filtering Events

Use the filter controls to narrow down events:

- **Event Type**: Filter by kind (posts, follows, DMs, etc.)
- **Author**: Filter by specific pubkey
- **Date Range**: Show events from a specific time period
- **Content Search**: Search within event content
- **Mentions Me**: Show only events that mention your pubkey

### Event Details

Click any event to view:
- Full event metadata (ID, pubkey, timestamp, kind)
- Content (rendered appropriately for the event type)
- Raw JSON view for technical details

### Deleting Events

Roostr supports NIP-09 deletion requests:

1. Click the delete icon on an event
2. Optionally provide a reason
3. Confirm deletion

The event will be queued for deletion and removed during the next cleanup cycle.

---

## Sync from Public Relays

Import your existing Nostr history from public relays.

### How Sync Works

1. Roostr connects to public relays as a client
2. Requests events from the pubkeys you specify
3. Verifies event signatures
4. Imports valid events into your relay (skipping duplicates)

### Starting a Sync

1. Go to Dashboard and click "Sync" (or navigate to the Sync page)
2. **Select Pubkeys**: Choose from your whitelist or enter pubkeys manually
3. **Select Relays**: Use the default relay list or add custom relays
4. **Choose Event Types**: Select which kinds to import
5. **Set Date Range** (optional): Only sync events after a certain date
6. Click "Start Sync"

### Monitoring Progress

The sync status shows:
- Events found
- Events imported (new)
- Events skipped (duplicates)
- Current relay being queried
- Estimated time remaining

### Sync History

View past sync jobs and their results.

---

## Paid Relay Access

Monetize your relay by requiring Lightning payments for access.

### Prerequisites

- A Lightning node (LND) accessible from your Umbrel/Start9 server
- Admin macaroon credentials

### Setting Up Lightning

1. Go to Configuration > Lightning
2. Enter your LND connection details:
   - **Host**: Your LND node address (e.g., `umbrel.local:8080`)
   - **Macaroon**: Admin macaroon in hex format
3. Test the connection
4. Save configuration

### Configuring Pricing

1. Go to Access Control > Pricing
2. Enable paid access
3. Configure pricing tiers:
   - **Name**: Display name (e.g., "Monthly", "Yearly", "Lifetime")
   - **Amount**: Price in satoshis
   - **Duration**: Access duration in days (leave empty for lifetime)
4. Save changes

### Public Signup Page

When paid access is enabled, a public signup page is available at `/signup`:
- Shows your relay name and description
- Displays available pricing tiers
- Users enter their npub and select a plan
- Lightning invoice is generated
- Payment is verified automatically
- User is added to whitelist upon payment

### Managing Paid Users

View and manage paid subscribers:
- See subscription status (active, expired, revoked)
- View expiration dates
- Revoke access if needed
- Export subscriber list

### Revenue Dashboard

Track your relay's earnings:
- Total revenue
- Active subscribers
- Revenue by tier
- Upcoming expirations

---

## Storage Management

Monitor and manage your relay's storage usage.

### Storage Status

View current storage metrics:
- Database size
- Available disk space
- Total events stored
- Oldest and newest events

### Retention Policies

Configure automatic event cleanup:

1. **Retention Period**: Delete events older than X days
2. **Exceptions**: Keep events from specific pubkeys regardless of age
3. **Honor NIP-09**: Automatically process deletion requests

### Manual Cleanup

Delete events before a specific date:

1. Go to Storage > Cleanup
2. Select a date
3. Preview events that will be deleted
4. Confirm cleanup

### Database Maintenance

**Vacuum**: Reclaim disk space after deleting events. SQLite doesn't automatically free disk space; VACUUM rebuilds the database.

**Integrity Check**: Verify database integrity and check for corruption.

---

## Export & Backup

Create backups of your relay's events.

### Export Formats

- **NDJSON** (recommended): One JSON event per line, ideal for large exports
- **JSON**: Single JSON array, easier to work with for small datasets

### Exporting Events

1. Go to Export
2. Select format (NDJSON or JSON)
3. Filter events (optional):
   - By event type
   - By date range
4. View size estimate
5. Click "Download"

### Backup Recommendations

- Export regularly (weekly or monthly)
- Store exports in multiple locations
- Consider encrypting sensitive backups (especially DMs)
- Test restoring from backups periodically

---

## Configuration

### Relay Settings

**Identity:**
- Relay name
- Description
- Contact (email or nostr pubkey)
- Icon URL

**Limits:**
- Maximum event size (bytes)
- Maximum WebSocket message size
- Rate limit (messages per second)
- Maximum subscriptions per connection

**Policies:**
- Proof of Work minimum difficulty
- Event kind allowlist (restrict which kinds are accepted)
- NIP-42 authentication (require AUTH for connections)

### Applying Changes

After changing configuration:
1. Click "Save"
2. Click "Reload Config" to apply without restarting

Some changes may require a full relay restart.

---

## Troubleshooting

### Relay Won't Start

1. Check relay logs (Dashboard > Logs or Relay Control page)
2. Verify database files exist and have correct permissions
3. Check available disk space
4. Ensure no other process is using port 7000

### Can't Connect from Nostr Client

1. Verify the relay is running (check Dashboard status)
2. Use the correct URL (local vs Tor)
3. If using Tor, ensure your client supports .onion addresses
4. Check if your pubkey is whitelisted (in Private/Paid modes)

### Sync Not Finding Events

1. Verify the pubkey is correct
2. Try different source relays
3. Check the date range filter
4. Some relays may rate-limit requests

### Lightning Payments Not Working

1. Test the Lightning connection in Configuration
2. Verify macaroon has correct permissions
3. Check LND node is online and synced
4. Ensure your node has inbound liquidity

### High Storage Usage

1. Check for large event types (kind 1063 file metadata, etc.)
2. Configure retention policy
3. Run manual cleanup for old events
4. Run VACUUM after cleanup to reclaim space

### Getting Help

- Check the [GitHub Issues](https://github.com/rdoiron/roostr/issues) for known issues
- Open a new issue with details about your problem
- Include relevant logs when reporting bugs
