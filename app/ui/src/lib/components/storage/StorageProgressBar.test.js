import { describe, it, expect } from 'vitest';
import { render, screen } from '@testing-library/svelte';
import StorageProgressBar from './StorageProgressBar.svelte';

describe('StorageProgressBar', () => {
	it('renders with default props', () => {
		const { container } = render(StorageProgressBar);
		// Should have the progress bar container
		const progressBar = container.querySelector('.rounded-full.bg-gray-200');
		expect(progressBar).toBeTruthy();
	});

	it('shows usage label by default', () => {
		render(StorageProgressBar, {
			props: {
				usedBytes: 1073741824, // 1 GB
				totalBytes: 10737418240 // 10 GB
			}
		});
		expect(screen.getByText(/1\.0 GB used of 10\.0 GB available/)).toBeTruthy();
	});

	it('hides label when showLabel is false', () => {
		render(StorageProgressBar, {
			props: {
				usedBytes: 1073741824,
				totalBytes: 10737418240,
				showLabel: false
			}
		});
		expect(screen.queryByText(/used of/)).toBeNull();
	});

	it('shows healthy status with green bar', () => {
		const { container } = render(StorageProgressBar, {
			props: {
				usedBytes: 1073741824,
				totalBytes: 10737418240,
				status: 'healthy'
			}
		});
		expect(screen.getByText(/Healthy/)).toBeTruthy();
		const greenBar = container.querySelector('.bg-green-500');
		expect(greenBar).toBeTruthy();
	});

	it('shows warning status with yellow bar', () => {
		const { container } = render(StorageProgressBar, {
			props: {
				usedBytes: 8589934592, // 8 GB
				totalBytes: 10737418240, // 10 GB
				status: 'warning'
			}
		});
		expect(screen.getByText(/Storage filling up/)).toBeTruthy();
		const yellowBar = container.querySelector('.bg-yellow-500');
		expect(yellowBar).toBeTruthy();
	});

	it('shows critical status with red bar', () => {
		const { container } = render(StorageProgressBar, {
			props: {
				usedBytes: 10200000000,
				totalBytes: 10737418240,
				status: 'critical'
			}
		});
		expect(screen.getByText(/Critical/)).toBeTruthy();
		const redBar = container.querySelector('.bg-red-500');
		expect(redBar).toBeTruthy();
	});

	it('shows low storage status with orange bar', () => {
		const { container } = render(StorageProgressBar, {
			props: {
				usedBytes: 7000000000,
				totalBytes: 10737418240,
				status: 'low'
			}
		});
		expect(screen.getByText(/Low storage/)).toBeTruthy();
		const orangeBar = container.querySelector('.bg-orange-500');
		expect(orangeBar).toBeTruthy();
	});

	it('calculates percentage correctly', () => {
		render(StorageProgressBar, {
			props: {
				usedBytes: 5368709120, // 5 GB
				totalBytes: 10737418240 // 10 GB
			}
		});
		expect(screen.getByText(/50\.0%/)).toBeTruthy();
	});

	it('handles zero total bytes', () => {
		render(StorageProgressBar, {
			props: {
				usedBytes: 0,
				totalBytes: 0
			}
		});
		expect(screen.getByText(/0\.0%/)).toBeTruthy();
	});

	it('applies small size class', () => {
		const { container } = render(StorageProgressBar, {
			props: {
				usedBytes: 1000000,
				totalBytes: 10000000,
				size: 'sm'
			}
		});
		const smallBar = container.querySelector('.h-2');
		expect(smallBar).toBeTruthy();
	});

	it('applies large size class', () => {
		const { container } = render(StorageProgressBar, {
			props: {
				usedBytes: 1000000,
				totalBytes: 10000000,
				size: 'lg'
			}
		});
		const largeBar = container.querySelector('.h-6');
		expect(largeBar).toBeTruthy();
	});
});
