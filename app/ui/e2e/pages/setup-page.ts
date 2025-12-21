/**
 * Setup wizard page object.
 */

import { Page, Locator, expect } from '@playwright/test';
import { BasePage } from './base-page';

export class SetupPage extends BasePage {
	readonly progressBar: Locator;
	readonly stepIndicator: Locator;
	readonly identityInput: Locator;
	readonly relayNameInput: Locator;
	readonly relayDescriptionInput: Locator;
	readonly accessModeRadios: Locator;
	readonly additionalPubkeyInput: Locator;

	constructor(page: Page) {
		super(page);
		this.progressBar = page.locator('.bg-purple-600').first();
		this.stepIndicator = page.locator('text=/Step \\d+ of \\d+/');
		this.identityInput = page.locator('#identity');
		this.relayNameInput = page.locator('#relay-name');
		this.relayDescriptionInput = page.locator('#description');
		this.accessModeRadios = page.locator('input[type="radio"]');
		this.additionalPubkeyInput = page.locator('input[placeholder*="npub"], input[placeholder*="NIP-05"]');
	}

	async goto() {
		await this.page.goto('/setup');
		await this.waitForPageLoad();
	}

	async getCurrentStep(): Promise<number> {
		const text = await this.stepIndicator.textContent().catch(() => null);
		if (!text) return 0;
		const match = text.match(/Step (\d+) of (\d+)/);
		return match ? parseInt(match[1]) : 0;
	}

	async isOnWelcomeStep(): Promise<boolean> {
		return this.page.locator('text=Welcome to Roostr').isVisible();
	}

	async clickGetStarted() {
		// Ensure app is fully hydrated first
		await this.page.waitForLoadState('domcontentloaded');
		await this.page.waitForLoadState('networkidle');

		// Wait a moment for Svelte hydration
		await this.page.waitForTimeout(1000);

		// Click the button
		await this.page.getByRole('button', { name: 'Get Started' }).click();

		// Wait for the "Your Identity" heading to appear
		await this.page.locator('h2:has-text("Your Identity")').waitFor({ state: 'visible', timeout: 10000 });
	}

	async clickContinue() {
		await this.page.getByRole('button', { name: 'Continue' }).click();
		// Wait for page to update
		await this.page.waitForTimeout(300);
	}

	async clickBack() {
		await this.page.getByRole('button', { name: 'Back' }).click();
		// Wait for page to update
		await this.page.waitForTimeout(300);
	}

	async clickFinishSetup() {
		await this.page.getByRole('button', { name: 'Finish Setup' }).click();
		// Wait for completion
		await this.page.waitForTimeout(500);
	}

	async enterIdentity(value: string) {
		await this.identityInput.fill(value);
		// Wait for validation debounce
		await this.page.waitForTimeout(600);
	}

	async enterRelayInfo(name: string, description: string) {
		await this.relayNameInput.fill(name);
		await this.relayDescriptionInput.fill(description);
	}

	async selectAccessMode(mode: 'private' | 'public' | 'paid') {
		// Access mode uses button elements, not radio inputs
		const modeLabels: Record<string, string> = {
			private: 'Private',
			public: 'Public',
			paid: 'Paid Access'
		};
		await this.page.locator(`button:has-text("${modeLabels[mode]}")`).click();
	}

	async addAdditionalPubkey(value: string) {
		await this.additionalPubkeyInput.fill(value);
		await this.page.getByRole('button', { name: 'Add' }).click();
	}

	async isIdentityValid(): Promise<boolean> {
		return this.page.locator('text=Valid').isVisible();
	}

	async isIdentityInvalid(): Promise<boolean> {
		return this.page.locator('text=Invalid').isVisible();
	}

	async isContinueEnabled(): Promise<boolean> {
		const button = this.page.getByRole('button', { name: /Continue|Get Started|Finish Setup/ });
		return !(await button.isDisabled());
	}

	async isSetupComplete(): Promise<boolean> {
		// The completion screen shows "You're All Set!"
		return this.page.locator('text=You\'re All Set').isVisible();
	}

	async goToDashboard() {
		await this.page.getByRole('link', { name: /dashboard/i }).click();
	}
}
