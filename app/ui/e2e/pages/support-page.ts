/**
 * Support page object.
 */

import { Page, Locator, expect } from '@playwright/test';
import { BasePage } from './base-page';

export class SupportPage extends BasePage {
	readonly donationSection: Locator;
	readonly lightningAddress: Locator;
	readonly bitcoinAddress: Locator;
	readonly qrCodes: Locator;
	readonly helpLinks: Locator;
	readonly aboutSection: Locator;
	readonly versionText: Locator;
	readonly githubLink: Locator;

	constructor(page: Page) {
		super(page);
		this.donationSection = page.locator('text=Support Roostr').locator('..');
		this.lightningAddress = page.locator('text=Lightning').locator('..');
		this.bitcoinAddress = page.locator('text=Bitcoin').locator('..');
		this.qrCodes = page.locator('canvas, img[alt*="QR"]').or(page.getByRole('img', { name: /qr/i }));
		this.helpLinks = page.locator('a[href*="github"]');
		this.aboutSection = page.locator('text=About').locator('..');
		this.versionText = page.locator('text=/v?\\d+\\.\\d+\\.\\d+/');
		this.githubLink = page.locator('a[href*="github.com"]');
	}

	async goto() {
		await this.page.goto('/support');
		await this.waitForPageLoad();
	}

	async hasLightningAddress(): Promise<boolean> {
		return this.lightningAddress.isVisible();
	}

	async hasBitcoinAddress(): Promise<boolean> {
		return this.bitcoinAddress.isVisible();
	}

	async hasQrCodes(): Promise<boolean> {
		return (await this.qrCodes.count()) > 0;
	}

	async getVersion(): Promise<string> {
		const text = await this.versionText.textContent();
		return text || '';
	}

	async clickGithubLink() {
		const link = this.githubLink.first();
		const href = await link.getAttribute('href');
		return href;
	}

	async expectDonationOptions() {
		// Check for donation toggle buttons (Bitcoin/Lightning)
		await expect(this.page.getByRole('button', { name: /bitcoin/i })).toBeVisible();
		await expect(this.page.getByRole('button', { name: /lightning/i })).toBeVisible();
	}

	async expectLightningSection() {
		await expect(this.page.locator('text=Lightning')).toBeVisible();
	}

	async expectBitcoinSection() {
		await expect(this.page.locator('text=Bitcoin')).toBeVisible();
	}

	async expectAboutSection() {
		await expect(this.page.locator('text=About')).toBeVisible();
	}

	async expectExternalLinks() {
		const githubLinks = await this.githubLink.count();
		expect(githubLinks).toBeGreaterThan(0);
	}
}
