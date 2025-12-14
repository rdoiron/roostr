<script>
	/**
	 * @type {'online' | 'offline' | 'degraded'}
	 */
	let { status = 'offline' } = $props();

	const colors = {
		online: 'bg-green-500',
		offline: 'bg-red-500',
		degraded: 'bg-yellow-500'
	};

	const labels = {
		online: 'Online',
		offline: 'Offline',
		degraded: 'Degraded'
	};
</script>

<div class="flex items-center gap-2">
	<span class="relative flex h-3 w-3">
		<span
			class="absolute inline-flex h-full w-full animate-ping rounded-full opacity-75 {colors[status]}"
		></span>
		<span class="relative inline-flex h-3 w-3 rounded-full {colors[status]}"></span>
	</span>
	<span class="font-medium">{labels[status]}</span>
</div>
