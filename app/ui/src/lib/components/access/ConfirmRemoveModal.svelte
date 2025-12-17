<script>
	import { access } from '$lib/api/client.js';
	import { notify } from '$lib/stores/app.svelte.js';
	import Button from '$lib/components/Button.svelte';

	let { entry, listType = 'whitelist', onClose, onConfirm } = $props();

	let removing = $state(false);
	let error = $state('');

	function truncateNpub(npub) {
		if (!npub) return '';
		return npub.slice(0, 12) + '...' + npub.slice(-8);
	}

	const displayName = $derived(entry?.nickname || truncateNpub(entry?.npub) || 'this pubkey');

	async function handleRemove() {
		removing = true;
		error = '';

		try {
			if (listType === 'whitelist') {
				await access.removeFromWhitelist(entry.pubkey);
				notify('success', `Removed ${displayName} from whitelist`);
			} else {
				await access.removeFromBlacklist(entry.pubkey);
				notify('success', `Removed ${displayName} from blacklist`);
			}
			onConfirm?.();
			onClose?.();
		} catch (e) {
			if (e.code === 'CANNOT_REMOVE_OPERATOR') {
				error = 'Cannot remove the operator from the whitelist';
			} else {
				error = e.message || 'Failed to remove from list';
			}
		} finally {
			removing = false;
		}
	}

	function handleBackdropClick(e) {
		if (e.target === e.currentTarget) {
			onClose?.();
		}
	}

	function handleKeydown(e) {
		if (e.key === 'Escape') {
			onClose?.();
		}
	}
</script>

<svelte:window onkeydown={handleKeydown} />

<!-- svelte-ignore a11y_click_events_have_key_events a11y_interactive_supports_focus -->
<!-- Modal backdrop -->
<div
	class="fixed inset-0 z-50 flex items-center justify-center bg-black/50 p-4"
	onclick={handleBackdropClick}
	role="dialog"
	aria-modal="true"
	aria-labelledby="modal-title"
>
	<!-- Modal content -->
	<div class="w-full max-w-md rounded-lg bg-white dark:bg-gray-800 shadow-xl">
		<!-- Header -->
		<div class="flex items-center justify-between border-b border-gray-200 dark:border-gray-700 px-6 py-4">
			<h2 id="modal-title" class="text-lg font-semibold text-gray-900 dark:text-gray-100">
				Remove from {listType === 'whitelist' ? 'Whitelist' : 'Blacklist'}
			</h2>
			<button
				type="button"
				onclick={onClose}
				aria-label="Close modal"
				class="p-1 text-gray-400 hover:text-gray-600 dark:hover:text-gray-300 transition-colors rounded hover:bg-gray-100 dark:hover:bg-gray-700"
			>
				<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
				</svg>
			</button>
		</div>

		<!-- Body -->
		<div class="p-6 space-y-4">
			<!-- Warning icon -->
			<div class="flex justify-center">
				<div class="w-12 h-12 rounded-full bg-red-100 dark:bg-red-900/30 flex items-center justify-center">
					<svg class="w-6 h-6 text-red-600 dark:text-red-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
					</svg>
				</div>
			</div>

			<!-- Message -->
			<div class="text-center">
				<p class="text-gray-900 dark:text-gray-100">
					Are you sure you want to remove <strong class="font-medium">{displayName}</strong> from the {listType}?
				</p>
				<p class="text-sm text-gray-500 dark:text-gray-400 mt-2">
					{#if listType === 'whitelist'}
						They will no longer be able to write events to your relay.
					{:else}
						They will be able to write events to your relay again.
					{/if}
				</p>
			</div>

			<!-- Pubkey info -->
			<div class="p-3 bg-gray-50 dark:bg-gray-700 rounded-lg">
				<p class="text-xs text-gray-500 dark:text-gray-400">Pubkey:</p>
				<p class="text-xs font-mono text-gray-700 dark:text-gray-200 truncate">{entry?.npub || entry?.pubkey}</p>
			</div>

			<!-- Error -->
			{#if error}
				<div class="p-3 bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-lg">
					<p class="text-sm text-red-700 dark:text-red-300">{error}</p>
				</div>
			{/if}
		</div>

		<!-- Footer -->
		<div class="flex justify-end space-x-3 border-t border-gray-200 dark:border-gray-700 px-6 py-4">
			<Button variant="secondary" onclick={onClose} disabled={removing}>
				Cancel
			</Button>
			<Button variant="danger" onclick={handleRemove} disabled={removing} loading={removing}>
				Remove
			</Button>
		</div>
	</div>
</div>
