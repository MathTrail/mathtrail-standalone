import type { Tiles } from "./model";

/**
 * jointOf is the side a tile shares with the tile after it, counted from its
 * first side, the one it shares with the tile before it: as the description
 * gives it, or else half its sides, rounded down.
 */
export function jointOf(tiles: Pick<Tiles, "sides" | "joint">): number {
	return tiles.joint ?? Math.floor(tiles.sides / 2);
}

/**
 * turnOf is how far each tile is turned from the one before it, in sides of a
 * tile, from less than half its sides back to half of them on: a tile laid
 * against side joint of the one before faces it with its own first side, so
 * it is turned half round and joint sides on. A row whose tiles turn by
 * nothing runs straight on, and a ring whose tiles' turns add up to as many
 * sides as a tile has goes round once.
 */
export function turnOf(sides: number, joint: number): number {
	const turn = (joint + sides / 2) % sides;
	return turn > sides / 2 ? turn - sides : turn;
}
