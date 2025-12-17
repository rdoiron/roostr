<script>
	let { label = '', value = '', subtext = '', tooltip = '', valueClass = 'text-gray-900 dark:text-gray-100' } = $props();
</script>

<div class="rounded-lg bg-white dark:bg-gray-800 p-6 shadow dark:shadow-gray-900/50">
	<p class="text-sm font-medium text-gray-500 dark:text-gray-400">{label}</p>
	<p class="mt-1 text-3xl font-bold {valueClass}" title={tooltip ? tooltip : undefined}>{value}</p>
	{#if subtext}
		<p class="mt-1 text-sm text-gray-500 dark:text-gray-400">{subtext}</p>
	{/if}
</div>
