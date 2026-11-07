import type { Lattice, LatticeLine } from "./model";
import { type PointLabel, pointLabelsOf } from "./pointLabels";
import { type Drawn, Label, r1, sizes } from "./shapes";
import { fitted, room, squeezed, widthOf, written } from "./text";

// A lattice, in the card's pixels: the largest a cell may be, the room kept
// round the paper for the dots on its edges, the radius of a dot, how far a
// point's label stands off its corner, and the room a mark keeps in its cell.
const largestCell = 40;
const edge = 4;
const dot = 3.5;
const offset = 4;
const inset = 4;

// Laid is a lattice laid out at a size of its labels: the room past each edge
// of its paper, and the side of a cell.
type Laid = {
	left: number;
	right: number;
	top: number;
	bottom: number;
	cell: number;
};

// Written is a point's label as the drawing writes it.
type Written = PointLabel & { text: string };

/**
 * drawLattice draws a figure on squared paper: the paper's grid faint, the
 * filled cells in the shading tone, the lines strong and the dashed lines
 * dashed, a dot at each point with its label beside it, in the cell round its
 * corner pointLabelsOf gives it, and each mark in its cell. Every label is
 * written at one size, the largest at which each label on the paper fits its
 * cell; the picture grows past an edge of the paper only by a label that
 * stands there.
 */
export function drawLattice(lattice: Lattice): Drawn {
	const labels = pointLabelsOf(lattice);
	if (labels === undefined) {
		throw new Error("a point's label has no cell of its own");
	}
	const marks = (lattice.marks ?? []).map((one) => ({
		at: one.at,
		text: written(one.label ?? ""),
	}));
	const placed: Written[] = labels.map((one) => ({
		...one,
		text: written(one.label),
	}));
	const fits = (size: number) => {
		const { cell } = laidOut(lattice, placed, size);
		return (
			marks.every((mark) => widthOf(mark.text, size) + inset <= cell) &&
			placed.every(
				(one) =>
					!onThePaper(lattice, one) ||
					(widthOf(one.text, size) + offset <= cell && size + offset <= cell),
			)
		);
	};
	const size = fitted([sizes.label, ...squeezed], fits);
	const { left, right, top, bottom, cell } = laidOut(lattice, placed, size);
	const x = (column: number) => left + column * cell;
	const y = (line: number) => top + line * cell;
	let paper = "";
	for (let line = 0; line <= lattice.height; line++) {
		paper += `M${r1(x(0))} ${r1(y(line))}H${r1(x(lattice.width))}`;
	}
	for (let column = 0; column <= lattice.width; column++) {
		paper += `M${r1(x(column))} ${r1(y(0))}V${r1(y(lattice.height))}`;
	}
	const path = (segments: readonly LatticeLine[]) =>
		segments
			.map(
				([x1, y1, x2, y2]) =>
					`M${r1(x(x1))} ${r1(y(y1))}L${r1(x(x2))} ${r1(y(y2))}`,
			)
			.join("");
	const lines = lattice.lines ?? [];
	const dashed = lattice.dashed ?? [];
	return {
		width: left + lattice.width * cell + right,
		height: top + lattice.height * cell + bottom,
		body: (
			<>
				{(lattice.cells ?? []).map(([column, line]) => (
					<rect
						key={`cell ${column} ${line}`}
						x={r1(x(column))}
						y={r1(y(line))}
						width={r1(cell)}
						height={r1(cell)}
						class="mt-pic-fill"
					/>
				))}
				<path d={paper} class="mt-pic-line mt-pic-faint" stroke-width={0.75} />
				{lines.length > 0 && (
					<path
						d={path(lines)}
						class="mt-pic-line"
						stroke-width={2.5}
						stroke-linecap="round"
					/>
				)}
				{dashed.length > 0 && (
					<path
						d={path(dashed)}
						class="mt-pic-line"
						stroke-width={2}
						stroke-dasharray="6 4"
					/>
				)}
				{(lattice.points ?? []).map(({ at: [column, line] }) => (
					<circle
						key={`point ${column} ${line}`}
						cx={r1(x(column))}
						cy={r1(y(line))}
						r={dot}
						class="mt-pic-ink"
					/>
				))}
				{placed.map(({ at: [column, line], text, quarter }) => (
					<Label
						key={`label ${column} ${line}`}
						x={x(column) + (quarter.dx === 0 ? offset : -offset)}
						y={y(line) + (quarter.dy === 0 ? 1 : -1) * (size / 2 + offset)}
						text={text}
						size={size}
						anchor={quarter.dx === 0 ? "start" : "end"}
					/>
				))}
				{marks.map((mark) => (
					<Label
						key={`mark ${mark.at[0]} ${mark.at[1]}`}
						x={x(mark.at[0]) + cell / 2}
						y={y(mark.at[1]) + cell / 2}
						text={mark.text}
						size={size}
					/>
				))}
			</>
		),
	};
}

// laidOut is how a lattice is laid out with its labels at a size: the room
// past each edge is what the labels standing past it take, and a cell is as
// large as the rest of the room allows.
function laidOut(
	lattice: Lattice,
	placed: readonly Written[],
	size: number,
): Laid {
	const past = (stands: (one: PointLabel) => boolean, across: boolean) => {
		const there = placed.filter(stands);
		const most = across
			? Math.max(0, ...there.map((one) => widthOf(one.text, size)))
			: size;
		return there.length === 0 ? edge : Math.max(edge, most + offset);
	};
	const column = (one: PointLabel) => one.at[0] + one.quarter.dx;
	const line = (one: PointLabel) => one.at[1] + one.quarter.dy;
	const left = past((one) => column(one) < 0, true);
	const right = past((one) => column(one) >= lattice.width, true);
	const top = past((one) => line(one) < 0, false);
	const bottom = past((one) => line(one) >= lattice.height, false);
	const cell = Math.min(
		largestCell,
		(room - left - right) / lattice.width,
		(room - top - bottom) / lattice.height,
	);
	return { left, right, top, bottom, cell };
}

// onThePaper says whether a point's label stands in a cell of the paper.
function onThePaper(lattice: Lattice, one: PointLabel): boolean {
	const column = one.at[0] + one.quarter.dx;
	const line = one.at[1] + one.quarter.dy;
	return (
		column >= 0 && line >= 0 && column < lattice.width && line < lattice.height
	);
}
