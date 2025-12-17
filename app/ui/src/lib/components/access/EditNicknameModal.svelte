<script>
	import { access } from '$lib/api/client.js';
	import { notify } from '$lib/stores/app.svelte.js';
	import Button from '$lib/components/Button.svelte';

	let { entry, onClose, onSave } = $props();

	let nickname = $state('');
	let saving = $state(false);
	let error = $state('');

	// Initialize nickname from entry prop
	$effect(() => {
		nickname = entry?.nickname || '';
	});

	function truncateNpub(npub) {
		if (!npub) return '';
		return npub.slice(0, 12) + '...' + npub.slice(-8);
	}

	async function handleSave() {
		saving = true;
		error = '';

		try {
			await access.updateWhitelist(entry.pubkey, { nickname: nickname.trim() });
			notify('success', 'Nickname updated');
			onSave?.();
			onClose?.();
		} catch (e) {
			error = e.message || 'Failed to update nickname';
		} finally {
			saving = false;
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
		} else if (e.key === 'Enter' && !saving) {
			handleSave();
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
			<h2 id="modal-title" class="text-lg font-semibold text-gray-900 dark:text-gray-100">Edit Nickname</h2>
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
			<!-- Pubkey info -->
			<div class="p-3 bg-gray-50 dark:bg-gray-700 rounded-lg">
				<p class="text-sm text-gray-500 dark:text-gray-400">Editing nickname for:</p>
				<p class="text-sm font-mono text-gray-700 dark:text-gray-200 truncate">{entry?.npub || truncateNpub(entry?.pubkey)}</p>
			</div>

			<!-- Nickname input -->
			<div>
				<label for="nickname-input" class="block text-sm font-medium text-gray-700 dark:text-gray-200 mb-2">
					Nickname
				</label>
				<input
					type="text"
					id="nickname-input"
					bind:value={nickname}
					placeholder="e.g., Family, Friend, Work colleague"
					class="input w-full"
					disabled={saving}
				/>
				<p class="text-xs text-gray-400 mt-1">Leave empty to remove the nickname</p>
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
			<Button variant="secondary" onclick={onClose} disabled={saving}>
				Cancel
			</Button>
			<Button variant="primary" onclick={handleSave} disabled={saving} loading={saving}>
				Save
			</Button>
		</div>
	</div>
</div>
