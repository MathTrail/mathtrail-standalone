import type { Ring } from "./model";
import { arrowhead, type Drawn, Label, r1, sizes, towards } from "./shapes";
import { widthOf, written } from "./text";

// A ring, in the card's pixels: how big a place is, how much of the ring each
// place takes at least, how small the ring may be, how long the arrow from the
// start's label to its place is at least, and the room kept clear around the
// label.
const placeRadius = 11;
const arcPerPlace = 32;
const smallestRadius = 48;
const shortestArrow = 10;
const aroundTheLabel = 3;

// Point is a point of the picture.
type Point = { x: number; y: number };

// directionOf is the direction of a place of a ring, in degrees clockwise
// from twelve: the first place stands at twelve.
function directionOf(place: number, count: number): number {
	return ((place - 1) * 360) / count;
}

// halfOf is half the room a label takes, across it and down it.
function halfOf(text: string): Point {
	return {
		x: widthOf(text, sizes.label) / 2 + aroundTheLabel,
		y: (sizes.label * 1.4) / 2 + aroundTheLabel,
	};
}

// reach is how far a label at the middle of the ring reaches out in a
// direction: to the edge of the room it takes.
function reach(half: Point, degrees: number): number {
	const along = towards(degrees, 1);
	const across =
		Math.abs(along.x) < 1e-9
			? Number.POSITIVE_INFINITY
			: half.x / Math.abs(along.x);
	const down =
		Math.abs(along.y) < 1e-9
			? Number.POSITIVE_INFINITY
			: half.y / Math.abs(along.y);
	return Math.min(across, down);
}

// radiusOf is the radius of a ring: as small as gives every place its share
// of the ring, and large enough for the start's label to stand at its middle
// clear of every place, with an arrow to the start.
function radiusOf(ring: Ring): number {
	const shared = (ring.count * arcPerPlace) / (2 * Math.PI);
	const label = ring.start?.label;
	if (ring.start === undefined || label === undefined) {
		return Math.max(smallestRadius, shared);
	}
	const half = halfOf(written(label));
	const clearOfPlaces = Math.hypot(half.x, half.y) + placeRadius;
	const withArrow =
		reach(half, directionOf(ring.start.at, ring.count)) +
		shortestArrow +
		placeRadius +
		2;
	return Math.max(smallestRadius, shared, clearOfPlaces, withArrow);
}

/**
 * drawRing draws places in a ring: each a small circle numbered by the card in
 * Latin digits, from 1 at the top on round clockwise, the way the ring is
 * travelled; an arrowhead on the ring between every two places, pointing on;
 * the place to start at shaded, and its label at the middle of the ring with
 * an arrow to it.
 */
export function drawRing(ring: Ring): Drawn {
	const radius = radiusOf(ring);
	const middle = radius + placeRadius + 1;
	const at = (degrees: number, distance: number): Point => {
		const offset = towards(degrees, distance);
		return { x: middle + offset.x, y: middle + offset.y };
	};
	const places = Array.from({ length: ring.count }, (_, before) => {
		const number = before + 1;
		return { number, ...at(directionOf(number, ring.count), radius) };
	});
	const arrows = places
		.map(({ number }) => {
			const degrees = directionOf(number + 0.5, ring.count);
			const tip = at(degrees + 3 * (180 / Math.PI / radius), radius);
			return arrowhead(tip.x, tip.y, degrees, 6, 3.5);
		})
		.join("");
	const start = ring.start;
	const label = start?.label === undefined ? undefined : written(start.label);
	const startDirection =
		start === undefined ? 0 : directionOf(start.at, ring.count);
	const from = label === undefined ? 0 : reach(halfOf(label), startDirection);
	const tail = at(startDirection, from);
	const back = at(startDirection, radius - placeRadius - 8);
	const tip = at(startDirection, radius - placeRadius - 2);
	return {
		width: 2 * middle,
		height: 2 * middle,
		body: (
			<>
				<circle
					cx={r1(middle)}
					cy={r1(middle)}
					r={r1(radius)}
					class="mt-pic-line"
					stroke-width={1.5}
				/>
				<path d={arrows} class="mt-pic-ink" />
				{places.map((place) => (
					<g key={place.number}>
						<circle
							cx={r1(place.x)}
							cy={r1(place.y)}
							r={placeRadius}
							class={
								place.number === start?.at ? "mt-pic-fill" : "mt-pic-plate"
							}
						/>
						<circle
							cx={r1(place.x)}
							cy={r1(place.y)}
							r={placeRadius}
							class="mt-pic-line"
							stroke-width={1.5}
						/>
						<Label
							x={place.x}
							y={place.y}
							text={String(place.number)}
							size={sizes.number}
							strong={false}
						/>
					</g>
				))}
				{label !== undefined && (
					<>
						<Label x={middle} y={middle} text={label} />
						<path
							d={`M${r1(tail.x)} ${r1(tail.y)}L${r1(back.x)} ${r1(back.y)}`}
							class="mt-pic-line"
							stroke-width={1.5}
						/>
						<path
							d={arrowhead(tip.x, tip.y, startDirection - 90, 8, 4.5)}
							class="mt-pic-ink"
						/>
					</>
				)}
			</>
		),
	};
}
