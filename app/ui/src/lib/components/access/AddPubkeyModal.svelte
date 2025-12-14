<script>
	import { access, setup } from '$lib/api/client.js';
	import { notify } from '$lib/stores/app.svelte.js';
	import Button from '$lib/components/Button.svelte';

	let { listType = 'whitelist', onClose, onAdd } = $props();

	let inputValue = $state('');
	let extraField = $state(''); // nickname for whitelist, reason for blacklist
	let validating = $state(false);
	let submitting = $state(false);
	let error = $state('');
	let validationResult = $state(null);

	let debounceTimer = null;

	function handleInput(e) {
		inputValue = e.target.value;
		error = '';
		validationResult = null;

		// Clear previous timer
		if (debounceTimer) clearTimeout(debounceTimer);

		// Don't validate empty input
		if (!inputValue.trim()) return;

		// Debounce validation
		debounceTimer = setTimeout(validateInput, 500);
	}

	async function validateInput() {
		if (!inputValue.trim()) return;

		validating = true;
		error = '';

		try {
			const result = await setup.validateIdentity(inputValue.trim());

			if (!result.valid) {
				error = result.error || 'Invalid identity';
				validationResult = null;
				return;
			}

			validationResult = result;
		} catch (e) {
			error = e.message || 'Failed to validate identity';
			validationResult = null;
		} finally {
			validating = false;
		}
	}

	async function handleSubmit() {
		if (!validationResult) {
			// Try to validate first
			await validateInput();
			if (!validationResult) return;
		}

		submitting = true;
		error = '';

		try {
			const data = {
				pubkey: validationResult.pubkey,
				npub: validationResult.npub
			};

			if (listType === 'whitelist') {
				data.nickname = extraField.trim() || '';
				await access.addToWhitelist(data);
				notify('success', `Added ${validationResult.npub?.slice(0, 12)}... to whitelist`);
			} else {
				data.reason = extraField.trim() || '';
				await access.addToBlacklist(data);
				notify('success', `Added ${validationResult.npub?.slice(0, 12)}... to blacklist`);
			}

			onAdd?.();
			onClose?.();
		} catch (e) {
			error = e.message || 'Failed to add pubkey';
		} finally {
			submitting = false;
		}
	}

	function handleBackdropClick(e) {
		if (e.target === e.currentTarget) {
			onClose?.();
		}
	}

	function handleKeydown(e) {
		if (e.key === 'Escape') {
			onClose?.();
		}
	}

	// Cleanup on unmount
	$effect(() => {
		return () => {
			if (debounceTimer) clearTimeout(debounceTimer);
		};
	});
</script>

<svelte:window onkeydown={handleKeydown} />

<!-- svelte-ignore a11y_click_events_have_key_events a11y_interactive_supports_focus -->
<!-- Modal backdrop -->
<div
	class="fixed inset-0 z-50 flex items-center justify-center bg-black/50 p-4"
	onclick={handleBackdropClick}
	role="dialog"
	aria-modal="true"
	aria-labelledby="modal-title"
>
	<!-- Modal content -->
	<div class="w-full max-w-md rounded-lg bg-white shadow-xl">
		<!-- Header -->
		<div class="flex items-center justify-between border-b px-6 py-4">
			<h2 id="modal-title" class="text-lg font-semibold text-gray-900">
				Add to {listType === 'whitelist' ? 'Whitelist' : 'Blacklist'}
			</h2>
			<button
				type="button"
				onclick={onClose}
				aria-label="Close modal"
				class="p-1 text-gray-400 hover:text-gray-600 transition-colors rounded hover:bg-gray-100"
			>
				<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
				</svg>
			</button>
		</div>

		<!-- Body -->
		<div class="p-6 space-y-4">
			<!-- Pubkey input -->
			<div>
				<label for="pubkey-input" class="block text-sm font-medium text-gray-700 mb-2">
					Pubkey or NIP-05 identifier
				</label>
				<div class="relative">
					<input
						type="text"
						id="pubkey-input"
						value={inputValue}
						oninput={handleInput}
						placeholder="npub1... or user@domain.com"
						class="input w-full pr-10"
						class:border-green-500={validationResult}
						class:border-red-500={error && !validating}
						disabled={submitting}
					/>
					{#if validating}
						<div class="absolute right-3 top-1/2 -translate-y-1/2">
							<div class="h-4 w-4 animate-spin rounded-full border-2 border-purple-600 border-t-transparent"></div>
						</div>
					{:else if validationResult}
						<div class="absolute right-3 top-1/2 -translate-y-1/2 text-green-500">
							<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
							</svg>
						</div>
					{/if}
				</div>
			</div>

			<!-- Extra field (nickname or reason) -->
			<div>
				<label for="extra-input" class="block text-sm font-medium text-gray-700 mb-2">
					{listType === 'whitelist' ? 'Nickname' : 'Reason'}
					<span class="text-gray-400">(optional)</span>
				</label>
				<input
					type="text"
					id="extra-input"
					bind:value={extraField}
					placeholder={listType === 'whitelist' ? 'e.g., Family, Friend' : 'e.g., Spam, Harassment'}
					class="input w-full"
					disabled={submitting}
				/>
			</div>

			<!-- Validation result -->
			{#if validationResult}
				<div class="p-3 bg-green-50 border border-green-200 rounded-lg">
					<div class="flex items-start space-x-2">
						<svg class="w-5 h-5 text-green-500 mt-0.5 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
						</svg>
						<div class="min-w-0">
							<p class="text-sm font-medium text-green-700">Validated</p>
							<p class="text-xs text-green-600 font-mono truncate">{validationResult.npub}</p>
							{#if validationResult.source === 'nip05'}
								<p class="text-xs text-green-600 mt-1">Resolved from NIP-05</p>
							{/if}
						</div>
					</div>
				</div>
			{/if}

			<!-- Error -->
			{#if error}
				<div class="p-3 bg-red-50 border border-red-200 rounded-lg">
					<div class="flex items-start space-x-2">
						<svg class="w-5 h-5 text-red-500 mt-0.5 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
						</svg>
						<p class="text-sm text-red-700">{error}</p>
					</div>
				</div>
			{/if}
		</div>

		<!-- Footer -->
		<div class="flex justify-end space-x-3 border-t px-6 py-4">
			<Button variant="secondary" onclick={onClose} disabled={submitting}>
				Cancel
			</Button>
			<Button
				variant="primary"
				onclick={handleSubmit}
				disabled={!validationResult || submitting}
				loading={submitting}
			>
				Add to {listType === 'whitelist' ? 'Whitelist' : 'Blacklist'}
			</Button>
		</div>
	</div>
</div>
