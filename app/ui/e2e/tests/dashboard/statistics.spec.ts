/**
 * E2E tests for the statistics page.
 */

import { test, expect } from '../../fixtures/test-fixtures';
import { StatisticsPage } from '../../pages/statistics-page';
import { mockDashboard } from '../../fixtures/api-mocks';

test.describe('Statistics', () => {
	test.beforeEach(async ({ page }) => {
		await mockDashboard(page);
	});

	test('displays statistics page', async ({ page }) => {
		const statsPage = new StatisticsPage(page);
		await statsPage.goto();

		await expect(page.locator('text=Statistics')).toBeVisible();
	});

	test('shows events over time chart', async ({ page }) => {
		const statsPage = new StatisticsPage(page);
		await statsPage.goto();

		expect(await statsPage.hasEventsOverTimeChart()).toBe(true);
	});

	test('shows events by kind chart', async ({ page }) => {
		const statsPage = new StatisticsPage(page);
		await statsPage.goto();

		expect(await statsPage.hasEventsByKindChart()).toBe(true);
	});

	test('shows top authors section', async ({ page }) => {
		const statsPage = new StatisticsPage(page);
		await statsPage.goto();

		await statsPage.expectTopAuthors();
	});

	test('displays time range selector', async ({ page }) => {
		const statsPage = new StatisticsPage(page);
		await statsPage.goto();

		await expect(statsPage.timeRangeSelector).toBeVisible();
	});

	test('can change time range', async ({ page }) => {
		const statsPage = new StatisticsPage(page);
		await statsPage.goto();

		await statsPage.selectTimeRange('30days');

		// Page should update without error
		await statsPage.expectCharts();
	});

	test('renders chart canvases', async ({ page }) => {
		const statsPage = new StatisticsPage(page);
		await statsPage.goto();

		const chartCount = await statsPage.getChartCount();
		expect(chartCount).toBeGreaterThan(0);
	});

	test('shows author nicknames in top authors', async ({ page }) => {
		const statsPage = new StatisticsPage(page);
		await statsPage.goto();

		// From mock data, Alice should be in top authors
		await expect(page.locator('text=Alice')).toBeVisible();
	});
});
