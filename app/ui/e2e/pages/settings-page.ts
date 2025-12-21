/**
 * Settings page object.
 */

import { Page, Locator, expect } from '@playwright/test';
import { BasePage } from './base-page';

export class SettingsPage extends BasePage {
	readonly appearanceSection: Locator;
	readonly identitySection: Locator;
	readonly limitsSection: Locator;
	readonly relayControlSection: Locator;
	readonly lightButton: Locator;
	readonly darkButton: Locator;
	readonly relayNameInput: Locator;
	readonly relayDescriptionInput: Locator;
	readonly contactInput: Locator;
	readonly maxEventBytesInput: Locator;
	readonly saveButton: Locator;
	readonly reloadButton: Locator;
	readonly restartButton: Locator;

	constructor(page: Page) {
		super(page);
		this.appearanceSection = page.locator('text=Appearance').locator('..');
		this.identitySection = page.locator('text=Relay Identity').locator('..');
		this.limitsSection = page.locator('text=Limits').locator('..');
		this.relayControlSection = page.locator('text=Relay Control').locator('..');
		this.lightButton = page.getByRole('button', { name: /light/i });
		this.darkButton = page.getByRole('button', { name: /dark/i });
		this.relayNameInput = page.locator('input#name, input[name="name"]').first();
		this.relayDescriptionInput = page.locator('textarea#description, textarea[name="description"]');
		this.contactInput = page.locator('input#contact, input[name="contact"]');
		this.maxEventBytesInput = page.locator('input#max_event_bytes, input[name="max_event_bytes"]');
		this.saveButton = page.getByRole('button', { name: /save/i });
		this.reloadButton = page.getByRole('button', { name: /reload/i });
		this.restartButton = page.getByRole('button', { name: /restart/i });
	}

	async goto() {
		await this.page.goto('/settings');
		await this.waitForPageLoad();
	}

	async selectLightTheme() {
		await this.lightButton.click();
	}

	async selectDarkTheme() {
		await this.darkButton.click();
	}

	async isDarkMode(): Promise<boolean> {
		return this.page.locator('html.dark').count().then((c) => c > 0);
	}

	async setRelayName(name: string) {
		await this.relayNameInput.fill(name);
	}

	async setRelayDescription(description: string) {
		await this.relayDescriptionInput.fill(description);
	}

	async setContact(contact: string) {
		await this.contactInput.fill(contact);
	}

	async setMaxEventBytes(bytes: number) {
		await this.maxEventBytesInput.fill(bytes.toString());
	}

	async saveConfig() {
		await this.saveButton.click();
	}

	async reloadRelay() {
		await this.reloadButton.click();
	}

	async restartRelay() {
		await this.restartButton.click();
	}

	async confirmRestart() {
		await this.page.getByRole('button', { name: /confirm|restart/i }).last().click();
	}

	async getRelayStatus(): Promise<string> {
		const status = this.page.locator('text=Running, text=Stopped').first();
		return (await status.textContent()) || '';
	}

	async expectAppearanceSection() {
		await expect(this.page.locator('text=Appearance')).toBeVisible();
	}

	async expectIdentitySection() {
		await expect(this.page.locator('text=Relay Identity')).toBeVisible();
	}

	async expectLimitsSection() {
		await expect(this.page.locator('text=Limits')).toBeVisible();
	}

	async expectRelayControlSection() {
		await expect(this.page.locator('text=Relay Control')).toBeVisible();
	}

	async expectSaveSuccess() {
		await expect(this.page.locator('text=/saved|success/i')).toBeVisible();
	}

	async expectValidationError() {
		await expect(this.page.locator('text=/invalid|error|required/i')).toBeVisible();
	}
}
