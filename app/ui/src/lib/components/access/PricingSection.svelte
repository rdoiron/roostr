<script>
	import { pricing } from '$lib/api/client.js';
	import { notify } from '$lib/stores/app.svelte.js';
	import Button from '$lib/components/Button.svelte';

	let { tiers = [], onUpdate = () => {} } = $props();

	let localTiers = $state([]);
	let saving = $state(false);
	let hasChanges = $state(false);

	// Initialize local state from props
	$effect(() => {
		if (tiers.length > 0 && localTiers.length === 0) {
			localTiers = tiers.map((t) => ({ ...t }));
		}
	});

	// Check for changes
	$effect(() => {
		if (tiers.length > 0 && localTiers.length > 0) {
			hasChanges = JSON.stringify(tiers) !== JSON.stringify(localTiers);
		}
	});

	function handleToggle(tierId) {
		localTiers = localTiers.map((t) =>
			t.id === tierId ? { ...t, enabled: !t.enabled } : t
		);
	}

	function handlePriceChange(tierId, value) {
		const sats = parseInt(value) || 0;
		localTiers = localTiers.map((t) =>
			t.id === tierId ? { ...t, amount_sats: sats } : t
		);
	}

	async function handleSave() {
		saving = true;
		try {
			await pricing.update(localTiers);
			notify('success', 'Pricing updated successfully');
			onUpdate();
		} catch (e) {
			notify('error', e.message || 'Failed to update pricing');
		} finally {
			saving = false;
		}
	}

	function handleReset() {
		localTiers = tiers.map((t) => ({ ...t }));
	}

	function getTierDescription(tier) {
		if (!tier.duration_days) return 'Pay once, access forever';
		if (tier.duration_days === 30) return 'Renews monthly';
		if (tier.duration_days === 90) return 'Renews quarterly';
		if (tier.duration_days === 365) return 'Renews annually';
		return `${tier.duration_days} days access`;
	}
</script>

<div class="rounded-lg bg-white p-6 shadow">
	<div class="flex items-center justify-between mb-4">
		<div>
			<h2 class="text-lg font-semibold text-gray-900">Pricing Tiers</h2>
			<p class="text-sm text-gray-500">Configure which pricing tiers to offer and their prices</p>
		</div>
	</div>

	{#if localTiers.length === 0}
		<div class="flex items-center justify-center py-8">
			<div class="h-6 w-6 animate-spin rounded-full border-2 border-purple-600 border-t-transparent"></div>
		</div>
	{:else}
		<div class="space-y-4">
			{#each localTiers as tier (tier.id)}
				<div class="flex items-center justify-between p-4 border rounded-lg {tier.enabled ? 'border-purple-200 bg-purple-50' : 'border-gray-200 bg-gray-50'}">
					<div class="flex items-center space-x-4">
						<label class="relative inline-flex items-center cursor-pointer">
							<input
								type="checkbox"
								checked={tier.enabled}
								onchange={() => handleToggle(tier.id)}
								class="sr-only peer"
							/>
							<div class="w-11 h-6 bg-gray-200 peer-focus:outline-none peer-focus:ring-4 peer-focus:ring-purple-300 rounded-full peer peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:border-gray-300 after:border after:rounded-full after:h-5 after:w-5 after:transition-all peer-checked:bg-purple-600"></div>
						</label>
						<div>
							<p class="font-medium text-gray-900">{tier.name}</p>
							<p class="text-sm text-gray-500">{getTierDescription(tier)}</p>
						</div>
					</div>
					<div class="flex items-center space-x-2">
						<input
							type="number"
							min="0"
							value={tier.amount_sats}
							oninput={(e) => handlePriceChange(tier.id, e.target.value)}
							disabled={!tier.enabled}
							class="w-28 px-3 py-2 border border-gray-300 rounded-lg text-right focus:ring-2 focus:ring-purple-500 focus:border-transparent disabled:bg-gray-100 disabled:text-gray-400"
						/>
						<span class="text-sm text-gray-500 w-8">sats</span>
					</div>
				</div>
			{/each}
		</div>

		<div class="flex items-center justify-between mt-6 pt-4 border-t">
			<p class="text-sm text-gray-500">
				Enable at least one tier for users to purchase access
			</p>
			<div class="flex items-center space-x-3">
				{#if hasChanges}
					<Button variant="secondary" onclick={handleReset} disabled={saving}>
						Reset
					</Button>
				{/if}
				<Button variant="primary" onclick={handleSave} disabled={!hasChanges} loading={saving}>
					Save Changes
				</Button>
			</div>
		</div>
	{/if}
</div>
