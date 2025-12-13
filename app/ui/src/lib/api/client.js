/**
 * API client module for Roostr.
 * All API calls go through this module for consistent error handling.
 */

const API_BASE = '/api/v1';

/**
 * Make a GET request to the API.
 * @param {string} path - API path (without base)
 * @returns {Promise<any>} Response data
 */
export async function get(path) {
	const res = await fetch(`${API_BASE}${path}`);
	if (!res.ok) {
		const error = await res.json().catch(() => ({ error: 'Unknown error' }));
		throw new Error(error.error || `HTTP ${res.status}`);
	}
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
	if (!res.ok) {
		const error = await res.json().catch(() => ({ error: 'Unknown error' }));
		throw new Error(error.error || `HTTP ${res.status}`);
	}
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
	if (!res.ok) {
		const error = await res.json().catch(() => ({ error: 'Unknown error' }));
		throw new Error(error.error || `HTTP ${res.status}`);
	}
	return res.json();
}

/**
 * Make a DELETE request to the API.
 * @param {string} path - API path (without base)
 * @returns {Promise<any>} Response data
 */
export async function del(path) {
	const res = await fetch(`${API_BASE}${path}`, { method: 'DELETE' });
	if (!res.ok) {
		const error = await res.json().catch(() => ({ error: 'Unknown error' }));
		throw new Error(error.error || `HTTP ${res.status}`);
	}
	return res.json();
}
