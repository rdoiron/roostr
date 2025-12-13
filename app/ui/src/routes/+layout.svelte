<script>
	import '../app.css';
	import { browser } from '$app/environment';
	import { goto } from '$app/navigation';
	import { page } from '$app/stores';
	import Sidebar from '$lib/components/Sidebar.svelte';
	import Header from '$lib/components/Header.svelte';
	import Loading from '$lib/components/Loading.svelte';
	import { setup } from '$lib/api/client.js';

	let { children } = $props();

	let sidebarOpen = $state(false);
	let loading = $state(true);
	let setupCompleted = $state(false);
	let initialized = $state(false);

	$effect(() => {
		if (browser && !initialized) {
			initialized = true;
			checkSetup();
		}
	});

	async function checkSetup() {
		try {
			const status = await setup.getStatus();
			setupCompleted = status.completed;
			loading = false;

			// Redirect to setup if not completed and not already on setup page
			if (!status.completed && !$page.url.pathname.startsWith('/setup')) {
				goto('/setup');
			}
		} catch (e) {
			console.error('Failed to check setup status:', e);
			loading = false;
			// On error, assume setup is needed
			if (!$page.url.pathname.startsWith('/setup')) {
				goto('/setup');
			}
		}
	}
</script>

<div class="min-h-screen bg-gray-50">
	{#if loading}
		<!-- Loading state while checking setup status -->
		<div class="min-h-screen flex items-center justify-center">
			<Loading text="Loading..." />
		</div>
	{:else if !setupCompleted}
		<!-- Setup wizard - no navigation -->
		<main class="min-h-screen">
			{@render children()}
		</main>
	{:else}
		<!-- Main app layout with sidebar -->
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
</div>
