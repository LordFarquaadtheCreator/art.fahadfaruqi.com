/** Bytes for a readout: 12400000 -> "11.8 MB". Binary units, one decimal. */
export function formatBytes(bytes: number): string | null {
	if (!Number.isFinite(bytes) || bytes <= 0) return null;

	const units = ['B', 'KB', 'MB', 'GB'];
	let value = bytes;
	let unit = 0;

	while (value >= 1024 && unit < units.length - 1) {
		value /= 1024;
		unit += 1;
	}

	return `${value.toFixed(unit === 0 ? 0 : 1)} ${units[unit]}`;
}
