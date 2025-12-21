/**
 * Access control page object.
 */

import { Page, Locator, expect } from '@playwright/test';
import { BasePage } from './base-page';

export class AccessPage extends BasePage {
	readonly modeSelector: Locator;
	readonly whitelistTab: Locator;
	readonly blacklistTab: Locator;
	readonly paidTab: Locator;
	readonly addButton: Locator;
	readonly pubkeyList: Locator;
	readonly searchInput: Locator;
	readonly importButton: Locator;
	readonly exportButton: Locator;

	constructor(page: Page) {
		super(page);
		this.modeSelector = page.locator('text=Access Mode').locator('..');
		this.whitelistTab = page.getByRole('tab', { name: /whitelist/i });
		this.blacklistTab = page.getByRole('tab', { name: /blacklist/i });
		this.paidTab = page.getByRole('tab', { name: /paid/i });
		this.addButton = page.getByRole('button', { name: /add/i }).first();
		this.pubkeyList = page.locator('[data-testid="pubkey-list"], .space-y-2');
		this.searchInput = page.locator('input[placeholder*="Search"]');
		this.importButton = page.getByRole('button', { name: /import/i });
		this.exportButton = page.getByRole('button', { name: /export/i });
	}

	async goto() {
		await this.page.goto('/access');
		await this.waitForPageLoad();
	}

	async selectMode(mode: 'whitelist' | 'blacklist' | 'paid' | 'open') {
		await this.page.locator(`input[value="${mode}"]`).check();
	}

	async clickWhitelistTab() {
		await this.page.locator('text=Whitelist').first().click();
	}

	async clickBlacklistTab() {
		await this.page.locator('text=Blacklist').first().click();
	}

	async clickPaidTab() {
		await this.page.locator('text=Paid').first().click();
	}

	async clickAddPubkey() {
		await this.addButton.click();
	}

	async addPubkeyToWhitelist(value: string, nickname?: string) {
		await this.clickAddPubkey();
		await this.page.locator('input[placeholder*="npub"]').fill(value);
		// Wait for validation to complete (green checkmark appears, button becomes enabled)
		await this.page.locator('text=Validated').waitFor({ timeout: 5000 });
		if (nickname) {
			// Nickname input has placeholder "e.g., Family, Friend"
			await this.page.locator('#extra-input').fill(nickname);
		}
		// Click the Add button in the modal footer and wait for modal to close
		await this.page.getByRole('button', { name: /add to whitelist/i }).click();
		// Wait for modal to close
		await this.page.locator('[role="dialog"]').waitFor({ state: 'hidden', timeout: 5000 });
	}

	async removePubkey(index: number = 0) {
		const removeButtons = this.page.locator('button').filter({ hasText: /remove/i });
		await removeButtons.nth(index).click();
		// Confirm deletion
		await this.page.getByRole('button', { name: /confirm|remove|delete/i }).click();
	}

	async searchPubkeys(query: string) {
		await this.searchInput.fill(query);
	}

	async getWhitelistCount(): Promise<number> {
		const entries = this.page.locator('[data-testid="whitelist-entry"]');
		return entries.count();
	}

	async exportWhitelist() {
		const downloadPromise = this.page.waitForEvent('download');
		await this.exportButton.click();
		return downloadPromise;
	}

	async isAddModalOpen(): Promise<boolean> {
		return this.page.locator('[role="dialog"]').isVisible();
	}

	async closeModal() {
		await this.page.keyboard.press('Escape');
	}

	async expectWhitelistEntries() {
		// Use .first() since there may be multiple npub elements
		await expect(this.page.locator('text=npub1').first()).toBeVisible();
	}

	async expectLightningSection() {
		// LightningSection has heading "Lightning Node"
		// Need longer timeout for async loading of paid mode sections
		await expect(this.page.locator('text=Lightning Node')).toBeVisible({ timeout: 10000 });
	}

	async expectPricingSection() {
		// PricingSection has heading "Pricing"
		await expect(this.page.locator('text=/pricing/i').first()).toBeVisible({ timeout: 10000 });
	}
}
