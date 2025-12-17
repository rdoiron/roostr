<script>
	let { children } = $props();
</script>

<div class="min-h-screen bg-gray-50 dark:bg-gray-900">
	<div class="max-w-2xl mx-auto px-4 py-8">
		{@render children()}
	</div>
</div>
