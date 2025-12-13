<script>
	import { onMount } from 'svelte';
	import { get } from '$lib/api';
	import { relayStatus } from '$lib/stores';

	let stats = $state(null);
	let loading = $state(true);
	let error = $state(null);

	onMount(async () => {
		try {
			const [healthRes, statsRes] = await Promise.all([
				get('/health'),
				get('/stats/summary')
			]);

			relayStatus.online = healthRes.status === 'ok';
			relayStatus.loading = false;

			if (statsRes.relay_connected) {
				stats = statsRes;
			}
		} catch (e) {
			error = e.message;
			relayStatus.online = false;
		} finally {
			loading = false;
			relayStatus.loading = false;
		}
	});
</script>

<div class="space-y-6">
	<div>
		<h1 class="text-2xl font-bold text-gray-900">Dashboard</h1>
		<p class="text-gray-600">Your private Nostr relay at a glance</p>
	</div>

	{#if loading}
		<div class="flex items-center justify-center py-12">
			<div class="h-8 w-8 animate-spin rounded-full border-4 border-purple-600 border-t-transparent"></div>
		</div>
	{:else if error}
		<div class="rounded-lg bg-red-50 p-4 text-red-700">
			<p class="font-medium">Error loading dashboard</p>
			<p class="text-sm">{error}</p>
		</div>
	{:else}
		<!-- Stats cards -->
		<div class="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
			<div class="rounded-lg bg-white p-6 shadow">
				<p class="text-sm font-medium text-gray-500">Total Events</p>
				<p class="mt-1 text-3xl font-bold text-gray-900">
					{stats?.total_events?.toLocaleString() ?? '—'}
				</p>
			</div>
			<div class="rounded-lg bg-white p-6 shadow">
				<p class="text-sm font-medium text-gray-500">Unique Authors</p>
				<p class="mt-1 text-3xl font-bold text-gray-900">
					{stats?.total_pubkeys?.toLocaleString() ?? '—'}
				</p>
			</div>
			<div class="rounded-lg bg-white p-6 shadow">
				<p class="text-sm font-medium text-gray-500">Relay Status</p>
				<p class="mt-1 text-3xl font-bold {relayStatus.online ? 'text-green-600' : 'text-red-600'}">
					{relayStatus.online ? 'Online' : 'Offline'}
				</p>
			</div>
			<div class="rounded-lg bg-white p-6 shadow">
				<p class="text-sm font-medium text-gray-500">Database</p>
				<p class="mt-1 text-3xl font-bold {stats?.relay_connected ? 'text-green-600' : 'text-yellow-600'}">
					{stats?.relay_connected ? 'Connected' : 'Waiting'}
				</p>
			</div>
		</div>

		<!-- Relay URLs -->
		<div class="rounded-lg bg-white p-6 shadow">
			<h2 class="text-lg font-semibold text-gray-900">Relay URLs</h2>
			<p class="mt-1 text-sm text-gray-500">Add these to your Nostr client</p>
			<div class="mt-4 space-y-3">
				<div class="flex items-center gap-3 rounded-lg bg-gray-50 p-3">
					<code class="flex-1 text-sm text-gray-700">ws://localhost:7000</code>
					<button class="text-purple-600 hover:text-purple-700">Copy</button>
				</div>
			</div>
		</div>

		<!-- Quick actions -->
		<div class="rounded-lg bg-white p-6 shadow">
			<h2 class="text-lg font-semibold text-gray-900">Quick Actions</h2>
			<div class="mt-4 flex flex-wrap gap-3">
				<a href="/access" class="rounded-lg bg-purple-600 px-4 py-2 text-sm font-medium text-white hover:bg-purple-700">
					Manage Access
				</a>
				<a href="/events" class="rounded-lg bg-gray-100 px-4 py-2 text-sm font-medium text-gray-700 hover:bg-gray-200">
					Browse Events
				</a>
				<button class="rounded-lg bg-gray-100 px-4 py-2 text-sm font-medium text-gray-700 hover:bg-gray-200">
					Sync from Public Relays
				</button>
			</div>
		</div>
	{/if}
</div>
