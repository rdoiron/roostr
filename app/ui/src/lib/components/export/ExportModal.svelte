<script>
	import { exportApi } from '$lib/api/client.js';
	import { formatBytes } from '$lib/utils/format.js';
	import Loading from '$lib/components/Loading.svelte';
	import Button from '$lib/components/Button.svelte';

	let { onClose, stats = {} } = $props();

	// Export mode: 'all' or 'selected'
	let exportMode = $state('all');

	// Selected event kinds (when exportMode === 'selected')
	let selectedKinds = $state({
		notes: true,      // kind 1
		reactions: true,  // kind 7
		follows: true,    // kind 3
		dms: false,       // kind 4, 14
		profile: true,    // kind 0
		reposts: true,    // kind 6
		other: false      // everything else
	});

	// Date mode: 'all' or 'custom'
	let dateMode = $state('all');
	let dateFrom = $state('');
	let dateTo = $state(new Date().toISOString().split('T')[0]);

	// Format: 'ndjson' or 'json'
	let format = $state('ndjson');

	// Estimate state
	let estimatedCount = $state(0);
	let estimatedBytes = $state(0);
	let estimateLoading = $state(false);
	let estimateError = $state(null);

	// Download state
	let downloading = $state(false);
	let downloadProgress = $state(0);
	let downloadError = $state(null);

	// Event kind counts from stats
	let kindCounts = $derived({
		notes: stats.events_by_kind?.posts ?? 0,
		reactions: stats.events_by_kind?.reactions ?? 0,
		follows: stats.events_by_kind?.follows ?? 0,
		dms: stats.events_by_kind?.dms ?? 0,
		profile: 0, // Not tracked separately in stats
		reposts: stats.events_by_kind?.reposts ?? 0,
		other: stats.events_by_kind?.other ?? 0
	});

	// Kind number mapping
	const kindNumbers = {
		notes: [1],
		reactions: [7],
		follows: [3],
		dms: [4, 14],
		profile: [0],
		reposts: [6]
	};

	// Build kinds array from selections
	function getSelectedKindNumbers() {
		if (exportMode === 'all') return null;

		const kinds = [];
		if (selectedKinds.notes) kinds.push(...kindNumbers.notes);
		if (selectedKinds.reactions) kinds.push(...kindNumbers.reactions);
		if (selectedKinds.follows) kinds.push(...kindNumbers.follows);
		if (selectedKinds.dms) kinds.push(...kindNumbers.dms);
		if (selectedKinds.profile) kinds.push(...kindNumbers.profile);
		if (selectedKinds.reposts) kinds.push(...kindNumbers.reposts);
		// Note: 'other' means all other kinds, which we can't specify, so we don't filter if selected
		if (selectedKinds.other) return null;

		return kinds.length > 0 ? kinds.join(',') : null;
	}

	// Build export params
	function getExportParams() {
		const params = { format };

		const kinds = getSelectedKindNumbers();
		if (kinds) params.kinds = kinds;

		if (dateMode === 'custom') {
			if (dateFrom) params.since = Math.floor(new Date(dateFrom).getTime() / 1000).toString();
			if (dateTo) params.until = Math.floor(new Date(dateTo + 'T23:59:59').getTime() / 1000).toString();
		}

		return params;
	}

	// Fetch estimate when filters change
	let estimateTimeout;
	$effect(() => {
		// Track dependencies by referencing them
		exportMode;
		selectedKinds;
		dateMode;
		dateFrom;
		dateTo;

		clearTimeout(estimateTimeout);
		estimateTimeout = setTimeout(fetchEstimate, 300);
	});

	async function fetchEstimate() {
		estimateLoading = true;
		estimateError = null;

		try {
			const params = getExportParams();
			delete params.format; // Not needed for estimate
			const result = await exportApi.getEstimate(params);
			estimatedCount = result.count;
			estimatedBytes = result.estimated_bytes;
		} catch (e) {
			estimateError = e.message;
			estimatedCount = 0;
			estimatedBytes = 0;
		} finally {
			estimateLoading = false;
		}
	}

	// Handle export download
	async function handleExport() {
		downloading = true;
		downloadProgress = 0;
		downloadError = null;

		try {
			const params = getExportParams();
			const url = exportApi.getExportUrl(params);

			const response = await fetch(url);
			if (!response.ok) {
				throw new Error(`Export failed: ${response.statusText}`);
			}

			// Get total count from header
			const totalCount = parseInt(response.headers.get('X-Total-Count') || '0', 10);
			const estimatedTotalBytes = totalCount * 500;

			// Read the stream
			const reader = response.body.getReader();
			const chunks = [];
			let receivedBytes = 0;

			while (true) {
				const { done, value } = await reader.read();
				if (done) break;

				chunks.push(value);
				receivedBytes += value.length;

				// Update progress
				if (estimatedTotalBytes > 0) {
					downloadProgress = Math.min((receivedBytes / estimatedTotalBytes) * 100, 99);
				}
			}

			downloadProgress = 100;

			// Create blob and trigger download
			const blob = new Blob(chunks, {
				type: format === 'ndjson' ? 'application/x-ndjson' : 'application/json'
			});
			const downloadUrl = URL.createObjectURL(blob);
			const a = document.createElement('a');
			a.href = downloadUrl;
			a.download = `nostr-backup-${new Date().toISOString().split('T')[0]}.${format === 'ndjson' ? 'ndjson' : 'json'}`;
			document.body.appendChild(a);
			a.click();
			document.body.removeChild(a);
			URL.revokeObjectURL(downloadUrl);

			// Close modal after successful download
			setTimeout(() => onClose(), 500);
		} catch (e) {
			downloadError = e.message;
		} finally {
			downloading = false;
		}
	}

	// Keyboard handling
	function handleKeydown(e) {
		if (e.key === 'Escape' && !downloading) {
			onClose();
		}
	}

	function handleBackdropClick(e) {
		if (e.target === e.currentTarget && !downloading) {
			onClose();
		}
	}

	// Today's date for max attribute
	const today = new Date().toISOString().split('T')[0];
</script>

<svelte:window onkeydown={handleKeydown} />

<div
	class="fixed inset-0 z-50 flex items-center justify-center bg-black/50 p-4"
	onclick={handleBackdropClick}
	onkeydown={handleKeydown}
	role="dialog"
	aria-modal="true"
	aria-labelledby="export-title"
	tabindex="-1"
>
	<div class="w-full max-w-lg rounded-lg bg-white dark:bg-gray-800 shadow-xl dark:shadow-gray-900/50">
		<!-- Header -->
		<div class="flex items-center justify-between border-b border-gray-200 dark:border-gray-700 px-6 py-4">
			<h2 id="export-title" class="text-lg font-semibold text-gray-900 dark:text-gray-100">
				Export Events
			</h2>
			<button
				type="button"
				onclick={onClose}
				disabled={downloading}
				aria-label="Close modal"
				class="p-1 text-gray-400 hover:text-gray-600 dark:hover:text-gray-300 transition-colors rounded hover:bg-gray-100 dark:hover:bg-gray-700 disabled:opacity-50"
			>
				<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
				</svg>
			</button>
		</div>

		<!-- Body -->
		<div class="p-6 space-y-6 max-h-[60vh] overflow-y-auto">
			<!-- What to export -->
			<fieldset>
				<legend class="mb-3 block text-sm font-medium text-gray-700 dark:text-gray-200">What to export:</legend>
				<div class="space-y-3">
					<label class="flex cursor-pointer items-center gap-3">
						<input
							type="radio"
							name="exportMode"
							value="all"
							bind:group={exportMode}
							disabled={downloading}
							class="h-4 w-4 text-purple-600"
						/>
						<span class="text-sm text-gray-700 dark:text-gray-200">
							All events
							{#if !estimateLoading && exportMode === 'all'}
								<span class="text-gray-500">({estimatedCount.toLocaleString()} events, ~{formatBytes(estimatedBytes)})</span>
							{/if}
						</span>
					</label>
					<label class="flex cursor-pointer items-center gap-3">
						<input
							type="radio"
							name="exportMode"
							value="selected"
							bind:group={exportMode}
							disabled={downloading}
							class="h-4 w-4 text-purple-600"
						/>
						<span class="text-sm text-gray-700 dark:text-gray-200">Selected event kinds</span>
					</label>
				</div>

				<!-- Kind checkboxes (visible when 'selected' mode) -->
				{#if exportMode === 'selected'}
					<div class="mt-4 ml-7 space-y-2">
						<label class="flex cursor-pointer items-center gap-3">
							<input
								type="checkbox"
								bind:checked={selectedKinds.notes}
								disabled={downloading}
								class="h-4 w-4 rounded border-gray-300 text-purple-600"
							/>
							<span class="text-sm text-gray-700 dark:text-gray-200">Notes (kind 1)</span>
							<span class="text-xs text-gray-400">{kindCounts.notes.toLocaleString()} events</span>
						</label>
						<label class="flex cursor-pointer items-center gap-3">
							<input
								type="checkbox"
								bind:checked={selectedKinds.reactions}
								disabled={downloading}
								class="h-4 w-4 rounded border-gray-300 text-purple-600"
							/>
							<span class="text-sm text-gray-700 dark:text-gray-200">Reactions (kind 7)</span>
							<span class="text-xs text-gray-400">{kindCounts.reactions.toLocaleString()} events</span>
						</label>
						<label class="flex cursor-pointer items-center gap-3">
							<input
								type="checkbox"
								bind:checked={selectedKinds.follows}
								disabled={downloading}
								class="h-4 w-4 rounded border-gray-300 text-purple-600"
							/>
							<span class="text-sm text-gray-700 dark:text-gray-200">Follow lists (kind 3)</span>
							<span class="text-xs text-gray-400">{kindCounts.follows.toLocaleString()} events</span>
						</label>
						<label class="flex cursor-pointer items-center gap-3">
							<input
								type="checkbox"
								bind:checked={selectedKinds.dms}
								disabled={downloading}
								class="h-4 w-4 rounded border-gray-300 text-purple-600"
							/>
							<span class="text-sm text-gray-700 dark:text-gray-200">DMs (kind 4, 14)</span>
							<span class="text-xs text-gray-400">{kindCounts.dms.toLocaleString()} events</span>
						</label>
						<label class="flex cursor-pointer items-center gap-3">
							<input
								type="checkbox"
								bind:checked={selectedKinds.profile}
								disabled={downloading}
								class="h-4 w-4 rounded border-gray-300 text-purple-600"
							/>
							<span class="text-sm text-gray-700 dark:text-gray-200">Profile (kind 0)</span>
						</label>
						<label class="flex cursor-pointer items-center gap-3">
							<input
								type="checkbox"
								bind:checked={selectedKinds.reposts}
								disabled={downloading}
								class="h-4 w-4 rounded border-gray-300 text-purple-600"
							/>
							<span class="text-sm text-gray-700 dark:text-gray-200">Reposts (kind 6)</span>
							<span class="text-xs text-gray-400">{kindCounts.reposts.toLocaleString()} events</span>
						</label>
						<label class="flex cursor-pointer items-center gap-3">
							<input
								type="checkbox"
								bind:checked={selectedKinds.other}
								disabled={downloading}
								class="h-4 w-4 rounded border-gray-300 text-purple-600"
							/>
							<span class="text-sm text-gray-700 dark:text-gray-200">Other kinds</span>
							<span class="text-xs text-gray-400">{kindCounts.other.toLocaleString()} events</span>
						</label>
					</div>
				{/if}
			</fieldset>

			<!-- Date range -->
			<fieldset>
				<legend class="mb-3 block text-sm font-medium text-gray-700 dark:text-gray-200">Date range:</legend>
				<div class="space-y-3">
					<label class="flex cursor-pointer items-center gap-3">
						<input
							type="radio"
							name="dateMode"
							value="all"
							bind:group={dateMode}
							disabled={downloading}
							class="h-4 w-4 text-purple-600"
						/>
						<span class="text-sm text-gray-700 dark:text-gray-200">All time</span>
					</label>
					<label class="flex cursor-pointer items-center gap-3">
						<input
							type="radio"
							name="dateMode"
							value="custom"
							bind:group={dateMode}
							disabled={downloading}
							class="h-4 w-4 text-purple-600"
						/>
						<span class="text-sm text-gray-700 dark:text-gray-200">Custom range</span>
					</label>
				</div>

				{#if dateMode === 'custom'}
					<div class="mt-4 ml-7 flex flex-wrap items-center gap-3">
						<div>
							<label for="date-from" class="sr-only">From date</label>
							<input
								id="date-from"
								type="date"
								bind:value={dateFrom}
								max={dateTo || today}
								disabled={downloading}
								class="rounded-lg border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100 px-3 py-2 text-sm focus:border-purple-500 focus:outline-none"
							/>
						</div>
						<span class="text-sm text-gray-500">to</span>
						<div>
							<label for="date-to" class="sr-only">To date</label>
							<input
								id="date-to"
								type="date"
								bind:value={dateTo}
								min={dateFrom}
								max={today}
								disabled={downloading}
								class="rounded-lg border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100 px-3 py-2 text-sm focus:border-purple-500 focus:outline-none"
							/>
						</div>
					</div>
				{/if}
			</fieldset>

			<!-- Format -->
			<fieldset>
				<legend class="mb-3 block text-sm font-medium text-gray-700 dark:text-gray-200">Format:</legend>
				<div class="space-y-3">
					<label class="flex cursor-pointer items-start gap-3">
						<input
							type="radio"
							name="format"
							value="ndjson"
							bind:group={format}
							disabled={downloading}
							class="mt-0.5 h-4 w-4 text-purple-600"
						/>
						<div>
							<span class="text-sm text-gray-700 dark:text-gray-200">NDJSON (Newline Delimited JSON)</span>
							<span class="ml-2 rounded bg-green-100 dark:bg-green-900/30 px-2 py-0.5 text-xs font-medium text-green-700 dark:text-green-400">Recommended</span>
							<p class="text-xs text-gray-500 dark:text-gray-400">One event per line. Best for large exports and streaming.</p>
						</div>
					</label>
					<label class="flex cursor-pointer items-start gap-3">
						<input
							type="radio"
							name="format"
							value="json"
							bind:group={format}
							disabled={downloading}
							class="mt-0.5 h-4 w-4 text-purple-600"
						/>
						<div>
							<span class="text-sm text-gray-700 dark:text-gray-200">JSON Array</span>
							<p class="text-xs text-gray-500 dark:text-gray-400">Standard JSON array format. Better compatibility.</p>
						</div>
					</label>
				</div>
			</fieldset>

			<!-- Estimate display -->
			<div class="rounded-lg bg-gray-50 dark:bg-gray-700 p-4">
				{#if estimateLoading}
					<div class="flex items-center gap-2 text-sm text-gray-500 dark:text-gray-400">
						<Loading size="sm" />
						Calculating estimate...
					</div>
				{:else if estimateError}
					<div class="text-sm text-red-600 dark:text-red-400">
						Error: {estimateError}
					</div>
				{:else}
					<div class="text-sm text-gray-700 dark:text-gray-200">
						<span class="font-medium">Estimated size:</span> ~{formatBytes(estimatedBytes)}
						<span class="text-gray-500 dark:text-gray-400">({estimatedCount.toLocaleString()} events)</span>
					</div>
				{/if}
			</div>

			<!-- Download progress -->
			{#if downloading}
				<div class="space-y-2">
					<div class="flex items-center justify-between text-sm">
						<span class="text-gray-600 dark:text-gray-400">Downloading...</span>
						<span class="text-gray-900 dark:text-gray-100 font-medium">{Math.round(downloadProgress)}%</span>
					</div>
					<div class="h-2 w-full rounded-full bg-gray-200 dark:bg-gray-700">
						<div
							class="h-2 rounded-full bg-purple-600 transition-all duration-300"
							style="width: {downloadProgress}%"
						></div>
					</div>
				</div>
			{/if}

			<!-- Download error -->
			{#if downloadError}
				<div class="p-3 bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-lg">
					<div class="flex items-start space-x-2">
						<svg class="w-5 h-5 text-red-500 dark:text-red-400 mt-0.5 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
						</svg>
						<p class="text-sm text-red-700 dark:text-red-300">{downloadError}</p>
					</div>
				</div>
			{/if}
		</div>

		<!-- Footer -->
		<div class="flex justify-end space-x-3 border-t border-gray-200 dark:border-gray-700 px-6 py-4">
			<Button variant="secondary" onclick={onClose} disabled={downloading}>
				Cancel
			</Button>
			<Button
				variant="primary"
				onclick={handleExport}
				disabled={downloading || estimateLoading || estimatedCount === 0}
				loading={downloading}
			>
				{downloading ? 'Downloading...' : 'Export & Download'}
			</Button>
		</div>
	</div>
</div>
