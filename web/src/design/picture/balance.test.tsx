import { describe, expect, test } from "vitest";
import { drawn, numberOf, shapesOf, textOf } from "./testing/svg";

describe("a balance", () => {
	test("stands each weight on its pan, with its label", () => {
		const drawing = drawn({ kind: "balance", left: ["?"], right: ["1", "3"] });
		const weights = shapesOf(drawing, "rect", "mt-pic-ring");

		expect(weights).toHaveLength(3);
		expect(numberOf(textOf(drawing, "?"), "x")).toBeLessThan(
			numberOf(textOf(drawing, "1"), "x"),
		);
		expect(numberOf(textOf(drawing, "1"), "x")).toBeLessThan(
			numberOf(textOf(drawing, "3"), "x"),
		);
	});

	test("stands four wide weights two to a row", () => {
		const drawing = drawn({
			kind: "balance",
			left: ["99999", "99998", "99997", "99996"],
			right: [],
		});
		const rows = new Set(
			shapesOf(drawing, "rect", "mt-pic-ring").map((weight) =>
				numberOf(weight, "y"),
			),
		);

		expect(rows.size).toBe(2);
	});

	test("keeps its beam level, whatever stands on its pans", () => {
		const drawing = drawn({
			kind: "balance",
			left: ["1", "2", "3"],
			right: [],
		});
		const beam = shapesOf(drawing, "path", "mt-pic-line").find(
			(path) => path.attributes["stroke-width"] === "3",
		);

		expect(beam?.attributes.d).toMatch(/^M[\d.]+ [\d.]+H[\d.]+$/);
	});
});
