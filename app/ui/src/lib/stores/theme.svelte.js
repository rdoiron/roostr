/**
 * Theme state store using Svelte 5 runes.
 * Handles light/dark/auto theme modes with localStorage persistence.
 * Uses class-based pattern for reliable cross-module reactivity.
 */

import { browser } from '$app/environment';

const STORAGE_KEY = 'roostr-theme';

class ThemeStore {
	preference = $state('auto');
	resolved = $state('light');
}

export const themeStore = new ThemeStore();

/**
 * Initialize theme from localStorage and set up system preference listener.
 * Call this once from the root layout on mount.
 */
export function initializeTheme() {
	if (!browser) return;

	// Load saved preference
	const saved = localStorage.getItem(STORAGE_KEY);
	if (saved && ['light', 'dark', 'auto'].includes(saved)) {
		themeStore.preference = saved;
	}

	// Resolve and apply theme
	resolveAndApplyTheme();

	// Listen for system preference changes (only matters in auto mode)
	window.matchMedia('(prefers-color-scheme: dark)').addEventListener('change', () => {
		if (themeStore.preference === 'auto') {
			resolveAndApplyTheme();
		}
	});
}

/**
 * Set theme preference.
 * @param {'light' | 'dark' | 'auto'} preference
 */
export function setTheme(preference) {
	if (!['light', 'dark', 'auto'].includes(preference)) return;

	themeStore.preference = preference;

	if (browser) {
		localStorage.setItem(STORAGE_KEY, preference);
		resolveAndApplyTheme();
	}
}

/**
 * Resolve the actual theme based on preference and system settings,
 * then apply it to the document.
 */
function resolveAndApplyTheme() {
	if (!browser) return;

	const resolved = themeStore.preference === 'auto'
		? (window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light')
		: themeStore.preference;

	themeStore.resolved = resolved;

	// Apply to document
	if (resolved === 'dark') {
		document.documentElement.classList.add('dark');
	} else {
		document.documentElement.classList.remove('dark');
	}
}
