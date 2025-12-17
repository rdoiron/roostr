<script>
	import { events } from '$lib/api/client.js';
	import { notify } from '$lib/stores/app.svelte.js';
	import Button from '$lib/components/Button.svelte';

	let { event, onClose, onConfirm } = $props();

	let deleting = $state(false);
	let error = $state('');
	let reason = $state('');

	const kindNames = {
		0: 'Metadata',
		1: 'Note',
		3: 'Follow List',
		4: 'DM',
		5: 'Deletion',
		6: 'Repost',
		7: 'Reaction',
		14: 'DM',
		10002: 'Relay List'
	};

	function getKindName(kind) {
		return kindNames[kind] || `Kind ${kind}`;
	}

	function truncateId(id) {
		if (!id || id.length < 16) return id || '';
		return id.slice(0, 12) + '...' + id.slice(-8);
	}

	async function handleDelete() {
		deleting = true;
		error = '';

		try {
			await events.delete(event.id);
			notify('success', 'Event deletion queued');
			onConfirm?.();
			onClose?.();
		} catch (e) {
			if (e.code === 'DUPLICATE_REQUEST') {
				error = 'Deletion already requested for this event';
			} else {
				error = e.message || 'Failed to delete event';
			}
		} finally {
			deleting = false;
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
<div
	class="fixed inset-0 z-50 flex items-center justify-center bg-black/50 p-4"
	onclick={handleBackdropClick}
	role="dialog"
	aria-modal="true"
	aria-labelledby="modal-title"
>
	<div class="w-full max-w-md rounded-lg bg-white dark:bg-gray-800 shadow-xl">
		<!-- Header -->
		<div class="flex items-center justify-between border-b border-gray-200 dark:border-gray-700 px-6 py-4">
			<h2 id="modal-title" class="text-lg font-semibold text-gray-900 dark:text-gray-100">Delete Event</h2>
			<button
				type="button"
				onclick={onClose}
				aria-label="Close modal"
				class="rounded p-1 text-gray-400 transition-colors hover:bg-gray-100 dark:hover:bg-gray-700 hover:text-gray-600 dark:hover:text-gray-300"
			>
				<svg class="h-5 w-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path
						stroke-linecap="round"
						stroke-linejoin="round"
						stroke-width="2"
						d="M6 18L18 6M6 6l12 12"
					/>
				</svg>
			</button>
		</div>

		<!-- Body -->
		<div class="space-y-4 p-6">
			<!-- Warning icon -->
			<div class="flex justify-center">
				<div class="flex h-12 w-12 items-center justify-center rounded-full bg-red-100 dark:bg-red-900/30">
					<svg class="h-6 w-6 text-red-600 dark:text-red-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path
							stroke-linecap="round"
							stroke-linejoin="round"
							stroke-width="2"
							d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z"
						/>
					</svg>
				</div>
			</div>

			<!-- Message -->
			<div class="text-center">
				<p class="text-gray-900 dark:text-gray-100">Are you sure you want to delete this event?</p>
				<p class="mt-2 text-sm text-gray-500 dark:text-gray-400">This action will queue the event for deletion from your relay.</p>
			</div>

			<!-- Event info -->
			<div class="rounded-lg bg-gray-50 dark:bg-gray-700 p-3">
				<div class="grid grid-cols-2 gap-2 text-sm">
					<div>
						<p class="text-xs text-gray-500 dark:text-gray-400">Event ID:</p>
						<p class="truncate font-mono text-gray-700 dark:text-gray-200">{truncateId(event?.id)}</p>
					</div>
					<div>
						<p class="text-xs text-gray-500 dark:text-gray-400">Kind:</p>
						<p class="text-gray-700 dark:text-gray-200">{event?.kind} ({getKindName(event?.kind)})</p>
					</div>
				</div>
			</div>

			<!-- Reason input (optional) -->
			<div>
				<label for="delete-reason" class="mb-1 block text-sm font-medium text-gray-700 dark:text-gray-200">
					Reason (optional)
				</label>
				<input
					id="delete-reason"
					type="text"
					bind:value={reason}
					placeholder="e.g., Spam, Test event, etc."
					class="input w-full"
				/>
			</div>

			<!-- Error -->
			{#if error}
				<div class="rounded-lg border border-red-200 dark:border-red-800 bg-red-50 dark:bg-red-900/20 p-3">
					<p class="text-sm text-red-700 dark:text-red-300">{error}</p>
				</div>
			{/if}
		</div>

		<!-- Footer -->
		<div class="flex justify-end gap-3 border-t border-gray-200 dark:border-gray-700 px-6 py-4">
			<Button variant="secondary" onclick={onClose} disabled={deleting}>Cancel</Button>
			<Button variant="danger" onclick={handleDelete} disabled={deleting} loading={deleting}>
				Delete
			</Button>
		</div>
	</div>
</div>
