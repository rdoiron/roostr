<script>
	import { sync, access } from '$lib/api/client.js';
	import { syncStatus, notify } from '$lib/stores';
	import Loading from '$lib/components/Loading.svelte';
	import Button from '$lib/components/Button.svelte';

	let { onClose, onBackground } = $props();

	// Modal phase: 'config' | 'syncing' | 'complete'
	let phase = $state('config');

	// Configuration state
	let whitelist = $state([]);
	let defaultRelays = $state([]);
	let loadingConfig = $state(true);
	let configError = $state(null);
	let configLoaded = $state(false);

	// Selected pubkeys (map of pubkey -> boolean)
	let selectedPubkeys = $state({});

	// Custom pubkey input
	let customPubkey = $state('');
	let customPubkeyError = $state(null);
	let validatingPubkey = $state(false);
	let addedCustomPubkeys = $state([]);

	// Selected relays
	let selectedRelays = $state({});

	// Custom relay input
	let customRelay = $state('');
	let customRelayError = $state(null);
	let addedCustomRelays = $state([]);

	// Event types
	let eventTypes = $state({
		posts: true,
		reactions: true,
		reposts: true,
		profile: true,
		follows: true,
		dms: false
	});

	// Sync state
	let syncing = $state(false);
	let syncJobId = $state(null);
	let syncProgress = $state(null);
	let pollInterval = null;

	// Complete state
	let syncResult = $state(null);

	// Event kind mapping
	const kindMap = {
		posts: [1],
		reactions: [7],
		reposts: [6],
		profile: [0],
		follows: [3],
		dms: [4, 14]
	};

	// Load configuration on component init
	$effect(() => {
		if (!configLoaded) {
			configLoaded = true;
			loadConfig();
		}
	});

	async function loadConfig() {
		try {
			const [whitelistRes, relaysRes] = await Promise.all([
				access.getWhitelist(),
				sync.getRelays()
			]);

			whitelist = whitelistRes.entries || [];
			defaultRelays = relaysRes.relays || [];

			// Initialize selected pubkeys (all checked by default)
			const pubkeySelections = {};
			whitelist.forEach((entry) => {
				pubkeySelections[entry.pubkey] = true;
			});
			selectedPubkeys = pubkeySelections;

			// Initialize selected relays (all checked by default)
			const relaySelections = {};
			defaultRelays.forEach((relay) => {
				relaySelections[relay] = true;
			});
			selectedRelays = relaySelections;

			// Check if a sync is already running
			if (syncStatus.running && syncStatus.jobId) {
				syncJobId = syncStatus.jobId;
				phase = 'syncing';
				startPolling();
			}
		} catch (e) {
			configError = e.message;
		} finally {
			loadingConfig = false;
		}
	}

	// Cleanup polling on unmount
	$effect(() => {
		return () => {
			stopPolling();
		};
	});

	// Get display name for pubkey
	function getDisplayName(entry) {
		if (entry.nickname) return entry.nickname;
		if (entry.npub) return truncateNpub(entry.npub);
		return truncatePubkey(entry.pubkey);
	}

	function truncateNpub(npub) {
		if (npub.length <= 20) return npub;
		return npub.slice(0, 12) + '...' + npub.slice(-8);
	}

	function truncatePubkey(pubkey) {
		if (pubkey.length <= 16) return pubkey;
		return pubkey.slice(0, 8) + '...' + pubkey.slice(-8);
	}

	// Format large numbers compactly (1.2K, 3.5M)
	function formatCompactNumber(num) {
		if (num >= 1000000) {
			return (num / 1000000).toFixed(1).replace(/\.0$/, '') + 'M';
		}
		if (num >= 1000) {
			return (num / 1000).toFixed(1).replace(/\.0$/, '') + 'K';
		}
		return num.toString();
	}

	// Add custom pubkey
	async function addCustomPubkey() {
		if (!customPubkey.trim()) return;

		customPubkeyError = null;
		validatingPubkey = true;

		try {
			let pubkey = customPubkey.trim();

			// If it looks like a NIP-05, resolve it
			if (pubkey.includes('@')) {
				const res = await access.resolveNip05(pubkey);
				pubkey = res.pubkey;
			} else if (pubkey.startsWith('npub')) {
				// Validate npub format - the API will handle conversion
				const res = await access.resolveNip05(pubkey);
				pubkey = res.pubkey;
			}

			// Check if already in whitelist
			if (whitelist.some((e) => e.pubkey === pubkey)) {
				customPubkeyError = 'Already in whitelist';
				return;
			}

			// Check if already added as custom
			if (addedCustomPubkeys.some((e) => e.pubkey === pubkey)) {
				customPubkeyError = 'Already added';
				return;
			}

			addedCustomPubkeys = [
				...addedCustomPubkeys,
				{ pubkey, nickname: customPubkey, npub: '' }
			];
			selectedPubkeys[pubkey] = true;
			customPubkey = '';
		} catch (e) {
			customPubkeyError = e.message || 'Invalid pubkey or NIP-05';
		} finally {
			validatingPubkey = false;
		}
	}

	// Add custom relay
	function addCustomRelay() {
		if (!customRelay.trim()) return;

		customRelayError = null;
		let relay = customRelay.trim();

		// Validate relay URL
		if (!relay.startsWith('wss://') && !relay.startsWith('ws://')) {
			customRelayError = 'Relay URL must start with wss:// or ws://';
			return;
		}

		try {
			new URL(relay);
		} catch {
			customRelayError = 'Invalid URL format';
			return;
		}

		// Check if already in defaults
		if (defaultRelays.includes(relay)) {
			customRelayError = 'Already in list';
			return;
		}

		// Check if already added as custom
		if (addedCustomRelays.includes(relay)) {
			customRelayError = 'Already added';
			return;
		}

		addedCustomRelays = [...addedCustomRelays, relay];
		selectedRelays[relay] = true;
		customRelay = '';
	}

	// Get selected kinds array
	function getSelectedKinds() {
		const kinds = [];
		Object.entries(eventTypes).forEach(([type, selected]) => {
			if (selected && kindMap[type]) {
				kinds.push(...kindMap[type]);
			}
		});
		return kinds;
	}

	// Get selected pubkeys array
	function getSelectedPubkeys() {
		return Object.entries(selectedPubkeys)
			.filter(([, selected]) => selected)
			.map(([pubkey]) => pubkey);
	}

	// Get selected relays array
	function getSelectedRelays() {
		return Object.entries(selectedRelays)
			.filter(([, selected]) => selected)
			.map(([relay]) => relay);
	}

	// Start sync
	async function startSync() {
		const pubkeys = getSelectedPubkeys();
		const relays = getSelectedRelays();
		const kinds = getSelectedKinds();

		if (pubkeys.length === 0) {
			notify('error', 'Please select at least one pubkey');
			return;
		}

		if (relays.length === 0) {
			notify('error', 'Please select at least one relay');
			return;
		}

		if (kinds.length === 0) {
			notify('error', 'Please select at least one event type');
			return;
		}

		syncing = true;

		try {
			const res = await sync.start({
				pubkeys,
				relays,
				event_kinds: kinds
			});

			syncJobId = res.job_id;
			syncStatus.running = true;
			syncStatus.jobId = res.job_id;
			phase = 'syncing';
			startPolling();
		} catch (e) {
			notify('error', e.message);
			syncing = false;
		}
	}

	// Poll for sync status
	function startPolling() {
		pollInterval = setInterval(async () => {
			try {
				const status = await sync.getStatus(syncJobId);
				syncProgress = status;
				syncStatus.progress = status;

				if (status.status === 'completed' || status.status === 'failed' || status.status === 'cancelled') {
					stopPolling();
					syncResult = status;
					syncStatus.running = false;
					syncStatus.jobId = null;
					syncStatus.progress = null;
					if (status.status === 'completed') {
						syncStatus.lastSyncTime = new Date().toISOString();
					}
					phase = 'complete';
				}
			} catch (e) {
				console.error('Failed to poll sync status:', e);
			}
		}, 1500);
	}

	function stopPolling() {
		if (pollInterval) {
			clearInterval(pollInterval);
			pollInterval = null;
		}
	}

	// Cancel sync
	async function cancelSync() {
		try {
			await sync.cancel();
			notify('info', 'Sync cancelled');
		} catch (e) {
			notify('error', e.message);
		}
	}

	// Send to background
	function sendToBackground() {
		stopPolling();
		if (onBackground) onBackground();
		onClose();
	}

	// Keyboard handling
	function handleKeydown(e) {
		if (e.key === 'Escape' && phase !== 'syncing') {
			onClose();
		}
	}

	function handleBackdropClick(e) {
		if (e.target === e.currentTarget && phase !== 'syncing') {
			onClose();
		}
	}

	// Derived values
	let canStartSync = $derived(
		getSelectedPubkeys().length > 0 &&
		getSelectedRelays().length > 0 &&
		getSelectedKinds().length > 0
	);

	let progressPercent = $derived(() => {
		if (!syncProgress) return 0;
		// Estimate progress based on events processed
		// This is approximate since we don't know total events
		const total = syncProgress.events_fetched || 0;
		if (total === 0) return 5; // Show some initial progress
		// Cap at 95% until complete
		return Math.min(95, 5 + (total / 100) * 10);
	});
</script>

<svelte:window onkeydown={handleKeydown} />

<div
	class="fixed inset-0 z-50 flex items-center justify-center bg-black/50 p-4"
	onclick={handleBackdropClick}
	onkeydown={handleKeydown}
	role="dialog"
	aria-modal="true"
	aria-labelledby="sync-title"
	tabindex="-1"
>
	<div class="w-full max-w-lg rounded-lg bg-white shadow-xl">
		<!-- Header -->
		<div class="flex items-center justify-between border-b px-6 py-4">
			<h2 id="sync-title" class="text-lg font-semibold text-gray-900">
				{#if phase === 'config'}
					Sync from Public Relays
				{:else if phase === 'syncing'}
					Syncing...
				{:else}
					Sync Complete
				{/if}
			</h2>
			{#if phase !== 'syncing'}
				<button
					type="button"
					onclick={onClose}
					aria-label="Close modal"
					class="rounded p-1 text-gray-400 transition-colors hover:bg-gray-100 hover:text-gray-600"
				>
					<svg class="h-5 w-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
					</svg>
				</button>
			{/if}
		</div>

		<!-- Body -->
		<div class="max-h-[60vh] overflow-y-auto p-6">
			{#if phase === 'config'}
				{#if loadingConfig}
					<div class="flex items-center justify-center py-8">
						<Loading text="Loading configuration..." />
					</div>
				{:else if configError}
					<div class="rounded-lg border border-red-200 bg-red-50 p-4">
						<p class="text-sm text-red-700">{configError}</p>
					</div>
				{:else}
					<div class="space-y-6">
						<p class="text-sm text-gray-600">
							Import your Nostr history from public relays into Roostr.
						</p>

						<!-- Pubkeys Section -->
						<fieldset>
							<legend class="mb-3 block text-sm font-medium text-gray-700">Pubkeys to sync:</legend>
							<div class="max-h-40 space-y-2 overflow-y-auto rounded-lg border border-gray-200 p-3">
								{#each whitelist as entry}
									<label class="flex cursor-pointer items-center gap-3">
										<input
											type="checkbox"
											bind:checked={selectedPubkeys[entry.pubkey]}
											class="h-4 w-4 rounded border-gray-300 text-purple-600"
										/>
										<span class="text-sm text-gray-700">{getDisplayName(entry)}</span>
										{#if entry.is_operator}
											<span class="rounded bg-purple-100 px-1.5 py-0.5 text-xs text-purple-700">You</span>
										{/if}
									</label>
								{/each}
								{#each addedCustomPubkeys as entry}
									<label class="flex cursor-pointer items-center gap-3">
										<input
											type="checkbox"
											bind:checked={selectedPubkeys[entry.pubkey]}
											class="h-4 w-4 rounded border-gray-300 text-purple-600"
										/>
										<span class="text-sm text-gray-700">{entry.nickname}</span>
										<span class="rounded bg-gray-100 px-1.5 py-0.5 text-xs text-gray-600">Custom</span>
									</label>
								{/each}
							</div>

							<!-- Add custom pubkey -->
							<div class="mt-3">
								<div class="flex gap-2">
									<input
										type="text"
										bind:value={customPubkey}
										placeholder="npub... or user@domain.com"
										disabled={validatingPubkey}
										class="flex-1 rounded-lg border border-gray-300 px-3 py-2 text-sm focus:border-purple-500 focus:outline-none disabled:bg-gray-50"
										onkeydown={(e) => e.key === 'Enter' && addCustomPubkey()}
									/>
									<Button
										variant="secondary"
										onclick={addCustomPubkey}
										disabled={validatingPubkey || !customPubkey.trim()}
										loading={validatingPubkey}
									>
										Add
									</Button>
								</div>
								{#if customPubkeyError}
									<p class="mt-1 text-xs text-red-600">{customPubkeyError}</p>
								{/if}
							</div>
						</fieldset>

						<!-- Relays Section -->
						<fieldset>
							<legend class="mb-3 block text-sm font-medium text-gray-700">Source relays:</legend>
							<div class="max-h-40 space-y-2 overflow-y-auto rounded-lg border border-gray-200 p-3">
								{#each defaultRelays as relay}
									<label class="flex cursor-pointer items-center gap-3">
										<input
											type="checkbox"
											bind:checked={selectedRelays[relay]}
											class="h-4 w-4 rounded border-gray-300 text-purple-600"
										/>
										<span class="truncate text-sm text-gray-700">{relay}</span>
									</label>
								{/each}
								{#each addedCustomRelays as relay}
									<label class="flex cursor-pointer items-center gap-3">
										<input
											type="checkbox"
											bind:checked={selectedRelays[relay]}
											class="h-4 w-4 rounded border-gray-300 text-purple-600"
										/>
										<span class="truncate text-sm text-gray-700">{relay}</span>
										<span class="rounded bg-gray-100 px-1.5 py-0.5 text-xs text-gray-600">Custom</span>
									</label>
								{/each}
							</div>

							<!-- Add custom relay -->
							<div class="mt-3">
								<div class="flex gap-2">
									<input
										type="text"
										bind:value={customRelay}
										placeholder="wss://relay.example.com"
										class="flex-1 rounded-lg border border-gray-300 px-3 py-2 text-sm focus:border-purple-500 focus:outline-none"
										onkeydown={(e) => e.key === 'Enter' && addCustomRelay()}
									/>
									<Button variant="secondary" onclick={addCustomRelay} disabled={!customRelay.trim()}>
										Add
									</Button>
								</div>
								{#if customRelayError}
									<p class="mt-1 text-xs text-red-600">{customRelayError}</p>
								{/if}
							</div>
						</fieldset>

						<!-- Event Types Section -->
						<fieldset>
							<legend class="mb-3 block text-sm font-medium text-gray-700">Event types to sync:</legend>
							<div class="flex flex-wrap gap-4">
								<label class="flex cursor-pointer items-center gap-2">
									<input
										type="checkbox"
										bind:checked={eventTypes.posts}
										class="h-4 w-4 rounded border-gray-300 text-purple-600"
									/>
									<span class="text-sm text-gray-700">Posts</span>
								</label>
								<label class="flex cursor-pointer items-center gap-2">
									<input
										type="checkbox"
										bind:checked={eventTypes.reactions}
										class="h-4 w-4 rounded border-gray-300 text-purple-600"
									/>
									<span class="text-sm text-gray-700">Reactions</span>
								</label>
								<label class="flex cursor-pointer items-center gap-2">
									<input
										type="checkbox"
										bind:checked={eventTypes.reposts}
										class="h-4 w-4 rounded border-gray-300 text-purple-600"
									/>
									<span class="text-sm text-gray-700">Reposts</span>
								</label>
								<label class="flex cursor-pointer items-center gap-2">
									<input
										type="checkbox"
										bind:checked={eventTypes.profile}
										class="h-4 w-4 rounded border-gray-300 text-purple-600"
									/>
									<span class="text-sm text-gray-700">Profile</span>
								</label>
								<label class="flex cursor-pointer items-center gap-2">
									<input
										type="checkbox"
										bind:checked={eventTypes.follows}
										class="h-4 w-4 rounded border-gray-300 text-purple-600"
									/>
									<span class="text-sm text-gray-700">Follows</span>
								</label>
								<label class="flex cursor-pointer items-center gap-2">
									<input
										type="checkbox"
										bind:checked={eventTypes.dms}
										class="h-4 w-4 rounded border-gray-300 text-purple-600"
									/>
									<span class="text-sm text-gray-700">DMs</span>
								</label>
							</div>
							<p class="mt-2 text-xs text-gray-500">DMs are end-to-end encrypted and safe to sync.</p>
						</fieldset>
					</div>
				{/if}

			{:else if phase === 'syncing'}
				<div class="space-y-6">
					<!-- Progress bar -->
					<div class="space-y-2">
						<div class="flex items-center justify-between text-sm">
							<span class="text-gray-600">Progress</span>
							<span class="font-medium text-gray-900">{progressPercent()}%</span>
						</div>
						<div class="h-3 w-full rounded-full bg-gray-200">
							<div
								class="h-3 rounded-full bg-purple-600 transition-all duration-300"
								style="width: {progressPercent()}%"
							></div>
						</div>
					</div>

					<!-- Stats -->
					{#if syncProgress}
						<div class="rounded-lg bg-gray-50 p-4">
							<div class="grid grid-cols-3 gap-2 sm:gap-4 text-center">
								<div>
									<div class="text-xl sm:text-2xl font-bold text-gray-900">
										{formatCompactNumber(syncProgress.events_fetched || 0)}
									</div>
									<div class="text-xs text-gray-500">Fetched</div>
								</div>
								<div>
									<div class="text-xl sm:text-2xl font-bold text-green-600">
										{formatCompactNumber(syncProgress.events_stored || 0)}
									</div>
									<div class="text-xs text-gray-500">New</div>
								</div>
								<div>
									<div class="text-xl sm:text-2xl font-bold text-gray-400">
										{formatCompactNumber(syncProgress.events_skipped || 0)}
									</div>
									<div class="text-xs text-gray-500">Duplicates</div>
								</div>
							</div>
						</div>
					{/if}

					<p class="text-center text-sm text-gray-500">
						This may take a few minutes depending on history size.
					</p>
				</div>

			{:else if phase === 'complete'}
				<div class="space-y-6">
					{#if syncResult?.status === 'completed'}
						<div class="text-center">
							<div class="mx-auto mb-4 flex h-16 w-16 items-center justify-center rounded-full bg-green-100">
								<svg class="h-8 w-8 text-green-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
								</svg>
							</div>
							<p class="text-gray-600">Successfully imported your Nostr history.</p>
						</div>
					{:else if syncResult?.status === 'cancelled'}
						<div class="text-center">
							<div class="mx-auto mb-4 flex h-16 w-16 items-center justify-center rounded-full bg-yellow-100">
								<svg class="h-8 w-8 text-yellow-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
								</svg>
							</div>
							<p class="text-gray-600">Sync was cancelled.</p>
						</div>
					{:else}
						<div class="text-center">
							<div class="mx-auto mb-4 flex h-16 w-16 items-center justify-center rounded-full bg-red-100">
								<svg class="h-8 w-8 text-red-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
								</svg>
							</div>
							<p class="text-gray-600">Sync failed: {syncResult?.error_message || 'Unknown error'}</p>
						</div>
					{/if}

					<!-- Summary -->
					{#if syncResult}
						<div class="rounded-lg border border-gray-200 bg-gray-50 p-4">
							<h3 class="mb-3 text-sm font-medium text-gray-900">Summary</h3>
							<div class="space-y-2 text-sm">
								<div class="flex justify-between">
									<span class="text-gray-600">Total events fetched:</span>
									<span class="font-medium text-gray-900">{(syncResult.events_fetched || 0).toLocaleString()}</span>
								</div>
								<div class="flex justify-between">
									<span class="text-gray-600">New events saved:</span>
									<span class="font-medium text-green-600">{(syncResult.events_stored || 0).toLocaleString()}</span>
								</div>
								<div class="flex justify-between">
									<span class="text-gray-600">Already existed:</span>
									<span class="font-medium text-gray-500">{(syncResult.events_skipped || 0).toLocaleString()}</span>
								</div>
							</div>
						</div>
					{/if}

					<p class="text-center text-sm text-gray-500">
						Your relay now has a backup of your Nostr activity!
					</p>
				</div>
			{/if}
		</div>

		<!-- Footer -->
		<div class="flex justify-end space-x-3 border-t px-6 py-4">
			{#if phase === 'config'}
				<Button variant="secondary" onclick={onClose}>Cancel</Button>
				<Button
					variant="primary"
					onclick={startSync}
					disabled={!canStartSync || syncing}
					loading={syncing}
				>
					Start Sync
				</Button>
			{:else if phase === 'syncing'}
				<Button variant="secondary" onclick={cancelSync}>Cancel</Button>
				<Button variant="primary" onclick={sendToBackground}>Background</Button>
			{:else}
				<Button variant="primary" onclick={onClose}>Done</Button>
			{/if}
		</div>
	</div>
</div>
