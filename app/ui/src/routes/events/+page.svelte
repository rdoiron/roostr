<script>
	import { browser } from '$app/environment';
	import { page } from '$app/stores';
	import { goto } from '$app/navigation';
	import { events, setup, access } from '$lib/api/client.js';
	import { notify } from '$lib/stores/app.svelte.js';
	import Loading from '$lib/components/Loading.svelte';
	import Error from '$lib/components/Error.svelte';
	import Empty from '$lib/components/Empty.svelte';
	import Button from '$lib/components/Button.svelte';
	import { EventCard, EventDetailModal, DeleteEventModal } from '$lib/components/events';

	// Event list state
	let eventList = $state([]);
	let loading = $state(true);
	let error = $state(null);

	// Pagination
	let offset = $state(0);
	let limit = $state(50);

	// Filters
	let kindFilter = $state('');
	let authorFilter = $state('');
	let startDate = $state('');
	let endDate = $state('');
	let mentionsMe = $state(false);
	let searchQuery = $state('');

	// Modal state
	let selectedEvent = $state(null);
	let showDetailModal = $state(false);
	let showDeleteModal = $state(false);

	// Author dropdown data
	let operatorPubkey = $state('');
	let whitelist = $state([]);

	// Kind options for dropdown
	const kindOptions = [
		{ value: '', label: 'All Kinds' },
		{ value: '1', label: 'Notes (kind 1)' },
		{ value: '7', label: 'Reactions (kind 7)' },
		{ value: '3', label: 'Follows (kind 3)' },
		{ value: '4,14', label: 'DMs (kind 4, 14)' },
		{ value: '0', label: 'Metadata (kind 0)' },
		{ value: '6', label: 'Reposts (kind 6)' },
		{ value: '10002', label: 'Relay Lists (kind 10002)' }
	];
	// Track if we've initialized
	let initialized = $state(false);

	$effect(() => {
		if (browser && !initialized) {
			initialized = true;
			initPage();
		}
	});

	async function initPage() {
		// Load operator pubkey and whitelist for filters (non-blocking)
		Promise.all([
			setup.getStatus(),
			access.getWhitelist()
		]).then(([setupRes, whitelistRes]) => {
			operatorPubkey = setupRes.operator_pubkey || '';
			whitelist = whitelistRes.entries || [];
		}).catch(() => {
			// Non-fatal, just won't have author dropdown populated
		});

		// Check for deep link
		const urlEventId = $page.url.searchParams.get('id');
		if (urlEventId) {
			await loadEventById(urlEventId);
		}

		await loadEvents();
	}

	async function loadEventById(id) {
		try {
			const event = await events.get(id);
			if (event) {
				selectedEvent = event;
				showDetailModal = true;
			}
		} catch {
			notify('error', 'Event not found');
			// Remove invalid id from URL
			goto('/events', { replaceState: true });
		}
	}

	async function loadEvents() {
		loading = true;
		error = null;

		try {
			const params = {
				limit: limit.toString(),
				offset: offset.toString()
			};

			if (kindFilter) params.kinds = kindFilter;
			if (authorFilter) params.authors = authorFilter;
			if (searchQuery) params.search = searchQuery;

			// Date range (convert to unix timestamps)
			if (startDate) {
				const since = Math.floor(new Date(startDate).getTime() / 1000);
				params.since = since.toString();
			}
			if (endDate) {
				// End of day
				const until = Math.floor(new Date(endDate + 'T23:59:59').getTime() / 1000);
				params.until = until.toString();
			}

			// Mentions filter
			if (mentionsMe && operatorPubkey) {
				params.mentions = operatorPubkey;
			}

			const res = await events.list(params);
			eventList = res.events || [];
		} catch (e) {
			error = e.message || 'Failed to load events';
		} finally {
			loading = false;
		}
	}

	function applyFilters() {
		offset = 0; // Reset pagination when filters change
		loadEvents();
	}

	function clearFilters() {
		kindFilter = '';
		authorFilter = '';
		startDate = '';
		endDate = '';
		mentionsMe = false;
		searchQuery = '';
		offset = 0;
		loadEvents();
	}

	function prevPage() {
		if (offset > 0) {
			offset = Math.max(0, offset - limit);
			loadEvents();
		}
	}

	function nextPage() {
		if (eventList.length === limit) {
			offset += limit;
			loadEvents();
		}
	}

	function handleViewRaw(event) {
		selectedEvent = event;
		showDetailModal = true;
		// Update URL for deep linking
		goto(`/events?id=${event.id}`, { replaceState: true });
	}

	function handleDelete(event) {
		selectedEvent = event;
		showDeleteModal = true;
	}

	function handleCloseDetailModal() {
		showDetailModal = false;
		selectedEvent = null;
		// Remove id from URL
		goto('/events', { replaceState: true });
	}

	function handleCloseDeleteModal() {
		showDeleteModal = false;
		// Don't clear selectedEvent in case detail modal is still open
	}

	function handleDeleteConfirm() {
		// Refresh the list after deletion
		loadEvents();
	}

	function handleDeleteFromDetail(event) {
		showDetailModal = false;
		selectedEvent = event;
		showDeleteModal = true;
	}

	// Computed values
	const hasFilters = $derived(
		kindFilter || authorFilter || startDate || endDate || mentionsMe || searchQuery
	);
	const showingStart = $derived(offset + 1);
	const showingEnd = $derived(offset + eventList.length);
	const hasPrev = $derived(offset > 0);
	const hasNext = $derived(eventList.length === limit);

	// Author options for dropdown
	const authorOptions = $derived(() => {
		const options = [{ value: '', label: 'All Authors' }];

		// Add operator
		if (operatorPubkey) {
			options.push({
				value: operatorPubkey,
				label: `You (${operatorPubkey.slice(0, 8)}...)`
			});
		}

		// Add whitelist entries
		for (const entry of whitelist) {
			if (entry.pubkey !== operatorPubkey) {
				const label = entry.nickname || `${entry.pubkey.slice(0, 8)}...`;
				options.push({ value: entry.pubkey, label });
			}
		}

		return options;
	});
</script>

<div class="space-y-6">
	<div>
		<h1 class="text-2xl font-bold text-gray-900 dark:text-gray-100">Event Browser</h1>
		<p class="text-gray-600 dark:text-gray-400">Explore and manage events stored on your relay</p>
	</div>

	<!-- Filters -->
	<div class="rounded-lg bg-white dark:bg-gray-800 p-4 shadow dark:shadow-gray-900/50">
		<div class="mb-4 flex items-center justify-between">
			<h2 class="font-medium text-gray-900 dark:text-gray-100">Filters</h2>
			{#if hasFilters}
				<button
					type="button"
					onclick={clearFilters}
					class="text-sm text-gray-500 dark:text-gray-400 hover:text-gray-700 dark:hover:text-gray-300"
				>
					Clear all
				</button>
			{/if}
		</div>

		<div class="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
			<!-- Kind filter -->
			<div>
				<label for="kind-filter" class="mb-1 block text-sm font-medium text-gray-700 dark:text-gray-200">Kind</label>
				<select
					id="kind-filter"
					bind:value={kindFilter}
					class="input w-full"
				>
					{#each kindOptions as option}
						<option value={option.value}>{option.label}</option>
					{/each}
				</select>
			</div>

			<!-- Author filter -->
			<div>
				<label for="author-filter" class="mb-1 block text-sm font-medium text-gray-700 dark:text-gray-200">Author</label>
				<select
					id="author-filter"
					bind:value={authorFilter}
					class="input w-full"
				>
					{#each authorOptions() as option}
						<option value={option.value}>{option.label}</option>
					{/each}
				</select>
			</div>

			<!-- Start date -->
			<div>
				<label for="start-date" class="mb-1 block text-sm font-medium text-gray-700 dark:text-gray-200">From Date</label>
				<input
					id="start-date"
					type="date"
					bind:value={startDate}
					class="input w-full"
				/>
			</div>

			<!-- End date -->
			<div>
				<label for="end-date" class="mb-1 block text-sm font-medium text-gray-700 dark:text-gray-200">To Date</label>
				<input
					id="end-date"
					type="date"
					bind:value={endDate}
					class="input w-full"
				/>
			</div>
		</div>

		<div class="mt-4 flex flex-wrap items-center gap-4">
			<!-- Search -->
			<div class="flex-1">
				<input
					type="text"
					placeholder="Search content..."
					bind:value={searchQuery}
					class="input w-full"
				/>
			</div>

			<!-- Mentions me checkbox -->
			<label class="flex cursor-pointer items-center gap-2">
				<input
					type="checkbox"
					bind:checked={mentionsMe}
					class="h-4 w-4 rounded border-gray-300 dark:border-gray-600 text-purple-600 focus:ring-purple-500"
				/>
				<span class="text-sm text-gray-700 dark:text-gray-200">Only events mentioning me</span>
			</label>

			<!-- Apply button -->
			<Button onclick={applyFilters}>
				Apply Filters
			</Button>
		</div>
	</div>

	<!-- Results header -->
	{#if !loading && !error}
		<div class="flex items-center justify-between">
			<p class="text-sm text-gray-600 dark:text-gray-400">
				{#if eventList.length > 0}
					Showing {showingStart}-{showingEnd} events
				{:else}
					No events found
				{/if}
			</p>

			{#if eventList.length > 0}
				<div class="flex gap-2">
					<Button variant="secondary" onclick={prevPage} disabled={!hasPrev}>
						<svg class="mr-1 h-4 w-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7" />
						</svg>
						Prev
					</Button>
					<Button variant="secondary" onclick={nextPage} disabled={!hasNext}>
						Next
						<svg class="ml-1 h-4 w-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" />
						</svg>
					</Button>
				</div>
			{/if}
		</div>
	{/if}

	<!-- Event list -->
	{#if loading}
		<Loading text="Loading events..." />
	{:else if error}
		<Error message={error} onRetry={loadEvents} />
	{:else if eventList.length === 0}
		<Empty
			title="No events found"
			message={hasFilters ? 'Try adjusting your filters' : 'Your relay has no events yet'}
		/>
	{:else}
		<div class="space-y-3">
			{#each eventList as event (event.id)}
				<EventCard
					{event}
					onViewRaw={handleViewRaw}
					onDelete={handleDelete}
				/>
			{/each}
		</div>

		<!-- Bottom pagination -->
		{#if eventList.length > 0}
			<div class="flex justify-center gap-2 pt-4">
				<Button variant="secondary" onclick={prevPage} disabled={!hasPrev}>
					<svg class="mr-1 h-4 w-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7" />
					</svg>
					Previous
				</Button>
				<Button variant="secondary" onclick={nextPage} disabled={!hasNext}>
					Next
					<svg class="ml-1 h-4 w-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" />
					</svg>
				</Button>
			</div>
		{/if}
	{/if}
</div>

<!-- Detail Modal -->
{#if showDetailModal && selectedEvent}
	<EventDetailModal
		event={selectedEvent}
		onClose={handleCloseDetailModal}
		onDelete={handleDeleteFromDetail}
	/>
{/if}

<!-- Delete Confirmation Modal -->
{#if showDeleteModal && selectedEvent}
	<DeleteEventModal
		event={selectedEvent}
		onClose={handleCloseDeleteModal}
		onConfirm={handleDeleteConfirm}
	/>
{/if}
