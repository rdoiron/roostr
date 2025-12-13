<script>
	let { title = '', children } = $props();
</script>

<div class="rounded-lg bg-white p-6 shadow">
	{#if title}
		<h2 class="text-lg font-semibold text-gray-900">{title}</h2>
	{/if}
	<div class={title ? 'mt-4' : ''}>
		{@render children()}
	</div>
</div>
