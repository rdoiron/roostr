<script>
	import { truncatePubkey, formatRelativeTime } from '$lib/utils/format.js';

	let { event, onViewRaw, onDelete } = $props();

	const kindNames = {
		0: 'METADATA',
		1: 'NOTE',
		3: 'FOLLOW LIST',
		4: 'DM',
		5: 'DELETION',
		6: 'REPOST',
		7: 'REACTION',
		14: 'DM',
		10002: 'RELAY LIST'
	};

	const kindColors = {
		0: 'bg-blue-100 text-blue-700',
		1: 'bg-purple-100 text-purple-700',
		3: 'bg-green-100 text-green-700',
		4: 'bg-yellow-100 text-yellow-700',
		5: 'bg-red-100 text-red-700',
		6: 'bg-indigo-100 text-indigo-700',
		7: 'bg-pink-100 text-pink-700',
		14: 'bg-yellow-100 text-yellow-700',
		10002: 'bg-cyan-100 text-cyan-700'
	};

	function getKindLabel(kind) {
		return kindNames[kind] || `KIND ${kind}`;
	}

	function getKindColor(kind) {
		return kindColors[kind] || 'bg-gray-100 text-gray-700';
	}

	// Kind-specific content rendering
	function renderContent(event) {
		switch (event.kind) {
			case 0: // Metadata
				return parseMetadata(event.content);
			case 1: // Note
				return truncateContent(event.content, 280);
			case 3: // Follow list
				return parseFollowList(event.tags);
			case 4: // DM (legacy)
			case 14: // DM (NIP-17)
				return { type: 'encrypted', text: 'Encrypted message' };
			case 7: // Reaction
				return parseReaction(event.content, event.tags);
			case 5: // Deletion
				return parseDeletion(event.tags);
			case 6: // Repost
				return { type: 'repost', text: 'Reposted an event' };
			default:
				return { type: 'raw', text: truncateContent(event.content || '(empty)', 100) };
		}
	}

	function parseMetadata(content) {
		try {
			const profile = JSON.parse(content);
			const name = profile.name || profile.display_name || 'Unknown';
			return { type: 'profile', text: `Profile update: ${name}` };
		} catch {
			return { type: 'profile', text: 'Profile update' };
		}
	}

	function parseFollowList(tags) {
		const following = tags.filter((t) => t[0] === 'p').length;
		return { type: 'follows', text: `Following ${following} accounts` };
	}

	function parseReaction(content, tags) {
		const emoji = content || '+';
		const eventRef = tags.find((t) => t[0] === 'e');
		const targetId = eventRef ? eventRef[1].slice(0, 12) + '...' : 'an event';
		return { type: 'reaction', emoji, text: `Reacted to ${targetId}` };
	}

	function parseDeletion(tags) {
		const count = tags.filter((t) => t[0] === 'e').length;
		return { type: 'deletion', text: `Requested deletion of ${count} event(s)` };
	}

	function truncateContent(text, maxLength) {
		if (!text) return { type: 'text', text: '' };
		const truncated = text.length > maxLength ? text.slice(0, maxLength) + '...' : text;
		return { type: 'text', text: truncated };
	}

	const content = $derived(renderContent(event));
	const formattedDate = $derived(formatRelativeTime(event.created_at));
	const authorDisplay = $derived(truncatePubkey(event.pubkey));
</script>

<div class="rounded-lg bg-white p-4 shadow transition-shadow hover:shadow-md">
	<div class="flex items-start justify-between">
		<div class="flex items-center gap-3">
			<span class="rounded px-2 py-1 text-xs font-semibold {getKindColor(event.kind)}">
				{getKindLabel(event.kind)}
			</span>
			<p class="font-mono text-sm text-gray-500">
				{authorDisplay}
			</p>
		</div>
		<p class="text-sm text-gray-400">{formattedDate}</p>
	</div>

	<div class="mt-3">
		{#if content.type === 'reaction'}
			<p class="text-gray-700">
				<span class="mr-2 text-xl">{content.emoji}</span>
				{content.text}
			</p>
		{:else if content.type === 'encrypted'}
			<p class="italic text-gray-400">
				<svg class="mr-1 inline h-4 w-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path
						stroke-linecap="round"
						stroke-linejoin="round"
						stroke-width="2"
						d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z"
					/>
				</svg>
				{content.text}
			</p>
		{:else}
			<p class="text-gray-700 line-clamp-3">{content.text}</p>
		{/if}
	</div>

	<div class="mt-3 flex gap-3 border-t pt-3">
		<button
			type="button"
			onclick={() => onViewRaw?.(event)}
			class="text-sm text-purple-600 hover:text-purple-700"
		>
			View Raw
		</button>
		<button
			type="button"
			onclick={() => onDelete?.(event)}
			class="text-sm text-red-600 hover:text-red-700"
		>
			Delete
		</button>
	</div>
</div>
