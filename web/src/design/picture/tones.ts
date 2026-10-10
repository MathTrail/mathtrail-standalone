/**
 * Tone is a colour a page may light one part of a picture in, beside the
 * card's own: cool and warm for two kinds of thing side by side, struck for a
 * part taken away, picked for the part the answer is, and start for where
 * something begins. A card lights no part, and its pictures keep its colours;
 * the site lights the parts of the pictures that show a solution step by
 * step, and its stylesheet says what each tone looks like.
 */
export type Tone = "cool" | "warm" | "struck" | "picked" | "start";

/**
 * Tones are the tones of a picture's parts, a part lit in one tone or more,
 * by the name its kind gives the part: a bar's piece as "piece <bar>.<piece>",
 * a row's item as "item <place>" and its line as "line", a table's cell as
 * "cell <row>.<column>" counted from its first row under the header, a grid's
 * cell as "cell <name>", and a number line's tick and mark as "tick <number>"
 * and "mark <number>". A kind that names no parts lights none.
 */
export type Tones = Readonly<Record<string, readonly Tone[]>>;

/**
 * toneOf is what the part called key is lit in, written as the data-tone of
 * its shapes, or nothing for a part left in the card's colours.
 */
export function toneOf(
	tones: Tones | undefined,
	key: string,
): string | undefined {
	const lit = tones?.[key];
	return lit === undefined || lit.length === 0 ? undefined : lit.join(" ");
}
