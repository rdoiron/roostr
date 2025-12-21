/**
 * Statistics page object.
 */

import { Page, Locator, expect } from '@playwright/test';
import { BasePage } from './base-page';

export class StatisticsPage extends BasePage {
	readonly timeRangeSelector: Locator;
	readonly eventsOverTimeChart: Locator;
	readonly eventsByKindChart: Locator;
	readonly topAuthorsSection: Locator;
	readonly chartCanvas: Locator;

	constructor(page: Page) {
		super(page);
		// TimeRangeSelector is a button group, not a select dropdown
		this.timeRangeSelector = page.locator('.flex.gap-1.rounded-lg').first();
		this.eventsOverTimeChart = page.locator('text=Events Over Time').locator('..');
		this.eventsByKindChart = page.locator('text=Events by Kind').locator('..');
		this.topAuthorsSection = page.locator('text=Top Authors').locator('..');
		this.chartCanvas = page.locator('canvas');
	}

	async goto() {
		await this.page.goto('/statistics');
		await this.waitForPageLoad();
	}

	async selectTimeRange(range: string) {
		// TimeRangeSelector uses buttons, not a select dropdown
		const labels: Record<string, string> = {
			today: 'Today',
			'7days': '7 Days',
			'30days': '30 Days',
			alltime: 'All Time'
		};
		const label = labels[range] || range;
		await this.page.getByRole('button', { name: label }).click();
	}

	async hasEventsOverTimeChart(): Promise<boolean> {
		return this.eventsOverTimeChart.isVisible();
	}

	async hasEventsByKindChart(): Promise<boolean> {
		return this.eventsByKindChart.isVisible();
	}

	async hasTopAuthors(): Promise<boolean> {
		return this.topAuthorsSection.isVisible();
	}

	async getChartCount(): Promise<number> {
		return this.chartCanvas.count();
	}

	async expectCharts() {
		await expect(this.page.locator('text=Events Over Time')).toBeVisible();
		await expect(this.page.locator('text=Events by Kind')).toBeVisible();
	}

	async expectTopAuthors() {
		await expect(this.page.locator('text=Top Authors')).toBeVisible();
	}
}
