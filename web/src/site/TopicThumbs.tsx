import type { VNode } from "preact";
import type { Drawing, Thumb } from "./drawings";
import { thumbs } from "./drawings";
import { useSiteWords } from "./words";

/**
 * ThumbProps are what a card's small drawing is drawn from: the way the page's
 * language writes numbers.
 */
type ThumbProps = { readonly numbers: Intl.NumberFormat };

/**
 * thumbDrawings are the cards' small drawings, by their names: every name has
 * one, or the site does not build. Each is a sign of its topic drawn in
 * coloured shapes, and holds no words but the letters of a knight and a liar,
 * which the site's words give, and numbers written the way the page's
 * language writes them.
 */
export const thumbDrawings: {
	readonly [T in Thumb]: (props: ThumbProps) => VNode;
} = {
	"lined-up": LinedUp,
	"knight-and-liar": KnightAndLiar,
	"two-circles": TwoCircles,
	"digit-pairs": DigitPairs,
	fence: Fence,
	hutches: Hutches,
	"grid-cells": GridCells,
	"clock-face": ClockFace,
	"two-weeks": TwoWeeks,
	"hundred-product": HundredProduct,
	"alternating-dots": AlternatingDots,
	"groups-and-remainder": GroupsAndRemainder,
	"three-quarters": ThreeQuarters,
	"one-in-ten": OneInTen,
	"two-to-three": TwoToThree,
	matches: Matches,
	balance: Balance,
};

/**
 * TopicThumb is a topic's small drawing where a link to the topic shows it,
 * if its card's drawing is one: a sign the screen reader passes over, since
 * the link's words name the topic.
 */
export function TopicThumb({
	drawing,
	numbers,
}: {
	drawing: Drawing | undefined;
	numbers: Intl.NumberFormat;
}) {
	if (
		drawing === undefined ||
		!("markup" in drawing) ||
		!(thumbs as readonly string[]).includes(drawing.markup)
	) {
		return null;
	}
	const Thumb = thumbDrawings[drawing.markup as Thumb];
	return (
		<span class="s-thumb-frame" aria-hidden="true" dir="ltr">
			<Thumb numbers={numbers} />
		</span>
	);
}

// Paint is a colour of a thumb's shapes, which its stylesheet names.
type Paint =
	| "red"
	| "blue"
	| "green"
	| "orange"
	| "sky"
	| "peach"
	| "teal"
	| "leaf"
	| "grey";

// Shapes are so many shapes of a kind side by side, each in its paint.
function Shapes({
	kind,
	paints,
}: {
	kind: string;
	paints: readonly Paint[];
}) {
	return (
		<>
			{paints.map((paint, at) => (
				<span key={at} class={`s-thumb-${kind}`} data-paint={paint} />
			))}
		</>
	);
}

// these are so many of one paint, then so many of another.
function these(
	count: number,
	paint: Paint,
	rest = 0,
	other: Paint = "grey",
): Paint[] {
	return [
		...Array.from({ length: count }, () => paint),
		...Array.from({ length: rest }, () => other),
	];
}

// LinedUp draws three people lined up by height, the tallest first.
function LinedUp() {
	return (
		<span class="s-thumb s-thumb-people">
			<span class="s-thumb-person" data-paint="red" data-height="tall" />
			<span class="s-thumb-person" data-paint="blue" data-height="middle" />
			<span class="s-thumb-person" data-paint="green" data-height="short" />
		</span>
	);
}

// KnightAndLiar draws a knight in green and a liar in red, each by the first
// letter of what they are.
function KnightAndLiar() {
	const words = useSiteWords();
	return (
		<span class="s-thumb">
			<span class="s-thumb-islander" data-role="knight">
				{words.text("thumb.knight")}
			</span>
			<span class="s-thumb-islander" data-role="liar">
				{words.text("thumb.liar")}
			</span>
		</span>
	);
}

// TwoCircles draws two circles that overlap, a blue one and an orange one.
function TwoCircles() {
	return (
		<span class="s-thumb s-thumb-circles">
			<span data-paint="blue" />
			<span data-paint="orange" />
		</span>
	);
}

// digitPaints are the colours of the digits of the pairs, each digit keeping
// its own.
const digitPaints: Readonly<Record<number, Paint>> = {
	1: "blue",
	2: "orange",
	3: "green",
};

// DigitPairs draws the first two-digit numbers of the digits 1, 2 and 3 with
// no digit twice, in the order they are listed.
function DigitPairs({ numbers }: ThumbProps) {
	const pairs = [
		[1, 2],
		[1, 3],
		[2, 1],
		[2, 3],
	] as const;
	return (
		<span class="s-thumb">
			{pairs.map((pair) => (
				<span key={pair.join("")} class="s-thumb-pair">
					{pair.map((digit) => (
						<span
							key={digit}
							class="s-thumb-tile"
							data-paint={digitPaints[digit]}
						>
							{numbers.format(digit)}
						</span>
					))}
				</span>
			))}
		</span>
	);
}

// Fence draws four posts and the three rails between them.
function Fence() {
	return (
		<span class="s-thumb s-thumb-fence">
			{Array.from({ length: 4 }, (_, at) => [
				at > 0 && <span key={`rail-${at}`} class="s-thumb-rail" />,
				<span key={`post-${at}`} class="s-thumb-post" />,
			])}
		</span>
	);
}

// Hutches draws three hutches with four rabbits, two in the first.
function Hutches() {
	return (
		<span class="s-thumb">
			{[2, 1, 1].map((rabbits, at) => (
				<span key={at} class="s-thumb-hutch">
					<Shapes kind="dot" paints={these(rabbits, "orange")} />
				</span>
			))}
		</span>
	);
}

// GridCells draws two rows of five cells with a shape of five cells filled.
function GridCells() {
	return (
		<span class="s-thumb s-thumb-grid" data-columns="5">
			<Shapes
				kind="cell"
				paints={[...these(3, "sky", 2), ...these(2, "sky", 3)]}
			/>
		</span>
	);
}

// ClockFace draws a clock's face with its two hands.
function ClockFace() {
	return (
		<span class="s-thumb">
			<span class="s-thumb-clock">
				<span class="s-thumb-hour" />
				<span class="s-thumb-minute" />
			</span>
		</span>
	);
}

// TwoWeeks draws two weeks of seven days, the same weekday marked in each.
function TwoWeeks() {
	const week = [...these(2, "grey"), "orange", ...these(4, "grey")] as Paint[];
	return (
		<span class="s-thumb s-thumb-grid" data-columns="7">
			<Shapes kind="day" paints={[...week, ...week]} />
		</span>
	);
}

// HundredProduct draws the product of 25 and 4, which makes a hundred.
function HundredProduct({ numbers }: ThumbProps) {
	return (
		<span class="s-thumb">
			<span class="s-thumb-sum">
				{numbers.format(25)} × {numbers.format(4)} = {numbers.format(100)}
			</span>
		</span>
	);
}

// AlternatingDots draws six dots, filled and empty by turns.
function AlternatingDots() {
	return (
		<span class="s-thumb s-thumb-dots">
			{Array.from({ length: 6 }, (_, at) => (
				<span key={at} class="s-thumb-turn" data-empty={at % 2 === 1 ? "" : undefined} />
			))}
		</span>
	);
}

// GroupsAndRemainder draws three groups of four dots and one dot left over.
function GroupsAndRemainder() {
	return (
		<span class="s-thumb s-thumb-groups">
			{Array.from({ length: 3 }, (_, at) => (
				<span key={at} class="s-thumb-grid" data-columns="2">
					<Shapes kind="dot" paints={these(4, "teal")} />
				</span>
			))}
			<Shapes kind="dot" paints={["orange"]} />
		</span>
	);
}

// ThreeQuarters draws a whole of four parts with three of them filled.
function ThreeQuarters() {
	return (
		<span class="s-thumb s-thumb-parts">
			<Shapes kind="part" paints={these(3, "peach", 1)} />
		</span>
	);
}

// OneInTen draws a whole of ten parts with one of them filled.
function OneInTen() {
	return (
		<span class="s-thumb s-thumb-parts">
			<Shapes kind="tenth" paints={these(1, "leaf", 9)} />
		</span>
	);
}

// TwoToThree draws two parts to three, the two in blue and the three in
// orange.
function TwoToThree() {
	return (
		<span class="s-thumb s-thumb-ratio">
			<Shapes kind="tile" paints={these(2, "sky")} />
			<span class="s-thumb-colon">:</span>
			<Shapes kind="tile" paints={these(3, "peach")} />
		</span>
	);
}

// Matches draws six matches, their heads up.
function Matches() {
	return (
		<span class="s-thumb s-thumb-matches">
			{Array.from({ length: 6 }, (_, at) => (
				<span key={at} class="s-thumb-match" />
			))}
		</span>
	);
}

// Balance draws a balance with three coins on each pan.
function Balance() {
	return (
		<span class="s-thumb s-thumb-balance">
			<span class="s-thumb-pan">
				<Shapes kind="coin" paints={these(3, "orange")} />
			</span>
			<span class="s-thumb-pivot" />
			<span class="s-thumb-pan">
				<Shapes kind="coin" paints={these(3, "orange")} />
			</span>
		</span>
	);
}
