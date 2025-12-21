import { describe, it, expect } from 'vitest';
import { render, screen } from '@testing-library/svelte';
import Loading from './Loading.svelte';

describe('Loading', () => {
	it('renders spinner', () => {
		render(Loading);
		// The spinner is a div with animate-spin class
		const spinner = document.querySelector('.animate-spin');
		expect(spinner).toBeTruthy();
	});

	it('shows text when provided', () => {
		render(Loading, { props: { text: 'Loading data...' } });
		expect(screen.getByText('Loading data...')).toBeTruthy();
	});

	it('hides text when not provided', () => {
		render(Loading);
		expect(screen.queryByText('Loading')).toBeNull();
	});

	it('applies size classes', () => {
		const { container } = render(Loading, { props: { size: 'lg' } });
		const spinner = container.querySelector('.animate-spin');
		expect(spinner.classList.contains('h-12')).toBe(true);
		expect(spinner.classList.contains('w-12')).toBe(true);
	});

	it('uses medium size by default', () => {
		const { container } = render(Loading);
		const spinner = container.querySelector('.animate-spin');
		expect(spinner.classList.contains('h-8')).toBe(true);
		expect(spinner.classList.contains('w-8')).toBe(true);
	});
});
