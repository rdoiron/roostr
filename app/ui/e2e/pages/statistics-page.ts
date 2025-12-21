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
		this.timeRangeSelector = page.locator('select').first();
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
		await this.timeRangeSelector.selectOption(range);
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
