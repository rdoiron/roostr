<script>
	import { importApi } from '$lib/api/client.js';
	import { formatBytes } from '$lib/utils/format.js';
	import Button from '$lib/components/Button.svelte';

	let { onClose, onSuccess } = $props();

	// File upload state
	let fileInput = $state(null);
	let selectedFile = $state(null);
	let fileSize = $state(0);
	let fileName = $state('');
	let detectedFormat = $state('');

	// Import options
	let verifySignatures = $state(true);
	let skipDuplicates = $state(true);
	let stopOnError = $state(false);

	// Import state
	let importing = $state(false);
	let importProgress = $state(null);
	let importError = $state(null);

	// Handle file selection
	function handleFileSelect(e) {
		const file = e.target.files?.[0];
		if (file) {
			selectedFile = file;
			fileSize = file.size;
			fileName = file.name;

			// Detect format from file extension
			if (file.name.endsWith('.ndjson') || file.name.endsWith('.jsonl')) {
				detectedFormat = 'NDJSON';
			} else if (file.name.endsWith('.json')) {
				detectedFormat = 'JSON';
			} else {
				detectedFormat = 'Unknown';
			}
		}
	}

	// Handle drag and drop
	function handleDrop(e) {
		e.preventDefault();
		const file = e.dataTransfer.files?.[0];
		if (file) {
			selectedFile = file;
			fileSize = file.size;
			fileName = file.name;

			// Detect format
			if (file.name.endsWith('.ndjson') || file.name.endsWith('.jsonl')) {
				detectedFormat = 'NDJSON';
			} else if (file.name.endsWith('.json')) {
				detectedFormat = 'JSON';
			} else {
				detectedFormat = 'Unknown';
			}
		}
	}

	function handleDragOver(e) {
		e.preventDefault();
	}

	// Trigger file input click
	function triggerFileInput() {
		fileInput?.click();
	}

	// Handle import
	async function handleImport() {
		if (!selectedFile) return;

		importing = true;
		importProgress = null;
		importError = null;

		try {
			const formData = new FormData();
			formData.append('file', selectedFile);
			formData.append('verify_signatures', verifySignatures.toString());
			formData.append('skip_duplicates', skipDuplicates.toString());
			formData.append('stop_on_error', stopOnError.toString());

			const result = await importApi.importEvents(formData);
			importProgress = result;

			// If successful, close after a brief delay
			if (result.errors === 0) {
				setTimeout(() => {
					if (onSuccess) onSuccess();
					onClose();
				}, 1500);
			}
		} catch (e) {
			importError = e.message;
		} finally {
			importing = false;
		}
	}

	// Keyboard handling
	function handleKeydown(e) {
		if (e.key === 'Escape' && !importing) {
			onClose();
		}
	}

	function handleBackdropClick(e) {
		if (e.target === e.currentTarget && !importing) {
			onClose();
		}
	}
</script>

<svelte:window onkeydown={handleKeydown} />

<div
	class="fixed inset-0 z-50 flex items-center justify-center bg-black/50 p-4"
	onclick={handleBackdropClick}
	onkeydown={handleKeydown}
	role="dialog"
	aria-modal="true"
	aria-labelledby="import-title"
	tabindex="-1"
>
	<div class="w-full max-w-lg rounded-lg bg-white dark:bg-gray-800 shadow-xl dark:shadow-gray-900/50">
		<!-- Header -->
		<div class="flex items-center justify-between border-b border-gray-200 dark:border-gray-700 px-6 py-4">
			<h2 id="import-title" class="text-lg font-semibold text-gray-900 dark:text-gray-100">
				Import Events
			</h2>
			<button
				type="button"
				onclick={onClose}
				disabled={importing}
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
			<!-- File upload area -->
			{#if !selectedFile}
				<div
					class="border-2 border-dashed border-gray-300 dark:border-gray-600 rounded-lg p-8 text-center cursor-pointer hover:border-purple-500 dark:hover:border-purple-400 transition-colors"
					ondrop={handleDrop}
					ondragover={handleDragOver}
					onclick={triggerFileInput}
					role="button"
					tabindex="0"
					onkeydown={(e) => e.key === 'Enter' && triggerFileInput()}
				>
					<svg class="mx-auto h-12 w-12 text-gray-400" stroke="currentColor" fill="none" viewBox="0 0 48 48">
						<path d="M28 8H12a4 4 0 00-4 4v20m32-12v8m0 0v8a4 4 0 01-4 4H12a4 4 0 01-4-4v-4m32-4l-3.172-3.172a4 4 0 00-5.656 0L28 28M8 32l9.172-9.172a4 4 0 015.656 0L28 28m0 0l4 4m4-24h8m-4-4v8m-12 4h.02" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" />
					</svg>
					<p class="mt-2 text-sm text-gray-600 dark:text-gray-400">
						<span class="font-semibold">Click to upload</span> or drag and drop
					</p>
					<p class="text-xs text-gray-500 dark:text-gray-500">
						NDJSON or JSON format (from Roostr, strfry, nosdump, etc.)
					</p>
				</div>
				<input
					type="file"
					bind:this={fileInput}
					onchange={handleFileSelect}
					accept=".json,.ndjson,.jsonl"
					class="hidden"
				/>
			{:else}
				<!-- Selected file display -->
				<div class="rounded-lg bg-gray-50 dark:bg-gray-700 p-4 space-y-3">
					<div class="flex items-start justify-between">
						<div class="flex-1">
							<div class="flex items-center gap-2">
								<svg class="w-5 h-5 text-purple-600 dark:text-purple-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
								</svg>
								<span class="text-sm font-medium text-gray-900 dark:text-gray-100">{fileName}</span>
							</div>
							<div class="mt-1 flex items-center gap-3 text-xs text-gray-500 dark:text-gray-400">
								<span>{formatBytes(fileSize)}</span>
								{#if detectedFormat}
									<span>•</span>
									<span class="font-medium">{detectedFormat}</span>
								{/if}
							</div>
						</div>
						<button
							type="button"
							onclick={() => { selectedFile = null; fileName = ''; fileSize = 0; detectedFormat = ''; }}
							disabled={importing}
							aria-label="Remove selected file"
							class="text-gray-400 hover:text-gray-600 dark:hover:text-gray-300"
						>
							<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
							</svg>
						</button>
					</div>
				</div>

				<!-- Import options -->
				<fieldset>
					<legend class="mb-3 block text-sm font-medium text-gray-700 dark:text-gray-200">Import options:</legend>
					<div class="space-y-3">
						<label class="flex cursor-pointer items-start gap-3">
							<input
								type="checkbox"
								bind:checked={verifySignatures}
								disabled={importing}
								class="mt-0.5 h-4 w-4 rounded border-gray-300 text-purple-600"
							/>
							<div>
								<span class="text-sm text-gray-700 dark:text-gray-200">Verify signatures</span>
								<span class="ml-2 rounded bg-green-100 dark:bg-green-900/30 px-2 py-0.5 text-xs font-medium text-green-700 dark:text-green-400">Recommended</span>
								<p class="text-xs text-gray-500 dark:text-gray-400">Validate event signatures before importing (slower but safer)</p>
							</div>
						</label>
						<label class="flex cursor-pointer items-start gap-3">
							<input
								type="checkbox"
								bind:checked={skipDuplicates}
								disabled={importing}
								class="mt-0.5 h-4 w-4 rounded border-gray-300 text-purple-600"
							/>
							<div>
								<span class="text-sm text-gray-700 dark:text-gray-200">Skip duplicates</span>
								<p class="text-xs text-gray-500 dark:text-gray-400">Silently skip events that already exist</p>
							</div>
						</label>
						<label class="flex cursor-pointer items-start gap-3">
							<input
								type="checkbox"
								bind:checked={stopOnError}
								disabled={importing}
								class="mt-0.5 h-4 w-4 rounded border-gray-300 text-purple-600"
							/>
							<div>
								<span class="text-sm text-gray-700 dark:text-gray-200">Stop on error</span>
								<p class="text-xs text-gray-500 dark:text-gray-400">Stop importing if any event fails (default: continue)</p>
							</div>
						</label>
					</div>
				</fieldset>
			{/if}

			<!-- Import progress -->
			{#if importProgress}
				<div class="rounded-lg bg-gray-50 dark:bg-gray-700 p-4 space-y-3">
					<div class="flex items-center justify-between text-sm">
						<span class="font-medium text-gray-700 dark:text-gray-200">Import complete!</span>
						{#if importProgress.errors === 0}
							<svg class="w-5 h-5 text-green-600 dark:text-green-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
							</svg>
						{:else}
							<svg class="w-5 h-5 text-yellow-600 dark:text-yellow-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
							</svg>
						{/if}
					</div>
					<div class="grid grid-cols-2 gap-3 text-sm">
						<div>
							<span class="text-gray-500 dark:text-gray-400">Total:</span>
							<span class="ml-2 font-medium text-gray-900 dark:text-gray-100">{importProgress.total}</span>
						</div>
						<div>
							<span class="text-gray-500 dark:text-gray-400">Processed:</span>
							<span class="ml-2 font-medium text-gray-900 dark:text-gray-100">{importProgress.processed}</span>
						</div>
						<div>
							<span class="text-gray-500 dark:text-gray-400">Added:</span>
							<span class="ml-2 font-medium text-green-600 dark:text-green-400">{importProgress.added}</span>
						</div>
						<div>
							<span class="text-gray-500 dark:text-gray-400">Duplicates:</span>
							<span class="ml-2 font-medium text-gray-600 dark:text-gray-400">{importProgress.duplicates}</span>
						</div>
						{#if importProgress.errors > 0}
							<div class="col-span-2">
								<span class="text-gray-500 dark:text-gray-400">Errors:</span>
								<span class="ml-2 font-medium text-red-600 dark:text-red-400">{importProgress.errors}</span>
							</div>
						{/if}
					</div>
					{#if importProgress.error_list && importProgress.error_list.length > 0}
						<details class="text-xs">
							<summary class="cursor-pointer text-gray-600 dark:text-gray-400 hover:text-gray-900 dark:hover:text-gray-200">
								Show errors ({importProgress.error_list.length})
							</summary>
							<div class="mt-2 space-y-1 max-h-32 overflow-y-auto">
								{#each importProgress.error_list as error}
									<div class="text-red-600 dark:text-red-400">{error}</div>
								{/each}
							</div>
						</details>
					{/if}
				</div>
			{/if}

			<!-- Import error -->
			{#if importError}
				<div class="p-3 bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-lg">
					<div class="flex items-start space-x-2">
						<svg class="w-5 h-5 text-red-500 dark:text-red-400 mt-0.5 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
						</svg>
						<p class="text-sm text-red-700 dark:text-red-300">{importError}</p>
					</div>
				</div>
			{/if}

			<!-- Info box -->
			{#if !selectedFile && !importing && !importProgress}
				<div class="rounded-lg bg-blue-50 dark:bg-blue-900/20 border border-blue-200 dark:border-blue-800 p-4">
					<div class="flex items-start gap-2">
						<svg class="w-5 h-5 text-blue-600 dark:text-blue-400 mt-0.5 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
						</svg>
						<div class="text-sm text-blue-700 dark:text-blue-300">
							<p class="font-medium mb-1">Compatible with:</p>
							<ul class="list-disc list-inside space-y-0.5 text-xs">
								<li>Roostr exports (NDJSON or JSON)</li>
								<li>strfry relay exports (<code>strfry export</code>)</li>
								<li>nosdump backups</li>
								<li>nostrudel exports</li>
								<li>Any standard Nostr event JSON file</li>
							</ul>
						</div>
					</div>
				</div>
			{/if}
		</div>

		<!-- Footer -->
		<div class="flex justify-end space-x-3 border-t border-gray-200 dark:border-gray-700 px-6 py-4">
			<Button variant="secondary" onclick={onClose} disabled={importing}>
				{importProgress && importProgress.errors === 0 ? 'Close' : 'Cancel'}
			</Button>
			{#if selectedFile && !importProgress}
				<Button
					variant="primary"
					onclick={handleImport}
					disabled={importing || !selectedFile}
					loading={importing}
				>
					{importing ? 'Importing...' : 'Import Events'}
				</Button>
			{/if}
		</div>
	</div>
</div>
