import { describe, expect, test } from "vitest";
import type { Stairs } from "./model";
import {
	drawn,
	numberOf,
	overlap,
	pointsOfPath,
	shapesOf,
	textBox,
	textOf,
	textsOf,
} from "./testing/svg";

// draw is a staircase as the card draws it.
const draw = (stairs: Omit<Stairs, "kind">) =>
	drawn({ kind: "stairs", ...stairs });

describe("a staircase", () => {
	test("numbers none of its steps: it writes only its marks", () => {
		const drawing = draw({ steps: 9, marks: [{ step: 4, label: "A" }] });

		expect(textsOf(drawing).map((one) => one.text)).toEqual(["A"]);
	});

	test("climbs to the right, a step at a time", () => {
		const drawing = draw({ steps: 5 });
		const outline = pointsOfPath(
			shapesOf(drawing, "path", "mt-pic-fill")[0]?.attributes.d ?? "",
		);
		const heights = [...new Set(outline.map((one) => one.y))];

		expect(heights).toHaveLength(6);
	});

	test("sets a mark above its step, higher for a higher step", () => {
		const drawing = draw({
			steps: 6,
			marks: [
				{ step: 1, label: "A" },
				{ step: 5, label: "B" },
			],
		});
		const a = textOf(drawing, "A");
		const b = textOf(drawing, "B");

		expect(numberOf(a, "x")).toBeLessThan(numberOf(b, "x"));
		expect(numberOf(a, "y")).toBeGreaterThan(numberOf(b, "y"));
	});

	test("raises a mark that would touch one set before it", () => {
		const drawing = draw({
			steps: 16,
			marks: [
				{ step: 7, label: "WWWWW" },
				{ step: 8, label: "MMMMM" },
			],
		});

		expect(
			overlap(
				textBox(textOf(drawing, "WWWWW")),
				textBox(textOf(drawing, "MMMMM")),
			),
		).toBe(false);
	});

	test("starts and ends a jump above the marks of its steps", () => {
		const drawing = draw({
			steps: 8,
			marks: [
				{ step: 0, label: "A" },
				{ step: 2, label: "B" },
			],
			jumps: [{ from: 0, to: 2 }],
		});
		const arc = shapesOf(drawing, "path", "mt-pic-line").find((one) =>
			one.attributes.d?.includes("Q"),
		);
		const [start, , landing] = pointsOfPath(arc?.attributes.d ?? "");

		expect(start?.y).toBeLessThan(textBox(textOf(drawing, "A")).top);
		expect(landing?.y).toBeLessThan(textBox(textOf(drawing, "B")).top);
	});

	test("draws a jump as an arc with an arrowhead where it lands", () => {
		const drawing = draw({ steps: 6, jumps: [{ from: 1, to: 4 }] });
		const arc = shapesOf(drawing, "path", "mt-pic-line").find((one) =>
			one.attributes.d?.includes("Q"),
		);
		const head = shapesOf(drawing, "path", "mt-pic-ink");
		const [, , landing] = pointsOfPath(arc?.attributes.d ?? "");

		expect(head).toHaveLength(1);
		expect(pointsOfPath(head[0]?.attributes.d ?? "")[0]).toEqual(landing);
	});
});
