<script>
	import { truncatePubkey } from '$lib/utils/format.js';

	let { authors = [] } = $props();
</script>

<div class="rounded-lg bg-white p-6 shadow">
	<h3 class="mb-4 text-lg font-semibold text-gray-900">Top Authors</h3>
	<div class="space-y-3">
		{#if authors?.length > 0}
			{#each authors as author, i}
				<div class="flex items-center justify-between rounded-lg bg-gray-50 px-4 py-3">
					<div class="flex items-center gap-3">
						<span
							class="flex h-6 w-6 items-center justify-center rounded-full text-xs font-bold
								{i === 0
									? 'bg-yellow-500 text-yellow-900'
									: i === 1
										? 'bg-gray-300 text-gray-700'
										: i === 2
											? 'bg-orange-400 text-orange-900'
											: 'bg-gray-200 text-gray-600'}"
						>
							{i + 1}
						</span>
						<code class="text-sm text-gray-600">{truncatePubkey(author.pubkey)}</code>
					</div>
					<span class="text-sm font-medium text-purple-600">
						{author.event_count.toLocaleString()} events
					</span>
				</div>
			{/each}
		{:else}
			<div class="flex h-32 items-center justify-center text-gray-500">No authors found</div>
		{/if}
	</div>
</div>
