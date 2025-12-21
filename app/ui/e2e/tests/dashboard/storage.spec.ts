/**
 * E2E tests for the storage management page.
 */

import { test, expect } from '../../fixtures/test-fixtures';
import { StoragePage } from '../../pages/storage-page';
import { mockDashboard } from '../../fixtures/api-mocks';

test.describe('Storage Management', () => {
	test.beforeEach(async ({ page }) => {
		await mockDashboard(page);
	});

	test('displays storage usage section', async ({ page }) => {
		const storagePage = new StoragePage(page);
		await storagePage.goto();

		await storagePage.expectUsageDisplay();
	});

	test('shows relay database size', async ({ page }) => {
		const storagePage = new StoragePage(page);
		await storagePage.goto();

		await expect(page.locator('text=Relay Database')).toBeVisible();
	});

	test('displays retention settings', async ({ page }) => {
		const storagePage = new StoragePage(page);
		await storagePage.goto();

		await storagePage.expectRetentionSettings();
	});

	test('shows maintenance tools', async ({ page }) => {
		const storagePage = new StoragePage(page);
		await storagePage.goto();

		await storagePage.expectMaintenanceTools();
	});

	test('can update retention days', async ({ page }) => {
		const storagePage = new StoragePage(page);
		await storagePage.goto();

		await storagePage.setRetentionDays(60);
		await storagePage.saveRetention();

		// Should show success feedback
		await expect(page.locator('text=/saved|success/i')).toBeVisible();
	});

	test('shows cleanup estimation', async ({ page }) => {
		const storagePage = new StoragePage(page);
		await storagePage.goto();

		await storagePage.setCleanupDate('2024-01-01');

		// Should show estimate from mock
		await expect(page.locator('text=/1,000 events|1000 events/i')).toBeVisible();
	});

	test('can run vacuum', async ({ page }) => {
		const storagePage = new StoragePage(page);
		await storagePage.goto();

		await storagePage.runVacuum();

		// Should show success
		await storagePage.expectVacuumSuccess();
	});

	test('can run integrity check', async ({ page }) => {
		const storagePage = new StoragePage(page);
		await storagePage.goto();

		await storagePage.runIntegrityCheck();

		// Should show passed status
		await storagePage.expectIntegrityCheckPassed();
	});

	test('shows cleanup confirmation dialog', async ({ page }) => {
		const storagePage = new StoragePage(page);
		await storagePage.goto();

		await storagePage.setCleanupDate('2024-01-01');
		await storagePage.deleteOldEvents();

		// Should show confirmation
		await expect(page.locator('text=/confirm|are you sure/i')).toBeVisible();
	});

	test('displays storage status indicator', async ({ page }) => {
		const storagePage = new StoragePage(page);
		await storagePage.goto();

		// From mock data, status is 'healthy'
		await expect(page.locator('text=/healthy|good|ok/i')).toBeVisible();
	});

	test('shows deletion requests section', async ({ page }) => {
		const storagePage = new StoragePage(page);
		await storagePage.goto();

		await expect(page.locator('text=Deletion Requests, text=NIP-09')).toBeVisible();
	});
});
