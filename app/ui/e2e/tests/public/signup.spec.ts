/**
 * E2E tests for the paid signup flow.
 */

import { test, expect } from '../../fixtures/test-fixtures';
import { setupApiMocks } from '../../fixtures/api-mocks';

test.describe('Signup Flow', () => {
	test('shows unavailable when paid access is disabled', async ({ page }) => {
		await setupApiMocks(page, { paidAccessEnabled: false });
		await page.goto('/signup');
		await page.waitForLoadState('networkidle');

		await expect(page.locator('text=Signup Unavailable')).toBeVisible();
	});

	test('displays pricing tiers when paid access is enabled', async ({ page }) => {
		await setupApiMocks(page, { paidAccessEnabled: true });
		await page.goto('/signup');
		await page.waitForLoadState('networkidle');

		await expect(page.locator('text=Choose Your Plan')).toBeVisible();
		await expect(page.locator('text=Monthly')).toBeVisible();
	});

	test('shows monthly and annual plans', async ({ page }) => {
		await setupApiMocks(page, { paidAccessEnabled: true });
		await page.goto('/signup');
		await page.waitForLoadState('networkidle');

		await expect(page.locator('text=Monthly')).toBeVisible();
		await expect(page.locator('text=Annual')).toBeVisible();
	});

	test('selecting a plan shows identity step', async ({ page }) => {
		await setupApiMocks(page, { paidAccessEnabled: true });
		await page.goto('/signup');
		await page.waitForLoadState('networkidle');

		// Click on the first plan card (monthly)
		await page.locator('button:has-text("Select")').first().click();

		// Should show identity step
		await expect(page.locator('text=Your Nostr Identity')).toBeVisible();
		await expect(page.locator('input#pubkey')).toBeVisible();
	});

	test('can go back from identity step to plan selection', async ({ page }) => {
		await setupApiMocks(page, { paidAccessEnabled: true });
		await page.goto('/signup');
		await page.waitForLoadState('networkidle');

		// Go to identity step
		await page.locator('button:has-text("Select")').first().click();
		await expect(page.locator('text=Your Nostr Identity')).toBeVisible();

		// Go back
		await page.locator('text=Back to plans').click();

		// Should see plans again
		await expect(page.locator('text=Choose Your Plan')).toBeVisible();
	});

	test('displays relay info on signup page', async ({ page }) => {
		await setupApiMocks(page, { paidAccessEnabled: true });
		await page.goto('/signup');
		await page.waitForLoadState('networkidle');

		// Should show relay name from mock data
		await expect(page.locator('text=Test Relay')).toBeVisible();
	});
});
