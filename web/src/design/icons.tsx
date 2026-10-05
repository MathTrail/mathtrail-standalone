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

/**
 * Mark is MathTrail's logo, the very drawing the site serves as its icon: one
 * ribbon folded into an M on a dark tile, and a star over it. A card loads
 * nothing from outside itself, so it draws the logo in place, and names its
 * gradients and its clip apart from those of any other logo on the page. It
 * is decoration: the name beside it says whose it is.
 */
export function Mark({ size = 32 }: { size?: number }) {
	const id = useScopedId();
	const name = (part: string) => `${id}-${part}`;
	const paint = (part: string) => `url(#${name(part)})`;
	return (
		<svg
			width={size}
			height={size}
			viewBox="0 0 48 48"
			class="mt-icon"
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
					from="#3d8ef3"
					to="#2766d6"
				/>
				<Gradient
					id={name("blue-band")}
					top={17.95}
					bottom={33.38}
					from="#76bcff"
					to="#4791f0"
				/>
				<Gradient
					id={name("green-leg")}
					top={21.99}
					bottom={36.4}
					from="#b9e276"
					to="#76c05b"
				/>
				<Gradient
					id={name("green-band")}
					top={21.99}
					bottom={33.38}
					from="#86ce69"
					to="#57ae55"
				/>
				<Gradient
					id={name("star")}
					top={11.3}
					bottom={19.9}
					from="#ffcd6e"
					to="#f5a238"
				/>
			</defs>
			<rect width="48" height="48" rx="6" fill="#0c1a33" />
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

// Gradient runs one part of the logo from its colour at the top to its colour
// at the bottom, measured on the logo's own grid.
function Gradient({
	id,
	top,
	bottom,
	from,
	to,
}: {
	id: string;
	top: number;
	bottom: number;
	from: string;
	to: string;
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
