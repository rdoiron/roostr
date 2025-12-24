/**
 * Timezone state store using Svelte 5 runes.
 * Handles timezone preference with backend persistence and localStorage cache.
 * Uses class-based pattern for reliable cross-module reactivity.
 */

import { browser } from '$app/environment';

const STORAGE_KEY = 'roostr-timezone';

class TimezoneStore {
	preference = $state('auto'); // 'auto' or IANA timezone string
	resolved = $state('UTC'); // Actual timezone being used
	loading = $state(true);
}

export const timezoneStore = new TimezoneStore();

// Common timezone options for the Settings UI dropdown
export const TIMEZONE_OPTIONS = [
	{ value: 'auto', label: 'Auto-detect' },
	{ value: 'UTC', label: 'UTC' },
	{ value: 'America/New_York', label: 'Eastern Time (US)' },
	{ value: 'America/Chicago', label: 'Central Time (US)' },
	{ value: 'America/Denver', label: 'Mountain Time (US)' },
	{ value: 'America/Los_Angeles', label: 'Pacific Time (US)' },
	{ value: 'Europe/London', label: 'London' },
	{ value: 'Europe/Paris', label: 'Paris' },
	{ value: 'Europe/Berlin', label: 'Berlin' },
	{ value: 'Asia/Tokyo', label: 'Tokyo' },
	{ value: 'Asia/Shanghai', label: 'Shanghai' },
	{ value: 'Australia/Sydney', label: 'Sydney' }
];

/**
 * Initialize timezone from localStorage (immediate) then sync with backend.
 * Call this once from the root layout on mount.
 */
export async function initializeTimezone() {
	if (!browser) return;

	// Load from localStorage first (for immediate UI)
	const cached = localStorage.getItem(STORAGE_KEY);
	if (cached) {
		timezoneStore.preference = cached;
		resolveTimezone();
	} else {
		// Default to auto-detect
		resolveTimezone();
	}

	// Then sync with backend
	try {
		const res = await fetch('/api/v1/settings/timezone');
		if (res.ok) {
			const data = await res.json();
			timezoneStore.preference = data.timezone || 'auto';
			localStorage.setItem(STORAGE_KEY, timezoneStore.preference);
			resolveTimezone();
		}
	} catch (e) {
		console.error('Failed to load timezone from backend:', e);
	} finally {
		timezoneStore.loading = false;
	}
}

/**
 * Set timezone preference.
 * Saves to localStorage immediately and persists to backend.
 * @param {string} timezone - 'auto' or IANA timezone string
 */
export async function setTimezone(timezone) {
	if (!browser) return;

	timezoneStore.preference = timezone;
	localStorage.setItem(STORAGE_KEY, timezone);
	resolveTimezone();

	// Persist to backend (fire and forget - don't block UI)
	try {
		const res = await fetch('/api/v1/settings/timezone', {
			method: 'PUT',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify({ timezone })
		});
		if (!res.ok) {
			console.error('Failed to save timezone to backend:', res.status);
		}
	} catch (e) {
		console.error('Failed to save timezone to backend:', e);
	}
}

/**
 * Resolve the actual timezone based on preference.
 */
function resolveTimezone() {
	if (!browser) return;

	if (timezoneStore.preference === 'auto') {
		timezoneStore.resolved = Intl.DateTimeFormat().resolvedOptions().timeZone;
	} else {
		timezoneStore.resolved = timezoneStore.preference;
	}
}

/**
 * Format a date in the user's selected timezone.
 * @param {Date|number|string} date - Date object, Unix timestamp, or ISO date string
 * @param {Intl.DateTimeFormatOptions} options - Formatting options
 * @returns {string} Formatted date string
 */
export function formatDateInTimezone(date, options = {}) {
	if (!browser) return '';

	let d;
	if (date instanceof Date) {
		d = date;
	} else if (typeof date === 'number') {
		// Unix timestamp - check if seconds or milliseconds
		// Timestamps before year 2001 in ms would be < 1e12
		d = new Date(date < 1e12 ? date * 1000 : date);
	} else if (typeof date === 'string') {
		d = new Date(date);
	} else {
		return ''; // Invalid input
	}

	// Check for invalid date
	if (isNaN(d.getTime())) {
		return '';
	}

	return new Intl.DateTimeFormat('en-US', {
		timeZone: timezoneStore.resolved,
		...options
	}).format(d);
}
