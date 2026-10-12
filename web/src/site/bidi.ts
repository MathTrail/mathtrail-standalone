/**
 * Text in Latin letters and digits inside a page written right to left: a
 * sum, a time, a citation or a piece of code keeps the order it is written
 * in, left to right, as a page's drawings and formulas do. A browser keeps a
 * word or a number whole in either direction, but lays the signs and spaces
 * between them out right to left: 12 ÷ 3 = 4 would read 4 = 3 ÷ 12, and /mcp
 * mcp/.
 */

// letters are what a run is made of: Latin and Greek letters, the digits 0
// to 9, and the digits written small or as a fraction, ² and ½.
const letters = String.raw`\p{sc=Latn}\p{sc=Grek}\d\p{No}`;

// run finds a stretch of text in letters, with the spaces, signs and marks
// between them, from its first letter; or from a bracket, slash or dot right
// before one, or a sign that stands before one, as in −3 or + Add. An en
// dash stays out of it: a range, 3–4, reads in the sentence's direction, as
// the card's do.
const run = new RegExp(
	String.raw`(?:[(/.](?=[${letters}])|[+\-−](?=\s?[${letters}])|[${letters}])[${letters}\s+\-−×÷=<>*/:·.,()%'’_@#&…]*`,
	"gu",
);

// ends are what a run may end in: a letter, a closing bracket, a percent
// sign, a star or a slash. A space or a full stop after it is the sentence's.
const ends = new RegExp(`[${letters})%*/]`, "u");

// lettered says whether a run holds a letter, and so is more than a mark of
// the sentence around it.
const lettered = new RegExp(`[${letters}]`, "u");

// joined says whether a run holds something between its letters, which the
// browser would lay out in the sentence's direction.
const joined = /[^\p{L}\p{N}]/u;

/**
 * Span is a stretch of a text, from its first character to the one after its
 * last.
 */
export type Span = readonly [start: number, end: number];

/**
 * leftToRightSpans are the stretches of a text written right to left that
 * must read left to right: each run that joins its letters and digits by
 * spaces, signs or marks, as a sum, a time or an address does. A word or a
 * number alone is left out, and so is a bracket the run does not both open
 * and close, which belongs to the sentence around it. wholes are stretches
 * the text holds as one, an emphasis or a slot's filling, which a run takes
 * whole or not at all.
 */
export function leftToRightSpans(
	text: string,
	wholes: readonly Span[] = [],
): Span[] {
	const spans: Span[] = [];
	for (const found of text.matchAll(run)) {
		const [start, end] = settled(text, wholes, [
			found.index,
			found.index + found[0].length,
		]);
		const stretch = text.slice(start, end);
		if (lettered.test(stretch) && joined.test(stretch)) {
			spans.push([start, end]);
		}
	}
	return spans;
}

// settled is a run cut back until its ends hold: it ends in what a run may
// end in, holds only brackets it both opens and closes, and cuts no stretch
// the text holds as one.
function settled(text: string, wholes: readonly Span[], found: Span): Span {
	let [start, end] = found;
	let was: Span;
	do {
		was = [start, end];
		while (start < end && /\s/u.test(text[start] ?? "")) {
			start += 1;
		}
		while (start < end && !ends.test(text[end - 1] ?? "")) {
			end -= 1;
		}
		[start, end] = bracketed(text, [start, end]);
		for (const [from, to] of wholes) {
			if (from < start && start < to) {
				start = to;
			}
			if (from < end && end < to) {
				end = from;
			}
		}
	} while (start !== was[0] || end !== was[1]);
	return start < end ? [start, end] : [start, start];
}

// bracketed is a stretch cut back to the brackets it both opens and closes:
// it begins after a closing bracket it never opened, which belongs to the
// sentence before it, and ends before an opening one it never closes.
function bracketed(text: string, [start, end]: Span): Span {
	const opened: number[] = [];
	for (let at = start; at < end; at += 1) {
		if (text[at] === "(") {
			opened.push(at);
		} else if (text[at] === ")") {
			if (opened.pop() === undefined) {
				return [at + 1, end];
			}
		}
	}
	return [start, opened[0] ?? end];
}
