/**
 * E2E tests for the setup wizard.
 */

import { test, expect } from '../../fixtures/test-fixtures';
import { SetupPage } from '../../pages/setup-page';
import { mockSetupFlow } from '../../fixtures/api-mocks';

test.describe('Setup Wizard', () => {
	test.beforeEach(async ({ page }) => {
		await mockSetupFlow(page);
	});

	test('displays welcome step initially', async ({ page }) => {
		const setupPage = new SetupPage(page);
		await setupPage.goto();

		await expect(page.locator('text=Welcome to Roostr')).toBeVisible();
		await expect(page.getByRole('button', { name: 'Get Started' })).toBeVisible();
	});

	test('progresses through all wizard steps', async ({ page }) => {
		const setupPage = new SetupPage(page);
		await setupPage.goto();

		// Step 0: Welcome
		expect(await setupPage.isOnWelcomeStep()).toBe(true);
		await setupPage.clickGetStarted();

		// Step 1: Identity
		expect(await setupPage.getCurrentStep()).toBe(1);
		await setupPage.enterIdentity('npub1test123abc456');
		await setupPage.clickContinue();

		// Step 2: Relay Info
		expect(await setupPage.getCurrentStep()).toBe(2);
		await setupPage.enterRelayInfo('My Private Relay', 'A secure relay for my notes');
		await setupPage.clickContinue();

		// Step 3: Access Mode
		expect(await setupPage.getCurrentStep()).toBe(3);
		await setupPage.selectAccessMode('private');
		await setupPage.clickContinue();

		// Step 4: Add Others (optional)
		expect(await setupPage.getCurrentStep()).toBe(4);
		await setupPage.clickFinishSetup();

		// Step 5: Complete
		expect(await setupPage.isSetupComplete()).toBe(true);
	});

	test('validates identity input - valid npub', async ({ page }) => {
		const setupPage = new SetupPage(page);
		await setupPage.goto();
		await setupPage.clickGetStarted();

		// Empty input - continue should be disabled
		expect(await setupPage.isContinueEnabled()).toBe(false);

		// Valid npub
		await setupPage.enterIdentity('npub1abc123xyz');
		expect(await setupPage.isIdentityValid()).toBe(true);
		expect(await setupPage.isContinueEnabled()).toBe(true);
	});

	test('validates identity input - invalid format', async ({ page }) => {
		const setupPage = new SetupPage(page);
		await setupPage.goto();
		await setupPage.clickGetStarted();

		await setupPage.enterIdentity('invalid-key');
		// Check for invalid indicator or continue button disabled
		const continueEnabled = await setupPage.isContinueEnabled();
		expect(continueEnabled).toBe(false);
	});

	test('validates identity input - NIP-05 format', async ({ page }) => {
		const setupPage = new SetupPage(page);
		await setupPage.goto();
		await setupPage.clickGetStarted();

		await setupPage.enterIdentity('alice@example.com');
		expect(await setupPage.isIdentityValid()).toBe(true);
		expect(await setupPage.isContinueEnabled()).toBe(true);
	});

	test('allows slow character-by-character typing without input locking', async ({ page }) => {
		const setupPage = new SetupPage(page);
		await setupPage.goto();
		await setupPage.clickGetStarted();

		const input = page.locator('#identity');
		const testValue = 'alice@example.com';

		// Type character by character with delays (simulates slow human typing)
		// This tests the bug where validation firing mid-typing would lock the input
		for (const char of testValue) {
			await input.pressSequentially(char, { delay: 100 });
		}

		// Wait for final validation to complete
		await page.waitForTimeout(700);

		// Verify all characters were accepted (not lost during validation)
		await expect(input).toHaveValue(testValue);

		// Validation should show valid
		expect(await setupPage.isIdentityValid()).toBe(true);
	});

	test('back button navigates to previous step', async ({ page }) => {
		const setupPage = new SetupPage(page);
		await setupPage.goto();
		await setupPage.clickGetStarted();

		// Go to step 2
		await setupPage.enterIdentity('npub1test123');
		await setupPage.clickContinue();
		expect(await setupPage.getCurrentStep()).toBe(2);

		// Go back to step 1
		await setupPage.clickBack();
		expect(await setupPage.getCurrentStep()).toBe(1);
	});

	test('preserves data when navigating back', async ({ page }) => {
		const setupPage = new SetupPage(page);
		await setupPage.goto();
		await setupPage.clickGetStarted();

		// Enter identity and proceed
		const testIdentity = 'npub1preservetest';
		await setupPage.enterIdentity(testIdentity);
		await setupPage.clickContinue();

		// Enter relay info
		await setupPage.enterRelayInfo('Test Relay', 'Test description');

		// Navigate back to identity step
		await setupPage.clickBack();
		await page.waitForTimeout(500);

		// Identity should be preserved (input has id="identity")
		const input = page.locator('#identity');
		await expect(input).toHaveValue(testIdentity);
	});

	test('relay info step validation', async ({ page }) => {
		const setupPage = new SetupPage(page);
		await setupPage.goto();
		await setupPage.clickGetStarted();

		// Complete identity step
		await setupPage.enterIdentity('npub1test123');
		await setupPage.clickContinue();

		// Relay name is required
		expect(await setupPage.isContinueEnabled()).toBe(false);

		await setupPage.enterRelayInfo('My Relay', '');
		expect(await setupPage.isContinueEnabled()).toBe(true);
	});

	test('public access mode can be selected', async ({ page }) => {
		const setupPage = new SetupPage(page);
		await setupPage.goto();
		await setupPage.clickGetStarted();

		// Complete identity and relay info
		await setupPage.enterIdentity('npub1test123');
		await setupPage.clickContinue();
		await setupPage.enterRelayInfo('Test Relay', 'Description');
		await setupPage.clickContinue();

		// Select public mode
		await setupPage.selectAccessMode('public');
		// Verify public button is selected (has purple border)
		const publicButton = page.locator('button:has-text("Public")');
		await expect(publicButton).toHaveClass(/border-purple/);
	});

	test.skip('shows error on setup failure', async ({ page }) => {
		// Skipped: Complex to test error states with mocked API
		// The error display depends on timing of API failure and Svelte re-render
	});
});
