/**
 * Mock data objects for E2E tests.
 * These mirror the API response shapes.
 */

// Setup status responses
export const setupStatus = {
	complete: {
		completed: true,
		operator_pubkey: 'abc123hexkey',
		relay_name: 'Test Relay',
		relay_description: 'A test relay'
	},
	incomplete: {
		completed: false
	}
};

// Relay URLs
export const relayUrls = {
	relay_port: '7000',
	relay_url: 'ws://localhost:7000',
	tor_available: true,
	tor: 'ws://abcdef1234567890.onion'
};

// Stats summary
export const statsSummary = {
	total_events: 12500,
	events_today: 150,
	whitelisted_count: 25,
	blacklisted_count: 3,
	relay_status: 'online',
	uptime_seconds: 86400,
	events_by_kind: {
		posts: 8000,
		reactions: 3000,
		dms: 500,
		reposts: 400,
		follows: 300,
		other: 300
	}
};

// Storage status
export const storageStatus = {
	total_size: 524288000,
	available_space: 10737418240,
	database_size: 500000000,
	app_database_size: 24288000,
	total_events: 12500,
	status: 'healthy',
	oldest_event: '2024-01-01T00:00:00Z',
	newest_event: '2024-12-20T00:00:00Z'
};

// Retention settings
export const retention = {
	enabled: true,
	max_age_days: 90,
	max_events: 100000,
	excluded_kinds: [0, 3, 10002]
};

// Whitelist entries
export const whitelist = {
	entries: [
		{
			pubkey: 'abc123hexkey',
			npub: 'npub1abc123...',
			nickname: 'Alice',
			event_count: 150,
			added_at: '2024-01-01T00:00:00Z'
		},
		{
			pubkey: 'def456hexkey',
			npub: 'npub1def456...',
			nickname: '',
			event_count: 75,
			added_at: '2024-02-01T00:00:00Z'
		},
		{
			pubkey: 'ghi789hexkey',
			npub: 'npub1ghi789...',
			nickname: 'Bob',
			event_count: 300,
			added_at: '2024-03-01T00:00:00Z'
		}
	]
};

// Blacklist entries
export const blacklist = {
	entries: [
		{
			pubkey: 'bad123hexkey',
			npub: 'npub1bad123...',
			reason: 'spam',
			added_at: '2024-06-01T00:00:00Z'
		}
	]
};

// Access mode
export const accessMode = {
	whitelist: { mode: 'whitelist' },
	blacklist: { mode: 'blacklist' },
	paid: { mode: 'paid' },
	open: { mode: 'open' }
};

// Events list
export const eventsList = {
	events: [
		{
			id: 'event001',
			kind: 1,
			pubkey: 'abc123hexkey',
			content: 'Hello world! This is a test post.',
			created_at: 1703001600,
			sig: 'sig123',
			tags: []
		},
		{
			id: 'event002',
			kind: 7,
			pubkey: 'def456hexkey',
			content: '+',
			created_at: 1703001500,
			sig: 'sig456',
			tags: [['e', 'event001']]
		},
		{
			id: 'event003',
			kind: 1,
			pubkey: 'ghi789hexkey',
			content: 'Another post with more content.',
			created_at: 1703001400,
			sig: 'sig789',
			tags: []
		},
		{
			id: 'event004',
			kind: 6,
			pubkey: 'abc123hexkey',
			content: '',
			created_at: 1703001300,
			sig: 'sigabc',
			tags: [['e', 'event003']]
		},
		{
			id: 'event005',
			kind: 4,
			pubkey: 'def456hexkey',
			content: 'encrypted content here',
			created_at: 1703001200,
			sig: 'sigdef',
			tags: [['p', 'abc123hexkey']]
		}
	],
	total: 12500,
	has_more: true
};

// Single event detail
export const eventDetail = {
	id: 'event001',
	kind: 1,
	pubkey: 'abc123hexkey',
	content: 'Hello world! This is a test post.',
	created_at: 1703001600,
	sig: 'sig123',
	tags: []
};

// Recent events (for dashboard)
export const recentEvents = eventsList.events.slice(0, 5);

// Relay status
export const relayStatus = {
	status: 'running',
	pid: 12345,
	memory_bytes: 52428800,
	uptime_seconds: 86400
};

// Relay config - matches expected API response structure
export const relayConfig = {
	info: {
		name: 'Test Relay',
		description: 'A private relay for testing',
		pubkey: 'abc123hexkey',
		contact: 'admin@example.com',
		relay_icon: ''
	},
	limits: {
		max_event_bytes: 65536,
		max_ws_message_bytes: 131072,
		messages_per_sec: 10,
		max_subs_per_conn: 10,
		min_pow_difficulty: 0
	},
	authorization: {
		nip42_auth: true,
		event_kind_allowlist: []
	}
};

// Events over time (for statistics)
export const eventsOverTime = {
	data: [
		{ date: '2024-12-14', count: 120 },
		{ date: '2024-12-15', count: 150 },
		{ date: '2024-12-16', count: 180 },
		{ date: '2024-12-17', count: 140 },
		{ date: '2024-12-18', count: 200 },
		{ date: '2024-12-19', count: 175 },
		{ date: '2024-12-20', count: 160 }
	]
};

// Events by kind (for statistics)
export const eventsByKind = {
	data: [
		{ kind: 1, count: 8000, label: 'Notes' },
		{ kind: 7, count: 3000, label: 'Reactions' },
		{ kind: 4, count: 500, label: 'DMs' },
		{ kind: 6, count: 400, label: 'Reposts' },
		{ kind: 3, count: 300, label: 'Follows' },
		{ kind: 0, count: 200, label: 'Metadata' }
	]
};

// Top authors (for statistics)
export const topAuthors = {
	authors: [
		{ pubkey: 'abc123hexkey', npub: 'npub1abc...', nickname: 'Alice', event_count: 300 },
		{ pubkey: 'ghi789hexkey', npub: 'npub1ghi...', nickname: 'Bob', event_count: 250 },
		{ pubkey: 'def456hexkey', npub: 'npub1def...', nickname: '', event_count: 150 }
	]
};

// Identity validation
export const identityValidation = {
	valid: {
		valid: true,
		pubkey: 'abc123hexkey',
		npub: 'npub1abc123xyz...'
	},
	invalid: {
		valid: false,
		error: 'Invalid public key format'
	}
};

// NIP-05 resolution
export const nip05Resolution = {
	success: {
		pubkey: 'abc123hexkey',
		npub: 'npub1abc123xyz...',
		nip05: 'alice@example.com'
	}
};

// Sync status
export const syncStatus = {
	idle: {
		running: false
	},
	running: {
		running: true,
		id: 'sync123',
		started_at: '2024-12-20T10:00:00Z',
		progress: {
			fetched: 500,
			imported: 450,
			duplicates: 50,
			errors: 0
		},
		pubkeys: ['abc123hexkey'],
		relays: ['wss://relay.example.com']
	},
	completed: {
		running: false,
		last_sync: {
			id: 'sync123',
			completed_at: '2024-12-20T10:05:00Z',
			result: {
				fetched: 1000,
				imported: 900,
				duplicates: 100,
				errors: 0
			}
		}
	}
};

// Sync history
export const syncHistory = {
	syncs: [
		{
			id: 'sync123',
			started_at: '2024-12-20T10:00:00Z',
			completed_at: '2024-12-20T10:05:00Z',
			status: 'completed',
			pubkeys: ['abc123hexkey'],
			relays: ['wss://relay.example.com'],
			result: { fetched: 1000, imported: 900, duplicates: 100, errors: 0 }
		}
	],
	total: 1
};

// Lightning status
export const lightningStatus = {
	connected: {
		configured: true,
		connected: true,
		node_alias: 'TestNode',
		balance_sats: 100000
	},
	disconnected: {
		configured: true,
		connected: false,
		error: 'Connection failed'
	},
	unconfigured: {
		configured: false,
		connected: false
	}
};

// Pricing tiers
export const pricingTiers = {
	tiers: [
		{ id: 'monthly', name: 'Monthly', amount_sats: 5000, duration_days: 30 },
		{ id: 'annual', name: 'Annual', amount_sats: 50000, duration_days: 365 }
	]
};

// Paid users
export const paidUsers = {
	users: [
		{
			pubkey: 'paid123hexkey',
			npub: 'npub1paid123...',
			tier: 'monthly',
			expires_at: '2025-01-20T00:00:00Z',
			created_at: '2024-12-20T00:00:00Z'
		}
	],
	total: 1
};

// Revenue
export const revenue = {
	total_sats: 55000,
	this_month_sats: 5000,
	active_subscribers: 1
};

// Public relay info (for signup)
export const publicRelayInfo = {
	enabled: {
		paid_access_enabled: true,
		name: 'Test Relay',
		description: 'A private relay',
		tiers: pricingTiers.tiers
	},
	disabled: {
		paid_access_enabled: false
	}
};

// Invoice
export const invoice = {
	payment_hash: 'hash123abc',
	bolt11: 'lnbc50u1ptest...',
	amount_sats: 5000,
	expires_at: '2024-12-20T11:00:00Z'
};

// Invoice status
export const invoiceStatus = {
	pending: { paid: false },
	paid: { paid: true }
};

// Deletion requests
export const deletionRequests = {
	requests: [
		{
			id: 'del001',
			event_id: 'event001',
			requester_pubkey: 'abc123hexkey',
			status: 'pending',
			created_at: '2024-12-20T00:00:00Z'
		}
	],
	total: 1
};

// Support config
export const supportConfig = {
	lightning_address: 'dev@getalby.com',
	bitcoin_address: 'bc1qtest...',
	github_repo: 'https://github.com/example/roostr',
	version: '0.1.1'
};

// Relay logs
export const relayLogs = {
	logs: [
		{ timestamp: '2024-12-20T10:00:00Z', level: 'INFO', message: 'Relay started' },
		{ timestamp: '2024-12-20T10:00:01Z', level: 'INFO', message: 'Listening on port 7000' },
		{ timestamp: '2024-12-20T10:00:02Z', level: 'DEBUG', message: 'Client connected' }
	]
};

// Cleanup estimate
export const cleanupEstimate = {
	event_count: 1000,
	estimated_space: 50000000
};

// Integrity check result
export const integrityCheck = {
	app_db: { ok: true, message: 'No issues found' },
	relay_db: { ok: true, message: 'No issues found' }
};

// Settings timezone
export const timezone = {
	timezone: 'America/New_York'
};
