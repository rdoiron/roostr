<script>
	import { lightning } from '$lib/api/client.js';
	import { notify } from '$lib/stores/app.svelte.js';
	import Button from '$lib/components/Button.svelte';

	let { status = null, onUpdate = () => {} } = $props();

	let config = $state({
		endpoint: '',
		macaroon: '',
		cert: ''
	});
	let saving = $state(false);
	let testing = $state(false);
	let detecting = $state(false);
	let testResult = $state(null);

	// Initialize config from status
	$effect(() => {
		if (status?.config) {
			config = {
				endpoint: status.config.endpoint || '',
				macaroon: status.config.macaroon || '',
				cert: status.config.cert || ''
			};
		}
	});

	async function handleDetect() {
		detecting = true;
		testResult = null;
		try {
			const result = await lightning.detect();
			if (result.detected) {
				config = {
					endpoint: result.endpoint || '',
					macaroon: result.macaroon || '',
					cert: result.cert || ''
				};
				notify('success', 'LND detected! Click "Test Connection" to verify.');
			} else {
				notify('warning', 'Could not auto-detect LND. Please enter credentials manually.');
			}
		} catch (e) {
			notify('error', e.message || 'Failed to detect LND');
		} finally {
			detecting = false;
		}
	}

	async function handleTest() {
		testing = true;
		testResult = null;
		try {
			const result = await lightning.test(config);
			testResult = {
				success: true,
				alias: result.alias,
				balance: result.balance
			};
			notify('success', `Connected to ${result.alias || 'LND node'}`);
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
			await lightning.updateConfig(config);
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
</script>

<div class="rounded-lg bg-white p-6 shadow">
	<div class="flex items-center justify-between mb-4">
		<div>
			<h2 class="text-lg font-semibold text-gray-900">Lightning Node</h2>
			<p class="text-sm text-gray-500">Connect your LND node to accept Lightning payments</p>
		</div>
		{#if status?.connected}
			<span class="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-green-100 text-green-800">
				<span class="w-2 h-2 mr-1.5 bg-green-500 rounded-full"></span>
				Connected
			</span>
		{:else}
			<span class="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-gray-100 text-gray-600">
				<span class="w-2 h-2 mr-1.5 bg-gray-400 rounded-full"></span>
				Not Connected
			</span>
		{/if}
	</div>

	{#if status?.connected && status?.node_info}
		<div class="mb-6 p-4 bg-green-50 border border-green-200 rounded-lg">
			<div class="flex items-center justify-between">
				<div>
					<p class="font-medium text-green-800">{status.node_info.alias || 'LND Node'}</p>
					<p class="text-sm text-green-600">Balance: {formatSats(status.node_info.balance || 0)}</p>
				</div>
				<svg class="w-8 h-8 text-green-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 10V3L4 14h7v7l9-11h-7z" />
				</svg>
			</div>
		</div>
	{/if}

	<div class="space-y-4">
		<div>
			<label for="endpoint" class="block text-sm font-medium text-gray-700 mb-1">
				LND REST Endpoint
			</label>
			<input
				type="text"
				id="endpoint"
				bind:value={config.endpoint}
				placeholder="umbrel.local:8080 or 127.0.0.1:8080"
				class="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-purple-500 focus:border-transparent"
			/>
			<p class="mt-1 text-xs text-gray-500">The REST API endpoint of your LND node</p>
		</div>

		<div>
			<label for="macaroon" class="block text-sm font-medium text-gray-700 mb-1">
				Admin Macaroon (hex)
			</label>
			<input
				type="password"
				id="macaroon"
				bind:value={config.macaroon}
				placeholder="0201036c6e6402..."
				class="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-purple-500 focus:border-transparent font-mono text-sm"
			/>
			<p class="mt-1 text-xs text-gray-500">Your admin.macaroon file encoded as hex</p>
		</div>

		<div>
			<label for="cert" class="block text-sm font-medium text-gray-700 mb-1">
				TLS Certificate Path (optional)
			</label>
			<input
				type="text"
				id="cert"
				bind:value={config.cert}
				placeholder="/path/to/tls.cert"
				class="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-purple-500 focus:border-transparent"
			/>
			<p class="mt-1 text-xs text-gray-500">Optional for local connections (Umbrel/Start9)</p>
		</div>
	</div>

	{#if testResult}
		<div class="mt-4 p-3 rounded-lg {testResult.success ? 'bg-green-50 border border-green-200' : 'bg-red-50 border border-red-200'}">
			{#if testResult.success}
				<div class="flex items-center text-green-700">
					<svg class="w-5 h-5 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
					</svg>
					<span class="font-medium">Connection successful!</span>
					{#if testResult.alias}
						<span class="ml-2 text-sm">({testResult.alias})</span>
					{/if}
				</div>
				{#if testResult.balance !== undefined}
					<p class="text-sm text-green-600 mt-1 ml-7">Balance: {formatSats(testResult.balance)}</p>
				{/if}
			{:else}
				<div class="flex items-center text-red-700">
					<svg class="w-5 h-5 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
					</svg>
					<span class="font-medium">Connection failed</span>
				</div>
				<p class="text-sm text-red-600 mt-1 ml-7">{testResult.error}</p>
			{/if}
		</div>
	{/if}

	<div class="flex items-center justify-between mt-6 pt-4 border-t">
		<Button variant="secondary" onclick={handleDetect} loading={detecting}>
			<svg class="w-4 h-4 mr-1" fill="none" stroke="currentColor" viewBox="0 0 24 24">
				<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
			</svg>
			Auto-detect
		</Button>
		<div class="flex items-center space-x-3">
			<Button variant="secondary" onclick={handleTest} loading={testing} disabled={!config.endpoint || !config.macaroon}>
				Test Connection
			</Button>
			<Button variant="primary" onclick={handleSave} loading={saving} disabled={!config.endpoint || !config.macaroon}>
				Save Configuration
			</Button>
		</div>
	</div>
</div>
