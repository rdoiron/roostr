<script>
	let { mode = 'private', onChange } = $props();

	let selectedMode = $state(mode);

	// Sync local state with prop when it changes (for back navigation)
	$effect(() => {
		if (mode !== selectedMode) {
			selectedMode = mode;
		}
	});

	function selectMode(newMode) {
		selectedMode = newMode;
		onChange(newMode);
	}

	const modes = [
		{
			id: 'private',
			name: 'Private',
			recommended: true,
			description: 'Only you and people you whitelist can write.',
			detail: 'Best for personal backup and family use.',
			icon: 'lock'
		},
		{
			id: 'paid',
			name: 'Paid Access',
			recommended: false,
			description: 'Whitelist + anyone who pays via Lightning.',
			detail: 'Requires Lightning node setup (can configure later).',
			icon: 'bolt'
		},
		{
			id: 'public',
			name: 'Public',
			recommended: false,
			description: 'Anyone can write. Not recommended for home servers.',
			detail: 'May fill your disk quickly with spam.',
			icon: 'globe',
			warning: true
		}
	];
</script>

<div>
	<h2 class="text-2xl font-bold text-gray-900 mb-2">Who Can Use Your Relay?</h2>
	<p class="text-gray-600 mb-6">
		Choose who can write events to your relay.
	</p>

	<div class="space-y-3">
		{#each modes as modeOption}
			<button
				type="button"
				onclick={() => selectMode(modeOption.id)}
				class="w-full text-left p-4 rounded-lg border-2 transition-all {selectedMode === modeOption.id ? 'border-purple-500 bg-purple-50' : 'border-gray-200 hover:border-gray-300 bg-white'}"
			>
				<div class="flex items-start space-x-3">
					<!-- Radio indicator -->
					<div class="mt-0.5">
						<div class="w-5 h-5 rounded-full border-2 flex items-center justify-center {selectedMode === modeOption.id ? 'border-purple-500' : 'border-gray-300'}">
							{#if selectedMode === modeOption.id}
								<div class="w-2.5 h-2.5 rounded-full bg-purple-500"></div>
							{/if}
						</div>
					</div>

					<!-- Content -->
					<div class="flex-1">
						<div class="flex items-center space-x-2">
							<span class="font-medium text-gray-900">{modeOption.name}</span>
							{#if modeOption.recommended}
								<span class="text-xs bg-purple-100 text-purple-700 px-2 py-0.5 rounded-full">Recommended</span>
							{/if}
							{#if modeOption.warning}
								<span class="text-xs bg-yellow-100 text-yellow-700 px-2 py-0.5 rounded-full">Caution</span>
							{/if}
						</div>
						<p class="text-sm text-gray-600 mt-1">{modeOption.description}</p>
						<p class="text-xs text-gray-500 mt-1">{modeOption.detail}</p>
					</div>

					<!-- Icon -->
					<div class="flex-shrink-0">
						{#if modeOption.icon === 'lock'}
							<svg class="w-6 h-6 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z" />
							</svg>
						{:else if modeOption.icon === 'bolt'}
							<svg class="w-6 h-6 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 10V3L4 14h7v7l9-11h-7z" />
							</svg>
						{:else if modeOption.icon === 'globe'}
							<svg class="w-6 h-6 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3.055 11H5a2 2 0 012 2v1a2 2 0 002 2 2 2 0 012 2v2.945M8 3.935V5.5A2.5 2.5 0 0010.5 8h.5a2 2 0 012 2 2 2 0 104 0 2 2 0 012-2h1.064M15 20.488V18a2 2 0 012-2h3.064M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
							</svg>
						{/if}
					</div>
				</div>
			</button>
		{/each}
	</div>
</div>
