<script>
	import { formatRelativeTime, getKindLabel, truncatePubkey } from '$lib/utils/format.js';

	let { events = [] } = $props();

	function getEventIcon(kind) {
		switch (kind) {
			case 1:
				return 'M7 8h10M7 12h4m1 8l-4-4H5a2 2 0 01-2-2V6a2 2 0 012-2h14a2 2 0 012 2v8a2 2 0 01-2 2h-3l-4 4z';
			case 7:
				return 'M4.318 6.318a4.5 4.5 0 000 6.364L12 20.364l7.682-7.682a4.5 4.5 0 00-6.364-6.364L12 7.636l-1.318-1.318a4.5 4.5 0 00-6.364 0z';
			case 4:
			case 14:
				return 'M3 8l7.89 5.26a2 2 0 002.22 0L21 8M5 19h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z';
			case 6:
				return 'M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15';
			case 3:
				return 'M12 4.354a4 4 0 110 5.292M15 21H3v-1a6 6 0 0112 0v1zm0 0h6v-1a6 6 0 00-9-5.197M13 7a4 4 0 11-8 0 4 4 0 018 0z';
			default:
				return 'M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z';
		}
	}
</script>

<div class="rounded-lg bg-white shadow">
	<div class="flex items-center justify-between border-b border-gray-100 p-4">
		<h2 class="text-lg font-semibold text-gray-900">Recent Activity</h2>
		<a href="/events" class="text-sm text-purple-600 hover:text-purple-700">View All</a>
	</div>

	{#if events.length === 0}
		<div class="p-8 text-center text-gray-500">
			<p>No recent events</p>
		</div>
	{:else}
		<ul class="divide-y divide-gray-100">
			{#each events as event (event.id)}
				<li>
					<a
						href="/events?id={event.id}"
						class="flex items-center gap-3 p-4 transition-colors hover:bg-gray-50"
					>
						<div
							class="flex h-8 w-8 flex-shrink-0 items-center justify-center rounded-full bg-gray-100"
						>
							<svg
								class="h-4 w-4 text-gray-600"
								fill="none"
								stroke="currentColor"
								viewBox="0 0 24 24"
							>
								<path
									stroke-linecap="round"
									stroke-linejoin="round"
									stroke-width="2"
									d={getEventIcon(event.kind)}
								/>
							</svg>
						</div>
						<div class="min-w-0 flex-1">
							<p class="text-sm font-medium text-gray-900">{getKindLabel(event.kind)} received</p>
							<p class="truncate text-xs text-gray-500">{truncatePubkey(event.pubkey)}</p>
						</div>
						<div class="flex flex-shrink-0 items-center gap-2">
							<span class="text-xs text-gray-400">{formatRelativeTime(event.created_at)}</span>
							<svg
								class="h-4 w-4 text-gray-400"
								fill="none"
								stroke="currentColor"
								viewBox="0 0 24 24"
							>
								<path
									stroke-linecap="round"
									stroke-linejoin="round"
									stroke-width="2"
									d="M9 5l7 7-7 7"
								/>
							</svg>
						</div>
					</a>
				</li>
			{/each}
		</ul>
	{/if}
</div>
