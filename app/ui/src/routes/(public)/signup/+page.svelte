<script>
	import { browser } from '$app/environment';
	import { signup } from '$lib/api/client.js';
	import PlanCard from '$lib/components/signup/PlanCard.svelte';
	import InvoiceDisplay from '$lib/components/signup/InvoiceDisplay.svelte';
	import PaymentConfirmation from '$lib/components/signup/PaymentConfirmation.svelte';

	// Flow state: 'loading' | 'unavailable' | 'plans' | 'identity' | 'payment' | 'success'
	let step = $state('loading');
	let relayInfo = $state(null);
	let error = $state(null);

	// User selections
	let selectedTier = $state(null);
	let pubkeyInput = $state('');
	let validatedPubkey = $state(null);
	let validating = $state(false);
	let validationError = $state(null);

	// Invoice state
	let invoice = $state(null);
	let creatingInvoice = $state(false);

	// Load relay info on mount
	$effect(() => {
		if (browser && step === 'loading') {
			loadRelayInfo();
		}
	});

	async function loadRelayInfo() {
		try {
			const info = await signup.getRelayInfo();
			relayInfo = info;

			if (!info.paid_access_enabled) {
				step = 'unavailable';
			} else if (!info.tiers || info.tiers.length === 0) {
				step = 'unavailable';
			} else {
				step = 'plans';
			}
		} catch (e) {
			error = e.message || 'Failed to load relay information';
			step = 'unavailable';
		}
	}

	function handleSelectPlan(tier) {
		selectedTier = tier;
		step = 'identity';
	}

	function handleBack() {
		if (step === 'identity') {
			step = 'plans';
			selectedTier = null;
		} else if (step === 'payment') {
			step = 'identity';
			invoice = null;
		}
	}

	async function validatePubkey() {
		if (!pubkeyInput.trim()) {
			validationError = 'Please enter your npub or NIP-05 address';
			return false;
		}

		validating = true;
		validationError = null;

		try {
			// Use the setup validation endpoint
			const res = await fetch(`/api/v1/setup/validate-identity?input=${encodeURIComponent(pubkeyInput.trim())}`);
			const result = await res.json();

			if (result.valid) {
				validatedPubkey = {
					pubkey: result.pubkey,
					npub: result.npub
				};
				return true;
			} else {
				validationError = result.error || 'Invalid pubkey or NIP-05 address';
				return false;
			}
		} catch (e) {
			validationError = e.message || 'Failed to validate pubkey';
			return false;
		} finally {
			validating = false;
		}
	}

	async function handleCreateInvoice() {
		const isValid = await validatePubkey();
		if (!isValid) return;

		creatingInvoice = true;
		error = null;

		try {
			const result = await signup.createInvoice({
				pubkey: validatedPubkey.pubkey,
				tier_id: selectedTier.id
			});

			invoice = result;
			step = 'payment';
		} catch (e) {
			error = e.message || 'Failed to create invoice';
		} finally {
			creatingInvoice = false;
		}
	}

	function handlePaymentConfirmed() {
		step = 'success';
	}

	function handleStartOver() {
		step = 'plans';
		selectedTier = null;
		pubkeyInput = '';
		validatedPubkey = null;
		invoice = null;
		error = null;
	}
</script>

<div class="text-center mb-8">
	<h1 class="text-3xl font-bold text-gray-900 dark:text-gray-100 mb-2">{relayInfo?.name || 'Private Relay'}</h1>
	<p class="text-gray-600 dark:text-gray-400">{relayInfo?.description || 'Private Nostr Relay'}</p>
</div>

{#if step === 'loading'}
	<div class="flex items-center justify-center py-16">
		<div class="h-8 w-8 animate-spin rounded-full border-4 border-purple-600 border-t-transparent"></div>
	</div>

{:else if step === 'unavailable'}
	<div class="rounded-lg bg-white dark:bg-gray-800 p-8 shadow dark:shadow-gray-900/50 text-center">
		<div class="w-16 h-16 mx-auto bg-gray-100 dark:bg-gray-700 rounded-full flex items-center justify-center mb-4">
			<svg class="w-8 h-8 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
				<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z" />
			</svg>
		</div>
		<h2 class="text-xl font-semibold text-gray-900 dark:text-gray-100 mb-2">Signup Unavailable</h2>
		<p class="text-gray-600 dark:text-gray-400">
			{error || 'This relay is not currently accepting new signups. Please contact the operator.'}
		</p>
		{#if relayInfo?.contact}
			<p class="text-sm text-gray-500 dark:text-gray-400 mt-4">
				Contact: <span class="font-mono">{relayInfo.contact}</span>
			</p>
		{/if}
	</div>

{:else if step === 'plans'}
	<div class="rounded-lg bg-white dark:bg-gray-800 p-6 shadow dark:shadow-gray-900/50 mb-6">
		<div class="text-center mb-6">
			<h2 class="text-xl font-semibold text-gray-900 dark:text-gray-100 mb-2">Choose Your Plan</h2>
			<p class="text-gray-600 dark:text-gray-400">Get reliable backup relay access for your Nostr activity</p>
		</div>

		<div class="grid gap-4 {relayInfo?.tiers?.length > 2 ? 'sm:grid-cols-2' : 'sm:grid-cols-' + relayInfo?.tiers?.length}">
			{#each relayInfo?.tiers || [] as tier (tier.id)}
				<PlanCard {tier} onSelect={() => handleSelectPlan(tier)} />
			{/each}
		</div>
	</div>

	<div class="rounded-lg bg-purple-50 dark:bg-purple-900/20 p-4 border border-purple-100 dark:border-purple-800">
		<h3 class="font-medium text-purple-900 dark:text-purple-100 mb-2">What you get:</h3>
		<ul class="space-y-1 text-sm text-purple-700 dark:text-purple-300">
			<li class="flex items-center">
				<svg class="w-4 h-4 mr-2 text-purple-500 dark:text-purple-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
				</svg>
				Always online, always syncing
			</li>
			<li class="flex items-center">
				<svg class="w-4 h-4 mr-2 text-purple-500 dark:text-purple-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
				</svg>
				Private &amp; censorship-resistant
			</li>
			<li class="flex items-center">
				<svg class="w-4 h-4 mr-2 text-purple-500 dark:text-purple-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
				</svg>
				Tor accessible from anywhere
			</li>
		</ul>
	</div>

{:else if step === 'identity'}
	<div class="rounded-lg bg-white dark:bg-gray-800 p-6 shadow dark:shadow-gray-900/50">
		<button
			type="button"
			onclick={handleBack}
			class="flex items-center text-sm text-gray-500 dark:text-gray-400 hover:text-gray-700 dark:hover:text-gray-300 mb-4"
		>
			<svg class="w-4 h-4 mr-1" fill="none" stroke="currentColor" viewBox="0 0 24 24">
				<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7" />
			</svg>
			Back to plans
		</button>

		<div class="text-center mb-6">
			<h2 class="text-xl font-semibold text-gray-900 dark:text-gray-100 mb-2">Your Nostr Identity</h2>
			<p class="text-gray-600 dark:text-gray-400">Enter your npub or NIP-05 address</p>
		</div>

		<!-- Selected plan summary -->
		<div class="mb-6 p-4 bg-purple-50 dark:bg-purple-900/20 rounded-lg border border-purple-100 dark:border-purple-800">
			<div class="flex items-center justify-between">
				<div>
					<p class="font-medium text-purple-900 dark:text-purple-100">{selectedTier?.name}</p>
					<p class="text-sm text-purple-600 dark:text-purple-400">
						{#if selectedTier?.duration_days}
							{selectedTier.duration_days === 30 ? 'Monthly' : selectedTier.duration_days === 90 ? 'Quarterly' : selectedTier.duration_days === 365 ? 'Annual' : `${selectedTier.duration_days} days`}
						{:else}
							Lifetime access
						{/if}
					</p>
				</div>
				<p class="text-xl font-bold text-purple-900 dark:text-purple-100">{selectedTier?.amount_sats?.toLocaleString()} sats</p>
			</div>
		</div>

		<div class="space-y-4">
			<div>
				<label for="pubkey" class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
					Your npub or NIP-05
				</label>
				<input
					type="text"
					id="pubkey"
					bind:value={pubkeyInput}
					placeholder="npub1... or alice@example.com"
					class="w-full px-4 py-3 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100 focus:ring-2 focus:ring-purple-500 focus:border-transparent"
					onkeydown={(e) => e.key === 'Enter' && handleCreateInvoice()}
				/>
				{#if validationError}
					<p class="mt-1 text-sm text-red-600 dark:text-red-400">{validationError}</p>
				{/if}
				{#if validatedPubkey}
					<p class="mt-1 text-sm text-green-600 dark:text-green-400 flex items-center">
						<svg class="w-4 h-4 mr-1" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
						</svg>
						Valid pubkey detected
					</p>
				{/if}
			</div>

			{#if error}
				<div class="p-3 bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-lg text-red-700 dark:text-red-400 text-sm">
					{error}
				</div>
			{/if}

			<button
				type="button"
				onclick={handleCreateInvoice}
				disabled={validating || creatingInvoice || !pubkeyInput.trim()}
				class="w-full py-3 px-4 bg-purple-600 hover:bg-purple-700 text-white font-medium rounded-lg transition-colors disabled:opacity-50 disabled:cursor-not-allowed flex items-center justify-center"
			>
				{#if validating || creatingInvoice}
					<div class="w-5 h-5 mr-2 animate-spin rounded-full border-2 border-white border-t-transparent"></div>
					{validating ? 'Validating...' : 'Creating Invoice...'}
				{:else}
					Continue to Payment
				{/if}
			</button>
		</div>
	</div>

{:else if step === 'payment'}
	<div class="rounded-lg bg-white dark:bg-gray-800 p-6 shadow dark:shadow-gray-900/50">
		<button
			type="button"
			onclick={handleBack}
			class="flex items-center text-sm text-gray-500 dark:text-gray-400 hover:text-gray-700 dark:hover:text-gray-300 mb-4"
		>
			<svg class="w-4 h-4 mr-1" fill="none" stroke="currentColor" viewBox="0 0 24 24">
				<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7" />
			</svg>
			Back
		</button>

		<div class="text-center mb-6">
			<h2 class="text-xl font-semibold text-gray-900 dark:text-gray-100 mb-2">Pay with Lightning</h2>
			<p class="text-gray-600 dark:text-gray-400">Scan the QR code or copy the invoice</p>
		</div>

		<InvoiceDisplay
			{invoice}
			tier={selectedTier}
			onPaymentConfirmed={handlePaymentConfirmed}
		/>
	</div>

{:else if step === 'success'}
	<PaymentConfirmation
		relayUrl={relayInfo?.relay_url}
		torUrl={relayInfo?.tor_url}
		onDone={handleStartOver}
	/>
{/if}

{#if relayInfo?.contact && step !== 'unavailable'}
	<div class="mt-8 text-center text-sm text-gray-500 dark:text-gray-400">
		Questions? Contact the operator: <span class="font-mono">{relayInfo.contact}</span>
	</div>
{/if}
