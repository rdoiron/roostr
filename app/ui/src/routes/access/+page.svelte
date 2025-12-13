<script>
	import { onMount } from 'svelte';
	import { get } from '$lib/api';

	let mode = $state('private');
	let whitelist = $state([]);
	let loading = $state(true);
	let error = $state(null);

	onMount(async () => {
		try {
			const [modeRes, listRes] = await Promise.all([
				get('/access/mode'),
				get('/access/whitelist')
			]);
			mode = modeRes.mode;
			whitelist = listRes.entries || [];
		} catch (e) {
			error = e.message;
		} finally {
			loading = false;
		}
	});
</script>

<div class="space-y-6">
	<div>
		<h1 class="text-2xl font-bold text-gray-900">Access Control</h1>
		<p class="text-gray-600">Manage who can use your relay</p>
	</div>

	{#if loading}
		<div class="flex items-center justify-center py-12">
			<div class="h-8 w-8 animate-spin rounded-full border-4 border-purple-600 border-t-transparent"></div>
		</div>
	{:else if error}
		<div class="rounded-lg bg-red-50 p-4 text-red-700">
			<p class="font-medium">Error loading access control</p>
			<p class="text-sm">{error}</p>
		</div>
	{:else}
		<!-- Access mode -->
		<div class="rounded-lg bg-white p-6 shadow">
			<h2 class="text-lg font-semibold text-gray-900">Access Mode</h2>
			<div class="mt-4 space-y-3">
				<label class="flex cursor-pointer items-start gap-3 rounded-lg border p-4 {mode === 'private' ? 'border-purple-500 bg-purple-50' : 'border-gray-200'}">
					<input type="radio" name="mode" value="private" bind:group={mode} class="mt-1" />
					<div>
						<p class="font-medium text-gray-900">Private</p>
						<p class="text-sm text-gray-500">Only whitelisted pubkeys can write</p>
					</div>
				</label>
				<label class="flex cursor-pointer items-start gap-3 rounded-lg border p-4 {mode === 'paid' ? 'border-purple-500 bg-purple-50' : 'border-gray-200'}">
					<input type="radio" name="mode" value="paid" bind:group={mode} class="mt-1" />
					<div>
						<p class="font-medium text-gray-900">Paid Access</p>
						<p class="text-sm text-gray-500">Whitelist + anyone who pays via Lightning</p>
					</div>
				</label>
				<label class="flex cursor-pointer items-start gap-3 rounded-lg border p-4 {mode === 'public' ? 'border-purple-500 bg-purple-50' : 'border-gray-200'}">
					<input type="radio" name="mode" value="public" bind:group={mode} class="mt-1" />
					<div>
						<p class="font-medium text-gray-900">Public</p>
						<p class="text-sm text-gray-500">Anyone can write (not recommended)</p>
					</div>
				</label>
			</div>
		</div>

		<!-- Whitelist -->
		<div class="rounded-lg bg-white p-6 shadow">
			<div class="flex items-center justify-between">
				<h2 class="text-lg font-semibold text-gray-900">Whitelist</h2>
				<button class="rounded-lg bg-purple-600 px-4 py-2 text-sm font-medium text-white hover:bg-purple-700">
					Add Pubkey
				</button>
			</div>
			<div class="mt-4">
				{#if whitelist.length === 0}
					<p class="text-gray-500">No pubkeys whitelisted yet</p>
				{:else}
					<div class="space-y-2">
						{#each whitelist as entry}
							<div class="flex items-center justify-between rounded-lg bg-gray-50 p-3">
								<div>
									<p class="font-medium text-gray-900">
										{entry.nickname || 'Unknown'}
										{#if entry.is_operator}
											<span class="ml-2 rounded bg-purple-100 px-2 py-0.5 text-xs text-purple-700">Operator</span>
										{/if}
									</p>
									<p class="text-sm text-gray-500 font-mono">{entry.npub?.slice(0, 20)}...</p>
								</div>
								<div class="text-sm text-gray-500">
									{entry.event_count ?? 0} events
								</div>
							</div>
						{/each}
					</div>
				{/if}
			</div>
		</div>
	{/if}
</div>
