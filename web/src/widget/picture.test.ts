import { readFileSync } from "node:fs";
import { join } from "node:path";
import { describe, expect, test } from "vitest";
import type { Picture } from "../design/picture/model";
import { kinds, paints } from "../design/picture/model";
import { extremes } from "../design/picture/testing/extremes";
import { pictureFormat, pictureLimits } from "./picture";
import { seeded } from "./testing/seeded";

// readPicture is a picture as the card's reading of the format takes it, or
// undefined for one it refuses: what the card's payload keeps of it.
function readPicture(value: unknown): Picture | undefined {
	const read = pictureFormat.safeParse(value);
	return read.success ? read.data : undefined;
}

// goSource is a file of the service's picture format, which the card's reading
// of a picture is held to.
function goSource(name: string): string {
	return readFileSync(
		join(import.meta.dirname, "../../../internal/domain/picture", name),
		"utf8",
	);
}

// examples are the example of each kind the content shows the model, by the
// file it is in.
const examples = Object.entries(
	import.meta.glob<{ picture: unknown }>("../../../content/pictures/*.json", {
		eager: true,
		import: "default",
	}),
).map(([file, example]) => [file, example.picture] as const);

// references are the pictures the reference tasks carry, by the task's id.
const references = Object.values(
	import.meta.glob<readonly { id: string; picture?: unknown }[]>(
		"../../../content/examples/*.json",
		{ eager: true, import: "default" },
	),
).flatMap((tasks) =>
	tasks.flatMap((task) =>
		task.picture === undefined ? [] : [[task.id, task.picture] as const],
	),
);

describe("the card's reading of a picture", () => {
	test("knows the kinds the service knows, in the order it lists them", () => {
		const listed = [
			...goSource("picture.go").matchAll(/^\t\w+\s+Kind = "(\w+)"$/gm),
		].map((match) => match[1]);

		expect(kinds).toEqual(listed);
	});

	test("knows the colours of the service's palette, in its order, and reads a word for each", () => {
		const listed = [
			...goSource("colors.go").matchAll(/^\t\w+\s+Paint = "(\w+)"$/gm),
		].map((match) => match[1]);
		const painted = pictureFormat.safeParse({
			kind: "flags",
			colors: Object.fromEntries(paints.map((paint) => [paint, paint])),
			groups: [
				{ flags: [["red", "yellow", "green"]] },
				{ flags: [["blue", "white", "black"]] },
			],
		});

		expect(paints).toEqual(listed);
		expect(painted.success).toBe(true);
	});

	test("holds a picture to the limits the service holds it to", () => {
		const named: Record<string, number> = {};
		for (const [, names = "", values = ""] of goSource("limits.go").matchAll(
			/^\t([\w, ]+?)\s*=\s*([-\d, ]+)$/gm,
		)) {
			const numbers = values.split(",").map(Number);
			names.split(",").forEach((name, at) => {
				named[name.trim()] = numbers[at] ?? Number.NaN;
			});
		}

		expect(named).toEqual(pictureLimits);
	});

	test.each(examples)("reads the example %s", (_, picture) => {
		expect(readPicture(picture)).toEqual(picture);
	});

	test("reads every picture the reference tasks carry", () => {
		expect(references).toHaveLength(95);
		for (const [id, picture] of references) {
			expect(readPicture(picture), id).toEqual(picture);
		}
	});

	test.each(extremes.map((one) => [one.name, one.picture]))(
		"reads a picture at its limits: %s",
		(_, picture) => {
			expect(readPicture(picture)).toEqual(picture);
		},
	);

	test("reads a member written as null or as an empty text as one left out", () => {
		expect(
			readPicture({
				kind: "clock",
				time: "4:30",
				hour_label: null,
				minute_label: "",
			}),
		).toEqual({ kind: "clock", time: "4:30" });
		expect(
			readPicture({
				kind: "number_line",
				from: 0,
				to: 6,
				step: null,
				marks: null,
			}),
		).toEqual({ kind: "number_line", from: 0, to: 6 });
		expect(
			readPicture({ kind: "row", items: [{}, { label: "" }], gaps: null }),
		).toEqual({ kind: "row", items: [{}, {}] });
	});

	test.each([
		["no object", "clock"],
		["a list", [{ kind: "clock", time: "4:30" }]],
		["no kind", { time: "4:30" }],
		["a kind the format does not have", { kind: "pie", parts: 3 }],
		[
			"a member its kind does not have",
			{ kind: "clock", time: "4:30", hands: 2 },
		],
		["a time past the day", { kind: "clock", time: "24:00" }],
		[
			"a time without its minutes in two digits",
			{ kind: "clock", time: "4:3" },
		],
		["a label too long", { kind: "clock", time: "4:30", hour_label: "ABCDEF" }],
		["a label of no text", { kind: "balance", left: [3], right: [] }],
		[
			"a table's row shorter than its header",
			{ kind: "table", header: ["A", "B"], rows: [["1"]] },
		],
		[
			"a table's row longer than its first",
			{ kind: "table", rows: [["1"], ["2", "3"]] },
		],
		[
			"a table of nine rows",
			{ kind: "table", rows: Array.from({ length: 9 }, () => ["1"]) },
		],
		[
			"a table's row of seven cells",
			{ kind: "table", rows: [Array(7).fill("1")] },
		],
		["a cell too long", { kind: "table", rows: [["ABCDEF"]] }],
		[
			"a line that ends where it starts",
			{ kind: "number_line", from: 3, to: 3 },
		],
		[
			"a line its step does not reach the end of",
			{ kind: "number_line", from: 0, to: 7, step: 2 },
		],
		["a line of seventeen ticks", { kind: "number_line", from: 0, to: 16 }],
		[
			"a line past its highest number",
			{ kind: "number_line", from: 99990, to: 100000, step: 5 },
		],
		[
			"a mark between two ticks",
			{ kind: "number_line", from: 0, to: 10, step: 2, marks: [{ at: 3 }] },
		],
		[
			"a mark past the line",
			{ kind: "number_line", from: 0, to: 6, marks: [{ at: 7 }] },
		],
		[
			"two marks on one tick",
			{ kind: "number_line", from: 0, to: 6, marks: [{ at: 2 }, { at: 2 }] },
		],
		["a row of one", { kind: "row", items: [{}] }],
		["a row of thirteen", { kind: "row", items: Array(13).fill({}) }],
		[
			"a row that starts with a skip",
			{ kind: "row", items: [{ skip: true }, {}, {}] },
		],
		[
			"a row that ends with a skip",
			{ kind: "row", items: [{}, {}, { skip: true }] },
		],
		[
			"two skips side by side",
			{ kind: "row", items: [{}, { skip: true }, { skip: true }, {}] },
		],
		[
			"a skip that holds a label",
			{ kind: "row", items: [{}, { skip: true, label: "A" }, {}] },
		],
		["a row drawn three times", { kind: "row", items: [{}, {}], copies: 3 }],
		[
			"a mark the format does not have",
			{ kind: "row", items: [{ mark: "star" }, {}] },
		],
		["a ring of two", { kind: "ring", count: 2 }],
		["a ring of twenty-five", { kind: "ring", count: 25 }],
		["a start past the ring", { kind: "ring", count: 12, start: { at: 13 } }],
		[
			"a grid of nine rows",
			{
				kind: "grid",
				rows: Array.from({ length: 9 }, (_, at) => String(at + 1)),
				cols: ["A"],
			},
		],
		[
			"a grid naming a row twice",
			{ kind: "grid", rows: ["A", "A"], cols: ["1"] },
		],
		[
			"a grid naming two cells alike",
			{ kind: "grid", rows: ["1", "12"], cols: ["22", "2"] },
		],
		[
			"a filled cell outside the grid",
			{ kind: "grid", rows: ["A"], cols: ["1"], filled: ["B1"] },
		],
		[
			"a cell filled twice",
			{ kind: "grid", rows: ["A"], cols: ["1"], filled: ["A1", "A1"] },
		],
		[
			"a mark outside the grid",
			{ kind: "grid", rows: ["A"], cols: ["1"], marks: { B1: "?" } },
		],
		["five bars", { kind: "bars", bars: Array(5).fill({ parts: 1 }) }],
		["a bar of thirteen parts", { kind: "bars", bars: [{ parts: 13 }] }],
		[
			"a bar of parts and segments",
			{
				kind: "bars",
				bars: [{ parts: 2, segments: [{ size: 1 }, { size: 1 }] }],
			},
		],
		[
			"more shaded than parts",
			{ kind: "bars", bars: [{ parts: 3, shaded: 4 }] },
		],
		[
			"segments shaded",
			{
				kind: "bars",
				bars: [{ segments: [{ size: 1 }, { size: 2 }], shaded: 1 }],
			},
		],
		[
			"a brace that ends where it starts",
			{
				kind: "bars",
				bars: [{ parts: 3, braces: [{ from: 1, to: 1, label: "A" }] }],
			},
		],
		[
			"a brace past its bar",
			{
				kind: "bars",
				bars: [{ parts: 3, braces: [{ from: 1, to: 4, label: "A" }] }],
			},
		],
		[
			"five braces",
			{
				kind: "bars",
				bars: [
					{ parts: 6, braces: Array(5).fill({ from: 0, to: 1, label: "A" }) },
				],
			},
		],
		["a bar too long", { kind: "bars", bars: [{ parts: 2, length: 61 }] }],
		[
			"four notes",
			{ kind: "bars", bars: [{ parts: 1 }], notes: Array(4).fill("A = 1") },
		],
		[
			"a note too long",
			{
				kind: "bars",
				bars: [{ parts: 1 }],
				notes: ["A + B + C + D + E + F = 99"],
			},
		],
		["one group", { kind: "venn", sets: [{ label: "A" }] }],
		["three groups", { kind: "venn", sets: Array(3).fill({ label: "A" }) }],
		["five on a pan", { kind: "balance", left: Array(5).fill("1"), right: [] }],
		["a balance with no right pan", { kind: "balance", left: ["1"] }],
		[
			"five containers",
			{ kind: "containers", items: Array(5).fill({ capacity: 1, amount: 0 }) },
		],
		[
			"a container too large",
			{ kind: "containers", items: [{ capacity: 21, amount: 0 }] },
		],
		[
			"a container that holds more than it can",
			{ kind: "containers", items: [{ capacity: 3, amount: 4 }] },
		],
		["seven piles", { kind: "piles", piles: Array(7).fill({ count: 1 }) }],
		["a pile of forty-one", { kind: "piles", piles: [{ count: 41 }] }],
		[
			"a pile cut short with no count",
			{ kind: "piles", piles: [{ shown: 4 }] },
		],
		[
			"a pile that shows all it has",
			{ kind: "piles", piles: [{ count: 6, shown: 6 }] },
		],
		[
			"a pile that shows one",
			{ kind: "piles", piles: [{ count: 6, shown: 1 }] },
		],
		["a group of none", { kind: "piles", piles: [{ count: 6, group: 0 }] }],
		[
			"a fill the format does not have",
			{ kind: "piles", piles: [{ count: 6, fill: "red" }] },
		],
		[
			"a skip that holds a count",
			{
				kind: "piles",
				piles: [{ count: 1 }, { skip: true, count: 3 }, { count: 1 }],
			},
		],
		[
			"piles that start with a skip",
			{ kind: "piles", piles: [{ skip: true }, { count: 1 }] },
		],
		["a week of a day before Monday", { kind: "calendar", first: 0, days: 30 }],
		["a week of a day past Sunday", { kind: "calendar", first: 8, days: 30 }],
		["a month of twenty-seven days", { kind: "calendar", first: 1, days: 27 }],
		["a month of thirty-two days", { kind: "calendar", first: 1, days: 32 }],
		[
			"a mark past the month",
			{ kind: "calendar", first: 1, days: 30, marks: { "31": "A" } },
		],
		[
			"a day written with a nought",
			{ kind: "calendar", first: 1, days: 30, marks: { "01": "A" } },
		],
		[
			"a week that starts on a Saturday",
			{ kind: "calendar", first: 1, days: 30, week_starts: "saturday" },
		],
		[
			"colours on a kind that paints none",
			{ kind: "clock", time: "4:30", colors: { red: "red" } },
		],
		[
			"a colour the palette does not have",
			{
				kind: "flags",
				colors: { purple: "purple" },
				groups: [{ flags: [["?", "?"]] }],
			},
		],
		[
			"a colour painted with no word for it",
			{
				kind: "flags",
				colors: { red: "red" },
				groups: [{ flags: [["red", "blue"]] }],
			},
		],
		[
			"a word for a colour that paints nothing",
			{
				kind: "flags",
				colors: { red: "red", blue: "blue" },
				groups: [{ flags: [["red", "?"]] }],
			},
		],
		[
			"a word past its length",
			{
				kind: "flags",
				colors: { red: "r".repeat(17) },
				groups: [{ flags: [["red", "?"]] }],
			},
		],
		[
			"a flag of four stripes",
			{ kind: "flags", groups: [{ flags: [["?", "?", "?", "?"]] }] },
		],
		[
			"a group of seven flags",
			{
				kind: "flags",
				groups: [{ flags: Array.from({ length: 7 }, () => ["?", "?"]) }],
			},
		],
		[
			"thirteen flags in all",
			{
				kind: "flags",
				groups: [6, 6, 1].map((count) => ({
					flags: Array.from({ length: count }, () => ["?", "?"]),
				})),
			},
		],
		[
			"a group named by a colour and a label",
			{
				kind: "flags",
				colors: { red: "red" },
				groups: [{ label: "A", color: "red", flags: [["red", "?"]] }],
			},
		],
	])("refuses %s", (_, description) => {
		expect(readPicture(description)).toBeUndefined();
	});

	test("never fails, whatever it is handed", () => {
		const random = seeded(70);
		const pick = <T>(from: readonly T[]): T =>
			from[Math.floor(random() * from.length)] as T;
		const keys = [
			"kind",
			...kinds,
			"time",
			"rows",
			"items",
			"marks",
			"piles",
			"bars",
			"skip",
			"at",
			"colors",
			"groups",
			"flags",
			"color",
			...paints,
			"__proto__",
			"constructor",
		];
		const anything = (depth: number): unknown => {
			switch (Math.floor(random() * (depth > 3 ? 4 : 7))) {
				case 0:
					return pick([null, true, false, undefined]);
				case 1:
					return pick([0, -1, 2.5, 1e9, Number.NaN, 7, 40]);
				case 2:
					return pick([
						"",
						"A",
						"?",
						"4:30",
						"…",
						"WWWWWW",
						"-3",
						...kinds,
						...paints,
					]);
				case 3:
					return pick(kinds);
				case 4:
					return Array.from({ length: Math.floor(random() * 5) }, () =>
						anything(depth + 1),
					);
				default:
					return Object.fromEntries(
						Array.from({ length: Math.floor(random() * 6) }, () => [
							pick(keys),
							anything(depth + 1),
						]),
					);
			}
		};
		for (let run = 0; run < 3000; run++) {
			const value = anything(0);
			expect(() => readPicture(value)).not.toThrow();
		}
	});
});
