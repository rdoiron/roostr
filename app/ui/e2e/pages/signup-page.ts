/**
 * Signup page object (public paid relay signup).
 */

import { Page, Locator, expect } from '@playwright/test';
import { BasePage } from './base-page';

export class SignupPage extends BasePage {
	readonly planCards: Locator;
	readonly identityInput: Locator;
	readonly continueButton: Locator;
	readonly invoiceDisplay: Locator;
	readonly qrCode: Locator;
	readonly bolt11Text: Locator;
	readonly paymentStatus: Locator;
	readonly successMessage: Locator;

	constructor(page: Page) {
		super(page);
		this.planCards = page.locator('[data-testid="plan-card"], .rounded-lg').filter({
			has: page.locator('text=/month|year|sats/i')
		});
		this.identityInput = page.locator('input[placeholder*="npub"]');
		this.continueButton = page.getByRole('button', { name: /continue|next/i });
		this.invoiceDisplay = page.locator('text=lnbc').locator('..');
		this.qrCode = page.locator('canvas, img[alt*="QR"]').first();
		this.bolt11Text = page.locator('text=/^lnbc/');
		this.paymentStatus = page.locator('text=/pending|checking|waiting/i');
		this.successMessage = page.locator('text=/confirmed|success|welcome/i');
	}

	async goto() {
		await this.page.goto('/signup');
		await this.waitForPageLoad();
	}

	async isUnavailable(): Promise<boolean> {
		return this.page.locator('text=/unavailable|not available|disabled/i').isVisible();
	}

	async getPlanCount(): Promise<number> {
		return this.planCards.count();
	}

	async selectPlan(index: number = 0) {
		await this.planCards.nth(index).click();
	}

	async enterIdentity(value: string) {
		await this.identityInput.fill(value);
		await this.page.waitForTimeout(600); // Wait for validation
	}

	async clickContinue() {
		await this.continueButton.click();
	}

	async hasInvoiceDisplayed(): Promise<boolean> {
		return this.invoiceDisplay.isVisible();
	}

	async hasQrCode(): Promise<boolean> {
		return this.qrCode.isVisible();
	}

	async getBolt11(): Promise<string> {
		const text = await this.bolt11Text.textContent();
		return text || '';
	}

	async isPaymentPending(): Promise<boolean> {
		return this.paymentStatus.isVisible();
	}

	async isPaymentConfirmed(): Promise<boolean> {
		return this.successMessage.isVisible();
	}

	async expectPlanSelection() {
		await expect(this.page.locator('text=Choose Your Plan')).toBeVisible();
	}

	async expectIdentityStep() {
		await expect(this.page.locator('text=Your Nostr Identity, text=Enter your')).toBeVisible();
	}

	async expectPaymentStep() {
		await expect(this.page.locator('text=Pay with Lightning')).toBeVisible();
	}

	async expectPaymentConfirmed() {
		await expect(this.successMessage).toBeVisible({ timeout: 10000 });
	}
}
