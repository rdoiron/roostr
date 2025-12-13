<script>
	import { relayStatus } from '$lib/stores';

	let { onMenuClick } = $props();
</script>

<header class="flex h-16 items-center justify-between border-b border-gray-200 bg-white px-4 lg:px-6">
	<!-- Mobile menu button -->
	<button
		type="button"
		class="rounded-lg p-2 text-gray-500 hover:bg-gray-100 lg:hidden"
		onclick={onMenuClick}
	>
		<svg class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
			<path stroke-linecap="round" stroke-linejoin="round" d="M4 6h16M4 12h16M4 18h16" />
		</svg>
	</button>

	<!-- Page title area (can be customized per page) -->
	<div class="flex-1 lg:ml-0">
		<!-- Empty for now, pages can use slots or context -->
	</div>

	<!-- Right side -->
	<div class="flex items-center gap-4">
		<!-- Relay status indicator -->
		<div class="flex items-center gap-2 text-sm">
			<span
				class="h-2.5 w-2.5 rounded-full {relayStatus.online ? 'bg-green-500' : 'bg-red-500'}"
			></span>
			<span class="hidden text-gray-600 sm:inline">
				Relay {relayStatus.online ? 'Online' : 'Offline'}
			</span>
		</div>

		<!-- Quick actions dropdown could go here -->
	</div>
</header>
