<script>
	import { paidUsers } from '$lib/api/client.js';
	import { notify } from '$lib/stores/app.svelte.js';
	import Button from '$lib/components/Button.svelte';
	import PaidUserCard from './PaidUserCard.svelte';

	let { users = [], total = 0, onUpdate = () => {} } = $props();

	let filter = $state('all');
	let search = $state('');
	let showConfirmModal = $state(false);
	let userToRevoke = $state(null);
	let revoking = $state(false);
	let page = $state(0);
	const pageSize = 10;

	const filteredUsers = $derived(() => {
		let result = users;

		// Filter by status
		if (filter === 'active') {
			result = result.filter((u) => u.status === 'active');
		} else if (filter === 'expired') {
			result = result.filter((u) => u.status === 'expired');
		}

		// Filter by search
		if (search) {
			const searchLower = search.toLowerCase();
			result = result.filter(
				(u) =>
					u.npub?.toLowerCase().includes(searchLower) ||
					u.nickname?.toLowerCase().includes(searchLower) ||
					u.pubkey?.toLowerCase().includes(searchLower)
			);
		}

		return result;
	});

	const paginatedUsers = $derived(() => {
		const start = page * pageSize;
		return filteredUsers().slice(start, start + pageSize);
	});

	const totalPages = $derived(Math.ceil(filteredUsers().length / pageSize));

	function openRevokeModal(user) {
		userToRevoke = user;
		showConfirmModal = true;
	}

	function closeModal() {
		showConfirmModal = false;
		userToRevoke = null;
	}

	async function confirmRevoke() {
		if (!userToRevoke) return;

		revoking = true;
		try {
			await paidUsers.revoke(userToRevoke.pubkey);
			notify('success', 'Access revoked successfully');
			closeModal();
			onUpdate();
		} catch (e) {
			notify('error', e.message || 'Failed to revoke access');
		} finally {
			revoking = false;
		}
	}

	function handleKeydown(e) {
		if (e.key === 'Escape') closeModal();
	}
</script>

<svelte:window onkeydown={handleKeydown} />

<div class="rounded-lg bg-white p-6 shadow">
	<div class="flex items-center justify-between mb-4">
		<div>
			<h2 class="text-lg font-semibold text-gray-900">
				Paid Users
				<span class="ml-2 text-sm font-normal text-gray-500">({total})</span>
			</h2>
			<p class="text-sm text-gray-500">Users who purchased access via Lightning payment</p>
		</div>
	</div>

	<!-- Filters -->
	<div class="flex items-center justify-between mb-4 pb-4 border-b">
		<div class="flex items-center space-x-2">
			<select
				bind:value={filter}
				onchange={() => (page = 0)}
				class="px-3 py-2 border border-gray-300 rounded-lg text-sm focus:ring-2 focus:ring-purple-500 focus:border-transparent"
			>
				<option value="all">All Users</option>
				<option value="active">Active</option>
				<option value="expired">Expired</option>
			</select>
		</div>
		<div class="relative">
			<input
				type="text"
				bind:value={search}
				oninput={() => (page = 0)}
				placeholder="Search by npub..."
				class="w-64 pl-10 pr-4 py-2 border border-gray-300 rounded-lg text-sm focus:ring-2 focus:ring-purple-500 focus:border-transparent"
			/>
			<svg class="absolute left-3 top-2.5 w-4 h-4 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
				<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
			</svg>
		</div>
	</div>

	<!-- User List -->
	{#if filteredUsers().length === 0}
		<div class="text-center py-8">
			<div class="w-12 h-12 mx-auto bg-gray-100 rounded-full flex items-center justify-center mb-3">
				<svg class="w-6 h-6 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0zm6 3a2 2 0 11-4 0 2 2 0 014 0zM7 10a2 2 0 11-4 0 2 2 0 014 0z" />
				</svg>
			</div>
			{#if search || filter !== 'all'}
				<p class="text-gray-500">No users match your filters</p>
				<button
					type="button"
					onclick={() => { search = ''; filter = 'all'; }}
					class="mt-2 text-sm text-purple-600 hover:text-purple-500"
				>
					Clear filters
				</button>
			{:else}
				<p class="text-gray-500">No paid users yet</p>
				<p class="text-sm text-gray-400 mt-1">Users will appear here after purchasing access</p>
			{/if}
		</div>
	{:else}
		<div class="space-y-2">
			{#each paginatedUsers() as user (user.pubkey)}
				<PaidUserCard {user} onRevoke={openRevokeModal} />
			{/each}
		</div>

		<!-- Pagination -->
		{#if totalPages > 1}
			<div class="flex items-center justify-between mt-4 pt-4 border-t">
				<p class="text-sm text-gray-500">
					Showing {page * pageSize + 1}-{Math.min((page + 1) * pageSize, filteredUsers().length)} of {filteredUsers().length}
				</p>
				<div class="flex items-center space-x-2">
					<button
						type="button"
						onclick={() => page--}
						disabled={page === 0}
						class="px-3 py-1.5 text-sm font-medium text-gray-700 bg-white border border-gray-300 rounded-lg hover:bg-gray-50 disabled:opacity-50 disabled:cursor-not-allowed"
					>
						Previous
					</button>
					<span class="text-sm text-gray-500">Page {page + 1} of {totalPages}</span>
					<button
						type="button"
						onclick={() => page++}
						disabled={page >= totalPages - 1}
						class="px-3 py-1.5 text-sm font-medium text-gray-700 bg-white border border-gray-300 rounded-lg hover:bg-gray-50 disabled:opacity-50 disabled:cursor-not-allowed"
					>
						Next
					</button>
				</div>
			</div>
		{/if}
	{/if}
</div>

<!-- Revoke Confirmation Modal -->
{#if showConfirmModal && userToRevoke}
	<!-- svelte-ignore a11y_click_events_have_key_events a11y_interactive_supports_focus -->
	<div
		class="fixed inset-0 z-50 flex items-center justify-center bg-black/50"
		onclick={(e) => e.target === e.currentTarget && closeModal()}
		role="dialog"
		aria-modal="true"
		aria-labelledby="revoke-modal-title"
	>
		<div class="w-full max-w-md rounded-lg bg-white p-6 shadow-xl">
			<div class="flex items-center space-x-3 mb-4">
				<div class="w-10 h-10 rounded-full bg-red-100 flex items-center justify-center">
					<svg class="w-5 h-5 text-red-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
					</svg>
				</div>
				<div>
					<h3 id="revoke-modal-title" class="text-lg font-semibold text-gray-900">Revoke Access</h3>
					<p class="text-sm text-gray-500">This action cannot be undone</p>
				</div>
			</div>

			<p class="text-gray-600 mb-6">
				Are you sure you want to revoke access for <span class="font-medium">{userToRevoke.nickname || userToRevoke.npub?.slice(0, 20) + '...'}</span>?
				They will no longer be able to write to your relay.
			</p>

			<div class="flex items-center justify-end space-x-3">
				<Button variant="secondary" onclick={closeModal} disabled={revoking}>
					Cancel
				</Button>
				<Button variant="danger" onclick={confirmRevoke} loading={revoking}>
					Revoke Access
				</Button>
			</div>
		</div>
	</div>
{/if}
