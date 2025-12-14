<script>
	import { onMount } from 'svelte';
	import { notify } from '$lib/stores/app.svelte.js';
	import Button from '$lib/components/Button.svelte';

	let { event, onClose, onDelete } = $props();

	let copyTimer = null;
	let copiedField = $state('');

	onMount(() => {
		return () => {
			if (copyTimer) clearTimeout(copyTimer);
		};
	});

	const kindNames = {
		0: 'Metadata',
		1: 'Short Text Note',
		3: 'Follow List',
		4: 'Encrypted DM (legacy)',
		5: 'Deletion',
		6: 'Repost',
		7: 'Reaction',
		14: 'Private Message',
		10002: 'Relay List'
	};

	function getKindName(kind) {
		return kindNames[kind] || `Unknown`;
	}

	function formatDate(timestamp) {
		const date = new Date(timestamp);
		return date.toLocaleString();
	}

	function getUnixTimestamp(timestamp) {
		return Math.floor(new Date(timestamp).getTime() / 1000);
	}

	async function copyToClipboard(text, field) {
		try {
			await navigator.clipboard.writeText(text);
			copiedField = field;
			notify('success', 'Copied to clipboard');
			if (copyTimer) clearTimeout(copyTimer);
			copyTimer = setTimeout(() => {
				copiedField = '';
			}, 2000);
		} catch {
			notify('error', 'Failed to copy');
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

	// Build raw JSON for display
	const rawJSON = $derived(
		JSON.stringify(
			{
				id: event.id,
				pubkey: event.pubkey,
				created_at: getUnixTimestamp(event.created_at),
				kind: event.kind,
				tags: event.tags,
				content: event.content,
				sig: event.sig
			},
			null,
			2
		)
	);

	const isEncrypted = $derived(event.kind === 4 || event.kind === 14);
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
	<div class="max-h-[90vh] w-full max-w-2xl overflow-y-auto rounded-lg bg-white shadow-xl">
		<!-- Header -->
		<div class="sticky top-0 flex items-center justify-between border-b bg-white px-6 py-4">
			<h2 id="modal-title" class="text-lg font-semibold text-gray-900">Event Details</h2>
			<button
				type="button"
				onclick={onClose}
				aria-label="Close modal"
				class="rounded p-1 text-gray-400 transition-colors hover:bg-gray-100 hover:text-gray-600"
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
			<!-- Event ID -->
			<div>
				<p class="mb-1 text-xs font-medium text-gray-500">Event ID</p>
				<div class="flex items-center gap-2">
					<code class="flex-1 truncate rounded bg-gray-100 px-3 py-2 font-mono text-sm">
						{event.id}
					</code>
					<button
						type="button"
						onclick={() => copyToClipboard(event.id, 'id')}
						class="rounded bg-gray-100 p-2 text-gray-600 transition-colors hover:bg-gray-200"
						title="Copy Event ID"
					>
						{#if copiedField === 'id'}
							<svg class="h-5 w-5 text-green-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
							</svg>
						{:else}
							<svg class="h-5 w-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 5H6a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2v-1M8 5a2 2 0 002 2h2a2 2 0 002-2M8 5a2 2 0 012-2h2a2 2 0 012 2m0 0h2a2 2 0 012 2v3m2 4H10m0 0l3-3m-3 3l3 3" />
							</svg>
						{/if}
					</button>
				</div>
			</div>

			<!-- Author -->
			<div>
				<p class="mb-1 text-xs font-medium text-gray-500">Author</p>
				<div class="flex items-center gap-2">
					<code class="flex-1 truncate rounded bg-gray-100 px-3 py-2 font-mono text-sm">
						{event.pubkey}
					</code>
					<button
						type="button"
						onclick={() => copyToClipboard(event.pubkey, 'pubkey')}
						class="rounded bg-gray-100 p-2 text-gray-600 transition-colors hover:bg-gray-200"
						title="Copy Pubkey"
					>
						{#if copiedField === 'pubkey'}
							<svg class="h-5 w-5 text-green-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
							</svg>
						{:else}
							<svg class="h-5 w-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 5H6a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2v-1M8 5a2 2 0 002 2h2a2 2 0 002-2M8 5a2 2 0 012-2h2a2 2 0 012 2m0 0h2a2 2 0 012 2v3m2 4H10m0 0l3-3m-3 3l3 3" />
							</svg>
						{/if}
					</button>
				</div>
			</div>

			<!-- Kind and Created -->
			<div class="grid grid-cols-2 gap-4">
				<div>
					<p class="mb-1 text-xs font-medium text-gray-500">Kind</p>
					<p class="rounded bg-gray-100 px-3 py-2 text-sm">
						{event.kind} ({getKindName(event.kind)})
					</p>
				</div>
				<div>
					<p class="mb-1 text-xs font-medium text-gray-500">Created</p>
					<p class="rounded bg-gray-100 px-3 py-2 text-sm">
						{formatDate(event.created_at)}
						<span class="text-gray-400">({getUnixTimestamp(event.created_at)})</span>
					</p>
				</div>
			</div>

			<!-- Content -->
			<div>
				<p class="mb-1 text-xs font-medium text-gray-500">Content</p>
				<div class="max-h-40 overflow-y-auto rounded bg-gray-100 px-3 py-2">
					{#if isEncrypted}
						<p class="italic text-gray-400">
							<svg class="mr-1 inline h-4 w-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z" />
							</svg>
							Encrypted content
						</p>
					{:else if event.content}
						<p class="whitespace-pre-wrap break-words text-sm">{event.content}</p>
					{:else}
						<p class="italic text-gray-400">(empty)</p>
					{/if}
				</div>
			</div>

			<!-- Tags -->
			{#if event.tags && event.tags.length > 0}
				<div>
					<p class="mb-1 text-xs font-medium text-gray-500">Tags ({event.tags.length})</p>
					<div class="max-h-32 overflow-y-auto rounded bg-gray-100 px-3 py-2">
						<pre class="font-mono text-xs text-gray-700">{JSON.stringify(event.tags, null, 2)}</pre>
					</div>
				</div>
			{/if}

			<!-- Raw JSON -->
			<div>
				<div class="mb-1 flex items-center justify-between">
					<p class="text-xs font-medium text-gray-500">Raw JSON</p>
					<button
						type="button"
						onclick={() => copyToClipboard(rawJSON, 'json')}
						class="rounded px-2 py-1 text-xs text-purple-600 transition-colors hover:bg-purple-50"
					>
						{copiedField === 'json' ? 'Copied!' : 'Copy JSON'}
					</button>
				</div>
				<div class="max-h-48 overflow-y-auto rounded bg-gray-900 p-3">
					<pre class="font-mono text-xs text-green-400">{rawJSON}</pre>
				</div>
			</div>
		</div>

		<!-- Footer -->
		<div class="sticky bottom-0 flex justify-end gap-3 border-t bg-white px-6 py-4">
			<Button variant="danger" onclick={() => onDelete?.(event)}>
				Delete Event
			</Button>
			<Button variant="secondary" onclick={onClose}>
				Close
			</Button>
		</div>
	</div>
</div>
