import type { Picture } from "../model";

/**
 * Extreme is a picture at the limits of its kind, and what it is: the most a
 * description of the kind may hold, with its labels as wide as labels come,
 * so that a picture laid out right here is laid out right for any picture of
 * its kind.
 */
export type Extreme = { name: string; picture: Picture };

// wide is a label as wide as a label of five characters comes.
const wide = "WWWWW";

// shortWide is a label as wide as a label of two characters comes, the most a
// lattice's labels may have.
const shortWide = "WW";

/** extremes are a picture of each kind at its limits. */
export const extremes: readonly Extreme[] = [
	{
		name: "clock, both hands labelled and close together",
		picture: {
			kind: "clock",
			time: "9:45",
			hour_label: wide,
			minute_label: "MMMMM",
		},
	},
	{
		name: "row of twelve items, every label, twice over",
		picture: {
			kind: "row",
			items: [
				{ label: "99999", below: wide, mark: "dot" },
				{ label: "99998", below: wide, mark: "ring" },
				{ label: "99997", below: wide, mark: "square" },
				{ label: "99996", below: wide },
				{ label: "99995", below: wide },
				{ skip: true },
				{ label: "99993", below: wide },
				{ label: "99992", below: wide },
				{ label: "99991", below: wide },
				{ label: "99990", below: wide },
				{ label: "99989", below: wide },
				{ label: "99988", below: wide },
			],
			gaps: "−9999",
			span: "ABCDE",
			copies: 2,
		},
	},
	{
		name: "four bars of twelve parts and segments, braces and notes",
		picture: {
			kind: "bars",
			bars: [
				{
					label: wide,
					parts: 12,
					shaded: 7,
					length: 60,
					value: wide,
					braces: [
						{ from: 0, to: 3, label: "99999" },
						{ from: 3, to: 6, label: "99998" },
						{ from: 6, to: 9, label: "99997" },
						{ from: 9, to: 12, label: "99996" },
					],
					span: "ABCDE",
				},
				{
					label: "MMMMM",
					segments: Array.from({ length: 12 }, () => ({
						size: 5,
						label: "99",
					})),
					value: "99999",
					braces: [{ from: 0, to: 12, label: wide }],
				},
				{ label: "A", value: "?" },
				{ label: "B", parts: 5, shaded: 5, length: 25 },
			],
			notes: [
				"WWWWW+WWWWW=WWWWW+WWWWW",
				"A + B = C − 99999",
				"(A + B) × 2 < 99999",
			],
		},
	},
	{
		name: "table of eight rows of six times, with a header",
		picture: {
			kind: "table",
			header: ["A", "B", "C", "D", "E", "F"],
			rows: Array.from({ length: 7 }, () => [
				"22:35",
				"23:59",
				"…",
				"10:05",
				"",
				"?",
			]),
		},
	},
	{
		name: "table of six columns of the widest labels",
		picture: {
			kind: "table",
			header: [wide, wide, wide, wide, wide, wide],
			rows: Array.from({ length: 7 }, () => [
				wide,
				"MMMMM",
				"99999",
				"−9999",
				"…",
				"",
			]),
		},
	},
	{
		name: "number line of sixteen five-digit ticks, a wide label on every one",
		picture: {
			kind: "number_line",
			from: -9999,
			to: 99996,
			step: 7333,
			marks: Array.from({ length: 16 }, (_, place) => ({
				at: -9999 + place * 7333,
				label: wide,
			})),
		},
	},
	{
		name: "ring of twenty-four places, started from the bottom with a wide label",
		picture: { kind: "ring", count: 24, start: { at: 13, label: wide } },
	},
	{
		name: "grid of eight rows and columns with wide names, filled and marked",
		picture: {
			kind: "grid",
			rows: [
				"WWWWA",
				"WWWWB",
				"WWWWC",
				"WWWWD",
				"WWWWE",
				"WWWWF",
				"WWWWG",
				"WWWWH",
			],
			cols: [
				"MMMMA",
				"MMMMB",
				"MMMMC",
				"MMMMD",
				"MMMME",
				"MMMMF",
				"MMMMG",
				"MMMMH",
			],
			filled: ["WWWWAMMMMA", "WWWWBMMMMB", "WWWWCMMMMC", "WWWWHMMMMH"],
			marks: { WWWWAMMMMB: wide, WWWWHMMMMA: "99999", WWWWDMMMMD: "?" },
		},
	},
	{
		name: "two groups of wide labels and counts, in both, neither and all",
		picture: {
			kind: "venn",
			sets: [
				{ label: wide, count: "99999" },
				{ label: "MMMMM", count: "99998" },
			],
			both: wide,
			neither: "99997",
			total: "99999",
		},
	},
	{
		name: "balance with four wide weights on each pan",
		picture: {
			kind: "balance",
			left: ["99999", "99998", "99997", "99996"],
			right: [wide, "MMMMM", "99995", "?"],
		},
	},
	{
		name: "four containers of twenty, empty, nearly empty, nearly full and full",
		picture: {
			kind: "containers",
			items: [
				{ capacity: 20, amount: 0, label: wide },
				{ capacity: 20, amount: 1, label: "MMMMM" },
				{ capacity: 20, amount: 19, label: "99999" },
				{ capacity: 1, amount: 1, label: "?" },
			],
		},
	},
	{
		name: "six piles of forty, one under another, in a box",
		picture: {
			kind: "piles",
			box: true,
			piles: [
				{ label: wide, value: "99999", count: 40, group: 5 },
				{ label: "MMMMM", value: "99998", count: 40, shown: 39, fill: "light" },
				{ skip: true },
				{ label: "99999", count: 40, group: 7, shape: "square", boxed: true },
				{ label: "?", value: "?", fill: "light", shape: "square" },
				{ label: "A", value: "0", count: 0, boxed: true },
			],
		},
	},
	{
		name: "six piles of forty side by side, in a box",
		picture: {
			kind: "piles",
			across: true,
			box: true,
			piles: [
				{ label: wide, value: "99999", count: 40, group: 5 },
				{ label: "MMMMM", value: "99998", count: 40, shown: 39, fill: "light" },
				{ skip: true },
				{ label: "99999", count: 40, group: 7, shape: "square", boxed: true },
				{ label: "?", value: "?", fill: "light", shape: "square" },
				{ label: "A", value: "0", count: 0, boxed: true },
			],
		},
	},
	{
		name: "month of six weeks, a wide label on every day of a week",
		picture: {
			kind: "calendar",
			first: 7,
			days: 31,
			marks: {
				"1": wide,
				"2": wide,
				"3": wide,
				"4": wide,
				"5": wide,
				"6": wide,
				"7": wide,
				"8": wide,
				"31": "?",
			},
		},
	},
	{
		name: "month starting on a Saturday, its weeks on Sundays",
		picture: {
			kind: "calendar",
			first: 6,
			days: 31,
			week_starts: "sunday",
			marks: { "12": "A", "30": "99999" },
		},
	},
	{
		name: "bar of one part, its length, its brace and a note as wide as they come",
		picture: {
			kind: "bars",
			bars: [
				{ parts: 1, span: wide, braces: [{ from: 0, to: 1, label: "MMMMM" }] },
			],
			notes: ["AB + CD = EF + GH - 99"],
		},
	},
	{
		name: "container of one, its label as wide as labels come",
		picture: {
			kind: "containers",
			items: [{ capacity: 1, amount: 0, label: wide }],
		},
	},
	{
		name: "row of two, its gaps and its length as wide as labels come",
		picture: { kind: "row", items: [{}, {}], gaps: wide, span: "MMMMM" },
	},
	{
		name: "number line of two ticks, a wide label on each",
		picture: {
			kind: "number_line",
			from: 0,
			to: 1,
			marks: [
				{ at: 0, label: wide },
				{ at: 1, label: "MMMMM" },
			],
		},
	},
	{
		name: "grid of one cell, its names as wide as labels come",
		picture: { kind: "grid", rows: ["A"], cols: [wide] },
	},
	{
		name: "pile of three side by side, its label and value as wide as they come",
		picture: {
			kind: "piles",
			across: true,
			piles: [{ label: wide, value: "MMMMM", count: 3 }],
		},
	},
	{
		name: "twelve flags of three stripes in four groups, each named as wide as labels come",
		picture: {
			kind: "flags",
			colors: { red: "red", yellow: "yellow", green: "green" },
			groups: [wide, "MMMMM", "99999", "WWWWW"].map((label) => ({
				label,
				flags: [
					["red", "yellow", "green"],
					["yellow", "green", "red"],
					["green", "red", "?"],
				],
			})),
		},
	},
	{
		name: "two groups of six flags, each named by a colour whose letters run to three",
		picture: {
			kind: "flags",
			colors: {
				red: "Wwwwwwwwwwwwwwww",
				blue: "Wwwwwwwwwwwwwwwm",
				black: "Wwwwwwwwwwwwwwwn",
			},
			groups: (["red", "blue"] as const).map((color) => ({
				color,
				flags: [
					["red", "blue", "black"],
					["blue", "black", "red"],
					["black", "red", "blue"],
					["red", "black", "blue"],
					["blue", "red", "black"],
					["black", "blue", "red"],
				],
			})),
		},
	},
	{
		name: "a flag of two unknown stripes, alone",
		picture: { kind: "flags", groups: [{ flags: [["?", "?"]] }] },
	},
	{
		name: "six flags of three stripes, their colours called in scripts whose letters reach high and low",
		picture: {
			kind: "flags",
			colors: {
				red: "أحمر",
				yellow: "पीला",
				green: "সবুজ",
				blue: "สีน้ำเงิน",
				white: "أبيض",
				black: "काला",
			},
			groups: [
				{
					color: "blue",
					flags: [
						["red", "yellow", "green"],
						["blue", "white", "black"],
						["white", "red", "blue"],
						["yellow", "black", "?"],
						["green", "blue", "red"],
						["black", "green", "white"],
					],
				},
			],
		},
	},
	{
		name: "flags of three stripes, their colours called in Chinese, Japanese and Korean",
		picture: {
			kind: "flags",
			colors: {
				red: "红色",
				yellow: "黄色い",
				green: "초록색",
				blue: "青",
				white: "白色",
				black: "검은색",
			},
			groups: [
				{
					label: "A",
					flags: [
						["red", "yellow", "green"],
						["blue", "white", "black"],
						["white", "red", "blue"],
					],
				},
				{
					color: "red",
					flags: [
						["yellow", "black", "?"],
						["green", "blue", "red"],
						["black", "green", "white"],
					],
				},
			],
		},
	},
	{
		name: "lattice of eight by eight, every cell filled, its lines, dashes, points and marks as many and as wide as they come",
		picture: {
			kind: "lattice",
			width: 8,
			height: 8,
			cells: cellsOf(8, 8, () => true),
			lines: [
				...unitLines(8, (x) => [x, 0, x + 1, 0]),
				...unitLines(8, (x) => [x, 8, x + 1, 8]),
				...unitLines(8, (x) => [x, 4, x + 1, 4]),
				...unitLines(8, (y) => [0, y, 0, y + 1]),
				...unitLines(8, (y) => [8, y, 8, y + 1]),
				...unitLines(8, (y) => [4, y, 4, y + 1]),
			],
			dashed: unitLines(8, (x) => [x, 2, x + 1, 2]),
			points: [5, 7].flatMap((y) =>
				Array.from({ length: 8 }, (_, x) => ({
					at: [x, y] as [number, number],
					label: shortWide,
				})),
			),
			marks: cellsOf(8, 5, (x, y) => y < 4 || x < 4).map((at) => ({
				at,
				label: shortWide,
			})),
		},
	},
	{
		name: "lattice of one cell, a mark in it and a wide label at each corner",
		picture: {
			kind: "lattice",
			width: 1,
			height: 1,
			points: [
				{ at: [0, 0], label: shortWide },
				{ at: [1, 0], label: "MM" },
				{ at: [0, 1], label: "99" },
				{ at: [1, 1], label: shortWide },
			],
			marks: [{ at: [0, 0], label: shortWide }],
		},
	},
	{
		name: "ring of ten pentagons, a two-digit number in every triangle",
		picture: {
			kind: "tiles",
			sides: 5,
			layout: "ring",
			tiles: Array.from({ length: 10 }, () => ["99", "88", "?", "00", "99"]),
		},
	},
	{
		name: "row of four hexagons, a two-digit number in every triangle",
		picture: {
			kind: "tiles",
			sides: 6,
			layout: "row",
			tiles: Array.from({ length: 4 }, () => [
				"99",
				"88",
				"?",
				"00",
				"99",
				"88",
			]),
		},
	},
	{
		name: "row of six squares, a two-digit number in every triangle",
		picture: {
			kind: "tiles",
			sides: 4,
			layout: "row",
			tiles: Array.from({ length: 6 }, () => ["99", "88", "?", "00"]),
		},
	},
	{
		name: "ring of six triangles, a two-digit number in every triangle",
		picture: {
			kind: "tiles",
			sides: 3,
			layout: "ring",
			joint: 1,
			tiles: Array.from({ length: 6 }, () => ["99", "88", "00"]),
		},
	},
	{
		name: "five rows of six shapes, each total as wide as labels come",
		picture: {
			kind: "shapes",
			rows: Array.from({ length: 5 }, () => ({
				shapes: ["triangle", "square", "circle", "diamond", "square", "circle"],
				total: wide,
			})),
		},
	},
	{
		name: "staircase of sixteen steps, four wide marks on its top steps and four jumps",
		picture: {
			kind: "stairs",
			steps: 16,
			marks: [13, 14, 15, 16].map((step) => ({ step, label: wide })),
			jumps: [
				{ from: 0, to: 16 },
				{ from: 16, to: 0 },
				{ from: 15, to: 16 },
				{ from: 1, to: 2 },
			],
		},
	},
	{
		name: "staircase of three steps, a wide mark on every step and the floor",
		picture: {
			kind: "stairs",
			steps: 3,
			marks: [0, 1, 2, 3].map((step) => ({ step, label: "MMMMM" })),
			jumps: [{ from: 0, to: 3 }],
		},
	},
];

// cellsOf are the cells of a lattice so wide and so tall that a test keeps,
// row by row.
function cellsOf(
	width: number,
	height: number,
	keeps: (x: number, y: number) => boolean,
): [number, number][] {
	return Array.from({ length: height }, (_, y) =>
		Array.from({ length: width }, (_, x) => [x, y] as [number, number]),
	)
		.flat()
		.filter(([x, y]) => keeps(x, y));
}

// unitLines are as many lines of a lattice as given, each a cell long, the nth
// where at says.
function unitLines(
	count: number,
	at: (n: number) => [number, number, number, number],
): [number, number, number, number][] {
	return Array.from({ length: count }, (_, n) => at(n));
}
