import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/svelte';
import Empty from './Empty.svelte';

describe('Empty', () => {
	it('renders with default props', () => {
		render(Empty);
		expect(screen.getByText('📭')).toBeTruthy();
		expect(screen.getByText('No data')).toBeTruthy();
	});

	it('renders custom icon', () => {
		render(Empty, { props: { icon: '🔍' } });
		expect(screen.getByText('🔍')).toBeTruthy();
	});

	it('renders custom title', () => {
		render(Empty, { props: { title: 'No events found' } });
		expect(screen.getByText('No events found')).toBeTruthy();
	});

	it('shows message when provided', () => {
		render(Empty, { props: { message: 'Try adjusting your filters' } });
		expect(screen.getByText('Try adjusting your filters')).toBeTruthy();
	});

	it('hides message when not provided', () => {
		const { container } = render(Empty);
		const paragraphs = container.querySelectorAll('p');
		expect(paragraphs.length).toBe(0);
	});

	it('shows action button when action provided', async () => {
		const action = vi.fn();
		render(Empty, {
			props: {
				action,
				actionLabel: 'Add Event'
			}
		});

		const button = screen.getByText('Add Event');
		expect(button).toBeTruthy();

		await fireEvent.click(button);
		expect(action).toHaveBeenCalledOnce();
	});

	it('hides action button when action is null', () => {
		render(Empty, { props: { actionLabel: 'Add Event' } });
		expect(screen.queryByText('Add Event')).toBeNull();
	});

	it('renders full state with all props', () => {
		const action = vi.fn();
		render(Empty, {
			props: {
				icon: '📝',
				title: 'No posts yet',
				message: 'Start by creating your first post',
				action,
				actionLabel: 'Create Post'
			}
		});

		expect(screen.getByText('📝')).toBeTruthy();
		expect(screen.getByText('No posts yet')).toBeTruthy();
		expect(screen.getByText('Start by creating your first post')).toBeTruthy();
		expect(screen.getByText('Create Post')).toBeTruthy();
	});
});
