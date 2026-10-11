import { type Dictionary, placeholder, textsOf, type Wording } from "./words";

/**
 * pseudoLocale is the tag of the pseudo-language: English as a longer
 * language would say it. It is the tag Android gives its own pseudo-language
 * of stretched, accented English, and it names English, so it counts, writes
 * numbers and runs the way English does.
 */
export const pseudoLocale = "en-XA";

/**
 * pseudoWords are the English words stretched further than any translation
 * stretches them: every letter accented, so that nothing drawn from an English
 * text can pass for pseudo, every vowel said twice, so that every word grows
 * as a word of a longer language does, and words added at the end until the
 * text is twice as long as its English one, at least ten letters longer,
 * since a short English label is the one a translation lengthens most, and
 * no shorter than the longest text any of the translations says for its key,
 * in any of its forms. A dictionary added among the translations lengthens
 * the pseudo-language where it says a key longer. Each text is set in
 * brackets, so that one cut short shows where. The slots stay as they are,
 * and so does each plural wording's set of forms.
 */
export function pseudoWords(
	english: Dictionary,
	translations: Iterable<Dictionary>,
): Dictionary {
	return Object.fromEntries(
		[...longestOf(english, translations)].map(([key, { wording, atLeast }]) => [
			key,
			stretched(wording, atLeast),
		]),
	);
}

// longestOf is, for each key of English, its wording and how long each of its
// texts is to be at least: as long as the longest text any translation says
// for it, in any of its forms, where that is longer than the shortest text the
// key grows to from its English alone. A key English lacks is no key of the
// pseudo-language. Counting letters is what costs, and every card of the
// preview counts them as its page loads: a text has no more letters than code
// units, so one no longer in code units than the length its key has reached
// is not counted at all.
function longestOf(
	english: Dictionary,
	translations: Iterable<Dictionary>,
): Map<string, { wording: Wording; atLeast: number }> {
	const longest = new Map(
		Object.entries(english).map(([key, wording]) => [
			key,
			{
				wording,
				atLeast: Math.min(...textsOf(wording).map((text) => grown(text))),
			},
		]),
	);
	for (const dictionary of translations) {
		for (const [key, wording] of Object.entries(dictionary)) {
			const kept = longest.get(key);
			if (kept === undefined) {
				continue;
			}
			for (const text of textsOf(wording)) {
				if (text.length > kept.atLeast) {
					kept.atLeast = Math.max(kept.atLeast, lengthOf(text));
				}
			}
		}
	}
	return longest;
}

// stretched is a wording as the pseudo-language says it: each of its texts,
// none shorter than atLeast.
function stretched(wording: Wording, atLeast: number): Wording {
	if (typeof wording === "string") {
		return pseudoText(wording, atLeast);
	}
	return Object.fromEntries(
		Object.entries(wording).map(([form, text]) => [
			form,
			pseudoText(text, atLeast),
		]),
	);
}

// The letters of English with the accented ones that stand in for them.
const accented: Readonly<Record<string, string>> = {
	a: "á",
	b: "ƀ",
	c: "ç",
	d: "ð",
	e: "é",
	f: "ƒ",
	g: "ĝ",
	h: "ĥ",
	i: "í",
	j: "ĵ",
	k: "ķ",
	l: "ļ",
	m: "ḿ",
	n: "ñ",
	o: "ó",
	p: "þ",
	q: "ǫ",
	r: "ŕ",
	s: "š",
	t: "ţ",
	u: "ú",
	v: "ṽ",
	w: "ŵ",
	x: "ẋ",
	y: "ý",
	z: "ž",
	A: "Å",
	B: "Ɓ",
	C: "Ç",
	D: "Ð",
	E: "É",
	F: "Ƒ",
	G: "Ĝ",
	H: "Ĥ",
	I: "Î",
	J: "Ĵ",
	K: "Ķ",
	L: "Ļ",
	M: "Ḿ",
	N: "Ñ",
	O: "Ö",
	P: "Þ",
	Q: "Ǫ",
	R: "Ŕ",
	S: "Š",
	T: "Ţ",
	U: "Û",
	V: "Ṽ",
	W: "Ŵ",
	X: "Ẋ",
	Y: "Ý",
	Z: "Ž",
};

// The accented letters a vowel is said again as.
const echoes: Readonly<Record<string, string>> = {
	a: "á",
	e: "é",
	i: "í",
	o: "ó",
	u: "ú",
	y: "ý",
	A: "á",
	E: "é",
	I: "í",
	O: "ó",
	U: "ú",
	Y: "ý",
};

// The words a text too short for its growth is made longer with, as Android's
// pseudo-language does: counting words, said longer like the rest, and counted
// again from one when a long text needs more of them than ten.
const padding = [
	"one",
	"two",
	"three",
	"four",
	"five",
	"six",
	"seven",
	"eight",
	"nine",
	"ten",
];

/**
 * pseudoText is one text of the pseudo-language: its English stretched and
 * padded to twice its length at least, ten letters more and atLeast letters,
 * its slots as they were, in brackets.
 */
export function pseudoText(text: string, atLeast = 0): string {
	const wanted = Math.max(grown(text), atLeast);
	let said = aroundSlots(text, saidLonger);
	// The two brackets count towards the length too. A padding word, after the
	// space that sets it apart, adds its own letters to those before it, and is
	// spelt in letters of one code point each.
	let length = lengthOf(said) + 2;
	for (const word of counting()) {
		if (length >= wanted) {
			break;
		}
		const added = ` ${saidLonger(word)}`;
		said += added;
		length += [...added].length;
	}
	return `[${said}]`;
}

// grown is how long a text of English grows to in the pseudo-language by
// itself: twice its length, and ten letters more at least.
function grown(text: string): number {
	const length = lengthOf(text);
	return Math.max(length * 2, length + 10);
}

// counting are the padding words, counted again from one after ten.
function* counting(): Generator<string> {
	for (;;) {
		yield* padding;
	}
}

// aroundSlots is text with what lies between its slots said by say, and the
// slots as they were.
function aroundSlots(text: string, say: (words: string) => string): string {
	let said = "";
	let from = 0;
	for (const slot of text.matchAll(placeholder)) {
		said += say(text.slice(from, slot.index)) + slot[0];
		from = slot.index + slot[0].length;
	}
	return said + say(text.slice(from));
}

// saidLonger is text with every letter accented and every vowel said twice.
function saidLonger(text: string): string {
	let said = "";
	for (const letter of text) {
		said += (accented[letter] ?? letter) + (echoes[letter] ?? "");
	}
	return said;
}

// letters split a text where a reader sees one letter end and the next begin:
// a letter with its marks, a syllable of a script of India and an emoji are
// one each, however many code points they take. Every language splits them
// the same way. It is made at the first count: the build that ships imports
// this module and counts nothing, and a platform that cannot split letters
// still draws its cards.
let letters: Intl.Segmenter | undefined;

/**
 * lengthOf is how many letters of a text a reader sees that are its own
 * rather than its slots', which the words filled in decide.
 */
export function lengthOf(text: string): number {
	letters ??= new Intl.Segmenter(undefined, { granularity: "grapheme" });
	let length = 0;
	for (const _letter of letters.segment(text.replace(placeholder, ""))) {
		length++;
	}
	return length;
}
