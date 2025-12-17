<script>
	import { setup } from '$lib/api/client.js';
	import Button from '$lib/components/Button.svelte';

	let { operatorPubkey = '', operatorNpub = '', pubkeys = [], onChange } = $props();

	let inputValue = $state('');
	let nickname = $state('');
	let validating = $state(false);
	let error = $state('');

	async function addPubkey() {
		if (!inputValue.trim()) return;

		validating = true;
		error = '';

		try {
			const result = await setup.validateIdentity(inputValue.trim());

			if (!result.valid) {
				error = result.error || 'Invalid identity';
				return;
			}

			// Check if already in list
			if (result.pubkey === operatorPubkey) {
				error = 'This is your operator pubkey (already included)';
				return;
			}

			if (pubkeys.some((p) => p.pubkey === result.pubkey)) {
				error = 'This pubkey is already in the list';
				return;
			}

			// Add to list
			const newPubkeys = [
				...pubkeys,
				{
					pubkey: result.pubkey,
					npub: result.npub,
					nickname: nickname.trim() || '',
					source: result.source,
					nip05Name: result.nip05_name || ''
				}
			];

			onChange(newPubkeys);

			// Clear inputs
			inputValue = '';
			nickname = '';
		} catch (e) {
			error = e.message || 'Failed to validate identity';
		} finally {
			validating = false;
		}
	}

	function removePubkey(pubkey) {
		const newPubkeys = pubkeys.filter((p) => p.pubkey !== pubkey);
		onChange(newPubkeys);
	}

	function truncateNpub(npub) {
		if (!npub) return '';
		return npub.slice(0, 12) + '...' + npub.slice(-8);
	}
</script>

<div>
	<h2 class="text-2xl font-bold text-gray-900 dark:text-white mb-2">Add Others</h2>
	<p class="text-gray-600 dark:text-gray-400 mb-6">
		Want to whitelist anyone else right now? You can always add more people later.
	</p>

	<!-- Add pubkey form -->
	<div class="space-y-3 mb-6">
		<div>
			<label for="pubkey-input" class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
				npub or NIP-05
			</label>
			<input
				type="text"
				id="pubkey-input"
				bind:value={inputValue}
				placeholder="npub1... or user@domain.com"
				class="input w-full"
				disabled={validating}
			/>
		</div>
		<div>
			<label for="nickname-input" class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
				Nickname <span class="text-gray-400 dark:text-gray-500">(optional)</span>
			</label>
			<div class="flex space-x-2">
				<input
					type="text"
					id="nickname-input"
					bind:value={nickname}
					placeholder="e.g., Family, Friend"
					class="input flex-1"
					disabled={validating}
				/>
				<Button variant="primary" onclick={addPubkey} disabled={!inputValue.trim() || validating} loading={validating}>
					Add
				</Button>
			</div>
		</div>

		{#if error}
			<div class="p-3 bg-red-50 dark:bg-red-900/30 border border-red-200 dark:border-red-700 rounded-lg text-red-700 dark:text-red-300 text-sm">
				{error}
			</div>
		{/if}
	</div>

	<!-- Whitelisted list -->
	<div class="space-y-2">
		<p class="text-sm font-medium text-gray-700 dark:text-gray-300">Whitelisted:</p>

		<!-- Operator (always first, non-removable) -->
		<div class="flex items-center justify-between p-3 bg-purple-50 dark:bg-purple-900/30 border border-purple-200 dark:border-purple-700 rounded-lg">
			<div class="flex items-center space-x-3">
				<div class="w-8 h-8 bg-purple-200 dark:bg-purple-800 rounded-full flex items-center justify-center">
					<svg class="w-4 h-4 text-purple-600 dark:text-purple-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
					</svg>
				</div>
				<div>
					<p class="text-sm font-medium text-gray-900 dark:text-white">You (Operator)</p>
					<p class="text-xs text-gray-500 dark:text-gray-400 font-mono">{truncateNpub(operatorNpub)}</p>
				</div>
			</div>
			<span class="text-xs bg-purple-100 dark:bg-purple-900/50 text-purple-700 dark:text-purple-300 px-2 py-1 rounded-full">Operator</span>
		</div>

		<!-- Additional pubkeys -->
		{#each pubkeys as pubkey}
			<div class="flex items-center justify-between p-3 bg-gray-50 dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg">
				<div class="flex items-center space-x-3">
					<div class="w-8 h-8 bg-green-100 dark:bg-green-900/50 rounded-full flex items-center justify-center">
						<svg class="w-4 h-4 text-green-600 dark:text-green-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
						</svg>
					</div>
					<div>
						<p class="text-sm font-medium text-gray-900 dark:text-white">
							{pubkey.nickname || (pubkey.nip05Name ? pubkey.nip05Name : 'No nickname')}
						</p>
						<p class="text-xs text-gray-500 dark:text-gray-400 font-mono">{truncateNpub(pubkey.npub)}</p>
					</div>
				</div>
				<button
					type="button"
					onclick={() => removePubkey(pubkey.pubkey)}
					class="p-1 text-gray-400 dark:text-gray-500 hover:text-red-500 dark:hover:text-red-400 transition-colors"
					title="Remove"
				>
					<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
					</svg>
				</button>
			</div>
		{/each}

		{#if pubkeys.length === 0}
			<p class="text-sm text-gray-500 dark:text-gray-500 italic py-2">
				No additional users added yet. You can skip this and add people later.
			</p>
		{/if}
	</div>

	<!-- Tip -->
	<div class="mt-6 p-3 bg-yellow-50 dark:bg-yellow-900/30 border border-yellow-200 dark:border-yellow-700 rounded-lg">
		<div class="flex items-start space-x-2">
			<svg class="w-5 h-5 text-yellow-500 dark:text-yellow-400 mt-0.5 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
				<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9.663 17h4.673M12 3v1m6.364 1.636l-.707.707M21 12h-1M4 12H3m3.343-5.657l-.707-.707m2.828 9.9a5 5 0 117.072 0l-.548.547A3.374 3.374 0 0014 18.469V19a2 2 0 11-4 0v-.531c0-.895-.356-1.754-.988-2.386l-.548-.547z" />
			</svg>
			<p class="text-sm text-yellow-700 dark:text-yellow-300">
				<strong>Tip:</strong> Your whitelisted users should add your relay URL to their Nostr client to start backing up their events.
			</p>
		</div>
	</div>
</div>
