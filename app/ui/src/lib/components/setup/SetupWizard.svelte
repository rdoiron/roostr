<script>
	import { setup, access } from '$lib/api/client.js';
	import { notify } from '$lib/stores';
	import Button from '$lib/components/Button.svelte';
	import Loading from '$lib/components/Loading.svelte';
	import WelcomeStep from './WelcomeStep.svelte';
	import IdentityStep from './IdentityStep.svelte';
	import RelayInfoStep from './RelayInfoStep.svelte';
	import AccessModeStep from './AccessModeStep.svelte';
	import AddOthersStep from './AddOthersStep.svelte';
	import CompleteStep from './CompleteStep.svelte';

	// Wizard state
	let currentStep = $state(0);
	let submitting = $state(false);
	let error = $state('');

	// Collected data across steps
	let wizardData = $state({
		operatorIdentity: '',
		operatorPubkey: '',
		operatorNpub: '',
		relayName: '',
		relayDescription: '',
		accessMode: 'private',
		additionalPubkeys: []
	});

	// Step validation states
	let stepValid = $state({
		identity: false,
		relayInfo: false,
		accessMode: true, // Always valid (has default)
		addOthers: true // Optional step
	});

	const totalSteps = 5;

	function canProceed() {
		switch (currentStep) {
			case 0:
				return true; // Welcome
			case 1:
				return stepValid.identity;
			case 2:
				return stepValid.relayInfo;
			case 3:
				return stepValid.accessMode;
			case 4:
				return stepValid.addOthers;
			default:
				return false;
		}
	}

	function goBack() {
		if (currentStep > 0) {
			currentStep--;
		}
	}

	async function goNext() {
		if (currentStep < 4) {
			currentStep++;
		} else if (currentStep === 4) {
			// Submit setup
			await completeSetup();
		}
	}

	async function completeSetup() {
		submitting = true;
		error = '';

		try {
			// 1. Complete setup with operator info
			await setup.complete({
				operator_identity: wizardData.operatorIdentity,
				relay_name: wizardData.relayName,
				relay_description: wizardData.relayDescription,
				access_mode: wizardData.accessMode
			});

			// 2. Add additional pubkeys to whitelist
			for (const pubkey of wizardData.additionalPubkeys) {
				try {
					await access.addToWhitelist({
						pubkey: pubkey.pubkey,
						npub: pubkey.npub,
						nickname: pubkey.nickname || ''
					});
				} catch (e) {
					console.error('Failed to add pubkey to whitelist:', pubkey, e);
					// Continue with other pubkeys
				}
			}

			// 3. Move to complete step
			currentStep = 5;
		} catch (e) {
			error = e.message || 'Failed to complete setup';
			notify('error', error);
		} finally {
			submitting = false;
		}
	}

	function handleIdentityChange(data) {
		wizardData.operatorIdentity = data.identity;
		wizardData.operatorPubkey = data.pubkey;
		wizardData.operatorNpub = data.npub;
		stepValid.identity = data.valid;
	}

	function handleRelayInfoChange(data) {
		wizardData.relayName = data.name;
		wizardData.relayDescription = data.description;
		stepValid.relayInfo = data.valid;
	}

	function handleAccessModeChange(mode) {
		wizardData.accessMode = mode;
	}

	function handleAddOthersChange(pubkeys) {
		wizardData.additionalPubkeys = pubkeys;
	}
</script>

<div class="min-h-screen bg-gray-50 flex items-center justify-center p-4">
	<div class="w-full max-w-2xl">
		<!-- Progress indicator (hidden on welcome and complete) -->
		{#if currentStep > 0 && currentStep < 5}
			<div class="mb-8">
				<div class="flex items-center justify-between mb-2">
					<span class="text-sm text-gray-500">Step {currentStep} of {totalSteps}</span>
				</div>
				<div class="h-2 bg-gray-200 rounded-full overflow-hidden">
					<div
						class="h-full bg-purple-600 transition-all duration-300"
						style="width: {(currentStep / totalSteps) * 100}%"
					></div>
				</div>
			</div>
		{/if}

		<!-- Step content -->
		<div class="bg-white rounded-xl shadow-lg p-8">
			{#if currentStep === 0}
				<WelcomeStep />
			{:else if currentStep === 1}
				<IdentityStep
					identity={wizardData.operatorIdentity}
					onChange={handleIdentityChange}
				/>
			{:else if currentStep === 2}
				<RelayInfoStep
					name={wizardData.relayName}
					description={wizardData.relayDescription}
					operatorNpub={wizardData.operatorNpub}
					onChange={handleRelayInfoChange}
				/>
			{:else if currentStep === 3}
				<AccessModeStep
					mode={wizardData.accessMode}
					onChange={handleAccessModeChange}
				/>
			{:else if currentStep === 4}
				<AddOthersStep
					operatorPubkey={wizardData.operatorPubkey}
					operatorNpub={wizardData.operatorNpub}
					pubkeys={wizardData.additionalPubkeys}
					onChange={handleAddOthersChange}
				/>
			{:else if currentStep === 5}
				<CompleteStep />
			{/if}

			<!-- Error display -->
			{#if error}
				<div class="mt-4 p-3 bg-red-50 border border-red-200 rounded-lg text-red-700 text-sm">
					{error}
				</div>
			{/if}

			<!-- Navigation buttons -->
			{#if currentStep < 5}
				<div class="mt-8 flex items-center justify-between">
					{#if currentStep > 0}
						<Button variant="secondary" onclick={goBack} disabled={submitting}>
							Back
						</Button>
					{:else}
						<div></div>
					{/if}

					<Button
						variant="primary"
						onclick={goNext}
						disabled={!canProceed() || submitting}
						loading={submitting}
					>
						{#if currentStep === 0}
							Get Started
						{:else if currentStep === 4}
							Finish Setup
						{:else}
							Continue
						{/if}
					</Button>
				</div>
			{/if}
		</div>
	</div>
</div>
