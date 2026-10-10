/**
 * The format of a task's picture, as the card draws it: one of thirteen kinds,
 * each with the members of its kind and no other. The service holds every
 * picture to the format before it hands it out, so what reaches a drawing is
 * a description it can draw; the reader in widget/picture.ts holds it to the
 * format again, as far as the drawing depends on it.
 *
 * A member left out is undefined: the reader turns a member written as null
 * or as an empty text into one left out, as the service does.
 */

/** kinds are every kind of picture, in the order the format lists them. */
export const kinds = [
	"clock",
	"table",
	"number_line",
	"row",
	"ring",
	"grid",
	"bars",
	"venn",
	"balance",
	"containers",
	"piles",
	"calendar",
	"flags",
] as const;

/** Kind is what a picture draws. */
export type Kind = (typeof kinds)[number];

/**
 * Clock is a clock face: the time its hands show, H:MM from 0:00 to 23:59,
 * and the labels at the tips of its hands where the question calls them so.
 */
export type Clock = {
	kind: "clock";
	time: string;
	hour_label?: string | undefined;
	minute_label?: string | undefined;
};

/**
 * Table is a table of cells: a header row, if it has one, and its rows. A
 * cell is a label, a time, an ellipsis for the cells left out, or empty.
 */
export type Table = {
	kind: "table";
	header?: string[] | undefined;
	rows: string[][];
};

/** LineMark is a mark on a tick of a number line, and its label. */
export type LineMark = { at: number; label?: string | undefined };

/**
 * NumberLine is a line of evenly spaced ticks the card numbers itself, from
 * one end to the other by its step, 1 when it gives none, with the marks the
 * question names standing on ticks.
 */
export type NumberLine = {
	kind: "number_line";
	from: number;
	to: number;
	step?: number | undefined;
	marks?: LineMark[] | undefined;
};

/**
 * RowItem is one object of a row, with the label over it, the one under it and
 * its mark, a dot when it names none; or a skip, a stretch of the row cut
 * short, which holds nothing else.
 */
export type RowItem = {
	label?: string | undefined;
	below?: string | undefined;
	mark?: "dot" | "ring" | "square" | undefined;
	skip?: boolean | undefined;
};

/**
 * Row is objects in a row — posts, trees, children in a queue — joined by a
 * line unless line is false, with a label under every gap drawn, a label of
 * the whole length, and drawn once or, with copies 2, twice, as the two sides
 * of a path.
 */
export type Row = {
	kind: "row";
	items: RowItem[];
	line?: boolean | undefined;
	gaps?: string | undefined;
	span?: string | undefined;
	copies?: number | undefined;
};

/** RingStart is the place a ring starts at, and its label. */
export type RingStart = { at: number; label?: string | undefined };

/**
 * Ring is places in a ring, which the card numbers from 1 in the order of
 * travel and draws clockwise, and the place to start at.
 */
export type Ring = {
	kind: "ring";
	count: number;
	start?: RingStart | undefined;
};

/**
 * Grid is a grid whose rows and columns are named, so that a cell is its
 * row's name and its column's, B2: the cells it fills, and the labels it sets
 * in cells, keyed by the cell.
 */
export type Grid = {
	kind: "grid";
	rows: string[];
	cols: string[];
	filled?: string[] | undefined;
	marks?: Record<string, string> | undefined;
};

/** Segment is a part of a bar of a size of its own, and its label. */
export type Segment = { size: number; label?: string | undefined };

/**
 * Brace is a brace under a bar, from one edge between its parts or segments to
 * a later one, the edges counted from 0 at the bar's start, and its label.
 */
export type Brace = { from: number; to: number; label: string };

/**
 * Bar is one bar: its label, the equal parts it is cut into and how many of
 * them are shaded, or the segments it is cut into; its length to the scale of
 * the others; the value at its end; the braces under it; and the label of its
 * whole length.
 */
export type Bar = {
	label?: string | undefined;
	parts?: number | undefined;
	shaded?: number | undefined;
	segments?: Segment[] | undefined;
	length?: number | undefined;
	value?: string | undefined;
	braces?: Brace[] | undefined;
	span?: string | undefined;
};

/** Bars is quantities as bars drawn to one scale, and notes under them all. */
export type Bars = {
	kind: "bars";
	bars: Bar[];
	notes?: string[] | undefined;
};

/** VennSet is one of two groups: its label, and how many it holds. */
export type VennSet = { label: string; count?: string | undefined };

/**
 * Venn is two groups that overlap, inside a box of everyone the question
 * counts: how many are in both, in neither, and in all.
 */
export type Venn = {
	kind: "venn";
	sets: VennSet[];
	both?: string | undefined;
	neither?: string | undefined;
	total?: string | undefined;
};

/** Balance is a pan balance and the labels of what stands on each pan. */
export type Balance = {
	kind: "balance";
	left: string[];
	right: string[];
};

/** Container is one container: how much it holds, how much is in it, and its label. */
export type Container = {
	capacity: number;
	amount: number;
	label?: string | undefined;
};

/** Containers is jugs, buckets and the like to pour between. */
export type Containers = {
	kind: "containers";
	items: Container[];
};

/**
 * Pile is one pile of counters, or a skip among the piles: its label at its
 * start, its value at its end, how many counters it has where it says, how
 * many of them it shows when it is cut short, a gap after every so many, the
 * counters' fill and shape, and whether it stands in a box of its own.
 */
export type Pile = {
	label?: string | undefined;
	value?: string | undefined;
	count?: number | undefined;
	shown?: number | undefined;
	group?: number | undefined;
	fill?: "dark" | "light" | undefined;
	shape?: "dot" | "square" | undefined;
	boxed?: boolean | undefined;
	skip?: boolean | undefined;
};

/**
 * Piles is the piles of a game, one under another or, with across, side by
 * side, all in one box with box.
 */
export type Piles = {
	kind: "piles";
	piles: Pile[];
	box?: boolean | undefined;
	across?: boolean | undefined;
};

/**
 * Calendar is the page of a month: the weekday of its 1st, 1 for Monday to 7
 * for Sunday, how many days it has, the day its weeks start on, and the labels
 * it sets on days, keyed by the day.
 */
export type Calendar = {
	kind: "calendar";
	first: number;
	days: number;
	week_starts?: "monday" | "sunday" | undefined;
	marks?: Record<string, string> | undefined;
};

/** paints are the colours of the palette, in the order the format lists them. */
export const paints = [
	"red",
	"yellow",
	"green",
	"blue",
	"white",
	"black",
] as const;

/** Paint is a colour of the palette a picture paints with. */
export type Paint = (typeof paints)[number];

/**
 * Colors are the colours a picture paints with, each with the word the lesson
 * calls it by, which the card writes beside its paint under the picture.
 */
export type Colors = Partial<Record<Paint, string>>;

/** Stripe is a stripe of a flag: a colour, or ? where its colour is the unknown. */
export type Stripe = Paint | "?";

/**
 * FlagGroup is a group of flags, each flag its stripes from top to bottom,
 * and what names the group under it: a label, or a colour.
 */
export type FlagGroup = {
	label?: string | undefined;
	color?: Paint | undefined;
	flags: Stripe[][];
};

/** Flags is flags in groups, painted with the colours its colors name. */
export type Flags = {
	kind: "flags";
	colors?: Colors | undefined;
	groups: FlagGroup[];
};

/** Picture is a description of a task's picture, of one of the kinds. */
export type Picture =
	| Clock
	| Table
	| NumberLine
	| Row
	| Ring
	| Grid
	| Bars
	| Venn
	| Balance
	| Containers
	| Piles
	| Calendar
	| Flags;
