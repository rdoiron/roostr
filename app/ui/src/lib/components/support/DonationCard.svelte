<script>
	import { onMount } from 'svelte';
	import { notify } from '$lib/stores';

	let { title = '', address = '', icon = '', children } = $props();

	let qrDataUrl = $state('');
	let copied = $state(false);
	let copyTimer = null;

	onMount(async () => {
		if (address && typeof window !== 'undefined') {
			const QRCode = (await import('qrcode')).default;
			// For Lightning addresses, use lightning: URI; for Bitcoin, use bitcoin: URI
			const uri = title.toLowerCase().includes('lightning')
				? `lightning:${address}`
				: `bitcoin:${address}`;
			qrDataUrl = await QRCode.toDataURL(uri, {
				width: 180,
				margin: 2,
				color: { dark: '#1f2937', light: '#ffffff' }
			});
		}
		return () => {
			if (copyTimer) clearTimeout(copyTimer);
		};
	});

	async function copyAddress() {
		try {
			await navigator.clipboard.writeText(address);
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

<div class="rounded-lg bg-white dark:bg-gray-800 p-6 shadow dark:shadow-gray-900/50">
	<div class="flex items-center gap-2">
		<span class="text-2xl">{icon}</span>
		<h2 class="text-lg font-semibold text-gray-900 dark:text-gray-100">{title}</h2>
	</div>

	<div class="mt-4 flex flex-col items-center">
		{#if qrDataUrl}
			<div class="rounded-lg border border-gray-200 dark:border-gray-600 bg-white p-2">
				<img src={qrDataUrl} alt="QR Code for {title}" class="rounded" />
			</div>
		{:else}
			<div class="flex h-[180px] w-[180px] items-center justify-center rounded-lg bg-gray-100 dark:bg-gray-700">
				<span class="text-gray-400">Loading...</span>
			</div>
		{/if}

		<div class="mt-4 w-full">
			<div class="flex items-center gap-2 rounded-lg bg-gray-50 dark:bg-gray-700 p-3">
				<code class="flex-1 truncate font-mono text-sm text-gray-700 dark:text-gray-200">{address}</code>
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
		</div>

		{#if children}
			<div class="mt-4 w-full">
				{@render children()}
			</div>
		{/if}
	</div>
</div>
