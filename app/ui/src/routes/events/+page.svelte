<script>
	import { onMount } from 'svelte';
	import { get } from '$lib/api';

	let events = $state([]);
	let loading = $state(true);
	let error = $state(null);

	// Filters
	let kindFilter = $state('');
	let searchQuery = $state('');

	const kindNames = {
		0: 'Metadata',
		1: 'Note',
		3: 'Contacts',
		4: 'DM',
		5: 'Deletion',
		6: 'Repost',
		7: 'Reaction',
		10002: 'Relay List'
	};

	onMount(async () => {
		await loadEvents();
	});

	async function loadEvents() {
		loading = true;
		error = null;
		try {
			const params = new URLSearchParams();
			if (kindFilter) params.set('kinds', kindFilter);
			if (searchQuery) params.set('search', searchQuery);

			const res = await get(`/events?${params}`);
			events = res.events || [];
		} catch (e) {
			error = e.message;
		} finally {
			loading = false;
		}
	}

	function formatDate(dateStr) {
		return new Date(dateStr).toLocaleString();
	}

	function getKindName(kind) {
		return kindNames[kind] || `Kind ${kind}`;
	}
</script>

<div class="space-y-6">
	<div>
		<h1 class="text-2xl font-bold text-gray-900">Event Browser</h1>
		<p class="text-gray-600">Explore events stored on your relay</p>
	</div>

	<!-- Filters -->
	<div class="flex flex-wrap gap-4 rounded-lg bg-white p-4 shadow">
		<input
			type="text"
			placeholder="Search content..."
			bind:value={searchQuery}
			class="flex-1 rounded-lg border border-gray-300 px-4 py-2 focus:border-purple-500 focus:outline-none"
		/>
		<select
			bind:value={kindFilter}
			class="rounded-lg border border-gray-300 px-4 py-2 focus:border-purple-500 focus:outline-none"
		>
			<option value="">All kinds</option>
			<option value="0">Metadata</option>
			<option value="1">Notes</option>
			<option value="3">Contacts</option>
			<option value="7">Reactions</option>
		</select>
		<button
			onclick={loadEvents}
			class="rounded-lg bg-purple-600 px-4 py-2 text-white hover:bg-purple-700"
		>
			Search
		</button>
	</div>

	{#if loading}
		<div class="flex items-center justify-center py-12">
			<div class="h-8 w-8 animate-spin rounded-full border-4 border-purple-600 border-t-transparent"></div>
		</div>
	{:else if error}
		<div class="rounded-lg bg-red-50 p-4 text-red-700">
			<p class="font-medium">Error loading events</p>
			<p class="text-sm">{error}</p>
		</div>
	{:else if events.length === 0}
		<div class="rounded-lg bg-gray-50 p-8 text-center">
			<p class="text-gray-500">No events found</p>
		</div>
	{:else}
		<div class="space-y-3">
			{#each events as event}
				<div class="rounded-lg bg-white p-4 shadow">
					<div class="flex items-start justify-between">
						<div>
							<span class="rounded bg-gray-100 px-2 py-1 text-xs font-medium text-gray-600">
								{getKindName(event.kind)}
							</span>
							<p class="mt-2 text-sm text-gray-500 font-mono">
								{event.pubkey?.slice(0, 16)}...
							</p>
						</div>
						<p class="text-sm text-gray-400">{formatDate(event.created_at)}</p>
					</div>
					{#if event.content}
						<p class="mt-3 text-gray-700 line-clamp-3">{event.content}</p>
					{/if}
					<div class="mt-3 flex gap-2">
						<button class="text-sm text-purple-600 hover:text-purple-700">View Details</button>
					</div>
				</div>
			{/each}
		</div>
	{/if}
</div>
