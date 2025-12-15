<script>
	import { truncatePubkey } from '$lib/utils/format.js';

	let { authors = [] } = $props();
</script>

<div class="rounded-lg bg-gray-800 p-6">
	<h3 class="mb-4 text-lg font-semibold text-white">Top Authors</h3>
	<div class="space-y-3">
		{#if authors.length > 0}
			{#each authors as author, i}
				<div class="flex items-center justify-between rounded-lg bg-gray-700/50 px-4 py-3">
					<div class="flex items-center gap-3">
						<span
							class="flex h-6 w-6 items-center justify-center rounded-full text-xs font-bold
								{i === 0
									? 'bg-yellow-500 text-yellow-900'
									: i === 1
										? 'bg-gray-300 text-gray-700'
										: i === 2
											? 'bg-orange-400 text-orange-900'
											: 'bg-gray-600 text-gray-300'}"
						>
							{i + 1}
						</span>
						<code class="text-sm text-gray-300">{truncatePubkey(author.pubkey)}</code>
					</div>
					<span class="text-sm font-medium text-purple-400">
						{author.event_count.toLocaleString()} events
					</span>
				</div>
			{/each}
		{:else}
			<div class="flex h-32 items-center justify-center text-gray-500">No authors found</div>
		{/if}
	</div>
</div>
