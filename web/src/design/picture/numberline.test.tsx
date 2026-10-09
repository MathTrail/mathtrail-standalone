import { describe, expect, test } from "vitest";
import type { NumberLine } from "./model";
import { drawingOf, numberOf, textsOf } from "./testing/svg";

// numbersOf are the numbers a line writes under its ticks, from the left, and
// the size it writes them at.
function numbersOf(line: NumberLine): { numbers: string[]; sizes: number[] } {
	const texts = textsOf(drawingOf(line) ?? { root: {}, shapes: [] }).filter(
		(text) => !text.attributes.class?.includes("mt-pic-strong"),
	);
	return {
		numbers: texts.map((text) => text.text),
		sizes: [...new Set(texts.map((text) => numberOf(text, "font-size")))],
	};
}

describe("a number line", () => {
	test("numbers every tick where the numbers fit", () => {
		expect(numbersOf({ kind: "number_line", from: 0, to: 6 })).toEqual({
			numbers: ["0", "1", "2", "3", "4", "5", "6"],
			sizes: [12],
		});
	});

	test("numbers every tick at the smaller size where only that fits them", () => {
		expect(numbersOf({ kind: "number_line", from: 100, to: 109 })).toEqual({
			numbers: [
				"100",
				"101",
				"102",
				"103",
				"104",
				"105",
				"106",
				"107",
				"108",
				"109",
			],
			sizes: [11],
		});
	});

	test("numbers every second tick where every one crowds, and always its ends", () => {
		expect(numbersOf({ kind: "number_line", from: 0, to: 15 }).numbers).toEqual(
			["0", "2", "4", "6", "8", "10", "12", "15"],
		);
	});

	test("always numbers the ticks its marks stand on, leaving out the numbers that would crowd them", () => {
		expect(
			numbersOf({
				kind: "number_line",
				from: 0,
				to: 15,
				marks: [{ at: 11, label: "P" }],
			}).numbers,
		).toEqual(["0", "2", "4", "6", "8", "11", "15"]);
	});

	test("numbers every fifth tick of a line of long numbers, and its marks and ends however they crowd", () => {
		const { numbers } = numbersOf({
			kind: "number_line",
			from: -9999,
			to: 99996,
			step: 7333,
			marks: [{ at: 63331 }, { at: 70664 }],
		});

		expect(numbers).toEqual(["−9999", "26666", "63331", "70664", "99996"]);
	});

	test("writes a number below zero with the sign of mathematics", () => {
		expect(numbersOf({ kind: "number_line", from: -3, to: 3 }).numbers[0]).toBe(
			"−3",
		);
	});

	test("points to every mark from over its tick, its label over the pointer", () => {
		const drawing = drawingOf({
			kind: "number_line",
			from: 0,
			to: 6,
			marks: [{ at: 0, label: "P" }, { at: 4 }],
		});
		const pointers = drawing?.shapes.find(
			(shape) =>
				shape.tag === "path" && shape.attributes.class === "mt-pic-fill",
		);
		const labels = textsOf(drawing ?? { root: {}, shapes: [] }).filter((text) =>
			text.attributes.class?.includes("mt-pic-strong"),
		);

		expect(pointers?.attributes.d?.match(/M/g)).toHaveLength(2);
		expect(labels.map((label) => label.text)).toEqual(["P"]);
	});
});
