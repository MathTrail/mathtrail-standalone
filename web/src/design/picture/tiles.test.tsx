import { describe, expect, test } from "vitest";
import type { Tiles } from "./model";
import {
	type Drawing,
	drawn,
	numberOf,
	pointsOfPath,
	shapesOf,
	textOf,
	textsOf,
} from "./testing/svg";

// draw is a picture of tiles as the card draws it.
const draw = (tiles: Omit<Tiles, "kind">) => drawn({ kind: "tiles", ...tiles });

// blank are as many tiles of so many sides as given, every triangle blank.
const blank = (count: number, sides: number) =>
	Array.from({ length: count }, () => Array<string>(sides).fill(""));

// middlesOf are the middles of a drawing's tiles, where their spokes meet.
function middlesOf(drawing: Drawing): { x: number; y: number }[] {
	return shapesOf(drawing, "path", "mt-pic-line")
		.filter((one) => numberOf(one, "stroke-width") === 1)
		.map((one) => pointsOfPath(one.attributes.d ?? "")[0] ?? { x: 0, y: 0 });
}

// apart is how far apart two points stand.
const apart = (
	one: { x: number; y: number },
	other: { x: number; y: number },
) => Math.hypot(one.x - other.x, one.y - other.y);

describe("tiles", () => {
	test("draw each tile's outline and its spokes, and a label only where a triangle has one", () => {
		const drawing = draw({
			sides: 4,
			layout: "row",
			tiles: [
				["1", "", "", "2"],
				["", "?", "", ""],
			],
		});

		expect(shapesOf(drawing, "path", "mt-pic-line")).toHaveLength(4);
		expect(textsOf(drawing).map((one) => one.text)).toEqual(["1", "2", "?"]);
	});

	test("lay a row straight on, its tiles evenly apart", () => {
		const middles = middlesOf(
			draw({ sides: 6, layout: "row", tiles: blank(4, 6) }),
		);
		const steps = middles
			.slice(1)
			.map((one, at) => apart(one, middles[at] ?? one));

		expect(new Set(middles.map((one) => one.y)).size).toBe(1);
		for (const step of steps) {
			expect(step).toBeCloseTo(steps[0] ?? 0, 0);
		}
	});

	test.each([
		["ten pentagons", 5, 10],
		["four squares joined at their next side", 4, 4],
		["six triangles", 3, 6],
	])(
		"close a ring of %s, its last tile beside its first",
		(_, sides, count) => {
			const middles = middlesOf(
				draw({
					sides,
					layout: "ring",
					...(sides === 4 ? { joint: 1 } : {}),
					tiles: blank(count, sides),
				}),
			);
			const first = middles[0] ?? { x: 0, y: 0 };
			const last = middles.at(-1) ?? first;

			expect(apart(first, last)).toBeCloseTo(
				apart(first, middles[1] ?? first),
				0,
			);
		},
	);

	test("set a tile's first label in the triangle facing the tile before it", () => {
		const drawing = draw({
			sides: 4,
			layout: "row",
			tiles: [
				["", "", "", ""],
				["A", "B", "C", "D"],
			],
		});
		const [first] = middlesOf(drawing);
		const from = (words: string) => {
			const text = textOf(drawing, words);
			return apart(first ?? { x: 0, y: 0 }, {
				x: numberOf(text, "x"),
				y: numberOf(text, "y"),
			});
		};

		expect(from("A")).toBeLessThan(Math.min(from("B"), from("C"), from("D")));
	});

	test("write their labels clockwise", () => {
		const drawing = draw({
			sides: 4,
			layout: "row",
			tiles: [
				["L", "T", "R", "B"],
				["", "", "", ""],
			],
		});
		const at = (words: string) => textOf(drawing, words);

		expect(numberOf(at("L"), "x")).toBeLessThan(numberOf(at("R"), "x"));
		expect(numberOf(at("T"), "y")).toBeLessThan(numberOf(at("B"), "y"));
	});
});
