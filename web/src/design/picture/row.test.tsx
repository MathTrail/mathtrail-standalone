import { describe, expect, test } from "vitest";
import { drawn, numberOf, shapesOf, textOf, textsOf } from "./testing/svg";

describe("a row", () => {
	test("draws each item with its mark: a dot unless it names a ring or a square", () => {
		const drawing = drawn({
			kind: "row",
			items: [{}, { mark: "ring" }, { mark: "square" }, { mark: "dot" }],
		});

		expect(shapesOf(drawing, "circle", "mt-pic-ink")).toHaveLength(2);
		expect(shapesOf(drawing, "circle", "mt-pic-ring")).toHaveLength(1);
		expect(shapesOf(drawing, "rect", "mt-pic-ink")).toHaveLength(1);
	});

	test("cuts itself short where a skip stands, with a broken line", () => {
		const drawing = drawn({ kind: "row", items: [{}, {}, { skip: true }, {}] });
		const lines = shapesOf(drawing, "path", "mt-pic-line");

		expect(shapesOf(drawing, "circle", "mt-pic-ink")).toHaveLength(3);
		expect(lines.map((line) => line.attributes["stroke-dasharray"])).toEqual([
			undefined,
			"2 5",
		]);
		expect(textsOf(drawing)).toHaveLength(0);
	});

	test("with no line, draws a skip as an ellipsis", () => {
		const drawing = drawn({
			kind: "row",
			items: [{}, { skip: true }, {}],
			line: false,
		});

		expect(shapesOf(drawing, "path", "mt-pic-line")).toHaveLength(0);
		expect(textsOf(drawing).map((text) => text.text)).toEqual(["…"]);
	});

	test("labels every gap it draws, and not the stretch it cuts short", () => {
		const drawing = drawn({
			kind: "row",
			items: [{}, {}, {}, { skip: true }, {}, {}],
			gaps: "3",
		});

		expect(textsOf(drawing).map((text) => text.text)).toEqual(["3", "3", "3"]);
	});

	test("writes its labels over its items and under them, and its whole length under all", () => {
		const drawing = drawn({
			kind: "row",
			items: [{ label: "A", below: "1" }, { label: "B" }],
			span: "?",
		});
		const [mark] = shapesOf(drawing, "circle", "mt-pic-ink");
		const heightOf = (words: string) => numberOf(textOf(drawing, words), "y");

		expect(heightOf("A")).toBeLessThan(
			numberOf(mark ?? textOf(drawing, "A"), "cy"),
		);
		expect(heightOf("1")).toBeGreaterThan(
			numberOf(mark ?? textOf(drawing, "A"), "cy"),
		);
		expect(heightOf("?")).toBeGreaterThan(heightOf("1"));
	});

	test("drawn twice stands as the two sides of a path, one under the other", () => {
		const once = drawn({ kind: "row", items: [{}, {}, {}] });
		const twice = drawn({ kind: "row", items: [{}, {}, {}], copies: 2 });
		const rowsOf = (drawing: typeof once) =>
			new Set(
				shapesOf(drawing, "circle", "mt-pic-ink").map((dot) =>
					numberOf(dot, "cy"),
				),
			).size;

		expect(rowsOf(once)).toBe(1);
		expect(rowsOf(twice)).toBe(2);
		expect(shapesOf(twice, "circle", "mt-pic-ink")).toHaveLength(6);
	});

	test("spreads its items as far apart as the card allows, and no further", () => {
		const xsOf = (count: number) =>
			shapesOf(
				drawn({ kind: "row", items: Array(count).fill({}) }),
				"circle",
				"mt-pic-ink",
			).map((dot) => numberOf(dot, "cx"));
		const steps = (xs: number[]) =>
			xs.slice(1).map((x, at) => x - (xs[at] ?? 0));

		expect(new Set(steps(xsOf(3)))).toEqual(new Set([44]));
		expect(Math.max(...steps(xsOf(12)))).toBeLessThan(44);
	});
});
