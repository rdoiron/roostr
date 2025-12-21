/**
 * E2E tests for the support page.
 */

import { test, expect } from '../../fixtures/test-fixtures';
import { SupportPage } from '../../pages/support-page';
import { mockDashboard } from '../../fixtures/api-mocks';

test.describe('Support', () => {
	test.beforeEach(async ({ page }) => {
		await mockDashboard(page);
	});

	test('displays support page', async ({ page }) => {
		const supportPage = new SupportPage(page);
		await supportPage.goto();

		// Page heading is "Support Roostr"
		await expect(page.getByRole('heading', { name: 'Support Roostr' })).toBeVisible();
	});

	test('shows donation section', async ({ page }) => {
		const supportPage = new SupportPage(page);
		await supportPage.goto();

		await supportPage.expectDonationOptions();
	});

	test('shows Lightning address', async ({ page }) => {
		const supportPage = new SupportPage(page);
		await supportPage.goto();

		await supportPage.expectLightningSection();

		// From mock data
		await expect(page.locator('text=dev@getalby.com')).toBeVisible();
	});

	test('shows Bitcoin address', async ({ page }) => {
		const supportPage = new SupportPage(page);
		await supportPage.goto();

		// Click Bitcoin button to switch to Bitcoin address
		await page.getByRole('button', { name: /bitcoin/i }).click();

		// From mock data - check that bitcoin address is displayed
		await expect(page.locator('text=bc1qtest')).toBeVisible();
	});

	test('displays QR codes', async ({ page }) => {
		const supportPage = new SupportPage(page);
		await supportPage.goto();

		// Wait for QR code to be visible
		await expect(supportPage.qrCodes.first()).toBeVisible();
	});

	test('shows about section', async ({ page }) => {
		const supportPage = new SupportPage(page);
		await supportPage.goto();

		await supportPage.expectAboutSection();
	});

	test('displays version number', async ({ page }) => {
		const supportPage = new SupportPage(page);
		await supportPage.goto();

		const version = await supportPage.getVersion();
		expect(version).toMatch(/\d+\.\d+\.\d+/);
	});

	test('has GitHub links', async ({ page }) => {
		const supportPage = new SupportPage(page);
		await supportPage.goto();

		await supportPage.expectExternalLinks();
	});

	test('GitHub links open in new tab', async ({ page }) => {
		const supportPage = new SupportPage(page);
		await supportPage.goto();

		const githubLink = page.locator('a[href*="github.com"]').first();
		await expect(githubLink).toHaveAttribute('target', '_blank');
	});
});
