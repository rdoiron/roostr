import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import {
	formatUptime,
	formatBytes,
	formatRelativeTime,
	getKindLabel,
	truncatePubkey,
	formatCompactNumber
} from './format.js';

describe('formatUptime', () => {
	it('returns 0m for zero or negative seconds', () => {
		expect(formatUptime(0)).toBe('0m');
		expect(formatUptime(-1)).toBe('0m');
		expect(formatUptime(null)).toBe('0m');
		expect(formatUptime(undefined)).toBe('0m');
	});

	it('formats minutes only', () => {
		expect(formatUptime(60)).toBe('1m');
		expect(formatUptime(120)).toBe('2m');
		expect(formatUptime(59)).toBe('0m');
	});

	it('formats hours and minutes', () => {
		expect(formatUptime(3600)).toBe('1h');
		expect(formatUptime(3660)).toBe('1h 1m');
		expect(formatUptime(7200)).toBe('2h');
		expect(formatUptime(7320)).toBe('2h 2m');
	});

	it('formats days, hours, and minutes', () => {
		expect(formatUptime(86400)).toBe('1d');
		expect(formatUptime(90000)).toBe('1d 1h');
		expect(formatUptime(90060)).toBe('1d 1h 1m');
		expect(formatUptime(172800)).toBe('2d');
	});

	it('handles large values', () => {
		expect(formatUptime(604800)).toBe('7d'); // 1 week
		expect(formatUptime(2592000)).toBe('30d'); // 30 days
	});
});

describe('formatBytes', () => {
	it('returns 0 B for zero or falsy values', () => {
		expect(formatBytes(0)).toBe('0 B');
		expect(formatBytes(null)).toBe('0 B');
		expect(formatBytes(undefined)).toBe('0 B');
	});

	it('formats bytes', () => {
		expect(formatBytes(1)).toBe('1 B');
		expect(formatBytes(500)).toBe('500 B');
		expect(formatBytes(1023)).toBe('1023 B');
	});

	it('formats kilobytes', () => {
		expect(formatBytes(1024)).toBe('1.0 KB');
		expect(formatBytes(1536)).toBe('1.5 KB');
		expect(formatBytes(10240)).toBe('10.0 KB');
	});

	it('formats megabytes', () => {
		expect(formatBytes(1048576)).toBe('1.0 MB');
		expect(formatBytes(5242880)).toBe('5.0 MB');
	});

	it('formats gigabytes', () => {
		expect(formatBytes(1073741824)).toBe('1.0 GB');
		expect(formatBytes(10737418240)).toBe('10.0 GB');
	});

	it('formats terabytes', () => {
		expect(formatBytes(1099511627776)).toBe('1.0 TB');
	});
});

describe('formatRelativeTime', () => {
	let now;

	beforeEach(() => {
		now = Math.floor(Date.now() / 1000);
		vi.useFakeTimers();
		vi.setSystemTime(now * 1000);
	});

	afterEach(() => {
		vi.useRealTimers();
	});

	it('returns "just now" for recent timestamps', () => {
		expect(formatRelativeTime(now)).toBe('just now');
		expect(formatRelativeTime(now - 30)).toBe('just now');
		expect(formatRelativeTime(now - 59)).toBe('just now');
	});

	it('returns minutes ago', () => {
		expect(formatRelativeTime(now - 60)).toBe('1 min ago');
		expect(formatRelativeTime(now - 120)).toBe('2 min ago');
		expect(formatRelativeTime(now - 3599)).toBe('59 min ago');
	});

	it('returns hours ago', () => {
		expect(formatRelativeTime(now - 3600)).toBe('1 hour ago');
		expect(formatRelativeTime(now - 7200)).toBe('2 hours ago');
		expect(formatRelativeTime(now - 86399)).toBe('23 hours ago');
	});

	it('returns days ago', () => {
		expect(formatRelativeTime(now - 86400)).toBe('1 day ago');
		expect(formatRelativeTime(now - 172800)).toBe('2 days ago');
		expect(formatRelativeTime(now - 604800)).toBe('7 days ago');
	});

	it('handles ISO date strings', () => {
		const isoDate = new Date((now - 3600) * 1000).toISOString();
		expect(formatRelativeTime(isoDate)).toBe('1 hour ago');
	});
});

describe('getKindLabel', () => {
	it('returns known kind labels', () => {
		expect(getKindLabel(0)).toBe('Profile');
		expect(getKindLabel(1)).toBe('Post');
		expect(getKindLabel(3)).toBe('Follows');
		expect(getKindLabel(4)).toBe('DM');
		expect(getKindLabel(6)).toBe('Repost');
		expect(getKindLabel(7)).toBe('Reaction');
		expect(getKindLabel(14)).toBe('DM');
	});

	it('returns generic label for unknown kinds', () => {
		expect(getKindLabel(2)).toBe('Kind 2');
		expect(getKindLabel(1000)).toBe('Kind 1000');
		expect(getKindLabel(30023)).toBe('Kind 30023');
	});
});

describe('truncatePubkey', () => {
	it('handles empty or short values', () => {
		expect(truncatePubkey('')).toBe('');
		expect(truncatePubkey(null)).toBe('');
		expect(truncatePubkey(undefined)).toBe('');
		expect(truncatePubkey('abc')).toBe('abc');
		expect(truncatePubkey('123456789abcdef')).toBe('123456789abcdef'); // 15 chars - no truncation
	});

	it('truncates long pubkeys', () => {
		const pubkey = '1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef';
		expect(truncatePubkey(pubkey)).toBe('12345678...cdef');
	});

	it('shows first 8 and last 4 chars', () => {
		const pubkey = 'abcdefgh12345678ijklmnopqrstuvwxyz';
		expect(truncatePubkey(pubkey)).toBe('abcdefgh...wxyz');
	});
});

describe('formatCompactNumber', () => {
	it('returns raw number for small values', () => {
		expect(formatCompactNumber(0)).toBe('0');
		expect(formatCompactNumber(1)).toBe('1');
		expect(formatCompactNumber(999)).toBe('999');
		expect(formatCompactNumber(null)).toBe('0');
		expect(formatCompactNumber(undefined)).toBe('0');
	});

	it('formats thousands with K suffix', () => {
		expect(formatCompactNumber(1000)).toBe('1.0K');
		expect(formatCompactNumber(1500)).toBe('1.5K');
		expect(formatCompactNumber(9999)).toBe('10.0K');
		expect(formatCompactNumber(10000)).toBe('10K');
		expect(formatCompactNumber(999999)).toBe('1000K');
	});

	it('formats millions with M suffix', () => {
		expect(formatCompactNumber(1000000)).toBe('1.0M');
		expect(formatCompactNumber(1500000)).toBe('1.5M');
		expect(formatCompactNumber(10000000)).toBe('10M');
		expect(formatCompactNumber(999999999)).toBe('1000M');
	});

	it('formats billions with B suffix', () => {
		expect(formatCompactNumber(1000000000)).toBe('1.0B');
		expect(formatCompactNumber(2500000000)).toBe('2.5B');
	});
});
