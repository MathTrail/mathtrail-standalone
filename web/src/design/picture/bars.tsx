import type { Bar, Bars } from "./model";
import {
	Brace,
	braceHeight,
	type Drawn,
	Label,
	type Lined,
	lineHeight,
	plateWidth,
	r1,
	Span,
	Stack,
	sizes,
	spanHeight,
	stackOf,
} from "./shapes";
import type { Placed } from "./text";
import {
	fitted,
	room,
	smallest,
	widest,
	widthOf,
	within,
	written,
	writtenNote,
} from "./text";

// A bar, in the card's pixels: how tall it is, how long one unit of length
// may be drawn at most, and the room around what is written beside it.
const barHeight = 28;
const widestUnit = 28;
const gap = 8;

// lengthOf is how long a bar is in the units all the bars share: its length
// where it gives one, else the sizes of its segments or the count of its
// parts, and one unit for a bar cut into nothing.
function lengthOf(bar: Bar): number {
	return (
		bar.length ??
		bar.segments?.reduce((sum, segment) => sum + segment.size, 0) ??
		bar.parts ??
		1
	);
}

// edgesOf are where the edges between a bar's parts or segments stand, from
// its start at 0 to its end at its drawn length: the points its braces span.
// Segments fill the bar in proportion to their sizes.
function edgesOf(bar: Bar, length: number): number[] {
	if (bar.segments !== undefined) {
		const whole = bar.segments.reduce((sum, segment) => sum + segment.size, 0);
		let reached = 0;
		return [
			0,
			...bar.segments.map(
				(segment) => (reached += (segment.size / whole) * length),
			),
		];
	}
	const parts = bar.parts ?? 1;
	return Array.from(
		{ length: parts + 1 },
		(_, before) => (before * length) / parts,
	);
}

// columnOf is how wide a column of labels beside the bars is: the widest of
// them with room to spare, or nothing where none has one.
function columnOf(texts: readonly (string | undefined)[]): number {
	const widths = texts.flatMap((text) =>
		text === undefined ? [] : [widthOf(written(text), sizes.label)],
	);
	return widths.length === 0 ? 0 : Math.max(...widths) + gap;
}

// Piece is one rectangle of a bar: a part or a segment, its place among them,
// and where it starts and ends along the bar.
type Piece = { at: number; from: number; end: number };

// Laid is a bar laid out with everything drawn beside it and under it, from
// the top it stands at to the height it takes.
type Laid = {
	bar: Bar;
	place: number;
	top: number;
	length: number;
	pieces: Piece[];
	inside: (Lined & Placed)[];
	outside: (Lined & Placed)[];
	braces: { key: string; from: number; to: number }[];
	braceLabels: (Lined & Placed)[];
	braceTop: number;
	labelsTop: number;
	spanAt: number;
	height: number;
};

// labelsOfSegments are the labels of a bar's segments, each where it fits: in
// its segment, or under the bar.
function labelsOfSegments(
	bar: Bar,
	place: number,
	start: number,
	pieces: readonly Piece[],
) {
	const inside: Lined[] = [];
	const outside: Lined[] = [];
	for (const piece of pieces) {
		const label = bar.segments?.[piece.at]?.label;
		if (label !== undefined) {
			const text = written(label);
			const lined = {
				id: `segment-${place}-${piece.at}`,
				x: start + (piece.from + piece.end) / 2,
				text,
			};
			(widthOf(text, sizes.label) + 6 <= piece.end - piece.from
				? inside
				: outside
			).push(lined);
		}
	}
	return { inside, outside };
}

// layOut lays one bar out from the top it stands at.
function layOut(
	bar: Bar,
	place: number,
	top: number,
	unit: number,
	start: number,
	width: number,
): Laid {
	const length = lengthOf(bar) * unit;
	const edges = edgesOf(bar, length);
	const pieces = edges
		.slice(1)
		.map((end, at) => ({ at, from: edges[at] ?? 0, end }));
	const segments = labelsOfSegments(bar, place, start, pieces);
	const outside = stackOf(segments.outside, 0, width);
	const braceTop = top + barHeight + outside.lines * lineHeight() + 2;
	const braces = (bar.braces ?? []).map((one, at) => ({
		key: `${at}`,
		from: start + (edges[one.from] ?? 0),
		to: start + (edges[one.to] ?? length),
		text: written(one.label),
	}));
	const braceLabels = stackOf(
		braces.map((one) => ({
			id: `brace-${place}-${one.key}`,
			x: (one.from + one.to) / 2,
			text: one.text,
		})),
		0,
		width,
	);
	const labelsTop = braceTop + braceHeight + 4 + lineHeight() / 2;
	const underBraces =
		braces.length === 0
			? braceTop
			: labelsTop + (braceLabels.lines - 0.5) * lineHeight();
	const spanAt = underBraces + spanHeight / 2 + 4;
	const bottom = bar.span === undefined ? underBraces : spanAt + spanHeight / 2;
	return {
		bar,
		place,
		top,
		length,
		pieces,
		inside: segments.inside.map((one) => ({ ...one, line: 0 })),
		outside: outside.placed,
		braces,
		braceLabels: braceLabels.placed,
		braceTop,
		labelsTop,
		spanAt,
		height: bottom - top + gap,
	};
}

// inset is how far a brace keeps in from the edges of what it holds: 2 px,
// or a quarter of a stretch too narrow for that.
function inset({ from, to }: { from: number; to: number }): number {
	return Math.min(2, (to - from) / 4);
}

// OneBar is one bar with everything drawn beside it and under it.
function OneBar({
	laid,
	start,
	width,
}: {
	laid: Laid;
	start: number;
	width: number;
}) {
	const { bar, top, length } = laid;
	const middle = top + barHeight / 2;
	return (
		<>
			{bar.label !== undefined && (
				<Label x={start / 2} y={middle} text={written(bar.label)} />
			)}
			{laid.pieces.map((piece) => (
				<g key={piece.at}>
					{piece.at < (bar.shaded ?? 0) && (
						<rect
							x={r1(start + piece.from)}
							y={r1(top)}
							width={r1(piece.end - piece.from)}
							height={barHeight}
							class="mt-pic-fill"
						/>
					)}
					<rect
						x={r1(start + piece.from)}
						y={r1(top)}
						width={r1(piece.end - piece.from)}
						height={barHeight}
						class="mt-pic-line"
						stroke-width={1.5}
					/>
				</g>
			))}
			<Stack placed={laid.inside} first={middle} />
			<Stack placed={laid.outside} first={top + barHeight + lineHeight() / 2} />
			{bar.value !== undefined && (
				<Label
					x={start + length + gap / 2}
					y={middle}
					text={written(bar.value)}
					anchor="start"
				/>
			)}
			{laid.braces.map((one) => (
				<Brace
					key={one.key}
					from={one.from + inset(one)}
					to={one.to - inset(one)}
					y={laid.braceTop}
				/>
			))}
			<Stack placed={laid.braceLabels} first={laid.labelsTop} />
			{bar.span !== undefined && (
				<Span
					from={start}
					to={start + length}
					y={laid.spanAt}
					text={written(bar.span)}
					at={within(
						start + length / 2,
						plateWidth(written(bar.span)),
						0,
						width,
					)}
				/>
			)}
		</>
	);
}

// noteSize is the size the notes under bars are written at: the largest at
// which the widest of them fits the card.
function noteSize(notes: readonly string[]): number {
	return fitted([sizes.label, 13, 12, smallest], (size) =>
		notes.every((note) => widthOf(note, size) <= room),
	);
}

// widestWord is how wide the widest thing written under or across the bars
// is: a note at its size, a brace's or a segment's label, or the plate of a
// bar's length. A picture is at least as wide, so that each stays in it.
function widestWord(
	bars: Bars,
	notes: readonly string[],
	size: number,
): number {
	const labels = bars.bars.flatMap((bar) => [
		...(bar.braces ?? []).map((one) => written(one.label)),
		...(bar.segments ?? []).flatMap((one) =>
			one.label === undefined ? [] : [written(one.label)],
		),
	]);
	const plates = bars.bars.flatMap((bar) =>
		bar.span === undefined ? [] : [plateWidth(written(bar.span))],
	);
	return Math.max(widest(notes, size), widest(labels, sizes.label), ...plates);
}

/**
 * drawBars draws quantities as bars to one scale: each bar's equal parts with
 * the shaded ones filled, or its segments of their own sizes, its label before
 * it and its value after it, the braces under it with their labels, the length
 * of the whole bar, and the notes under all the bars.
 */
export function drawBars(bars: Bars): Drawn {
	const start = columnOf(bars.bars.map((bar) => bar.label));
	const after = columnOf(bars.bars.map((bar) => bar.value));
	const longest = Math.max(...bars.bars.map(lengthOf));
	const unit = Math.min(widestUnit, (room - start - after) / longest);
	const notes = (bars.notes ?? []).map(writtenNote);
	const size = noteSize(notes);
	const width = Math.max(
		Math.min(room, start + longest * unit + after),
		widestWord(bars, notes, size),
	);
	let top = 2;
	const laids = bars.bars.map((bar, place) => {
		const laid = layOut(bar, place, top, unit, start, width);
		top += laid.height;
		return laid;
	});
	const placed = notes.map((text, at) => ({ at, text }));
	return {
		width,
		height: top + placed.length * lineHeight(size) + 2,
		body: (
			<>
				{laids.map((laid) => (
					<OneBar key={laid.place} laid={laid} start={start} width={width} />
				))}
				{placed.map((note) => (
					<Label
						key={note.at}
						x={0}
						y={top + (note.at + 0.5) * lineHeight(size)}
						text={note.text}
						size={size}
						anchor="start"
					/>
				))}
			</>
		),
	};
}
