import { describe, expect, test } from "vitest";
import { drawn, numberOf, shapesOf, textsOf } from "./testing/svg";

describe("containers", () => {
	const drawing = drawn({
		kind: "containers",
		items: [
			{ capacity: 8, amount: 8 },
			{ capacity: 5, amount: 0, label: "B" },
			{ capacity: 3, amount: 1 },
		],
	});

	test("write what each holds and what it can hold, and the labels over them", () => {
		expect(textsOf(drawing).map((text) => text.text)).toEqual([
			"B",
			"8",
			"8",
			"0",
			"5",
			"1",
			"3",
		]);
	});

	test("draw each as tall as it can hold, to one scale, with its water as deep as it holds", () => {
		const walls = shapesOf(drawing, "path", "mt-pic-line").filter(
			(path) => path.attributes["stroke-width"] === "2.5",
		);
		const heights = walls.map((path) => {
			const [, top = 0, bottom = 0] =
				(path.attributes.d ?? "")
					.match(/^M[\d.]+ ([\d.]+)V([\d.]+)/)
					?.map(Number) ?? [];
			return bottom - top;
		});
		const water = shapesOf(drawing, "rect", "mt-pic-fill").map(
			(rect) => numberOf(rect, "height") + 1.25,
		);

		expect((heights[1] ?? 0) / (heights[0] ?? 1)).toBeCloseTo(5 / 8, 2);
		expect((heights[2] ?? 0) / (heights[0] ?? 1)).toBeCloseTo(3 / 8, 2);
		expect(water).toHaveLength(2);
		expect((water[1] ?? 0) / (water[0] ?? 1)).toBeCloseTo(1 / 8, 2);
	});

	test("mark every unit on a container's wall", () => {
		const marks = shapesOf(drawing, "path", "mt-pic-line").filter(
			(path) => path.attributes["stroke-width"] === "1",
		);

		expect(marks.map((path) => path.attributes.d?.match(/M/g)?.length)).toEqual(
			[7, 4, 2],
		);
	});
});
