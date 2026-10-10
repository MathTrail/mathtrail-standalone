import { type Colors, type Paint, paints } from "./model";

/**
 * Painted is a colour a picture paints with, as the card writes it: the paint,
 * the word the lesson calls it by, and the letters written on the paint, which
 * tell it apart from the picture's other colours where the colours alone do
 * not — to a child who cannot tell them apart, on a screen that shows none.
 */
export type Painted = { paint: Paint; word: string; letters: string };

// mostLetters is how many letters a paint carries at most: as many as the
// narrowest stripe has room for.
const mostLetters = 3;

/**
 * paintedOf are the colours a picture paints with, in the order of the
 * palette, each with the letters its paint carries: the fewest letters of its
 * word, counted as a reader sees them, that no other word of the picture
 * shares, three at most, the first a capital where the language writes
 * capitals. A word of two is told by the start of its first word and the
 * first letter of its second, as xanh lá and xanh dương are by Xl and Xd.
 * Words that still share three letters are told by their first letter and
 * the first one where they part, as vermelho and verde are by Vm and Vd.
 */
export function paintedOf(
	colors: Colors | undefined,
	locale: string,
): Painted[] {
	const named = paints.flatMap((paint) => {
		const word = colors?.[paint];
		return word === undefined ? [] : [{ paint, word }];
	});
	const choices = named.map(({ word }) => choicesOf(word, locale));
	const spelled = named.map(({ word }) =>
		graphemesOf(word.replaceAll(" ", ""), locale),
	);
	return named.map(({ paint, word }, at) => {
		const own = choices[at] ?? [];
		const sharing = (count: number) =>
			choices.flatMap((other, there) =>
				there !== at &&
				folded(other, count, locale) === folded(own, count, locale)
					? [spelled[there] ?? []]
					: [],
			);
		let count = 1;
		while (
			count < mostLetters &&
			count < own.length &&
			sharing(count).length > 0
		) {
			count++;
		}
		const others = sharing(count);
		const letters =
			others.length > 0
				? (partedOf(spelled[at] ?? [], others, locale) ?? own[count - 1])
				: own[count - 1];
		return { paint, word, letters: capitalised(letters ?? [], locale) };
	});
}

// partedOf are a word's first letter and the first of its letters after it
// that each of the others has another letter in place of, or nothing where the
// word parts from them nowhere: where it is the start of one of them.
function partedOf(
	word: readonly string[],
	others: readonly (readonly string[])[],
	locale: string,
): string[] | undefined {
	const fold = (letter: string | undefined) =>
		letter?.toLocaleLowerCase(locale);
	for (let at = 1; at < word.length; at++) {
		if (others.every((other) => fold(other[at]) !== fold(word[at]))) {
			return [word[0] ?? "", word[at] ?? ""];
		}
	}
	return undefined;
}

// graphemesOf are the letters of a text as a reader sees them, each with the
// marks it carries.
function graphemesOf(text: string, locale: string): string[] {
	const segmenter = new Intl.Segmenter(locale, { granularity: "grapheme" });
	return [...segmenter.segment(text)].map((piece) => piece.segment);
}

// choicesOf are the letters a word's paint may carry, one letter more at each
// choice: of a word of one, its start; of a word of two, the start of its
// first word, from the second choice on with the first letter of its second
// word after it.
function choicesOf(word: string, locale: string): string[][] {
	const [first = "", second] = word.split(" ");
	const graphemes = (text: string) => graphemesOf(text, locale);
	const head = graphemes(first);
	if (second === undefined) {
		return head.map((_, at) => head.slice(0, at + 1));
	}
	const initial = graphemes(second)[0] ?? "";
	return head.map((_, at) =>
		at === 0 ? head.slice(0, 1) : [...head.slice(0, at), initial],
	);
}

// folded is a word's choice of a count of letters, or its last where it has
// fewer, with their case folded, as two words' letters are compared.
function folded(
	choices: readonly string[][],
	count: number,
	locale: string,
): string {
	const at = Math.min(count, choices.length) - 1;
	return (choices[at] ?? []).join("").toLocaleLowerCase(locale);
}

// capitalised is letters written with the first of them a capital, where the
// language writes capitals.
function capitalised(letters: readonly string[], locale: string): string {
	const [first = "", ...rest] = letters;
	return first.toLocaleUpperCase(locale) + rest.join("");
}
