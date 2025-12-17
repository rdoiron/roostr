<script>
	import { onMount } from 'svelte';
	import { Chart, BarController, BarElement, LinearScale, CategoryScale, Tooltip } from 'chart.js';

	let { kinds = [], total = 0 } = $props();

	let canvas = $state();
	let chart;
	let registered = false;

	const colors = [
		'#9333ea', // purple-600
		'#a855f7', // purple-500
		'#c084fc', // purple-400
		'#d8b4fe', // purple-300
		'#e9d5ff', // purple-200
		'#f3e8ff' // purple-100
	];

	function createChart() {
		if (!registered) {
			Chart.register(BarController, BarElement, LinearScale, CategoryScale, Tooltip);
			registered = true;
		}

		if (chart) {
			chart.destroy();
		}

		const labels = kinds.map((k) => k.label.charAt(0).toUpperCase() + k.label.slice(1));
		const values = kinds.map((k) => k.count);

		chart = new Chart(canvas, {
			type: 'bar',
			data: {
				labels,
				datasets: [
					{
						label: 'Events',
						data: values,
						backgroundColor: kinds.map((_, i) => colors[i % colors.length]),
						borderRadius: 4,
						barThickness: 24
					}
				]
			},
			options: {
				indexAxis: 'y',
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
							label: (context) => {
								const kind = kinds[context.dataIndex];
								return `${context.parsed.x.toLocaleString()} events (${kind.percent.toFixed(1)}%)`;
							}
						}
					}
				},
				scales: {
					x: {
						beginAtZero: true,
						grid: {
							color: '#e5e7eb',
							drawBorder: false
						},
						ticks: {
							color: '#6b7280',
							callback: (value) => value.toLocaleString()
						}
					},
					y: {
						grid: {
							display: false
						},
						ticks: {
							color: '#374151'
						}
					}
				}
			}
		});
	}

	onMount(() => {
		if (kinds?.length > 0) {
			createChart();
		}
		return () => {
			if (chart) {
				chart.destroy();
			}
		};
	});

	$effect(() => {
		if (canvas && kinds?.length > 0) {
			createChart();
		}
	});
</script>

<div class="rounded-lg bg-white p-6 shadow">
	<div class="mb-4 flex items-center justify-between">
		<h3 class="text-lg font-semibold text-gray-900">Events by Kind</h3>
		<span class="text-sm text-gray-500">{total.toLocaleString()} total</span>
	</div>
	<div class="h-64">
		{#if kinds?.length > 0}
			<canvas bind:this={canvas}></canvas>
		{:else}
			<div class="flex h-full items-center justify-center text-gray-500">No data available</div>
		{/if}
	</div>
</div>
