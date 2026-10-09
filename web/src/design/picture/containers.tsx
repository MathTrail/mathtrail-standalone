import type { Container, Containers } from "./model";
import {
	type Drawn,
	Label,
	lineHeight,
	r1,
	Stack,
	sizes,
	stackOf,
} from "./shapes";
import { widest, written } from "./text";

// Containers, in the card's pixels: how wide a container is, the room between
// two, how tall the largest is drawn at most, how tall a unit of what one
// holds may be at most, how thick its walls are, how long a mark on its wall
// is, and the least room a number needs to be written in.
const containerWidth = 48;
const between = 22;
const tallest = 150;
const largestUnit = 30;
const wall = 2.5;
const markLength = 6;
const roomForANumber = 20;

// overTheTop says whether the number of what a container holds is written
// over the container, having room neither in its water nor above it.
function overTheTop(container: Container, unit: number): boolean {
	return (
		container.amount * unit < roomForANumber &&
		(container.capacity - container.amount) * unit < roomForANumber
	);
}

// amountAt is where the number of what a container holds is written: in the
// water where the water is deep enough, else above it where the container has
// room there, and else over the container.
function amountAt(
	container: Container,
	top: number,
	bottom: number,
	unit: number,
): number {
	const surface = bottom - container.amount * unit;
	if (container.amount * unit >= roomForANumber) {
		return (surface + bottom) / 2;
	}
	return overTheTop(container, unit)
		? top - roomForANumber / 2
		: surface - roomForANumber / 2;
}

/**
 * drawContainers draws containers to pour between, side by side on one floor:
 * each open at the top and as tall as it can hold, to one scale, with a mark
 * on its wall for every unit; the water in it in the shading tone, with the
 * number of what it holds written in the water and the number of what it can
 * hold under it, both in Latin digits; and its label over all the containers.
 */
export function drawContainers(containers: Containers): Drawn {
	const most = Math.max(
		...containers.items.map((container) => container.capacity),
	);
	const unit = Math.min(largestUnit, tallest / most);
	// Containers narrower than the widest of their labels stand in the middle
	// of a picture as wide as that label, so that the label stays in it.
	const own =
		containers.items.length * (containerWidth + between) - between + wall;
	const labels = containers.items.flatMap((container) =>
		container.label === undefined ? [] : [written(container.label)],
	);
	const shift = Math.max(0, widest(labels, sizes.label) - own) / 2;
	const width = own + 2 * shift;
	const xs = containers.items.map(
		(_, before) => shift + wall / 2 + before * (containerWidth + between),
	);
	const over = stackOf(
		containers.items.flatMap((container, place) =>
			container.label === undefined
				? []
				: [
						{
							id: `label-${place}`,
							x: (xs[place] ?? 0) + containerWidth / 2,
							text: written(container.label),
						},
					],
		),
		0,
		width,
	);
	const roomOver = Math.max(
		0,
		...containers.items.map((container) =>
			overTheTop(container, unit)
				? roomForANumber - (most - container.capacity) * unit
				: 0,
		),
	);
	const bottom =
		over.lines * lineHeight() +
		(over.lines > 0 ? 4 : 0) +
		roomOver +
		most * unit +
		2;
	const laid = containers.items.map((container, place) => ({
		place,
		container,
		x: xs[place] ?? 0,
		top: bottom - container.capacity * unit,
	}));
	return {
		width,
		height: bottom + wall / 2 + 4 + lineHeight(sizes.number),
		body: (
			<>
				<Stack
					placed={over.placed}
					first={(over.lines - 0.5) * lineHeight()}
					upward
				/>
				{laid.map(({ place, container, x, top }) => {
					let marks = "";
					for (let level = 1; level < container.capacity; level++) {
						marks += `M${r1(x)} ${r1(bottom - level * unit)}H${r1(x + markLength)}`;
					}
					return (
						<g key={place}>
							{container.amount > 0 && (
								<rect
									x={r1(x + wall / 2)}
									y={r1(bottom - container.amount * unit)}
									width={r1(containerWidth - wall)}
									height={r1(container.amount * unit - wall / 2)}
									class="mt-pic-fill"
								/>
							)}
							{marks !== "" && (
								<path d={marks} class="mt-pic-line" stroke-width={1} />
							)}
							<path
								d={`M${r1(x)} ${r1(top)}V${r1(bottom)}H${r1(x + containerWidth)}V${r1(top)}`}
								class="mt-pic-line"
								stroke-width={wall}
								stroke-linejoin="round"
							/>
							<Label
								x={x + containerWidth / 2}
								y={amountAt(container, top, bottom, unit)}
								text={String(container.amount)}
								size={sizes.number}
								strong={false}
							/>
							<Label
								x={x + containerWidth / 2}
								y={bottom + wall / 2 + 4 + lineHeight(sizes.number) / 2}
								text={String(container.capacity)}
								size={sizes.number}
								strong={false}
							/>
						</g>
					);
				})}
			</>
		),
	};
}
