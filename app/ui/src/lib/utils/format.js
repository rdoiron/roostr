/**
 * Formatting utility functions for the dashboard.
 */

/**
 * Format seconds into "Xd Xh Xm" uptime string.
 * @param {number} seconds - Total seconds
 * @returns {string} Formatted uptime string
 */
export function formatUptime(seconds) {
	if (!seconds || seconds < 0) return '0m';

	const days = Math.floor(seconds / 86400);
	const hours = Math.floor((seconds % 86400) / 3600);
	const minutes = Math.floor((seconds % 3600) / 60);

	const parts = [];
	if (days > 0) parts.push(`${days}d`);
	if (hours > 0) parts.push(`${hours}h`);
	if (minutes > 0 || parts.length === 0) parts.push(`${minutes}m`);

	return parts.join(' ');
}

/**
 * Format bytes into human-readable string (KB, MB, GB).
 * @param {number} bytes - Byte count
 * @returns {string} Formatted string
 */
export function formatBytes(bytes) {
	if (!bytes || bytes === 0) return '0 B';

	const units = ['B', 'KB', 'MB', 'GB', 'TB'];
	const i = Math.floor(Math.log(bytes) / Math.log(1024));
	const value = bytes / Math.pow(1024, i);

	return `${value.toFixed(i > 0 ? 1 : 0)} ${units[i]}`;
}

/**
 * Format a timestamp into relative time string (e.g., "2 min ago").
 * @param {number|string} timestamp - Unix timestamp in seconds or ISO date string
 * @returns {string} Relative time string
 */
export function formatRelativeTime(timestamp) {
	let seconds;
	if (typeof timestamp === 'string') {
		seconds = Math.floor(new Date(timestamp).getTime() / 1000);
	} else {
		seconds = timestamp;
	}

	const now = Math.floor(Date.now() / 1000);
	const diff = now - seconds;

	if (diff < 60) return 'just now';
	if (diff < 3600) {
		const mins = Math.floor(diff / 60);
		return `${mins} min ago`;
	}
	if (diff < 86400) {
		const hours = Math.floor(diff / 3600);
		return `${hours} hour${hours > 1 ? 's' : ''} ago`;
	}
	const days = Math.floor(diff / 86400);
	return `${days} day${days > 1 ? 's' : ''} ago`;
}

/**
 * Get human-readable label for Nostr event kind.
 * @param {number} kind - Nostr event kind
 * @returns {string} Human-readable label
 */
export function getKindLabel(kind) {
	const kinds = {
		0: 'Profile',
		1: 'Post',
		3: 'Follows',
		4: 'DM',
		6: 'Repost',
		7: 'Reaction',
		14: 'DM'
	};
	return kinds[kind] || `Kind ${kind}`;
}

/**
 * Truncate a pubkey for display.
 * @param {string} pubkey - Full pubkey hex string
 * @returns {string} Truncated pubkey
 */
export function truncatePubkey(pubkey) {
	if (!pubkey || pubkey.length < 16) return pubkey || '';
	return pubkey.slice(0, 8) + '...' + pubkey.slice(-4);
}

/**
 * Format a number into compact form (e.g., 1.2K, 3.5M).
 * @param {number} num - Number to format
 * @returns {string} Compact formatted string
 */
export function formatCompactNumber(num) {
	if (!num || num < 1000) return (num || 0).toString();
	if (num < 1000000) return (num / 1000).toFixed(num < 10000 ? 1 : 0) + 'K';
	if (num < 1000000000) return (num / 1000000).toFixed(num < 10000000 ? 1 : 0) + 'M';
	return (num / 1000000000).toFixed(1) + 'B';
}
