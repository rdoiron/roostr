/**
 * Common test helper utilities for E2E tests.
 */

import { Page, expect } from '@playwright/test';

/**
 * Wait for loading indicators to disappear.
 */
export async function waitForPageLoad(page: Page) {
	// Wait for any loading spinners to disappear
	const spinner = page.locator('.animate-spin, [data-testid="loading"]').first();
	await expect(spinner).toBeHidden({ timeout: 10000 }).catch(() => {
		// Spinner might not exist, that's fine
	});
}

/**
 * Navigate to a page and wait for load.
 */
export async function navigateTo(page: Page, path: string) {
	await page.goto(path);
	await page.waitForLoadState('networkidle');
	await waitForPageLoad(page);
}

/**
 * Expect a toast/notification message to appear.
 */
export async function expectToast(page: Page, message: string) {
	await expect(page.locator(`text=${message}`)).toBeVisible({ timeout: 5000 });
}

/**
 * Fill a form field and optionally wait for validation.
 */
export async function fillField(page: Page, selector: string, value: string, waitForValidation = false) {
	await page.locator(selector).fill(value);
	if (waitForValidation) {
		await page.waitForTimeout(500);
	}
}

/**
 * Click a button and wait for response.
 */
export async function clickButton(page: Page, name: string) {
	await page.getByRole('button', { name }).click();
}

/**
 * Wait for an API call to complete.
 */
export async function waitForApiCall(page: Page, pathPattern: string) {
	await page.waitForRequest((request) => request.url().includes(pathPattern));
}

/**
 * Get text content from an element.
 */
export async function getText(page: Page, selector: string): Promise<string> {
	const text = await page.locator(selector).textContent();
	return text || '';
}

/**
 * Check if element is visible.
 */
export async function isVisible(page: Page, selector: string): Promise<boolean> {
	return page.locator(selector).isVisible();
}

/**
 * Open a modal by clicking a trigger.
 */
export async function openModal(page: Page, triggerSelector: string) {
	await page.locator(triggerSelector).click();
	await expect(page.locator('[role="dialog"]')).toBeVisible();
}

/**
 * Close a modal.
 */
export async function closeModal(page: Page) {
	// Try ESC key first
	await page.keyboard.press('Escape');
	await expect(page.locator('[role="dialog"]')).toBeHidden().catch(async () => {
		// If that doesn't work, try clicking close button
		await page.locator('[role="dialog"] button:has-text("Close")').click();
	});
}

/**
 * Format bytes for display comparison.
 */
export function formatBytes(bytes: number): string {
	const units = ['B', 'KB', 'MB', 'GB'];
	let size = bytes;
	let unitIndex = 0;
	while (size >= 1024 && unitIndex < units.length - 1) {
		size /= 1024;
		unitIndex++;
	}
	return `${size.toFixed(1)} ${units[unitIndex]}`;
}

/**
 * Format number with K/M suffix for comparison.
 */
export function formatNumber(num: number): string {
	if (num >= 1000000) {
		return `${(num / 1000000).toFixed(1)}M`;
	}
	if (num >= 1000) {
		return `${(num / 1000).toFixed(1)}K`;
	}
	return num.toString();
}
