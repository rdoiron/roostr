<script>
	import { access } from '$lib/api/client.js';
	import { notify } from '$lib/stores/app.svelte.js';

	let { mode = 'whitelist', onChange = null } = $props();

	let saving = $state(false);
	let selectedMode = $state('');

	// Initialize selectedMode from prop
	$effect(() => {
		if (!saving) {
			selectedMode = mode;
		}
	});

	const modes = [
		{
			id: 'open',
			name: 'Open',
			description: 'Anyone can write to your relay',
			icon: 'globe',
			warning: 'Not recommended for private relays'
		},
		{
			id: 'whitelist',
			name: 'Whitelist',
			description: 'Only allowed pubkeys can write',
			icon: 'shield',
			recommended: true
		},
		{
			id: 'paid',
			name: 'Paid Access',
			description: 'Whitelist + anyone who pays via Lightning',
			icon: 'lightning',
			note: 'Configure pricing in Paid Access settings'
		},
		{
			id: 'blacklist',
			name: 'Blacklist',
			description: 'Block specific pubkeys only',
			icon: 'ban'
		}
	];

	async function selectMode(newMode) {
		if (newMode === selectedMode || saving) return;

		const previousMode = selectedMode;
		selectedMode = newMode;
		saving = true;

		try {
			await access.setMode(newMode);
			notify('success', `Access mode changed to ${modes.find((m) => m.id === newMode)?.name}`);
			onChange?.(newMode);
		} catch (e) {
			selectedMode = previousMode;
			notify('error', e.message || 'Failed to change access mode');
		} finally {
			saving = false;
		}
	}
</script>

<div class="space-y-3">
	{#each modes as modeOption}
		<button
			type="button"
			onclick={() => selectMode(modeOption.id)}
			disabled={saving}
			class="w-full text-left p-4 rounded-lg border-2 transition-all {selectedMode === modeOption.id
				? 'border-purple-500 bg-purple-50 dark:bg-purple-900/30'
				: 'border-gray-200 dark:border-gray-600 hover:border-gray-300 dark:hover:border-gray-500 bg-white dark:bg-gray-700'} {saving ? 'opacity-50 cursor-wait' : ''}"
		>
			<div class="flex items-start space-x-3">
				<!-- Radio indicator -->
				<div
					class="w-5 h-5 rounded-full border-2 flex items-center justify-center mt-0.5 flex-shrink-0 {selectedMode === modeOption.id
						? 'border-purple-500'
						: 'border-gray-300 dark:border-gray-500'}"
				>
					{#if selectedMode === modeOption.id}
						<div class="w-2.5 h-2.5 rounded-full bg-purple-500"></div>
					{/if}
				</div>

				<!-- Icon -->
				<div
					class="w-8 h-8 rounded-lg flex items-center justify-center flex-shrink-0 {selectedMode === modeOption.id
						? 'bg-purple-100 dark:bg-purple-900/50 text-purple-600 dark:text-purple-400'
						: 'bg-gray-100 dark:bg-gray-600 text-gray-500 dark:text-gray-400'}"
				>
					{#if modeOption.icon === 'globe'}
						<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3.055 11H5a2 2 0 012 2v1a2 2 0 002 2 2 2 0 012 2v2.945M8 3.935V5.5A2.5 2.5 0 0010.5 8h.5a2 2 0 012 2 2 2 0 104 0 2 2 0 012-2h1.064M15 20.488V18a2 2 0 012-2h3.064M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
						</svg>
					{:else if modeOption.icon === 'shield'}
						<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.042-.133-2.052-.382-3.016z" />
						</svg>
					{:else if modeOption.icon === 'lightning'}
						<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 10V3L4 14h7v7l9-11h-7z" />
						</svg>
					{:else if modeOption.icon === 'ban'}
						<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M18.364 18.364A9 9 0 005.636 5.636m12.728 12.728A9 9 0 015.636 5.636m12.728 12.728L5.636 5.636" />
						</svg>
					{/if}
				</div>

				<!-- Content -->
				<div class="flex-1 min-w-0">
					<div class="flex flex-wrap items-center gap-x-2 gap-y-1">
						<p class="font-medium text-gray-900 dark:text-gray-100">{modeOption.name}</p>
						{#if modeOption.recommended}
							<span class="flex-shrink-0 rounded bg-green-100 dark:bg-green-900/30 px-2 py-0.5 text-xs font-medium text-green-700 dark:text-green-400">Recommended</span>
						{/if}
						{#if selectedMode === modeOption.id && saving}
							<div class="h-4 w-4 animate-spin rounded-full border-2 border-purple-600 border-t-transparent"></div>
						{/if}
					</div>
					<p class="text-sm text-gray-500 dark:text-gray-400">{modeOption.description}</p>
					{#if modeOption.warning}
						<p class="text-xs text-amber-600 mt-1">{modeOption.warning}</p>
					{/if}
					{#if modeOption.note && selectedMode === modeOption.id}
						<p class="text-xs text-purple-600 mt-1">{modeOption.note}</p>
					{/if}
				</div>
			</div>
		</button>
	{/each}

	<p class="text-xs text-gray-500 dark:text-gray-400 flex items-center space-x-1 mt-2">
		<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
			<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
		</svg>
		<span>Changes take effect immediately</span>
	</p>
</div>
