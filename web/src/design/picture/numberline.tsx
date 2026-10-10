import type { NumberLine } from "./model";
import {
	arrowhead,
	type Drawn,
	Label,
	lineHeight,
	r1,
	Stack,
	sizes,
	stackOf,
} from "./shapes";
import {
	lines,
	room,
	smallest,
	stacked,
	widest,
	widthOf,
	written,
} from "./text";
import { type Tones, toneOf } from "./tones";

// A number line, in the card's pixels: how far apart two ticks may stand at
// most, how far a tick reaches over and under the line, how far the line runs
// on past its last tick to the tip of its arrowhead, and how tall and how wide
// a mark's pointer is.
const widestStep = 40;
const tickReach = 6;
const pastTheEnd = 18;
const pointerTall = 10;
const pointerHalf = 6;

// tilePad is how far the tile under a tick's number a page lights reaches
// past the number on either side, where the next tick leaves it the room.
const tilePad = 5;

// everyFew are how many ticks apart a crowded line numbers its ticks: every
// tick, every second, every fifth or every tenth.
const everyFew = [1, 2, 5, 10];

// numberSizes are the sizes the numbers of the ticks are written at, the
// smaller only where the larger crowds them.
const numberSizes = [sizes.number, smallest];

// clear is the least room between two numbers side by side.
const clear = 4;

// Tick is one tick of a line: its place among the ticks, counted from 0, its
// number as the card writes it, and where it stands.
type Tick = { place: number; text: string; x: number };

// apart says whether the numbers of two ticks, at a size, stand clear of each
// other.
function apart(one: Tick, other: Tick, size: number): boolean {
	return (
		Math.abs(one.x - other.x) >=
		(widthOf(one.text, size) + widthOf(other.text, size)) / 2 + clear
	);
}

// numbered are the ticks a line writes the numbers of, and the size it writes
// them at: every tick where their numbers stand clear of each other, else
// every second, fifth or tenth from the first, at the larger size where it
// can; and, whatever else is numbered, the two ends and every tick a mark
// stands on, which the child reads the task by. A tick of the regular run that
// would crowd one of those is left unnumbered.
function numbered(
	ticks: readonly Tick[],
	always: ReadonlySet<number>,
): { chosen: Tick[]; size: number } {
	const forced = ticks.filter((tick) => always.has(tick.place));
	for (const every of everyFew) {
		for (const size of numberSizes) {
			const regular = ticks.filter((tick) => tick.place % every === 0);
			const spaced = regular.every(
				(tick, at) => at === 0 || apart(regular[at - 1] as Tick, tick, size),
			);
			if (spaced) {
				const kept = new Set(
					regular.filter((tick) =>
						forced.every((other) => other === tick || apart(other, tick, size)),
					),
				);
				const chosen = ticks.filter(
					(tick) => always.has(tick.place) || kept.has(tick),
				);
				return { chosen, size };
			}
		}
	}
	return { chosen: forced, size: smallest };
}

/**
 * drawNumberLine draws a number line: its ticks evenly spaced, as far apart
 * as the card allows, and numbered by the card in Latin digits — every tick,
 * or, where the numbers crowd, every few, with both ends and every marked tick
 * always numbered and set on a second line where they crowd each other — the
 * line running on past its last tick to an arrowhead; and each mark a pointer
 * over its tick in the shading tone, with its label over it. A tick's number
 * a page lights is written on a tile of its tone, and a mark it lights is
 * drawn in its tone.
 */
export function drawNumberLine(
	line: NumberLine,
	_locale?: string,
	tones?: Tones,
): Drawn {
	const step = line.step ?? 1;
	const count = Math.round((line.to - line.from) / step) + 1;
	const texts = Array.from({ length: count }, (_, place) =>
		written(String(line.from + place * step)),
	);
	const placeOf = (at: number) => Math.round((at - line.from) / step);
	const start =
		Math.max(widthOf(texts[0] ?? "", sizes.number) / 2, pointerHalf) + 2;
	const end =
		Math.max(widthOf(texts[count - 1] ?? "", sizes.number) / 2, pastTheEnd) + 2;
	const gap = Math.min(widestStep, (room - start - end) / (count - 1));
	// A line shorter than the widest label of its marks stands in the middle
	// of a picture as wide as that label, so that the label stays in it.
	const own = start + (count - 1) * gap + end;
	const labels = (line.marks ?? []).flatMap((mark) =>
		mark.label === undefined ? [] : [written(mark.label)],
	);
	const shift = Math.max(0, widest(labels, sizes.label) - own) / 2;
	const width = own + 2 * shift;
	const origin = start + shift;
	const ticks = texts.map((text, place) => ({
		place,
		text,
		x: origin + place * gap,
	}));
	const marks = (line.marks ?? []).map((mark) => ({
		...mark,
		x: origin + placeOf(mark.at) * gap,
	}));
	const { chosen, size } = numbered(
		ticks,
		new Set([0, count - 1, ...marks.map((mark) => placeOf(mark.at))]),
	);
	const numbers = stacked(
		chosen.map((tick) => ({ x: tick.x, width: widthOf(tick.text, size) })),
		0,
		width,
		clear,
	).map((spot, at) => ({ ...(chosen[at] as Tick), ...spot }));
	const over = stackOf(
		marks.flatMap((mark) =>
			mark.label === undefined
				? []
				: [{ id: `mark-${mark.at}`, x: mark.x, text: written(mark.label) }],
		),
		0,
		width,
	);
	const pointerTop = over.lines * lineHeight() + (over.lines > 0 ? 3 : 0) + 2;
	const axis =
		(marks.length > 0 ? pointerTop + pointerTall + 3 : 2) + tickReach;
	const numbersAt = axis + tickReach + 3 + lineHeight(size) / 2;
	const height = numbersAt + (lines(numbers) - 0.5) * lineHeight(size) + 1;
	const last = origin + (count - 1) * gap;
	const pointerOf = (mark: { x: number }) =>
		`M${r1(mark.x - pointerHalf)} ${r1(pointerTop)}` +
		`L${r1(mark.x + pointerHalf)} ${r1(pointerTop)}` +
		`L${r1(mark.x)} ${r1(pointerTop + pointerTall)}Z`;
	const markTone = (mark: { at: number }) => toneOf(tones, `mark ${mark.at}`);
	const pointers = marks
		.filter((mark) => markTone(mark) === undefined)
		.map(pointerOf)
		.join("");
	const lit = marks.flatMap((mark) => {
		const tone = markTone(mark);
		return tone === undefined ? [] : [{ mark, tone }];
	});
	return {
		width,
		height,
		body: (
			<>
				<Stack
					placed={over.placed}
					first={(over.lines - 0.5) * lineHeight()}
					upward
				/>
				{pointers !== "" && (
					<>
						<path d={pointers} class="mt-pic-fill" />
						<path
							d={pointers}
							class="mt-pic-line"
							stroke-width={1.5}
							stroke-linejoin="round"
						/>
					</>
				)}
				{lit.map(({ mark, tone }) => (
					<g key={mark.at}>
						<path d={pointerOf(mark)} class="mt-pic-fill" data-tone={tone} />
						<path
							d={pointerOf(mark)}
							class="mt-pic-line"
							stroke-width={1.5}
							stroke-linejoin="round"
							data-tone={tone}
						/>
					</g>
				))}
				<path
					d={`M${r1(Math.max(0, origin - 10))} ${r1(axis)}H${r1(last + pastTheEnd - 6)}`}
					class="mt-pic-line"
					stroke-width={2}
				/>
				<path
					d={arrowhead(last + pastTheEnd, axis, 0, 9, 5)}
					class="mt-pic-ink"
				/>
				<path
					d={ticks
						.map(
							(tick) =>
								`M${r1(tick.x)} ${r1(axis - tickReach)}V${r1(axis + tickReach)}`,
						)
						.join("")}
					class="mt-pic-line"
					stroke-width={1.5}
				/>
				{numbers.map((number) => {
					const y = numbersAt + number.line * lineHeight(size);
					const tone = toneOf(tones, `tick ${line.from + number.place * step}`);
					const text = widthOf(number.text, size);
					const tile = Math.min(
						text + 2 * tilePad,
						Math.max(text + 2, gap - 2),
					);
					return (
						<g key={number.place}>
							{tone !== undefined && (
								<rect
									x={r1(number.x - tile / 2)}
									y={r1(y - lineHeight(size) / 2)}
									width={r1(tile)}
									height={r1(lineHeight(size))}
									rx={4}
									class="mt-pic-fill"
									data-tone={tone}
								/>
							)}
							<Label
								x={number.x}
								y={y}
								text={number.text}
								size={size}
								strong={false}
								tone={tone}
							/>
						</g>
					);
				})}
			</>
		),
	};
}
