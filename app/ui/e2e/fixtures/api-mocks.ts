/**
 * API route interception for E2E tests.
 * Uses Playwright's route API to mock all backend responses.
 */

import { Page } from '@playwright/test';
import * as mockData from './mock-data';

export type MockOptions = {
	setupComplete?: boolean;
	accessMode?: 'whitelist' | 'blacklist' | 'paid' | 'open';
	relayOnline?: boolean;
	hasEvents?: boolean;
	lightningConnected?: boolean;
	paidAccessEnabled?: boolean;
};

/**
 * Set up all API mocks for a test.
 * Call this before navigating to pages.
 */
export async function setupApiMocks(page: Page, options: MockOptions = {}) {
	const {
		setupComplete = true,
		accessMode = 'whitelist',
		relayOnline = true,
		hasEvents = true,
		lightningConnected = false,
		paidAccessEnabled = false
	} = options;

	// Register all routes in parallel to avoid race conditions in faster browsers
	await Promise.all([
		// Setup status
		page.route('**/api/v1/setup/status', async (route) => {
			await route.fulfill({
				status: 200,
				contentType: 'application/json',
				body: JSON.stringify(
					setupComplete ? mockData.setupStatus.complete : mockData.setupStatus.incomplete
				)
			});
		}),

		// Setup validation
		page.route('**/api/v1/setup/validate-identity**', async (route) => {
			const url = new URL(route.request().url());
			const input = url.searchParams.get('input') || '';
			const isValid = input.startsWith('npub') || input.includes('@');
			await route.fulfill({
				status: 200,
				contentType: 'application/json',
				body: JSON.stringify(
					isValid ? mockData.identityValidation.valid : mockData.identityValidation.invalid
				)
			});
		}),

		// Setup complete
		page.route('**/api/v1/setup/complete', async (route) => {
			await route.fulfill({
				status: 200,
				contentType: 'application/json',
				body: JSON.stringify({ success: true })
			});
		}),

		// Stats stream (SSE)
		page.route('**/api/v1/stats/stream', async (route) => {
			const sseData = {
				stats: mockData.statsSummary,
				recentEvents: mockData.recentEvents,
				storage: mockData.storageStatus
			};
			await route.fulfill({
				status: 200,
				contentType: 'text/event-stream',
				body: `event: connected\ndata: {}\n\nevent: stats\ndata: ${JSON.stringify(sseData)}\n\n`
			});
		}),

		// Stats summary
		page.route('**/api/v1/stats/summary', async (route) => {
			await route.fulfill({ json: mockData.statsSummary });
		}),

		// Stats endpoints
		page.route('**/api/v1/stats/events-over-time**', async (route) => {
			await route.fulfill({ json: mockData.eventsOverTime });
		}),

		page.route('**/api/v1/stats/events-by-kind**', async (route) => {
			await route.fulfill({ json: mockData.eventsByKind });
		}),

		page.route('**/api/v1/stats/top-authors**', async (route) => {
			await route.fulfill({ json: mockData.topAuthors });
		}),

		// Relay URLs
		page.route('**/api/v1/relay/urls', async (route) => {
			await route.fulfill({ json: mockData.relayUrls });
		}),

		// Relay status
		page.route('**/api/v1/relay/status', async (route) => {
			if (route.request().method() === 'GET') {
				const status = relayOnline
					? mockData.relayStatus
					: { ...mockData.relayStatus, status: 'stopped', pid: 0 };
				await route.fulfill({ json: status });
			} else {
				await route.fulfill({ json: { success: true } });
			}
		}),

		// Relay control
		page.route('**/api/v1/relay/reload', async (route) => {
			await route.fulfill({ json: { success: true } });
		}),

		page.route('**/api/v1/relay/restart', async (route) => {
			await route.fulfill({ json: { success: true } });
		}),

		page.route('**/api/v1/relay/logs**', async (route) => {
			await route.fulfill({ json: mockData.relayLogs });
		}),

		// Access mode
		page.route('**/api/v1/access/mode', async (route) => {
			if (route.request().method() === 'GET') {
				await route.fulfill({ json: mockData.accessMode[accessMode] });
			} else {
				await route.fulfill({ json: { success: true } });
			}
		}),

		// Whitelist (pattern must match query strings)
		page.route(/\/api\/v1\/access\/whitelist(\?.*)?$/, async (route) => {
			if (route.request().method() === 'GET') {
				await route.fulfill({ json: mockData.whitelist });
			} else {
				await route.fulfill({ json: { success: true } });
			}
		}),

		page.route('**/api/v1/access/whitelist/**', async (route) => {
			await route.fulfill({ json: { success: true } });
		}),

		// Blacklist (pattern must match query strings)
		page.route(/\/api\/v1\/access\/blacklist(\?.*)?$/, async (route) => {
			if (route.request().method() === 'GET') {
				await route.fulfill({ json: mockData.blacklist });
			} else {
				await route.fulfill({ json: { success: true } });
			}
		}),

		page.route('**/api/v1/access/blacklist/**', async (route) => {
			await route.fulfill({ json: { success: true } });
		}),

		// NIP-05 resolution
		page.route('**/api/v1/nip05/**', async (route) => {
			await route.fulfill({ json: mockData.nip05Resolution.success });
		}),

		// Events list (pattern must match query strings)
		page.route(/\/api\/v1\/events(\?.*)?$/, async (route) => {
			if (route.request().method() === 'GET') {
				await route.fulfill({
					json: hasEvents ? mockData.eventsList : { events: [], total: 0, has_more: false }
				});
			}
		}),

		// Single event
		page.route(/\/api\/v1\/events\/[^/]+$/, async (route) => {
			if (route.request().method() === 'GET') {
				await route.fulfill({ json: mockData.eventDetail });
			} else if (route.request().method() === 'DELETE') {
				await route.fulfill({ json: { success: true } });
			}
		}),

		// Recent events
		page.route('**/api/v1/events/recent', async (route) => {
			await route.fulfill({ json: { events: mockData.recentEvents } });
		}),

		// Export estimate
		page.route('**/api/v1/events/export/estimate**', async (route) => {
			await route.fulfill({ json: { event_count: 1000, estimated_size: 50000000 } });
		}),

		// Config
		page.route('**/api/v1/config', async (route) => {
			if (route.request().method() === 'GET') {
				await route.fulfill({ json: mockData.relayConfig });
			} else {
				await route.fulfill({ json: { success: true } });
			}
		}),

		page.route('**/api/v1/config/reload', async (route) => {
			await route.fulfill({ json: { success: true } });
		}),

		// Storage
		page.route('**/api/v1/storage/status', async (route) => {
			await route.fulfill({ json: mockData.storageStatus });
		}),

		page.route('**/api/v1/storage/retention', async (route) => {
			if (route.request().method() === 'GET') {
				await route.fulfill({ json: mockData.retention });
			} else {
				await route.fulfill({ json: { success: true } });
			}
		}),

		page.route('**/api/v1/storage/cleanup', async (route) => {
			await route.fulfill({ json: { deleted_count: 100, space_freed: 10000000 } });
		}),

		page.route('**/api/v1/storage/vacuum', async (route) => {
			await route.fulfill({ json: { space_reclaimed: 10000000 } });
		}),

		page.route('**/api/v1/storage/estimate**', async (route) => {
			await route.fulfill({ json: mockData.cleanupEstimate });
		}),

		page.route('**/api/v1/storage/integrity-check', async (route) => {
			await route.fulfill({ json: mockData.integrityCheck });
		}),

		page.route('**/api/v1/storage/deletion-requests**', async (route) => {
			await route.fulfill({ json: mockData.deletionRequests });
		}),

		// Sync
		page.route('**/api/v1/sync/status**', async (route) => {
			await route.fulfill({ json: mockData.syncStatus.idle });
		}),

		page.route('**/api/v1/sync/start', async (route) => {
			await route.fulfill({ json: { id: 'sync123', started: true } });
		}),

		page.route('**/api/v1/sync/cancel', async (route) => {
			await route.fulfill({ json: { success: true } });
		}),

		page.route('**/api/v1/sync/history**', async (route) => {
			await route.fulfill({ json: mockData.syncHistory });
		}),

		page.route('**/api/v1/sync/relays', async (route) => {
			await route.fulfill({
				json: { relays: ['wss://relay.damus.io', 'wss://nos.lol', 'wss://relay.nostr.band'] }
			});
		}),

		// Lightning
		page.route('**/api/v1/lightning/status', async (route) => {
			const status = lightningConnected
				? mockData.lightningStatus.connected
				: mockData.lightningStatus.unconfigured;
			await route.fulfill({ json: status });
		}),

		page.route('**/api/v1/lightning/config', async (route) => {
			await route.fulfill({ json: { success: true } });
		}),

		page.route('**/api/v1/lightning/test', async (route) => {
			await route.fulfill({ json: { success: true, node_alias: 'TestNode' } });
		}),

		// Pricing
		page.route('**/api/v1/access/pricing', async (route) => {
			if (route.request().method() === 'GET') {
				await route.fulfill({ json: mockData.pricingTiers });
			} else {
				await route.fulfill({ json: { success: true } });
			}
		}),

		// Paid users (pattern must match query strings)
		page.route(/\/api\/v1\/access\/paid-users(\?.*)?$/, async (route) => {
			await route.fulfill({ json: mockData.paidUsers });
		}),

		page.route('**/api/v1/access/paid-users/**', async (route) => {
			await route.fulfill({ json: { success: true } });
		}),

		page.route('**/api/v1/access/revenue', async (route) => {
			await route.fulfill({ json: mockData.revenue });
		}),

		// Support
		page.route('**/api/v1/support/config', async (route) => {
			await route.fulfill({ json: mockData.supportConfig });
		}),

		// Settings
		page.route('**/api/v1/settings/timezone', async (route) => {
			if (route.request().method() === 'GET') {
				await route.fulfill({ json: mockData.timezone });
			} else {
				await route.fulfill({ json: { success: true } });
			}
		}),

		// Public endpoints (no /api/v1 prefix)
		page.route('**/public/relay-info', async (route) => {
			const info = paidAccessEnabled
				? mockData.publicRelayInfo.enabled
				: mockData.publicRelayInfo.disabled;
			await route.fulfill({ json: info });
		}),

		page.route('**/public/create-invoice', async (route) => {
			await route.fulfill({ json: mockData.invoice });
		}),

		page.route('**/public/invoice-status/**', async (route) => {
			await route.fulfill({ json: mockData.invoiceStatus.pending });
		})
	]);
}

/**
 * Mock setup flow specifically (incomplete setup state).
 */
export async function mockSetupFlow(page: Page) {
	await setupApiMocks(page, { setupComplete: false });
}

/**
 * Mock dashboard with all features enabled.
 */
export async function mockDashboard(page: Page, options: Partial<MockOptions> = {}) {
	await setupApiMocks(page, {
		setupComplete: true,
		relayOnline: true,
		hasEvents: true,
		...options
	});
}

/**
 * Mock paid access signup flow.
 */
export async function mockSignupFlow(page: Page) {
	await setupApiMocks(page, {
		setupComplete: true,
		paidAccessEnabled: true
	});

	// Override invoice status to simulate payment after delay
	let checkCount = 0;
	await page.route('**/public/invoice-status/**', async (route) => {
		checkCount++;
		const isPaid = checkCount > 2; // Paid after 2 checks
		await route.fulfill({
			json: isPaid ? mockData.invoiceStatus.paid : mockData.invoiceStatus.pending
		});
	});
}

/**
 * Mock API error response.
 */
export async function mockApiError(page: Page, pathPattern: string, status: number, message: string) {
	await page.route(`**${pathPattern}`, async (route) => {
		await route.fulfill({
			status,
			contentType: 'application/json',
			body: JSON.stringify({ error: message, code: 'ERROR' })
		});
	});
}
