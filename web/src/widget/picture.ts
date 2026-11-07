import * as z from "zod";
import { type Picture, paints, shapeNames } from "../design/picture/model";
import { pointLabelsOf } from "../design/picture/pointLabels";
import { jointOf, turnOf } from "../design/picture/tiling";

/**
 * pictureLimits are the limits a picture is held to, under the names the
 * service gives them, so that the two are seen to be one list: the card reads
 * a picture against the limits its drawing is laid out for.
 */
export const pictureLimits = {
	MaxLabelCharacters: 5,
	MaxNoteCharacters: 24,
	MaxTotalCharacters: 32,
	minTableRows: 1,
	maxTableRows: 8,
	minTableCells: 1,
	maxTableCells: 6,
	minTicks: 2,
	maxTicks: 16,
	lowestTick: -9999,
	highestTick: 99999,
	minRowItems: 2,
	maxRowItems: 12,
	maxCopies: 2,
	minRingPlaces: 3,
	maxRingPlaces: 24,
	minGridSide: 1,
	maxGridSide: 8,
	minBars: 1,
	maxBars: 4,
	minParts: 1,
	maxParts: 12,
	minLength: 1,
	maxLength: 60,
	maxBraces: 4,
	maxNotes: 3,
	vennSets: 2,
	maxOnAPan: 4,
	minContainers: 1,
	maxContainers: 4,
	minCapacity: 1,
	maxCapacity: 20,
	minPiles: 1,
	maxPiles: 6,
	maxPileCount: 40,
	fewestShown: 2,
	minGroup: 1,
	maxGroup: 40,
	firstWeekday: 1,
	lastWeekday: 7,
	fewestDays: 28,
	mostDays: 31,
	minFlagGroups: 1,
	maxFlagGroups: 4,
	minFlags: 1,
	maxFlagsInGroup: 6,
	maxFlags: 12,
	minStripes: 2,
	maxStripes: 3,
	minLatticeSide: 1,
	maxLatticeSide: 8,
	maxLatticeCells: 64,
	maxLatticeLines: 48,
	maxLatticeDashed: 8,
	maxLatticePoints: 16,
	maxLatticeMarks: 36,
	maxShortLabelCharacters: 2,
	minTileSides: 3,
	maxTileSides: 6,
	minTiles: 2,
	maxTiles: 12,
	maxTilesInRow: 6,
	maxHexagonsInRow: 4,
	maxTileDigits: 2,
	minShapeRows: 1,
	maxShapeRows: 5,
	maxShapesInRow: 6,
	minSteps: 3,
	maxSteps: 16,
	maxStairMarks: 4,
	maxStairJumps: 4,
	maxColorWords: 2,
	maxColorCharacters: 16,
} as const;

const limits = pictureLimits;

// absent is what a member written as null or as an empty text stands for: a
// member left out, as the service reads it.
function absent(value: unknown): unknown {
	return value === null || value === "" ? undefined : value;
}

// optional is a member a description may leave out, written as null or as an
// empty text all the same.
function optional<T extends z.ZodType>(schema: T) {
	return z.preprocess(absent, schema.optional());
}

// whole is a whole number between two limits: JSON writes 3.0 as 3.
function whole(least: number, most: number) {
	return z.number().int().min(least).max(most);
}

// A label, a time and a note as the card reads them: their grammar, and the
// decimal mark of the lesson's language, are the service's to hold, and the
// card needs only what its drawing depends on — how long each may be.
const label = z.string().min(1).max(limits.MaxLabelCharacters);
const time = z.string().regex(/^([01]?\d|2[0-3]):([0-5]\d)$/);
const note = z.string().min(1).max(limits.MaxNoteCharacters);
const cell = z.string().max(limits.MaxLabelCharacters);

const clock = z.strictObject({
	kind: z.literal("clock"),
	time,
	hour_label: optional(label),
	minute_label: optional(label),
});

const tableRow = z
	.array(cell)
	.min(limits.minTableCells)
	.max(limits.maxTableCells);

const table = z
	.strictObject({
		kind: z.literal("table"),
		header: optional(tableRow),
		rows: z.array(tableRow).min(limits.minTableRows).max(limits.maxTableRows),
	})
	.superRefine((read, problems) => {
		const width = read.header?.length ?? read.rows[0]?.length ?? 0;
		if (read.rows.some((row) => row.length !== width)) {
			problems.addIssue({
				code: "custom",
				message: "every row is as long as the header, or as the first row",
			});
		}
	});

const tick = whole(limits.lowestTick, limits.highestTick);

const numberLine = z
	.strictObject({
		kind: z.literal("number_line"),
		from: tick,
		to: tick,
		step: optional(whole(1, limits.highestTick - limits.lowestTick)),
		marks: optional(
			z
				.array(z.strictObject({ at: tick, label: optional(label) }))
				.max(limits.maxTicks),
		),
	})
	.superRefine((read, problems) => {
		const step = read.step ?? 1;
		const span = read.to - read.from;
		const ticks = span / step + 1;
		if (
			span <= 0 ||
			span % step !== 0 ||
			ticks < limits.minTicks ||
			ticks > limits.maxTicks
		) {
			problems.addIssue({
				code: "custom",
				message: `the ends and the step make ${limits.minTicks} to ${limits.maxTicks} ticks`,
			});
			return;
		}
		const taken = new Set<number>();
		for (const mark of read.marks ?? []) {
			const onATick =
				mark.at >= read.from &&
				mark.at <= read.to &&
				(mark.at - read.from) % step === 0;
			if (!onATick || taken.has(mark.at)) {
				problems.addIssue({
					code: "custom",
					message: "every mark stands on a tick no other mark stands on",
				});
			}
			taken.add(mark.at);
		}
	});

// skipsPlaced says whether the skips among some items stand where a skip may:
// never first or last, and never next to another skip.
function skipsPlaced(skips: boolean[]): boolean {
	return skips.every(
		(skip, at) => !skip || (at > 0 && at < skips.length - 1 && !skips[at - 1]),
	);
}

// aSkipAlone says whether an item that is a skip holds nothing else.
function aSkipAlone(item: Record<string, unknown>): boolean {
	return (
		item.skip !== true ||
		Object.entries(item).every(
			([name, value]) => name === "skip" || value === undefined,
		)
	);
}

const rowItem = z.strictObject({
	label: optional(label),
	below: optional(label),
	mark: optional(z.enum(["dot", "ring", "square"])),
	skip: optional(z.boolean()),
});

const row = z
	.strictObject({
		kind: z.literal("row"),
		items: z.array(rowItem).min(limits.minRowItems).max(limits.maxRowItems),
		line: optional(z.boolean()),
		gaps: optional(label),
		span: optional(label),
		copies: optional(whole(1, limits.maxCopies)),
	})
	.superRefine((read, problems) => {
		if (
			!skipsPlaced(read.items.map((item) => item.skip === true)) ||
			!read.items.every(aSkipAlone)
		) {
			problems.addIssue({
				code: "custom",
				message:
					"a skip holds nothing, is never first or last, and is never next to a skip",
			});
		}
	});

const ring = z
	.strictObject({
		kind: z.literal("ring"),
		count: whole(limits.minRingPlaces, limits.maxRingPlaces),
		start: optional(
			z.strictObject({
				at: whole(1, limits.maxRingPlaces),
				label: optional(label),
			}),
		),
	})
	.superRefine((read, problems) => {
		if (read.start !== undefined && read.start.at > read.count) {
			problems.addIssue({
				code: "custom",
				message: "the start is a place of the ring",
			});
		}
	});

const gridNames = z
	.array(label)
	.min(limits.minGridSide)
	.max(limits.maxGridSide);

const grid = z
	.strictObject({
		kind: z.literal("grid"),
		rows: gridNames,
		cols: gridNames,
		filled: optional(
			z.array(z.string()).max(limits.maxGridSide * limits.maxGridSide),
		),
		marks: optional(z.record(z.string(), label)),
	})
	.superRefine((read, problems) => {
		const cells = new Set<string>();
		let alike =
			new Set(read.rows).size < read.rows.length ||
			new Set(read.cols).size < read.cols.length;
		for (const name of read.rows) {
			for (const col of read.cols) {
				alike ||= cells.has(name + col);
				cells.add(name + col);
			}
		}
		const filled = read.filled ?? [];
		const inTheGrid = [...filled, ...Object.keys(read.marks ?? {})].every(
			(at) => cells.has(at),
		);
		if (alike || !inTheGrid || new Set(filled).size < filled.length) {
			problems.addIssue({
				code: "custom",
				message:
					"every line and cell of the grid is named once, and every cell it names is in it",
			});
		}
	});

const brace = z.strictObject({
	from: whole(0, limits.maxParts),
	to: whole(0, limits.maxParts),
	label,
});

const bar = z
	.strictObject({
		label: optional(label),
		parts: optional(whole(limits.minParts, limits.maxParts)),
		shaded: optional(whole(0, limits.maxParts)),
		segments: optional(
			z
				.array(
					z.strictObject({
						size: whole(limits.minLength, limits.maxLength),
						label: optional(label),
					}),
				)
				.min(limits.minParts)
				.max(limits.maxParts),
		),
		length: optional(whole(limits.minLength, limits.maxLength)),
		value: optional(label),
		braces: optional(z.array(brace).max(limits.maxBraces)),
		span: optional(label),
	})
	.superRefine((read, problems) => {
		const divisions = read.parts ?? read.segments?.length ?? 1;
		const cut = read.parts === undefined || read.segments === undefined;
		const shaded =
			read.shaded === undefined ||
			(read.parts !== undefined && read.shaded <= read.parts);
		const braced = (read.braces ?? []).every(
			(one) => one.from < one.to && one.to <= divisions,
		);
		if (!cut || !shaded || !braced) {
			problems.addIssue({
				code: "custom",
				message:
					"a bar is cut into parts or segments, shades no more than its parts, and has its braces within it",
			});
		}
	});

const bars = z.strictObject({
	kind: z.literal("bars"),
	bars: z.array(bar).min(limits.minBars).max(limits.maxBars),
	notes: optional(z.array(note).max(limits.maxNotes)),
});

const venn = z.strictObject({
	kind: z.literal("venn"),
	sets: z
		.array(z.strictObject({ label, count: optional(label) }))
		.length(limits.vennSets),
	both: optional(label),
	neither: optional(label),
	total: optional(label),
});

const pan = z.array(label).max(limits.maxOnAPan);

const balance = z.strictObject({
	kind: z.literal("balance"),
	left: pan,
	right: pan,
});

const containers = z
	.strictObject({
		kind: z.literal("containers"),
		items: z
			.array(
				z.strictObject({
					capacity: whole(limits.minCapacity, limits.maxCapacity),
					amount: whole(0, limits.maxCapacity),
					label: optional(label),
				}),
			)
			.min(limits.minContainers)
			.max(limits.maxContainers),
	})
	.superRefine((read, problems) => {
		if (read.items.some((item) => item.amount > item.capacity)) {
			problems.addIssue({
				code: "custom",
				message: "no container holds more than it can",
			});
		}
	});

const pile = z
	.strictObject({
		label: optional(label),
		value: optional(label),
		count: optional(whole(0, limits.maxPileCount)),
		shown: optional(whole(limits.fewestShown, limits.maxPileCount - 1)),
		group: optional(whole(limits.minGroup, limits.maxGroup)),
		fill: optional(z.enum(["dark", "light"])),
		shape: optional(z.enum(["dot", "square"])),
		boxed: optional(z.boolean()),
		skip: optional(z.boolean()),
	})
	.superRefine((read, problems) => {
		if (
			read.shown !== undefined &&
			(read.count === undefined || read.shown >= read.count)
		) {
			problems.addIssue({
				code: "custom",
				message: "a pile cut short has a count, and shows less than it",
			});
		}
	});

const piles = z
	.strictObject({
		kind: z.literal("piles"),
		piles: z.array(pile).min(limits.minPiles).max(limits.maxPiles),
		box: optional(z.boolean()),
		across: optional(z.boolean()),
	})
	.superRefine((read, problems) => {
		if (
			!skipsPlaced(read.piles.map((one) => one.skip === true)) ||
			!read.piles.every(aSkipAlone)
		) {
			problems.addIssue({
				code: "custom",
				message:
					"a skip holds nothing, is never first or last, and is never next to a skip",
			});
		}
	});

const calendar = z
	.strictObject({
		kind: z.literal("calendar"),
		first: whole(limits.firstWeekday, limits.lastWeekday),
		days: whole(limits.fewestDays, limits.mostDays),
		week_starts: optional(z.enum(["monday", "sunday"])),
		marks: optional(z.record(z.string(), label)),
	})
	.superRefine((read, problems) => {
		const days = Object.keys(read.marks ?? {});
		if (
			!days.every((day) => /^[1-9]\d?$/.test(day) && Number(day) <= read.days)
		) {
			problems.addIssue({
				code: "custom",
				message: "every mark stands on a day of the month",
			});
		}
	});

// The word of a colour as the card reads it: its grammar is the service's to
// hold, and the card needs only how long it may be.
const colorWord = z.string().min(1).max(limits.maxColorCharacters);
const paint = z.enum(paints);
const colors = z.strictObject({
	red: optional(colorWord),
	yellow: optional(colorWord),
	green: optional(colorWord),
	blue: optional(colorWord),
	white: optional(colorWord),
	black: optional(colorWord),
});

const flagGroup = z
	.strictObject({
		label: optional(label),
		color: optional(paint),
		flags: z
			.array(
				z
					.array(z.union([paint, z.literal("?")]))
					.min(limits.minStripes)
					.max(limits.maxStripes),
			)
			.min(limits.minFlags)
			.max(limits.maxFlagsInGroup),
	})
	.refine((read) => read.label === undefined || read.color === undefined, {
		message: "a group is named by a colour or a label, never both",
	});

const flags = z
	.strictObject({
		kind: z.literal("flags"),
		colors: optional(colors),
		groups: z
			.array(flagGroup)
			.min(limits.minFlagGroups)
			.max(limits.maxFlagGroups),
	})
	.superRefine((read, problems) => {
		const painted = new Set<string>();
		let total = 0;
		for (const group of read.groups) {
			total += group.flags.length;
			for (const name of [group.color, ...group.flags.flat()]) {
				if (name !== undefined && name !== "?") {
					painted.add(name);
				}
			}
		}
		const named = paints.filter((name) => read.colors?.[name] !== undefined);
		if (total > limits.maxFlags) {
			problems.addIssue({
				code: "custom",
				message: "a picture holds at most 12 flags in all",
			});
		}
		if (
			named.length !== painted.size ||
			!named.every((name) => painted.has(name))
		) {
			problems.addIssue({
				code: "custom",
				message: "every colour painted has its word, and every word paints",
			});
		}
	});

// A lattice's corners and cells as the card reads them: [x, y], two whole
// numbers, each at most its end.
function pairWithin(width: number, height: number) {
	return (pair: readonly number[]) =>
		(pair[0] ?? 0) <= width && (pair[1] ?? 0) <= height;
}

const coordinate = whole(0, limits.maxLatticeSide);
const corner = z.tuple([coordinate, coordinate]);
const segment = z.tuple([coordinate, coordinate, coordinate, coordinate]);
// A lattice's label is counted in characters, as the service counts them,
// rather than in the halves of a character some scripts are written in.
const shortLabel = label.refine(
	(text) => [...text].length <= limits.maxShortLabelCharacters,
);
const latticeLabel = z.strictObject({
	at: corner,
	label: optional(shortLabel),
});

// keyOf is what tells a cell, a corner or a line from the others, a line the
// same whichever end it is written from.
function keyOf(numbers: readonly number[]): string {
	const ends = [numbers.slice(0, 2).join(","), numbers.slice(2).join(",")];
	return ends.sort().join(" ");
}

// once says whether no two of some places, cells or lines, are the same.
function once(places: readonly (readonly number[])[]): boolean {
	return new Set(places.map(keyOf)).size === places.length;
}

const lattice = z
	.strictObject({
		kind: z.literal("lattice"),
		width: whole(limits.minLatticeSide, limits.maxLatticeSide),
		height: whole(limits.minLatticeSide, limits.maxLatticeSide),
		cells: optional(z.array(corner).max(limits.maxLatticeCells)),
		lines: optional(z.array(segment).max(limits.maxLatticeLines)),
		dashed: optional(z.array(segment).max(limits.maxLatticeDashed)),
		points: optional(z.array(latticeLabel).max(limits.maxLatticePoints)),
		marks: optional(z.array(latticeLabel).max(limits.maxLatticeMarks)),
	})
	.superRefine((read, problems) => {
		const inACell = pairWithin(read.width - 1, read.height - 1);
		const onACorner = pairWithin(read.width, read.height);
		const cells = read.cells ?? [];
		const marks = (read.marks ?? []).map((one) => one.at);
		const points = (read.points ?? []).map((one) => one.at);
		const segments = [...(read.lines ?? []), ...(read.dashed ?? [])];
		const placed =
			[...cells, ...marks].every(inACell) &&
			points.every(onACorner) &&
			segments.every(
				(one) => onACorner(one.slice(0, 2)) && onACorner(one.slice(2)),
			);
		const apart =
			once(cells) &&
			once(marks) &&
			once(points) &&
			once(segments) &&
			segments.every((one) => one[0] !== one[2] || one[1] !== one[3]);
		const marked = (read.marks ?? []).every((one) => one.label !== undefined);
		if (!placed || !apart || !marked) {
			problems.addIssue({
				code: "custom",
				message:
					"every cell, mark, point and line lies in the lattice, once, a line joins two corners and a mark has its label",
			});
		} else if (pointLabelsOf(read) === undefined) {
			problems.addIssue({
				code: "custom",
				message:
					"every point's label has a cell of its own round its corner, one no mark holds",
			});
		}
	});

// A tile's triangle holds a whole number of at most two digits, ? for the
// one asked about, or nothing.
const triangle = z.union([
	z.literal(""),
	z.literal("?"),
	z.string().regex(new RegExp(`^\\d{1,${limits.maxTileDigits}}$`)),
]);

// laidOut says whether tiles lie as their layout says: a row runs straight
// on, each tile laid against the side opposite the one it shares with the
// tile before, and holds at most so many tiles, fewer of hexagons, whose
// triangles are narrower; a ring goes round once and closes, its last tile
// against its first.
function laidOut(sides: number, joint: number, count: number, ring: boolean) {
	const turn = turnOf(sides, joint);
	const most = sides === 6 ? limits.maxHexagonsInRow : limits.maxTilesInRow;
	return ring ? count * Math.abs(turn) === sides : turn === 0 && count <= most;
}

const tiles = z
	.strictObject({
		kind: z.literal("tiles"),
		sides: whole(limits.minTileSides, limits.maxTileSides),
		layout: z.enum(["row", "ring"]),
		joint: optional(whole(1, limits.maxTileSides - 1)),
		tiles: z.array(z.array(triangle)).min(limits.minTiles).max(limits.maxTiles),
	})
	.superRefine((read, problems) => {
		const joint = jointOf(read);
		const shaped =
			read.tiles.every((one) => one.length === read.sides) &&
			joint < read.sides;
		if (
			!shaped ||
			!laidOut(read.sides, joint, read.tiles.length, read.layout === "ring")
		) {
			problems.addIssue({
				code: "custom",
				message: `every tile has a label for each of its sides; a row of at most ${limits.maxTilesInRow} tiles, ${limits.maxHexagonsInRow} hexagons, runs straight, and a ring closes once round`,
			});
		}
	});

const shapes = z.strictObject({
	kind: z.literal("shapes"),
	rows: z
		.array(
			z.strictObject({
				shapes: z.array(z.enum(shapeNames)).min(1).max(limits.maxShapesInRow),
				total: label,
			}),
		)
		.min(limits.minShapeRows)
		.max(limits.maxShapeRows),
});

const step = whole(0, limits.maxSteps);

const stairs = z
	.strictObject({
		kind: z.literal("stairs"),
		steps: whole(limits.minSteps, limits.maxSteps),
		marks: optional(
			z.array(z.strictObject({ step, label })).max(limits.maxStairMarks),
		),
		jumps: optional(
			z
				.array(z.strictObject({ from: step, to: step }))
				.max(limits.maxStairJumps),
		),
	})
	.superRefine((read, problems) => {
		const marks = (read.marks ?? []).map((one) => one.step);
		const jumps = read.jumps ?? [];
		const onTheStairs = [
			...marks,
			...jumps.flatMap((one) => [one.from, one.to]),
		].every((at) => at <= read.steps);
		const apart =
			new Set(marks).size === marks.length &&
			new Set(jumps.map((one) => `${one.from} ${one.to}`)).size ===
				jumps.length &&
			jumps.every((one) => one.from !== one.to);
		if (!onTheStairs || !apart) {
			problems.addIssue({
				code: "custom",
				message:
					"every mark and jump stands on a step of the stairs, a mark alone on its step, and a jump, drawn once, goes from one step to another",
			});
		}
	});

/**
 * pictureFormat reads a picture of one of the kinds, held to the format as
 * far as its drawing depends on it. A description it refuses is one the card
 * could not draw as the service read it.
 */
export const pictureFormat: z.ZodType<Picture> = z.discriminatedUnion("kind", [
	clock,
	table,
	numberLine,
	row,
	ring,
	grid,
	bars,
	venn,
	balance,
	containers,
	piles,
	calendar,
	flags,
	lattice,
	tiles,
	shapes,
	stairs,
]);
