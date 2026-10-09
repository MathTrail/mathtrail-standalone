import { drawBalance } from "./balance";
import { drawBars } from "./bars";
import { drawCalendar } from "./calendar";
import { drawClock } from "./clock";
import { drawContainers } from "./containers";
import { drawGrid } from "./grid";
import type { Kind, Picture } from "./model";
import { drawNumberLine } from "./numberline";
import { drawPiles } from "./piles";
import { drawRing } from "./ring";
import { drawRow } from "./row";
import { type Drawn, r1 } from "./shapes";
import { drawTable } from "./table";
import { clearance } from "./text";
import { drawVenn } from "./venn";

// Drawing is how a picture of one kind is drawn, in the language of the task
// it belongs to: the words a kind draws itself, a month's weekdays, are words
// of the task, as its question is.
type Drawing<K extends Kind> = (
	picture: Extract<Picture, { kind: K }>,
	locale: string,
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
};

// drawnOf is a picture laid out, or undefined where it cannot be: a picture
// that fails to draw is left out, and the task stands on its words, rather
// than take the card down with it.
function drawnOf(picture: Picture, locale: string): Drawn | undefined {
	try {
		return drawings[picture.kind](picture as never, locale);
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
 */
export function Diagram({
	picture,
	label,
	locale,
}: {
	picture: Picture;
	label: string;
	locale: string;
}) {
	const drawn = drawnOf(picture, locale);
	if (drawn === undefined) {
		return null;
	}
	const width = r1(drawn.width + 2 * clearance);
	const height = r1(drawn.height + 2 * clearance);
	return (
		<svg
			class="mt-picture"
			role="img"
			aria-label={label}
			direction="ltr"
			width={width}
			height={height}
			viewBox={`${-clearance} ${-clearance} ${width} ${height}`}
		>
			{drawn.body}
		</svg>
	);
}
