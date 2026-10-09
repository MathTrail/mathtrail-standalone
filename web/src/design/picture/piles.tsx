import type { Pile, Piles } from "./model";
import {
	type Drawn,
	Label,
	lineHeight,
	r1,
	Stack,
	sizes,
	stackOf,
} from "./shapes";
import { room, widest, written } from "./text";

// Piles, in the card's pixels: how big a counter is, how far apart the
// middles of two counters side by side stand, the room between two groups,
// how wide the ellipsis of a pile cut short is, how tall a line of counters
// is, the room between piles and around what is beside them, the room inside
// a box, and how round its corners are.
const counterRadius = 5.5;
const pitch = 14;
const groupGap = 6;
const ellipsisWidth = 20;
const lineTall = 16;
const pileGap = 10;
const aside = 8;
const padding = 5;
const corner = 5;

// widestColumn is how wide a pile set beside others may be at most.
const widestColumn = 76;

// firstLineTall is how tall the first line of a pile laid out under another
// is: as tall as its label's line, which stands beside it.
const firstLineTall = Math.max(lineTall, lineHeight());

// Token is one thing of a pile as it is laid out: a counter, a gap between
// two groups of counters, or the ellipsis where a pile is cut short.
type Token = "counter" | "gap" | "ellipsis";

// Laid is a counter or an ellipsis laid out: its place in the pile, its left
// edge and the line it is on.
type Laid = {
	at: number;
	token: "counter" | "ellipsis";
	x: number;
	line: number;
};

// Flow is a pile's counters laid out in lines: each counter and ellipsis, how
// many lines they take, and how far the widest line reaches.
type Flow = { laid: Laid[]; lines: number; widest: number };

// tokensOf are what a pile shows, in its order: a pile cut short shows half
// the counters it shows, an ellipsis and the other half; a heap of no stated
// size shows one counter, an ellipsis and one more; any other pile shows its
// every counter, with a gap after every group of them.
function tokensOf(pile: Pile): Token[] {
	if (pile.count === undefined) {
		return ["counter", "ellipsis", "counter"];
	}
	if (pile.shown !== undefined) {
		const first = Math.ceil(pile.shown / 2);
		return [
			...Array<Token>(first).fill("counter"),
			"ellipsis",
			...Array<Token>(pile.shown - first).fill("counter"),
		];
	}
	const tokens: Token[] = [];
	for (let counter = 0; counter < pile.count; counter++) {
		if (pile.group !== undefined && counter > 0 && counter % pile.group === 0) {
			tokens.push("gap");
		}
		tokens.push("counter");
	}
	return tokens;
}

// widthOfGroup is how wide the group of counters that starts at a place of a
// pile is drawn: up to the next gap or ellipsis.
function widthOfGroup(tokens: readonly Token[], from: number): number {
	let counters = 0;
	for (let at = from; at < tokens.length && tokens[at] === "counter"; at++) {
		counters++;
	}
	return (counters - 1) * pitch + 2 * counterRadius;
}

// flow lays a pile's counters out in lines no wider than given, breaking a
// line between groups where a group would not fit the rest of it.
function flow(tokens: readonly Token[], width: number): Flow {
	const laid: Laid[] = [];
	let x = 0;
	let line = 0;
	let reach = 0;
	const breakLine = () => {
		x = 0;
		line++;
	};
	tokens.forEach((token, at) => {
		if (token === "gap") {
			const group = widthOfGroup(tokens, at + 1);
			if (x + groupGap + group > width && group <= width) {
				breakLine();
			} else {
				x += groupGap;
			}
			return;
		}
		const wide =
			token === "counter"
				? 2 * counterRadius
				: ellipsisWidth - (pitch - 2 * counterRadius);
		if (x > 0 && x + wide > width) {
			breakLine();
		}
		laid.push({ at, token, x, line });
		reach = Math.max(reach, x + wide);
		x += token === "counter" ? pitch : ellipsisWidth;
	});
	return { laid, lines: line + 1, widest: reach };
}

// Counter is one counter of a pile, its left edge at x and its middle at y:
// a dot or a square, dark or light.
function Counter({ pile, x, y }: { pile: Pile; x: number; y: number }) {
	const light = pile.fill === "light";
	const inset = light ? 0.75 : 0;
	const radius = counterRadius - inset;
	const className = light ? "mt-pic-ring" : "mt-pic-ink";
	const width = light ? 1.5 : undefined;
	if (pile.shape === "square") {
		return (
			<rect
				x={r1(x + inset)}
				y={r1(y - radius)}
				width={r1(2 * radius)}
				height={r1(2 * radius)}
				class={className}
				stroke-width={width}
			/>
		);
	}
	return (
		<circle
			cx={r1(x + counterRadius)}
			cy={r1(y)}
			r={r1(radius)}
			class={className}
			stroke-width={width}
		/>
	);
}

// Dots are three small dots in a row or a column, centred on a point: where
// a pile is cut short, and where a skip stands among the piles.
function Dots({
	x,
	y,
	down = false,
}: {
	x: number;
	y: number;
	down?: boolean;
}) {
	return (
		<>
			{[-5, 0, 5].map((offset) => (
				<circle
					key={offset}
					cx={r1(down ? x : x + offset)}
					cy={r1(down ? y + offset : y)}
					r={1.6}
					class="mt-pic-ink"
				/>
			))}
		</>
	);
}

// Counters are a pile's counters laid out, from a left edge and the middle of
// their first line; a pile of none is an empty place, outlined in dashes.
function Counters({
	pile,
	flowed,
	left,
	top,
}: {
	pile: Pile;
	flowed: Flow;
	left: number;
	top: number;
}) {
	if (pile.count === 0) {
		return (
			<rect
				x={r1(left)}
				y={r1(top - counterRadius - 2)}
				width={r1(flowed.widest)}
				height={2 * counterRadius + 4}
				rx={3}
				class="mt-pic-line"
				stroke-width={1.5}
				stroke-dasharray="3 3"
			/>
		);
	}
	return (
		<>
			{flowed.laid.map((one) =>
				one.token === "counter" ? (
					<Counter
						key={one.at}
						pile={pile}
						x={left + one.x}
						y={top + one.line * lineTall}
					/>
				) : (
					<Dots
						key={one.at}
						x={left + one.x + (ellipsisWidth - (pitch - 2 * counterRadius)) / 2}
						y={top + one.line * lineTall}
					/>
				),
			)}
		</>
	);
}

// emptyWidth is how wide a pile of no counters is drawn at most: its dashed
// place.
const emptyWidth = 3 * pitch;

// flowOf lays a pile out in lines no wider than given; a pile of no counters
// takes one line, as wide as its place.
function flowOf(pile: Pile, width: number): Flow {
	return pile.count === 0
		? { laid: [], lines: 1, widest: Math.min(emptyWidth, width) }
		: flow(tokensOf(pile), width);
}

// Box is a box around something, from its left and top to its right and
// bottom.
function Box({
	left,
	top,
	right,
	bottom,
}: {
	left: number;
	top: number;
	right: number;
	bottom: number;
}) {
	return (
		<rect
			x={r1(left)}
			y={r1(top)}
			width={r1(right - left)}
			height={r1(bottom - top)}
			rx={corner}
			class="mt-pic-line"
			stroke-width={1.5}
		/>
	);
}

// widestOf is how wide the widest of the piles' labels, or of their values,
// is written, and nothing where none has one.
function widestOf(texts: readonly (string | undefined)[]): number {
	return widest(
		texts.flatMap((text) => (text === undefined ? [] : [written(text)])),
		sizes.label,
	);
}

// downward lays piles out one under another: each pile's label before it, its
// counters in as many lines as they take, and its value after its last line.
function downward(piles: Piles): Drawn {
	const outer = piles.box === true ? padding + 1 : 0;
	const labels = widestOf(piles.piles.map((pile) => pile.label));
	const values = widestOf(piles.piles.map((pile) => pile.value));
	const left = outer + (labels > 0 ? labels + aside : 0);
	const space = room - left - (values > 0 ? values + aside : 0) - outer;
	let top = outer;
	let right = left;
	const laid = piles.piles.map((pile, place) => {
		const inner = pile.boxed === true ? padding : 0;
		const flowed = flowOf(pile, space - 2 * inner);
		const at = { pile, place, flowed, top, inner };
		top += firstLineTall + (flowed.lines - 1) * lineTall + 2 * inner + pileGap;
		right = Math.max(right, left + flowed.widest + 2 * inner);
		return at;
	});
	const bottom = top - pileGap + outer;
	const width = Math.min(
		room,
		right + (values > 0 ? values + aside : 0) + outer,
	);
	return {
		width,
		height: bottom,
		body: (
			<>
				{piles.box === true && (
					<Box
						left={0.75}
						top={0.75}
						right={width - 0.75}
						bottom={bottom - 0.75}
					/>
				)}
				{laid.map(({ pile, place, flowed, top: pileTop, inner }) => {
					const firstLine = pileTop + inner + firstLineTall / 2;
					const lastLine = firstLine + (flowed.lines - 1) * lineTall;
					const end = left + flowed.widest + 2 * inner;
					if (pile.skip === true) {
						return <Dots key={place} x={left + 2 * pitch} y={firstLine} down />;
					}
					return (
						<g key={place}>
							{pile.label !== undefined && (
								<Label
									x={left - aside}
									y={firstLine}
									text={written(pile.label)}
									anchor="end"
								/>
							)}
							{pile.boxed === true && (
								<Box
									left={left}
									top={pileTop}
									right={end}
									bottom={lastLine + firstLineTall / 2 + inner}
								/>
							)}
							<Counters
								pile={pile}
								flowed={flowed}
								left={left + inner}
								top={firstLine}
							/>
							{pile.value !== undefined && (
								<Label
									x={end + aside}
									y={lastLine}
									text={written(pile.value)}
									anchor="start"
								/>
							)}
						</g>
					);
				})}
			</>
		),
	};
}

// across lays piles out side by side: each a column of counters, its label
// over it and its value under it, on as many lines as keep them apart, and as
// many counters to a line as its column holds.
function across(piles: Piles): Drawn {
	const outer = piles.box === true ? padding + 1 : 0;
	const count = piles.piles.length;
	const column = Math.min(
		widestColumn,
		(room - 2 * outer - (count - 1) * pileGap) / count,
	);
	const width = 2 * outer + count * column + (count - 1) * pileGap;
	const placed = piles.piles.map((pile, place) => {
		const inner = pile.boxed === true ? padding : 0;
		const fits = Math.max(
			1,
			Math.floor((column - 2 * inner + pitch - 2 * counterRadius) / pitch),
		);
		const flowed = flowOf(pile, fits * pitch - (pitch - 2 * counterRadius));
		return {
			pile,
			place,
			flowed,
			inner,
			middle: outer + place * (column + pileGap) + column / 2,
		};
	});
	const named = (text: (pile: Pile) => string | undefined) =>
		stackOf(
			placed.flatMap(({ pile, place, middle }) => {
				const said = pile.skip === true ? undefined : text(pile);
				return said === undefined
					? []
					: [{ id: `${place}`, x: middle, text: written(said) }];
			}),
			outer,
			width - outer,
		);
	const labels = named((pile) => pile.label);
	const values = named((pile) => pile.value);
	const top = outer + (labels.lines > 0 ? labels.lines * lineHeight() + 2 : 0);
	const tallest = Math.max(
		...placed.map(({ flowed, inner }) => flowed.lines * lineTall + 2 * inner),
	);
	const bottomOfCounters = top + tallest;
	const bottom =
		bottomOfCounters +
		(values.lines > 0 ? values.lines * lineHeight() + 2 : 0) +
		outer;
	return {
		width,
		height: bottom,
		body: (
			<>
				{piles.box === true && (
					<Box
						left={0.75}
						top={0.75}
						right={width - 0.75}
						bottom={bottom - 0.75}
					/>
				)}
				<Stack
					placed={labels.placed}
					first={top - 2 - lineHeight() / 2}
					upward
				/>
				<Stack
					placed={values.placed}
					first={bottomOfCounters + 2 + lineHeight() / 2}
				/>
				{placed.map(({ pile, place, flowed, middle, inner }) => {
					const start = middle - flowed.widest / 2;
					if (pile.skip === true) {
						return <Dots key={place} x={middle} y={top + tallest / 2} />;
					}
					return (
						<g key={place}>
							{pile.boxed === true && (
								<Box
									left={start - inner}
									top={top}
									right={start + flowed.widest + inner}
									bottom={top + flowed.lines * lineTall + 2 * inner}
								/>
							)}
							<Counters
								pile={pile}
								flowed={flowed}
								left={start}
								top={top + inner + lineTall / 2}
							/>
						</g>
					);
				})}
			</>
		),
	};
}

/**
 * drawPiles draws the piles of a game: one under another, or side by side
 * with across, all in one box with box. Each is a run of counters of its
 * shape and fill, a gap after every group, cut short to half the counters it
 * shows, an ellipsis and the other half where it says how many it shows; a
 * heap of no stated size is a counter, an ellipsis and a counter, and a pile
 * of none is an empty place. Its label stands before it or over it, its value
 * after it or under it, a pile in a box of its own with boxed, and a skip
 * among the piles is three dots. How many counters a pile has is never
 * written: the counters show it, or the wording says it.
 */
export function drawPiles(piles: Piles): Drawn {
	return piles.across === true ? across(piles) : downward(piles);
}
