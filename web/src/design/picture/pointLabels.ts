import type { Corner, Lattice } from "./model";

/**
 * Quarter is one of the four cells round a corner of a lattice, as steps from
 * the corner's own cell, the one below and to the right of it.
 */
export type Quarter = { dx: 0 | -1; dy: 0 | -1 };

// quarters are the four cells round a corner, in the order a point's label
// is set in them: above and to the right, above and to the left, below and to
// the right, below and to the left.
const quarters: readonly Quarter[] = [
	{ dx: 0, dy: -1 },
	{ dx: -1, dy: -1 },
	{ dx: 0, dy: 0 },
	{ dx: -1, dy: 0 },
];

/** PointLabel is a point's label and the quarter of its corner it stands in. */
export type PointLabel = { at: Corner; label: string; quarter: Quarter };

/**
 * pointLabelsOf are the labels of a lattice's points, each given a cell of its
 * own round its corner, one no mark holds, or undefined where some label has
 * none. A cell of the paper is given before a cell past its edge, and each
 * label the first of its cells in the order of quarters that the others
 * leave it; where the others have taken all of them, one of the others is
 * moved to another of its own.
 */
export function pointLabelsOf(lattice: Lattice): PointLabel[] | undefined {
	const marked = new Set(
		(lattice.marks ?? []).map(({ at }) => keyOf(at[0], at[1])),
	);
	const onThePaper = ([column, line]: readonly [number, number]) =>
		column >= 0 && line >= 0 && column < lattice.width && line < lattice.height;
	const labelled = (lattice.points ?? []).flatMap((point) =>
		point.label === undefined ? [] : [{ at: point.at, label: point.label }],
	);
	// Each label's cells, a cell of the paper before one past its edge.
	const choices = labelled.map(({ at: [column, line] }) => {
		const cells = quarters
			.map((quarter) => ({
				quarter,
				cell: [column + quarter.dx, line + quarter.dy] as const,
			}))
			.filter(({ cell }) => !marked.has(keyOf(...cell)));
		return [
			...cells.filter(({ cell }) => onThePaper(cell)),
			...cells.filter(({ cell }) => !onThePaper(cell)),
		];
	});
	const holder = new Map<string, number>();
	const chosen: Quarter[] = [];
	// seat gives label i a cell, moving a label that holds one to another of
	// its own where it can be moved.
	const seat = (i: number, tried: Set<string>): boolean => {
		for (const { quarter, cell } of choices[i] ?? []) {
			const key = keyOf(...cell);
			if (tried.has(key)) {
				continue;
			}
			tried.add(key);
			const held = holder.get(key);
			if (held === undefined || seat(held, tried)) {
				holder.set(key, i);
				chosen[i] = quarter;
				return true;
			}
		}
		return false;
	};
	for (const [i] of labelled.entries()) {
		if (!seat(i, new Set())) {
			return undefined;
		}
	}
	return labelled.map((one, i) => ({
		...one,
		quarter: chosen[i] as Quarter,
	}));
}

// keyOf is what tells a cell from the others.
function keyOf(column: number, line: number): string {
	return `${column},${line}`;
}
