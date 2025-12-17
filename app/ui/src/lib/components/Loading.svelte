<script>
	let { size = 'md', text = '' } = $props();

	const sizes = {
		sm: 'h-4 w-4 border-2',
		md: 'h-8 w-8 border-4',
		lg: 'h-12 w-12 border-4'
	};
</script>

<div class="flex flex-col items-center justify-center gap-3">
	<div
		class="animate-spin rounded-full border-purple-600 dark:border-purple-400 border-t-transparent {sizes[size]}"
	></div>
	{#if text}
		<p class="text-sm text-gray-500 dark:text-gray-400">{text}</p>
	{/if}
</div>
