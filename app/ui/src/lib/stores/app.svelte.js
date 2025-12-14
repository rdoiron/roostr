/**
 * Application-wide state stores using Svelte 5 runes.
 */

// Relay status
export const relayStatus = $state({
	online: false,
	uptime: 0,
	eventCount: 0,
	loading: true
});

// Setup state
export const setupState = $state({
	completed: false,
	currentStep: 0,
	loading: true
});

// Sync status for cross-component communication
export const syncStatus = $state({
	running: false,
	jobId: null,
	progress: null,
	lastSyncTime: null
});

// User notifications
export const notifications = $state([]);

/**
 * Add a notification.
 * @param {'info' | 'success' | 'warning' | 'error'} type
 * @param {string} message
 */
export function notify(type, message) {
	const id = Date.now();
	notifications.push({ id, type, message });
	// Auto-remove after 5 seconds
	setTimeout(() => {
		const index = notifications.findIndex((n) => n.id === id);
		if (index !== -1) notifications.splice(index, 1);
	}, 5000);
}
