<script>
	import { formatRelativeTime } from '$lib/utils/format.js';
	import { formatDateInTimezone } from '$lib/stores/timezone.svelte.js';

	let { user, onRevoke = () => {} } = $props();

	const displayName = $derived(user.nickname || truncateNpub(user.npub) || 'Unknown');
	const isExpiringSoon = $derived(user.expires_at && daysUntilExpiry(user.expires_at) <= 7);
	const isExpired = $derived(user.status === 'expired');

	function truncateNpub(npub) {
		if (!npub) return null;
		return npub.slice(0, 12) + '...' + npub.slice(-8);
	}

	function daysUntilExpiry(expiresAt) {
		if (!expiresAt) return Infinity;
		const now = new Date();
		const expiry = new Date(expiresAt);
		const diff = expiry - now;
		return Math.ceil(diff / (1000 * 60 * 60 * 24));
	}

	function formatExpiryDate(expiresAt) {
		if (!expiresAt) return 'Never';
		const days = daysUntilExpiry(expiresAt);
		if (days < 0) return 'Expired';
		if (days === 0) return 'Expires today';
		if (days === 1) return 'Expires tomorrow';
		if (days <= 7) return `Expires in ${days} days`;
		return formatDateInTimezone(expiresAt, { month: 'short', day: 'numeric', year: 'numeric' });
	}

	function formatSats(sats) {
		return sats.toLocaleString() + ' sats';
	}
</script>

<div class="flex items-center justify-between p-4 border rounded-lg {isExpired ? 'border-gray-200 dark:border-gray-600 bg-gray-50 dark:bg-gray-700' : isExpiringSoon ? 'border-amber-200 dark:border-amber-700 bg-amber-50 dark:bg-amber-900/20' : 'border-green-200 dark:border-green-700 bg-green-50 dark:bg-green-900/20'}">
	<div class="flex items-center space-x-4">
		<div class="w-10 h-10 rounded-full flex items-center justify-center {isExpired ? 'bg-gray-200 dark:bg-gray-600' : isExpiringSoon ? 'bg-amber-200 dark:bg-amber-900/50' : 'bg-green-200 dark:bg-green-900/50'}">
			<svg class="w-5 h-5 {isExpired ? 'text-gray-600 dark:text-gray-400' : isExpiringSoon ? 'text-amber-600 dark:text-amber-400' : 'text-green-600 dark:text-green-400'}" fill="none" stroke="currentColor" viewBox="0 0 24 24">
				<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z" />
			</svg>
		</div>
		<div>
			<div class="flex items-center space-x-2">
				<p class="font-medium text-gray-900 dark:text-gray-100">{displayName}</p>
				<span class="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium {isExpired ? 'bg-gray-200 dark:bg-gray-600 text-gray-700 dark:text-gray-300' : 'bg-purple-100 dark:bg-purple-900/30 text-purple-700 dark:text-purple-300'}">
					{user.tier}
				</span>
				{#if isExpired}
					<span class="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium bg-red-100 dark:bg-red-900/30 text-red-700 dark:text-red-400">
						Expired
					</span>
				{:else if isExpiringSoon}
					<span class="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium bg-amber-100 dark:bg-amber-900/30 text-amber-700 dark:text-amber-400">
						Expiring Soon
					</span>
				{/if}
			</div>
			<p class="text-sm text-gray-500 dark:text-gray-400 font-mono">{user.npub?.slice(0, 20)}...</p>
			<div class="flex items-center space-x-4 mt-1 text-xs text-gray-500 dark:text-gray-400">
				<span>Paid: {formatSats(user.amount_sats)}</span>
				<span>Date: {formatRelativeTime(user.created_at)}</span>
				{#if user.expires_at}
					<span class="{isExpiringSoon && !isExpired ? 'text-amber-600 dark:text-amber-400 font-medium' : ''}">
						{formatExpiryDate(user.expires_at)}
					</span>
				{:else}
					<span class="text-green-600 dark:text-green-400">Lifetime access</span>
				{/if}
			</div>
		</div>
	</div>
	<div class="flex items-center space-x-2">
		{#if user.event_count !== undefined}
			<span class="text-sm text-gray-500 dark:text-gray-400">{user.event_count.toLocaleString()} events</span>
		{/if}
		{#if !isExpired}
			<button
				type="button"
				onclick={() => onRevoke(user)}
				class="px-3 py-1.5 text-sm font-medium text-red-600 dark:text-red-400 hover:text-red-700 dark:hover:text-red-300 hover:bg-red-50 dark:hover:bg-red-900/20 rounded-lg transition-colors"
			>
				Revoke
			</button>
		{/if}
	</div>
</div>
