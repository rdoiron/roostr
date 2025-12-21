/**
 * E2E tests for the dashboard home page.
 */

import { test, expect } from '../../fixtures/test-fixtures';
import { DashboardPage } from '../../pages/dashboard-page';
import { mockDashboard } from '../../fixtures/api-mocks';

test.describe('Dashboard', () => {
	test.beforeEach(async ({ page }) => {
		await mockDashboard(page);
	});

	test('displays relay status as online', async ({ page }) => {
		const dashboard = new DashboardPage(page);
		await dashboard.goto();

		await expect(page.locator('text=Relay Status')).toBeVisible();
		expect(await dashboard.isRelayOnline()).toBe(true);
	});

	test('shows statistics cards with data', async ({ page }) => {
		const dashboard = new DashboardPage(page);
		await dashboard.goto();

		await dashboard.expectStatCards();

		// Check for formatted numbers
		await expect(page.locator('text=Total Events').locator('..').locator('text=/\\d/')).toBeVisible();
	});

	test('displays event type breakdown', async ({ page }) => {
		const dashboard = new DashboardPage(page);
		await dashboard.goto();

		await dashboard.expectEventTypeBreakdown();
		await expect(page.locator('text=Posts')).toBeVisible();
		await expect(page.locator('text=Reactions')).toBeVisible();
		await expect(page.locator('text=DMs')).toBeVisible();
		await expect(page.locator('text=Reposts')).toBeVisible();
		await expect(page.locator('text=Follows')).toBeVisible();
	});

	test('shows relay URLs section', async ({ page }) => {
		const dashboard = new DashboardPage(page);
		await dashboard.goto();

		expect(await dashboard.hasLocalUrl()).toBe(true);
	});

	test('shows Tor URL when available', async ({ page }) => {
		const dashboard = new DashboardPage(page);
		await dashboard.goto();

		expect(await dashboard.hasTorUrl()).toBe(true);
	});

	test('shows recent activity feed', async ({ page }) => {
		const dashboard = new DashboardPage(page);
		await dashboard.goto();

		await expect(page.locator('text=Recent Activity')).toBeVisible();
	});

	test('shows quick actions', async ({ page }) => {
		const dashboard = new DashboardPage(page);
		await dashboard.goto();

		await expect(page.locator('text=Quick Actions')).toBeVisible();
		await expect(page.locator('text=Sync from Relays')).toBeVisible();
		await expect(page.locator('text=Export Events')).toBeVisible();
	});

	test('shows storage card with progress', async ({ page }) => {
		const dashboard = new DashboardPage(page);
		await dashboard.goto();

		await expect(page.locator('text=Storage')).toBeVisible();
	});

	test('displays uptime', async ({ page }) => {
		const dashboard = new DashboardPage(page);
		await dashboard.goto();

		await expect(page.locator('text=Uptime')).toBeVisible();
	});

	test('shows offline status when relay is down', async ({ page }) => {
		await mockDashboard(page, { relayOnline: false });

		const dashboard = new DashboardPage(page);
		await dashboard.goto();

		expect(await dashboard.isRelayOffline()).toBe(true);
	});

	test('SSE connection receives stats updates', async ({ page }) => {
		await mockDashboard(page);

		const dashboard = new DashboardPage(page);
		await dashboard.goto();

		// Stats should be populated from SSE
		await expect(page.locator('text=Total Events').locator('..').locator('text=/\\d/')).toBeVisible();
	});

	test('handles SSE connection error gracefully', async ({ page }) => {
		// Mock SSE to fail
		await page.route('**/api/v1/stats/stream', async (route) => {
			await route.abort('failed');
		});

		// Still mock other endpoints
		await mockDashboard(page);

		const dashboard = new DashboardPage(page);
		await dashboard.goto();

		// Page should still be functional even if SSE fails initially
		await expect(page.locator('text=Dashboard')).toBeVisible();
	});
});
