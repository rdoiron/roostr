<script>
	import { browser } from '$app/environment';
	import { storage } from '$lib/api/client.js';
	import { notify } from '$lib/stores/app.svelte.js';
	import { formatBytes, formatRelativeTime } from '$lib/utils/format.js';
	import { formatDateInTimezone } from '$lib/stores/timezone.svelte.js';

	import Loading from '$lib/components/Loading.svelte';
	import Error from '$lib/components/Error.svelte';
	import Button from '$lib/components/Button.svelte';
	import StorageProgressBar from '$lib/components/storage/StorageProgressBar.svelte';

	let loading = $state(true);
	let error = $state(null);

	// Storage status data
	let storageStatus = $state(null);

	// Retention policy data
	let retentionPolicy = $state({
		retention_days: 0,
		exceptions: [],
		honor_nip09: true
	});
	let retentionMode = $state('never'); // never, 1year, 6months, 90days, 30days, custom
	let customDays = $state(90);

	// Cleanup form
	let cleanupDate = $state('');
	let cleanupEstimate = $state(null);
	let estimateLoading = $state(false);
	let applyCleanupExceptions = $state(false);

	// Operation states
	let savingRetention = $state(false);
	let runningRetention = $state(false);
	let runningCleanup = $state(false);
	let runningVacuum = $state(false);
	let runningIntegrity = $state(false);
	let integrityResult = $state(null);


	// Cleanup confirmation
	let showCleanupConfirm = $state(false);

	// Exception checkboxes
	let exceptOperator = $state(false);
	let exceptProfiles = $state(false);
	let exceptFollows = $state(false);
	let exceptDMs = $state(false);

	async function loadData() {
		try {
			const [statusRes, retentionRes] = await Promise.all([
				storage.getStatus(),
				storage.getRetention()
			]);

			storageStatus = statusRes;
			retentionPolicy = retentionRes;

			// Set UI state from retention policy
			setRetentionUIFromPolicy(retentionRes);

			error = null;
		} catch (e) {
			error = e.message;
		} finally {
			loading = false;
		}
	}

	function setRetentionUIFromPolicy(policy) {
		// Set retention mode
		const days = policy.retention_days;
		if (days === 0) {
			retentionMode = 'never';
		} else if (days === 365) {
			retentionMode = '1year';
		} else if (days === 180) {
			retentionMode = '6months';
		} else if (days === 90) {
			retentionMode = '90days';
		} else if (days === 30) {
			retentionMode = '30days';
		} else {
			retentionMode = 'custom';
			customDays = days;
		}

		// Set exception checkboxes
		const exc = policy.exceptions || [];
		exceptOperator = exc.includes('pubkey:operator');
		exceptProfiles = exc.includes('kind:0');
		exceptFollows = exc.includes('kind:3');
		exceptDMs = exc.includes('kind:4') || exc.includes('kind:14');
	}

	function getRetentionDays() {
		switch (retentionMode) {
			case 'never': return 0;
			case '1year': return 365;
			case '6months': return 180;
			case '90days': return 90;
			case '30days': return 30;
			case 'custom': return parseInt(customDays) || 90;
			default: return 0;
		}
	}

	function buildExceptions() {
		const exceptions = [];
		if (exceptOperator) exceptions.push('pubkey:operator');
		if (exceptProfiles) exceptions.push('kind:0');
		if (exceptFollows) exceptions.push('kind:3');
		if (exceptDMs) {
			exceptions.push('kind:4');
			exceptions.push('kind:14');
		}
		return exceptions;
	}

	async function saveRetentionPolicy() {
		savingRetention = true;
		try {
			await storage.updateRetention({
				retention_days: getRetentionDays(),
				exceptions: buildExceptions(),
				honor_nip09: retentionPolicy.honor_nip09
			});
			notify('success', 'Retention policy saved');
		} catch (e) {
			notify('error', e.message);
		} finally {
			savingRetention = false;
		}
	}

	async function runRetentionNow() {
		runningRetention = true;
		try {
			const result = await storage.runRetentionNow();
			if (result.disabled) {
				notify('info', 'Retention policy is set to "keep forever" - no events deleted');
			} else {
				notify('success', `Retention completed: deleted ${result.events_deleted} events`);
			}
			// Reload storage status and retention policy
			const [statusRes, retentionRes] = await Promise.all([
				storage.getStatus(),
				storage.getRetention()
			]);
			storageStatus = statusRes;
			retentionPolicy = retentionRes;
		} catch (e) {
			notify('error', e.message);
		} finally {
			runningRetention = false;
		}
	}

	async function loadCleanupEstimate() {
		if (!cleanupDate) return;

		estimateLoading = true;
		try {
			const isoDate = new Date(cleanupDate).toISOString();
			cleanupEstimate = await storage.getEstimate(isoDate, applyCleanupExceptions);
		} catch {
			cleanupEstimate = null;
		} finally {
			estimateLoading = false;
		}
	}

	async function runCleanup() {
		if (!cleanupDate) return;

		runningCleanup = true;
		try {
			const isoDate = new Date(cleanupDate).toISOString();
			const result = await storage.cleanup({
				before_date: isoDate,
				apply_exceptions: applyCleanupExceptions
			});
			notify('success', `Deleted ${result.deleted_count} events. Run VACUUM to reclaim space.`);
			showCleanupConfirm = false;
			cleanupDate = '';
			cleanupEstimate = null;
			// Reload storage status
			storageStatus = await storage.getStatus();
		} catch (e) {
			notify('error', e.message);
		} finally {
			runningCleanup = false;
		}
	}

	async function runVacuum() {
		runningVacuum = true;
		try {
			const result = await storage.vacuum();
			notify('success', `VACUUM completed. Reclaimed ${formatBytes(result.space_reclaimed)}.`);
			// Reload storage status
			storageStatus = await storage.getStatus();
		} catch (e) {
			notify('error', e.message);
		} finally {
			runningVacuum = false;
		}
	}

	async function runIntegrityCheck() {
		runningIntegrity = true;
		integrityResult = null;
		try {
			const result = await storage.integrityCheck();
			integrityResult = result;
			if (result.app_db?.ok && result.relay_db?.ok) {
				notify('success', 'Integrity check passed');
			} else {
				notify('error', 'Integrity check found issues');
			}
		} catch (e) {
			notify('error', e.message);
		} finally {
			runningIntegrity = false;
		}
	}

	// Load data on mount (using $effect for Svelte 5 compatibility)
	let initialized = $state(false);
	$effect(() => {
		if (browser && !initialized) {
			initialized = true;
			loadData();
		}
	});

	// Watch cleanup date and exceptions for estimate
	$effect(() => {
		// Track both cleanupDate and applyCleanupExceptions for reactivity
		void applyCleanupExceptions;
		if (cleanupDate) {
			loadCleanupEstimate();
		}
	});
</script>

<div class="space-y-6">
	<!-- Page Header -->
	<div>
		<h1 class="text-2xl font-bold text-gray-900 dark:text-gray-100">Storage Management</h1>
		<p class="text-gray-600 dark:text-gray-400">Monitor and manage relay storage</p>
	</div>

	{#if loading}
		<div class="flex items-center justify-center py-12">
			<Loading text="Loading storage data..." />
		</div>
	{:else if error}
		<Error title="Error loading storage data" message={error} onRetry={loadData} />
	{:else}
		<!-- Current Usage -->
		<div class="rounded-lg bg-white dark:bg-gray-800 p-6 shadow dark:shadow-gray-900/50">
			<h2 class="mb-4 text-lg font-semibold text-gray-900 dark:text-gray-100">Current Usage</h2>

			<StorageProgressBar
				usedBytes={storageStatus?.total_size ?? 0}
				totalBytes={storageStatus?.available_space ?? 0}
				status={storageStatus?.status ?? 'healthy'}
				size="lg"
			/>

			<div class="mt-6 grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
				<div class="rounded-lg bg-gray-50 dark:bg-gray-700 p-4">
					<div class="text-sm text-gray-500 dark:text-gray-400">Relay Database</div>
					<div class="text-lg font-semibold text-gray-900 dark:text-gray-100">
						{formatBytes(storageStatus?.database_size ?? 0)}
					</div>
				</div>
				<div class="rounded-lg bg-gray-50 dark:bg-gray-700 p-4">
					<div class="text-sm text-gray-500 dark:text-gray-400">App Database</div>
					<div class="text-lg font-semibold text-gray-900 dark:text-gray-100">
						{formatBytes(storageStatus?.app_database_size ?? 0)}
					</div>
				</div>
				<div class="rounded-lg bg-gray-50 dark:bg-gray-700 p-4">
					<div class="text-sm text-gray-500 dark:text-gray-400">Available Space</div>
					<div class="text-lg font-semibold text-gray-900 dark:text-gray-100">
						{formatBytes(storageStatus?.available_space ?? 0)}
					</div>
				</div>
				<div class="rounded-lg bg-gray-50 dark:bg-gray-700 p-4">
					<div class="text-sm text-gray-500 dark:text-gray-400">Total Events</div>
					<div class="text-lg font-semibold text-gray-900 dark:text-gray-100">
						{storageStatus?.total_events?.toLocaleString() ?? 0}
					</div>
				</div>
			</div>

			{#if storageStatus?.oldest_event && storageStatus?.newest_event}
				<div class="mt-4 text-sm text-gray-500 dark:text-gray-400">
					Events span from {formatDateInTimezone(storageStatus.oldest_event, { month: 'short', day: 'numeric', year: 'numeric' })} to {formatDateInTimezone(storageStatus.newest_event, { month: 'short', day: 'numeric', year: 'numeric' })}
				</div>
			{/if}
		</div>

		<!-- Retention Policy -->
		<div class="rounded-lg bg-white dark:bg-gray-800 p-6 shadow dark:shadow-gray-900/50">
			<h2 class="mb-4 text-lg font-semibold text-gray-900 dark:text-gray-100">Retention Policy</h2>

			<div class="space-y-4">
				<fieldset>
					<legend class="mb-2 block text-sm font-medium text-gray-700 dark:text-gray-200">
						Auto-delete events older than:
					</legend>
					<div class="space-y-2">
						<label class="flex cursor-pointer items-center gap-2">
							<input
								type="radio"
								name="retention"
								value="never"
								bind:group={retentionMode}
								class="h-4 w-4 text-purple-600"
							/>
							<span class="text-sm text-gray-700 dark:text-gray-200">Never (keep forever)</span>
						</label>
						<label class="flex cursor-pointer items-center gap-2">
							<input
								type="radio"
								name="retention"
								value="1year"
								bind:group={retentionMode}
								class="h-4 w-4 text-purple-600"
							/>
							<span class="text-sm text-gray-700 dark:text-gray-200">1 year</span>
						</label>
						<label class="flex cursor-pointer items-center gap-2">
							<input
								type="radio"
								name="retention"
								value="6months"
								bind:group={retentionMode}
								class="h-4 w-4 text-purple-600"
							/>
							<span class="text-sm text-gray-700 dark:text-gray-200">6 months</span>
						</label>
						<label class="flex cursor-pointer items-center gap-2">
							<input
								type="radio"
								name="retention"
								value="90days"
								bind:group={retentionMode}
								class="h-4 w-4 text-purple-600"
							/>
							<span class="text-sm text-gray-700 dark:text-gray-200">90 days</span>
						</label>
						<label class="flex cursor-pointer items-center gap-2">
							<input
								type="radio"
								name="retention"
								value="30days"
								bind:group={retentionMode}
								class="h-4 w-4 text-purple-600"
							/>
							<span class="text-sm text-gray-700 dark:text-gray-200">30 days</span>
						</label>
						<label class="flex cursor-pointer items-center gap-2">
							<input
								type="radio"
								name="retention"
								value="custom"
								bind:group={retentionMode}
								class="h-4 w-4 text-purple-600"
							/>
							<span class="text-sm text-gray-700 dark:text-gray-200">Custom:</span>
							<input
								type="number"
								min="1"
								bind:value={customDays}
								disabled={retentionMode !== 'custom'}
								class="w-20 rounded-lg border border-gray-300 dark:border-gray-600 px-2 py-1 text-sm bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100 disabled:bg-gray-100 dark:disabled:bg-gray-600 disabled:text-gray-400"
							/>
							<span class="text-sm text-gray-700 dark:text-gray-200">days</span>
						</label>
					</div>
				</fieldset>

				<fieldset>
					<legend class="mb-2 block text-sm font-medium text-gray-700 dark:text-gray-200">
						Exceptions (never auto-delete):
					</legend>
					<div class="space-y-2">
						<label class="flex cursor-pointer items-start gap-3">
							<input
								type="checkbox"
								bind:checked={exceptOperator}
								class="mt-0.5 h-4 w-4 rounded border-gray-300 dark:border-gray-600 text-purple-600"
							/>
							<div>
								<span class="text-sm font-medium text-gray-700 dark:text-gray-200">My events (operator pubkey)</span>
								<p class="text-xs text-gray-500 dark:text-gray-400">Keep all events from the relay operator</p>
							</div>
						</label>
						<label class="flex cursor-pointer items-start gap-3">
							<input
								type="checkbox"
								bind:checked={exceptProfiles}
								class="mt-0.5 h-4 w-4 rounded border-gray-300 dark:border-gray-600 text-purple-600"
							/>
							<div>
								<span class="text-sm font-medium text-gray-700 dark:text-gray-200">Profile metadata (kind 0)</span>
								<p class="text-xs text-gray-500 dark:text-gray-400">Keep user profile information</p>
							</div>
						</label>
						<label class="flex cursor-pointer items-start gap-3">
							<input
								type="checkbox"
								bind:checked={exceptFollows}
								class="mt-0.5 h-4 w-4 rounded border-gray-300 dark:border-gray-600 text-purple-600"
							/>
							<div>
								<span class="text-sm font-medium text-gray-700 dark:text-gray-200">Follow lists (kind 3)</span>
								<p class="text-xs text-gray-500 dark:text-gray-400">Keep contact/follow lists</p>
							</div>
						</label>
						<label class="flex cursor-pointer items-start gap-3">
							<input
								type="checkbox"
								bind:checked={exceptDMs}
								class="mt-0.5 h-4 w-4 rounded border-gray-300 dark:border-gray-600 text-purple-600"
							/>
							<div>
								<span class="text-sm font-medium text-gray-700 dark:text-gray-200">DMs (kind 4, 14)</span>
								<p class="text-xs text-gray-500 dark:text-gray-400">Keep encrypted direct messages</p>
							</div>
						</label>
					</div>
				</fieldset>

				<div class="flex items-center gap-2 text-xs text-gray-500 dark:text-gray-400 border-t border-gray-200 dark:border-gray-700 pt-4">
					<svg class="h-4 w-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
					</svg>
					Retention policy runs daily at midnight
					{#if retentionPolicy.last_run}
						<span class="text-gray-400 dark:text-gray-500">|</span>
						Last run: {formatRelativeTime(retentionPolicy.last_run)}
					{/if}
				</div>

				<div class="flex justify-end gap-3">
					<Button variant="secondary" onclick={runRetentionNow} loading={runningRetention}>
						Run Now
					</Button>
					<Button variant="primary" onclick={saveRetentionPolicy} loading={savingRetention}>
						Save Retention Policy
					</Button>
				</div>
			</div>
		</div>

		<!-- Manual Cleanup -->
		<div class="rounded-lg bg-white dark:bg-gray-800 p-6 shadow dark:shadow-gray-900/50">
			<h2 class="mb-4 text-lg font-semibold text-gray-900 dark:text-gray-100">Manual Cleanup</h2>

			<div class="space-y-4">
				<div>
					<label for="cleanup-date" class="mb-1 block text-sm font-medium text-gray-700 dark:text-gray-200">
						Delete all events before:
					</label>
					<input
						id="cleanup-date"
						type="date"
						bind:value={cleanupDate}
						max={new Date().toISOString().split('T')[0]}
						class="rounded-lg border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100 px-4 py-2 focus:border-purple-500 focus:outline-none"
					/>
				</div>

				<label class="flex cursor-pointer items-start gap-3">
					<input
						type="checkbox"
						bind:checked={applyCleanupExceptions}
						class="mt-0.5 h-4 w-4 rounded border-gray-300 dark:border-gray-600 text-purple-600"
					/>
					<div>
						<span class="text-sm font-medium text-gray-700 dark:text-gray-200">Apply retention exceptions</span>
						<p class="text-xs text-gray-500 dark:text-gray-400">
							Protect event types selected in Retention Policy above
							{#if applyCleanupExceptions && retentionPolicy.exceptions?.length > 0}
								<span class="text-purple-600 dark:text-purple-400">
									(protecting: {retentionPolicy.exceptions.map(e => e.replace('kind:', 'kind ').replace('pubkey:operator', 'operator')).join(', ')})
								</span>
							{/if}
						</p>
					</div>
				</label>

				{#if cleanupEstimate}
					<div class="rounded-lg bg-gray-50 dark:bg-gray-700 p-4">
						<div class="text-sm text-gray-600 dark:text-gray-300">
							<strong>{cleanupEstimate.event_count?.toLocaleString()}</strong> events will be deleted
						</div>
						<div class="text-sm text-gray-600 dark:text-gray-300">
							Estimated space freed: <strong>{formatBytes(cleanupEstimate.estimated_space ?? 0)}</strong>
						</div>
					</div>
				{:else if estimateLoading}
					<div class="flex items-center gap-2 text-sm text-gray-500 dark:text-gray-400">
						<Loading size="sm" />
						Calculating estimate...
					</div>
				{/if}

				<div class="flex items-start gap-2 rounded-lg bg-yellow-50 dark:bg-yellow-900/20 p-3 text-sm text-yellow-800 dark:text-yellow-200">
					<svg class="mt-0.5 h-4 w-4 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
					</svg>
					<span>This cannot be undone. Consider exporting events first.</span>
				</div>

				<div class="flex justify-end">
					<Button
						variant="danger"
						onclick={() => showCleanupConfirm = true}
						disabled={!cleanupDate || !cleanupEstimate || cleanupEstimate.event_count === 0}
					>
						Delete Old Events
					</Button>
				</div>
			</div>
		</div>

		<!-- Database Maintenance -->
		<div class="rounded-lg bg-white dark:bg-gray-800 p-6 shadow dark:shadow-gray-900/50">
			<h2 class="mb-4 text-lg font-semibold text-gray-900 dark:text-gray-100">Database Maintenance</h2>

			<div class="space-y-4">
				<div class="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
					<div>
						<h3 class="font-medium text-gray-900 dark:text-gray-100">VACUUM Database</h3>
						<p class="text-sm text-gray-500 dark:text-gray-400">
							Reclaims disk space after deletions. May take a few minutes for large databases.
						</p>
					</div>
					<Button variant="secondary" onclick={runVacuum} loading={runningVacuum}>
						Run VACUUM
					</Button>
				</div>

				<div class="border-t border-gray-200 dark:border-gray-700 pt-4">
					<div class="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
						<div>
							<h3 class="font-medium text-gray-900 dark:text-gray-100">Integrity Check</h3>
							<p class="text-sm text-gray-500 dark:text-gray-400">
								Verify database integrity and check for corruption.
							</p>
						</div>
						<Button variant="secondary" onclick={runIntegrityCheck} loading={runningIntegrity}>
							Check Integrity
						</Button>
					</div>

					{#if integrityResult}
						<div class="mt-3 rounded-lg p-3 {integrityResult.app_db?.ok && integrityResult.relay_db?.ok ? 'bg-green-50 dark:bg-green-900/20' : 'bg-red-50 dark:bg-red-900/20'}">
							<div class="flex items-center gap-2 text-sm {integrityResult.app_db?.ok && integrityResult.relay_db?.ok ? 'text-green-800 dark:text-green-200' : 'text-red-800 dark:text-red-200'}">
								{#if integrityResult.app_db?.ok && integrityResult.relay_db?.ok}
									<svg class="h-4 w-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
										<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
									</svg>
									All databases passed integrity check
								{:else}
									<svg class="h-4 w-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
										<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
									</svg>
									Integrity issues found
								{/if}
							</div>
							{#if !integrityResult.app_db?.ok}
								<div class="mt-1 text-xs text-red-600 dark:text-red-400">
									App DB: {integrityResult.app_db?.result}
								</div>
							{/if}
							{#if !integrityResult.relay_db?.ok}
								<div class="mt-1 text-xs text-red-600 dark:text-red-400">
									Relay DB: {integrityResult.relay_db?.result}
								</div>
							{/if}
						</div>
					{/if}
				</div>
			</div>
		</div>
	{/if}
</div>

<!-- Cleanup Confirmation Modal -->
{#if showCleanupConfirm}
	<div
		class="fixed inset-0 z-50 flex items-center justify-center bg-black/50"
		onclick={() => showCleanupConfirm = false}
		onkeydown={(e) => e.key === 'Escape' && (showCleanupConfirm = false)}
		role="dialog"
		aria-modal="true"
		tabindex="-1"
	>
		<!-- svelte-ignore a11y_click_events_have_key_events -->
		<!-- svelte-ignore a11y_no_static_element_interactions -->
		<div
			class="mx-4 w-full max-w-md rounded-lg bg-white dark:bg-gray-800 p-6 shadow-xl"
			onclick={(e) => e.stopPropagation()}
		>
			<h3 class="text-lg font-semibold text-gray-900 dark:text-gray-100">Confirm Cleanup</h3>
			<p class="mt-2 text-sm text-gray-600 dark:text-gray-300">
				You are about to delete <strong>{cleanupEstimate?.event_count?.toLocaleString()}</strong> events
				created before <strong>{formatDateInTimezone(cleanupDate, { month: 'short', day: 'numeric', year: 'numeric' })}</strong>.
			</p>
			<p class="mt-2 text-sm text-gray-600 dark:text-gray-300">
				This will free approximately <strong>{formatBytes(cleanupEstimate?.estimated_space ?? 0)}</strong>
				of storage space.
			</p>
			<div class="mt-4 rounded-lg bg-red-50 dark:bg-red-900/20 p-3 text-sm text-red-800 dark:text-red-200">
				This action cannot be undone. Make sure you have exported any important data.
			</div>
			<div class="mt-6 flex justify-end gap-3">
				<Button variant="secondary" onclick={() => showCleanupConfirm = false}>
					Cancel
				</Button>
				<Button variant="danger" onclick={runCleanup} loading={runningCleanup}>
					Delete Events
				</Button>
			</div>
		</div>
	</div>
{/if}

