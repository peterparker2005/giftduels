export function formatFloat(
	value?: number | string,
	maxFractionDigits = 2,
): string {
	if (value === undefined || value === null || value === "") {
		return "0";
	}

	const n =
		typeof value === "number"
			? value
			: parseFloat(String(value).replace(/\s+/g, ""));

	if (Number.isNaN(n)) {
		return String(value);
	}

	const [intPart, rawFracPart = ""] = n.toFixed(maxFractionDigits).split(".");

	const fracPart = rawFracPart.replace(/0+$/, "");

	const intWithSpaces = intPart.replace(/\B(?=(\d{3})+(?!\d))/g, " ");

	if (fracPart === "") {
		return intWithSpaces;
	}

	return `${intWithSpaces}.${fracPart}`;
}
