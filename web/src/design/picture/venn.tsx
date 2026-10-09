import type { Venn } from "./model";
import {
	type Drawn,
	Label,
	lineHeight,
	Plate,
	r1,
	Stack,
	sizes,
	stackOf,
} from "./shapes";
import { room, widthOf, written } from "./text";

// Two groups, in the card's pixels: how wide the picture is at most, how big
// each circle is, how far apart their middles stand, the room kept inside the
// box around them, and how round the box's corners are.
const widest = 280;
const circleRadius = 58;
const apartBy = 68;
const inset = 10;
const corner = 8;

// captionOf is what is written over a group: its label, and how many it holds
// where the description says.
function captionOf({ label, count }: Venn["sets"][number]): string {
	return count === undefined
		? written(label)
		: `${written(label)} ${written(count)}`;
}

/**
 * drawVenn draws two groups that overlap, inside a box of everyone the
 * question counts: each group a circle, its label and its count over its outer
 * half; how many are in both in the overlap, on a plate where it is wider than
 * the overlap; how many are in neither in the box's lower corner; and how many
 * there are in all on the box's top edge.
 */
export function drawVenn(venn: Venn): Drawn {
	const width = Math.min(room, widest);
	const middle = width / 2;
	const centres = [middle - apartBy / 2, middle + apartBy / 2] as const;
	const boxTop = venn.total === undefined ? 1 : (sizes.label * 1.4) / 2;
	const captions = stackOf(
		venn.sets.map((set, at) => ({
			id: `set-${at}`,
			x:
				at === 0
					? centres[0] - circleRadius / 2
					: centres[1] + circleRadius / 2,
			text: captionOf(set),
		})),
		inset,
		width - inset,
	);
	const captionsAt = boxTop + 6 + lineHeight() / 2;
	const circlesAt =
		captionsAt + (captions.lines - 0.5) * lineHeight() + 6 + circleRadius;
	const neitherAt = circlesAt + circleRadius + 3 + lineHeight() / 2;
	const boxBottom =
		venn.neither === undefined
			? circlesAt + circleRadius + inset
			: neitherAt + lineHeight() / 2 + 3;
	const lens = 2 * circleRadius - apartBy;
	const both = venn.both === undefined ? undefined : written(venn.both);
	const total = venn.total === undefined ? undefined : written(venn.total);
	return {
		width,
		height: boxBottom + 1,
		body: (
			<>
				<rect
					x={1}
					y={r1(boxTop)}
					width={r1(width - 2)}
					height={r1(boxBottom - boxTop)}
					rx={corner}
					class="mt-pic-line"
					stroke-width={1.5}
				/>
				{centres.map((x) => (
					<circle
						key={x}
						cx={r1(x)}
						cy={r1(circlesAt)}
						r={circleRadius}
						class="mt-pic-line"
						stroke-width={2}
					/>
				))}
				<Stack placed={captions.placed} first={captionsAt} />
				{both !== undefined &&
					(widthOf(both, sizes.label) + 6 <= lens ? (
						<Label x={middle} y={circlesAt} text={both} />
					) : (
						<Plate x={middle} y={circlesAt} text={both} />
					))}
				{venn.neither !== undefined && (
					<Label
						x={width - inset}
						y={neitherAt}
						text={written(venn.neither)}
						anchor="end"
					/>
				)}
				{total !== undefined && (
					<Plate
						x={width - inset - corner - widthOf(total, sizes.label) / 2 - 3}
						y={boxTop}
						text={total}
					/>
				)}
			</>
		),
	};
}
