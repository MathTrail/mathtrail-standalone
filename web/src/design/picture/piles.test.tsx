import { describe, expect, test } from "vitest";
import type { Pile, Piles } from "./model";
import {
	type Drawing,
	drawn,
	numberOf,
	shapesOf,
	textOf,
	textsOf,
} from "./testing/svg";

// draw is a picture of piles as the card draws it.
const draw = (piles: Pile[], how: Omit<Piles, "kind" | "piles"> = {}) =>
	drawn({ kind: "piles", piles, ...how });

// countersOf are the counters a drawing draws: dots and squares, dark and
// light, by where their middles stand.
function countersOf(drawing: Drawing): { x: number; y: number }[] {
	return drawing.shapes
		.filter(
			(shape) =>
				(shape.tag === "circle" && numberOf(shape, "r") > 2) ||
				(shape.tag === "rect" && shape.attributes.class !== "mt-pic-line"),
		)
		.map((shape) =>
			shape.tag === "circle"
				? { x: numberOf(shape, "cx"), y: numberOf(shape, "cy") }
				: {
						x: numberOf(shape, "x") + numberOf(shape, "width") / 2,
						y: numberOf(shape, "y") + numberOf(shape, "height") / 2,
					},
		);
}

// dotsOf are how many small dots a drawing draws: three to every ellipsis.
function dotsOf(drawing: Drawing): number {
	return drawing.shapes.filter(
		(shape) => shape.tag === "circle" && numberOf(shape, "r") < 2,
	).length;
}

describe("piles", () => {
	test("never write how many counters a pile has", () => {
		const drawing = draw([
			{ label: "A", count: 13 },
			{ count: 40, shown: 6 },
		]);

		expect(textsOf(drawing).map((text) => text.text)).toEqual(["A"]);
	});

	test("write a pile's value after it, and its label before it", () => {
		const drawing = draw([{ label: "A", count: 3, value: "13" }]);
		const xs = countersOf(drawing).map((counter) => counter.x);

		expect(textsOf(drawing).map((text) => text.text)).toEqual(["A", "13"]);
		expect(numberOf(textOf(drawing, "A"), "x")).toBeLessThan(Math.min(...xs));
		expect(numberOf(textOf(drawing, "13"), "x")).toBeGreaterThan(
			Math.max(...xs),
		);
	});

	test("draw every counter of a pile not cut short", () => {
		expect(countersOf(draw([{ count: 12 }]))).toHaveLength(12);
	});

	test("cut a pile short to the counters it shows, half on either side of an ellipsis", () => {
		const drawing = draw([{ count: 40, shown: 7 }]);
		const xs = countersOf(drawing)
			.map((counter) => counter.x)
			.sort((one, other) => one - other);

		expect(xs).toHaveLength(7);
		expect(dotsOf(drawing)).toBe(3);
		expect((xs[4] ?? 0) - (xs[3] ?? 0)).toBeGreaterThan(
			(xs[1] ?? 0) - (xs[0] ?? 0),
		);
	});

	test("draw a heap of no stated size as a counter, an ellipsis and a counter", () => {
		const drawing = draw([{ label: "A" }]);

		expect(countersOf(drawing)).toHaveLength(2);
		expect(dotsOf(drawing)).toBe(3);
	});

	test("draw a pile of none as an empty place, outlined in dashes", () => {
		const drawing = draw([{ count: 0 }]);
		const place = shapesOf(drawing, "rect", "mt-pic-line");

		expect(countersOf(drawing)).toHaveLength(0);
		expect(place.map((one) => one.attributes["stroke-dasharray"])).toEqual([
			"3 3",
		]);
	});

	test("leave a gap after every group", () => {
		const xs = countersOf(draw([{ count: 10, group: 5 }])).map(
			(counter) => counter.x,
		);
		const steps = xs.slice(1).map((x, at) => x - (xs[at] ?? 0));

		expect(steps[4]).toBeGreaterThan(steps[0] ?? 0);
		expect(new Set(steps.filter((_, at) => at !== 4)).size).toBe(1);
	});

	test("draw counters of the shape and fill they name", () => {
		const drawing = draw([
			{ count: 1 },
			{ count: 1, fill: "light" },
			{ count: 1, shape: "square" },
			{ count: 1, shape: "square", fill: "light" },
		]);

		expect(shapesOf(drawing, "circle", "mt-pic-ink")).toHaveLength(1);
		expect(shapesOf(drawing, "circle", "mt-pic-ring")).toHaveLength(1);
		expect(shapesOf(drawing, "rect", "mt-pic-ink")).toHaveLength(1);
		expect(shapesOf(drawing, "rect", "mt-pic-ring")).toHaveLength(1);
	});

	test("stand one under another, or side by side", () => {
		const under = countersOf(draw([{ count: 1 }, { count: 1 }]));
		const across = countersOf(
			draw([{ count: 1 }, { count: 1 }], { across: true }),
		);

		expect(under[0]?.x).toBe(under[1]?.x);
		expect(under[1]?.y).toBeGreaterThan(under[0]?.y ?? 0);
		expect(across[0]?.y).toBe(across[1]?.y);
		expect(across[1]?.x).toBeGreaterThan(across[0]?.x ?? 0);
	});

	test("stand in a box all together, or each in a box of its own", () => {
		const boxes = (how: Omit<Piles, "kind" | "piles">, piles: Pile[]) =>
			shapesOf(draw(piles, how), "rect", "mt-pic-line").length;

		expect(boxes({}, [{ count: 2 }, { count: 2 }])).toBe(0);
		expect(boxes({ box: true }, [{ count: 2 }, { count: 2 }])).toBe(1);
		expect(boxes({}, [{ count: 2, boxed: true }, { count: 2 }])).toBe(1);
	});

	test("mark a skip among them with three dots and nothing else", () => {
		const drawing = draw([{ count: 1 }, { skip: true }, { count: 1 }]);

		expect(countersOf(drawing)).toHaveLength(2);
		expect(dotsOf(drawing)).toBe(3);
	});
});
