// Vitest setup file
import { vi } from 'vitest';

// Mock localStorage
const localStorageMock = {
	store: {},
	getItem: vi.fn((key) => localStorageMock.store[key] || null),
	setItem: vi.fn((key, value) => {
		localStorageMock.store[key] = value;
	}),
	removeItem: vi.fn((key) => {
		delete localStorageMock.store[key];
	}),
	clear: vi.fn(() => {
		localStorageMock.store = {};
	})
};

Object.defineProperty(globalThis, 'localStorage', {
	value: localStorageMock,
	writable: true
});

// Reset localStorage between tests
beforeEach(() => {
	localStorageMock.store = {};
	localStorageMock.getItem.mockClear();
	localStorageMock.setItem.mockClear();
	localStorageMock.removeItem.mockClear();
	localStorageMock.clear.mockClear();
});

// Mock fetch globally
globalThis.fetch = vi.fn();

beforeEach(() => {
	vi.mocked(fetch).mockClear();
});
