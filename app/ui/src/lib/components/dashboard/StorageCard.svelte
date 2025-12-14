<script>
	import StorageProgressBar from '$lib/components/storage/StorageProgressBar.svelte';

	/**
	 * @type {{ usedBytes: number, totalBytes: number, status?: string }}
	 */
	let { usedBytes = 0, totalBytes = 0, status = 'healthy' } = $props();
</script>

<div class="rounded-lg bg-white p-6 shadow">
	<div class="mb-4 flex items-center justify-between">
		<h2 class="text-lg font-semibold text-gray-900">Storage</h2>
		<a
			href="/storage"
			class="text-sm font-medium text-purple-600 hover:text-purple-700 hover:underline"
		>
			Manage Storage
		</a>
	</div>

	<StorageProgressBar {usedBytes} {totalBytes} {status} />
</div>
