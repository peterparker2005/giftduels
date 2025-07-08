export function formatThousands(n: number | undefined): string {
	if (!n) return "0";
	return n.toString().replace(/\B(?=(\d{3})+(?!\d))/g, " ");
}
