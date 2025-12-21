import { describe, it, expect } from 'vitest';
import { render, screen } from '@testing-library/svelte';
import StatusIndicator from './StatusIndicator.svelte';

describe('StatusIndicator', () => {
	it('shows offline by default', () => {
		render(StatusIndicator);
		expect(screen.getByText('Offline')).toBeTruthy();
	});

	it('shows online status', () => {
		const { container } = render(StatusIndicator, { props: { status: 'online' } });
		expect(screen.getByText('Online')).toBeTruthy();

		// Check for green color class
		const indicator = container.querySelector('.bg-green-500');
		expect(indicator).toBeTruthy();
	});

	it('shows offline status', () => {
		const { container } = render(StatusIndicator, { props: { status: 'offline' } });
		expect(screen.getByText('Offline')).toBeTruthy();

		// Check for red color class
		const indicator = container.querySelector('.bg-red-500');
		expect(indicator).toBeTruthy();
	});

	it('shows degraded status', () => {
		const { container } = render(StatusIndicator, { props: { status: 'degraded' } });
		expect(screen.getByText('Degraded')).toBeTruthy();

		// Check for yellow color class
		const indicator = container.querySelector('.bg-yellow-500');
		expect(indicator).toBeTruthy();
	});

	it('has animated ping indicator', () => {
		const { container } = render(StatusIndicator, { props: { status: 'online' } });
		const pingIndicator = container.querySelector('.animate-ping');
		expect(pingIndicator).toBeTruthy();
	});
});
