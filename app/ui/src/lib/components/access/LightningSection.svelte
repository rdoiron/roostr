<script>
	import { lightning } from '$lib/api/client.js';
	import { notify } from '$lib/stores/app.svelte.js';
	import Button from '$lib/components/Button.svelte';

	let { status = null, onUpdate = () => {} } = $props();

	let config = $state({
		host: '',
		macaroon_hex: '',
		tls_cert_path: ''
	});
	let saving = $state(false);
	let testing = $state(false);
	let testResult = $state(null);

	// Initialize config from status
	$effect(() => {
		if (status?.config) {
			config = {
				host: status.config.host || '',
				macaroon_hex: status.config.macaroon_hex || '',
				tls_cert_path: status.config.tls_cert_path || ''
			};
		}
	});

	async function handleTest() {
		testing = true;
		testResult = null;
		try {
			const normalizedConfig = { ...config, host: normalizeHost(config.host) };
			const result = await lightning.test(normalizedConfig);
			testResult = {
				success: result.success,
				alias: result.node_info?.alias,
				balance: result.node_info?.balance
			};
			if (result.success) {
				notify('success', `Connected to ${result.node_info?.alias || 'LND node'}`);
			} else {
				testResult.error = result.error || result.message || 'Connection failed';
				notify('error', testResult.error);
			}
		} catch (e) {
			testResult = {
				success: false,
				error: e.message || 'Connection failed'
			};
			notify('error', e.message || 'Connection test failed');
		} finally {
			testing = false;
		}
	}

	async function handleSave() {
		saving = true;
		try {
			const normalizedConfig = { ...config, host: normalizeHost(config.host) };
			await lightning.updateConfig(normalizedConfig);
			notify('success', 'Lightning configuration saved');
			onUpdate();
		} catch (e) {
			notify('error', e.message || 'Failed to save configuration');
		} finally {
			saving = false;
		}
	}

	function formatSats(sats) {
		if (sats >= 100000000) {
			return (sats / 100000000).toFixed(2) + ' BTC';
		}
		return sats.toLocaleString() + ' sats';
	}

	// Normalize host - strip protocol prefix and trailing slash
	function normalizeHost(host) {
		return host
			.replace(/^https?:\/\//i, '')
			.replace(/\/$/, '');
	}
</script>

<div class="rounded-lg bg-white dark:bg-gray-800 p-6 shadow dark:shadow-gray-900/50">
	<div class="flex items-center justify-between mb-4">
		<div>
			<h2 class="text-lg font-semibold text-gray-900 dark:text-gray-100">Lightning Node</h2>
			<p class="text-sm text-gray-500 dark:text-gray-400">Connect your LND node to accept Lightning payments</p>
		</div>
		{#if status?.connected}
			<span class="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-green-100 dark:bg-green-900/30 text-green-800 dark:text-green-400">
				<span class="w-2 h-2 mr-1.5 bg-green-500 rounded-full"></span>
				Connected
			</span>
		{:else}
			<span class="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-gray-100 dark:bg-gray-700 text-gray-600 dark:text-gray-400">
				<span class="w-2 h-2 mr-1.5 bg-gray-400 rounded-full"></span>
				Not Connected
			</span>
		{/if}
	</div>

	{#if status?.connected && status?.node_info}
		<div class="mb-6 p-4 bg-green-50 dark:bg-green-900/20 border border-green-200 dark:border-green-800 rounded-lg">
			<div class="flex items-center justify-between">
				<div>
					<p class="font-medium text-green-800 dark:text-green-300">{status.node_info.alias || 'LND Node'}</p>
					<p class="text-sm text-green-600 dark:text-green-400">Balance: {formatSats(status.balance?.local_balance || 0)}</p>
				</div>
				<svg class="w-8 h-8 text-green-500 dark:text-green-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 10V3L4 14h7v7l9-11h-7z" />
				</svg>
			</div>
		</div>
	{/if}

	<div class="space-y-4">
		<div>
			<label for="host" class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
				LND REST Endpoint
			</label>
			<input
				type="text"
				id="host"
				bind:value={config.host}
				placeholder="umbrel.local:8080 or 127.0.0.1:8080"
				class="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100 focus:ring-2 focus:ring-purple-500 focus:border-transparent"
			/>
			<p class="mt-1 text-xs text-gray-500 dark:text-gray-400">The REST API endpoint of your LND node</p>
		</div>

		<div>
			<label for="macaroon_hex" class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
				Admin Macaroon (hex)
			</label>
			<input
				type="password"
				id="macaroon_hex"
				bind:value={config.macaroon_hex}
				placeholder="0201036c6e6402..."
				class="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100 focus:ring-2 focus:ring-purple-500 focus:border-transparent font-mono text-sm"
			/>
			<p class="mt-1 text-xs text-gray-500 dark:text-gray-400">Your admin.macaroon file encoded as hex</p>
		</div>

		<div>
			<label for="tls_cert_path" class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
				TLS Certificate Path (optional)
			</label>
			<input
				type="text"
				id="tls_cert_path"
				bind:value={config.tls_cert_path}
				placeholder="/path/to/tls.cert"
				class="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100 focus:ring-2 focus:ring-purple-500 focus:border-transparent"
			/>
			<p class="mt-1 text-xs text-gray-500 dark:text-gray-400">Optional for local connections (Umbrel/Start9)</p>
		</div>
	</div>

	{#if testResult}
		<div class="mt-4 p-3 rounded-lg {testResult.success ? 'bg-green-50 dark:bg-green-900/20 border border-green-200 dark:border-green-800' : 'bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800'}">
			{#if testResult.success}
				<div class="flex items-center text-green-700 dark:text-green-400">
					<svg class="w-5 h-5 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
					</svg>
					<span class="font-medium">Connection successful!</span>
					{#if testResult.alias}
						<span class="ml-2 text-sm">({testResult.alias})</span>
					{/if}
				</div>
				{#if testResult.balance !== undefined}
					<p class="text-sm text-green-600 dark:text-green-400 mt-1 ml-7">Balance: {formatSats(testResult.balance)}</p>
				{/if}
			{:else}
				<div class="flex items-center text-red-700 dark:text-red-400">
					<svg class="w-5 h-5 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
					</svg>
					<span class="font-medium">Connection failed</span>
				</div>
				<p class="text-sm text-red-600 dark:text-red-400 mt-1 ml-7">{testResult.error}</p>
			{/if}
		</div>
	{/if}

	<div class="flex flex-col gap-3 mt-6 pt-4 border-t dark:border-gray-700 sm:flex-row sm:items-center sm:justify-end">
		<div class="flex flex-col gap-2 sm:flex-row sm:items-center sm:gap-3">
			<Button variant="secondary" onclick={handleTest} loading={testing} disabled={!config.host || !config.macaroon_hex}>
				Test Connection
			</Button>
			<Button variant="primary" onclick={handleSave} loading={saving} disabled={!config.host || !config.macaroon_hex}>
				Save Configuration
			</Button>
		</div>
	</div>
</div>
