<script>
	let status = $state('checking...');

	$effect(() => {
		fetch('/api/v1/health')
			.then((res) => res.json())
			.then((data) => {
				status = data.status || 'ok';
			})
			.catch(() => {
				status = 'offline';
			});
	});
</script>

<main class="container mx-auto px-4 py-8">
	<h1 class="text-3xl font-bold text-gray-900 mb-4">Roostr</h1>
	<p class="text-gray-600 mb-8">Your Private Roost on Nostr</p>

	<div class="bg-white rounded-lg shadow p-6">
		<h2 class="text-xl font-semibold mb-2">Relay Status</h2>
		<p class="text-gray-700">
			API Status:
			<span
				class={status === 'ok' ? 'text-green-600' : status === 'checking...' ? 'text-yellow-600' : 'text-red-600'}
			>
				{status}
			</span>
		</p>
	</div>
</main>
