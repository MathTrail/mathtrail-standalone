import { classes } from "./classes";

// The line icons are one stroke each, on a grid of 16, in the colour of the
// text around them: a state that recolours its text recolours its icon too.
const lines = {
	"chevron-right": { d: "M6 4l4 4-4 4", width: 1.75 },
	"chevron-left": { d: "M10 4l-4 4 4 4", width: 1.75 },
	check: { d: "M3.5 8.5l3 3 6-7", width: 2 },
	cross: { d: "M4.5 4.5l7 7M11.5 4.5l-7 7", width: 2 },
	dash: { d: "M4.5 8h7", width: 2 },
	hint: {
		d: "M6 12.5h4M6.5 14.5h3M8 1.5a4.5 4.5 0 0 0-2.6 8.2c.4.3.6.7.6 1.1v.2h4v-.2c0-.4.2-.8.6-1.1A4.5 4.5 0 0 0 8 1.5z",
		width: 1.4,
	},
	trap: { d: "M8 2.2L14.3 13.3H1.7L8 2.2zM8 6.5v3M8 11.4v.1", width: 1.4 },
} as const;

// The marks of a verdict, and of a step done, are a tick or cross on a disc:
// a verdict's in its own colour, a step's in the colour of the strongest ink.
const marks = {
	"verdict-correct": {
		fill: "var(--correct)",
		ink: "var(--on-signal)",
		d: "M6 10.2l2.6 2.6L14 7.4",
	},
	"verdict-wrong": {
		fill: "var(--wrong)",
		ink: "var(--on-signal)",
		d: "M7 7l6 6M13 7l-6 6",
	},
	"step-done": {
		fill: "var(--ink-strong)",
		ink: "var(--surface)",
		d: "M6 10.2l2.6 2.6L14 7.4",
	},
} as const;

/**
 * IconName names an icon of the design: a line icon, the spinner of a check
 * under way, the mark of a verdict, or how a step stands — done, or to come.
 */
export type IconName =
	| keyof typeof lines
	| keyof typeof marks
	| "spinner"
	| "step-waiting";

/**
 * Icon is one of the design's icons. It is decoration, hidden from a screen
 * reader: the words beside it say what it shows. The spinner's track takes
 * the colour a token names, so that it shows on the tint of the row it turns
 * in.
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
	if (name === "spinner") {
		return (
			<svg
				width={size}
				height={size}
				viewBox="0 0 20 20"
				fill="none"
				aria-hidden="true"
				class={classes("mt-icon", "mt-spin", className)}
			>
				<circle
					cx="10"
					cy="10"
					r="8"
					stroke={`var(--${track})`}
					stroke-width="2.5"
				/>
				<path
					d="M10 2a8 8 0 0 1 8 8"
					stroke="currentColor"
					stroke-width="2.5"
					stroke-linecap="round"
				/>
			</svg>
		);
	}
	if (name === "step-waiting") {
		return (
			<svg
				width={size}
				height={size}
				viewBox="0 0 20 20"
				fill="none"
				aria-hidden="true"
				class={classes("mt-icon", className)}
			>
				<circle
					cx="10"
					cy="10"
					r="8.5"
					stroke="var(--track)"
					stroke-width="1.5"
				/>
			</svg>
		);
	}
	if (isMark(name)) {
		const { fill, ink, d } = marks[name];
		return (
			<svg
				width={size}
				height={size}
				viewBox="0 0 20 20"
				fill="none"
				aria-hidden="true"
				class={classes("mt-icon", className)}
			>
				<circle cx="10" cy="10" r="10" fill={fill} />
				<path
					d={d}
					stroke={ink}
					stroke-width="2"
					stroke-linecap="round"
					stroke-linejoin="round"
				/>
			</svg>
		);
	}
	const { d, width } = lines[name];
	return (
		<svg
			width={size}
			height={size}
			viewBox="0 0 16 16"
			fill="none"
			aria-hidden="true"
			class={classes("mt-icon", className)}
		>
			<path
				d={d}
				fill="none"
				stroke="currentColor"
				stroke-width={width}
				stroke-linecap="round"
				stroke-linejoin="round"
			/>
		</svg>
	);
}

// isMark says whether name is one of the marks drawn on a disc.
function isMark(name: IconName): name is keyof typeof marks {
	return Object.hasOwn(marks, name);
}

/**
 * Mark is MathTrail's sign: a trail rising to a point, on a disc. It is
 * decoration: the name beside it says whose it is.
 */
export function Mark({ size = 32 }: { size?: number }) {
	return (
		<svg
			width={size}
			height={size}
			viewBox="0 0 32 32"
			class="mt-icon"
			aria-hidden="true"
		>
			<circle cx="16" cy="16" r="16" fill="var(--mark-fill)" />
			<path
				d="M8.5 21.5L13.5 16.5L17 19L22.5 12.5"
				fill="none"
				stroke="var(--mark-ink)"
				stroke-width="2.2"
				stroke-linecap="round"
				stroke-linejoin="round"
			/>
			<circle cx="23" cy="12" r="2.6" fill="var(--mark-ink)" />
		</svg>
	);
}

/**
 * Avatar is the child's picture, the same for every child: a card shows a
 * pseudonym and nothing that could tell children apart.
 */
export function Avatar({ size = 24 }: { size?: number }) {
	return (
		<svg
			width={size}
			height={size}
			viewBox="0 0 24 24"
			aria-hidden="true"
			class="mt-icon"
		>
			<circle cx="12" cy="12" r="12" fill="var(--avatar-fill)" />
			<circle cx="12" cy="9.5" r="3.6" fill="var(--avatar-ink)" />
			<path
				d="M5 19.6c1.3-3.2 3.9-4.8 7-4.8s5.7 1.6 7 4.8"
				fill="var(--avatar-ink)"
			/>
		</svg>
	);
}
