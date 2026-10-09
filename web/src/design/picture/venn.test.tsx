import { describe, expect, test } from "vitest";
import { drawn, numberOf, shapesOf, textOf } from "./testing/svg";

describe("two groups", () => {
	test("write each group's label and count over it, and the counts in both, in neither and in all", () => {
		const drawing = drawn({
			kind: "venn",
			total: "25",
			sets: [
				{ label: "F", count: "15" },
				{ label: "C", count: "12" },
			],
			both: "5",
			neither: "?",
		});
		const [first, second] = shapesOf(drawing, "circle", "mt-pic-line");
		const circleTop =
			numberOf(first ?? textOf(drawing, "5"), "cy") -
			numberOf(first ?? textOf(drawing, "5"), "r");

		expect(numberOf(textOf(drawing, "F 15"), "y")).toBeLessThan(circleTop);
		expect(numberOf(textOf(drawing, "F 15"), "x")).toBeLessThan(
			numberOf(first ?? textOf(drawing, "5"), "cx"),
		);
		expect(numberOf(textOf(drawing, "C 12"), "x")).toBeGreaterThan(
			numberOf(second ?? textOf(drawing, "5"), "cx"),
		);
		expect(numberOf(textOf(drawing, "5"), "x")).toBeCloseTo(
			(numberOf(first ?? textOf(drawing, "5"), "cx") +
				numberOf(second ?? textOf(drawing, "5"), "cx")) /
				2,
			1,
		);
		expect(numberOf(textOf(drawing, "?"), "y")).toBeGreaterThan(
			numberOf(first ?? textOf(drawing, "5"), "cy"),
		);
		expect(numberOf(textOf(drawing, "25"), "y")).toBeLessThan(
			numberOf(textOf(drawing, "F 15"), "y"),
		);
	});

	test("set the count in both on a plate where the overlap is too narrow for it", () => {
		const plated = (both: string) =>
			shapesOf(
				drawn({ kind: "venn", sets: [{ label: "A" }, { label: "B" }], both }),
				"rect",
				"mt-pic-plate",
			).length;

		expect(plated("5")).toBe(0);
		expect(plated("WWWWW")).toBe(1);
	});
});
