<script>
	import ExportModal from '$lib/components/export/ExportModal.svelte';
	import ImportModal from '$lib/components/import/ImportModal.svelte';
	import SyncModal from '$lib/components/sync/SyncModal.svelte';
	import { syncStatus } from '$lib/stores';

	let { stats = {} } = $props();

	let showExportModal = $state(false);
	let showImportModal = $state(false);
	let showSyncModal = $state(false);

	// Callback when import succeeds to refresh the page data
	function handleImportSuccess() {
		// Reload the page to show updated stats
		window.location.reload();
	}
</script>

<div class="rounded-lg bg-white dark:bg-gray-800 p-6 shadow dark:shadow-gray-900/50">
	<h2 class="mb-4 text-lg font-semibold text-gray-900 dark:text-gray-100">Quick Actions</h2>
	<div class="flex flex-wrap gap-3">
		<a
			href="/access"
			class="inline-flex items-center gap-2 rounded-lg bg-purple-600 dark:bg-purple-500 px-4 py-2 text-sm font-medium text-white hover:bg-purple-700 dark:hover:bg-purple-600"
		>
			<svg class="h-4 w-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
				<path
					stroke-linecap="round"
					stroke-linejoin="round"
					stroke-width="2"
					d="M18 9v3m0 0v3m0-3h3m-3 0h-3m-2-5a4 4 0 11-8 0 4 4 0 018 0zM3 20a6 6 0 0112 0v1H3v-1z"
				/>
			</svg>
			Add to Whitelist
		</a>
		<a
			href="/events"
			class="inline-flex items-center gap-2 rounded-lg bg-gray-100 dark:bg-gray-700 px-4 py-2 text-sm font-medium text-gray-700 dark:text-gray-200 hover:bg-gray-200 dark:hover:bg-gray-600"
		>
			<svg class="h-4 w-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
				<path
					stroke-linecap="round"
					stroke-linejoin="round"
					stroke-width="2"
					d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10"
				/>
			</svg>
			Browse Events
		</a>
		<button
			type="button"
			onclick={() => (showSyncModal = true)}
			class="inline-flex items-center gap-2 rounded-lg bg-gray-100 dark:bg-gray-700 px-4 py-2 text-sm font-medium text-gray-700 dark:text-gray-200 hover:bg-gray-200 dark:hover:bg-gray-600"
		>
			{#if syncStatus.running}
				<svg class="h-4 w-4 animate-spin text-purple-600 dark:text-purple-400" fill="none" viewBox="0 0 24 24">
					<circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
					<path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
				</svg>
				Syncing...
			{:else}
				<svg class="h-4 w-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path
						stroke-linecap="round"
						stroke-linejoin="round"
						stroke-width="2"
						d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15"
					/>
				</svg>
				Sync from Relays
			{/if}
		</button>
		<button
			type="button"
			onclick={() => (showImportModal = true)}
			class="inline-flex items-center gap-2 rounded-lg bg-gray-100 dark:bg-gray-700 px-4 py-2 text-sm font-medium text-gray-700 dark:text-gray-200 hover:bg-gray-200 dark:hover:bg-gray-600"
		>
			<svg class="h-4 w-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
				<path
					stroke-linecap="round"
					stroke-linejoin="round"
					stroke-width="2"
					d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-8l-4-4m0 0L8 8m4-4v12"
				/>
			</svg>
			Import Backup
		</button>
		<button
			type="button"
			onclick={() => (showExportModal = true)}
			class="inline-flex items-center gap-2 rounded-lg bg-gray-100 dark:bg-gray-700 px-4 py-2 text-sm font-medium text-gray-700 dark:text-gray-200 hover:bg-gray-200 dark:hover:bg-gray-600"
		>
			<svg class="h-4 w-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
				<path
					stroke-linecap="round"
					stroke-linejoin="round"
					stroke-width="2"
					d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-4l-4 4m0 0l-4-4m4 4V4"
				/>
			</svg>
			Export Backup
		</button>
	</div>
</div>

{#if showExportModal}
	<ExportModal
		{stats}
		onClose={() => (showExportModal = false)}
	/>
{/if}

{#if showImportModal}
	<ImportModal
		onClose={() => (showImportModal = false)}
		onSuccess={handleImportSuccess}
	/>
{/if}

{#if showSyncModal}
	<SyncModal onClose={() => (showSyncModal = false)} />
{/if}
