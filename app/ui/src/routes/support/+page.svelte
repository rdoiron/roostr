<script>
	import { onMount } from 'svelte';
	import { support } from '$lib/api/client.js';
	import DonationCard from '$lib/components/support/DonationCard.svelte';
	import WebLNTipButton from '$lib/components/support/WebLNTipButton.svelte';

	let config = $state(null);
	let loading = $state(true);
	let error = $state(null);

	onMount(async () => {
		try {
			config = await support.getConfig();
		} catch (err) {
			error = err.message || 'Failed to load support configuration';
		} finally {
			loading = false;
		}
	});
</script>

<div class="space-y-6">
	<!-- Header -->
	<div>
		<h1 class="text-2xl font-bold text-gray-900">Support Roostr</h1>
		<p class="text-gray-600">Help keep the project alive and thriving</p>
	</div>

	<!-- Support Development Intro -->
	<div class="rounded-lg bg-gradient-to-r from-purple-600 to-indigo-600 p-6 text-white shadow">
		<div class="flex items-center gap-3">
			<span class="text-3xl">⚡</span>
			<div>
				<h2 class="text-xl font-bold">Support Development</h2>
				<p class="mt-1 text-purple-100">
					Roostr is free and open source. If you find it useful, consider supporting development!
				</p>
			</div>
		</div>
	</div>

	{#if loading}
		<div class="flex items-center justify-center py-12">
			<div class="h-8 w-8 animate-spin rounded-full border-4 border-gray-200 border-t-purple-600"></div>
		</div>
	{:else if error}
		<div class="rounded-lg bg-red-50 p-4 text-red-700">
			<p>{error}</p>
		</div>
	{:else if config}
		<!-- Donation Cards -->
		<div class="grid gap-6 md:grid-cols-2">
			<!-- Lightning -->
			<DonationCard
				title="Lightning"
				address={config.lightning_address}
				icon="⚡"
			>
				<WebLNTipButton lightningAddress={config.lightning_address} />
			</DonationCard>

			<!-- Bitcoin -->
			<DonationCard
				title="Bitcoin"
				address={config.bitcoin_address}
				icon="₿"
			/>
		</div>

		<!-- Get Help Section -->
		<div class="rounded-lg bg-white p-6 shadow">
			<h2 class="text-lg font-semibold text-gray-900">Get Help</h2>
			<div class="mt-4 space-y-3">
				<a
					href="{config.github_repo}"
					target="_blank"
					rel="noopener"
					class="flex items-center gap-3 rounded-lg bg-gray-50 p-3 transition-colors hover:bg-gray-100"
				>
					<span class="text-xl">📖</span>
					<div>
						<p class="font-medium text-gray-900">Documentation</p>
						<p class="text-sm text-gray-500">Read the docs on GitHub</p>
					</div>
				</a>
				<a
					href="{config.github_repo}/issues"
					target="_blank"
					rel="noopener"
					class="flex items-center gap-3 rounded-lg bg-gray-50 p-3 transition-colors hover:bg-gray-100"
				>
					<span class="text-xl">🐛</span>
					<div>
						<p class="font-medium text-gray-900">Report a Bug</p>
						<p class="text-sm text-gray-500">Open an issue on GitHub</p>
					</div>
				</a>
				<a
					href="{config.github_repo}/discussions"
					target="_blank"
					rel="noopener"
					class="flex items-center gap-3 rounded-lg bg-gray-50 p-3 transition-colors hover:bg-gray-100"
				>
					<span class="text-xl">💬</span>
					<div>
						<p class="font-medium text-gray-900">Discussions</p>
						<p class="text-sm text-gray-500">Join the community on GitHub</p>
					</div>
				</a>
				{#if config.developer_npub && config.developer_npub !== 'npub1...'}
					<a
						href="https://njump.me/{config.developer_npub}"
						target="_blank"
						rel="noopener"
						class="flex items-center gap-3 rounded-lg bg-gray-50 p-3 transition-colors hover:bg-gray-100"
					>
						<span class="text-xl">🦩</span>
						<div>
							<p class="font-medium text-gray-900">Follow on Nostr</p>
							<p class="text-sm text-gray-500">Stay updated on development</p>
						</div>
					</a>
				{/if}
			</div>
		</div>

		<!-- About Section -->
		<div class="rounded-lg bg-white p-6 shadow">
			<h2 class="text-lg font-semibold text-gray-900">About</h2>
			<div class="mt-4 space-y-3">
				<div class="flex items-center gap-3">
					<span class="text-2xl">🐓</span>
					<div>
						<p class="font-bold text-gray-900">Roostr v{config.version}</p>
						<p class="text-gray-600">"Your private roost on Nostr"</p>
					</div>
				</div>
				<div class="border-t border-gray-100 pt-3">
					<p class="text-sm text-gray-500">
						Built with Svelte, Go, and nostr-rs-relay
					</p>
					<p class="mt-1 text-sm text-gray-500">
						License: MIT
					</p>
				</div>
			</div>
		</div>
	{/if}
</div>
