<script>
	import { browser } from '$app/environment';
	import { stats, relay, events, storage } from '$lib/api/client.js';
	import { relayStatus } from '$lib/stores';
	import { formatUptime, formatBytes } from '$lib/utils/format.js';

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
	let initialized = $state(false);

	let dashboardData = $state({
		stats: null,
		urls: null,
		recentEvents: [],
		storage: null
	});

	const REFRESH_INTERVAL = 30000; // 30 seconds

	async function loadDashboard() {
		try {
			const [statsRes, urlsRes, eventsRes, storageRes] = await Promise.all([
				stats.getSummary(),
				relay.getURLs(),
				events.getRecent(),
				storage.getStatus()
			]);

			dashboardData.stats = statsRes;
			dashboardData.urls = urlsRes;
			dashboardData.recentEvents = eventsRes.events || [];
			dashboardData.storage = storageRes;

			// Update global relay status
			relayStatus.online = statsRes.relay_status === 'online';
			relayStatus.uptime = statsRes.uptime_seconds;
			relayStatus.loading = false;

			error = null;
		} catch (e) {
			error = e.message;
			relayStatus.online = false;
		} finally {
			loading = false;
			relayStatus.loading = false;
		}
	}

	// Initialize on client side
	$effect(() => {
		if (browser && !initialized) {
			initialized = true;
			loadDashboard();

			// Auto-refresh every 30 seconds
			const refreshInterval = setInterval(loadDashboard, REFRESH_INTERVAL);

			// Cleanup on unmount
			return () => {
				clearInterval(refreshInterval);
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
		<h1 class="text-2xl font-bold text-gray-900">Dashboard</h1>
		<p class="text-gray-600">Your private Nostr relay at a glance</p>
	</div>

	{#if loading}
		<div class="flex items-center justify-center py-12">
			<Loading text="Loading dashboard..." />
		</div>
	{:else if error}
		<Error title="Error loading dashboard" message={error} onRetry={loadDashboard} />
	{:else}
		<!-- Relay Status Card -->
		<div class="rounded-lg bg-white p-6 shadow">
			<div class="mb-6 flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
				<div class="flex items-center gap-4">
					<h2 class="text-lg font-semibold text-gray-900">Relay Status</h2>
					<StatusIndicator status={statusType} />
				</div>
				<div class="text-sm text-gray-500">
					Uptime: <span class="font-medium text-gray-900"
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
				value={dashboardData.stats?.total_events?.toLocaleString() ?? '0'}
				subtext="+{dashboardData.stats?.events_today ?? 0} today"
			/>
			<StatCard
				label="Storage Used"
				value={formatBytes(dashboardData.storage?.total_size ?? 0)}
			/>
			<StatCard
				label="Whitelisted Pubkeys"
				value={dashboardData.stats?.whitelisted_count?.toString() ?? '0'}
			/>
		</div>

		<!-- Storage Card with Progress Bar -->
		<StorageCard
			usedBytes={dashboardData.storage?.total_size ?? 0}
			totalBytes={dashboardData.storage?.total_space ?? 0}
			status={dashboardData.storage?.status ?? 'healthy'}
		/>

		<!-- Event Type Breakdown -->
		<div class="grid grid-cols-2 gap-4 sm:grid-cols-3 lg:grid-cols-6">
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
