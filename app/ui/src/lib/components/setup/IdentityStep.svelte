<script>
	import { setup } from '$lib/api/client.js';

	let { identity = '', onChange } = $props();

	let inputValue = $state('');
	let validating = $state(false);
	let validationResult = $state(null);
	let debounceTimer = null;

	// Sync local state with prop (handles initial value and back navigation)
	$effect(() => {
		if (identity && identity !== inputValue) {
			inputValue = identity;
		}
	});

	// Cleanup on unmount
	$effect(() => {
		return () => {
			if (debounceTimer) clearTimeout(debounceTimer);
		};
	});

	// Debounced validation
	function handleInput(e) {
		inputValue = e.target.value;

		// Clear previous timer
		if (debounceTimer) {
			clearTimeout(debounceTimer);
		}

		// Clear validation if empty
		if (!inputValue.trim()) {
			validationResult = null;
			onChange({ identity: '', pubkey: '', npub: '', valid: false });
			return;
		}

		// Debounce validation
		debounceTimer = setTimeout(async () => {
			await validateIdentity();
		}, 500);
	}

	async function validateIdentity() {
		if (!inputValue.trim()) return;

		validating = true;
		try {
			const result = await setup.validateIdentity(inputValue.trim());
			validationResult = result;

			if (result.valid) {
				onChange({
					identity: inputValue.trim(),
					pubkey: result.pubkey,
					npub: result.npub,
					valid: true
				});
			} else {
				onChange({ identity: inputValue.trim(), pubkey: '', npub: '', valid: false });
			}
		} catch (e) {
			validationResult = { valid: false, error: e.message };
			onChange({ identity: inputValue.trim(), pubkey: '', npub: '', valid: false });
		} finally {
			validating = false;
		}
	}
</script>

<div>
	<h2 class="text-2xl font-bold text-gray-900 dark:text-white mb-2">Your Identity</h2>
	<p class="text-gray-600 dark:text-gray-400 mb-6">
		Enter your Nostr public key (npub) or NIP-05 identifier. You'll be automatically whitelisted as
		the relay operator.
	</p>

	<!-- Input field -->
	<div class="mb-4">
		<label for="identity" class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
			Public Key or NIP-05
		</label>
		<div class="relative">
			<input
				type="text"
				id="identity"
				value={inputValue}
				oninput={handleInput}
				placeholder="npub1... or you@domain.com"
				class="input w-full pr-10"
				class:border-green-500={validationResult?.valid}
				class:border-red-500={validationResult && !validationResult.valid}
			/>
			{#if validating}
				<div class="absolute right-3 top-1/2 -translate-y-1/2">
					<div class="w-5 h-5 border-2 border-purple-600 border-t-transparent rounded-full animate-spin"></div>
				</div>
			{:else if validationResult?.valid}
				<div class="absolute right-3 top-1/2 -translate-y-1/2">
					<svg class="w-5 h-5 text-green-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
					</svg>
				</div>
			{:else if validationResult && !validationResult.valid}
				<div class="absolute right-3 top-1/2 -translate-y-1/2">
					<svg class="w-5 h-5 text-red-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
					</svg>
				</div>
			{/if}
		</div>
	</div>

	<!-- Validation feedback -->
	{#if validationResult}
		{#if validationResult.valid}
			<div class="p-3 bg-green-50 dark:bg-green-900/30 border border-green-200 dark:border-green-700 rounded-lg mb-4">
				<div class="flex items-start space-x-2">
					<svg class="w-5 h-5 text-green-500 dark:text-green-400 mt-0.5 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
					</svg>
					<div class="text-sm">
						<p class="font-medium text-green-800 dark:text-green-300">Valid {validationResult.source === 'nip05' ? 'NIP-05' : 'pubkey'}</p>
						<p class="text-green-700 dark:text-green-400 font-mono text-xs break-all mt-1">
							{validationResult.npub}
						</p>
						{#if validationResult.nip05_name}
							<p class="text-green-600 dark:text-green-400 mt-1">Resolved from: {validationResult.nip05_name}</p>
						{/if}
					</div>
				</div>
			</div>
		{:else}
			<div class="p-3 bg-red-50 dark:bg-red-900/30 border border-red-200 dark:border-red-700 rounded-lg mb-4">
				<div class="flex items-start space-x-2">
					<svg class="w-5 h-5 text-red-500 dark:text-red-400 mt-0.5 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
					</svg>
					<div class="text-sm">
						<p class="font-medium text-red-800 dark:text-red-300">Invalid identity</p>
						<p class="text-red-700 dark:text-red-400 mt-1">{validationResult.error || 'Please enter a valid npub, hex pubkey, or NIP-05 identifier.'}</p>
					</div>
				</div>
			</div>
		{/if}
	{/if}

	<!-- Info note -->
	<div class="p-3 bg-blue-50 dark:bg-blue-900/30 border border-blue-200 dark:border-blue-700 rounded-lg">
		<div class="flex items-start space-x-2">
			<svg class="w-5 h-5 text-blue-500 dark:text-blue-400 mt-0.5 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
				<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
			</svg>
			<p class="text-sm text-blue-700 dark:text-blue-300">
				This pubkey will have full access to your relay and cannot be removed without resetting Roostr.
			</p>
		</div>
	</div>
</div>
