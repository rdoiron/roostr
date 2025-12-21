import { describe, it, expect, vi, beforeEach } from 'vitest';
import {
	ApiError,
	get,
	post,
	put,
	patch,
	del,
	setup,
	access,
	stats,
	events,
	relay,
	config,
	storage,
	sync,
	pricing,
	lightning,
	paidUsers
} from './client.js';

describe('ApiError', () => {
	it('creates error with message, code, and status', () => {
		const error = new ApiError('Not found', 'NOT_FOUND', 404);
		expect(error.message).toBe('Not found');
		expect(error.code).toBe('NOT_FOUND');
		expect(error.status).toBe(404);
		expect(error.name).toBe('ApiError');
	});

	it('is instanceof Error', () => {
		const error = new ApiError('Test', 'TEST', 500);
		expect(error instanceof Error).toBe(true);
	});
});

describe('HTTP methods', () => {
	beforeEach(() => {
		vi.mocked(fetch).mockReset();
	});

	describe('get', () => {
		it('makes GET request and returns JSON', async () => {
			const mockData = { id: 1, name: 'test' };
			vi.mocked(fetch).mockResolvedValue({
				ok: true,
				json: () => Promise.resolve(mockData)
			});

			const result = await get('/test');

			expect(fetch).toHaveBeenCalledWith('/api/v1/test');
			expect(result).toEqual(mockData);
		});

		it('throws ApiError on non-ok response', async () => {
			vi.mocked(fetch).mockResolvedValue({
				ok: false,
				status: 404,
				json: () => Promise.resolve({ error: 'Not found', code: 'NOT_FOUND' })
			});

			await expect(get('/missing')).rejects.toThrow(ApiError);
			await expect(get('/missing')).rejects.toMatchObject({
				message: 'Not found',
				code: 'NOT_FOUND',
				status: 404
			});
		});

		it('handles JSON parse error in error response', async () => {
			vi.mocked(fetch).mockResolvedValue({
				ok: false,
				status: 500,
				json: () => Promise.reject(new Error('Invalid JSON'))
			});

			await expect(get('/error')).rejects.toMatchObject({
				message: 'HTTP 500',
				code: 'UNKNOWN',
				status: 500
			});
		});
	});

	describe('post', () => {
		it('makes POST request with JSON body', async () => {
			const requestData = { name: 'test' };
			const responseData = { id: 1, name: 'test' };
			vi.mocked(fetch).mockResolvedValue({
				ok: true,
				json: () => Promise.resolve(responseData)
			});

			const result = await post('/items', requestData);

			expect(fetch).toHaveBeenCalledWith('/api/v1/items', {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify(requestData)
			});
			expect(result).toEqual(responseData);
		});

		it('throws ApiError on failure', async () => {
			vi.mocked(fetch).mockResolvedValue({
				ok: false,
				status: 400,
				json: () => Promise.resolve({ error: 'Invalid data', code: 'INVALID' })
			});

			await expect(post('/items', {})).rejects.toThrow(ApiError);
		});
	});

	describe('put', () => {
		it('makes PUT request with JSON body', async () => {
			const requestData = { name: 'updated' };
			vi.mocked(fetch).mockResolvedValue({
				ok: true,
				json: () => Promise.resolve(requestData)
			});

			await put('/items/1', requestData);

			expect(fetch).toHaveBeenCalledWith('/api/v1/items/1', {
				method: 'PUT',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify(requestData)
			});
		});
	});

	describe('patch', () => {
		it('makes PATCH request with JSON body', async () => {
			const requestData = { name: 'patched' };
			vi.mocked(fetch).mockResolvedValue({
				ok: true,
				json: () => Promise.resolve(requestData)
			});

			await patch('/items/1', requestData);

			expect(fetch).toHaveBeenCalledWith('/api/v1/items/1', {
				method: 'PATCH',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify(requestData)
			});
		});
	});

	describe('del', () => {
		it('makes DELETE request', async () => {
			vi.mocked(fetch).mockResolvedValue({
				ok: true,
				json: () => Promise.resolve({ success: true })
			});

			await del('/items/1');

			expect(fetch).toHaveBeenCalledWith('/api/v1/items/1', { method: 'DELETE' });
		});
	});
});

describe('API function groups', () => {
	beforeEach(() => {
		vi.mocked(fetch).mockReset();
		vi.mocked(fetch).mockResolvedValue({
			ok: true,
			json: () => Promise.resolve({})
		});
	});

	describe('setup', () => {
		it('getStatus calls correct endpoint', async () => {
			await setup.getStatus();
			expect(fetch).toHaveBeenCalledWith('/api/v1/setup/status');
		});

		it('complete calls POST with data', async () => {
			const data = { pubkey: 'abc123' };
			await setup.complete(data);
			expect(fetch).toHaveBeenCalledWith('/api/v1/setup/complete', {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify(data)
			});
		});

		it('validateIdentity encodes input', async () => {
			await setup.validateIdentity('user@example.com');
			expect(fetch).toHaveBeenCalledWith('/api/v1/setup/validate-identity?input=user%40example.com');
		});
	});

	describe('access', () => {
		it('getMode calls correct endpoint', async () => {
			await access.getMode();
			expect(fetch).toHaveBeenCalledWith('/api/v1/access/mode');
		});

		it('setMode calls PUT with mode', async () => {
			await access.setMode('private');
			expect(fetch).toHaveBeenCalledWith('/api/v1/access/mode', {
				method: 'PUT',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ mode: 'private' })
			});
		});

		it('getWhitelist calls correct endpoint', async () => {
			await access.getWhitelist();
			expect(fetch).toHaveBeenCalledWith('/api/v1/access/whitelist');
		});

		it('addToWhitelist posts data', async () => {
			const data = { pubkey: 'abc', nickname: 'test' };
			await access.addToWhitelist(data);
			expect(fetch).toHaveBeenCalledWith('/api/v1/access/whitelist', {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify(data)
			});
		});

		it('removeFromWhitelist deletes by pubkey', async () => {
			await access.removeFromWhitelist('abc123');
			expect(fetch).toHaveBeenCalledWith('/api/v1/access/whitelist/abc123', { method: 'DELETE' });
		});

		it('resolveNip05 encodes identifier', async () => {
			await access.resolveNip05('user@example.com');
			expect(fetch).toHaveBeenCalledWith('/api/v1/nip05/user%40example.com');
		});
	});

	describe('stats', () => {
		it('getSummary calls correct endpoint', async () => {
			await stats.getSummary();
			expect(fetch).toHaveBeenCalledWith('/api/v1/stats/summary');
		});

		it('getEventsOverTime uses default time range', async () => {
			await stats.getEventsOverTime();
			expect(fetch).toHaveBeenCalledWith('/api/v1/stats/events-over-time?time_range=7days');
		});

		it('getEventsOverTime accepts custom params', async () => {
			await stats.getEventsOverTime('30days', 'America/New_York');
			expect(fetch).toHaveBeenCalledWith(
				'/api/v1/stats/events-over-time?time_range=30days&timezone=America%2FNew_York'
			);
		});

		it('getEventsByKind uses default time range', async () => {
			await stats.getEventsByKind();
			expect(fetch).toHaveBeenCalledWith('/api/v1/stats/events-by-kind?time_range=alltime');
		});

		it('getTopAuthors uses defaults', async () => {
			await stats.getTopAuthors();
			expect(fetch).toHaveBeenCalledWith('/api/v1/stats/top-authors?time_range=alltime&limit=10');
		});

		it('getTopAuthors accepts custom params', async () => {
			await stats.getTopAuthors('7days', 20);
			expect(fetch).toHaveBeenCalledWith('/api/v1/stats/top-authors?time_range=7days&limit=20');
		});
	});

	describe('events', () => {
		it('list calls correct endpoint', async () => {
			await events.list();
			expect(fetch).toHaveBeenCalledWith('/api/v1/events');
		});

		it('list builds query string from params', async () => {
			await events.list({ kind: '1', limit: '50' });
			expect(fetch).toHaveBeenCalledWith('/api/v1/events?kind=1&limit=50');
		});

		it('get fetches by id', async () => {
			await events.get('abc123');
			expect(fetch).toHaveBeenCalledWith('/api/v1/events/abc123');
		});

		it('delete removes by id', async () => {
			await events.delete('abc123');
			expect(fetch).toHaveBeenCalledWith('/api/v1/events/abc123', { method: 'DELETE' });
		});

		it('getRecent calls correct endpoint', async () => {
			await events.getRecent();
			expect(fetch).toHaveBeenCalledWith('/api/v1/events/recent');
		});
	});

	describe('relay', () => {
		it('getStatus calls correct endpoint', async () => {
			await relay.getStatus();
			expect(fetch).toHaveBeenCalledWith('/api/v1/relay/status');
		});

		it('getURLs calls correct endpoint', async () => {
			await relay.getURLs();
			expect(fetch).toHaveBeenCalledWith('/api/v1/relay/urls');
		});

		it('reload posts to correct endpoint', async () => {
			await relay.reload();
			expect(fetch).toHaveBeenCalledWith('/api/v1/relay/reload', {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: '{}'
			});
		});

		it('restart posts to correct endpoint', async () => {
			await relay.restart();
			expect(fetch).toHaveBeenCalledWith('/api/v1/relay/restart', {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: '{}'
			});
		});

		it('getLogs uses default limit', async () => {
			await relay.getLogs();
			expect(fetch).toHaveBeenCalledWith('/api/v1/relay/logs?limit=100');
		});

		it('getLogs accepts custom limit', async () => {
			await relay.getLogs(50);
			expect(fetch).toHaveBeenCalledWith('/api/v1/relay/logs?limit=50');
		});
	});

	describe('config', () => {
		it('get calls correct endpoint', async () => {
			await config.get();
			expect(fetch).toHaveBeenCalledWith('/api/v1/config');
		});

		it('update patches config', async () => {
			const data = { name: 'My Relay' };
			await config.update(data);
			expect(fetch).toHaveBeenCalledWith('/api/v1/config', {
				method: 'PATCH',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify(data)
			});
		});

		it('reload posts to correct endpoint', async () => {
			await config.reload();
			expect(fetch).toHaveBeenCalledWith('/api/v1/config/reload', {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: '{}'
			});
		});
	});

	describe('storage', () => {
		it('getStatus calls correct endpoint', async () => {
			await storage.getStatus();
			expect(fetch).toHaveBeenCalledWith('/api/v1/storage/status');
		});

		it('getRetention calls correct endpoint', async () => {
			await storage.getRetention();
			expect(fetch).toHaveBeenCalledWith('/api/v1/storage/retention');
		});

		it('updateRetention puts data', async () => {
			const data = { days: 30 };
			await storage.updateRetention(data);
			expect(fetch).toHaveBeenCalledWith('/api/v1/storage/retention', {
				method: 'PUT',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify(data)
			});
		});

		it('cleanup posts data', async () => {
			const data = { before: '2024-01-01' };
			await storage.cleanup(data);
			expect(fetch).toHaveBeenCalledWith('/api/v1/storage/cleanup', {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify(data)
			});
		});

		it('vacuum posts to correct endpoint', async () => {
			await storage.vacuum();
			expect(fetch).toHaveBeenCalledWith('/api/v1/storage/vacuum', {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: '{}'
			});
		});

		it('getDeletionRequests calls with optional status', async () => {
			await storage.getDeletionRequests();
			expect(fetch).toHaveBeenCalledWith('/api/v1/storage/deletion-requests');

			await storage.getDeletionRequests('pending');
			expect(fetch).toHaveBeenCalledWith('/api/v1/storage/deletion-requests?status=pending');
		});

		it('getEstimate encodes date', async () => {
			await storage.getEstimate('2024-01-01');
			expect(fetch).toHaveBeenCalledWith('/api/v1/storage/estimate?before_date=2024-01-01');
		});
	});

	describe('sync', () => {
		it('start posts data', async () => {
			const data = { pubkeys: ['abc'], relays: ['wss://relay.example.com'] };
			await sync.start(data);
			expect(fetch).toHaveBeenCalledWith('/api/v1/sync/start', {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify(data)
			});
		});

		it('getStatus calls with optional id', async () => {
			await sync.getStatus();
			expect(fetch).toHaveBeenCalledWith('/api/v1/sync/status');

			await sync.getStatus('123');
			expect(fetch).toHaveBeenCalledWith('/api/v1/sync/status?id=123');
		});

		it('cancel posts to correct endpoint', async () => {
			await sync.cancel();
			expect(fetch).toHaveBeenCalledWith('/api/v1/sync/cancel', {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: '{}'
			});
		});

		it('getHistory builds query params', async () => {
			await sync.getHistory();
			expect(fetch).toHaveBeenCalledWith('/api/v1/sync/history');

			await sync.getHistory({ limit: 10, offset: 5 });
			expect(fetch).toHaveBeenCalledWith('/api/v1/sync/history?limit=10&offset=5');
		});

		it('getRelays calls correct endpoint', async () => {
			await sync.getRelays();
			expect(fetch).toHaveBeenCalledWith('/api/v1/sync/relays');
		});
	});

	describe('pricing', () => {
		it('get calls correct endpoint', async () => {
			await pricing.get();
			expect(fetch).toHaveBeenCalledWith('/api/v1/access/pricing');
		});

		it('update puts tiers', async () => {
			const tiers = [{ name: 'Basic', sats: 1000 }];
			await pricing.update(tiers);
			expect(fetch).toHaveBeenCalledWith('/api/v1/access/pricing', {
				method: 'PUT',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ tiers })
			});
		});
	});

	describe('lightning', () => {
		it('getStatus calls correct endpoint', async () => {
			await lightning.getStatus();
			expect(fetch).toHaveBeenCalledWith('/api/v1/lightning/status');
		});

		it('updateConfig puts config', async () => {
			const cfg = { host: 'localhost:8080' };
			await lightning.updateConfig(cfg);
			expect(fetch).toHaveBeenCalledWith('/api/v1/lightning/config', {
				method: 'PUT',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify(cfg)
			});
		});

		it('test posts config', async () => {
			const cfg = { host: 'localhost:8080' };
			await lightning.test(cfg);
			expect(fetch).toHaveBeenCalledWith('/api/v1/lightning/test', {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify(cfg)
			});
		});
	});

	describe('paidUsers', () => {
		it('list calls correct endpoint', async () => {
			await paidUsers.list();
			expect(fetch).toHaveBeenCalledWith('/api/v1/access/paid-users');
		});

		it('list builds query from params', async () => {
			await paidUsers.list({ status: 'active' });
			expect(fetch).toHaveBeenCalledWith('/api/v1/access/paid-users?status=active');
		});

		it('revoke deletes by pubkey', async () => {
			await paidUsers.revoke('abc123');
			expect(fetch).toHaveBeenCalledWith('/api/v1/access/paid-users/abc123', { method: 'DELETE' });
		});

		it('getRevenue calls correct endpoint', async () => {
			await paidUsers.getRevenue();
			expect(fetch).toHaveBeenCalledWith('/api/v1/access/revenue');
		});
	});
});
