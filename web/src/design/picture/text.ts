/**
 * The words of a picture, laid out without a browser: the card draws a
 * picture where it is shown and the site draws it at build time, so a
 * picture's text is placed by how wide it is known to be rather than by
 * measuring it. Every width is an estimate on the generous side, so that a
 * label placed by it has room in every font a card meets.
 */

/**
 * clearance is the margin a picture keeps clear on each side, for the strokes
 * along its edges, which reach half their width past the lines they draw.
 */
export const clearance = 2;

/**
 * room is the width a picture lays itself out in on the narrowest card: 320
 * px less the card's border and the padding of its body, and its own
 * clearance on each side. A picture no wider than this is drawn at its own
 * size on every card, so its labels keep their size.
 */
export const room = 286 - 2 * clearance;

/** smallest is the size below which no label of a picture is drawn. */
export const smallest = 11;

/**
 * squeezed are the sizes a crowded label is written at, largest first: down
 * to the smallest, and below it only for the widest labels of a table or a
 * grid of many columns, which fit at no larger size.
 */
export const squeezed: readonly number[] = [
	13,
	12,
	smallest,
	10.5,
	10,
	9.5,
	9,
	8.5,
	8,
	7.5,
	7,
	6.5,
	6,
];

// advances are the widths of the characters a picture writes in — its labels,
// times and notes, and the names of the days of a week in a language written
// in Latin letters — in ems: for each, the widest of the fonts the study
// measured a card in — DejaVu Sans, bold and regular, and Roboto — draws it.
const advances: Readonly<Record<string, number>> = {
	A: 0.774,
	B: 0.762,
	C: 0.734,
	D: 0.83,
	E: 0.683,
	F: 0.683,
	G: 0.821,
	H: 0.837,
	I: 0.372,
	J: 0.552,
	K: 0.775,
	L: 0.637,
	M: 0.995,
	N: 0.837,
	O: 0.85,
	P: 0.733,
	Q: 0.85,
	R: 0.77,
	S: 0.72,
	T: 0.682,
	U: 0.812,
	V: 0.774,
	W: 1.103,
	X: 0.771,
	Y: 0.724,
	Z: 0.725,
	a: 0.675,
	b: 0.716,
	c: 0.593,
	d: 0.716,
	e: 0.678,
	f: 0.435,
	g: 0.716,
	h: 0.712,
	i: 0.343,
	j: 0.343,
	k: 0.665,
	l: 0.343,
	m: 1.042,
	n: 0.712,
	o: 0.687,
	p: 0.716,
	q: 0.716,
	r: 0.493,
	s: 0.595,
	t: 0.478,
	u: 0.712,
	v: 0.652,
	w: 0.924,
	x: 0.645,
	y: 0.652,
	z: 0.582,
	"0": 0.696,
	"1": 0.696,
	"2": 0.696,
	"3": 0.696,
	"4": 0.696,
	"5": 0.696,
	"6": 0.696,
	"7": 0.696,
	"8": 0.696,
	"9": 0.696,
	"?": 0.58,
	"−": 0.838,
	"-": 0.415,
	".": 0.38,
	",": 0.38,
	":": 0.4,
	"…": 1,
	" ": 0.348,
	"+": 0.838,
	"×": 0.838,
	"÷": 0.838,
	"=": 0.838,
	"<": 0.838,
	">": 0.838,
	"(": 0.457,
	")": 0.457,
};

// mark is a character that takes no room of its own: an accent or another
// mark a letter carries, or a character that only says how letters join.
const mark = /[\p{M}\p{Cf}]/u;

// arabic is a letter of the Arabic script, which joins its neighbours and is
// narrower for it: the widest measured in a word took under half an em.
const arabic =
	/[\u0600-\u06FF\u0750-\u077F\u08A0-\u08FF\uFB50-\uFDFF\uFE70-\uFEFF]/u;

/**
 * widthOf is how wide a text is at a size, at most: a letter of the Arabic
 * script is taken as half an em, and any other character the table does not
 * know, a letter of another script, as wide as an em.
 */
export function widthOf(text: string, size: number): number {
	let ems = 0;
	for (const character of text) {
		if (!mark.test(character)) {
			ems += advances[character] ?? (arabic.test(character) ? 0.5 : 1);
		}
	}
	return ems * size;
}

/**
 * fitted is the first of some sizes, largest first, at which fits says a text
 * fits, and the last of them when none does.
 */
export function fitted(
	sizes: readonly number[],
	fits: (size: number) => boolean,
): number {
	for (const size of sizes) {
		if (fits(size)) {
			return size;
		}
	}
	return sizes.at(-1) ?? smallest;
}

/** Placed is a label laid out among others: where its middle stands, and on which line. */
export type Placed = { x: number; line: number };

/**
 * within is where the middle of a thing as wide as given stands, as near a
 * point as keeps the thing between from and to.
 */
export function within(
	middle: number,
	wide: number,
	from: number,
	to: number,
): number {
	return Math.min(Math.max(middle, from + wide / 2), to - wide / 2);
}

/**
 * stacked lays out labels centred on points of a line, each as wide as it is
 * said to be, on as few lines as keep any two of them apart by clear: each
 * goes on the first line where it touches no label already there. A label
 * that would stand past from or to is moved in, so that none leaves the
 * picture.
 */
export function stacked(
	labels: readonly { x: number; width: number }[],
	from: number,
	to: number,
	clear = 4,
): Placed[] {
	const lines: { start: number; end: number }[][] = [];
	return labels.map(({ x, width }) => {
		const middle = within(x, width, from, to);
		const start = middle - width / 2 - clear / 2;
		const end = middle + width / 2 + clear / 2;
		let line = lines.findIndex((taken) =>
			taken.every((one) => one.end <= start || one.start >= end),
		);
		if (line < 0) {
			line = lines.length;
			lines.push([]);
		}
		lines[line]?.push({ start, end });
		return { x: middle, line };
	});
}

/**
 * written is a label as a picture writes it: a number below zero with the
 * sign of mathematics, which a child reads as minus, wherever the description
 * wrote a hyphen.
 */
export function written(label: string): string {
	return /^-\d/.test(label) ? `−${label.slice(1)}` : label;
}

/**
 * writtenNote is a note as a picture writes it: every hyphen the sign of
 * mathematics, since a note's hyphen is a minus, of a subtraction or of a
 * number below zero, and a child reads either as minus.
 */
export function writtenNote(note: string): string {
	return note.replaceAll("-", "−");
}

/** widest is how wide the widest of some texts is at a size, and nothing for none. */
export function widest(texts: readonly string[], size: number): number {
	return Math.max(0, ...texts.map((text) => widthOf(text, size)));
}

/** lines are how many lines a stack of labels laid out by stacked takes. */
export function lines(placed: readonly Placed[]): number {
	return placed.reduce((most, one) => Math.max(most, one.line + 1), 0);
}
