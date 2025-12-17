<script>
	import CopyButton from './CopyButton.svelte';
	import QRButton from './QRButton.svelte';

	let { label = '', url = '' } = $props();
</script>

<div class="flex items-center gap-3 rounded-lg bg-gray-50 dark:bg-gray-700/50 p-3">
	<div class="min-w-0 flex-1">
		<p class="mb-1 text-xs text-gray-500 dark:text-gray-400">{label}</p>
		<code class="block truncate text-sm text-gray-700 dark:text-gray-300">{url}</code>
	</div>
	<div class="flex flex-shrink-0 items-center gap-1">
		<CopyButton text={url} />
		<QRButton {url} />
	</div>
</div>
