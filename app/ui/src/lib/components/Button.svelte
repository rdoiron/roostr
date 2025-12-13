<script>
	/**
	 * Reusable button component.
	 * @prop {'primary' | 'secondary' | 'danger'} variant - Button style variant
	 * @prop {boolean} disabled - Whether button is disabled
	 * @prop {boolean} loading - Show loading spinner
	 */
	let { variant = 'primary', disabled = false, loading = false, onclick, children } = $props();

	const baseClasses = 'px-4 py-2 rounded-lg font-medium transition-colors focus:outline-none focus:ring-2 focus:ring-offset-2';

	const variantClasses = {
		primary: 'bg-purple-600 text-white hover:bg-purple-700 focus:ring-purple-500 disabled:bg-purple-300',
		secondary: 'bg-gray-200 text-gray-900 hover:bg-gray-300 focus:ring-gray-500 disabled:bg-gray-100',
		danger: 'bg-red-600 text-white hover:bg-red-700 focus:ring-red-500 disabled:bg-red-300'
	};
</script>

<button
	class="{baseClasses} {variantClasses[variant]}"
	disabled={disabled || loading}
	{onclick}
>
	{#if loading}
		<span class="inline-block animate-spin mr-2">⟳</span>
	{/if}
	{@render children()}
</button>
