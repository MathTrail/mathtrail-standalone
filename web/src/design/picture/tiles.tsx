import type { Tiles } from "./model";
import { type Drawn, Label, r1, sizes } from "./shapes";
import { fitted, room, squeezed, widthOf } from "./text";
import { jointOf, turnOf } from "./tiling";

// Tiles, in the card's pixels: the tallest a picture of them is drawn, how
// far a label is kept inside its tile's outline and apart from the labels
// beside it, and how far from a tile's middle its labels may stand, as shares
// of the way to its sides, the nearest tried first.
const tallest = 240;
const inside = 1.5;
const apart = 3;
const depths = [0.6, 0.65, 0.7, 0.75, 0.8];

type Point = { x: number; y: number };

// Placed is a tile laid out: its middle, and the direction its first side
// faces, in radians clockwise from the right.
type Placed = { middle: Point; turn: number };

// Box is the room a label takes about its middle, as the drawing estimates it.
type Box = { left: number; top: number; right: number; bottom: number };

/**
 * drawTiles draws identical tiles, each a regular polygon cut into triangles
 * from its middle, with a label in each triangle that has one. The first tile
 * faces left with its first side, and each tile after it is laid against the
 * side joint of the tile before, facing it with its own first side: so a row
 * runs on and a ring closes round, by the tiles' own geometry, with the
 * labels of each tile going clockwise from the side it shares with the tile
 * before it. The labels are written as large as every one of them stands
 * inside its tile's outline, toward its own side, apart from the others.
 */
export function drawTiles(tiles: Tiles): Drawn {
	const sides = tiles.sides;
	const placed = laidOut(sides, jointOf(tiles), tiles.tiles.length);
	// Every length so far is of a tile whose corners lie a unit from its middle.
	const corners = placed.flatMap((tile) =>
		Array.from({ length: sides }, (_, k) => cornerOf(tile, sides, k)),
	);
	const left = Math.min(...corners.map((one) => one.x));
	const top = Math.min(...corners.map((one) => one.y));
	const spanX = Math.max(...corners.map((one) => one.x)) - left;
	const spanY = Math.max(...corners.map((one) => one.y)) - top;
	const scale = Math.min(room / spanX, tallest / spanY);
	const at = (one: Point) => ({
		x: (one.x - left) * scale,
		y: (one.y - top) * scale,
	});
	const apothem = Math.cos(Math.PI / sides) * scale;
	// The nearest share of the way to its sides at which every label of every
	// tile fits, at a size, if any does.
	const depthAt = (size: number) =>
		depths.find((share) =>
			placed.every((tile, i) =>
				labelsFit(tiles.tiles[i] ?? [], tile.turn, sides, {
					apothem,
					depth: share * apothem,
					size,
				}),
			),
		);
	const size = fitted(
		[sizes.label, ...squeezed],
		(one) => depthAt(one) !== undefined,
	);
	const depth = (depthAt(size) ?? depths.at(-1) ?? 0) * apothem;
	return {
		width: spanX * scale,
		height: spanY * scale,
		body: (
			<>
				{placed.map((tile, i) => {
					const outline = Array.from({ length: sides }, (_, k) =>
						at(cornerOf(tile, sides, k)),
					);
					const middle = at(tile.middle);
					const spokes = outline
						.map(
							(one) =>
								`M${r1(middle.x)} ${r1(middle.y)}L${r1(one.x)} ${r1(one.y)}`,
						)
						.join("");
					return (
						<g key={`tile ${r1(middle.x)} ${r1(middle.y)}`}>
							<path
								d={closedPath(outline)}
								class="mt-pic-line"
								stroke-width={2}
							/>
							<path d={spokes} class="mt-pic-line" stroke-width={1} />
							{(tiles.tiles[i] ?? []).map((text, k) => {
								const spot = toward(middle, sideAngle(tile, sides, k), depth);
								return text === "" ? null : (
									<Label
										key={`label ${r1(spot.x)} ${r1(spot.y)}`}
										{...spot}
										text={text}
										size={size}
									/>
								);
							})}
						</g>
					);
				})}
			</>
		),
	};
}

// laidOut are the middles and the turns of as many tiles as given, the first
// facing left with its first side, each after it laid against the side joint
// of the one before and turned as turnOf says.
function laidOut(sides: number, joint: number, count: number): Placed[] {
	const across = 2 * Math.cos(Math.PI / sides);
	const step = (turnOf(sides, joint) * 2 * Math.PI) / sides;
	const placed: Placed[] = [{ middle: { x: 0, y: 0 }, turn: Math.PI }];
	for (let i = 1; i < count; i++) {
		const before = placed[i - 1] as Placed;
		placed.push({
			middle: toward(before.middle, sideAngle(before, sides, joint), across),
			turn: before.turn + step,
		});
	}
	return placed;
}

// sideAngle is the direction side k of a tile faces from its middle.
function sideAngle(tile: Placed, sides: number, k: number): number {
	return tile.turn + (k * 2 * Math.PI) / sides;
}

// cornerOf is the corner of a tile where its side k ends and side k + 1
// begins, counted clockwise from its first side.
function cornerOf(tile: Placed, sides: number, k: number): Point {
	return toward(tile.middle, sideAngle(tile, sides, k + 0.5), 1);
}

// toward is the point a distance from another in a direction.
function toward(from: Point, angle: number, distance: number): Point {
	return {
		x: from.x + distance * Math.cos(angle),
		y: from.y + distance * Math.sin(angle),
	};
}

// closedPath is the outline through some points, closed back to the first.
function closedPath(points: readonly Point[]): string {
	const through = points.map((one) => `${r1(one.x)} ${r1(one.y)}`);
	return `M${through.join("L")}Z`;
}

// labelsFit says whether the labels of a tile turned as given, each written
// at a size as far from the tile's middle as depth, toward its own side,
// stand inside the tile's outline and apart from each other.
function labelsFit(
	labels: readonly string[],
	turn: number,
	sides: number,
	{ apothem, depth, size }: { apothem: number; depth: number; size: number },
): boolean {
	const tile = { middle: { x: 0, y: 0 }, turn };
	const boxes: Box[] = labels.flatMap((text, k) => {
		if (text === "") {
			return [];
		}
		const spot = toward(tile.middle, sideAngle(tile, sides, k), depth);
		const half = widthOf(text, size) / 2;
		return [
			{
				left: spot.x - half,
				right: spot.x + half,
				top: spot.y - size / 2,
				bottom: spot.y + size / 2,
			},
		];
	});
	const normals = Array.from({ length: sides }, (_, k) =>
		sideAngle(tile, sides, k),
	);
	// A box is inside the outline when each of its corners is on the inner
	// side of every side's line, by the room kept inside.
	const within = (box: Box) =>
		[
			[box.left, box.top],
			[box.right, box.top],
			[box.left, box.bottom],
			[box.right, box.bottom],
		].every(([x = 0, y = 0]) =>
			normals.every(
				(angle) =>
					x * Math.cos(angle) + y * Math.sin(angle) <= apothem - inside,
			),
		);
	const clear = (one: Box, other: Box) =>
		one.right + apart <= other.left ||
		other.right + apart <= one.left ||
		one.bottom + apart <= other.top ||
		other.bottom + apart <= one.top;
	return boxes.every(
		(one, at) =>
			within(one) && boxes.slice(at + 1).every((other) => clear(one, other)),
	);
}
