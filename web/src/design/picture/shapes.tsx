import type { ComponentChildren } from "preact";
import { lines, type Placed, stacked, widthOf } from "./text";

/**
 * Drawn is a picture laid out: how wide and how tall it is, and what it
 * draws, every length of it in the same units as the card's pixels.
 */
export type Drawn = { width: number; height: number; body: ComponentChildren };

/** r1 is a length to a tenth, so that a picture's markup stays short. */
export function r1(value: number): number {
	return Math.round(value * 10) / 10;
}

/** The sizes of what a picture writes, in the card's pixels. */
export const sizes = {
	/** label is the size of a label the description gives. */
	label: 14,
	/** number is the size of a number the card writes itself. */
	number: 12,
	/** day is the size of a day of a month's page. */
	day: 13,
	/** numeral is the size of a number of a clock face. */
	numeral: 15,
} as const;

/**
 * Label is one text of a picture, centred on a point unless anchored at its
 * start or its end, in the colour of the card's text, or of the tone a page
 * lights it in: a label the description gives is drawn strong, a number the
 * card writes itself plain.
 */
export function Label({
	x,
	y,
	text,
	size = sizes.label,
	strong = true,
	anchor = "middle",
	tone,
}: {
	x: number;
	y: number;
	text: string;
	size?: number;
	strong?: boolean;
	anchor?: "middle" | "start" | "end";
	tone?: string | undefined;
}) {
	return (
		<text
			x={r1(x)}
			y={r1(y)}
			dy="0.35em"
			text-anchor={anchor}
			font-size={size}
			class={strong ? "mt-pic-text mt-pic-strong" : "mt-pic-text"}
			data-tone={tone}
		>
			{text}
		</text>
	);
}

/**
 * Plate is a label on a patch of the card's own colour, which hides what is
 * drawn under it, so that a label written over a line can be read.
 */
export function Plate({
	x,
	y,
	text,
	size = sizes.label,
}: {
	x: number;
	y: number;
	text: string;
	size?: number;
}) {
	const width = plateWidth(text, size);
	const height = size * 1.4;
	return (
		<>
			<rect
				x={r1(x - width / 2)}
				y={r1(y - height / 2)}
				width={r1(width)}
				height={r1(height)}
				class="mt-pic-plate"
			/>
			<Label x={x} y={y} text={text} size={size} />
		</>
	);
}

/**
 * arrowhead is the path of an arrowhead with its tip at a point, pointing in
 * a direction in degrees clockwise from the right — 0 points right, 90 down —
 * as long as given and twice half as wide.
 */
export function arrowhead(
	x: number,
	y: number,
	degrees: number,
	length = 8,
	half = 5,
): string {
	const radians = (degrees * Math.PI) / 180;
	const back = {
		x: x - length * Math.cos(radians),
		y: y - length * Math.sin(radians),
	};
	const side = { x: -half * Math.sin(radians), y: half * Math.cos(radians) };
	return (
		`M${r1(x)} ${r1(y)}` +
		`L${r1(back.x + side.x)} ${r1(back.y + side.y)}` +
		`L${r1(back.x - side.x)} ${r1(back.y - side.y)}Z`
	);
}

/**
 * Span is the length of something between two points of a line, drawn as a
 * line with an arrowhead at each end and its label on a plate in the middle,
 * or where at says when the middle would leave the plate past an edge.
 */
export function Span({
	from,
	to,
	y,
	text,
	at = (from + to) / 2,
}: {
	from: number;
	to: number;
	y: number;
	text: string;
	at?: number;
}) {
	return (
		<>
			<path
				d={`M${r1(from)} ${r1(y)}L${r1(to)} ${r1(y)}`}
				class="mt-pic-line"
				stroke-width={1.5}
			/>
			<path
				d={arrowhead(from, y, 180) + arrowhead(to, y, 0)}
				class="mt-pic-ink"
			/>
			<Plate x={at} y={y} text={text} />
		</>
	);
}

/** plateWidth is how wide a Plate of a label is drawn. */
export function plateWidth(text: string, size: number = sizes.label): number {
	return widthOf(text, size) + 6;
}

/**
 * towards is the offset of a point a distance away in a direction, in degrees
 * clockwise from twelve.
 */
export function towards(
	degrees: number,
	distance: number,
): { x: number; y: number } {
	const radians = (degrees * Math.PI) / 180;
	return { x: distance * Math.sin(radians), y: -distance * Math.cos(radians) };
}

/**
 * spanHeight is how tall a Span is drawn, over its line and under it, for a
 * picture to leave room for it.
 */
export const spanHeight = sizes.label * 1.4;

/**
 * Brace is a curly brace under a stretch of a line, from one point to
 * another, opening upward toward what it holds.
 */
export function Brace({
	from,
	to,
	y,
}: {
	from: number;
	to: number;
	y: number;
}) {
	const middle = (from + to) / 2;
	const curl = Math.max(0, Math.min(7, (to - from) / 4));
	const d =
		`M${r1(from)} ${r1(y)}` +
		`Q${r1(from)} ${r1(y + curl)} ${r1(from + curl)} ${r1(y + curl)}` +
		`L${r1(middle - curl)} ${r1(y + curl)}` +
		`Q${r1(middle)} ${r1(y + curl)} ${r1(middle)} ${r1(y + 2 * curl)}` +
		`Q${r1(middle)} ${r1(y + curl)} ${r1(middle + curl)} ${r1(y + curl)}` +
		`L${r1(to - curl)} ${r1(y + curl)}` +
		`Q${r1(to)} ${r1(y + curl)} ${r1(to)} ${r1(y)}`;
	return <path d={d} class="mt-pic-line" stroke-width={1.5} />;
}

/** braceHeight is how far a Brace reaches under the line it stands on. */
export const braceHeight = 14;

/**
 * Lined is a label to lay out among others on a line: what tells it apart
 * from them, where its middle wants to stand, and what it says, at what size,
 * how strongly and in what tone.
 */
export type Lined = {
	id: string;
	x: number;
	text: string;
	size?: number;
	strong?: boolean;
	tone?: string | undefined;
};

/** lineHeight is how far apart the lines of a stack of labels stand, at a size. */
export function lineHeight(size: number = sizes.label): number {
	return size * 1.4;
}

/**
 * stackOf lays labels out on as few lines as keep them apart, each moved in
 * where it would stand past from or to, and says how many lines they take.
 */
export function stackOf(
	labels: readonly Lined[],
	from: number,
	to: number,
): { placed: (Lined & Placed)[]; lines: number } {
	const placed = stacked(
		labels.map((one) => ({
			x: one.x,
			width: widthOf(one.text, one.size ?? sizes.label),
		})),
		from,
		to,
	).map((spot, at) => ({ ...(labels[at] as Lined), ...spot }));
	return { placed, lines: lines(placed) };
}

/**
 * Stack draws labels laid out by stackOf: line 0 centred on first, and each
 * further line a line's height further down, or further up.
 */
export function Stack({
	placed,
	first,
	upward = false,
	size = sizes.label,
}: {
	placed: readonly (Lined & Placed)[];
	first: number;
	upward?: boolean;
	size?: number;
}) {
	const step = upward ? -lineHeight(size) : lineHeight(size);
	return (
		<>
			{placed.map((one) => (
				<Label
					key={one.id}
					x={one.x}
					y={first + one.line * step}
					text={one.text}
					size={one.size ?? size}
					strong={one.strong ?? true}
					tone={one.tone}
				/>
			))}
		</>
	);
}
