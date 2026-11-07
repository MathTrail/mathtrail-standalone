import { describe, expect, test } from "vitest";
import type { ShapeRow } from "./model";
import {
	boxOfShape,
	type Drawing,
	drawn,
	numberOf,
	shapesOf,
	textsOf,
} from "./testing/svg";

// draw is rows of shapes as the card draws them.
const draw = (rows: ShapeRow[]) => drawn({ kind: "shapes", rows });

// outlinesOf are the outlines a drawing draws its shapes with, in its order.
function outlinesOf(drawing: Drawing) {
	return drawing.shapes.filter(
		(shape) => shape.attributes.class === "mt-pic-line",
	);
}

describe("rows of shapes", () => {
	test("draw the same shape alike wherever it stands", () => {
		const drawing = draw([
			{ shapes: ["circle", "triangle"], total: "8" },
			{ shapes: ["triangle", "triangle", "circle"], total: "8" },
		]);
		const sizeOf = (tag: string) =>
			shapesOf(drawing, tag, "mt-pic-line").map((one) => {
				const box = boxOfShape(one);
				return box && [box.right - box.left, box.bottom - box.top];
			});

		expect(sizeOf("circle")).toHaveLength(2);
		expect(new Set(sizeOf("circle").map(String)).size).toBe(1);
		expect(sizeOf("path")).toHaveLength(3);
		expect(new Set(sizeOf("path").map(String)).size).toBe(1);
	});

	test("draw a square and a diamond apart from a triangle", () => {
		const drawing = draw([
			{ shapes: ["square", "diamond", "triangle"], total: "?" },
		]);

		expect(shapesOf(drawing, "rect", "mt-pic-line")).toHaveLength(1);
		expect(shapesOf(drawing, "path", "mt-pic-line")).toHaveLength(2);
	});

	test("write each row's total after an equals sign, lined up after the longest row", () => {
		const drawing = draw([
			{ shapes: ["circle"], total: "4" },
			{ shapes: ["circle", "circle", "square"], total: "?" },
		]);
		const totals = textsOf(drawing);
		const right = Math.max(
			...outlinesOf(drawing).map(
				(one) =>
					numberOf(one, "cx") + numberOf(one, "r") ||
					numberOf(one, "x") + numberOf(one, "width"),
			),
		);

		expect(totals.map((one) => one.text)).toEqual(["= 4", "= ?"]);
		expect(new Set(totals.map((one) => one.attributes.x)).size).toBe(1);
		for (const total of totals) {
			expect(numberOf(total, "x")).toBeGreaterThan(right);
		}
	});
});
