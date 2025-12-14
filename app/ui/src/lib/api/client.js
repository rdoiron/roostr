/**
 * API client module for Roostr.
 * All API calls go through this module for consistent error handling.
 */

const API_BASE = '/api/v1';

/**
 * Custom API error with code.
 */
export class ApiError extends Error {
	constructor(message, code, status) {
		super(message);
		this.name = 'ApiError';
		this.code = code;
		this.status = status;
	}
}

/**
 * Parse error response.
 */
async function parseError(res) {
	try {
		const data = await res.json();
		return new ApiError(data.error || `HTTP ${res.status}`, data.code, res.status);
	} catch {
		return new ApiError(`HTTP ${res.status}`, 'UNKNOWN', res.status);
	}
}

/**
 * Make a GET request to the API.
 * @param {string} path - API path (without base)
 * @returns {Promise<any>} Response data
 */
export async function get(path) {
	const res = await fetch(`${API_BASE}${path}`);
	if (!res.ok) throw await parseError(res);
	return res.json();
}

/**
 * Make a POST request to the API.
 * @param {string} path - API path (without base)
 * @param {any} data - Request body
 * @returns {Promise<any>} Response data
 */
export async function post(path, data) {
	const res = await fetch(`${API_BASE}${path}`, {
		method: 'POST',
		headers: { 'Content-Type': 'application/json' },
		body: JSON.stringify(data)
	});
	if (!res.ok) throw await parseError(res);
	return res.json();
}

/**
 * Make a PUT request to the API.
 * @param {string} path - API path (without base)
 * @param {any} data - Request body
 * @returns {Promise<any>} Response data
 */
export async function put(path, data) {
	const res = await fetch(`${API_BASE}${path}`, {
		method: 'PUT',
		headers: { 'Content-Type': 'application/json' },
		body: JSON.stringify(data)
	});
	if (!res.ok) throw await parseError(res);
	return res.json();
}

/**
 * Make a PATCH request to the API.
 * @param {string} path - API path (without base)
 * @param {any} data - Request body
 * @returns {Promise<any>} Response data
 */
export async function patch(path, data) {
	const res = await fetch(`${API_BASE}${path}`, {
		method: 'PATCH',
		headers: { 'Content-Type': 'application/json' },
		body: JSON.stringify(data)
	});
	if (!res.ok) throw await parseError(res);
	return res.json();
}

/**
 * Make a DELETE request to the API.
 * @param {string} path - API path (without base)
 * @returns {Promise<any>} Response data
 */
export async function del(path) {
	const res = await fetch(`${API_BASE}${path}`, { method: 'DELETE' });
	if (!res.ok) throw await parseError(res);
	return res.json();
}

// API function groups for better organization
export const setup = {
	getStatus: () => get('/setup/status'),
	complete: (data) => post('/setup/complete', data),
	validateIdentity: (input) => get(`/setup/validate-identity?input=${encodeURIComponent(input)}`)
};

export const access = {
	getMode: () => get('/access/mode'),
	setMode: (mode) => put('/access/mode', { mode }),
	getWhitelist: () => get('/access/whitelist'),
	addToWhitelist: (data) => post('/access/whitelist', data),
	removeFromWhitelist: (pubkey) => del(`/access/whitelist/${pubkey}`),
	updateWhitelist: (pubkey, data) => patch(`/access/whitelist/${pubkey}`, data),
	getBlacklist: () => get('/access/blacklist'),
	addToBlacklist: (data) => post('/access/blacklist', data),
	removeFromBlacklist: (pubkey) => del(`/access/blacklist/${pubkey}`),
	resolveNip05: (identifier) => get(`/nip05/${encodeURIComponent(identifier)}`)
};

export const stats = {
	getSummary: () => get('/stats/summary')
};

export const events = {
	list: (params = {}) => {
		const query = new URLSearchParams(params).toString();
		return get(`/events${query ? '?' + query : ''}`);
	},
	get: (id) => get(`/events/${id}`),
	delete: (id) => del(`/events/${id}`),
	getRecent: () => get('/events/recent')
};

export const relay = {
	getStatus: () => get('/relay/status'),
	getURLs: () => get('/relay/urls')
};
