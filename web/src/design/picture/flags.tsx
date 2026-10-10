import type { FlagGroup, Flags, Paint, Stripe } from "./model";
import { type Painted, paintedOf } from "./paint";
import { type Drawn, Label, r1, sizes } from "./shapes";
import { room, smallest, widthOf } from "./text";

// Flags, in the card's pixels: how thick a pole is, how wide a flag's cloth,
// how far the pole reaches under the tallest cloth, the room between two
// flags of a group, between two groups side by side and between two rows of
// groups, the room under the poles above the bracket, how deep the bracket
// is, and the room between it and the group's name under it.
const pole = 2;
const clothWide = 34;
const poleBelow = 9;
const flagGap = 9;
const groupGap = 18;
const rowGap = 12;
const bracketAbove = 5;
const bracketDeep = 5;
const nameAbove = 5;

// A stripe is as tall as the letters on it need: the fonts a card meets draw
// the letters of Latin, Greek, Cyrillic, Chinese, Japanese and Korean within
// a stripe of 15 px, and those of the other scripts — Arabic, the scripts of
// India, Thai — reach higher and lower, and take 19.
const stripeTall = 15;
const tallStripe = 19;

// compact are letters of the scripts a stripe of the usual height holds, with
// the marks they carry.
const compact =
	/^[\p{Script=Latin}\p{Script=Greek}\p{Script=Cyrillic}\p{Script=Han}\p{Script=Hiragana}\p{Script=Katakana}\p{Script=Hangul}\p{M}\p{N}\p{P}\s]*$/u;

// stripeOf is how tall every stripe of a picture's flags is: as tall as the
// letters of any of its paints need.
function stripeOf(painted: readonly Painted[]): number {
	return painted.every(({ letters }) => compact.test(letters))
		? stripeTall
		: tallStripe;
}

// A colour that names a group is a swatch of its paint, as tall as a stripe,
// with the paint's letters on it.
const swatchAside = 6;

// flagWide is how much room a flag takes across: its pole and its cloth.
const flagWide = pole + clothWide;

// Laid is a group of flags laid out: where its left edge stands, on which row
// of groups, and how wide it is.
type Laid = { group: FlagGroup; x: number; row: number; width: number };

/**
 * drawFlags draws flags in groups, side by side as many as fit the card and
 * the rest on rows under them: each flag a pole with its cloth at the top,
 * the stripes from top to bottom in their paints, each with its paint's
 * letters on it, a stripe whose colour is the unknown left blank with a ?;
 * and under each group's poles a bracket, and the label or the colour that
 * names it.
 */
export function drawFlags(flags: Flags, locale: string): Drawn {
	const painted = paintedOf(flags.colors, locale);
	const stripe = stripeOf(painted);
	const stripes = Math.max(
		...flags.groups.flatMap((group) => group.flags.map((flag) => flag.length)),
	);
	const flagsTall = stripes * stripe + poleBelow;
	const named = flags.groups.some(
		(group) => group.label !== undefined || group.color !== undefined,
	);
	const nameTall = Math.max(sizes.label * 1.4, stripe);
	const groupTall =
		flagsTall + (named ? bracketAbove + bracketDeep + nameAbove + nameTall : 0);
	const laid = laidOut(flags.groups, painted);
	const rows = Math.max(...laid.map((one) => one.row)) + 1;
	const width = Math.max(
		...Array.from({ length: rows }, (_, row) => rowWidth(laid, row)),
	);
	return {
		width,
		height: rows * groupTall + (rows - 1) * rowGap,
		body: (
			<>
				{laid.map((one) => {
					const left = one.x + (width - rowWidth(laid, one.row)) / 2;
					const top = one.row * (groupTall + rowGap);
					return (
						<Group
							key={`${left} ${top}`}
							laid={one}
							left={left}
							top={top}
							flagsTall={flagsTall}
							stripe={stripe}
							painted={painted}
						/>
					);
				})}
			</>
		),
	};
}

// laidOut lays the groups out in rows, each on the first row it fits beside
// those already there, in their order.
function laidOut(groups: readonly FlagGroup[], painted: Painted[]): Laid[] {
	const laid: Laid[] = [];
	let row = 0;
	let x = 0;
	for (const group of groups) {
		const width = groupWidth(group, painted);
		if (x > 0 && x + width > room) {
			row++;
			x = 0;
		}
		laid.push({ group, x, row, width });
		x += width + groupGap;
	}
	return laid;
}

// rowWidth is how wide a row of groups is.
function rowWidth(laid: readonly Laid[], row: number): number {
	const inRow = laid.filter((one) => one.row === row);
	return (
		inRow.reduce((sum, one) => sum + one.width, 0) +
		groupGap * Math.max(0, inRow.length - 1)
	);
}

// groupWidth is how wide a group is: its flags side by side, or the name under
// them where that is wider.
function groupWidth(group: FlagGroup, painted: Painted[]): number {
	const flagsWide =
		group.flags.length * flagWide + (group.flags.length - 1) * flagGap;
	return Math.max(flagsWide, nameWidth(group, painted));
}

// nameWidth is how wide what names a group is.
function nameWidth(group: FlagGroup, painted: Painted[]): number {
	if (group.color !== undefined) {
		return swatchWidth(lettersOf(painted, group.color));
	}
	return group.label === undefined ? 0 : widthOf(group.label, sizes.label);
}

// swatchWidth is how wide a swatch is that carries letters.
function swatchWidth(letters: string): number {
	return Math.max(
		clothWide * 0.75,
		widthOf(letters, smallest) + 2 * swatchAside,
	);
}

// lettersOf are the letters a paint carries.
function lettersOf(painted: readonly Painted[], paint: Paint): string {
	return painted.find((one) => one.paint === paint)?.letters ?? "";
}

// Group is one group of flags at its place, its flags centred over its
// bracket and its name, every stripe and swatch as tall as stripe.
function Group({
	laid,
	left,
	top,
	flagsTall,
	stripe,
	painted,
}: {
	laid: Laid;
	left: number;
	top: number;
	flagsTall: number;
	stripe: number;
	painted: Painted[];
}) {
	const { group, width } = laid;
	const flagsWide =
		group.flags.length * flagWide + (group.flags.length - 1) * flagGap;
	const first = left + (width - flagsWide) / 2;
	const bracketAt = top + flagsTall + bracketAbove;
	const nameAt = bracketAt + bracketDeep + nameAbove;
	const middle = left + width / 2;
	return (
		<>
			{group.flags.map((stripes, at) => {
				const x = first + at * (flagWide + flagGap);
				return (
					<Flag
						key={x}
						x={x}
						y={top}
						tall={flagsTall}
						stripe={stripe}
						stripes={stripes}
						painted={painted}
					/>
				);
			})}
			{(group.label !== undefined || group.color !== undefined) && (
				<>
					<path
						d={`M${r1(first)} ${r1(bracketAt)}V${r1(bracketAt + bracketDeep)}H${r1(first + flagsWide)}V${r1(bracketAt)}`}
						class="mt-pic-line"
						stroke-width={1.5}
						stroke-linejoin="round"
					/>
					{group.color === undefined ? (
						<Label
							x={middle}
							y={nameAt + (sizes.label * 1.4) / 2}
							text={group.label ?? ""}
						/>
					) : (
						<Swatch
							x={middle}
							y={nameAt}
							tall={stripe}
							paint={group.color}
							letters={lettersOf(painted, group.color)}
						/>
					)}
				</>
			)}
		</>
	);
}

// Flag is one flag, its pole's top at a point and as tall as given: the
// pole, and the cloth at its top, stripe by stripe, each as tall as stripe,
// outlined.
function Flag({
	x,
	y,
	tall,
	stripe,
	stripes,
	painted,
}: {
	x: number;
	y: number;
	tall: number;
	stripe: number;
	stripes: readonly Stripe[];
	painted: Painted[];
}) {
	const clothLeft = x + pole;
	return (
		<>
			<path
				d={`M${r1(x + pole / 2)} ${r1(y)}V${r1(y + tall)}`}
				class="mt-pic-line"
				stroke-width={pole}
				stroke-linecap="round"
			/>
			{stripes.map((paint, at) => {
				const top = y + at * stripe;
				return (
					<StripeOf
						key={top}
						x={clothLeft}
						y={top}
						tall={stripe}
						stripe={paint}
						painted={painted}
					/>
				);
			})}
			<rect
				x={r1(clothLeft)}
				y={r1(y)}
				width={clothWide}
				height={stripes.length * stripe}
				class="mt-pic-line"
				stroke-width={1}
			/>
		</>
	);
}

// StripeOf is one stripe of a flag, as tall as given: its paint with the
// paint's letters on it, or, where its colour is the unknown, the card's own
// colour with a ?.
function StripeOf({
	x,
	y,
	tall,
	stripe,
	painted,
}: {
	x: number;
	y: number;
	tall: number;
	stripe: Stripe;
	painted: Painted[];
}) {
	const middle = { x: x + clothWide / 2, y: y + tall / 2 };
	if (stripe === "?") {
		return (
			<>
				<rect
					x={r1(x)}
					y={r1(y)}
					width={clothWide}
					height={tall}
					class="mt-pic-plate"
				/>
				<Label x={middle.x} y={middle.y} text="?" size={smallest} />
			</>
		);
	}
	return (
		<>
			<rect
				x={r1(x)}
				y={r1(y)}
				width={clothWide}
				height={tall}
				class={`mt-paint mt-paint-${stripe}`}
			/>
			<Letters
				x={middle.x}
				y={middle.y}
				paint={stripe}
				letters={lettersOf(painted, stripe)}
			/>
		</>
	);
}

// Swatch is a colour that names a group: its paint, as tall as given, with its
// letters on it, centred under the group, and outlined, as a swatch of the key
// is, so that a paint of the card's own colour shows.
function Swatch({
	x,
	y,
	tall,
	paint,
	letters,
}: {
	x: number;
	y: number;
	tall: number;
	paint: Paint;
	letters: string;
}) {
	const wide = swatchWidth(letters);
	const box = { x: r1(x - wide / 2), y: r1(y), width: r1(wide), height: tall };
	return (
		<>
			<rect {...box} rx={3} class={`mt-paint mt-paint-${paint}`} />
			<rect {...box} rx={3} class="mt-pic-line" stroke-width={1} />
			<Letters x={x} y={y + tall / 2} paint={paint} letters={letters} />
		</>
	);
}

// Letters are a paint's letters, written on it in the colour that reads on
// that paint.
function Letters({
	x,
	y,
	paint,
	letters,
}: {
	x: number;
	y: number;
	paint: Paint;
	letters: string;
}) {
	return (
		<text
			x={r1(x)}
			y={r1(y)}
			dy="0.35em"
			text-anchor="middle"
			font-size={smallest}
			class={`mt-paint-ink mt-paint-ink-${paint}`}
		>
			{letters}
		</text>
	);
}
