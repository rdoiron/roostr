<script>
	import { onMount } from 'svelte';

	let { url = '' } = $props();
	let showQR = $state(false);
	let qrDataUrl = $state('');

	onMount(async () => {
		if (url && typeof window !== 'undefined') {
			const QRCode = (await import('qrcode')).default;
			qrDataUrl = await QRCode.toDataURL(url, {
				width: 256,
				margin: 2,
				color: { dark: '#1f2937', light: '#ffffff' }
			});
		}
	});

	function closeModal() {
		showQR = false;
	}

	function handleBackdropClick(e) {
		if (e.target === e.currentTarget) {
			closeModal();
		}
	}

	function handleKeydown(e) {
		if (e.key === 'Escape') {
			closeModal();
		}
	}
</script>

<svelte:window onkeydown={showQR ? handleKeydown : undefined} />

<div>
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
		<!-- Modal backdrop -->
		<div
			class="fixed inset-0 z-50 flex items-start justify-center bg-black/50 pt-20"
			onclick={handleBackdropClick}
			onkeydown={handleKeydown}
			role="dialog"
			aria-modal="true"
			tabindex="-1"
		>
			<!-- Modal content -->
			<div class="relative rounded-xl bg-white p-4 shadow-2xl">
				<img
					src={qrDataUrl}
					alt="QR Code for {url}"
					class="h-64 w-64 rounded-lg"
				/>
				<p class="mt-2 max-w-64 truncate text-center text-xs text-gray-500">{url}</p>
				<button
					type="button"
					onclick={closeModal}
					class="absolute -right-3 -top-3 flex h-8 w-8 items-center justify-center rounded-full bg-gray-800 text-white shadow-lg hover:bg-gray-700"
					aria-label="Close QR code"
				>
					<svg class="h-4 w-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
					</svg>
				</button>
			</div>
		</div>
	{/if}
</div>
