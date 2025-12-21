/**
 * Storage management page object.
 */

import { Page, Locator, expect } from '@playwright/test';
import { BasePage } from './base-page';

export class StoragePage extends BasePage {
	readonly usageSection: Locator;
	readonly retentionSection: Locator;
	readonly cleanupSection: Locator;
	readonly maintenanceSection: Locator;
	readonly progressBar: Locator;
	readonly retentionDaysInput: Locator;
	readonly cleanupDateInput: Locator;
	readonly vacuumButton: Locator;
	readonly integrityCheckButton: Locator;
	readonly saveRetentionButton: Locator;
	readonly deleteOldEventsButton: Locator;

	constructor(page: Page) {
		super(page);
		this.usageSection = page.locator('text=Current Usage').locator('..');
		this.retentionSection = page.locator('text=Retention Policy').locator('..');
		this.cleanupSection = page.locator('text=Manual Cleanup').locator('..');
		this.maintenanceSection = page.locator('text=Maintenance').locator('..');
		this.progressBar = page.locator('[role="progressbar"], .bg-purple-600');
		this.retentionDaysInput = page.locator('input[type="number"]').first();
		this.cleanupDateInput = page.locator('input[type="date"]');
		this.vacuumButton = page.getByRole('button', { name: /vacuum/i });
		this.integrityCheckButton = page.getByRole('button', { name: /check integrity/i });
		this.saveRetentionButton = page.getByRole('button', { name: /save/i });
		this.deleteOldEventsButton = page.getByRole('button', { name: /delete.*events/i });
	}

	async goto() {
		await this.page.goto('/storage');
		await this.waitForPageLoad();
	}

	async getUsageText(): Promise<string> {
		const usage = this.page.locator('text=Relay Database').locator('..').locator('span, p');
		return (await usage.first().textContent()) || '';
	}

	async setRetentionDays(days: number) {
		await this.retentionDaysInput.fill(days.toString());
	}

	async saveRetention() {
		await this.saveRetentionButton.click();
	}

	async setCleanupDate(date: string) {
		await this.cleanupDateInput.fill(date);
	}

	async deleteOldEvents() {
		await this.deleteOldEventsButton.click();
	}

	async confirmCleanup() {
		await this.page.getByRole('button', { name: /confirm|delete/i }).last().click();
	}

	async runVacuum() {
		await this.vacuumButton.click();
	}

	async runIntegrityCheck() {
		await this.integrityCheckButton.click();
	}

	async getCleanupEstimate(): Promise<string> {
		const estimate = this.page.locator('text=/\\d+ events/');
		return (await estimate.textContent()) || '';
	}

	async expectUsageDisplay() {
		await expect(this.page.locator('text=Relay Database')).toBeVisible();
	}

	async expectRetentionSettings() {
		await expect(this.page.getByRole('heading', { name: 'Retention Policy' })).toBeVisible();
	}

	async expectMaintenanceTools() {
		// The section has heading "Database Maintenance" and a button "Run VACUUM"
		await expect(this.page.getByRole('heading', { name: 'Database Maintenance' })).toBeVisible();
		await expect(this.page.getByRole('button', { name: /vacuum/i })).toBeVisible();
	}

	async expectVacuumSuccess() {
		await expect(this.page.locator('text=/vacuum.*complete|reclaimed/i')).toBeVisible();
	}

	async expectIntegrityCheckPassed() {
		// Check for the specific result message showing all databases passed
		await expect(this.page.getByText('All databases passed integrity check')).toBeVisible();
	}
}
