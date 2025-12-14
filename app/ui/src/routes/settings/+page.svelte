<script>
	import { onMount } from 'svelte';
	import { config } from '$lib/api/client.js';
	import { notify } from '$lib/stores/app.svelte.js';
	import Loading from '$lib/components/Loading.svelte';
	import Error from '$lib/components/Error.svelte';

	// Default values for reset
	const DEFAULTS = {
		info: { name: '', description: '', contact: '', relay_icon: '' },
		limits: {
			max_event_bytes: 131072,
			max_ws_message_bytes: 131072,
			messages_per_sec: 3,
			max_subs_per_conn: 10,
			min_pow_difficulty: 0
		},
		authorization: { nip42_auth: false, event_kind_allowlist: [] }
	};

	// State
	let loading = $state(true);
	let error = $state(null);
	let saving = $state(false);

	// Form data
	let formData = $state({
		info: { ...DEFAULTS.info },
		limits: { ...DEFAULTS.limits },
		authorization: { ...DEFAULTS.authorization }
	});

	// Event kinds UI state
	let kindsMode = $state('all'); // 'all' or 'specific'
	let kindsInput = $state(''); // comma-separated string

	// PoW UI state
	let powEnabled = $state(false);

	// Validation errors
	let errors = $state({});

	// Load configuration
	async function loadConfig() {
		loading = true;
		error = null;
		try {
			const data = await config.get();
			formData.info = { ...DEFAULTS.info, ...data.info };
			formData.limits = { ...DEFAULTS.limits, ...data.limits };
			formData.authorization = { ...DEFAULTS.authorization, ...data.authorization };

			// Set event kinds mode and input
			const allowlist = formData.authorization.event_kind_allowlist || [];
			if (allowlist.length > 0) {
				kindsMode = 'specific';
				kindsInput = allowlist.join(', ');
			} else {
				kindsMode = 'all';
				kindsInput = '';
			}

			// Set PoW enabled state
			powEnabled = formData.limits.min_pow_difficulty > 0;
		} catch (e) {
			error = e.message;
		} finally {
			loading = false;
		}
	}

	// Validate a single field
	function validateField(section, field, value) {
		const key = `${section}.${field}`;

		// Clear previous error
		delete errors[key];

		if (section === 'info') {
			if (field === 'name' && value && value.length > 64) {
				errors[key] = 'Name must be 64 characters or less';
			}
			if (field === 'description' && value && value.length > 500) {
				errors[key] = 'Description must be 500 characters or less';
			}
			if (field === 'relay_icon' && value) {
				try {
					new URL(value);
				} catch {
					errors[key] = 'Must be a valid URL';
				}
			}
		}

		if (section === 'limits') {
			if (field === 'max_event_bytes' || field === 'max_ws_message_bytes') {
				if (value < 1024 || value > 16777216) {
					errors[key] = 'Must be between 1,024 and 16,777,216 bytes';
				}
			}
			if (field === 'messages_per_sec' || field === 'max_subs_per_conn') {
				if (value < 1 || value > 100) {
					errors[key] = 'Must be between 1 and 100';
				}
			}
			if (field === 'min_pow_difficulty') {
				if (value < 0 || value > 32) {
					errors[key] = 'Must be between 0 and 32';
				}
			}
		}

		// Trigger reactivity
		errors = { ...errors };
	}

	// Validate event kinds input
	function validateKinds() {
		delete errors['authorization.event_kind_allowlist'];

		if (kindsMode === 'specific' && kindsInput.trim()) {
			const parts = kindsInput.split(',').map((s) => s.trim()).filter(Boolean);
			for (const part of parts) {
				const num = parseInt(part, 10);
				if (isNaN(num) || num < 0) {
					errors['authorization.event_kind_allowlist'] = 'All event kinds must be non-negative integers';
					break;
				}
			}
		}

		errors = { ...errors };
	}

	// Parse event kinds from input
	function parseKinds() {
		if (kindsMode === 'all') {
			return [];
		}
		if (!kindsInput.trim()) {
			return [];
		}
		return kindsInput
			.split(',')
			.map((s) => s.trim())
			.filter(Boolean)
			.map((s) => parseInt(s, 10))
			.filter((n) => !isNaN(n) && n >= 0);
	}

	// Check if form has any validation errors
	let hasErrors = $derived(Object.keys(errors).length > 0);

	// Handle save
	async function handleSave() {
		// Run all validations
		validateField('info', 'name', formData.info.name);
		validateField('info', 'description', formData.info.description);
		validateField('info', 'relay_icon', formData.info.relay_icon);
		validateField('limits', 'max_event_bytes', formData.limits.max_event_bytes);
		validateField('limits', 'max_ws_message_bytes', formData.limits.max_ws_message_bytes);
		validateField('limits', 'messages_per_sec', formData.limits.messages_per_sec);
		validateField('limits', 'max_subs_per_conn', formData.limits.max_subs_per_conn);
		validateField('limits', 'min_pow_difficulty', formData.limits.min_pow_difficulty);
		validateKinds();

		if (hasErrors) {
			notify('error', 'Please fix validation errors before saving');
			return;
		}

		saving = true;
		try {
			// Build update payload
			const payload = {
				info: {
					name: formData.info.name,
					description: formData.info.description,
					contact: formData.info.contact,
					relay_icon: formData.info.relay_icon
				},
				limits: {
					max_event_bytes: formData.limits.max_event_bytes,
					max_ws_message_bytes: formData.limits.max_ws_message_bytes,
					messages_per_sec: formData.limits.messages_per_sec,
					max_subs_per_conn: formData.limits.max_subs_per_conn,
					min_pow_difficulty: powEnabled ? formData.limits.min_pow_difficulty : 0
				},
				authorization: {
					nip42_auth: formData.authorization.nip42_auth,
					event_kind_allowlist: parseKinds()
				}
			};

			await config.update(payload);
			notify('success', 'Configuration saved successfully');
		} catch (e) {
			notify('error', e.message || 'Failed to save configuration');
		} finally {
			saving = false;
		}
	}

	// Handle reset to defaults
	function handleReset() {
		if (!confirm('Reset all settings to defaults? This will not be saved until you click Save.')) {
			return;
		}

		formData.info = { ...DEFAULTS.info };
		formData.limits = { ...DEFAULTS.limits };
		formData.authorization = { ...DEFAULTS.authorization };
		kindsMode = 'all';
		kindsInput = '';
		powEnabled = false;
		errors = {};
		notify('info', 'Settings reset to defaults. Click Save to apply.');
	}

	// Format bytes for display
	function formatBytes(bytes) {
		if (bytes < 1024) return `${bytes} B`;
		if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`;
		return `${(bytes / (1024 * 1024)).toFixed(1)} MB`;
	}

	onMount(() => {
		loadConfig();
	});
</script>

<div class="space-y-6">
	<!-- Header -->
	<div>
		<h1 class="text-2xl font-bold text-gray-900">Relay Configuration</h1>
		<p class="text-gray-600">Configure your relay's identity, limits, and policies</p>
	</div>

	{#if loading}
		<div class="flex justify-center py-12">
			<Loading text="Loading configuration..." />
		</div>
	{:else if error}
		<Error message={error} onRetry={loadConfig} />
	{:else}
		<!-- Relay Identity Section -->
		<div class="rounded-lg bg-white p-6 shadow">
			<h2 class="text-lg font-semibold text-gray-900">Relay Identity (NIP-11)</h2>
			<p class="mt-1 text-sm text-gray-500">
				Public metadata visible to clients connecting to your relay
			</p>

			<div class="mt-4 space-y-4">
				<!-- Name -->
				<div>
					<label for="name" class="block text-sm font-medium text-gray-700">
						Name
						<span class="ml-1 text-xs text-gray-400">
							({formData.info.name?.length || 0}/64)
						</span>
					</label>
					<input
						id="name"
						type="text"
						maxlength="64"
						placeholder="My Private Relay"
						bind:value={formData.info.name}
						oninput={() => validateField('info', 'name', formData.info.name)}
						class="mt-1 w-full rounded-lg border px-4 py-2 focus:border-purple-500 focus:outline-none {errors['info.name'] ? 'border-red-300' : 'border-gray-300'}"
					/>
					{#if errors['info.name']}
						<p class="mt-1 text-xs text-red-600">{errors['info.name']}</p>
					{/if}
				</div>

				<!-- Description -->
				<div>
					<label for="description" class="block text-sm font-medium text-gray-700">
						Description
						<span class="ml-1 text-xs text-gray-400">
							({formData.info.description?.length || 0}/500)
						</span>
					</label>
					<textarea
						id="description"
						maxlength="500"
						rows="3"
						placeholder="A private Nostr relay for personal use..."
						bind:value={formData.info.description}
						oninput={() => validateField('info', 'description', formData.info.description)}
						class="mt-1 w-full rounded-lg border px-4 py-2 focus:border-purple-500 focus:outline-none {errors['info.description'] ? 'border-red-300' : 'border-gray-300'}"
					></textarea>
					{#if errors['info.description']}
						<p class="mt-1 text-xs text-red-600">{errors['info.description']}</p>
					{/if}
				</div>

				<!-- Contact -->
				<div>
					<label for="contact" class="block text-sm font-medium text-gray-700">
						Contact
					</label>
					<input
						id="contact"
						type="text"
						placeholder="admin@example.com or npub1..."
						bind:value={formData.info.contact}
						class="mt-1 w-full rounded-lg border border-gray-300 px-4 py-2 focus:border-purple-500 focus:outline-none"
					/>
					<p class="mt-1 text-xs text-gray-500">Email address or npub for relay contact</p>
				</div>

				<!-- Icon URL -->
				<div>
					<label for="relay_icon" class="block text-sm font-medium text-gray-700">
						Icon URL
						<span class="ml-1 text-xs text-gray-400">(optional)</span>
					</label>
					<input
						id="relay_icon"
						type="url"
						placeholder="https://example.com/relay-icon.png"
						bind:value={formData.info.relay_icon}
						oninput={() => validateField('info', 'relay_icon', formData.info.relay_icon)}
						class="mt-1 w-full rounded-lg border px-4 py-2 focus:border-purple-500 focus:outline-none {errors['info.relay_icon'] ? 'border-red-300' : 'border-gray-300'}"
					/>
					{#if errors['info.relay_icon']}
						<p class="mt-1 text-xs text-red-600">{errors['info.relay_icon']}</p>
					{/if}
				</div>
			</div>
		</div>

		<!-- Limits Section -->
		<div class="rounded-lg bg-white p-6 shadow">
			<h2 class="text-lg font-semibold text-gray-900">Limits</h2>
			<p class="mt-1 text-sm text-gray-500">
				Configure rate limits and size restrictions to protect your relay
			</p>

			<div class="mt-4 grid gap-4 sm:grid-cols-2">
				<!-- Max Event Size -->
				<div>
					<label for="max_event_bytes" class="block text-sm font-medium text-gray-700">
						Max Event Size
					</label>
					<div class="mt-1 flex items-center gap-2">
						<input
							id="max_event_bytes"
							type="number"
							min="1024"
							max="16777216"
							step="1024"
							bind:value={formData.limits.max_event_bytes}
							oninput={() => validateField('limits', 'max_event_bytes', formData.limits.max_event_bytes)}
							class="w-full rounded-lg border px-4 py-2 focus:border-purple-500 focus:outline-none {errors['limits.max_event_bytes'] ? 'border-red-300' : 'border-gray-300'}"
						/>
						<span class="text-sm text-gray-500 whitespace-nowrap">
							({formatBytes(formData.limits.max_event_bytes)})
						</span>
					</div>
					{#if errors['limits.max_event_bytes']}
						<p class="mt-1 text-xs text-red-600">{errors['limits.max_event_bytes']}</p>
					{:else}
						<p class="mt-1 text-xs text-gray-500">Maximum size of a single event in bytes</p>
					{/if}
				</div>

				<!-- Max WebSocket Message -->
				<div>
					<label for="max_ws_message_bytes" class="block text-sm font-medium text-gray-700">
						Max WebSocket Message
					</label>
					<div class="mt-1 flex items-center gap-2">
						<input
							id="max_ws_message_bytes"
							type="number"
							min="1024"
							max="16777216"
							step="1024"
							bind:value={formData.limits.max_ws_message_bytes}
							oninput={() => validateField('limits', 'max_ws_message_bytes', formData.limits.max_ws_message_bytes)}
							class="w-full rounded-lg border px-4 py-2 focus:border-purple-500 focus:outline-none {errors['limits.max_ws_message_bytes'] ? 'border-red-300' : 'border-gray-300'}"
						/>
						<span class="text-sm text-gray-500 whitespace-nowrap">
							({formatBytes(formData.limits.max_ws_message_bytes)})
						</span>
					</div>
					{#if errors['limits.max_ws_message_bytes']}
						<p class="mt-1 text-xs text-red-600">{errors['limits.max_ws_message_bytes']}</p>
					{:else}
						<p class="mt-1 text-xs text-gray-500">Maximum WebSocket message size in bytes</p>
					{/if}
				</div>

				<!-- Messages Per Second -->
				<div>
					<label for="messages_per_sec" class="block text-sm font-medium text-gray-700">
						Messages Per Second
					</label>
					<div class="mt-1 flex items-center gap-2">
						<input
							id="messages_per_sec"
							type="number"
							min="1"
							max="100"
							bind:value={formData.limits.messages_per_sec}
							oninput={() => validateField('limits', 'messages_per_sec', formData.limits.messages_per_sec)}
							class="w-full rounded-lg border px-4 py-2 focus:border-purple-500 focus:outline-none {errors['limits.messages_per_sec'] ? 'border-red-300' : 'border-gray-300'}"
						/>
						<span class="text-sm text-gray-500 whitespace-nowrap">per IP</span>
					</div>
					{#if errors['limits.messages_per_sec']}
						<p class="mt-1 text-xs text-red-600">{errors['limits.messages_per_sec']}</p>
					{:else}
						<p class="mt-1 text-xs text-gray-500">Rate limit per IP address</p>
					{/if}
				</div>

				<!-- Max Subscriptions -->
				<div>
					<label for="max_subs_per_conn" class="block text-sm font-medium text-gray-700">
						Max Subscriptions
					</label>
					<div class="mt-1 flex items-center gap-2">
						<input
							id="max_subs_per_conn"
							type="number"
							min="1"
							max="100"
							bind:value={formData.limits.max_subs_per_conn}
							oninput={() => validateField('limits', 'max_subs_per_conn', formData.limits.max_subs_per_conn)}
							class="w-full rounded-lg border px-4 py-2 focus:border-purple-500 focus:outline-none {errors['limits.max_subs_per_conn'] ? 'border-red-300' : 'border-gray-300'}"
						/>
						<span class="text-sm text-gray-500 whitespace-nowrap">per connection</span>
					</div>
					{#if errors['limits.max_subs_per_conn']}
						<p class="mt-1 text-xs text-red-600">{errors['limits.max_subs_per_conn']}</p>
					{:else}
						<p class="mt-1 text-xs text-gray-500">Maximum concurrent subscriptions per client</p>
					{/if}
				</div>
			</div>
		</div>

		<!-- Event Policies Section -->
		<div class="rounded-lg bg-white p-6 shadow">
			<h2 class="text-lg font-semibold text-gray-900">Event Policies</h2>
			<p class="mt-1 text-sm text-gray-500">
				Control which events your relay accepts
			</p>

			<div class="mt-4 space-y-6">
				<!-- Accepted Event Kinds -->
				<fieldset>
					<legend class="block text-sm font-medium text-gray-700 mb-2">
						Accepted Event Kinds
					</legend>
					<div class="space-y-2">
						<label class="flex items-center gap-2 cursor-pointer">
							<input
								type="radio"
								name="kindsMode"
								value="all"
								bind:group={kindsMode}
								class="h-4 w-4 text-purple-600 focus:ring-purple-500"
							/>
							<span class="text-sm text-gray-700">All kinds</span>
							<span class="text-xs text-gray-500">(no filtering)</span>
						</label>
						<label class="flex items-center gap-2 cursor-pointer">
							<input
								type="radio"
								name="kindsMode"
								value="specific"
								bind:group={kindsMode}
								class="h-4 w-4 text-purple-600 focus:ring-purple-500"
							/>
							<span class="text-sm text-gray-700">Specific kinds only</span>
						</label>
					</div>

					{#if kindsMode === 'specific'}
						<div class="mt-3">
							<input
								type="text"
								placeholder="0, 1, 3, 4, 5, 6, 7, 10002"
								bind:value={kindsInput}
								oninput={validateKinds}
								class="w-full rounded-lg border px-4 py-2 focus:border-purple-500 focus:outline-none {errors['authorization.event_kind_allowlist'] ? 'border-red-300' : 'border-gray-300'}"
							/>
							{#if errors['authorization.event_kind_allowlist']}
								<p class="mt-1 text-xs text-red-600">{errors['authorization.event_kind_allowlist']}</p>
							{:else}
								<p class="mt-1 text-xs text-gray-500">
									Comma-separated list of event kind numbers (e.g., 0=metadata, 1=note, 3=contacts, 7=reaction)
								</p>
							{/if}
						</div>
					{/if}
				</fieldset>

				<!-- NIP-42 Authentication -->
				<div class="border-t border-gray-200 pt-4">
					<label class="flex items-start gap-3 cursor-pointer">
						<input
							type="checkbox"
							bind:checked={formData.authorization.nip42_auth}
							class="mt-0.5 h-4 w-4 rounded border-gray-300 text-purple-600 focus:ring-purple-500"
						/>
						<div>
							<span class="text-sm font-medium text-gray-700">Require NIP-42 Authentication</span>
							<p class="text-xs text-gray-500 mt-0.5">
								Clients must authenticate with their private key before writing events
							</p>
						</div>
					</label>
				</div>

				<!-- Proof of Work -->
				<div class="border-t border-gray-200 pt-4">
					<label class="flex items-start gap-3 cursor-pointer">
						<input
							type="checkbox"
							bind:checked={powEnabled}
							class="mt-0.5 h-4 w-4 rounded border-gray-300 text-purple-600 focus:ring-purple-500"
						/>
						<div class="flex-1">
							<span class="text-sm font-medium text-gray-700">Require Proof of Work</span>
							<p class="text-xs text-gray-500 mt-0.5">
								Events must include proof of work to be accepted (spam prevention)
							</p>
						</div>
					</label>

					{#if powEnabled}
						<div class="mt-3 ml-7">
							<label for="min_pow_difficulty" class="block text-sm font-medium text-gray-700">
								Minimum Difficulty
							</label>
							<div class="mt-1 flex items-center gap-2">
								<input
									id="min_pow_difficulty"
									type="number"
									min="1"
									max="32"
									bind:value={formData.limits.min_pow_difficulty}
									oninput={() => validateField('limits', 'min_pow_difficulty', formData.limits.min_pow_difficulty)}
									class="w-24 rounded-lg border px-4 py-2 focus:border-purple-500 focus:outline-none {errors['limits.min_pow_difficulty'] ? 'border-red-300' : 'border-gray-300'}"
								/>
								<span class="text-sm text-gray-500">bits (1-32)</span>
							</div>
							{#if errors['limits.min_pow_difficulty']}
								<p class="mt-1 text-xs text-red-600">{errors['limits.min_pow_difficulty']}</p>
							{:else}
								<p class="mt-1 text-xs text-gray-500">
									Higher values require more computational work (16 is typical)
								</p>
							{/if}
						</div>
					{/if}
				</div>
			</div>
		</div>

		<!-- Action Buttons -->
		<div class="flex justify-end gap-3">
			<button
				type="button"
				onclick={handleReset}
				disabled={saving}
				class="rounded-lg border border-gray-300 bg-white px-4 py-2 text-sm font-medium text-gray-700 hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-purple-500 focus:ring-offset-2 disabled:opacity-50"
			>
				Reset Defaults
			</button>
			<button
				type="button"
				onclick={handleSave}
				disabled={saving || hasErrors}
				class="rounded-lg bg-purple-600 px-4 py-2 text-sm font-medium text-white hover:bg-purple-700 focus:outline-none focus:ring-2 focus:ring-purple-500 focus:ring-offset-2 disabled:opacity-50 flex items-center gap-2"
			>
				{#if saving}
					<svg class="animate-spin h-4 w-4" viewBox="0 0 24 24">
						<circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" fill="none"></circle>
						<path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
					</svg>
					Saving...
				{:else}
					Save Configuration
				{/if}
			</button>
		</div>
	{/if}
</div>
