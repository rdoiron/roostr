/**
 * E2E tests for the access control page.
 */

import { test, expect } from '../../fixtures/test-fixtures';
import { AccessPage } from '../../pages/access-page';
import { mockDashboard } from '../../fixtures/api-mocks';

test.describe('Access Control', () => {
	test.beforeEach(async ({ page }) => {
		await mockDashboard(page, { accessMode: 'whitelist' });
	});

	test('displays access mode selector', async ({ page }) => {
		const accessPage = new AccessPage(page);
		await accessPage.goto();

		await expect(page.locator('text=Access Mode')).toBeVisible();
	});

	test('shows whitelist mode as selected by default', async ({ page }) => {
		const accessPage = new AccessPage(page);
		await accessPage.goto();

		const whitelistRadio = page.locator('input[value="whitelist"]');
		await expect(whitelistRadio).toBeChecked();
	});

	test('displays whitelist entries', async ({ page }) => {
		const accessPage = new AccessPage(page);
		await accessPage.goto();

		await accessPage.expectWhitelistEntries();
	});

	test('shows add pubkey button', async ({ page }) => {
		const accessPage = new AccessPage(page);
		await accessPage.goto();

		await expect(page.getByRole('button', { name: /add/i }).first()).toBeVisible();
	});

	test('opens add pubkey modal', async ({ page }) => {
		const accessPage = new AccessPage(page);
		await accessPage.goto();

		await accessPage.clickAddPubkey();
		expect(await accessPage.isAddModalOpen()).toBe(true);
	});

	test('can close add pubkey modal', async ({ page }) => {
		const accessPage = new AccessPage(page);
		await accessPage.goto();

		await accessPage.clickAddPubkey();
		await accessPage.closeModal();
		expect(await accessPage.isAddModalOpen()).toBe(false);
	});

	test('can add pubkey to whitelist', async ({ page }) => {
		await mockDashboard(page);

		// Mock successful add
		await page.route('**/api/v1/access/whitelist', async (route) => {
			if (route.request().method() === 'POST') {
				await route.fulfill({ json: { success: true } });
			} else {
				await route.continue();
			}
		});

		const accessPage = new AccessPage(page);
		await accessPage.goto();

		await accessPage.addPubkeyToWhitelist('npub1newuser123', 'New User');

		// Modal should close after successful add
		expect(await accessPage.isAddModalOpen()).toBe(false);
	});

	test('displays blacklist when blacklist mode is selected', async ({ page }) => {
		await mockDashboard(page, { accessMode: 'blacklist' });

		const accessPage = new AccessPage(page);
		await accessPage.goto();

		const blacklistRadio = page.locator('input[value="blacklist"]');
		await expect(blacklistRadio).toBeChecked();
	});

	test('shows export button', async ({ page }) => {
		const accessPage = new AccessPage(page);
		await accessPage.goto();

		await expect(page.getByRole('button', { name: /export/i })).toBeVisible();
	});

	test('shows import button', async ({ page }) => {
		const accessPage = new AccessPage(page);
		await accessPage.goto();

		await expect(page.getByRole('button', { name: /import/i })).toBeVisible();
	});

	test('shows paid mode options when paid mode selected', async ({ page }) => {
		await mockDashboard(page, { accessMode: 'paid' });

		const accessPage = new AccessPage(page);
		await accessPage.goto();

		await accessPage.expectLightningSection();
		await accessPage.expectPricingSection();
	});

	test('displays pubkey nicknames', async ({ page }) => {
		const accessPage = new AccessPage(page);
		await accessPage.goto();

		// From mock data, Alice should be visible
		await expect(page.locator('text=Alice')).toBeVisible();
	});

	test('displays event counts for pubkeys', async ({ page }) => {
		const accessPage = new AccessPage(page);
		await accessPage.goto();

		// Should show event count from mock data
		await expect(page.locator('text=/\\d+ events?/i').first()).toBeVisible();
	});
});
