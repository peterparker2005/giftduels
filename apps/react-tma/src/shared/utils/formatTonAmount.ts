// src/shared/utils/formatTonAmount.ts

/**
 * Форматирует TON-сумму:
 * - всегда показывает fractionDigits после точки (по умолчанию 2),
 * - группирует тысячи пробелами,
 * - оставляет точку как разделитель дробной части.
 *
 * @param value — входная сумма (number или строка).
 * @param fractionDigits — сколько цифр после точки (дефолт 2).
 * @returns Строка вида "1 234 567.89".
 */
export function formatTonAmount(
	value?: number | string,
	fractionDigits = 2,
): string {
	if (!value) return "0";
	// Приводим к числу
	const n =
		typeof value === "number" ? value : parseFloat(value.replace(/\s+/g, ""));
	if (Number.isNaN(n)) {
		// если не число — возвращаем оригинал
		return String(value);
	}

	// Формируем fixed-строку, разбиваем на целую и дробную части
	const [intPart, fracPart] = n.toFixed(fractionDigits).split(".");

	// Вставляем пробелы в тысячах
	const intWithSpaces = intPart.replace(/\B(?=(\d{3})+(?!\d))/g, " ");

	return fractionDigits > 0 ? `${intWithSpaces}.${fracPart}` : intWithSpaces;
}
