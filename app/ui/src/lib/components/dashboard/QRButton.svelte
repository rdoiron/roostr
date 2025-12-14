<script>
	import { onMount } from 'svelte';

	let { url = '' } = $props();
	let showQR = $state(false);
	let qrDataUrl = $state('');

	onMount(async () => {
		if (url && typeof window !== 'undefined') {
			const QRCode = (await import('qrcode')).default;
			qrDataUrl = await QRCode.toDataURL(url, {
				width: 200,
				margin: 2,
				color: { dark: '#1f2937', light: '#ffffff' }
			});
		}
	});
</script>

<div class="relative">
	<button
		type="button"
		onclick={() => (showQR = !showQR)}
		class="inline-flex items-center justify-center rounded-lg bg-gray-100 p-2 text-gray-700 transition-colors hover:bg-gray-200"
		title="QR Code"
	>
		<svg class="h-5 w-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
			<path
				stroke-linecap="round"
				stroke-linejoin="round"
				stroke-width="2"
				d="M12 4v1m6 11h2m-6 0h-2v4m0-11v3m0 0h.01M12 12h4.01M16 20h4M4 12h4m12 0h.01M5 8h2a1 1 0 001-1V5a1 1 0 00-1-1H5a1 1 0 00-1 1v2a1 1 0 001 1zm12 0h2a1 1 0 001-1V5a1 1 0 00-1-1h-2a1 1 0 00-1 1v2a1 1 0 001 1zM5 20h2a1 1 0 001-1v-2a1 1 0 00-1-1H5a1 1 0 00-1 1v2a1 1 0 001 1z"
			/>
		</svg>
	</button>

	{#if showQR && qrDataUrl}
		<div
			class="absolute right-0 z-10 mt-2 rounded-lg border border-gray-200 bg-white p-2 shadow-lg"
		>
			<img src={qrDataUrl} alt="QR Code for {url}" class="rounded" />
			<button
				type="button"
				onclick={() => (showQR = false)}
				class="absolute -right-2 -top-2 flex h-6 w-6 items-center justify-center rounded-full bg-gray-800 text-xs text-white"
			>
				&times;
			</button>
		</div>
	{/if}
</div>
