import type { Said } from "../controls";
import { drawBalance } from "./balance";
import { drawBars } from "./bars";
import { drawCalendar } from "./calendar";
import { drawClock } from "./clock";
import { drawContainers } from "./containers";
import { drawFlags } from "./flags";
import { drawGrid } from "./grid";
import type { Kind, Picture } from "./model";
import { drawNumberLine } from "./numberline";
import { type Painted, paintedOf } from "./paint";
import { drawPiles } from "./piles";
import { drawRing } from "./ring";
import { drawRow } from "./row";
import { type Drawn, r1 } from "./shapes";
import { drawTable } from "./table";
import { clearance } from "./text";
import type { Tones } from "./tones";
import { drawVenn } from "./venn";

// Drawing is how a picture of one kind is drawn, in the language of the task
// it belongs to: the words a kind draws itself, a month's weekdays, are words
// of the task, as its question is. A kind that names its parts lights them in
// the tones a page gives.
type Drawing<K extends Kind> = (
	picture: Extract<Picture, { kind: K }>,
	locale: string,
	tones?: Tones,
) => Drawn;

// drawings are how each kind of picture is drawn: every kind has one, or the
// card does not build. A kind is added with a drawing of its own and a line
// here.
const drawings: { [K in Kind]: Drawing<K> } = {
	clock: drawClock,
	table: drawTable,
	number_line: drawNumberLine,
	row: drawRow,
	ring: drawRing,
	grid: drawGrid,
	bars: drawBars,
	venn: drawVenn,
	balance: drawBalance,
	containers: drawContainers,
	piles: drawPiles,
	calendar: drawCalendar,
	flags: drawFlags,
};

/**
 * drawPicture is a picture laid out in a language: its shapes and words, and
 * the room they take, its parts lit in the tones a page gives, if any. It
 * throws where the picture cannot be drawn, and what to draw in its place, if
 * anything, is for the caller to decide.
 */
export function drawPicture(
	picture: Picture,
	locale: string,
	tones?: Tones,
): Drawn {
	return drawings[picture.kind](picture as never, locale, tones);
}

// drawnOf is a picture laid out, or undefined where it cannot be: a picture
// that fails to draw is left out, and the task stands on its words, rather
// than take the card down with it.
function drawnOf(picture: Picture, locale: string): Drawn | undefined {
	try {
		return drawPicture(picture, locale);
	} catch (error) {
		console.error("widget: the picture could not be drawn", error);
		return undefined;
	}
}

/**
 * Diagram is a task's picture, drawn from its description in the colours of
 * the card's theme: a picture a screen reader is told what it shows by its
 * label, laid out left to right in every language, and no wider than its own
 * size or the card, whichever is narrower — it shrinks to fit and never
 * scrolls. The words of a task carry every fact its picture shows.
 *
 * A picture that paints has its key under it: each colour's paint with its
 * letters, beside the word the lesson calls it by, written as the task's
 * words are said and listed under keyLabel for a screen reader.
 */
export function Diagram({
	picture,
	label,
	locale,
	keyLabel,
	said,
}: {
	picture: Picture;
	label: string;
	locale: string;
	keyLabel?: string | undefined;
	said?: Said | undefined;
}) {
	const drawn = drawnOf(picture, locale);
	if (drawn === undefined) {
		return null;
	}
	const painted = "colors" in picture ? paintedOf(picture.colors, locale) : [];
	if (painted.length === 0) {
		return <PictureFrame drawn={drawn} label={label} />;
	}
	return (
		<div class="mt-picture-keyed">
			<PictureFrame drawn={drawn} label={label} />
			<PictureKey painted={painted} label={keyLabel} said={said} />
		</div>
	);
}

// PictureKey is the key under a picture that paints: each colour's paint,
// with the letters it carries in the picture, beside its word.
function PictureKey({
	painted,
	label,
	said,
}: {
	painted: readonly Painted[];
	label: string | undefined;
	said: Said | undefined;
}) {
	return (
		<ul class="mt-pic-key" aria-label={label} lang={said?.lang} dir={said?.dir}>
			{painted.map(({ paint, word, letters }) => (
				<li key={paint}>
					<span
						class={`mt-pic-swatch mt-paint mt-paint-${paint}`}
						aria-hidden="true"
					>
						<span class={`mt-paint-ink mt-paint-ink-${paint}`}>{letters}</span>
					</span>
					<span>{word}</span>
				</li>
			))}
		</ul>
	);
}

/**
 * PictureFrame is a picture already laid out, drawn as Diagram draws one: in
 * the colours of the theme, left to right, and no wider than its own size or
 * its room. A screen reader is told what it shows by its label, and passes
 * over one with no label, which stands beside words that say it already.
 */
export function PictureFrame({
	drawn,
	label,
}: {
	drawn: Drawn;
	label?: string;
}) {
	const width = r1(drawn.width + 2 * clearance);
	const height = r1(drawn.height + 2 * clearance);
	return (
		<svg
			class="mt-picture"
			role={label === undefined ? undefined : "img"}
			aria-label={label}
			aria-hidden={label === undefined ? "true" : undefined}
			direction="ltr"
			width={width}
			height={height}
			viewBox={`${-clearance} ${-clearance} ${width} ${height}`}
		>
			{drawn.body}
		</svg>
	);
}
