<script>
	import { stats } from '$lib/api/client.js';
	import Loading from '$lib/components/Loading.svelte';
	import Error from '$lib/components/Error.svelte';
	import TimeRangeSelector from '$lib/components/statistics/TimeRangeSelector.svelte';
	import TopAuthorsList from '$lib/components/statistics/TopAuthorsList.svelte';

	let loading = $state(true);
	let error = $state(null);
	let timeRange = $state('7days');
	let initialized = $state(false);

	let eventsOverTime = $state({ data: [], total: 0 });
	let eventsByKind = $state({ kinds: [], total: 0 });
	let topAuthors = $state({ authors: [] });

	// Dynamically imported chart components (Chart.js doesn't work with SSR)
	let EventsOverTimeChart = $state(null);
	let EventsByKindChart = $state(null);

	async function loadData(range) {
		try {
			const [overTimeRes, byKindRes, authorsRes] = await Promise.all([
				stats.getEventsOverTime(range),
				stats.getEventsByKind(range),
				stats.getTopAuthors(range, 10)
			]);

			eventsOverTime = overTimeRes;
			eventsByKind = byKindRes;
			topAuthors = authorsRes;
			error = null;
		} catch (e) {
			error = e.message || 'Failed to load statistics';
		} finally {
			loading = false;
		}
	}

	function handleTimeRangeChange(newRange) {
		timeRange = newRange;
		loading = true;
		loadData(newRange);
	}

	// Use $effect for initialization instead of onMount
	$effect(() => {
		if (!initialized) {
			initialized = true;

			// Dynamically import chart components
			Promise.all([
				import('$lib/components/statistics/EventsOverTimeChart.svelte'),
				import('$lib/components/statistics/EventsByKindChart.svelte')
			]).then(([overTimeModule, byKindModule]) => {
				EventsOverTimeChart = overTimeModule.default;
				EventsByKindChart = byKindModule.default;
			});

			// Load initial data
			loadData('7days');
		}
	});
</script>

<svelte:head>
	<title>Statistics - Roostr</title>
</svelte:head>

<div class="space-y-6">
	<!-- Header -->
	<div class="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
		<div>
			<h1 class="text-2xl font-bold text-gray-900">Statistics</h1>
			<p class="mt-1 text-sm text-gray-600">Detailed insights into your relay activity</p>
		</div>
		<TimeRangeSelector value={timeRange} onchange={handleTimeRangeChange} />
	</div>

	{#if loading}
		<Loading message="Loading statistics..." />
	{:else if error}
		<Error message={error} onRetry={() => loadData(timeRange)} />
	{:else}
		<!-- Events Over Time Chart (full width) -->
		{#if EventsOverTimeChart}
			<EventsOverTimeChart data={eventsOverTime.data ?? []} total={eventsOverTime.total ?? 0} />
		{:else}
			<div class="rounded-lg bg-white p-6 shadow">
				<h3 class="text-lg font-semibold text-gray-900">Events Over Time</h3>
				<p class="text-gray-500">{eventsOverTime.total} total</p>
			</div>
		{/if}

		<!-- Events by Kind and Top Authors (two columns on larger screens) -->
		<div class="grid gap-6 lg:grid-cols-2">
			{#if EventsByKindChart}
				<EventsByKindChart kinds={eventsByKind.kinds ?? []} total={eventsByKind.total ?? 0} />
			{:else}
				<div class="rounded-lg bg-white p-6 shadow">
					<h3 class="text-lg font-semibold text-gray-900">Events by Kind</h3>
					<p class="text-gray-500">{eventsByKind.total} total</p>
				</div>
			{/if}
			<TopAuthorsList authors={topAuthors.authors ?? []} />
		</div>
	{/if}
</div>
