import type { Stairs } from "./model";
import { arrowhead, type Drawn, Label, lineHeight, r1, sizes } from "./shapes";
import { room, widthOf, within, written } from "./text";

// A staircase, in the card's pixels: the tallest a step may be, how far a
// label stands above its step, how high a jump's arc rises over the higher of
// its two steps, and how far apart two labels are kept.
const tallestStep = 18;
const overStep = 4;
const arcRise = 22;
const clear = 4;

// Box is the room a label takes, by its left and right ends and its middle's
// height above the floor.
type Box = { left: number; right: number; y: number };

/**
 * drawStairs draws a staircase climbing to the right from the floor, step 0,
 * to its top step, its steps unnumbered: each label above its step, moved in
 * where it would stand past an edge and up, a line at a time, where it would
 * touch a label set before it, and each jump an arc from the step it leaves
 * to the step it lands on, above their labels, with an arrowhead where it
 * lands.
 */
export function drawStairs(stairs: Stairs): Drawn {
	const steps = stairs.steps;
	const tread = room / (steps + 1);
	const rise = Math.min(tallestStep, tread);
	// Heights are first taken from the floor, upward, and turned into the
	// picture's own, downward from its top, once the top is known.
	const height = (step: number) => step * rise;
	const middle = (step: number) => (step + 0.5) * tread;
	const width = (steps + 1) * tread;
	const labels = placedLabels(stairs, tread, height, width);
	// A jump leaves its step, and lands on it, above the step's mark, if the
	// step has one, so that its arc never runs through the mark.
	const over = new Map(
		labels.map((one) => [one.step, one.box.y + sizes.label / 2 + overStep]),
	);
	const end = (step: number) => over.get(step) ?? height(step);
	const jumps = (stairs.jumps ?? []).map(({ from, to }) => ({
		from,
		to,
		lift: Math.max(end(from), end(to)) + arcRise,
	}));
	const top = Math.max(
		height(steps),
		...labels.map((one) => one.box.y + sizes.label / 2),
		// A jump is held by its arc's control point, which stands above it.
		...jumps.map((one) => one.lift),
	);
	const y = (above: number) => top - above;
	let outline = `M0 ${r1(y(0))}`;
	for (let step = 0; step <= steps; step++) {
		outline += `H${r1((step + 1) * tread)}`;
		if (step < steps) {
			outline += `V${r1(y(height(step + 1)))}`;
		}
	}
	outline += `V${r1(y(0))}Z`;
	return {
		width,
		height: top,
		body: (
			<>
				<path d={outline} class="mt-pic-fill" />
				<path d={outline} class="mt-pic-line" stroke-width={1.5} />
				{jumps.map(({ from, to, lift }) => {
					const x1 = middle(from);
					const x2 = middle(to);
					const y1 = y(end(from));
					const y2 = y(end(to));
					const control = { x: (x1 + x2) / 2, y: y(lift) };
					const landing =
						(Math.atan2(y2 - control.y, x2 - control.x) * 180) / Math.PI;
					return (
						<g key={`jump ${from} ${to}`}>
							<path
								d={`M${r1(x1)} ${r1(y1)}Q${r1(control.x)} ${r1(control.y)} ${r1(x2)} ${r1(y2)}`}
								class="mt-pic-line"
								stroke-width={1.5}
							/>
							<path d={arrowhead(x2, y2, landing, 7, 4)} class="mt-pic-ink" />
						</g>
					);
				})}
				{labels.map(({ step, text, box }) => (
					<Label
						key={`mark ${step}`}
						x={(box.left + box.right) / 2}
						y={y(box.y)}
						text={text}
					/>
				))}
			</>
		),
	};
}

// placedLabels are the labels of a staircase's marks, each above its step,
// moved in to stay within the picture, clear of the steps under it, and
// raised a line at a time until it touches no label placed before it.
function placedLabels(
	stairs: Stairs,
	tread: number,
	height: (step: number) => number,
	width: number,
): { step: number; text: string; box: Box }[] {
	const placed: { step: number; text: string; box: Box }[] = [];
	const touches = (one: Box, other: Box) =>
		one.left < other.right + clear &&
		other.left < one.right + clear &&
		Math.abs(one.y - other.y) < lineHeight();
	for (const { step, label } of stairs.marks ?? []) {
		const text = written(label);
		const wide = widthOf(text, sizes.label);
		const at = within((step + 0.5) * tread, wide, 0, width);
		// The steps climb to the right, so the highest under a label is the
		// one under its right end.
		const under = Math.min(stairs.steps, Math.floor((at + wide / 2) / tread));
		const box = {
			left: at - wide / 2,
			right: at + wide / 2,
			y: height(Math.max(step, under)) + overStep + sizes.label / 2,
		};
		while (placed.some((other) => touches(box, other.box))) {
			box.y += lineHeight();
		}
		placed.push({ step, text, box });
	}
	return placed;
}
