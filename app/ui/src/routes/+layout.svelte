<script>
	import '../app.css';
	import { browser } from '$app/environment';
	import { notifications } from '$lib/stores/app.svelte.js';
	import { initializeTheme } from '$lib/stores/theme.svelte.js';

	let { children } = $props();

	let themeInitialized = $state(false);

	// Initialize theme on mount
	$effect(() => {
		if (browser && !themeInitialized) {
			initializeTheme();
			themeInitialized = true;
		}
	});
</script>

<div class="min-h-screen bg-gray-50 dark:bg-gray-900">
	{@render children()}

	<!-- Notification toasts -->
	{#if notifications.length > 0}
		<div class="fixed bottom-4 right-4 z-50 flex flex-col gap-2">
			{#each notifications as notification (notification.id)}
				<div
					class="rounded-lg px-4 py-3 shadow-lg text-sm font-medium flex items-center gap-2 animate-in slide-in-from-right {notification.type === 'success'
						? 'bg-green-100 text-green-800 dark:bg-green-900/50 dark:text-green-300'
						: notification.type === 'error'
							? 'bg-red-100 text-red-800 dark:bg-red-900/50 dark:text-red-300'
							: notification.type === 'warning'
								? 'bg-yellow-100 text-yellow-800 dark:bg-yellow-900/50 dark:text-yellow-300'
								: 'bg-blue-100 text-blue-800 dark:bg-blue-900/50 dark:text-blue-300'}"
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
