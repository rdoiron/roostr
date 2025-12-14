<script>
	let { entry, listType = 'whitelist', onEdit = null, onRemove = null } = $props();

	function truncateNpub(npub) {
		if (!npub) return '';
		return npub.slice(0, 12) + '...' + npub.slice(-8);
	}

	function formatDate(timestamp) {
		if (!timestamp) return 'Unknown';
		const date = new Date(timestamp * 1000);
		return date.toLocaleDateString('en-US', { month: 'short', day: 'numeric', year: 'numeric' });
	}

	const isOperator = $derived(entry.is_operator);
	const displayName = $derived(entry.nickname || (entry.npub ? truncateNpub(entry.npub) : 'Unknown'));
</script>

<div class="flex items-center justify-between rounded-lg border p-4 {isOperator ? 'border-purple-200 bg-purple-50' : 'border-gray-200 bg-gray-50'}">
	<div class="flex items-center space-x-3 min-w-0 flex-1">
		<div class="w-10 h-10 rounded-full flex items-center justify-center flex-shrink-0 {isOperator ? 'bg-purple-200' : listType === 'blacklist' ? 'bg-red-100' : 'bg-green-100'}">
			{#if listType === 'blacklist'}
				<svg class="w-5 h-5 text-red-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M18.364 18.364A9 9 0 005.636 5.636m12.728 12.728A9 9 0 015.636 5.636m12.728 12.728L5.636 5.636" />
				</svg>
			{:else}
				<svg class="w-5 h-5 {isOperator ? 'text-purple-600' : 'text-green-600'}" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z" />
				</svg>
			{/if}
		</div>
		<div class="min-w-0 flex-1">
			<div class="flex items-center space-x-2">
				<p class="font-medium text-gray-900 truncate">{displayName}</p>
				{#if isOperator}
					<span class="flex-shrink-0 rounded bg-purple-100 px-2 py-0.5 text-xs font-medium text-purple-700">Operator</span>
				{/if}
			</div>
			<p class="text-sm text-gray-500 font-mono truncate">{entry.npub || entry.pubkey}</p>
			<div class="flex items-center space-x-3 mt-1 text-xs text-gray-400">
				<span>Added {formatDate(entry.added_at)}</span>
				{#if listType === 'whitelist' && entry.event_count !== undefined}
					<span class="flex items-center space-x-1">
						<svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
						</svg>
						<span>{entry.event_count.toLocaleString()} events</span>
					</span>
				{/if}
				{#if listType === 'blacklist' && entry.reason}
					<span class="text-red-500 truncate">Reason: {entry.reason}</span>
				{/if}
			</div>
		</div>
	</div>
	<div class="flex items-center space-x-2 flex-shrink-0 ml-4">
		{#if onEdit && listType === 'whitelist'}
			<button
				type="button"
				onclick={() => onEdit(entry)}
				class="p-2 text-gray-400 hover:text-gray-600 transition-colors rounded hover:bg-gray-100"
				title="Edit nickname"
			>
				<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15.232 5.232l3.536 3.536m-2.036-5.036a2.5 2.5 0 113.536 3.536L6.5 21.036H3v-3.572L16.732 3.732z" />
				</svg>
			</button>
		{/if}
		{#if onRemove}
			<button
				type="button"
				onclick={() => onRemove(entry)}
				disabled={isOperator}
				class="p-2 transition-colors rounded {isOperator ? 'text-gray-300 cursor-not-allowed' : 'text-gray-400 hover:text-red-500 hover:bg-red-50'}"
				title={isOperator ? 'Cannot remove operator' : 'Remove'}
			>
				<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
				</svg>
			</button>
		{/if}
	</div>
</div>
