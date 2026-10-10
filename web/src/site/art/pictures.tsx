import { drawPicture, PictureFrame } from "../../design/picture/diagram";
import type { Picture } from "../../design/picture/model";
import type { Tone, Tones } from "../../design/picture/tones";

/**
 * Kind is a picture of a kind the card draws, its parts lit in tones, as the
 * card draws it in a language.
 */
export function Kind({
	picture,
	locale,
	tones = {},
}: {
	picture: Picture;
	locale: string;
	tones?: Tones;
}) {
	return <PictureFrame drawn={drawPicture(picture, locale, tones)} />;
}

/** lit lights every part named in the same tones. */
export function lit(names: readonly string[], ...tones: Tone[]): Tones {
	return Object.fromEntries(names.map((name) => [name, tones]));
}

/**
 * pieces are the names of the pieces of a bar, by its place, from the first
 * piece up to the one before the last given.
 */
export function pieces(bar: number, count: number, from = 0): string[] {
	return Array.from(
		{ length: count - from },
		(_, at) => `piece ${bar}.${from + at}`,
	);
}
