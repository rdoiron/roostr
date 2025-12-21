/**
 * Events browser page object.
 */

import { Page, Locator, expect } from '@playwright/test';
import { BasePage } from './base-page';

export class EventsPage extends BasePage {
	readonly eventList: Locator;
	readonly kindFilter: Locator;
	readonly authorFilter: Locator;
	readonly searchInput: Locator;
	readonly startDateInput: Locator;
	readonly endDateInput: Locator;
	readonly applyFiltersButton: Locator;
	readonly clearFiltersButton: Locator;
	readonly pagination: Locator;
	readonly prevButton: Locator;
	readonly nextButton: Locator;
	readonly eventDetailModal: Locator;

	constructor(page: Page) {
		super(page);
		this.eventList = page.locator('[data-testid="event-list"], .space-y-4');
		this.kindFilter = page.locator('select#kind-filter');
		this.authorFilter = page.locator('select#author-filter');
		this.searchInput = page.locator('input[placeholder*="Search content"]');
		this.startDateInput = page.locator('input#start-date');
		this.endDateInput = page.locator('input#end-date');
		this.applyFiltersButton = page.getByRole('button', { name: /apply filters/i });
		this.clearFiltersButton = page.locator('text=Clear all');
		this.pagination = page.locator('text=Showing');
		this.prevButton = page.getByRole('button', { name: /prev/i }).first();
		this.nextButton = page.getByRole('button', { name: /^next$/i }).first();
		this.eventDetailModal = page.locator('[role="dialog"]');
	}

	async goto() {
		await this.page.goto('/events');
		await this.waitForPageLoad();
	}

	async gotoWithEventId(eventId: string) {
		await this.page.goto(`/events?id=${eventId}`);
		await this.waitForPageLoad();
	}

	async filterByKind(kind: string) {
		await this.kindFilter.selectOption(kind);
	}

	async filterByAuthor(author: string) {
		// Author filter is a select dropdown
		await this.authorFilter.selectOption(author);
	}

	async filterBySearch(query: string) {
		await this.searchInput.fill(query);
	}

	async filterByDateRange(start: string, end: string) {
		await this.startDateInput.fill(start);
		await this.endDateInput.fill(end);
	}

	async applyFilters() {
		await this.applyFiltersButton.click();
	}

	async clearFilters() {
		await this.clearFiltersButton.click();
	}

	async goToNextPage() {
		await this.nextButton.click();
	}

	async goToPrevPage() {
		await this.prevButton.click();
	}

	async getEventCount(): Promise<number> {
		const cards = this.page.locator('[data-testid="event-card"], .bg-white.rounded-lg.shadow').filter({
			has: this.page.locator('text=/kind \\d+|Note|Reaction|Repost|DM/i')
		});
		return cards.count();
	}

	async clickViewRaw(index: number = 0) {
		const viewButtons = this.page.locator('button').filter({ hasText: /view|raw|json/i });
		await viewButtons.nth(index).click();
	}

	async clickDeleteEvent(index: number = 0) {
		const deleteButtons = this.page.locator('button').filter({ hasText: /delete/i });
		await deleteButtons.nth(index).click();
	}

	async confirmDelete() {
		await this.page.getByRole('button', { name: /delete|confirm/i }).last().click();
	}

	async isEventDetailModalOpen(): Promise<boolean> {
		return this.eventDetailModal.isVisible();
	}

	async closeEventDetailModal() {
		await this.page.keyboard.press('Escape');
	}

	async getPaginationText(): Promise<string> {
		return (await this.pagination.textContent()) || '';
	}

	async expectEventList() {
		await expect(this.page.getByRole('heading', { name: 'Event Browser' })).toBeVisible();
	}

	async expectNoEvents() {
		await expect(this.page.getByRole('heading', { name: 'No events found' })).toBeVisible();
	}
}
