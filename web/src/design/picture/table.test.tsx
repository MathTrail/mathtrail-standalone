import { describe, expect, test } from "vitest";
import { drawn, numberOf, shapesOf, textOf, textsOf } from "./testing/svg";

describe("a table", () => {
	test("writes its header strong on a band, its names strong and its numbers and times plain", () => {
		const drawing = drawn({
			kind: "table",
			header: ["1", "2"],
			rows: [
				["A", "22:35"],
				["?", "15"],
			],
		});
		const strong = (words: string) =>
			textOf(drawing, words).attributes.class?.includes("mt-pic-strong");

		expect(shapesOf(drawing, "rect", "mt-pic-band")).toHaveLength(1);
		expect(["1", "2", "A", "?"].map(strong)).toEqual([true, true, true, true]);
		expect(["22:35", "15"].map(strong)).toEqual([false, false]);
	});

	test("writes nothing in an empty cell and an ellipsis where cells are left out", () => {
		const drawing = drawn({ kind: "table", rows: [["A", "", "…"]] });

		expect(textsOf(drawing).map((text) => text.text)).toEqual(["A", "…"]);
	});

	test("makes each column as wide as its widest cell, centred in it", () => {
		const drawing = drawn({
			kind: "table",
			rows: [
				["A", "22:35"],
				["B", "7:05"],
			],
		});
		const [a, b, late] = [
			textOf(drawing, "A"),
			textOf(drawing, "B"),
			textOf(drawing, "22:35"),
		];

		expect(numberOf(a, "x")).toBe(numberOf(b, "x"));
		expect(numberOf(textOf(drawing, "7:05"), "x")).toBe(numberOf(late, "x"));
		expect(numberOf(late, "x") - numberOf(a, "x")).toBeGreaterThan(
			numberOf(a, "x"),
		);
	});

	test.each([
		["a small table", [["A", "22:35", "6:10"]], 13],
		[
			"eight rows of six times",
			Array(8).fill(["22:35", "23:59", "10:05", "16:10", "17:05", "19:45"]),
			11,
		],
	])(
		"writes %s at the largest size at which it fits the card",
		(_, rows, size) => {
			const sizes = new Set(
				textsOf(drawn({ kind: "table", rows })).map((text) =>
					numberOf(text, "font-size"),
				),
			);

			expect([...sizes]).toEqual([size]);
		},
	);
});
