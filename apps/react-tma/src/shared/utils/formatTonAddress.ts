/**
 * Форматирует длинный TON-адрес, оставляя первые `start` и последние `end` символов,
 * между ними вставляет многоточие.
 *
 * @param address — исходный адрес (например, из useTonAddress()).
 * @param start — сколько символов оставить спереди (по умолчанию 4).
 * @param end — сколько символов оставить в конце (по умолчанию 4).
 * @returns Короче строка вида "UQBp***CwNF".
 */
export function formatTonAddress(address: string, start = 4, end = 4): string {
	if (!address) return "";
	if (address.length <= start + end) return address;
	const prefix = address.slice(0, start);
	const suffix = address.slice(-end);
	return `${prefix}…${suffix}`;
}
