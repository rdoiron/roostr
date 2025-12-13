<script>
	import '../app.css';
	import Sidebar from '$lib/components/Sidebar.svelte';
	import Header from '$lib/components/Header.svelte';
	import { setupState } from '$lib/stores';

	let { children } = $props();

	let sidebarOpen = $state(false);
</script>

<div class="min-h-screen bg-gray-50">
	{#if !setupState.completed && setupState.loading === false}
		<!-- Setup wizard - no navigation -->
		<main class="min-h-screen">
			{@render children()}
		</main>
	{:else}
		<!-- Main app layout with sidebar -->
		<div class="flex h-screen overflow-hidden">
			<!-- Sidebar -->
			<Sidebar bind:open={sidebarOpen} />

			<!-- Main content area -->
			<div class="flex flex-1 flex-col overflow-hidden">
				<!-- Header -->
				<Header onMenuClick={() => (sidebarOpen = !sidebarOpen)} />

				<!-- Page content -->
				<main class="flex-1 overflow-y-auto p-6">
					{@render children()}
				</main>
			</div>
		</div>
	{/if}
</div>
