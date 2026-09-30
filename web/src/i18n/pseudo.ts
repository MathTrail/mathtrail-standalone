import { type Dictionary, placeholder, type Wording } from "./words";

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
 * text is twice as long as its English one, and at least ten characters
 * longer, since a short English label is the one a translation lengthens most.
 * Each text is set in brackets, so that one cut short shows where. The slots
 * stay as they are, and so does each plural wording's set of forms.
 */
export function pseudoWords(english: Dictionary): Dictionary {
	return Object.fromEntries(
		Object.entries(english).map(([key, wording]) => [key, stretched(wording)]),
	);
}

// stretched is a wording as the pseudo-language says it: each of its texts.
function stretched(wording: Wording): Wording {
	if (typeof wording === "string") {
		return pseudoText(wording);
	}
	return Object.fromEntries(
		Object.entries(wording).map(([form, text]) => [form, pseudoText(text)]),
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

const vowels = "aeiouyAEIOUY";

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
 * padded to twice its length at least, and ten characters more, its slots as
 * they were, in brackets.
 */
export function pseudoText(text: string): string {
	const length = literalLength(text);
	const wanted = Math.max(length * 2, length + 10);
	let said = aroundSlots(text, saidLonger);
	// The two brackets count towards the length too.
	for (let at = 0; literalLength(said) + 2 < wanted; at++) {
		said += ` ${saidLonger(padding[at % padding.length] ?? "")}`;
	}
	return `[${said}]`;
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
		const shown = accented[letter] ?? letter;
		said += vowels.includes(letter)
			? shown + (accented[letter.toLowerCase()] ?? letter)
			: shown;
	}
	return said;
}

// literalLength is how many characters of text are its own rather than its
// slots', which the words filled in decide.
function literalLength(text: string): number {
	return [...text.replace(placeholder, "")].length;
}
