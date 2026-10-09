import type { Clock } from "./model";
import { type Drawn, Label, Plate, r1, sizes, towards } from "./shapes";
import { widthOf, written } from "./text";

// The face, in the card's pixels: its middle, its rim, the ring its numbers
// stand on, and how long its hands are. The minute hand stops short of the
// numbers, with room between its tip and each of them at every minute: the
// two digits of 10, which it points across at ten to the hour, come nearest.
const middle = 100;
const rim = 94;
const numerals = 72;
const hourHand = 46;
const minuteHand = 56;
const hourWidth = 6;
const minuteWidth = 3.5;
// handLabelAside is how far a hand's label stands from its hand.
const handLabelAside = 16;

// twelve are the numbers of the face, in the order they go round.
const twelve = Array.from({ length: 12 }, (_, before) => before + 1);

/**
 * hands are where a clock's hands point at a time, in degrees clockwise from
 * twelve: the hour hand moved on by the minutes, and a time after 12:59 shown
 * as a twelve-hour face shows it.
 */
export function hands(time: string): { hour: number; minute: number } {
	const [hours = 0, minutes = 0] = time.split(":").map(Number);
	return { hour: (30 * (hours % 12) + minutes / 2) % 360, minute: 6 * minutes };
}

// at is the point a distance from the middle of the face in a direction.
function at(degrees: number, distance: number): { x: number; y: number } {
	const offset = towards(degrees, distance);
	return { x: middle + offset.x, y: middle + offset.y };
}

// ticks is the path of the face's minute marks: the long ones at the hours,
// or the short ones between them.
function ticks(long: boolean): string {
	let d = "";
	for (let minute = 0; minute < 60; minute++) {
		if ((minute % 5 === 0) === long) {
			const inner = at(minute * 6, long ? 84 : 87);
			const outer = at(minute * 6, 91);
			d += `M${r1(inner.x)} ${r1(inner.y)}L${r1(outer.x)} ${r1(outer.y)}`;
		}
	}
	return d;
}

// clockwiseFrom says whether a direction lies clockwise from another, the
// short way round.
function clockwiseFrom(from: number, to: number): boolean {
	const turn = (((to - from) % 360) + 360) % 360;
	return turn < 180;
}

// Hand is one hand of the clock, from the middle to its tip.
function Hand({
	direction,
	length,
	width,
}: {
	direction: number;
	length: number;
	width: number;
}) {
	const tip = at(direction, length);
	return (
		<path
			d={`M${middle} ${middle}L${r1(tip.x)} ${r1(tip.y)}`}
			class="mt-pic-line"
			stroke-width={width}
			stroke-linecap="round"
		/>
	);
}

// Box is the room a text takes: its left, top, right and bottom.
type Box = { left: number; top: number; right: number; bottom: number };

// boxOf is the room a text centred on a point takes at a size, with a little
// to spare all round.
function boxOf(
	point: { x: number; y: number },
	text: string,
	size: number,
): Box {
	const half = widthOf(text, size) / 2 + 3;
	const tall = (size * 1.4) / 2;
	return {
		left: point.x - half,
		top: point.y - tall,
		right: point.x + half,
		bottom: point.y + tall,
	};
}

// overlap says whether two boxes share any room.
function overlap(one: Box, other: Box): boolean {
	return (
		one.left < other.right &&
		other.left < one.right &&
		one.top < other.bottom &&
		other.top < one.bottom
	);
}

// numeralBoxes are the room the numbers of the face take.
const numeralBoxes = twelve.map((numeral) =>
	boxOf(at(numeral * 30, numerals), String(numeral), sizes.numeral),
);

// insideTheRing says whether a box stands clear of the face's minute marks.
function insideTheRing(box: Box): boolean {
	return [box.left, box.right].every((x) =>
		[box.top, box.bottom].every((y) => Math.hypot(x - middle, y - middle) < 82),
	);
}

// Hands are the clock's two hands as a label must stay clear of them: where
// each points, how long it is, and how wide it is drawn.
type Hands = readonly { direction: number; length: number; width: number }[];

// coversAHand says whether a box lies over any part of a hand.
function coversAHand(box: Box, hands: Hands): boolean {
	return hands.some(({ direction, length, width }) => {
		for (let along = 0; along <= length; along += 3) {
			const point = at(direction, along);
			const reach = width / 2;
			if (
				point.x > box.left - reach &&
				point.x < box.right + reach &&
				point.y > box.top - reach &&
				point.y < box.bottom + reach
			) {
				return true;
			}
		}
		return false;
	});
}

// troubleOf is how many things a label's box gets in the way of: the face's
// marks, each hand and each number or label it covers.
function troubleOf(box: Box, taken: readonly Box[], hands: Hands): number {
	return (
		(insideTheRing(box) ? 0 : 1) +
		hands.filter((hand) => coversAHand(box, [hand])).length +
		[...numeralBoxes, ...taken].filter((other) => overlap(other, box)).length
	);
}

// spotsBeside are the points a hand's label may stand at, in the order they
// are tried: on the side of the hand it is given, then on the other; nearest
// the hand first; and from near its tip in toward the middle of the face.
function spotsBeside(
	direction: number,
	length: number,
	clockwise: boolean,
): { x: number; y: number }[] {
	return [clockwise, !clockwise].flatMap((side) =>
		[handLabelAside, handLabelAside + 8, handLabelAside + 16].flatMap(
			(away) => {
				const aside = towards(direction + (side ? 90 : -90), away);
				return [0.62, 0.5, 0.38, 0.26, 0.14].map((along) => {
					const point = at(direction, length * along);
					return { x: point.x + aside.x, y: point.y + aside.y };
				});
			},
		),
	);
}

// besideTheHand is where a hand's label stands: beside the hand rather than on
// it, so that neither hand, which tell the time, is hidden; at the first spot
// that keeps it clear of the face's numbers and marks, of the hands and of the
// labels already placed. A label too wide to stand clear of them all stands
// where it gets in the way of fewest.
function besideTheHand(
	text: string,
	direction: number,
	length: number,
	clockwise: boolean,
	taken: readonly Box[],
	hands: Hands,
): { x: number; y: number; box: Box } {
	let best: { x: number; y: number; box: Box } | undefined;
	let fewest = Number.POSITIVE_INFINITY;
	for (const spot of spotsBeside(direction, length, clockwise)) {
		const box = boxOf(spot, text, sizes.label);
		const trouble = troubleOf(box, taken, hands);
		if (trouble < fewest) {
			best = { ...spot, box };
			fewest = trouble;
		}
		if (trouble === 0) {
			break;
		}
	}
	return best as { x: number; y: number; box: Box };
}

// TipLabels are the labels of the hands, each beside its hand and on the side
// of it away from the other hand where it can be.
function TipLabels({
	clock,
	hour,
	minute,
}: {
	clock: Clock;
	hour: number;
	minute: number;
}) {
	const minuteAhead = clockwiseFrom(hour, minute);
	const hands: Hands = [
		{ direction: hour, length: hourHand, width: hourWidth },
		{ direction: minute, length: minuteHand, width: minuteWidth },
	];
	const taken: Box[] = [];
	const placed: { at: number; text: string; x: number; y: number }[] = [];
	for (const [label, direction, length, side] of [
		[clock.minute_label, minute, minuteHand, minuteAhead],
		[clock.hour_label, hour, hourHand, !minuteAhead],
	] as const) {
		if (label !== undefined) {
			const text = written(label);
			const spot = besideTheHand(text, direction, length, side, taken, hands);
			taken.push(spot.box);
			placed.push({ at: placed.length, text, x: spot.x, y: spot.y });
		}
	}
	return (
		<>
			{placed.map((one) => (
				<Plate key={one.at} x={one.x} y={one.y} text={one.text} />
			))}
		</>
	);
}

/**
 * drawClock draws a clock face: its rim, its sixty minute marks, its numbers
 * 1 to 12, and its hands where the time puts them.
 */
export function drawClock(clock: Clock): Drawn {
	const { hour, minute } = hands(clock.time);
	return {
		width: 2 * middle,
		height: 2 * middle,
		body: (
			<>
				<circle
					cx={middle}
					cy={middle}
					r={rim}
					class="mt-pic-line"
					stroke-width={2.5}
				/>
				<path d={ticks(false)} class="mt-pic-line" stroke-width={1} />
				<path d={ticks(true)} class="mt-pic-line" stroke-width={2} />
				{twelve.map((numeral) => {
					const point = at(numeral * 30, numerals);
					return (
						<Label
							key={numeral}
							x={point.x}
							y={point.y}
							text={String(numeral)}
							size={sizes.numeral}
						/>
					);
				})}
				<Hand direction={hour} length={hourHand} width={hourWidth} />
				<Hand direction={minute} length={minuteHand} width={minuteWidth} />
				<circle cx={middle} cy={middle} r={4.5} class="mt-pic-ink" />
				<TipLabels clock={clock} hour={hour} minute={minute} />
			</>
		),
	};
}
