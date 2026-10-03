/**
 * byCodeUnits orders strings by their code units, the same on every machine
 * whatever language it is set to, so that an order never depends on where it
 * was worked out.
 */
export function byCodeUnits(a: string, b: string): number {
	if (a === b) {
		return 0;
	}
	return a < b ? -1 : 1;
}
