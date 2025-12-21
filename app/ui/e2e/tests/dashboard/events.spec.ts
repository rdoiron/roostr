/**
 * E2E tests for the events browser page.
 */

import { test, expect } from '../../fixtures/test-fixtures';
import { EventsPage } from '../../pages/events-page';
import { mockDashboard } from '../../fixtures/api-mocks';

test.describe('Events Browser', () => {
	test.beforeEach(async ({ page }) => {
		await mockDashboard(page);
	});

	test('displays event browser page', async ({ page }) => {
		const eventsPage = new EventsPage(page);
		await eventsPage.goto();

		await eventsPage.expectEventList();
	});

	test('shows event list with events', async ({ page }) => {
		const eventsPage = new EventsPage(page);
		await eventsPage.goto();

		const count = await eventsPage.getEventCount();
		expect(count).toBeGreaterThan(0);
	});

	test('displays filter controls', async ({ page }) => {
		const eventsPage = new EventsPage(page);
		await eventsPage.goto();

		await expect(eventsPage.kindFilter).toBeVisible();
		await expect(eventsPage.searchInput).toBeVisible();
	});

	test('can filter by event kind', async ({ page }) => {
		const eventsPage = new EventsPage(page);
		await eventsPage.goto();

		await eventsPage.filterByKind('1'); // Notes
		await eventsPage.applyFilters();

		// Should still show events (mock returns same data)
		const count = await eventsPage.getEventCount();
		expect(count).toBeGreaterThan(0);
	});

	test('can search events', async ({ page }) => {
		const eventsPage = new EventsPage(page);
		await eventsPage.goto();

		await eventsPage.filterBySearch('hello');
		await eventsPage.applyFilters();

		// Page should not error
		await eventsPage.expectEventList();
	});

	test('shows pagination controls', async ({ page }) => {
		const eventsPage = new EventsPage(page);
		await eventsPage.goto();

		await expect(page.locator('text=Showing')).toBeVisible();
	});

	// TODO: Frontend uses eventList.length === limit for pagination, not has_more from API
	// Would need to mock 50+ events to enable the Next button
	test.skip('can navigate to next page', async ({ page }) => {
		const eventsPage = new EventsPage(page);
		await eventsPage.goto();

		await expect(eventsPage.nextButton).toBeVisible();
		await eventsPage.goToNextPage();

		await expect(page.locator('text=Showing')).toBeVisible();
	});

	test('opens event detail modal when clicking view', async ({ page }) => {
		const eventsPage = new EventsPage(page);
		await eventsPage.goto();

		await eventsPage.clickViewRaw(0);
		expect(await eventsPage.isEventDetailModalOpen()).toBe(true);
	});

	test('can close event detail modal', async ({ page }) => {
		const eventsPage = new EventsPage(page);
		await eventsPage.goto();

		await eventsPage.clickViewRaw(0);
		await eventsPage.closeEventDetailModal();
		expect(await eventsPage.isEventDetailModalOpen()).toBe(false);
	});

	test('deep link opens specific event', async ({ page }) => {
		await mockDashboard(page);

		const eventsPage = new EventsPage(page);
		await eventsPage.gotoWithEventId('event001');

		// Modal should auto-open
		expect(await eventsPage.isEventDetailModalOpen()).toBe(true);
	});

	test('shows delete button for events', async ({ page }) => {
		const eventsPage = new EventsPage(page);
		await eventsPage.goto();

		const deleteButtons = page.locator('button').filter({ hasText: /delete/i });
		expect(await deleteButtons.count()).toBeGreaterThan(0);
	});

	test('can delete an event', async ({ page }) => {
		await mockDashboard(page);

		// Mock delete endpoint
		await page.route(/\/api\/v1\/events\/[^/]+$/, async (route) => {
			if (route.request().method() === 'DELETE') {
				await route.fulfill({ json: { success: true } });
			} else {
				await route.continue();
			}
		});

		const eventsPage = new EventsPage(page);
		await eventsPage.goto();

		await eventsPage.clickDeleteEvent(0);
		await eventsPage.confirmDelete();

		// Should not error
		await eventsPage.expectEventList();
	});

	test('shows no events message when empty', async ({ page }) => {
		await mockDashboard(page, { hasEvents: false });

		const eventsPage = new EventsPage(page);
		await eventsPage.goto();

		await eventsPage.expectNoEvents();
	});

	test('can filter by date range', async ({ page }) => {
		const eventsPage = new EventsPage(page);
		await eventsPage.goto();

		await eventsPage.filterByDateRange('2024-01-01', '2024-12-31');
		await eventsPage.applyFilters();

		await eventsPage.expectEventList();
	});

	test('displays event content preview', async ({ page }) => {
		const eventsPage = new EventsPage(page);
		await eventsPage.goto();

		// From mock data, "Hello world" should be visible
		await expect(page.locator('text=Hello world')).toBeVisible();
	});
});
