/**
 * E2E tests for the settings page.
 */

import { test, expect } from '../../fixtures/test-fixtures';
import { SettingsPage } from '../../pages/settings-page';
import { mockDashboard } from '../../fixtures/api-mocks';

test.describe('Settings', () => {
	test.beforeEach(async ({ page }) => {
		await mockDashboard(page);
	});

	test('displays settings page', async ({ page }) => {
		const settingsPage = new SettingsPage(page);
		await settingsPage.goto();

		await expect(page.locator('text=Settings')).toBeVisible();
	});

	test('shows appearance section', async ({ page }) => {
		const settingsPage = new SettingsPage(page);
		await settingsPage.goto();

		await settingsPage.expectAppearanceSection();
	});

	test('shows relay identity section', async ({ page }) => {
		const settingsPage = new SettingsPage(page);
		await settingsPage.goto();

		await settingsPage.expectIdentitySection();
	});

	test('shows limits section', async ({ page }) => {
		const settingsPage = new SettingsPage(page);
		await settingsPage.goto();

		await settingsPage.expectLimitsSection();
	});

	test('shows relay control section', async ({ page }) => {
		const settingsPage = new SettingsPage(page);
		await settingsPage.goto();

		await settingsPage.expectRelayControlSection();
	});

	test('can toggle dark mode', async ({ page }) => {
		const settingsPage = new SettingsPage(page);
		await settingsPage.goto();

		await settingsPage.selectDarkTheme();

		// HTML should have dark class
		expect(await settingsPage.isDarkMode()).toBe(true);
	});

	test('can toggle light mode', async ({ page }) => {
		const settingsPage = new SettingsPage(page);
		await settingsPage.goto();

		// First set dark
		await settingsPage.selectDarkTheme();
		expect(await settingsPage.isDarkMode()).toBe(true);

		// Then set light
		await settingsPage.selectLightTheme();
		expect(await settingsPage.isDarkMode()).toBe(false);
	});

	test('can update relay name', async ({ page }) => {
		const settingsPage = new SettingsPage(page);
		await settingsPage.goto();

		await settingsPage.setRelayName('Updated Relay Name');
		await settingsPage.saveConfig();

		await settingsPage.expectSaveSuccess();
	});

	test('can update relay description', async ({ page }) => {
		const settingsPage = new SettingsPage(page);
		await settingsPage.goto();

		await settingsPage.setRelayDescription('Updated description for my relay');
		await settingsPage.saveConfig();

		await settingsPage.expectSaveSuccess();
	});

	test('shows reload button', async ({ page }) => {
		const settingsPage = new SettingsPage(page);
		await settingsPage.goto();

		await expect(page.getByRole('button', { name: /reload/i })).toBeVisible();
	});

	test('shows restart button', async ({ page }) => {
		const settingsPage = new SettingsPage(page);
		await settingsPage.goto();

		await expect(page.getByRole('button', { name: /restart/i })).toBeVisible();
	});

	test('can reload relay config', async ({ page }) => {
		const settingsPage = new SettingsPage(page);
		await settingsPage.goto();

		await settingsPage.reloadRelay();

		// Should show success feedback
		await expect(page.locator('text=/reloaded|success/i')).toBeVisible();
	});

	test('shows relay status', async ({ page }) => {
		const settingsPage = new SettingsPage(page);
		await settingsPage.goto();

		// From mock data, relay is running
		await expect(page.locator('text=Running')).toBeVisible();
	});

	test('loads existing config values', async ({ page }) => {
		const settingsPage = new SettingsPage(page);
		await settingsPage.goto();

		// From mock data
		await expect(settingsPage.relayNameInput).toHaveValue('Test Relay');
	});
});
