<script>
	import { relayStatus, syncStatus } from '$lib/stores';
	import SyncModal from '$lib/components/sync/SyncModal.svelte';

	let { onMenuClick } = $props();

	let showSyncModal = $state(false);

	// Format relative time
	function formatRelativeTime(isoString) {
		if (!isoString) return null;
		const date = new Date(isoString);
		const now = new Date();
		const diffMs = now - date;
		const diffSec = Math.floor(diffMs / 1000);
		const diffMin = Math.floor(diffSec / 60);
		const diffHours = Math.floor(diffMin / 60);
		const diffDays = Math.floor(diffHours / 24);

		if (diffSec < 60) return 'just now';
		if (diffMin < 60) return `${diffMin}m ago`;
		if (diffHours < 24) return `${diffHours}h ago`;
		return `${diffDays}d ago`;
	}

	let lastSyncText = $derived(formatRelativeTime(syncStatus.lastSyncTime));
</script>

<header class="relative flex h-16 items-center justify-between border-b border-gray-200 bg-white px-4 lg:px-6">
	<!-- Mobile menu button -->
	<button
		type="button"
		class="rounded-lg p-2 text-gray-500 hover:bg-gray-100 lg:hidden"
		onclick={onMenuClick}
		aria-label="Open menu"
	>
		<svg class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
			<path stroke-linecap="round" stroke-linejoin="round" d="M4 6h16M4 12h16M4 18h16" />
		</svg>
	</button>

	<!-- Mobile branding - absolutely positioned for true centering -->
	<div class="absolute left-1/2 flex -translate-x-1/2 items-center lg:hidden">
		<img src="/roostr-icon.svg" alt="Roostr" class="h-6 w-6 rounded" />
		<span class="ml-2 text-lg font-bold text-gray-900">Roostr</span>
	</div>

	<!-- Spacer to push right content to the right -->
	<div class="flex-1"></div>

	<!-- Right side -->
	<div class="flex items-center gap-4">
		<!-- Sync button with last sync time -->
		<div class="flex items-center gap-2">
			{#if lastSyncText}
				<span class="hidden text-xs text-gray-500 sm:inline">
					Last synced: {lastSyncText}
				</span>
			{/if}
			<button
				type="button"
				onclick={() => (showSyncModal = true)}
				class="flex items-center gap-1.5 rounded-lg px-3 py-1.5 text-sm text-gray-600 transition-colors hover:bg-gray-100 hover:text-gray-900"
				title="Sync from public relays"
			>
				{#if syncStatus.running}
					<svg class="h-4 w-4 animate-spin text-purple-600" fill="none" viewBox="0 0 24 24">
						<circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
						<path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
					</svg>
					<span class="hidden sm:inline">Syncing...</span>
				{:else}
					<svg class="h-4 w-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
					</svg>
					<span class="hidden sm:inline">Sync</span>
				{/if}
			</button>
		</div>

		<!-- Relay status indicator -->
		<div class="flex items-center gap-2 text-sm">
			<span
				class="h-2.5 w-2.5 rounded-full {relayStatus.online ? 'bg-green-500' : 'bg-red-500'}"
			></span>
			<span class="hidden text-gray-600 sm:inline">
				Relay {relayStatus.online ? 'Online' : 'Offline'}
			</span>
		</div>
	</div>
</header>

{#if showSyncModal}
	<SyncModal onClose={() => (showSyncModal = false)} />
{/if}
