<script>
	import { formatBytes } from '$lib/utils/format.js';

	/**
	 * @type {{ usedBytes: number, totalBytes: number, status?: string, showLabel?: boolean, size?: 'sm' | 'md' | 'lg' }}
	 */
	let { usedBytes = 0, totalBytes = 0, status = 'healthy', showLabel = true, size = 'md' } = $props();

	// Calculate percentage (avoid division by zero)
	// Use minimum 0.5% width when there's any data so the bar is visible
	let rawPercentage = $derived(totalBytes > 0 ? Math.min((usedBytes / totalBytes) * 100, 100) : 0);
	let percentage = $derived(usedBytes > 0 && rawPercentage < 0.5 ? 0.5 : rawPercentage);

	// Determine color based on status
	let barColor = $derived(() => {
		switch (status) {
			case 'critical':
				return 'bg-red-500';
			case 'low':
				return 'bg-orange-500';
			case 'warning':
				return 'bg-yellow-500';
			default:
				return 'bg-green-500';
		}
	});

	// Status text and color
	let statusInfo = $derived(() => {
		switch (status) {
			case 'critical':
				return { text: 'Critical - relay may stop accepting events', color: 'text-red-600' };
			case 'low':
				return { text: 'Low storage - consider cleanup', color: 'text-orange-600' };
			case 'warning':
				return { text: 'Storage filling up', color: 'text-yellow-600' };
			default:
				return { text: 'Healthy - plenty of space available', color: 'text-green-600' };
		}
	});

	// Bar height based on size
	let barHeight = $derived(() => {
		switch (size) {
			case 'sm':
				return 'h-2';
			case 'lg':
				return 'h-6';
			default:
				return 'h-4';
		}
	});
</script>

<div class="w-full">
	<!-- Progress bar -->
	<div class="w-full rounded-full bg-gray-200 dark:bg-gray-700 {barHeight()}">
		<div
			class="rounded-full transition-all duration-300 {barHeight()} {barColor()}"
			style="width: {percentage}%"
		></div>
	</div>

	{#if showLabel}
		<!-- Usage label -->
		<div class="mt-2 flex flex-col gap-1 text-sm sm:flex-row sm:items-center sm:justify-between">
			<span class="text-gray-600 dark:text-gray-400">
				{formatBytes(usedBytes)} used of {formatBytes(totalBytes)} available
				<span class="ml-1 text-gray-400 dark:text-gray-500">({rawPercentage.toFixed(1)}%)</span>
			</span>
			<span class="text-right {statusInfo().color}">
				{statusInfo().text}
			</span>
		</div>
	{/if}
</div>
