import { describe, expect, test } from "vitest";
import {
	drawn,
	numberOf,
	pointsOfPath,
	shapesOf,
	textOf,
	textsOf,
} from "./testing/svg";

describe("a ring", () => {
	test("numbers its places from 1 at the top on round clockwise", () => {
		const drawing = drawn({ kind: "ring", count: 4 });
		const middle = numberOf(drawing.shapes[0] ?? textOf(drawing, "1"), "cx");
		const at = (words: string) => {
			const text = textOf(drawing, words);
			return {
				x: Math.round(numberOf(text, "x") - middle),
				y: Math.round(numberOf(text, "y") - middle),
			};
		};

		expect(textsOf(drawing).map((text) => text.text)).toEqual([
			"1",
			"2",
			"3",
			"4",
		]);
		expect(at("1").x).toBe(0);
		expect(at("1").y).toBeLessThan(0);
		expect(at("2").x).toBeGreaterThan(0);
		expect(at("3").y).toBeGreaterThan(0);
		expect(at("4").x).toBeLessThan(0);
	});

	test("points the way round between every two places", () => {
		const drawing = drawn({ kind: "ring", count: 7 });
		const [arrows] = shapesOf(drawing, "path", "mt-pic-ink");

		expect(arrows?.attributes.d?.match(/M/g)).toHaveLength(7);
	});

	test("shades the place it starts at, and points to it from its label at the middle", () => {
		const drawing = drawn({
			kind: "ring",
			count: 12,
			start: { at: 4, label: "S" },
		});
		const middle = numberOf(drawing.shapes[0] ?? textOf(drawing, "S"), "cx");
		const shaded = shapesOf(drawing, "circle", "mt-pic-fill");
		const tips = shapesOf(drawing, "path", "mt-pic-ink").map(
			(path) => pointsOfPath(path.attributes.d ?? "")[0],
		);

		expect(shaded).toHaveLength(1);
		expect(numberOf(shaded[0] ?? textOf(drawing, "S"), "cy")).toBeCloseTo(
			middle,
			0,
		);
		expect(numberOf(textOf(drawing, "S"), "x")).toBeCloseTo(middle, 1);
		expect(tips.at(-1)?.y).toBeCloseTo(middle, 0);
		expect(tips.at(-1)?.x).toBeGreaterThan(middle);
	});

	test("with a start of no label shades it and draws no arrow to it", () => {
		const drawing = drawn({ kind: "ring", count: 12, start: { at: 4 } });

		expect(shapesOf(drawing, "circle", "mt-pic-fill")).toHaveLength(1);
		expect(shapesOf(drawing, "path", "mt-pic-ink")).toHaveLength(1);
	});
});
