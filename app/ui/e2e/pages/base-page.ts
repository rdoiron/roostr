/**
 * Base page object with common locators and methods.
 */

import { Page, Locator, expect } from '@playwright/test';

export class BasePage {
	readonly page: Page;
	readonly loadingIndicator: Locator;
	readonly errorAlert: Locator;

	constructor(page: Page) {
		this.page = page;
		this.loadingIndicator = page.locator('.animate-spin, [data-testid="loading"]').first();
		this.errorAlert = page.locator('.bg-red-50, [role="alert"]').first();
	}

	async waitForPageLoad() {
		await expect(this.loadingIndicator).toBeHidden({ timeout: 10000 }).catch(() => {
			// Loading indicator might not exist
		});
	}

	async hasError(): Promise<boolean> {
		return this.errorAlert.isVisible();
	}

	async getErrorMessage(): Promise<string> {
		const text = await this.errorAlert.textContent();
		return text || '';
	}

	async clickButton(name: string) {
		await this.page.getByRole('button', { name }).click();
	}

	async expectVisible(text: string) {
		await expect(this.page.locator(`text=${text}`)).toBeVisible();
	}

	async expectHidden(text: string) {
		await expect(this.page.locator(`text=${text}`)).toBeHidden();
	}
}
