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
		if (nickname) {
			await this.page.locator('input[placeholder*="nickname"]').fill(nickname);
		}
		await this.page.getByRole('button', { name: /add/i }).last().click();
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
		await expect(this.page.locator('text=npub1')).toBeVisible();
	}

	async expectLightningSection() {
		await expect(this.page.locator('text=Lightning')).toBeVisible();
	}

	async expectPricingSection() {
		await expect(this.page.locator('text=Pricing')).toBeVisible();
	}
}
