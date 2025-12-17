<script>
	let { name = '', description = '', operatorNpub = '', onChange } = $props();

	let nameValue = $state(name);
	let descriptionValue = $state(description);
	let initialized = $state(false);

	// Sync local state with props when they change (for back navigation)
	$effect(() => {
		if (!initialized && name) {
			nameValue = name;
			descriptionValue = description;
			initialized = true;
		}
	});

	// Notify parent of changes
	$effect(() => {
		const valid = nameValue.trim().length > 0;
		onChange({
			name: nameValue,
			description: descriptionValue,
			valid
		});
	});
</script>

<div>
	<h2 class="text-2xl font-bold text-gray-900 dark:text-white mb-2">Name Your Relay</h2>
	<p class="text-gray-600 dark:text-gray-400 mb-6">
		Give your relay a name and description. This is shown to clients that connect.
	</p>

	<!-- Relay Name -->
	<div class="mb-4">
		<label for="relay-name" class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
			Relay Name <span class="text-red-500">*</span>
		</label>
		<input
			type="text"
			id="relay-name"
			bind:value={nameValue}
			placeholder="My Private Relay"
			class="input w-full"
		/>
	</div>

	<!-- Description -->
	<div class="mb-4">
		<label for="description" class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
			Description <span class="text-gray-400 dark:text-gray-500">(optional)</span>
		</label>
		<textarea
			id="description"
			bind:value={descriptionValue}
			placeholder="Personal backup relay for family and friends."
			rows="3"
			class="input w-full resize-none"
		></textarea>
	</div>

	<!-- Contact (auto-filled, read-only) -->
	<div class="mb-4">
		<label for="contact" class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
			Contact <span class="text-gray-400 dark:text-gray-500">(auto-filled)</span>
		</label>
		<input
			type="text"
			id="contact"
			value={operatorNpub}
			readonly
			class="input w-full bg-gray-50 dark:bg-gray-800 text-gray-500 dark:text-gray-400 cursor-not-allowed font-mono text-sm"
		/>
		<p class="text-xs text-gray-500 dark:text-gray-400 mt-1">
			Your pubkey from the previous step will be used as the relay contact.
		</p>
	</div>
</div>
