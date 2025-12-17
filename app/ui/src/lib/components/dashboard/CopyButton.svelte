<script>
	import { onMount } from 'svelte';
	import { notify } from '$lib/stores';

	let { text = '' } = $props();
	let copied = $state(false);
	let resetTimer = null;

	onMount(() => {
		return () => {
			if (resetTimer) clearTimeout(resetTimer);
		};
	});

	// Fallback copy method for non-secure contexts (HTTP)
	function fallbackCopy(str) {
		const textarea = document.createElement('textarea');
		textarea.value = str;
		textarea.style.position = 'fixed';
		textarea.style.left = '-9999px';
		textarea.style.top = '-9999px';
		document.body.appendChild(textarea);
		textarea.focus();
		textarea.select();
		try {
			document.execCommand('copy');
			return true;
		} catch {
			return false;
		} finally {
			document.body.removeChild(textarea);
		}
	}

	async function copy() {
		try {
			// Try modern Clipboard API first (requires HTTPS or localhost)
			if (navigator.clipboard && window.isSecureContext) {
				await navigator.clipboard.writeText(text);
			} else {
				// Fallback for HTTP contexts
				const success = fallbackCopy(text);
				if (!success) throw new Error('Fallback copy failed');
			}
			copied = true;
			notify('success', 'Copied to clipboard');
			if (resetTimer) clearTimeout(resetTimer);
			resetTimer = setTimeout(() => {
				copied = false;
			}, 2000);
		} catch {
			notify('error', 'Failed to copy');
		}
	}
</script>

<button
	type="button"
	onclick={copy}
	class="inline-flex items-center justify-center rounded-lg bg-gray-100 dark:bg-gray-600 p-2 text-gray-700 dark:text-gray-300 transition-colors hover:bg-gray-200 dark:hover:bg-gray-500"
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
