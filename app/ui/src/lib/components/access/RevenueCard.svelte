<script>
	let { revenue = null } = $props();

	function formatSats(sats) {
		if (sats >= 100000000) {
			return (sats / 100000000).toFixed(4) + ' BTC';
		}
		if (sats >= 1000000) {
			return (sats / 1000000).toFixed(2) + 'M sats';
		}
		if (sats >= 1000) {
			return (sats / 1000).toFixed(1) + 'k sats';
		}
		return sats.toLocaleString() + ' sats';
	}

	function formatUSD(sats) {
		// Rough estimate at ~$100k/BTC
		const usd = (sats / 100000000) * 100000;
		if (usd < 0.01) return '';
		if (usd < 1) return `~$${usd.toFixed(2)}`;
		if (usd < 100) return `~$${usd.toFixed(0)}`;
		return `~$${usd.toLocaleString()}`;
	}
</script>

<div class="rounded-lg bg-white p-6 shadow">
	<div class="flex items-center justify-between mb-4">
		<div>
			<h2 class="text-lg font-semibold text-gray-900">Revenue Summary</h2>
			<p class="text-sm text-gray-500">Lifetime earnings from paid relay access</p>
		</div>
		<svg class="w-8 h-8 text-purple-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
			<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 10V3L4 14h7v7l9-11h-7z" />
		</svg>
	</div>

	{#if !revenue}
		<div class="flex items-center justify-center py-8">
			<div class="h-6 w-6 animate-spin rounded-full border-2 border-purple-600 border-t-transparent"></div>
		</div>
	{:else}
		<div class="grid grid-cols-2 lg:grid-cols-4 gap-4">
			<!-- Total Revenue -->
			<div class="p-4 bg-purple-50 rounded-lg">
				<p class="text-sm text-purple-600 font-medium">Total Revenue</p>
				<p class="text-2xl font-bold text-purple-900">{formatSats(revenue.total_sats || 0)}</p>
				{#if revenue.total_sats > 0}
					<p class="text-xs text-purple-500">{formatUSD(revenue.total_sats)}</p>
				{/if}
			</div>

			<!-- Active Subscribers -->
			<div class="p-4 bg-green-50 rounded-lg">
				<p class="text-sm text-green-600 font-medium">Active Users</p>
				<p class="text-2xl font-bold text-green-900">{revenue.active_count || 0}</p>
				<p class="text-xs text-green-500">with relay access</p>
			</div>

			<!-- Expiring Soon -->
			<div class="p-4 {revenue.expiring_soon > 0 ? 'bg-amber-50' : 'bg-gray-50'} rounded-lg">
				<p class="text-sm {revenue.expiring_soon > 0 ? 'text-amber-600' : 'text-gray-600'} font-medium">Expiring Soon</p>
				<p class="text-2xl font-bold {revenue.expiring_soon > 0 ? 'text-amber-900' : 'text-gray-900'}">{revenue.expiring_soon || 0}</p>
				<p class="text-xs {revenue.expiring_soon > 0 ? 'text-amber-500' : 'text-gray-500'}">within 7 days</p>
			</div>

			<!-- Total Payments -->
			<div class="p-4 bg-gray-50 rounded-lg">
				<p class="text-sm text-gray-600 font-medium">Total Payments</p>
				<p class="text-2xl font-bold text-gray-900">{revenue.payment_count || 0}</p>
				<p class="text-xs text-gray-500">transactions</p>
			</div>
		</div>

		<!-- Revenue by Tier -->
		{#if revenue.by_tier && Object.keys(revenue.by_tier).length > 0}
			<div class="mt-4 pt-4 border-t">
				<p class="text-sm font-medium text-gray-700 mb-2">Revenue by Tier</p>
				<div class="space-y-2">
					{#each Object.entries(revenue.by_tier) as [tier, amount]}
						<div class="flex items-center justify-between text-sm">
							<span class="text-gray-600">{tier}</span>
							<span class="font-medium text-gray-900">{formatSats(amount)}</span>
						</div>
					{/each}
				</div>
			</div>
		{/if}
	{/if}
</div>
