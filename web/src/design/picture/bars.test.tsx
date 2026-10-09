import { describe, expect, test } from "vitest";
import type { Bars } from "./model";
import {
	type Drawing,
	drawn,
	numberOf,
	pointsOfPath,
	type Shape,
	shapesOf,
	textOf,
	textsOf,
} from "./testing/svg";

// outlinesOf are the outlines of a drawing's parts and segments, bar by bar
// from the top: each part or segment a rectangle.
function outlinesOf(drawing: Drawing): Shape[][] {
	const rows = new Map<number, Shape[]>();
	for (const part of shapesOf(drawing, "rect", "mt-pic-line")) {
		const top = numberOf(part, "y");
		rows.set(top, [...(rows.get(top) ?? []), part]);
	}
	return [...rows.entries()]
		.sort(([one], [other]) => one - other)
		.map(([, parts]) => parts);
}

// lengthOf is how long a run of parts is drawn.
function lengthOf(parts: readonly Shape[]): number {
	return parts.reduce((sum, part) => sum + numberOf(part, "width"), 0);
}

// draw is a picture of bars as the card draws it.
const draw = (bars: Omit<Bars, "kind">) => drawn({ kind: "bars", ...bars });

describe("bars", () => {
	test("are drawn to one scale, a bar as long as its parts or its length", () => {
		const [three, four, six] = outlinesOf(
			draw({ bars: [{ parts: 3 }, { parts: 4 }, { parts: 2, length: 6 }] }),
		).map(lengthOf);

		expect((four ?? 0) / (three ?? 1)).toBeCloseTo(4 / 3, 2);
		expect((six ?? 0) / (three ?? 1)).toBeCloseTo(2, 2);
	});

	test("cut into equal parts shade as many as they say, from the start", () => {
		const drawing = draw({ bars: [{ parts: 5, shaded: 2 }] });
		const shaded = shapesOf(drawing, "rect", "mt-pic-fill");
		const [parts = []] = outlinesOf(drawing);

		expect(parts).toHaveLength(5);
		expect(shaded.map((one) => numberOf(one, "x"))).toEqual(
			parts.slice(0, 2).map((part) => numberOf(part, "x")),
		);
	});

	test("cut into segments size each by its share, and brace from edge to edge", () => {
		const drawing = draw({
			bars: [
				{
					segments: [{ size: 1 }, { size: 2 }, { size: 3 }],
					braces: [{ from: 1, to: 3, label: "B" }],
				},
			],
		});
		const [segments = []] = outlinesOf(drawing);
		const brace = shapesOf(drawing, "path", "mt-pic-line").find((path) =>
			path.attributes.d?.includes("Q"),
		);
		const xs = pointsOfPath(brace?.attributes.d ?? "").map((point) => point.x);
		const edge = (at: number) =>
			numberOf(segments[0] ?? textOf(drawing, "B"), "x") +
			lengthOf(segments.slice(0, at));

		expect(segments.map((segment) => numberOf(segment, "width"))).toEqual(
			[1, 2, 3].map((size) =>
				expect.closeTo((size / 6) * lengthOf(segments), 0),
			),
		);
		expect(Math.min(...xs)).toBeCloseTo(edge(1) + 2, 0);
		expect(Math.max(...xs)).toBeCloseTo(edge(3) - 2, 0);
		expect(numberOf(textOf(drawing, "B"), "x")).toBeCloseTo(
			(edge(1) + edge(3)) / 2,
			0,
		);
	});

	test("write a segment's label in it where it fits, and under the bar where not", () => {
		const drawing = draw({
			bars: [
				{
					segments: [
						{ size: 1, label: "WWWWW" },
						{ size: 59, label: "B" },
					],
				},
			],
		});
		const [bar = []] = outlinesOf(drawing);
		const middle = numberOf(bar[0] ?? textOf(drawing, "B"), "y") + 14;

		expect(numberOf(textOf(drawing, "B"), "y")).toBeCloseTo(middle, 0);
		expect(numberOf(textOf(drawing, "WWWWW"), "y")).toBeGreaterThan(
			middle + 14,
		);
	});

	test("write a bar's label before it, its value after it and its length under it", () => {
		const drawing = draw({
			bars: [{ label: "A", parts: 3, value: "?", span: "12" }],
		});
		const [parts = []] = outlinesOf(drawing);
		const start = numberOf(parts[0] ?? textOf(drawing, "A"), "x");

		expect(numberOf(textOf(drawing, "A"), "x")).toBeLessThan(start);
		expect(numberOf(textOf(drawing, "?"), "x")).toBeGreaterThan(
			start + lengthOf(parts),
		);
		expect(numberOf(textOf(drawing, "12"), "y")).toBeGreaterThan(
			numberOf(textOf(drawing, "A"), "y"),
		);
	});

	test("write their notes under them all, in order, from the start of the line", () => {
		const drawing = draw({
			bars: [{ parts: 1 }],
			notes: ["A + B = 56", "A = ?"],
		});
		const notes = textsOf(drawing);

		expect(notes.map((note) => note.text)).toEqual(["A + B = 56", "A = ?"]);
		expect(notes.map((note) => note.attributes["text-anchor"])).toEqual([
			"start",
			"start",
		]);
		expect(numberOf(notes[1] ?? textOf(drawing, "A = ?"), "y")).toBeGreaterThan(
			numberOf(notes[0] ?? textOf(drawing, "A = ?"), "y"),
		);
	});

	test("write a note's hyphen as the sign of mathematics, as a label's is", () => {
		const drawing = draw({
			bars: [{ label: "-3", parts: 1 }],
			notes: ["A = -3", "B - 2 = A"],
		});

		expect(textsOf(drawing).map((text) => text.text)).toEqual([
			"−3",
			"A = −3",
			"B − 2 = A",
		]);
	});

	test("open a brace over a part too narrow for its curl under the bar, never over it", () => {
		const drawing = draw({
			bars: [
				{ length: 60 },
				{ length: 1, parts: 12, braces: [{ from: 0, to: 1, label: "A" }] },
			],
		});
		const brace = shapesOf(drawing, "path", "mt-pic-line").find((path) =>
			path.attributes.d?.includes("Q"),
		);
		const [top = { y: 0 }, ...rest] = pointsOfPath(brace?.attributes.d ?? "");

		for (const point of rest) {
			expect(point.y).toBeGreaterThanOrEqual(top.y);
		}
	});
});
