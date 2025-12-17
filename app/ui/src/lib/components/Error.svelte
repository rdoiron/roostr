<script>
	let { title = 'Error', message = '', code = '', onRetry = null } = $props();
</script>

<div class="rounded-lg bg-red-50 dark:bg-red-900/20 p-4">
	<div class="flex items-start gap-3">
		<div class="flex-shrink-0">
			<svg class="h-5 w-5 text-red-400 dark:text-red-500" viewBox="0 0 20 20" fill="currentColor">
				<path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.28 7.22a.75.75 0 00-1.06 1.06L8.94 10l-1.72 1.72a.75.75 0 101.06 1.06L10 11.06l1.72 1.72a.75.75 0 101.06-1.06L11.06 10l1.72-1.72a.75.75 0 00-1.06-1.06L10 8.94 8.28 7.22z" clip-rule="evenodd" />
			</svg>
		</div>
		<div class="flex-1">
			<h3 class="text-sm font-medium text-red-800 dark:text-red-300">{title}</h3>
			{#if message}
				<p class="mt-1 text-sm text-red-700 dark:text-red-400">{message}</p>
			{/if}
			{#if code}
				<p class="mt-1 text-xs text-red-500 dark:text-red-400 font-mono">{code}</p>
			{/if}
			{#if onRetry}
				<button
					type="button"
					onclick={onRetry}
					class="mt-3 text-sm font-medium text-red-700 dark:text-red-400 hover:text-red-600 dark:hover:text-red-300"
				>
					Try again
				</button>
			{/if}
		</div>
	</div>
</div>
