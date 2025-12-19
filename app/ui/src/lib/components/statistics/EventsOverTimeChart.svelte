<script>
	import { onMount } from 'svelte';
	import { Chart, LineController, LineElement, PointElement, LinearScale, CategoryScale, Tooltip, Filler } from 'chart.js';
	import { themeStore } from '$lib/stores/theme.svelte.js';
	import { formatDateInTimezone } from '$lib/stores/timezone.svelte.js';

	let { data = [], total = 0 } = $props();

	function getChartColors() {
		const isDark = themeStore.resolved === 'dark';
		return {
			grid: isDark ? '#374151' : '#e5e7eb',
			ticks: isDark ? '#9ca3af' : '#6b7280'
		};
	}

	let canvas = $state();
	let chart;
	let registered = false;

	function formatDate(dateStr) {
		// Check if this is hourly data (contains space and :00)
		if (dateStr.includes(' ') && dateStr.includes(':')) {
			// Parse hourly format: "2024-12-17 09:00" - already in user's timezone from backend
			const hour = parseInt(dateStr.split(' ')[1].split(':')[0], 10);
			if (hour === 0) return '12 AM';
			if (hour === 12) return '12 PM';
			if (hour < 12) return `${hour} AM`;
			return `${hour - 12} PM`;
		}
		// Daily format: "2024-12-17" - already in user's timezone from backend
		// Parse as local date to avoid timezone shifts
		const [year, month, day] = dateStr.split('-').map(Number);
		const date = new Date(year, month - 1, day);
		return formatDateInTimezone(date, { month: 'short', day: 'numeric' });
	}

	function createChart() {
		if (!registered) {
			Chart.register(LineController, LineElement, PointElement, LinearScale, CategoryScale, Tooltip, Filler);
			registered = true;
		}

		if (chart) {
			chart.destroy();
		}

		const labels = data.map((d) => formatDate(d.date));
		const values = data.map((d) => d.count);

		const colors = getChartColors();

		chart = new Chart(canvas, {
			type: 'line',
			data: {
				labels,
				datasets: [
					{
						label: 'Events',
						data: values,
						borderColor: '#9333ea',
						backgroundColor: 'rgba(147, 51, 234, 0.1)',
						fill: true,
						tension: 0.3,
						pointRadius: 4,
						pointHoverRadius: 6,
						pointBackgroundColor: '#9333ea',
						pointBorderColor: '#ffffff',
						pointBorderWidth: 2
					}
				]
			},
			options: {
				responsive: true,
				maintainAspectRatio: false,
				plugins: {
					tooltip: {
						backgroundColor: '#1f2937',
						titleColor: '#f9fafb',
						bodyColor: '#d1d5db',
						borderColor: '#374151',
						borderWidth: 1,
						padding: 12,
						displayColors: false,
						callbacks: {
							label: (context) => `${context.parsed.y.toLocaleString()} events`
						}
					}
				},
				scales: {
					x: {
						grid: {
							color: colors.grid,
							drawBorder: false
						},
						ticks: {
							color: colors.ticks
						}
					},
					y: {
						beginAtZero: true,
						grid: {
							color: colors.grid,
							drawBorder: false
						},
						ticks: {
							color: colors.ticks,
							callback: (value) => value.toLocaleString()
						}
					}
				}
			}
		});
	}

	onMount(() => {
		if (data?.length > 0) {
			createChart();
		}
		return () => {
			if (chart) {
				chart.destroy();
			}
		};
	});

	$effect(() => {
		// Subscribe to theme changes (accessing resolved triggers re-run on change)
		themeStore.resolved;
		if (canvas && data?.length > 0) {
			createChart();
		}
	});
</script>

<div class="rounded-lg bg-white dark:bg-gray-800 p-6 shadow dark:shadow-gray-900/50">
	<div class="mb-4 flex items-center justify-between">
		<h3 class="text-lg font-semibold text-gray-900 dark:text-gray-100">Events Over Time</h3>
		<span class="text-sm text-gray-500 dark:text-gray-400">{total.toLocaleString()} total</span>
	</div>
	<div class="h-64">
		{#if data?.length > 0}
			<canvas bind:this={canvas}></canvas>
		{:else}
			<div class="flex h-full items-center justify-center text-gray-500 dark:text-gray-400">No data available</div>
		{/if}
	</div>
</div>
