import type { Grid } from "./model";
import {
	type Drawn,
	Label,
	lineHeight,
	r1,
	Stack,
	sizes,
	stackOf,
} from "./shapes";
import { fitted, room, squeezed, widest, widthOf, written } from "./text";
import { type Tones, toneOf } from "./tones";

// A grid, in the card's pixels: how large a cell may be, and the room between
// the names of the rows and columns and the grid.
const largestCell = 40;
const nameGap = 6;

// Cell is one cell of a grid: its name, its row's name run into its column's,
// and its row and column, counted from the top and from the left.
type Cell = { name: string; line: number; column: number };

/**
 * drawGrid draws a grid of square cells, as large as the card allows: the
 * names of its rows to the left and of its columns above, each column's on as
 * many lines as keep them apart; the filled cells in the shading tone, and
 * each mark in its cell, written as large as the widest of them fits a cell.
 * A cell a page lights is filled, edged and written in its tone instead.
 */
export function drawGrid(grid: Grid, _locale?: string, tones?: Tones): Drawn {
	const rows = grid.rows.map((name, line) => ({
		name,
		line,
		text: written(name),
	}));
	const left =
		Math.max(...rows.map((row) => widthOf(row.text, sizes.label))) + nameGap;
	const cell = Math.min(largestCell, (room - left) / grid.cols.length);
	const right = left + grid.cols.length * cell;
	const names = grid.cols.map((name, column) => ({
		id: `col-${column}`,
		x: left + (column + 0.5) * cell,
		text: written(name),
	}));
	// The names of the columns may stand over the corner by the rows' names,
	// and a picture is as wide as its widest name, so that every one stays in
	// it.
	const width = Math.max(
		right,
		widest(
			names.map((one) => one.text),
			sizes.label,
		),
	);
	const columns = stackOf(names, 0, width);
	const top = columns.lines * lineHeight() + nameGap;
	const height = top + grid.rows.length * cell;
	const cells: Cell[] = grid.rows.flatMap((row, line) =>
		grid.cols.map((col, column) => ({ name: `${row}${col}`, line, column })),
	);
	const filled = new Set(grid.filled ?? []);
	const marks = new Map(
		Object.entries(grid.marks ?? {}).map(([name, label]) => [
			name,
			written(label),
		]),
	);
	const size = fitted(squeezed, (one) =>
		[...marks.values()].every((text) => widthOf(text, one) + 4 <= cell),
	);
	const at = ({ line, column }: Cell) => ({
		x: left + column * cell,
		y: top + line * cell,
	});
	const toneAt = (one: Cell) => toneOf(tones, `cell ${one.name}`);
	let lines = "";
	for (let line = 0; line <= grid.rows.length; line++) {
		lines += `M${r1(left)} ${r1(top + line * cell)}H${r1(right)}`;
	}
	for (let column = 0; column <= grid.cols.length; column++) {
		lines += `M${r1(left + column * cell)} ${r1(top)}V${r1(height)}`;
	}
	return {
		width,
		height,
		body: (
			<>
				{cells
					.filter((one) => filled.has(one.name) && toneAt(one) === undefined)
					.map((one) => (
						<rect
							key={one.name}
							x={r1(at(one).x)}
							y={r1(at(one).y)}
							width={r1(cell)}
							height={r1(cell)}
							class="mt-pic-fill"
						/>
					))}
				<path d={lines} class="mt-pic-line" stroke-width={1.5} />
				{cells
					.filter((one) => toneAt(one) !== undefined)
					.map((one) => (
						<rect
							key={`lit-${one.name}`}
							x={r1(at(one).x + 3)}
							y={r1(at(one).y + 3)}
							width={r1(cell - 6)}
							height={r1(cell - 6)}
							rx={4}
							class="mt-pic-fill"
							data-tone={toneAt(one)}
						/>
					))}
				<Stack
					placed={columns.placed}
					first={top - nameGap - lineHeight() / 2}
					upward
				/>
				{rows.map((row) => (
					<Label
						key={row.name}
						x={(left - nameGap) / 2}
						y={top + (row.line + 0.5) * cell}
						text={row.text}
					/>
				))}
				{cells
					.filter((one) => marks.has(one.name))
					.map((one) => (
						<Label
							key={one.name}
							x={at(one).x + cell / 2}
							y={at(one).y + cell / 2}
							text={marks.get(one.name) ?? ""}
							size={size}
							tone={toneAt(one)}
						/>
					))}
			</>
		),
	};
}
