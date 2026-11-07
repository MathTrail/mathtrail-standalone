import type { ShapeName, Shapes } from "./model";
import { type Drawn, Label, r1, sizes } from "./shapes";
import { room, widthOf, written } from "./text";

// A row of shapes, in the card's pixels: the largest a shape is drawn, the
// gap between two shapes and between two rows, and the room before a total.
const largestShape = 30;
const gap = 8;
const beforeTotal = 10;

/**
 * drawShapes draws rows of shapes, each shape drawn alike wherever it stands,
 * the rows' shapes in columns, and each row's total after an equals sign, the
 * totals lined up after the longest row. The file is named for the totals,
 * since shapes.tsx holds what every drawing draws with.
 */
export function drawShapes(shapes: Shapes): Drawn {
	const most = Math.max(...shapes.rows.map((row) => row.shapes.length));
	const totals = shapes.rows.map((row) => `= ${written(row.total)}`);
	const totalWidth = Math.max(
		...totals.map((text) => widthOf(text, sizes.label)),
	);
	const side = Math.min(
		largestShape,
		(room - beforeTotal - totalWidth - (most - 1) * gap) / most,
	);
	const totalAt = most * side + (most - 1) * gap + beforeTotal;
	return {
		width: totalAt + totalWidth,
		height: shapes.rows.length * (side + gap) - gap,
		body: (
			<>
				{shapes.rows.map((row, line) => {
					const y = line * (side + gap);
					return (
						<g key={`row ${y}`}>
							{row.shapes.map((name, column) => (
								<Shape
									key={`${name} ${column}`}
									name={name}
									x={column * (side + gap)}
									y={y}
									side={side}
								/>
							))}
							<Label
								x={totalAt}
								y={y + side / 2}
								text={totals[line] ?? ""}
								anchor="start"
							/>
						</g>
					);
				})}
			</>
		),
	};
}

// Shape is one shape of a row, drawn in a square of the side given whose top
// left corner stands at x and y.
function Shape({
	name,
	x,
	y,
	side,
}: {
	name: ShapeName;
	x: number;
	y: number;
	side: number;
}) {
	const half = side / 2;
	const inset = 1;
	switch (name) {
		case "circle":
			return (
				<circle
					cx={r1(x + half)}
					cy={r1(y + half)}
					r={r1(half - inset)}
					class="mt-pic-line"
					stroke-width={2}
				/>
			);
		case "square":
			return (
				<rect
					x={r1(x + inset)}
					y={r1(y + inset)}
					width={r1(side - 2 * inset)}
					height={r1(side - 2 * inset)}
					class="mt-pic-line"
					stroke-width={2}
				/>
			);
		case "triangle":
			return (
				<path
					d={`M${r1(x + half)} ${r1(y + inset)}L${r1(x + side - inset)} ${r1(y + side - inset)}L${r1(x + inset)} ${r1(y + side - inset)}Z`}
					class="mt-pic-line"
					stroke-width={2}
				/>
			);
		case "diamond":
			return (
				<path
					d={`M${r1(x + half)} ${r1(y + inset)}L${r1(x + side - inset)} ${r1(y + half)}L${r1(x + half)} ${r1(y + side - inset)}L${r1(x + inset)} ${r1(y + half)}Z`}
					class="mt-pic-line"
					stroke-width={2}
				/>
			);
	}
}
