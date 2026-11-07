import { describe, expect, test } from "vitest";
import type { Lattice } from "./model";
import {
	type Drawing,
	drawn,
	numberOf,
	pointsOfPath,
	shapesOf,
	textBox,
	textOf,
	textsOf,
} from "./testing/svg";

// draw is a lattice as the card draws it.
const draw = (lattice: Omit<Lattice, "kind">) =>
	drawn({ kind: "lattice", ...lattice });

// paperOf is the faint grid of a lattice's paper.
function paperOf(drawing: Drawing) {
	return shapesOf(drawing, "path", "mt-pic-line mt-pic-faint");
}

// cellOf is the room a cell of a lattice takes in its drawing, read from the
// faint grid of its paper: its corners are the paper's, a cell apart.
function cellOf(drawing: Drawing, column: number, line: number) {
	const points = pointsOfPath(paperOf(drawing)[0]?.attributes.d ?? "");
	const xs = [...new Set(points.map((one) => one.x))].sort((a, b) => a - b);
	const ys = [...new Set(points.map((one) => one.y))].sort((a, b) => a - b);
	return {
		left: xs[column] ?? Number.NaN,
		right: xs[column + 1] ?? Number.NaN,
		top: ys[line] ?? Number.NaN,
		bottom: ys[line + 1] ?? Number.NaN,
	};
}

// within says whether a text stands inside a cell of a lattice.
function within(drawing: Drawing, words: string, column: number, line: number) {
	const box = textBox(textOf(drawing, words));
	const cell = cellOf(drawing, column, line);
	return (
		box.left >= cell.left &&
		box.right <= cell.right &&
		box.top >= cell.top &&
		box.bottom <= cell.bottom
	);
}

describe("a lattice", () => {
	test("draws its paper faint, and what the description draws on it strong", () => {
		const drawing = draw({
			width: 3,
			height: 2,
			cells: [[1, 0]],
			lines: [[0, 0, 3, 0]],
			dashed: [[0, 1, 3, 1]],
		});

		expect(paperOf(drawing)).toHaveLength(1);
		expect(shapesOf(drawing, "rect", "mt-pic-fill")).toHaveLength(1);
		expect(
			shapesOf(drawing, "path", "mt-pic-line").map(
				(one) => one.attributes["stroke-dasharray"],
			),
		).toEqual([undefined, "6 4"]);
	});

	test("draws no line it is not given, plain or dashed", () => {
		const drawing = draw({ width: 2, height: 2 });

		expect(shapesOf(drawing, "path", "mt-pic-line")).toHaveLength(0);
	});

	test("draws a line from corner to corner of its paper", () => {
		const drawing = draw({ width: 2, height: 2, lines: [[0, 0, 2, 2]] });
		const [from, to] = pointsOfPath(
			shapesOf(drawing, "path", "mt-pic-line")[0]?.attributes.d ?? "",
		);
		const first = cellOf(drawing, 0, 0);
		const last = cellOf(drawing, 1, 1);

		expect(from).toEqual({ x: first.left, y: first.top });
		expect(to).toEqual({ x: last.right, y: last.bottom });
	});

	test("writes each mark in its cell", () => {
		const drawing = draw({
			width: 3,
			height: 3,
			marks: [
				{ at: [0, 0], label: "S" },
				{ at: [2, 1], label: "E" },
			],
		});

		expect(within(drawing, "S", 0, 0)).toBe(true);
		expect(within(drawing, "E", 2, 1)).toBe(true);
	});

	test("sets a dot at each point, labelled or not", () => {
		const drawing = draw({
			width: 2,
			height: 2,
			points: [{ at: [1, 1], label: "P" }, { at: [2, 0] }],
		});

		expect(shapesOf(drawing, "circle", "mt-pic-ink")).toHaveLength(2);
		expect(textsOf(drawing).map((one) => one.text)).toEqual(["P"]);
	});

	test("labels a point above it and to its right", () => {
		const drawing = draw({
			width: 2,
			height: 2,
			points: [{ at: [1, 1], label: "P" }],
		});

		expect(within(drawing, "P", 1, 0)).toBe(true);
	});

	test("labels a point in the next quarter of its corner where a mark stands in the first", () => {
		const drawing = draw({
			width: 2,
			height: 2,
			points: [{ at: [1, 1], label: "P" }],
			marks: [{ at: [1, 0], label: "T" }],
		});

		expect(within(drawing, "P", 0, 0)).toBe(true);
	});

	test("labels a point on the paper before it labels it past an edge", () => {
		const drawing = draw({
			width: 2,
			height: 2,
			points: [{ at: [0, 0], label: "P" }],
		});

		expect(within(drawing, "P", 0, 0)).toBe(true);
	});

	test("moves a label to another of its cells where a label beside it has no other", () => {
		const drawing = draw({
			width: 2,
			height: 2,
			points: [
				{ at: [2, 1], label: "B" },
				{ at: [1, 1], label: "A" },
			],
			marks: [
				{ at: [0, 0], label: "1" },
				{ at: [0, 1], label: "2" },
				{ at: [1, 1], label: "3" },
			],
		});

		expect(within(drawing, "A", 1, 0)).toBe(true);
		expect(textBox(textOf(drawing, "B")).left).toBeGreaterThan(
			cellOf(drawing, 1, 0).right,
		);
	});

	test("grows past an edge only by a label that stands there", () => {
		const alone = draw({ width: 1, height: 1, points: [{ at: [0, 0] }] });
		const outside = draw({
			width: 1,
			height: 1,
			points: [{ at: [1, 1], label: "P" }],
			marks: [{ at: [0, 0], label: "M" }],
		});
		const box = textBox(textOf(outside, "P"));
		const paper = cellOf(outside, 0, 0);

		expect(cellOf(alone, 0, 0).left).toBe(4);
		expect(box.left).toBeGreaterThan(paper.right);
		expect(box.bottom).toBeLessThanOrEqual(paper.bottom);
		expect(numberOf(textOf(outside, "P"), "font-size")).toBe(14);
	});
});
