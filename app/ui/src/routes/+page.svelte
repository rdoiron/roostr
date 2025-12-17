<script>
	import { browser } from '$app/environment';
	import { relay } from '$lib/api/client.js';
	import { formatUptime, formatBytes, formatCompactNumber } from '$lib/utils/format.js';

	import Loading from '$lib/components/Loading.svelte';
	import Error from '$lib/components/Error.svelte';
	import StatCard from '$lib/components/dashboard/StatCard.svelte';
	import StatusIndicator from '$lib/components/dashboard/StatusIndicator.svelte';
	import RelayURLCard from '$lib/components/dashboard/RelayURLCard.svelte';
	import EventTypeCard from '$lib/components/dashboard/EventTypeCard.svelte';
	import RecentActivityFeed from '$lib/components/dashboard/RecentActivityFeed.svelte';
	import QuickActions from '$lib/components/dashboard/QuickActions.svelte';
	import StorageCard from '$lib/components/dashboard/StorageCard.svelte';

	let loading = $state(true);
	let error = $state(null);
	let initialized = false; // Not reactive - prevents $effect re-run
	let connected = $state(false);

	let dashboardData = $state({
		stats: null,
		urls: null,
		recentEvents: [],
		storage: null
	});

	// Fetch relay URLs (only needed once, built client-side)
	async function loadRelayUrls() {
		try {
			const urlsRes = await relay.getURLs();
			const relayPort = urlsRes.relay_port || '7000';
			const localUrl = `ws://${window.location.hostname}:${relayPort}`;
			dashboardData.urls = {
				...urlsRes,
				local: localUrl
			};
		} catch (e) {
			console.error('Failed to load relay URLs:', e);
		}
	}

	// Initialize SSE connection for real-time dashboard updates
	$effect(() => {
		if (browser && !initialized) {
			initialized = true;

			// Load relay URLs once (they don't change)
			loadRelayUrls();

			// Connect to SSE stream for real-time stats
			const eventSource = new EventSource('/api/v1/stats/stream');

			eventSource.addEventListener('connected', () => {
				connected = true;
				error = null;
			});

			eventSource.addEventListener('stats', (e) => {
				try {
					const data = JSON.parse(e.data);

					// Update stats
					if (data.stats) {
						dashboardData.stats = data.stats;
					}

					// Update recent events
					if (data.recentEvents) {
						dashboardData.recentEvents = data.recentEvents;
					}

					// Update storage
					if (data.storage) {
						dashboardData.storage = data.storage;
					}

					loading = false;
					error = null;
				} catch (err) {
					console.error('Failed to parse SSE data:', err);
				}
			});

			eventSource.onerror = () => {
				connected = false;
				// EventSource will automatically reconnect
				// Only show error if we haven't loaded data yet
				if (loading) {
					error = 'Connection lost. Reconnecting...';
				}
			};

			// Cleanup on unmount
			return () => {
				eventSource.close();
			};
		}
	});

	// Derived values
	let statusType = $derived(
		dashboardData.stats?.relay_status === 'online' ? 'online' : 'offline'
	);

	let eventsByKind = $derived(dashboardData.stats?.events_by_kind || {});
</script>

<div class="space-y-6">
	<!-- Page Header -->
	<div>
		<h1 class="text-2xl font-bold text-gray-900 dark:text-gray-100">Dashboard</h1>
		<p class="text-gray-600 dark:text-gray-400">Your private Nostr relay at a glance</p>
	</div>

	{#if loading}
		<div class="flex items-center justify-center py-12">
			<Loading text="Loading dashboard..." />
		</div>
	{:else if error}
		<Error title="Error loading dashboard" message={error} onRetry={loadDashboard} />
	{:else}
		<!-- Relay Status Card -->
		<div class="rounded-lg bg-white dark:bg-gray-800 p-6 shadow dark:shadow-gray-900/50">
			<div class="mb-6 flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
				<div class="flex items-center gap-4">
					<h2 class="text-lg font-semibold text-gray-900 dark:text-gray-100">Relay Status</h2>
					<StatusIndicator status={statusType} />
				</div>
				<div class="text-sm text-gray-500 dark:text-gray-400">
					Uptime: <span class="font-medium text-gray-900 dark:text-gray-100"
						>{formatUptime(dashboardData.stats?.uptime_seconds || 0)}</span
					>
				</div>
			</div>

			<div class="space-y-3">
				{#if dashboardData.urls?.local}
					<RelayURLCard label="Local Network" url={dashboardData.urls.local} />
				{/if}
				{#if dashboardData.urls?.tor_available && dashboardData.urls?.tor}
					<RelayURLCard label="Tor (Remote Access)" url={dashboardData.urls.tor} />
				{/if}
			</div>
		</div>

		<!-- Primary Stats Cards -->
		<div class="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
			<StatCard
				label="Total Events"
				value={formatCompactNumber(dashboardData.stats?.total_events ?? 0)}
				tooltip={(dashboardData.stats?.total_events ?? 0).toLocaleString()}
				subtext="+{formatCompactNumber(dashboardData.stats?.events_today ?? 0)} today"
			/>
			<StatCard
				label="Storage Used"
				value={formatBytes(dashboardData.storage?.total_size ?? 0)}
			/>
			<StatCard
				label="Whitelisted Pubkeys"
				value={formatCompactNumber(dashboardData.stats?.whitelisted_count ?? 0)}
				tooltip={(dashboardData.stats?.whitelisted_count ?? 0).toLocaleString()}
			/>
		</div>

		<!-- Storage Card with Progress Bar -->
		<StorageCard
			usedBytes={dashboardData.storage?.total_size ?? 0}
			totalBytes={dashboardData.storage?.available_space ?? 0}
			status={dashboardData.storage?.status ?? 'healthy'}
		/>

		<!-- Event Type Breakdown -->
		<div class="grid grid-cols-2 gap-4 md:grid-cols-3 xl:grid-cols-6">
			<EventTypeCard label="Posts" count={eventsByKind.posts ?? 0} icon="post" />
			<EventTypeCard label="Reactions" count={eventsByKind.reactions ?? 0} icon="reaction" />
			<EventTypeCard label="DMs" count={eventsByKind.dms ?? 0} icon="dm" />
			<EventTypeCard label="Reposts" count={eventsByKind.reposts ?? 0} icon="repost" />
			<EventTypeCard label="Follows" count={eventsByKind.follows ?? 0} icon="follow" />
			<EventTypeCard label="Other" count={eventsByKind.other ?? 0} icon="other" />
		</div>

		<!-- Recent Activity Feed -->
		<RecentActivityFeed events={dashboardData.recentEvents} />

		<!-- Quick Actions -->
		<QuickActions stats={dashboardData.stats} />
	{/if}
</div>
