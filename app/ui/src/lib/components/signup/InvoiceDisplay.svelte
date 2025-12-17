<script>
	import { browser } from '$app/environment';
	import { onDestroy } from 'svelte';
	import { signup } from '$lib/api/client.js';
	import QRCode from 'qrcode';

	let { invoice, tier, onPaymentConfirmed = () => {} } = $props();

	let qrCodeUrl = $state('');
	let copied = $state(false);
	let timeRemaining = $state(0);
	let pollInterval = $state(null);
	let countdownInterval = $state(null);
	let checking = $state(false);
	let webLNAvailable = $state(false);
	let payingWithWebLN = $state(false);

	// Generate QR code
	$effect(() => {
		if (browser && invoice?.payment_request) {
			QRCode.toDataURL(invoice.payment_request.toUpperCase(), {
				width: 256,
				margin: 2,
				color: {
					dark: '#000000',
					light: '#ffffff'
				}
			}).then((url) => {
				qrCodeUrl = url;
			});
		}
	});

	// Check for WebLN
	$effect(() => {
		if (browser) {
			webLNAvailable = typeof window.webln !== 'undefined';
		}
	});

	// Calculate time remaining
	$effect(() => {
		if (browser && invoice?.expires_at) {
			const updateCountdown = () => {
				const now = new Date();
				const expiry = new Date(invoice.expires_at);
				const diff = Math.max(0, Math.floor((expiry - now) / 1000));
				timeRemaining = diff;

				if (diff <= 0 && countdownInterval) {
					clearInterval(countdownInterval);
				}
			};

			updateCountdown();
			countdownInterval = setInterval(updateCountdown, 1000);
		}
	});

	// Poll for payment status
	$effect(() => {
		if (browser && invoice?.payment_hash) {
			const checkPayment = async () => {
				if (checking) return;
				checking = true;

				try {
					const status = await signup.checkInvoice(invoice.payment_hash);
					if (status.status === 'paid') {
						cleanup();
						onPaymentConfirmed();
					}
				} catch (e) {
					console.error('Failed to check invoice status:', e);
				} finally {
					checking = false;
				}
			};

			// Check immediately
			checkPayment();
			// Then poll every 3 seconds
			pollInterval = setInterval(checkPayment, 3000);
		}
	});

	function cleanup() {
		if (pollInterval) {
			clearInterval(pollInterval);
			pollInterval = null;
		}
		if (countdownInterval) {
			clearInterval(countdownInterval);
			countdownInterval = null;
		}
	}

	onDestroy(cleanup);

	async function copyInvoice() {
		try {
			await navigator.clipboard.writeText(invoice.payment_request);
			copied = true;
			setTimeout(() => (copied = false), 2000);
		} catch (e) {
			console.error('Failed to copy:', e);
		}
	}

	async function payWithWebLN() {
		if (!window.webln) return;

		payingWithWebLN = true;
		try {
			await window.webln.enable();
			await window.webln.sendPayment(invoice.payment_request);
			// Payment successful - the polling will catch it
		} catch (e) {
			console.error('WebLN payment failed:', e);
			// User cancelled or error - they can still scan QR
		} finally {
			payingWithWebLN = false;
		}
	}

	function formatTime(seconds) {
		const mins = Math.floor(seconds / 60);
		const secs = seconds % 60;
		return `${mins}:${secs.toString().padStart(2, '0')}`;
	}

	function truncateInvoice(inv) {
		if (!inv) return '';
		if (inv.length <= 40) return inv;
		return inv.slice(0, 20) + '...' + inv.slice(-20);
	}
</script>

<div class="space-y-6">
	<!-- Amount Summary -->
	<div class="text-center p-4 bg-purple-50 dark:bg-purple-900/20 rounded-lg">
		<p class="text-sm text-purple-600 dark:text-purple-400 mb-1">{tier?.name || 'Relay Access'}</p>
		<p class="text-2xl font-bold text-purple-900 dark:text-purple-100">{invoice?.amount_sats?.toLocaleString()} sats</p>
	</div>

	<!-- QR Code -->
	<div class="flex justify-center">
		{#if qrCodeUrl}
			<div class="p-4 bg-white rounded-xl border-2 border-gray-200 dark:border-gray-600 shadow-sm">
				<img src={qrCodeUrl} alt="Lightning Invoice QR Code" class="w-56 h-56" />
			</div>
		{:else}
			<div class="w-64 h-64 bg-gray-100 dark:bg-gray-700 rounded-xl animate-pulse flex items-center justify-center">
				<div class="h-8 w-8 animate-spin rounded-full border-4 border-purple-600 border-t-transparent"></div>
			</div>
		{/if}
	</div>

	<!-- Invoice String -->
	<div class="space-y-2">
		<div class="flex items-center justify-between p-3 bg-gray-50 dark:bg-gray-700 rounded-lg border border-gray-200 dark:border-gray-600">
			<code class="text-xs text-gray-600 dark:text-gray-300 font-mono truncate mr-2">
				{truncateInvoice(invoice?.payment_request)}
			</code>
			<button
				type="button"
				onclick={copyInvoice}
				class="flex-shrink-0 px-3 py-1.5 text-sm font-medium rounded-lg transition-colors {copied ? 'bg-green-100 dark:bg-green-900/30 text-green-700 dark:text-green-300' : 'bg-gray-200 dark:bg-gray-600 hover:bg-gray-300 dark:hover:bg-gray-500 text-gray-700 dark:text-gray-200'}"
			>
				{copied ? 'Copied!' : 'Copy'}
			</button>
		</div>
	</div>

	<!-- WebLN Button -->
	{#if webLNAvailable}
		<button
			type="button"
			onclick={payWithWebLN}
			disabled={payingWithWebLN}
			class="w-full py-3 px-4 bg-amber-500 hover:bg-amber-600 text-white font-medium rounded-lg transition-colors flex items-center justify-center disabled:opacity-50"
		>
			{#if payingWithWebLN}
				<div class="w-5 h-5 mr-2 animate-spin rounded-full border-2 border-white border-t-transparent"></div>
				Processing...
			{:else}
				<svg class="w-5 h-5 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 10V3L4 14h7v7l9-11h-7z" />
				</svg>
				Pay with WebLN
			{/if}
		</button>
	{/if}

	<!-- Timer and Status -->
	<div class="text-center space-y-2">
		{#if timeRemaining > 0}
			<div class="flex items-center justify-center text-gray-600 dark:text-gray-300">
				<svg class="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
				</svg>
				<span class="text-sm">
					Invoice expires in <span class="font-mono font-medium">{formatTime(timeRemaining)}</span>
				</span>
			</div>
		{:else}
			<div class="text-amber-600 dark:text-amber-400 font-medium">Invoice expired</div>
		{/if}

		<div class="flex items-center justify-center text-sm text-gray-500 dark:text-gray-400">
			<div class="w-2 h-2 mr-2 bg-green-500 rounded-full animate-pulse"></div>
			Waiting for payment...
		</div>
	</div>
</div>
