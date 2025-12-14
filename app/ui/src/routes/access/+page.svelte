<script>
	import { browser } from '$app/environment';
	import { access, pricing, lightning, paidUsers } from '$lib/api/client.js';
	import { notify } from '$lib/stores/app.svelte.js';
	import AccessModeSelector from '$lib/components/access/AccessModeSelector.svelte';
	import PubkeyCard from '$lib/components/access/PubkeyCard.svelte';
	import AddPubkeyModal from '$lib/components/access/AddPubkeyModal.svelte';
	import EditNicknameModal from '$lib/components/access/EditNicknameModal.svelte';
	import ConfirmRemoveModal from '$lib/components/access/ConfirmRemoveModal.svelte';
	import PricingSection from '$lib/components/access/PricingSection.svelte';
	import LightningSection from '$lib/components/access/LightningSection.svelte';
	import PaidUsersSection from '$lib/components/access/PaidUsersSection.svelte';
	import RevenueCard from '$lib/components/access/RevenueCard.svelte';
	import Button from '$lib/components/Button.svelte';

	let mode = $state('whitelist');
	let whitelist = $state([]);
	let blacklist = $state([]);
	let pricingTiers = $state([]);
	let lightningStatus = $state(null);
	let paidUsersList = $state([]);
	let paidUsersTotal = $state(0);
	let revenueData = $state(null);
	let loading = $state(true);
	let error = $state(null);
	let initialized = $state(false);

	// Modal states
	let showAddModal = $state(false);
	let showEditModal = $state(false);
	let showRemoveModal = $state(false);
	let selectedEntry = $state(null);
	let activeListType = $state('whitelist');

	// Import state
	let importing = $state(false);
	let fileInput = $state(null);

	async function loadData() {
		try {
			const [modeRes, whitelistRes, blacklistRes] = await Promise.all([
				access.getMode(),
				access.getWhitelist(),
				access.getBlacklist()
			]);
			mode = modeRes.mode;
			whitelist = whitelistRes.entries || [];
			blacklist = blacklistRes.entries || [];

			// Load paid access data if mode is "paid"
			if (mode === 'paid') {
				await loadPaidAccessData();
			}
		} catch (e) {
			error = e.message;
		} finally {
			loading = false;
		}
	}

	async function loadPaidAccessData() {
		// Load each resource independently so one failure doesn't block others
		try {
			const pricingRes = await pricing.get();
			pricingTiers = pricingRes.tiers || [];
		} catch (e) {
			console.error('Failed to load pricing:', e);
		}

		try {
			const lightningRes = await lightning.getStatus();
			lightningStatus = lightningRes;
		} catch (e) {
			// Lightning not configured is expected, set a default status
			lightningStatus = { connected: false, error: e.message };
		}

		try {
			const paidUsersRes = await paidUsers.list({ limit: 100 });
			paidUsersList = paidUsersRes.users || [];
			paidUsersTotal = paidUsersRes.total || 0;
		} catch (e) {
			console.error('Failed to load paid users:', e);
		}

		try {
			const revenueRes = await paidUsers.getRevenue();
			revenueData = revenueRes;
		} catch (e) {
			console.error('Failed to load revenue:', e);
		}
	}

	// Load data on mount (browser only)
	$effect(() => {
		if (browser && !initialized) {
			initialized = true;
			loadData();
		}
	});

	async function handleModeChange(newMode) {
		mode = newMode;
		// Load paid access data when switching to paid mode
		if (newMode === 'paid' && pricingTiers.length === 0) {
			await loadPaidAccessData();
		}
	}

	// Whitelist actions
	function openAddWhitelist() {
		activeListType = 'whitelist';
		showAddModal = true;
	}

	function openEditNickname(entry) {
		selectedEntry = entry;
		showEditModal = true;
	}

	function openRemoveWhitelist(entry) {
		selectedEntry = entry;
		activeListType = 'whitelist';
		showRemoveModal = true;
	}

	// Blacklist actions
	function openAddBlacklist() {
		activeListType = 'blacklist';
		showAddModal = true;
	}

	function openRemoveBlacklist(entry) {
		selectedEntry = entry;
		activeListType = 'blacklist';
		showRemoveModal = true;
	}

	// Modal callbacks
	function handleAddComplete() {
		loadData();
	}

	function handleEditComplete() {
		loadData();
	}

	function handleRemoveComplete() {
		loadData();
	}

	function closeModals() {
		showAddModal = false;
		showEditModal = false;
		showRemoveModal = false;
		selectedEntry = null;
	}

	// Import/Export functionality
	async function handleImport(e) {
		const file = e.target.files?.[0];
		if (!file) return;

		importing = true;
		const reader = new FileReader();

		reader.onload = async (event) => {
			try {
				const content = event.target.result;
				let entries = [];

				// Try to parse as JSON first
				try {
					const json = JSON.parse(content);
					if (Array.isArray(json)) {
						entries = json;
					} else if (json.entries) {
						entries = json.entries;
					}
				} catch {
					// Fallback to newline-separated npubs
					entries = content.split('\n').filter((line) => line.trim()).map((line) => ({ npub: line.trim() }));
				}

				let added = 0;
				let failed = 0;

				for (const entry of entries) {
					try {
						const npub = entry.npub || entry;
						if (typeof npub !== 'string' || !npub.startsWith('npub')) continue;

						// Use the validate endpoint to get the pubkey
						const validation = await fetch(`/api/v1/setup/validate-identity?input=${encodeURIComponent(npub)}`);
						const result = await validation.json();

						if (result.valid) {
							await access.addToWhitelist({
								pubkey: result.pubkey,
								npub: result.npub,
								nickname: entry.nickname || ''
							});
							added++;
						} else {
							failed++;
						}
					} catch {
						failed++;
					}
				}

				notify('success', `Imported ${added} entries${failed > 0 ? `, ${failed} failed` : ''}`);
				loadData();
			} catch (e) {
				notify('error', 'Failed to parse import file');
			} finally {
				importing = false;
				// Reset file input
				if (fileInput) fileInput.value = '';
			}
		};

		reader.onerror = () => {
			notify('error', 'Failed to read file');
			importing = false;
		};

		reader.readAsText(file);
	}

	function handleExport() {
		const exportData = {
			exported_at: new Date().toISOString(),
			entries: whitelist.map((e) => ({
				npub: e.npub,
				pubkey: e.pubkey,
				nickname: e.nickname || ''
			}))
		};

		const blob = new Blob([JSON.stringify(exportData, null, 2)], { type: 'application/json' });
		const url = URL.createObjectURL(blob);
		const a = document.createElement('a');
		a.href = url;
		a.download = `whitelist-${new Date().toISOString().split('T')[0]}.json`;
		document.body.appendChild(a);
		a.click();
		document.body.removeChild(a);
		URL.revokeObjectURL(url);
		notify('success', 'Whitelist exported');
	}

	// Show list based on mode
	const showWhitelist = $derived(mode === 'whitelist' || mode === 'paid');
	const showBlacklist = $derived(mode === 'blacklist');
</script>

<div class="space-y-6">
	<div>
		<h1 class="text-2xl font-bold text-gray-900">Access Control</h1>
		<p class="text-gray-600">Manage who can write to your relay</p>
	</div>

	{#if loading}
		<div class="flex items-center justify-center py-12">
			<div class="h-8 w-8 animate-spin rounded-full border-4 border-purple-600 border-t-transparent"></div>
		</div>
	{:else if error}
		<div class="rounded-lg bg-red-50 p-4 text-red-700">
			<p class="font-medium">Error loading access control</p>
			<p class="text-sm">{error}</p>
			<button onclick={loadData} class="mt-2 text-sm font-medium text-red-600 hover:text-red-500">
				Try again
			</button>
		</div>
	{:else}
		<!-- Access Mode Selector -->
		<div class="rounded-lg bg-white p-6 shadow">
			<h2 class="text-lg font-semibold text-gray-900 mb-4">Access Mode</h2>
			<AccessModeSelector {mode} onChange={handleModeChange} />
		</div>

		<!-- Paid Access Sections (only when mode is "paid") -->
		{#if mode === 'paid'}
			<!-- Lightning Node Configuration -->
			<LightningSection status={lightningStatus} onUpdate={loadPaidAccessData} />

			<!-- Pricing Configuration -->
			<PricingSection tiers={pricingTiers} onUpdate={loadPaidAccessData} />

			<!-- Revenue Summary -->
			<RevenueCard revenue={revenueData} />

			<!-- Paid Users -->
			<PaidUsersSection users={paidUsersList} total={paidUsersTotal} onUpdate={loadPaidAccessData} />
		{/if}

		<!-- Whitelist Section -->
		{#if showWhitelist}
			<div class="rounded-lg bg-white p-6 shadow">
				<div class="flex items-center justify-between mb-4">
					<div>
						<h2 class="text-lg font-semibold text-gray-900">
							Whitelisted Pubkeys
							<span class="ml-2 text-sm font-normal text-gray-500">({whitelist.length})</span>
						</h2>
						{#if mode === 'paid'}
							<p class="text-sm text-gray-500 mt-1">These users plus paid subscribers can write to your relay.</p>
						{/if}
					</div>
					<Button variant="primary" onclick={openAddWhitelist}>
						<svg class="w-4 h-4 mr-1" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
						</svg>
						Add
					</Button>
				</div>

				{#if whitelist.length === 0}
					<div class="text-center py-8">
						<div class="w-12 h-12 mx-auto bg-gray-100 rounded-full flex items-center justify-center mb-3">
							<svg class="w-6 h-6 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0zm6 3a2 2 0 11-4 0 2 2 0 014 0zM7 10a2 2 0 11-4 0 2 2 0 014 0z" />
							</svg>
						</div>
						<p class="text-gray-500">No pubkeys whitelisted yet</p>
						<p class="text-sm text-gray-400 mt-1">Add pubkeys to allow them to write to your relay</p>
					</div>
				{:else}
					<div class="space-y-2">
						{#each whitelist as entry (entry.pubkey)}
							<PubkeyCard
								{entry}
								listType="whitelist"
								onEdit={openEditNickname}
								onRemove={openRemoveWhitelist}
							/>
						{/each}
					</div>
				{/if}

				<!-- Import/Export -->
				<div class="flex items-center space-x-3 mt-6 pt-4 border-t">
					<input
						type="file"
						accept=".json,.txt"
						onchange={handleImport}
						bind:this={fileInput}
						class="hidden"
						id="import-input"
					/>
					<label
						for="import-input"
						class="inline-flex items-center px-3 py-2 text-sm font-medium text-gray-700 bg-white border border-gray-300 rounded-lg hover:bg-gray-50 cursor-pointer {importing ? 'opacity-50 pointer-events-none' : ''}"
					>
						{#if importing}
							<div class="w-4 h-4 mr-2 animate-spin rounded-full border-2 border-gray-600 border-t-transparent"></div>
						{:else}
							<svg class="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-8l-4-4m0 0L8 8m4-4v12" />
							</svg>
						{/if}
						Import
					</label>
					<button
						type="button"
						onclick={handleExport}
						disabled={whitelist.length === 0}
						class="inline-flex items-center px-3 py-2 text-sm font-medium text-gray-700 bg-white border border-gray-300 rounded-lg hover:bg-gray-50 disabled:opacity-50 disabled:cursor-not-allowed"
					>
						<svg class="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-4l-4 4m0 0l-4-4m4 4V4" />
						</svg>
						Export
					</button>
				</div>
			</div>
		{/if}

		<!-- Blacklist Section -->
		{#if showBlacklist}
			<div class="rounded-lg bg-white p-6 shadow">
				<div class="flex items-center justify-between mb-4">
					<div>
						<h2 class="text-lg font-semibold text-gray-900">
							Blacklisted Pubkeys
							<span class="ml-2 text-sm font-normal text-gray-500">({blacklist.length})</span>
						</h2>
						<p class="text-sm text-gray-500 mt-1">These users are blocked from writing to your relay.</p>
					</div>
					<Button variant="primary" onclick={openAddBlacklist}>
						<svg class="w-4 h-4 mr-1" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
						</svg>
						Add
					</Button>
				</div>

				{#if blacklist.length === 0}
					<div class="text-center py-8">
						<div class="w-12 h-12 mx-auto bg-gray-100 rounded-full flex items-center justify-center mb-3">
							<svg class="w-6 h-6 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M18.364 18.364A9 9 0 005.636 5.636m12.728 12.728A9 9 0 015.636 5.636m12.728 12.728L5.636 5.636" />
							</svg>
						</div>
						<p class="text-gray-500">No pubkeys blacklisted</p>
						<p class="text-sm text-gray-400 mt-1">Add pubkeys to block them from your relay</p>
					</div>
				{:else}
					<div class="space-y-2">
						{#each blacklist as entry (entry.pubkey)}
							<PubkeyCard
								{entry}
								listType="blacklist"
								onRemove={openRemoveBlacklist}
							/>
						{/each}
					</div>
				{/if}
			</div>
		{/if}

		<!-- Open mode info -->
		{#if mode === 'open'}
			<div class="rounded-lg bg-amber-50 border border-amber-200 p-4">
				<div class="flex items-start space-x-3">
					<svg class="w-5 h-5 text-amber-500 mt-0.5 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
					</svg>
					<div>
						<p class="font-medium text-amber-800">Open Mode Active</p>
						<p class="text-sm text-amber-700 mt-1">
							Anyone can write to your relay. This is not recommended for private relays as it may fill up your storage with unwanted content.
						</p>
					</div>
				</div>
			</div>
		{/if}
	{/if}
</div>

<!-- Modals -->
{#if showAddModal}
	<AddPubkeyModal
		listType={activeListType}
		onClose={closeModals}
		onAdd={handleAddComplete}
	/>
{/if}

{#if showEditModal && selectedEntry}
	<EditNicknameModal
		entry={selectedEntry}
		onClose={closeModals}
		onSave={handleEditComplete}
	/>
{/if}

{#if showRemoveModal && selectedEntry}
	<ConfirmRemoveModal
		entry={selectedEntry}
		listType={activeListType}
		onClose={closeModals}
		onConfirm={handleRemoveComplete}
	/>
{/if}
