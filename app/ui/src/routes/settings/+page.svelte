<script>
	// Settings page - placeholder
</script>

<div class="space-y-6">
	<div>
		<h1 class="text-2xl font-bold text-gray-900">Settings</h1>
		<p class="text-gray-600">Configure your relay</p>
	</div>

	<div class="rounded-lg bg-white p-6 shadow">
		<h2 class="text-lg font-semibold text-gray-900">Relay Identity</h2>
		<div class="mt-4 space-y-4">
			<div>
				<label class="block text-sm font-medium text-gray-700">Relay Name</label>
				<input
					type="text"
					placeholder="My Private Relay"
					class="mt-1 w-full rounded-lg border border-gray-300 px-4 py-2 focus:border-purple-500 focus:outline-none"
				/>
			</div>
			<div>
				<label class="block text-sm font-medium text-gray-700">Description</label>
				<textarea
					placeholder="A private Nostr relay..."
					rows="3"
					class="mt-1 w-full rounded-lg border border-gray-300 px-4 py-2 focus:border-purple-500 focus:outline-none"
				></textarea>
			</div>
		</div>
	</div>

	<div class="rounded-lg bg-white p-6 shadow">
		<h2 class="text-lg font-semibold text-gray-900">Limits</h2>
		<p class="mt-2 text-gray-500">Configure rate limits and size restrictions</p>
		<p class="mt-4 text-sm text-gray-400">Coming soon...</p>
	</div>

	<div class="rounded-lg bg-white p-6 shadow">
		<h2 class="text-lg font-semibold text-gray-900">Relay Control</h2>
		<div class="mt-4 flex gap-3">
			<button class="rounded-lg bg-yellow-100 px-4 py-2 text-sm font-medium text-yellow-700 hover:bg-yellow-200">
				Reload Config
			</button>
			<button class="rounded-lg bg-red-100 px-4 py-2 text-sm font-medium text-red-700 hover:bg-red-200">
				Restart Relay
			</button>
		</div>
	</div>
</div>
