<script>
	import { onMount } from 'svelte';
	import { Chart, LineController, LineElement, PointElement, LinearScale, CategoryScale, Tooltip, Filler } from 'chart.js';

	let { data = [], total = 0 } = $props();

	let canvas = $state();
	let chart;
	let registered = false;

	function formatDate(dateStr) {
		const date = new Date(dateStr);
		return date.toLocaleDateString('en-US', { month: 'short', day: 'numeric' });
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
							color: '#e5e7eb',
							drawBorder: false
						},
						ticks: {
							color: '#6b7280'
						}
					},
					y: {
						beginAtZero: true,
						grid: {
							color: '#e5e7eb',
							drawBorder: false
						},
						ticks: {
							color: '#6b7280',
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
		if (canvas && data?.length > 0) {
			createChart();
		}
	});
</script>

<div class="rounded-lg bg-white p-6 shadow">
	<div class="mb-4 flex items-center justify-between">
		<h3 class="text-lg font-semibold text-gray-900">Events Over Time</h3>
		<span class="text-sm text-gray-500">{total.toLocaleString()} total</span>
	</div>
	<div class="h-64">
		{#if data?.length > 0}
			<canvas bind:this={canvas}></canvas>
		{:else}
			<div class="flex h-full items-center justify-center text-gray-500">No data available</div>
		{/if}
	</div>
</div>
