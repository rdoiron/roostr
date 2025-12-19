<script>
	import { browser } from '$app/environment';
	import { goto } from '$app/navigation';
	import Sidebar from '$lib/components/Sidebar.svelte';
	import Header from '$lib/components/Header.svelte';
	import Loading from '$lib/components/Loading.svelte';
	import { setup, relay } from '$lib/api/client.js';
	import { relayStatus } from '$lib/stores/app.svelte.js';
	import { initializeTimezone } from '$lib/stores/timezone.svelte.js';

	let { children } = $props();

	let sidebarOpen = $state(false);
	let loading = $state(true);
	let setupCompleted = $state(false);
	let statusInterval = null;

	// Check setup status on initial load
	$effect(() => {
		if (browser && loading) {
			checkSetup();
		}
	});

	async function checkSetup() {
		try {
			const status = await setup.getStatus();
			setupCompleted = status.completed;
			loading = false;

			// Redirect to setup if not completed
			if (!status.completed) {
				goto('/setup');
			}
		} catch (e) {
			console.error('Failed to check setup status:', e);
			loading = false;
			// On error, redirect to setup
			goto('/setup');
		}
	}

	async function checkRelayStatus() {
		try {
			const res = await relay.getStatus();
			relayStatus.online = res.status === 'running';
			relayStatus.uptime = res.uptime_seconds || 0;
			relayStatus.loading = false;
		} catch (e) {
			console.error('Failed to check relay status:', e);
			relayStatus.online = false;
			relayStatus.loading = false;
		}
	}

	// Start relay status checking and initialize timezone when setup is complete
	$effect(() => {
		if (browser && setupCompleted && !statusInterval) {
			checkRelayStatus();
			statusInterval = setInterval(checkRelayStatus, 30000);
			// Initialize timezone preference from backend
			initializeTimezone();
		}

		return () => {
			if (statusInterval) {
				clearInterval(statusInterval);
				statusInterval = null;
			}
		};
	});
</script>

{#if loading}
	<div class="min-h-screen flex items-center justify-center">
		<Loading text="Loading..." />
	</div>
{:else if setupCompleted}
	<div class="flex h-screen overflow-hidden">
		<!-- Sidebar -->
		<Sidebar bind:open={sidebarOpen} />

		<!-- Main content area -->
		<div class="flex flex-1 flex-col overflow-hidden">
			<!-- Header -->
			<Header onMenuClick={() => (sidebarOpen = !sidebarOpen)} />

			<!-- Page content -->
			<main class="flex-1 overflow-y-auto p-6">
				{@render children()}
			</main>
		</div>
	</div>
{/if}
