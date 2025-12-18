<script>
	let { relayUrl = '', torUrl = '', onDone = () => {} } = $props();

	let copiedRelay = $state(false);
	let copiedTor = $state(false);

	async function copyUrl(url, type) {
		try {
			await navigator.clipboard.writeText(url);
			if (type === 'relay') {
				copiedRelay = true;
				setTimeout(() => (copiedRelay = false), 2000);
			} else {
				copiedTor = true;
				setTimeout(() => (copiedTor = false), 2000);
			}
		} catch (e) {
			console.error('Failed to copy:', e);
		}
	}
</script>

<div class="rounded-lg bg-white dark:bg-gray-800 p-8 shadow dark:shadow-gray-900/50 text-center">
	<!-- Success Animation -->
	<div class="w-20 h-20 mx-auto mb-6 bg-green-100 dark:bg-green-900/30 rounded-full flex items-center justify-center">
		<svg class="w-10 h-10 text-green-600 dark:text-green-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
			<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
		</svg>
	</div>

	<h2 class="text-2xl font-bold text-gray-900 dark:text-gray-100 mb-2">Payment Confirmed!</h2>
	<p class="text-gray-600 dark:text-gray-400 mb-8">Your relay access has been activated.</p>

	<!-- Relay URLs -->
	<div class="space-y-4 mb-8">
		<div class="text-left">
			<p class="text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">Add this relay to your Nostr client:</p>

			{#if relayUrl}
				<div class="flex items-center justify-between p-3 bg-gray-50 dark:bg-gray-700 rounded-lg border border-gray-200 dark:border-gray-600 mb-2">
					<code class="text-sm text-gray-900 dark:text-gray-100 font-mono truncate mr-2">{relayUrl}</code>
					<button
						type="button"
						onclick={() => copyUrl(relayUrl, 'relay')}
						class="flex-shrink-0 px-3 py-1.5 text-sm font-medium rounded-lg transition-colors {copiedRelay ? 'bg-green-100 dark:bg-green-900/30 text-green-700 dark:text-green-300' : 'bg-purple-100 dark:bg-purple-900/30 hover:bg-purple-200 dark:hover:bg-purple-900/50 text-purple-700 dark:text-purple-300'}"
					>
						{copiedRelay ? 'Copied!' : 'Copy'}
					</button>
				</div>
			{/if}

			{#if torUrl}
				<div class="flex items-center justify-between p-3 bg-gray-50 dark:bg-gray-700 rounded-lg border border-gray-200 dark:border-gray-600">
					<div class="flex items-center mr-2 min-w-0">
						<span class="flex-shrink-0 px-2 py-0.5 text-xs font-medium bg-purple-100 dark:bg-purple-900/30 text-purple-700 dark:text-purple-300 rounded mr-2">Tor</span>
						<code class="text-sm text-gray-900 dark:text-gray-100 font-mono truncate">{torUrl}</code>
					</div>
					<button
						type="button"
						onclick={() => copyUrl(torUrl, 'tor')}
						class="flex-shrink-0 px-3 py-1.5 text-sm font-medium rounded-lg transition-colors {copiedTor ? 'bg-green-100 dark:bg-green-900/30 text-green-700 dark:text-green-300' : 'bg-purple-100 dark:bg-purple-900/30 hover:bg-purple-200 dark:hover:bg-purple-900/50 text-purple-700 dark:text-purple-300'}"
					>
						{copiedTor ? 'Copied!' : 'Copy'}
					</button>
				</div>
			{/if}
		</div>
	</div>

	<!-- Instructions -->
	<div class="bg-purple-50 dark:bg-purple-900/20 rounded-lg p-4 mb-8 text-left border border-purple-100 dark:border-purple-800">
		<h3 class="font-medium text-purple-900 dark:text-purple-100 mb-2">Next steps:</h3>
		<ol class="space-y-1 text-sm text-purple-700 dark:text-purple-300 list-decimal list-inside">
			<li>Open your favorite Nostr client</li>
			<li>Go to relay settings</li>
			<li>Add the relay URL above</li>
			<li>Start posting!</li>
		</ol>
	</div>

	<button
		type="button"
		onclick={onDone}
		class="px-6 py-2.5 bg-gray-900 dark:bg-gray-700 hover:bg-gray-800 dark:hover:bg-gray-600 text-white font-medium rounded-lg transition-colors"
	>
		Done
	</button>
</div>
