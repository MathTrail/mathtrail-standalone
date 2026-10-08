import type { SVGAttributes } from "preact";
import { classes } from "./classes";
import { useScopedId } from "./ids";

// The icons are one set: each is drawn on a grid of 24, in lines of one width
// rounded at their ends and corners and never filled, and in one colour — the
// text's, so that a state that recolours its text recolours its icon too, or
// the token of what the icon marks. The colour is the drawing's own attribute
// rather than a style, which a host's page may refuse.
const lineWidth = 2.5;
const text = "currentColor";
const ring = "M2 12a10 10 0 1 0 20 0 10 10 0 1 0-20 0";
const tick = "M7.5 12.5l3 3 6-6.5";

const drawings = {
	"chevron-right": { d: "M9.5 7l5 5-5 5", tone: text },
	"chevron-left": { d: "M14.5 7l-5 5 5 5", tone: text },
	check: { d: "M5 12.5l4.5 4.5L19 7.5", tone: text },
	cross: { d: "M6.5 6.5l11 11M17.5 6.5l-11 11", tone: text },
	dash: { d: "M6 12h12", tone: text },
	hint: { d: "M9.5 17v-2.5a5.5 5.5 0 1 1 5 0V17ZM10.5 20h3", tone: text },
	trap: { d: "M12 3.5 21.5 20h-19ZM12 9.5V14M12 17v.01", tone: text },
	menu: { d: "M4 7h16M4 12h16M4 17h16", tone: text },
	chat: {
		d: "M21 12a8 8 0 0 1-11.6 7.1L4 20l1-4.6A8 8 0 1 1 21 12Z",
		tone: text,
	},
	renew: {
		d: "M4 12a8 8 0 0 1 13.66-5.66L20 8.5M20 4v4.5h-4.5M20 12a8 8 0 0 1-13.66 5.66L4 15.5M4 20v-4.5h4.5",
		tone: text,
	},
	// The spark is the logo's star, with a smaller one beside it.
	sparkle: {
		d: "M10 5Q12 11 18 13 12 15 10 21 8 15 2 13 8 11 10 5ZM19 2.5v5M16.5 5h5",
		tone: text,
	},
	tag: {
		d: "M3 5a2 2 0 0 1 2-2h6.17a2 2 0 0 1 1.41.59l8.3 8.3a2 2 0 0 1 0 2.83l-6.17 6.17a2 2 0 0 1-2.83 0l-8.3-8.3A2 2 0 0 1 3 11.17ZM7.5 7.5h.01",
		tone: text,
	},
	// A verdict is its tick or cross in a ring, in its own colour; a step done
	// is ticked in the colour of the strongest ink, and a step to come is the
	// ring alone.
	"verdict-correct": { d: `${ring}${tick}`, tone: "var(--correct)" },
	"verdict-wrong": {
		d: `${ring}M8.5 8.5l7 7M15.5 8.5l-7 7`,
		tone: "var(--wrong)",
	},
	"step-done": { d: `${ring}${tick}`, tone: "var(--ink-strong)" },
	"step-waiting": { d: ring, tone: "var(--track)" },
	// The child's avatar is the same for every child: a card shows a pseudonym
	// and nothing that could tell children apart.
	avatar: {
		d: `${ring}M15.25 9.5a3.25 3.25 0 1 1-6.5 0 3.25 3.25 0 0 1 6.5 0M6 19a7.5 7.5 0 0 1 12 0`,
		tone: "var(--text-muted)",
	},
} as const;

/**
 * IconName names an icon of the design: a line icon, the spinner of a check
 * under way, the mark of a verdict, how a step stands — done, or to come —
 * and the child's avatar.
 */
export type IconName = keyof typeof drawings | "spinner";

/**
 * Icon is one of the design's icons. It is decoration, hidden from a screen
 * reader: the words beside it say what it shows. The spinner turns its arc,
 * in the text's colour, on a track of the colour a token names, so that it
 * shows on the tint of the row it turns in.
 */
export function Icon({
	name,
	size = 16,
	track = "track",
	className,
}: {
	name: IconName;
	size?: number;
	track?: string;
	className?: string;
}) {
	return (
		<svg
			width={size}
			height={size}
			viewBox="0 0 24 24"
			fill="none"
			aria-hidden="true"
			class={classes("mt-icon", name === "spinner" && "mt-spin", className)}
		>
			{name === "spinner" ? (
				<>
					<Line d={ring} tone={`var(--${track})`} />
					<Line d="M12 2a10 10 0 0 1 10 10" tone={text} />
				</>
			) : (
				<Line {...drawings[name]} />
			)}
		</svg>
	);
}

// Line is one stroke of an icon, in the set's width.
function Line({ d, tone }: { d: string; tone: string }) {
	return (
		<path
			d={d}
			stroke={tone}
			stroke-width={lineWidth}
			stroke-linecap="round"
			stroke-linejoin="round"
		/>
	);
}

// Ink is the colour one part of the logo runs through, from its colour at the
// top to its colour at the bottom.
type Ink = { readonly from: string; readonly to: string };

// Look is one of the logo's two drawings: the theme it is shown in, its tile,
// and the inks of the ribbon's four parts and of the star.
type Look = {
	readonly theme: "light" | "dark";
	readonly tile: SVGAttributes<SVGRectElement>;
	readonly blueLeg: Ink;
	readonly blueBand: Ink;
	readonly greenLeg: Ink;
	readonly greenBand: Ink;
	readonly star: Ink;
};

// onDark is the very drawing the site serves as its icon: the ribbon and the
// star on a dark navy tile.
const onDark: Look = {
	theme: "dark",
	tile: { width: "48", height: "48", rx: "6", fill: "#0c1a33" },
	blueLeg: { from: "#3d8ef3", to: "#2766d6" },
	blueBand: { from: "#76bcff", to: "#4791f0" },
	greenLeg: { from: "#b9e276", to: "#76c05b" },
	greenBand: { from: "#86ce69", to: "#57ae55" },
	star: { from: "#ffcd6e", to: "#f5a238" },
};

// onLight is the same ribbon and star on a white tile under a hairline that
// stays within the square, the lightest of their inks deepened, so that the
// tops of the bands and of the green leg, and the star, hold on a light
// ground.
const onLight: Look = {
	theme: "light",
	tile: {
		x: "0.75",
		y: "0.75",
		width: "46.5",
		height: "46.5",
		rx: "5.5",
		fill: "#ffffff",
		stroke: "#d6dce6",
		"stroke-width": "1.5",
	},
	blueLeg: onDark.blueLeg,
	blueBand: { from: "#5aa8f7", to: "#3b7fe3" },
	greenLeg: { from: "#8fcf55", to: "#4fa648" },
	greenBand: { from: "#6fc35a", to: "#3f9a45" },
	star: { from: "#ffb938", to: "#ec8a12" },
};

/**
 * Mark is MathTrail's logo — one ribbon folded into an M, and a star over it —
 * drawn for either theme, of which the theme's tokens show one: in the dark
 * theme the very drawing the site serves as its icon, on a dark tile, and in
 * the light theme the same ribbon and star on a white tile, whose dark tile
 * would weigh on a light card. A card loads nothing from outside itself, so it
 * draws the logo in place. It is decoration: the name beside it says whose it
 * is.
 */
export function Mark({ size = 32 }: { size?: number }) {
	return (
		<>
			<MarkIn look={onLight} size={size} />
			<MarkIn look={onDark} size={size} />
		</>
	);
}

// MarkIn is the logo as one look draws it, named after that look's theme so
// that the theme's tokens can show it alone; it names its gradients and its
// clip apart from those of any other logo on the page.
function MarkIn({ look, size }: { look: Look; size: number }) {
	const id = useScopedId();
	const name = (part: string) => `${id}-${part}`;
	const paint = (part: string) => `url(#${name(part)})`;
	return (
		<svg
			width={size}
			height={size}
			viewBox="0 0 48 48"
			class={`mt-icon mt-mark-${look.theme}`}
			aria-hidden="true"
		>
			<defs>
				<clipPath id={name("ground")}>
					<rect width="48" height="36.4" />
				</clipPath>
				<Gradient
					id={name("blue-leg")}
					top={17.95}
					bottom={36.4}
					ink={look.blueLeg}
				/>
				<Gradient
					id={name("blue-band")}
					top={17.95}
					bottom={33.38}
					ink={look.blueBand}
				/>
				<Gradient
					id={name("green-leg")}
					top={21.99}
					bottom={36.4}
					ink={look.greenLeg}
				/>
				<Gradient
					id={name("green-band")}
					top={21.99}
					bottom={33.38}
					ink={look.greenBand}
				/>
				<Gradient id={name("star")} top={11.3} bottom={19.9} ink={look.star} />
			</defs>
			<rect {...look.tile} />
			<g fill="none" stroke-width="6.9" stroke-linecap="round">
				<path
					d="M8.61 42.44L20.2 21.4"
					stroke={paint("blue-leg")}
					clip-path={paint("ground")}
				/>
				<path d="M20.2 21.4L24.25 29.93" stroke={paint("blue-band")} />
				<path
					d="M31.28 25.44L39.51 42.62"
					stroke={paint("green-leg")}
					clip-path={paint("ground")}
				/>
				<path
					d="M24.25 29.93L31.28 25.44"
					stroke={paint("green-band")}
					stroke-opacity="0.86"
				/>
			</g>
			<path
				d="M31.3 11.3Q31.8 15.1 36 15.6 31.8 16.1 31.3 19.9 30.8 16.1 26.6 15.6 30.8 15.1 31.3 11.3Z"
				fill={paint("star")}
				stroke={paint("star")}
				stroke-width="0.8"
				stroke-linejoin="round"
			/>
		</svg>
	);
}

// Gradient runs one part of the logo through its ink, from the top to the
// bottom given, measured on the logo's own grid.
function Gradient({
	id,
	top,
	bottom,
	ink: { from, to },
}: {
	id: string;
	top: number;
	bottom: number;
	ink: Ink;
}) {
	return (
		<linearGradient
			id={id}
			x1="0"
			y1={top}
			x2="0"
			y2={bottom}
			gradientUnits="userSpaceOnUse"
		>
			<stop offset="0" stop-color={from} />
			<stop offset="1" stop-color={to} />
		</linearGradient>
	);
}
