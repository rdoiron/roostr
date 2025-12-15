<script>
	import '../app.css';
	import { browser } from '$app/environment';
	import { goto } from '$app/navigation';
	import { page } from '$app/stores';
	import Sidebar from '$lib/components/Sidebar.svelte';
	import Header from '$lib/components/Header.svelte';
	import Loading from '$lib/components/Loading.svelte';
	import { setup } from '$lib/api/client.js';
	import { notifications } from '$lib/stores/app.svelte.js';

	let { children } = $props();

	let sidebarOpen = $state(false);
	let loading = $state(true);
	let setupCompleted = $state(false);
	let lastPathname = $state('');

	// Check setup status on initial load and when navigating away from /setup
	$effect(() => {
		if (browser) {
			const currentPath = $page.url.pathname;
			const wasOnSetup = lastPathname.startsWith('/setup');
			const nowOnSetup = currentPath.startsWith('/setup');

			// Re-check if: first load OR navigating away from setup
			if (!lastPathname || (wasOnSetup && !nowOnSetup)) {
				checkSetup();
			}
			lastPathname = currentPath;
		}
	});

	async function checkSetup() {
		loading = true;
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

	<!-- Notification toasts -->
	{#if notifications.length > 0}
		<div class="fixed bottom-4 right-4 z-50 flex flex-col gap-2">
			{#each notifications as notification (notification.id)}
				<div
					class="rounded-lg px-4 py-3 shadow-lg text-sm font-medium flex items-center gap-2 animate-in slide-in-from-right {notification.type === 'success'
						? 'bg-green-100 text-green-800'
						: notification.type === 'error'
							? 'bg-red-100 text-red-800'
							: notification.type === 'warning'
								? 'bg-yellow-100 text-yellow-800'
								: 'bg-blue-100 text-blue-800'}"
				>
					{#if notification.type === 'success'}
						<svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
						</svg>
					{:else if notification.type === 'error'}
						<svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
						</svg>
					{:else if notification.type === 'warning'}
						<svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
						</svg>
					{:else}
						<svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
						</svg>
					{/if}
					{notification.message}
				</div>
			{/each}
		</div>
	{/if}
</div>
