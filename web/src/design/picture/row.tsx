import type { Row, RowItem } from "./model";
import {
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
import { room, widest, widthOf, written } from "./text";
import { type Tones, toneOf } from "./tones";

// The row, in the card's pixels: how big a mark is, how far apart two items
// may stand at most, and how far below the first row its copy runs.
const markRadius = 6;
const widestStep = 44;
const copyGap = 28;

// Spot is one item of a row and where it stands.
type Spot = { item: RowItem; place: number; x: number };

// endRoom is how far an item at an end of the row needs from the edge of the
// picture: half its widest label, and its mark at least.
function endRoom(item: RowItem | undefined): number {
	const label = written(item?.label ?? "");
	const below = written(item?.below ?? "");
	return Math.max(
		markRadius + 6,
		widthOf(label, sizes.label) / 2 + 2,
		widthOf(below, sizes.label) / 2 + 2,
	);
}

// spotsOf are where a row's items stand: evenly, as far apart as the card's
// width allows, up to widestStep.
function spotsOf(items: readonly RowItem[]): { spots: Spot[]; width: number } {
	const start = endRoom(items[0]);
	const end = endRoom(items.at(-1));
	const step = Math.min(widestStep, (room - start - end) / (items.length - 1));
	const spots = items.map((item, place) => ({
		item,
		place,
		x: start + place * step,
	}));
	return { spots, width: start + (items.length - 1) * step + end };
}

// Line is the line a row's marks stand on: solid between two items, and
// broken where a skip cuts the row short, in the tone a page lights it in.
function Line({
	spots,
	y,
	tone,
}: {
	spots: readonly Spot[];
	y: number;
	tone: string | undefined;
}) {
	let solid = "";
	let broken = "";
	for (const spot of spots.slice(1)) {
		const before = spots[spot.place - 1] as Spot;
		const piece = `M${r1(before.x)} ${r1(y)}L${r1(spot.x)} ${r1(y)}`;
		if (spot.item.skip === true || before.item.skip === true) {
			broken += piece;
		} else {
			solid += piece;
		}
	}
	return (
		<>
			{solid !== "" && (
				<path d={solid} class="mt-pic-line" stroke-width={2} data-tone={tone} />
			)}
			{broken !== "" && (
				<path
					d={broken}
					class="mt-pic-line"
					stroke-width={2}
					stroke-dasharray="2 5"
					data-tone={tone}
				/>
			)}
		</>
	);
}

// Mark is one item of a row on its line: a dot, a ring or a square, in the
// tone a page lights it in, or, where the row is drawn with no line, the
// ellipsis of a skip.
function Mark({
	spot,
	y,
	line,
	tone,
}: {
	spot: Spot;
	y: number;
	line: boolean;
	tone: string | undefined;
}) {
	const { item, x } = spot;
	if (item.skip === true) {
		return line ? null : <Label x={x} y={y} text="…" />;
	}
	switch (item.mark ?? "dot") {
		case "ring":
			return (
				<circle
					cx={r1(x)}
					cy={r1(y)}
					r={markRadius}
					class="mt-pic-ring"
					stroke-width={2}
					data-tone={tone}
				/>
			);
		case "square":
			return (
				<rect
					x={r1(x - markRadius)}
					y={r1(y - markRadius)}
					width={2 * markRadius}
					height={2 * markRadius}
					class="mt-pic-ink"
					data-tone={tone}
				/>
			);
		default:
			return (
				<circle
					cx={r1(x)}
					cy={r1(y)}
					r={markRadius}
					class="mt-pic-ink"
					data-tone={tone}
				/>
			);
	}
}

// above are the labels over a row's items.
function above(spots: readonly Spot[]): Lined[] {
	return spots.flatMap(({ item, place, x }) =>
		item.label === undefined
			? []
			: [{ id: `label-${place}`, x, text: written(item.label) }],
	);
}

// under are the labels under a row: under each item, and under each gap drawn
// between two items that are not skips, in the tone of the line.
function under(row: Row, spots: readonly Spot[], tones?: Tones): Lined[] {
	const labels: Lined[] = [];
	for (const { item, place, x } of spots) {
		if (item.below !== undefined) {
			labels.push({ id: `below-${place}`, x, text: written(item.below) });
		}
		const before = spots[place - 1];
		if (
			row.gaps !== undefined &&
			before !== undefined &&
			before.item.skip !== true &&
			item.skip !== true
		) {
			labels.push({
				id: `gap-${place}`,
				x: (before.x + x) / 2,
				text: written(row.gaps),
				tone: toneOf(tones, "line"),
			});
		}
	}
	return labels;
}

// widened is a row laid out no narrower than the widest of its labels and the
// plate of its length, its items moved to the middle of the room gained, so
// that every label stays within the picture.
function widened(
	spots: readonly Spot[],
	width: number,
	row: Row,
): { spots: Spot[]; width: number } {
	const texts = [...above(spots), ...under(row, spots)].map((one) => one.text);
	const need = Math.max(
		widest(texts, sizes.label),
		row.span === undefined ? 0 : plateWidth(written(row.span)),
	);
	const shift = Math.max(0, need - width) / 2;
	return {
		spots: spots.map((spot) => ({ ...spot, x: spot.x + shift })),
		width: width + 2 * shift,
	};
}

/**
 * drawRow draws objects in a row on a line, as far apart as the card allows:
 * the labels over them and under them on as many lines as keep them apart,
 * the label of every gap drawn, a stretch cut short where a skip stands, the
 * row again under itself as the other side of a path, and the length of the
 * whole row. An item a page lights, and the line, are drawn in their tones.
 */
export function drawRow(row: Row, _locale?: string, tones?: Tones): Drawn {
	const laid = spotsOf(row.items);
	const { spots, width } = widened(laid.spots, laid.width, row);
	const line = row.line !== false;
	const over = stackOf(above(spots), 0, width);
	const below = stackOf(under(row, spots, tones), 0, width);
	const first = 4 + over.lines * lineHeight() + markRadius + 4;
	const copies = Array.from(
		{ length: row.copies ?? 1 },
		(_, before) => first + before * copyGap,
	);
	const last = Math.max(first, ...copies);
	const underTop = last + markRadius + 4 + lineHeight() / 2;
	const spanAt =
		underTop -
		lineHeight() / 2 +
		below.lines * lineHeight() +
		spanHeight / 2 +
		4;
	const height =
		row.span === undefined
			? spanAt - spanHeight / 2
			: spanAt + spanHeight / 2 + 2;
	const xs = spots.map((spot) => spot.x);
	const ends = { from: Math.min(...xs), to: Math.max(...xs) };
	return {
		width,
		height,
		body: (
			<>
				<Stack
					placed={over.placed}
					first={first - markRadius - 4 - lineHeight() / 2}
					upward
				/>
				{copies.map((y) => (
					<g key={y}>
						{line && <Line spots={spots} y={y} tone={toneOf(tones, "line")} />}
						{spots.map((spot) => (
							<Mark
								key={spot.place}
								spot={spot}
								y={y}
								line={line}
								tone={toneOf(tones, `item ${spot.place}`)}
							/>
						))}
					</g>
				))}
				<Stack placed={below.placed} first={underTop} />
				{row.span !== undefined && (
					<Span
						from={ends.from}
						to={ends.to}
						y={spanAt}
						text={written(row.span)}
					/>
				)}
			</>
		),
	};
}
