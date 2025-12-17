<script>
	let { title = '', children } = $props();
</script>

<div class="rounded-lg bg-white dark:bg-gray-800 p-6 shadow dark:shadow-gray-900/50">
	{#if title}
		<h2 class="text-lg font-semibold text-gray-900 dark:text-gray-100">{title}</h2>
	{/if}
	<div class={title ? 'mt-4' : ''}>
		{@render children()}
	</div>
</div>
