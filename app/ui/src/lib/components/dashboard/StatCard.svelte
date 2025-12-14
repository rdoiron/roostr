<script>
	let { label = '', value = '', subtext = '', valueClass = 'text-gray-900' } = $props();
</script>

<div class="rounded-lg bg-white p-6 shadow">
	<p class="text-sm font-medium text-gray-500">{label}</p>
	<p class="mt-1 text-3xl font-bold {valueClass}">
		{value}
	</p>
	{#if subtext}
		<p class="mt-1 text-sm text-gray-500">{subtext}</p>
	{/if}
</div>
