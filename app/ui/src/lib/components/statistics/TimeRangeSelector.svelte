<script>
	let { value = '7days', onchange = () => {} } = $props();

	const options = [
		{ value: 'today', label: 'Today' },
		{ value: '7days', label: '7 Days' },
		{ value: '30days', label: '30 Days' },
		{ value: 'alltime', label: 'All Time' }
	];
</script>

<div class="flex gap-1 rounded-lg bg-gray-100 dark:bg-gray-700 p-1">
	{#each options as option}
		<button
			class="rounded-md px-3 py-1.5 text-sm font-medium transition-colors
				{value === option.value
					? 'bg-purple-600 text-white'
					: 'text-gray-600 dark:text-gray-300 hover:bg-gray-200 dark:hover:bg-gray-600 hover:text-gray-900 dark:hover:text-gray-100'}"
			onclick={() => onchange(option.value)}
		>
			{option.label}
		</button>
	{/each}
</div>
