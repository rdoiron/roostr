/**
 * Dashboard page object.
 */

import { Page, Locator, expect } from '@playwright/test';
import { BasePage } from './base-page';

export class DashboardPage extends BasePage {
	readonly statusIndicator: Locator;
	readonly totalEventsCard: Locator;
	readonly storageCard: Locator;
	readonly whitelistedCard: Locator;
	readonly relayUrlCards: Locator;
	readonly eventTypeCards: Locator;
	readonly recentActivityFeed: Locator;
	readonly quickActions: Locator;

	constructor(page: Page) {
		super(page);
		this.statusIndicator = page.locator('text=online, text=offline').first();
		this.totalEventsCard = page.locator('text=Total Events').locator('..');
		this.storageCard = page.locator('text=Storage Used').locator('..');
		this.whitelistedCard = page.locator('text=Whitelisted Pubkeys').locator('..');
		this.relayUrlCards = page.locator('.rounded-lg').filter({ has: page.locator('text=ws://') });
		this.eventTypeCards = page.locator('text=Posts, text=Reactions, text=DMs, text=Reposts, text=Follows, text=Other');
		this.recentActivityFeed = page.locator('text=Recent Activity').locator('..');
		this.quickActions = page.locator('text=Quick Actions').locator('..');
	}

	async goto() {
		await this.page.goto('/');
		await this.waitForPageLoad();
	}

	async isRelayOnline(): Promise<boolean> {
		return this.page.getByText('Online', { exact: true }).first().isVisible();
	}

	async isRelayOffline(): Promise<boolean> {
		return this.page.getByText('Offline', { exact: true }).first().isVisible();
	}

	async getTotalEvents(): Promise<string> {
		const card = this.page.locator('text=Total Events').locator('..').locator('p, span').first();
		return (await card.textContent()) || '';
	}

	async getStorageUsed(): Promise<string> {
		const card = this.page.locator('text=Storage Used').locator('..').locator('p, span').first();
		return (await card.textContent()) || '';
	}

	async getWhitelistedCount(): Promise<string> {
		const card = this.page.locator('text=Whitelisted Pubkeys').locator('..').locator('p, span').first();
		return (await card.textContent()) || '';
	}

	async copyRelayUrl(index: number = 0) {
		const copyButtons = this.page.locator('button').filter({ hasText: /copy/i });
		await copyButtons.nth(index).click();
	}

	async hasLocalUrl(): Promise<boolean> {
		return this.page.locator('text=Local Network').isVisible();
	}

	async hasTorUrl(): Promise<boolean> {
		// The label is "Tor (Remote Access)"
		return this.page.locator('text=Tor (Remote Access)').isVisible();
	}

	async clickSyncFromRelays() {
		await this.page.locator('text=Sync from Relays').click();
	}

	async clickExportEvents() {
		await this.page.locator('text=Export Events').click();
	}

	async getRecentActivityCount(): Promise<number> {
		const feed = this.page.locator('text=Recent Activity').locator('..').locator('a, button');
		return feed.count();
	}

	async expectStatCards() {
		await expect(this.page.locator('text=Total Events')).toBeVisible();
		await expect(this.page.locator('text=Storage Used')).toBeVisible();
		await expect(this.page.locator('text=Whitelisted Pubkeys')).toBeVisible();
	}

	async expectEventTypeBreakdown() {
		await expect(this.page.getByText('Posts', { exact: true })).toBeVisible();
		await expect(this.page.getByText('Reactions', { exact: true })).toBeVisible();
		await expect(this.page.getByText('DMs', { exact: true })).toBeVisible();
	}
}
