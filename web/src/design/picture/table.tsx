import type { Table } from "./model";
import { type Drawn, Label, r1 } from "./shapes";
import { fitted, room, squeezed, widthOf, written } from "./text";
import { type Tones, toneOf } from "./tones";

// A table, in the card's pixels: the room between a cell's text and the lines
// on either side of it, the room over and under it, and how narrow a column
// may be.
const aside = 5;
const above = 5;
const narrowest = 24;

// Cell is one cell of a table as it is drawn: its row and its column, counted
// from the header's, and what it writes.
type Cell = { line: number; column: number; text: string };

// named says whether a cell names something — a run of capitals, or the ? of
// what is asked — and is written strong, as a label is; a number or a time is
// written plain.
function named(text: string): boolean {
	return /^[A-Z]+$/.test(text) || text === "?";
}

// widthsOf are how wide a table's columns are at a size: each as wide as its
// widest cell with room on either side, and never narrower than narrowest.
function widthsOf(
	rows: readonly (readonly string[])[],
	size: number,
): number[] {
	const columns = rows[0]?.length ?? 0;
	return Array.from({ length: columns }, (_, column) =>
		Math.max(
			narrowest,
			...rows.map((row) => widthOf(row[column] ?? "", size) + 2 * aside),
		),
	);
}

// total is the sum of some lengths.
function total(lengths: readonly number[]): number {
	return lengths.reduce((sum, length) => sum + length, 0);
}

/**
 * drawTable draws a table: its cells in thin lines, the header row on a band
 * and written strong, the names in the cells strong and the numbers and times
 * plain, every cell centred in its column, each column as wide as its widest
 * cell, and the whole written at the largest size at which it fits the card.
 * A cell a page lights is filled and written in its tone.
 */
export function drawTable(
	table: Table,
	_locale?: string,
	tones?: Tones,
): Drawn {
	const header = table.header === undefined ? [] : [table.header.map(written)];
	const rows = [...header, ...table.rows.map((row) => row.map(written))];
	const size = fitted(squeezed, (one) => total(widthsOf(rows, one)) <= room);
	const widths = widthsOf(rows, size);
	let edge = 0;
	const edges = [edge];
	for (const column of widths) {
		edge += column;
		edges.push(edge);
	}
	const tall = size * 1.4 + 2 * above;
	const width = total(widths);
	const height = rows.length * tall;
	const cells: Cell[] = rows.flatMap((row, line) =>
		row.map((text, column) => ({ line, column, text })),
	);
	// toneAt is what a cell is lit in: a cell of the header is none of the
	// rows a page counts.
	const toneAt = (cell: Cell) =>
		cell.line < header.length
			? undefined
			: toneOf(tones, `cell ${cell.line - header.length}.${cell.column}`);
	const lit = cells.flatMap((cell) => {
		const tone = toneAt(cell);
		return tone === undefined ? [] : [{ cell, tone }];
	});
	let inner = "";
	for (let line = 1; line < rows.length; line++) {
		inner += `M0 ${r1(line * tall)}H${r1(width)}`;
	}
	for (const x of edges.slice(1, -1)) {
		inner += `M${r1(x)} 0V${r1(height)}`;
	}
	return {
		width,
		height,
		body: (
			<>
				{header.length > 0 && (
					<rect
						x={0}
						y={0}
						width={r1(width)}
						height={r1(tall)}
						class="mt-pic-band"
					/>
				)}
				{lit.map(({ cell, tone }) => (
					<rect
						key={`lit-${cell.line}-${cell.column}`}
						x={r1((edges[cell.column] ?? 0) + 2)}
						y={r1(cell.line * tall + 2)}
						width={r1((widths[cell.column] ?? 0) - 4)}
						height={r1(tall - 4)}
						rx={4}
						class="mt-pic-fill"
						data-tone={tone}
					/>
				))}
				{cells.map(
					(cell) =>
						cell.text !== "" && (
							<Label
								key={`${cell.line}-${cell.column}`}
								x={
									((edges[cell.column] ?? 0) + (edges[cell.column + 1] ?? 0)) /
									2
								}
								y={(cell.line + 0.5) * tall}
								text={cell.text}
								size={size}
								strong={cell.line < header.length || named(cell.text)}
								tone={toneAt(cell)}
							/>
						),
				)}
				{inner !== "" && (
					<path d={inner} class="mt-pic-line" stroke-width={1} />
				)}
				<rect
					x={0}
					y={0}
					width={r1(width)}
					height={r1(height)}
					class="mt-pic-line"
					stroke-width={1.5}
				/>
			</>
		),
	};
}
