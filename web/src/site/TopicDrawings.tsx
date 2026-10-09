import { type ComponentChildren, Fragment, type VNode } from "preact";
import { drawPicture, PictureFrame } from "../design/picture/diagram";
import type { Picture } from "../design/picture/model";
import type { Drawn } from "../design/picture/shapes";
import { Chain, DigitTree, type DrawingProps } from "./Art";
import type { Drawing, Markup } from "./drawings";
import type { PageReader } from "./reader";

/**
 * TopicDrawing is a drawing a topic's card or page shows: a picture, drawn as
 * a card of the widget draws one, or a drawing of the site's own, its words
 * the page's under at. The words around it say what it shows, so a screen
 * reader passes it over. A page is drawn before anybody reads it, so a picture
 * that cannot be drawn in the page's language stops the build, under the name
 * of where it stands, rather than leaving its place empty.
 */
export function TopicDrawing({
	drawing,
	page,
	at,
	where,
}: {
	drawing: Drawing;
	page: PageReader;
	at: string;
	where: string;
}) {
	if ("picture" in drawing) {
		return (
			<div class="s-picture" aria-hidden="true">
				<PictureFrame drawn={laidOut(drawing.picture, page.locale, where)} />
			</div>
		);
	}
	const OwnDrawing = drawings[drawing.markup];
	return (
		<div class="s-art" aria-hidden="true" dir="ltr">
			<OwnDrawing
				page={page}
				at={at}
				numbers={new Intl.NumberFormat(page.locale)}
			/>
		</div>
	);
}

// laidOut is a picture laid out in a language, or an error that names where
// the picture stands.
function laidOut(picture: Picture, locale: string, where: string): Drawn {
	try {
		return drawPicture(picture, locale);
	} catch (error) {
		throw new Error(`${where}: the picture cannot be drawn in ${locale}`, {
			cause: error,
		});
	}
}

// drawings are the drawings of the site's own, by their names: every name has
// one, or the site does not build.
const drawings: { readonly [M in Markup]: (props: DrawingProps) => VNode } = {
	islanders: Islanders,
	"product-regrouped": ProductRegrouped,
	"number-tree": NumberTree,
	"through-the-hour": ThroughTheHour,
	"sum-regrouped": SumRegrouped,
	"round-table": RoundTable,
	"two-stripe-flags": TwoStripeFlags,
	"strip-cuts": StripCuts,
	"cube-corners": CubeCorners,
	"eggs-backwards": EggsBackwards,
	"price-changes": PriceChanges,
	daisy: Daisy,
};

// islanders are the two islanders who speak, by the keys of their words.
const islanders = ["first", "second"] as const;

// Islanders draws two islanders, each with what they say.
function Islanders({ page, at }: DrawingProps) {
	return (
		<div class="s-speech">
			{islanders.map((one) => (
				<p key={one} class="s-speech-line">
					<span class="s-art-dot">{page.text(`${at}.${one}`)}</span>
					<span class="s-speech-says">{page.text(`${at}.${one}-says`)}</span>
				</p>
			))}
		</div>
	);
}

// ProductRegrouped draws a product whose two factors that make a hundred are
// marked, and under it the product they make it.
function ProductRegrouped({ numbers }: DrawingProps) {
	return (
		<p class="s-sum">
			<span class="s-art-chip s-art-chip-accent">{numbers.format(25)}</span>
			<span>×</span>
			<span>{numbers.format(7)}</span>
			<span>×</span>
			<span class="s-art-chip s-art-chip-accent">{numbers.format(4)}</span>
			<span class="s-sum-break" />
			<span>=</span>
			<span>{numbers.format(7)}</span>
			<span>×</span>
			<span class="s-art-chip s-art-chip-accent">{numbers.format(100)}</span>
		</p>
	);
}

// NumberTree draws the two-digit numbers with different digits made of 1, 2
// and 3, grown from their first digit.
function NumberTree({ numbers }: DrawingProps) {
	return <DigitTree digits={[1, 2, 3]} numbers={numbers} />;
}

// ThroughTheHour draws a time stepped on to the whole hour, then past it, each
// step's minutes on its arrow.
function ThroughTheHour({ page, at }: DrawingProps) {
	return (
		<Chain
			start="4:50"
			steps={[
				{ by: page.text(`${at}.step`, { count: 10 }), to: "5:00" },
				{ by: page.text(`${at}.step`, { count: 15 }), to: "5:15" },
			]}
		/>
	);
}

// pairs are the terms of the sum regrouped, two by two.
const pairs = [
	[19, 81],
	[23, 77],
] as const;

// SumRegrouped draws a sum regrouped into pairs, each pair over the hundred it
// makes.
function SumRegrouped({ numbers }: DrawingProps) {
	return (
		<div class="s-sum">
			{pairs.map(([first, second], at) => (
				<Fragment key={first}>
					{at > 0 && <span>+</span>}
					<span class="s-sum-pair">
						<span>
							{numbers.format(first)} + {numbers.format(second)}
						</span>
						<span class="s-art-chip s-art-chip-accent">
							{numbers.format(first + second)}
						</span>
					</span>
				</Fragment>
			))}
		</div>
	);
}

// seats are the places round the table, a knight in the first and the kinds
// taking turns from there.
const seats = 10;

// RoundTable draws the table with its answer: knights and liars taking turns
// round it, in the letters the solution names them by.
function RoundTable({ page, at }: DrawingProps) {
	return (
		<Ring count={seats} kind="s-round-table">
			{(seat) => {
				const knight = seat % 2 === 0;
				return (
					<span class={knight ? "s-art-dot s-art-dot-accent" : "s-art-dot"}>
						{page.text(knight ? `${at}.knight` : `${at}.liar`)}
					</span>
				);
			}}
		</Ring>
	);
}

// stripes are the colours a flag's stripe may be.
const stripes = ["red", "yellow", "green"] as const;

// TwoStripeFlags draws every flag of two stripes of different colours, in the
// order the solution lists them: by the top stripe, then by the bottom one.
function TwoStripeFlags() {
	return (
		<div class="s-flags">
			{stripes.flatMap((top) =>
				stripes
					.filter((bottom) => bottom !== top)
					.map((bottom) => (
						<span key={`${top}-${bottom}`} class="s-flag-pair">
							<span class={`s-flag-stripe s-flag-${top}`} />
							<span class={`s-flag-stripe s-flag-${bottom}`} />
						</span>
					)),
			)}
		</div>
	);
}

// cuts are the ways to cut a strip of 2 by 4 cells into two pieces of four,
// row by row: # a cell of one piece, . a cell of the other.
const cuts = [
	["####", "...."],
	["###.", "#..."],
	["##..", "##.."],
	["#...", "###."],
] as const;

// StripCuts draws each way to cut the strip, its two pieces in two shades.
function StripCuts() {
	return (
		<div class="s-cuts">
			{cuts.map((rows) => (
				<span key={rows.join("")} class="s-cut">
					{rows.flatMap((row, line) =>
						[...row].map((cell, place) => (
							<span
								key={`${line}-${place}`}
								class={cell === "#" ? "s-cut-cell s-cut-first" : "s-cut-cell"}
							/>
						)),
					)}
				</span>
			))}
		</div>
	);
}

// Corner is a corner of a cube, by its three coordinates, each 0 or 1.
type Corner = readonly [number, number, number];

// corners are the eight corners of a cube.
const corners: readonly Corner[] = [0, 1].flatMap((x) =>
	[0, 1].flatMap((y) => [0, 1].map((z): Corner => [x, y, z])),
);

// edges are the twelve edges of a cube: the pairs of corners that differ in
// one coordinate.
const edges = corners.flatMap((one, at) =>
	corners
		.slice(at + 1)
		.filter(
			(other) =>
				one.filter((coordinate, axis) => coordinate !== other[axis]).length ===
				1,
		)
		.map((other) => [one, other] as const),
);

// The cube as it is drawn: a side of the front face, how far up and right
// the back face stands, the margin round it and the radius of a corner.
const side = 96;
const depth = 44;
const margin = 12;
const cubeSize = margin + depth + side + margin;
const cornerRadius = 7;

// pointOf is where a corner is drawn: the front face at the bottom left, the
// back one up and to the right of it.
function pointOf([x, y, z]: Corner): { x: number; y: number } {
	return {
		x: margin + side * x + depth * z,
		y: margin + depth * (1 - z) + side * (1 - y),
	};
}

// behind says whether a corner is hidden behind the front face: the back
// corner at the bottom left, whose edges are drawn dashed.
function behind([x, y, z]: Corner): boolean {
	return x === 0 && y === 0 && z === 1;
}

// dark says whether a corner is drawn black: those an odd number of edges from
// the front corner at the bottom left, which is white.
function dark([x, y, z]: Corner): boolean {
	return (x + y + z) % 2 === 1;
}

// CubeCorners draws a cube of wire whose corners take turns in colour: every
// edge joins a black corner and a white one.
function CubeCorners() {
	return (
		<svg
			class="s-cube"
			aria-hidden="true"
			width={cubeSize}
			height={cubeSize}
			viewBox={`0 0 ${cubeSize} ${cubeSize}`}
		>
			{edges.map(([one, other]) => {
				const from = pointOf(one);
				const to = pointOf(other);
				return (
					<line
						key={`${one.join("")}-${other.join("")}`}
						class={
							behind(one) || behind(other)
								? "s-cube-edge s-cube-behind"
								: "s-cube-edge"
						}
						x1={from.x}
						y1={from.y}
						x2={to.x}
						y2={to.y}
					/>
				);
			})}
			{corners.map((corner) => {
				const { x, y } = pointOf(corner);
				return (
					<circle
						key={corner.join("")}
						class={dark(corner) ? "s-cube-corner s-cube-dark" : "s-cube-corner"}
						cx={x}
						cy={y}
						r={cornerRadius}
					/>
				);
			})}
		</svg>
	);
}

// EggsBackwards draws the eggs Granny has before each buyer, and after the
// last.
function EggsBackwards({ numbers }: DrawingProps) {
	return (
		<Chain
			start={numbers.format(7)}
			steps={[3, 1, 0].map((left) => ({ to: numbers.format(left) }))}
		/>
	);
}

// PriceChanges draws a price raised by a quarter and then cut by a fifth of
// the new price, each change in percent on its arrow.
function PriceChanges({ numbers }: DrawingProps) {
	return (
		<Chain
			start={numbers.format(400)}
			steps={[
				{ by: `+${numbers.format(25)}%`, to: numbers.format(500) },
				{ by: `−${numbers.format(20)}%`, to: numbers.format(400) },
			]}
		/>
	);
}

// petals are the daisy's petals, and torn those the first two moves tore off,
// one by each player, opposite each other.
const petals = 12;
const torn: ReadonlySet<number> = new Set([0, petals / 2]);

// Daisy draws the daisy after the first two moves: two rows of five petals
// left between the two torn off.
function Daisy() {
	return (
		<Ring count={petals} kind="s-daisy">
			{(petal) => (
				<span class={torn.has(petal) ? "s-petal s-petal-torn" : "s-petal"} />
			)}
		</Ring>
	);
}

// Ring stands count places evenly round a circle, the first at the top and
// the rest clockwise, each turned by its share of a full turn, which its
// styles read as --turn; kind names the ring's look.
function Ring({
	count,
	kind,
	children,
}: {
	count: number;
	kind: string;
	children: (place: number) => ComponentChildren;
}) {
	return (
		<div class={`s-ring ${kind}`}>
			{Array.from({ length: count }, (_, place) => (
				<span
					key={place}
					class="s-ring-place"
					style={{ "--turn": `${(360 * place) / count}deg` }}
				>
					{children(place)}
				</span>
			))}
		</div>
	);
}
