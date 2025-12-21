import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/svelte';
import Error from './Error.svelte';

describe('Error', () => {
	it('renders with default title', () => {
		render(Error);
		expect(screen.getByText('Error')).toBeTruthy();
	});

	it('renders custom title', () => {
		render(Error, { props: { title: 'Connection Failed' } });
		expect(screen.getByText('Connection Failed')).toBeTruthy();
	});

	it('shows message when provided', () => {
		render(Error, { props: { message: 'Could not connect to server' } });
		expect(screen.getByText('Could not connect to server')).toBeTruthy();
	});

	it('shows error code when provided', () => {
		render(Error, { props: { code: 'ERR_NETWORK' } });
		expect(screen.getByText('ERR_NETWORK')).toBeTruthy();
	});

	it('shows retry button when onRetry provided', async () => {
		const onRetry = vi.fn();
		render(Error, { props: { onRetry } });

		const retryButton = screen.getByText('Try again');
		expect(retryButton).toBeTruthy();

		await fireEvent.click(retryButton);
		expect(onRetry).toHaveBeenCalledOnce();
	});

	it('hides retry button when onRetry is null', () => {
		render(Error, { props: { onRetry: null } });
		expect(screen.queryByText('Try again')).toBeNull();
	});

	it('renders all props together', () => {
		render(Error, {
			props: {
				title: 'Server Error',
				message: 'Internal server error',
				code: '500'
			}
		});
		expect(screen.getByText('Server Error')).toBeTruthy();
		expect(screen.getByText('Internal server error')).toBeTruthy();
		expect(screen.getByText('500')).toBeTruthy();
	});
});
