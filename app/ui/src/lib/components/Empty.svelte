<script>
	let { icon = '📭', title = 'No data', message = '', action = null, actionLabel = '' } = $props();
</script>

<div class="flex flex-col items-center justify-center rounded-lg bg-gray-50 dark:bg-gray-800 px-6 py-12 text-center">
	<span class="text-4xl">{icon}</span>
	<h3 class="mt-4 text-lg font-medium text-gray-900 dark:text-gray-100">{title}</h3>
	{#if message}
		<p class="mt-1 text-sm text-gray-500 dark:text-gray-400">{message}</p>
	{/if}
	{#if action}
		<button
			type="button"
			onclick={action}
			class="mt-4 rounded-lg bg-purple-600 px-4 py-2 text-sm font-medium text-white hover:bg-purple-700"
		>
			{actionLabel}
		</button>
	{/if}
</div>
