<script>
	import { onMount } from 'svelte';
	import { notify } from '$lib/stores';

	let { lightningAddress = '' } = $props();

	let webLNAvailable = $state(false);
	let loading = $state(false);

	onMount(async () => {
		// Check if WebLN is available (Alby, etc.)
		if (typeof window !== 'undefined' && window.webln) {
			try {
				await window.webln.enable();
				webLNAvailable = true;
			} catch {
				// User denied or no provider
				webLNAvailable = false;
			}
		}
	});

	async function tip() {
		if (!webLNAvailable || !lightningAddress) return;

		loading = true;
		try {
			// Parse lightning address (user@domain.com)
			const [name, domain] = lightningAddress.split('@');
			if (!name || !domain) {
				throw new Error('Invalid Lightning address');
			}

			// Resolve Lightning address to LNURL-pay endpoint
			const lnurlPayUrl = `https://${domain}/.well-known/lnurlp/${name}`;
			const response = await fetch(lnurlPayUrl);
			if (!response.ok) {
				throw new Error('Failed to resolve Lightning address');
			}

			const lnurlPayData = await response.json();
			if (lnurlPayData.status === 'ERROR') {
				throw new Error(lnurlPayData.reason || 'LNURL error');
			}

			// Use WebLN to make the payment
			// webln.lnurl handles the full LNURL-pay flow
			const result = await window.webln.lnurl(lnurlPayUrl);

			if (result) {
				notify('success', 'Thank you for your support!');
			}
		} catch (err) {
			// User may have cancelled - that's ok
			if (err.message && !err.message.includes('User rejected')) {
				notify('error', err.message || 'Failed to complete tip');
			}
		} finally {
			loading = false;
		}
	}
</script>

{#if webLNAvailable}
	<button
		type="button"
		onclick={tip}
		disabled={loading}
		class="flex w-full items-center justify-center gap-2 rounded-lg bg-amber-500 px-4 py-3 font-medium text-white transition-colors hover:bg-amber-600 disabled:opacity-50"
	>
		{#if loading}
			<svg class="h-5 w-5 animate-spin" fill="none" viewBox="0 0 24 24">
				<circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" />
				<path
					class="opacity-75"
					fill="currentColor"
					d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
				/>
			</svg>
			<span>Processing...</span>
		{:else}
			<svg class="h-5 w-5" fill="currentColor" viewBox="0 0 24 24">
				<path d="M13 10V3L4 14h7v7l9-11h-7z" />
			</svg>
			<span>Tip with WebLN</span>
		{/if}
	</button>
{/if}
