import { describe, expect, test } from "vitest";
import { drawn, numberOf, shapesOf, textOf } from "./testing/svg";

describe("a grid", () => {
	const drawing = drawn({
		kind: "grid",
		rows: ["A", "B", "C"],
		cols: ["1", "2", "3"],
		filled: ["A1", "B2"],
		marks: { C3: "?" },
	});

	test("names its rows to its left and its columns over it", () => {
		const [first] = shapesOf(drawing, "rect", "mt-pic-fill");
		const left = numberOf(first ?? textOf(drawing, "A"), "x");
		const top = numberOf(first ?? textOf(drawing, "A"), "y");

		expect(numberOf(textOf(drawing, "A"), "x")).toBeLessThan(left);
		expect(numberOf(textOf(drawing, "B"), "y")).toBeGreaterThan(
			numberOf(textOf(drawing, "A"), "y"),
		);
		expect(numberOf(textOf(drawing, "1"), "y")).toBeLessThan(top);
		expect(numberOf(textOf(drawing, "2"), "x")).toBeGreaterThan(
			numberOf(textOf(drawing, "1"), "x"),
		);
	});

	test("fills the cells it names, each under its row's and its column's names", () => {
		const filled = shapesOf(drawing, "rect", "mt-pic-fill").map((cell) => ({
			x: numberOf(cell, "x") + numberOf(cell, "width") / 2,
			y: numberOf(cell, "y") + numberOf(cell, "height") / 2,
		}));

		expect(filled.map((cell) => Math.round(cell.x))).toEqual(
			["1", "2"].map((name) =>
				Math.round(numberOf(textOf(drawing, name), "x")),
			),
		);
		expect(filled.map((cell) => Math.round(cell.y))).toEqual(
			["A", "B"].map((name) =>
				Math.round(numberOf(textOf(drawing, name), "y")),
			),
		);
	});

	test("writes a mark in the middle of its cell", () => {
		expect(numberOf(textOf(drawing, "?"), "x")).toBeCloseTo(
			numberOf(textOf(drawing, "3"), "x"),
			1,
		);
		expect(numberOf(textOf(drawing, "?"), "y")).toBeCloseTo(
			numberOf(textOf(drawing, "C"), "y"),
			1,
		);
	});
});
