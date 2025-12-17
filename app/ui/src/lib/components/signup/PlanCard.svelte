<script>
	let { tier, onSelect = () => {} } = $props();

	function formatDuration(days) {
		if (!days) return 'Forever';
		if (days === 30) return '/month';
		if (days === 90) return '/quarter';
		if (days === 365) return '/year';
		return `/${days} days`;
	}

	function getDurationLabel(days) {
		if (!days) return 'Pay once, access forever';
		if (days === 30) return 'Auto-expires after 30 days';
		if (days === 90) return 'Auto-expires after 90 days';
		if (days === 365) return 'Auto-expires after 1 year';
		return `Auto-expires after ${days} days`;
	}

	function formatUSD(sats) {
		// Rough estimate at ~$100k/BTC
		const usd = (sats / 100000000) * 100000;
		if (usd < 0.01) return null;
		return `~$${usd.toFixed(2)}`;
	}

	const isLifetime = $derived(!tier.duration_days);
</script>

<div class="relative p-6 border-2 rounded-xl transition-all hover:border-purple-300 dark:hover:border-purple-600 hover:shadow-lg {isLifetime ? 'border-purple-200 dark:border-purple-700 bg-purple-50 dark:bg-purple-900/20' : 'border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-800'}">
	{#if isLifetime}
		<div class="absolute -top-3 left-1/2 transform -translate-x-1/2">
			<span class="px-3 py-1 text-xs font-semibold text-purple-700 dark:text-purple-300 bg-purple-100 dark:bg-purple-900/50 rounded-full border border-purple-200 dark:border-purple-700">
				Best Value
			</span>
		</div>
	{/if}

	<div class="text-center">
		<div class="w-12 h-12 mx-auto mb-4 rounded-full flex items-center justify-center {isLifetime ? 'bg-purple-200 dark:bg-purple-900/50' : 'bg-gray-100 dark:bg-gray-700'}">
			{#if isLifetime}
				<svg class="w-6 h-6 text-purple-600 dark:text-purple-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 10V3L4 14h7v7l9-11h-7z" />
				</svg>
			{:else}
				<svg class="w-6 h-6 text-gray-600 dark:text-gray-300" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z" />
				</svg>
			{/if}
		</div>

		<h3 class="text-lg font-semibold text-gray-900 dark:text-gray-100 mb-1">{tier.name}</h3>

		<div class="mb-2">
			<span class="text-3xl font-bold text-gray-900 dark:text-gray-100">{tier.amount_sats.toLocaleString()}</span>
			<span class="text-gray-500 dark:text-gray-400"> sats{formatDuration(tier.duration_days)}</span>
		</div>

		{#if formatUSD(tier.amount_sats)}
			<p class="text-sm text-gray-400 mb-4">{formatUSD(tier.amount_sats)}</p>
		{:else}
			<div class="mb-4"></div>
		{/if}

		<p class="text-sm text-gray-600 dark:text-gray-400 mb-6">{getDurationLabel(tier.duration_days)}</p>

		<button
			type="button"
			onclick={() => onSelect(tier)}
			class="w-full py-2.5 px-4 rounded-lg font-medium transition-colors {isLifetime ? 'bg-purple-600 hover:bg-purple-700 text-white' : 'bg-gray-900 dark:bg-gray-700 hover:bg-gray-800 dark:hover:bg-gray-600 text-white'}"
		>
			Select
		</button>
	</div>
</div>
