<script>
	import { browser } from '$app/environment';
	import { goto } from '$app/navigation';
	import { relay } from '$lib/api/client.js';
	import { notify } from '$lib/stores';
	import Button from '$lib/components/Button.svelte';
	import Loading from '$lib/components/Loading.svelte';

	let loading = $state(true);
	let initialized = $state(false);
	let urls = $state({ local: '', tor: '' });
	let qrLocal = $state('');
	let qrTor = $state('');
	let showQrLocal = $state(false);
	let showQrTor = $state(false);

	$effect(() => {
		if (browser && !initialized) {
			initialized = true;
			loadRelayInfo();
		}
	});

	async function loadRelayInfo() {
		try {
			// Get relay URLs from API
			const status = await relay.getStatus();
			urls = {
				local: status.websocket_url || 'ws://localhost:4848',
				tor: status.tor_url || ''
			};

			// Generate QR codes
			const QRCode = (await import('qrcode')).default;

			if (urls.local) {
				qrLocal = await QRCode.toDataURL(urls.local, {
					width: 200,
					margin: 2,
					color: { dark: '#1f2937', light: '#ffffff' }
				});
			}

			if (urls.tor) {
				qrTor = await QRCode.toDataURL(urls.tor, {
					width: 200,
					margin: 2,
					color: { dark: '#1f2937', light: '#ffffff' }
				});
			}
		} catch (e) {
			console.error('Failed to load relay status:', e);
			// Use fallback values
			urls = {
				local: 'ws://localhost:4848',
				tor: ''
			};
		} finally {
			loading = false;
		}
	}

	async function copyToClipboard(text) {
		try {
			await navigator.clipboard.writeText(text);
			notify('success', 'Copied to clipboard');
		} catch {
			notify('error', 'Failed to copy');
		}
	}

	function goToDashboard() {
		goto('/');
	}
</script>

<div class="text-center">
	<!-- Success icon -->
	<div class="mb-6">
		<div class="inline-flex items-center justify-center w-20 h-20 bg-green-100 dark:bg-green-900/50 rounded-full">
			<svg class="w-10 h-10 text-green-600 dark:text-green-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
				<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
			</svg>
		</div>
	</div>

	<h2 class="text-3xl font-bold text-gray-900 dark:text-white mb-2">You're All Set!</h2>
	<p class="text-gray-600 dark:text-gray-400 mb-8">Your private relay is ready to use.</p>

	{#if loading}
		<div class="py-8">
			<Loading text="Loading relay information..." />
		</div>
	{:else}
		<!-- Relay URLs -->
		<div class="bg-gray-50 dark:bg-gray-800 rounded-lg p-6 mb-6 text-left">
			<h3 class="text-sm font-semibold text-gray-700 dark:text-gray-300 uppercase tracking-wide mb-4">Your Relay URLs</h3>

			<!-- Local URL -->
			<div class="mb-4">
				<label for="local-url" class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">Local Network</label>
				<div class="flex items-center space-x-2">
					<input
						type="text"
						id="local-url"
						value={urls.local}
						readonly
						class="input flex-1 font-mono text-sm bg-white dark:bg-gray-700"
					/>
					<button
						type="button"
						onclick={() => copyToClipboard(urls.local)}
						class="inline-flex items-center justify-center p-2 rounded-lg bg-gray-100 dark:bg-gray-700 text-gray-700 dark:text-gray-300 hover:bg-gray-200 dark:hover:bg-gray-600 transition-colors"
						title="Copy"
					>
						<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 5H6a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2v-1M8 5a2 2 0 002 2h2a2 2 0 002-2M8 5a2 2 0 012-2h2a2 2 0 012 2m0 0h2a2 2 0 012 2v3m2 4H10m0 0l3-3m-3 3l3 3" />
						</svg>
					</button>
					<button
						type="button"
						onclick={() => (showQrLocal = !showQrLocal)}
						class="inline-flex items-center justify-center p-2 rounded-lg bg-gray-100 dark:bg-gray-700 text-gray-700 dark:text-gray-300 hover:bg-gray-200 dark:hover:bg-gray-600 transition-colors"
						title="QR Code"
					>
						<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v1m6 11h2m-6 0h-2v4m0-11v3m0 0h.01M12 12h4.01M16 20h4M4 12h4m12 0h.01M5 8h2a1 1 0 001-1V5a1 1 0 00-1-1H5a1 1 0 00-1 1v2a1 1 0 001 1zm12 0h2a1 1 0 001-1V5a1 1 0 00-1-1h-2a1 1 0 00-1 1v2a1 1 0 001 1zM5 20h2a1 1 0 001-1v-2a1 1 0 00-1-1H5a1 1 0 00-1 1v2a1 1 0 001 1z" />
						</svg>
					</button>
				</div>
				{#if showQrLocal && qrLocal}
					<div class="mt-3 flex justify-center">
						<img src={qrLocal} alt="Local URL QR Code" class="rounded-lg border border-gray-200 dark:border-gray-700" />
					</div>
				{/if}
			</div>

			<!-- Tor URL (if available) -->
			{#if urls.tor}
				<div>
					<label for="tor-url" class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">Tor (Remote Access)</label>
					<div class="flex items-center space-x-2">
						<input
							type="text"
							id="tor-url"
							value={urls.tor}
							readonly
							class="input flex-1 font-mono text-sm bg-white dark:bg-gray-700"
						/>
						<button
							type="button"
							onclick={() => copyToClipboard(urls.tor)}
							class="inline-flex items-center justify-center p-2 rounded-lg bg-gray-100 dark:bg-gray-700 text-gray-700 dark:text-gray-300 hover:bg-gray-200 dark:hover:bg-gray-600 transition-colors"
							title="Copy"
						>
							<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 5H6a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2v-1M8 5a2 2 0 002 2h2a2 2 0 002-2M8 5a2 2 0 012-2h2a2 2 0 012 2m0 0h2a2 2 0 012 2v3m2 4H10m0 0l3-3m-3 3l3 3" />
							</svg>
						</button>
						<button
							type="button"
							onclick={() => (showQrTor = !showQrTor)}
							class="inline-flex items-center justify-center p-2 rounded-lg bg-gray-100 dark:bg-gray-700 text-gray-700 dark:text-gray-300 hover:bg-gray-200 dark:hover:bg-gray-600 transition-colors"
							title="QR Code"
						>
							<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v1m6 11h2m-6 0h-2v4m0-11v3m0 0h.01M12 12h4.01M16 20h4M4 12h4m12 0h.01M5 8h2a1 1 0 001-1V5a1 1 0 00-1-1H5a1 1 0 00-1 1v2a1 1 0 001 1zm12 0h2a1 1 0 001-1V5a1 1 0 00-1-1h-2a1 1 0 00-1 1v2a1 1 0 001 1zM5 20h2a1 1 0 001-1v-2a1 1 0 00-1-1H5a1 1 0 00-1 1v2a1 1 0 001 1z" />
							</svg>
						</button>
					</div>
					{#if showQrTor && qrTor}
						<div class="mt-3 flex justify-center">
							<img src={qrTor} alt="Tor URL QR Code" class="rounded-lg border border-gray-200 dark:border-gray-700" />
						</div>
					{/if}
				</div>
			{/if}
		</div>

		<!-- Next steps -->
		<div class="bg-blue-50 dark:bg-blue-900/30 rounded-lg p-6 mb-8 text-left">
			<h3 class="text-sm font-semibold text-gray-700 dark:text-gray-300 uppercase tracking-wide mb-4">Next Steps</h3>
			<ol class="space-y-3 text-sm text-gray-700 dark:text-gray-300">
				<li class="flex items-start space-x-3">
					<span class="flex-shrink-0 w-6 h-6 bg-blue-100 dark:bg-blue-900/50 text-blue-600 dark:text-blue-300 rounded-full flex items-center justify-center font-medium text-xs">1</span>
					<span>Add your relay URL to your Nostr client (Damus: Settings &rarr; Relays &rarr; Add Relay)</span>
				</li>
				<li class="flex items-start space-x-3">
					<span class="flex-shrink-0 w-6 h-6 bg-blue-100 dark:bg-blue-900/50 text-blue-600 dark:text-blue-300 rounded-full flex items-center justify-center font-medium text-xs">2</span>
					<span>Sync your existing posts from public relays (Click the Sync button on the dashboard)</span>
				</li>
				<li class="flex items-start space-x-3">
					<span class="flex-shrink-0 w-6 h-6 bg-blue-100 dark:bg-blue-900/50 text-blue-600 dark:text-blue-300 rounded-full flex items-center justify-center font-medium text-xs">3</span>
					<span>Share your Tor URL with whitelisted friends so they can connect remotely</span>
				</li>
			</ol>
		</div>

		<!-- Go to Dashboard button -->
		<Button variant="primary" onclick={goToDashboard}>
			Go to Dashboard
		</Button>
	{/if}
</div>
