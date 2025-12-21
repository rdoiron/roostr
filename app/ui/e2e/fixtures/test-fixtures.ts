/**
 * Extended Playwright test fixtures for Roostr E2E tests.
 */

import { test as base, expect } from '@playwright/test';
import { setupApiMocks, mockDashboard, mockSetupFlow, MockOptions } from './api-mocks';

type TestFixtures = {
	/**
	 * Set up API mocks with custom options.
	 */
	mockApi: (options?: MockOptions) => Promise<void>;

	/**
	 * Navigate to dashboard with mocked API (setup complete).
	 */
	dashboardPage: void;

	/**
	 * Navigate to setup wizard with mocked API (setup incomplete).
	 */
	setupPage: void;
};

export const test = base.extend<TestFixtures>({
	mockApi: async ({ page }, use) => {
		const setupMocks = async (options?: MockOptions) => {
			await setupApiMocks(page, options);
		};
		await use(setupMocks);
	},

	dashboardPage: async ({ page }, use) => {
		await mockDashboard(page);
		await page.goto('/');
		// Wait for initial load
		await page.waitForLoadState('networkidle');
		await use();
	},

	setupPage: async ({ page }, use) => {
		await mockSetupFlow(page);
		await page.goto('/setup');
		await page.waitForLoadState('networkidle');
		await use();
	}
});

export { expect };
