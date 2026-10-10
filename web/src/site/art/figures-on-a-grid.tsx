import type { ComponentChildren } from "preact";
import type { PageReader } from "../reader";
import { Key, Line, Op, Say, Stack, Versus } from "../Sketch";
import type { TopicArt } from "./art";

// Point is a point of the paper's grid, its column and its row counted in
// cells from the top left corner.
type Point = readonly [number, number];

// figureCells are the characters of the cells that belong to a figure:
// blue, peach and half filled; the paper around a figure is not part of it.
const figureCells = new Set(["c", "w", "t"]);

// borderOf is the border of a figure, the sides of its cells that face no
// other cell of it, as segments from point to point.
function borderOf(cells: readonly string[]): [Point, Point][] {
	const inFigure = (row: number, column: number) =>
		figureCells.has(cells[row]?.[column] ?? " ");
	const sides: [Point, Point][] = [];
	cells.forEach((line, row) => {
		[...line].forEach((_, column) => {
			if (!inFigure(row, column)) {
				return;
			}
			if (!inFigure(row - 1, column)) {
				sides.push([
					[column, row],
					[column + 1, row],
				]);
			}
			if (!inFigure(row + 1, column)) {
				sides.push([
					[column, row + 1],
					[column + 1, row + 1],
				]);
			}
			if (!inFigure(row, column - 1)) {
				sides.push([
					[column, row],
					[column, row + 1],
				]);
			}
			if (!inFigure(row, column + 1)) {
				sides.push([
					[column + 1, row],
					[column + 1, row + 1],
				]);
			}
		});
	});
	return sides;
}

// pathOf is the path through points of the grid, in pixels of a cell's size.
function pathOf(points: readonly Point[], size: number, at: number): string {
	return points
		.map(([x, y], place) => `${place === 0 ? "M" : "L"}${at + x * size} ${at + y * size}`)
		.join("");
}

// Figure is a figure of cells on squared paper: rows of characters, a cell
// painted by its character — "c" blue, "w" peach, "t" its lower left half
// blue, "." the empty paper, " " nothing at all — with the number written in
// a cell by "row.column", the figure's border in the accent, the cuts along
// the grid's lines in ink, the sides counted twice in dashed red, and a dot
// in the corner of a cell.
function Figure({
	cells,
	size = 20,
	marks = {},
	border = false,
	cuts = [],
	twice = [],
	dots = [],
}: {
	cells: readonly string[];
	size?: number;
	marks?: Readonly<Record<string, string>>;
	border?: boolean;
	cuts?: readonly (readonly Point[])[];
	twice?: readonly (readonly Point[])[];
	dots?: readonly Point[];
}) {
	const columns = Math.max(...cells.map((line) => line.length));
	const at = 2;
	const width = columns * size + 2 * at;
	const height = cells.length * size + 2 * at;
	return (
		<svg
			class="s-grid-figure"
			width={width}
			height={height}
			viewBox={`0 0 ${width} ${height}`}
		>
			{cells.flatMap((line, row) =>
				[...line].map((paint, column) => {
					if (paint === " ") {
						return null;
					}
					const x = at + column * size;
					const y = at + row * size;
					return (
						<g key={`${row}.${column}`}>
							<rect
								class="s-grid-cell"
								data-paint={paint}
								x={x}
								y={y}
								width={size}
								height={size}
							/>
							{paint === "t" && (
								<path
									class="s-grid-half"
									d={`M${x} ${y}L${x} ${y + size}L${x + size} ${y + size}Z`}
								/>
							)}
						</g>
					);
				}),
			)}
			{Object.entries(marks).map(([cell, text]) => {
				const [row = 0, column = 0] = cell.split(".").map(Number);
				return (
					<text
						key={cell}
						class="s-grid-mark"
						x={at + (column + 0.5) * size}
						y={at + (row + 0.5) * size}
						text-anchor="middle"
						dominant-baseline="central"
					>
						{text}
					</text>
				);
			})}
			{twice.map((points, place) => (
				<path key={place} class="s-grid-twice" d={pathOf(points, size, at)} />
			))}
			{border && (
				<path
					class="s-grid-border"
					d={borderOf(cells)
						.map((side) => pathOf(side, size, at))
						.join("")}
				/>
			)}
			{cuts.map((points, place) => (
				<path key={place} class="s-grid-cut" d={pathOf(points, size, at)} />
			))}
			{dots.map(([column, row]) => (
				<circle
					key={`${column}.${row}`}
					class="s-grid-dot"
					cx={at + column * size + 6}
					cy={at + row * size + 6}
					r={2.5}
				/>
			))}
		</svg>
	);
}

// Crossing is where a segment crosses a line of the grid: across a line
// that runs down, one that runs across, or both at once at a node.
type Crossing = { x: number; y: number; line: "down" | "across" | "node" };

// crossingsOf are the points where the segment from the bottom left corner
// of a rectangle of so many rows and columns to its top right one crosses
// the grid's lines inside it.
function crossingsOf(rows: number, columns: number): Crossing[] {
	const crossings: Crossing[] = [];
	for (let x = 1; x < columns; x++) {
		const y = rows - (rows * x) / columns;
		crossings.push({ x, y, line: Number.isInteger(y) ? "node" : "down" });
	}
	for (let y = 1; y < rows; y++) {
		const x = (columns * (rows - y)) / rows;
		if (!Number.isInteger(x)) {
			crossings.push({ x, y, line: "across" });
		}
	}
	return crossings;
}

// passedOf are the cells whose inside the segment passes through, by
// "row.column": between each two crossings it runs inside one cell.
function passedOf(rows: number, columns: number): Set<string> {
	const along = [0, 1, ...crossingsOf(rows, columns).map((one) => one.x / columns)]
		.sort((one, other) => one - other);
	const passed = new Set<string>();
	for (let place = 1; place < along.length; place++) {
		const middle = ((along[place - 1] ?? 0) + (along[place] ?? 0)) / 2;
		const x = middle * columns;
		const y = rows - middle * rows;
		passed.add(`${Math.floor(y)}.${Math.floor(x)}`);
	}
	return passed;
}

// Diagonal is a rectangle of squared paper with the segment from its bottom
// left corner to its top right one: where it crosses the grid's lines, blue
// on the lines that run down, orange on those that run across and red at a
// node; and the cells it passes through, green.
function Diagonal({
	rows,
	columns,
	points = false,
	passed = false,
}: {
	rows: number;
	columns: number;
	points?: boolean;
	passed?: boolean;
}) {
	const size = 30;
	const at = 4;
	const width = columns * size + 2 * at;
	const height = rows * size + 2 * at;
	const through = passed ? passedOf(rows, columns) : new Set<string>();
	return (
		<svg
			class="s-grid-figure"
			width={width}
			height={height}
			viewBox={`0 0 ${width} ${height}`}
		>
			{Array.from({ length: rows }, (_, row) =>
				Array.from({ length: columns }, (_, column) => (
					<rect
						key={`${row}.${column}`}
						class="s-grid-cell"
						data-paint={through.has(`${row}.${column}`) ? "passed" : "paper"}
						x={at + column * size}
						y={at + row * size}
						width={size}
						height={size}
					/>
				)),
			)}
			<path
				class="s-grid-segment"
				d={`M${at} ${at + rows * size}L${at + columns * size} ${at}`}
			/>
			{points &&
				crossingsOf(rows, columns).map((one) => (
					<circle
						key={`${one.x}.${one.y}`}
						class="s-grid-crossing"
						data-line={one.line}
						cx={at + one.x * size}
						cy={at + one.y * size}
						r={one.line === "node" ? 5.5 : 4}
					/>
				))}
		</svg>
	);
}

// Spot is a dot of the colour of the crossings on one kind of line, as the
// sample of a key.
function Spot({ line }: { line: "down" | "across" }) {
	return <span class="s-grid-spot" data-line={line} />;
}

// strips are the four ways to cut a strip of 2 by 4 into two whole parts of
// 4 cells: along it, across it, by a step, and by the step the other way.
const strips = {
	along: { cells: ["cccc", "wwww"], cut: [[0, 1], [4, 1]] },
	across: { cells: ["ccww", "ccww"], cut: [[2, 0], [2, 2]] },
	step: { cells: ["cccw", "cwww"], cut: [[3, 0], [3, 1], [1, 1], [1, 2]] },
	back: { cells: ["cwww", "cccw"], cut: [[1, 0], [1, 1], [3, 1], [3, 2]] },
} as const satisfies Record<
	string,
	{ cells: readonly string[]; cut: readonly Point[] }
>;

// Strip is a strip of 2 by 4 cut one of the four ways, its first part blue
// and its second peach, small in a trap's row, with a dot in its top left
// cell where it shows which part that cell keeps.
function Strip({
	way,
	size = 20,
	dot = false,
}: {
	way: keyof typeof strips;
	size?: number;
	dot?: boolean;
}) {
	const strip = strips[way];
	return (
		<Figure
			cells={strip.cells}
			size={size}
			cuts={[strip.cut]}
			dots={dot ? [[0, 0]] : []}
		/>
	);
}

// Ringed is a drawing ringed as a whole: dashed red where it is wrong and
// green where it is right.
function Ringed({
	way,
	children,
}: {
	way: "wrong" | "right";
	children: ComponentChildren;
}) {
	return (
		<span class="s-grid-ringed" data-way={way}>
			{children}
		</span>
	);
}

// Sum is a count written out in bold.
function Sum({ children }: { children: ComponentChildren }) {
	return <span class="s-grid-sum">{children}</span>;
}

// Perimeter is the words of a cell of the key idea's table in the accent of
// a border: the perimeter it gives, read from the cell's own words.
function Perimeter({ page, row }: { page: PageReader; row: number }) {
	return (
		<span class="s-grid-perimeter">
			{page.text(`idea.table.rows.${row}.3`)}
		</span>
	);
}

// cross is the cross of five cells, on the paper of its corners.
const cross = [".c.", "ccc", ".c."];

/**
 * art are the drawings of figures on squared paper: the cells that make a
 * figure's area, the border that makes its perimeter with the number of
 * sides each cell turns outward, strips cut into two whole parts, and a
 * segment across a rectangle with the lines it crosses and the cells it
 * passes through.
 */
export const art: TopicArt = {
	heroLine: ({ page, at }) => page.text(`${at}.line`),
	hero: ({ page, at }) => (
		<>
			<Line gap={24} align="end">
				<Stack gap={8}>
					<Figure
						cells={["c.", "cc"]}
						size={40}
						marks={{ "0.0": "1", "1.0": "2", "1.1": "3" }}
					/>
					<Say>{page.text(`${at}.area`)}</Say>
				</Stack>
				<Stack gap={8}>
					<Figure
						cells={["c.", "cc"]}
						size={40}
						marks={{ "0.0": "3", "1.0": "2", "1.1": "3" }}
						border
					/>
					<span class="s-grid-accent">{page.text(`${at}.perimeter`)}</span>
				</Stack>
			</Line>
			<Say>{page.text(`${at}.out`)}</Say>
		</>
	),
	basis: [
		({ page, at }) => (
			<Line gap={20} align="end">
				<Stack gap={6}>
					<Figure
						cells={["cc", "c."]}
						size={30}
						marks={{ "0.0": "1", "0.1": "2", "1.0": "3" }}
					/>
					<Say>{page.text(`${at}.cells`)}</Say>
				</Stack>
				<Stack gap={6}>
					<Figure cells={["tt"]} size={30} />
					<Say>½ + ½ = 1</Say>
				</Stack>
			</Line>
		),
		({ page, at }) => (
			<Stack gap={8}>
				<Figure
					cells={["cc"]}
					size={38}
					border
					twice={[
						[
							[1, 0],
							[1, 1],
						],
					]}
				/>
				<Say way="wrong">{page.text(`${at}.shared`)}</Say>
			</Stack>
		),
		({ page, at }) => (
			<Line gap={20} align="end">
				<Stack gap={6}>
					<Figure cells={["cccc"]} size={22} border />
					<Say>{page.text(`${at}.perimeter`, { count: 10 })}</Say>
				</Stack>
				<Stack gap={6}>
					<Figure cells={["cc", "cc"]} size={22} border />
					<Say>{page.text(`${at}.perimeter`, { count: 8 })}</Say>
				</Stack>
			</Line>
		),
	],
	idea: [
		{
			column: 0,
			words: "beside",
			rows: [
				() => <Figure cells={["c"]} size={18} marks={{ "0.0": "4" }} border />,
				() => (
					<Figure
						cells={["cc"]}
						size={18}
						marks={{ "0.0": "3", "0.1": "3" }}
						border
					/>
				),
				() => (
					<Figure
						cells={["cccc"]}
						size={18}
						marks={{ "0.0": "3", "0.1": "2", "0.2": "2", "0.3": "3" }}
						border
					/>
				),
				() => (
					<Figure
						cells={["cc", "cc"]}
						size={18}
						marks={{ "0.0": "2", "0.1": "2", "1.0": "2", "1.1": "2" }}
						border
					/>
				),
			],
		},
		{
			column: 2,
			rows: [1, 2, 3, 4].map(
				(row) =>
					({ page }: { page: PageReader }) => <Perimeter page={page} row={row} />,
			),
		},
	],
	legend: ({ page, at }) => (
		<>
			<Key sample={<span class="s-grid-line" data-line="border" />}>
				{page.text(`${at}.border`)}
			</Key>
			<Key sample={<span class="s-grid-line" data-line="cut" />}>
				{page.text(`${at}.cut`)}
			</Key>
			<Key sample={<span class="s-grid-sample">4</span>}>
				{page.text(`${at}.count`)}
			</Key>
		</>
	),
	examples: [
		{
			task: () => <Figure cells={cross} size={30} />,
			steps: [
				({ page, at }) => (
					<>
						<Figure
							cells={cross}
							size={30}
							marks={{ "0.1": "3", "1.0": "3", "1.2": "3", "2.1": "3" }}
						/>
						<Say>{page.text(`${at}.ends`)}</Say>
					</>
				),
				({ page, at }) => (
					<>
						<Figure cells={cross} size={30} marks={{ "1.1": "0" }} />
						<Say>{page.text(`${at}.middle`)}</Say>
					</>
				),
				({ page, at }) => (
					<>
						<Figure cells={cross} size={30} border />
						<span class="s-grid-accent">{page.text(`${at}.border`)}</span>
					</>
				),
			],
			note: ({ page, at }) => (
				<Line gap={14}>
					<Figure
						cells={cross}
						size={26}
						twice={[
							[
								[1, 1],
								[2, 1],
								[2, 2],
								[1, 2],
								[1, 1],
							],
						]}
					/>
					<Stack gap={4} align="start">
						<Say way="wrong">{page.text(`${at}.inner`)}</Say>
						<Sum>20 − 2 × 4 = 12</Sum>
					</Stack>
				</Line>
			),
		},
		{
			task: () => <Figure cells={["cccc", "cccc"]} size={26} />,
			steps: [
				({ page, at }) => (
					<>
						<Strip way="along" size={24} />
						<Say>{page.text(`${at}.along`)}</Say>
					</>
				),
				({ page, at }) => (
					<>
						<Strip way="across" size={24} />
						<Say>{page.text(`${at}.across`)}</Say>
					</>
				),
				({ page, at }) => (
					<Line gap={12}>
						<Stack gap={6}>
							<Strip way="step" size={24} />
							<Say>{page.text(`${at}.step`)}</Say>
						</Stack>
						<Stack gap={6}>
							<Strip way="back" size={24} />
							<Say>{page.text(`${at}.back`)}</Say>
						</Stack>
					</Line>
				),
			],
			note: ({ page, at }) => (
				<Stack gap={10}>
					<Line gap={8}>
						<Strip way="along" />
						<Op>=</Op>
						<Figure cells={["wwww", "cccc"]} size={20} cuts={[strips.along.cut]} />
						<Say way="wrong">{page.text(`${at}.same`)}</Say>
					</Line>
					<Line gap={6}>
						<Strip way="along" size={16} dot />
						<Strip way="across" size={16} dot />
						<Strip way="step" size={16} dot />
						<Strip way="back" size={16} dot />
					</Line>
					<Say>{page.text(`${at}.dot`)}</Say>
				</Stack>
			),
		},
		{
			task: () => <Diagonal rows={3} columns={5} />,
			steps: [
				({ page, at }) => (
					<>
						<Diagonal rows={3} columns={5} points />
						<Say>{page.text(`${at}.points`)}</Say>
					</>
				),
				({ page, at }) => (
					<>
						<Diagonal rows={3} columns={5} points />
						<Line gap={14}>
							<Key sample={<Spot line="down" />}>{page.text(`${at}.down`)}</Key>
							<Key sample={<Spot line="across" />}>
								{page.text(`${at}.across`)}
							</Key>
						</Line>
					</>
				),
				({ page, at }) => (
					<>
						<Diagonal rows={3} columns={5} passed />
						<Say>{page.text(`${at}.green`)}</Say>
					</>
				),
			],
			note: ({ page, at }) => (
				<>
					<Diagonal rows={4} columns={6} points passed />
					<Say>{page.text(`${at}.node`)}</Say>
				</>
			),
		},
	],
	traps: {
		double_count: ({ page, at }) => (
			<>
				<Say>{page.text(`${at}.question`)}</Say>
				<Versus
					child={
						<>
							<Figure
								cells={["cc"]}
								size={20}
								twice={[
									[
										[1, 0],
										[1, 1],
									],
								]}
							/>
							<Sum>4 + 4 = 8</Sum>
						</>
					}
					right={
						<>
							<Figure cells={["cc"]} size={20} border />
							<Sum>= 6</Sum>
						</>
					}
				/>
			</>
		),
		missed_case: ({ page, at }) => (
			<>
				<Say>{page.text(`${at}.question`)}</Say>
				<Versus
					child={
						<>
							<Strip way="along" size={12} />
							<Strip way="across" size={12} />
							<Strip way="step" size={12} />
							<Sum>= 3</Sum>
						</>
					}
					right={
						<>
							<Strip way="along" size={12} />
							<Strip way="across" size={12} />
							<Strip way="step" size={12} />
							<Strip way="back" size={12} />
							<Sum>= 4</Sum>
						</>
					}
				/>
			</>
		),
		ignored_condition: ({ page, at }) => (
			<>
				<Say>{page.text(`${at}.question`)}</Say>
				<Versus
					child={
						<>
							<Ringed way="wrong">
								<Figure
									cells={["cwwc", "cwwc"]}
									size={18}
									cuts={[
										[
											[1, 0],
											[1, 2],
										],
										[
											[3, 0],
											[3, 2],
										],
									]}
								/>
							</Ringed>
							<Say way="wrong">{page.text(`${at}.broken`)}</Say>
						</>
					}
					right={
						<>
							<Ringed way="right">
								<Strip way="across" size={18} />
							</Ringed>
							<Say way="right">{page.text(`${at}.whole`)}</Say>
						</>
					}
				/>
			</>
		),
	},
};
