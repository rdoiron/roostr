import { describe, it, expect } from 'vitest';
import { render, screen } from '@testing-library/svelte';
import StatCard from './StatCard.svelte';

describe('StatCard', () => {
	it('renders label and value', () => {
		render(StatCard, { props: { label: 'Total Events', value: '1,234' } });
		expect(screen.getByText('Total Events')).toBeTruthy();
		expect(screen.getByText('1,234')).toBeTruthy();
	});

	it('shows subtext when provided', () => {
		render(StatCard, {
			props: {
				label: 'Uptime',
				value: '5d 12h',
				subtext: 'Since last restart'
			}
		});
		expect(screen.getByText('Since last restart')).toBeTruthy();
	});

	it('hides subtext when not provided', () => {
		const { container } = render(StatCard, {
			props: { label: 'Events', value: '100' }
		});
		// Only 2 p elements: label and value
		const paragraphs = container.querySelectorAll('p');
		expect(paragraphs.length).toBe(2);
	});

	it('applies custom value class', () => {
		const { container } = render(StatCard, {
			props: {
				label: 'Status',
				value: 'Online',
				valueClass: 'text-green-600'
			}
		});
		const valueElement = container.querySelector('.text-green-600');
		expect(valueElement).toBeTruthy();
		expect(valueElement.textContent).toBe('Online');
	});

	it('sets tooltip when provided', () => {
		const { container } = render(StatCard, {
			props: {
				label: 'Events',
				value: '1.2K',
				tooltip: '1,234 total events'
			}
		});
		const valueElement = container.querySelector('[title="1,234 total events"]');
		expect(valueElement).toBeTruthy();
	});

	it('does not set title attribute when no tooltip', () => {
		const { container } = render(StatCard, {
			props: { label: 'Events', value: '100' }
		});
		const valueElement = container.querySelector('.text-3xl');
		expect(valueElement.hasAttribute('title')).toBe(false);
	});
});
