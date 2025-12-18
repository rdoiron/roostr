<script>
	import { onMount } from 'svelte';
	import { notify } from '$lib/stores';
	import WebLNTipButton from './WebLNTipButton.svelte';

	let { lightningAddress = '', bitcoinAddress = '' } = $props();

	let paymentMethod = $state('lightning');
	let qrDataUrl = $state('');
	let copied = $state(false);
	let copyTimer = null;

	let currentAddress = $derived(paymentMethod === 'lightning' ? lightningAddress : bitcoinAddress);

	async function generateQR(address, method) {
		if (!address || typeof window === 'undefined') return '';
		const QRCode = (await import('qrcode')).default;
		const uri = method === 'lightning' ? address : `bitcoin:${address}`;
		return await QRCode.toDataURL(uri, {
			width: 180,
			margin: 2,
			color: { dark: '#1f2937', light: '#ffffff' }
		});
	}

	$effect(() => {
		generateQR(currentAddress, paymentMethod).then((url) => {
			qrDataUrl = url;
		});
	});

	onMount(() => {
		return () => {
			if (copyTimer) clearTimeout(copyTimer);
		};
	});

	async function copyAddress() {
		try {
			await navigator.clipboard.writeText(currentAddress);
			copied = true;
			notify('success', 'Copied to clipboard');
			if (copyTimer) clearTimeout(copyTimer);
			copyTimer = setTimeout(() => {
				copied = false;
			}, 2000);
		} catch {
			notify('error', 'Failed to copy');
		}
	}
</script>

<div class="space-y-6">
	<!-- Payment Method Toggle -->
	<div class="flex justify-center">
		<div class="inline-flex rounded-lg bg-gray-100 dark:bg-gray-700 p-1">
			<button
				type="button"
				onclick={() => paymentMethod = 'bitcoin'}
				class="flex items-center gap-2 rounded-md px-4 py-2 text-sm font-medium transition-colors {paymentMethod === 'bitcoin'
					? 'bg-white dark:bg-gray-600 text-gray-900 dark:text-gray-100 shadow'
					: 'text-gray-600 dark:text-gray-400 hover:text-gray-900 dark:hover:text-gray-200'}"
			>
				<img src="/bitcoin-logo.svg" alt="Bitcoin" class="h-5 w-5" />
				Bitcoin
			</button>
			<button
				type="button"
				onclick={() => paymentMethod = 'lightning'}
				class="flex items-center gap-2 rounded-md px-4 py-2 text-sm font-medium transition-colors {paymentMethod === 'lightning'
					? 'bg-white dark:bg-gray-600 text-gray-900 dark:text-gray-100 shadow'
					: 'text-gray-600 dark:text-gray-400 hover:text-gray-900 dark:hover:text-gray-200'}"
			>
				<img src="/lightning-logo.svg" alt="Lightning" class="h-5 w-5" />
				Lightning
			</button>
		</div>
	</div>

	<!-- QR Code -->
	<div class="flex flex-col items-center">
		{#if qrDataUrl}
			<div class="rounded-lg border border-gray-200 dark:border-gray-600 bg-white p-2">
				<img src={qrDataUrl} alt="QR Code" class="rounded" />
			</div>
		{:else}
			<div class="flex h-[180px] w-[180px] items-center justify-center rounded-lg bg-gray-100 dark:bg-gray-700">
				<span class="text-gray-400">Loading...</span>
			</div>
		{/if}
	</div>

	<!-- Address -->
	<div class="flex items-center gap-2 rounded-lg bg-gray-50 dark:bg-gray-700 p-3">
		<code class="flex-1 truncate font-mono text-sm text-gray-700 dark:text-gray-200">{currentAddress}</code>
		<button
			type="button"
			onclick={copyAddress}
			class="inline-flex shrink-0 items-center justify-center rounded-lg bg-gray-100 dark:bg-gray-600 p-2 text-gray-700 dark:text-gray-200 transition-colors hover:bg-gray-200 dark:hover:bg-gray-500"
			title="Copy"
		>
			{#if copied}
				<svg class="h-5 w-5 text-green-600 dark:text-green-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
				</svg>
			{:else}
				<svg class="h-5 w-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path
						stroke-linecap="round"
						stroke-linejoin="round"
						stroke-width="2"
						d="M8 5H6a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2v-1M8 5a2 2 0 002 2h2a2 2 0 002-2M8 5a2 2 0 012-2h2a2 2 0 012 2m0 0h2a2 2 0 012 2v3m2 4H10m0 0l3-3m-3 3l3 3"
					/>
				</svg>
			{/if}
		</button>
	</div>

	<!-- WebLN Button (Lightning only) -->
	{#if paymentMethod === 'lightning'}
		<WebLNTipButton lightningAddress={lightningAddress} />
	{/if}
</div>
