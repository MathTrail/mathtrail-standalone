import type { ComponentChildren } from "preact";
import { useSiteWords } from "./words";

/**
 * The parts the site's drawings of a topic are built of, where no kind of
 * picture the card draws shows what a step means: numbers and words in boxes,
 * the signs between them, groups ringed with what they make, arrows, the
 * lines of words under a drawing, and the rows a drawing sets side by side.
 * Every part is markup with a class, its colours the stylesheet's, and a
 * drawing is hidden from a screen reader by the frame it stands in.
 */

/**
 * Tone is the colour a box is drawn in: plain on white, cool and warm for two
 * kinds of thing side by side, green and red for a knight and a liar or for
 * right and wrong, grey for a cell left out, pale for a box on a tinted
 * ground, faint for a box still to fill, and unknown, dashed, for what the
 * task asks.
 */
export type Tone =
	| "plain"
	| "cool"
	| "warm"
	| "green"
	| "red"
	| "grey"
	| "pale"
	| "faint"
	| "unknown";

/** Size is how large a box is, from the largest to the smallest. */
export type Size = "xl" | "lg" | "md" | "sm" | "xs" | "xxs" | "mini";

/**
 * Ring is what a box is ringed as: the one a step finds, the right answer,
 * or a wrong one.
 */
export type Ring = "picked" | "right" | "wrong";

/** Gap is how far apart the parts of a row or a column stand, in pixels. */
export type Gap = 1 | 2 | 3 | 4 | 6 | 8 | 10 | 12 | 14 | 16 | 20 | 24;

/** Box is a number, a word or a sign in a box of a tone and a size. */
export function Box({
	tone = "plain",
	size = "md",
	ring,
	children,
}: {
	tone?: Tone;
	size?: Size;
	ring?: Ring;
	children?: ComponentChildren;
}) {
	return (
		<span class="s-sk-box" data-tone={tone} data-size={size} data-ring={ring}>
			{children}
		</span>
	);
}

/**
 * Act is one action of a chain, such as adding 3 or halving: plain where it
 * is done and warm where it undoes another.
 */
export function Act({
	tone = "plain",
	children,
}: {
	tone?: "plain" | "warm";
	children: ComponentChildren;
}) {
	return (
		<span class="s-sk-act" data-tone={tone}>
			{children}
		</span>
	);
}

/** Unknown is the number a chain starts from and the task asks for. */
export function Unknown() {
	return <span class="s-sk-unknown">?</span>;
}

/** Op is a sign between two boxes, such as plus or equals. */
export function Op({ children }: { children: ComponentChildren }) {
	return <span class="s-sk-op">{children}</span>;
}

/**
 * Say is a line of words under a drawing or between its rows: red under what
 * is wrong, and green under what is right.
 */
export function Say({
	way,
	children,
}: {
	way?: "wrong" | "right";
	children: ComponentChildren;
}) {
	return (
		<span class="s-sk-say" data-way={way}>
			{children}
		</span>
	);
}

/**
 * Down is the arrow from one row of a drawing down to the next, beside the
 * words of what happens between them, if any.
 */
export function Down({ children }: { children?: ComponentChildren }) {
	return (
		<span class="s-sk-down">
			<span class="s-sk-down-arrow">↓</span>
			{children}
		</span>
	);
}

/**
 * Group is boxes ringed together in a tone, with what they make written
 * under the ring.
 */
export function Group({
	tone = "cool",
	label,
	children,
}: {
	tone?: "cool" | "warm" | "green" | "plain";
	label?: ComponentChildren;
	children: ComponentChildren;
}) {
	return (
		<span class="s-sk-group" data-tone={tone}>
			<span class="s-sk-ring">{children}</span>
			{label !== undefined && <span class="s-sk-label">{label}</span>}
		</span>
	);
}

/**
 * Then is the arrow from one part of a drawing to the next: forward, or back
 * where a chain is undone, plain or in the warm of undoing.
 */
export function Then({
	back = false,
	tone = "plain",
}: {
	back?: boolean;
	tone?: "plain" | "warm" | "cool";
}) {
	return (
		<span
			class="s-sk-then"
			data-back={back ? "" : undefined}
			data-tone={tone}
		/>
	);
}

/**
 * Line is parts of a drawing side by side, in the middle of their height,
 * going on to the next line where a narrow screen leaves no room.
 */
export function Line({
	gap = 6,
	align,
	children,
}: {
	gap?: Gap;
	align?: "start" | "end";
	children: ComponentChildren;
}) {
	return (
		<span class="s-sk-line" data-gap={gap} data-align={align}>
			{children}
		</span>
	);
}

/** Stack is parts of a drawing one under another, in the middle of its width. */
export function Stack({
	gap = 8,
	align,
	children,
}: {
	gap?: Gap;
	align?: "start" | "end";
	children: ComponentChildren;
}) {
	return (
		<span class="s-sk-stack" data-gap={gap} data-align={align}>
			{children}
		</span>
	);
}

/**
 * Name is the name of a row of a drawing at its start: who did it, or which
 * way it goes. A small name heads a row of a chain, and a wide one a row whose
 * name is a few words.
 */
export function Name({
	way,
	size,
	children,
}: {
	way?: "wrong" | "right";
	size?: "small" | "wide";
	children: ComponentChildren;
}) {
	return (
		<span class="s-sk-name" data-way={way} data-size={size}>
			{children}
		</span>
	);
}

/**
 * Row is a named row of a drawing: its name at its start, then its parts.
 */
export function Row({
	name,
	way,
	size,
	gap = 6,
	children,
}: {
	name: ComponentChildren;
	way?: "wrong" | "right";
	size?: "small" | "wide";
	gap?: Gap;
	children: ComponentChildren;
}) {
	return (
		<span class="s-sk-row" data-size={size}>
			<Name way={way} size={size}>
				{name}
			</Name>
			<Line gap={gap}>{children}</Line>
		</span>
	);
}

/**
 * Versus is a trap set against the right way: what the child did, then what
 * is right, each row under its name.
 */
export function Versus({
	child,
	right,
}: {
	child: ComponentChildren;
	right: ComponentChildren;
}) {
	const words = useSiteWords();
	return (
		<>
			<Row name={words.text("topic.child")} way="wrong">
				{child}
			</Row>
			<Row name={words.text("topic.right")} way="right">
				{right}
			</Row>
		</>
	);
}

/**
 * Total is a number written under the boxes above it, in the accent of what
 * they add up to.
 */
export function Total({ children }: { children: ComponentChildren }) {
	return <span class="s-sk-total">{children}</span>;
}

/**
 * Dots are so many dots in rows of so many, in a tone: a count drawn as an
 * array.
 */
export function Dots({
	count,
	columns,
	tone = "cool",
}: {
	count: number;
	columns: number;
	tone?: "cool" | "warm" | "green" | "grey";
}) {
	return (
		<span
			class="s-sk-dots"
			data-tone={tone}
			style={{ "--s-columns": String(columns) }}
		>
			{Array.from({ length: count }, (_, at) => (
				<span key={at} />
			))}
		</span>
	);
}

/**
 * Swatch is a block of a tone and a length, in pixels: the sample of a key
 * that names what a colour of a drawing stands for.
 */
export function Swatch({
	tone = "cool",
	length = 18,
}: {
	tone?: "cool" | "warm" | "pale";
	length?: number;
}) {
	return (
		<span
			class="s-sk-swatch"
			data-tone={tone}
			style={{ "--s-length": `${length}px` }}
		/>
	);
}

/**
 * Stretch is how far it is from one point of a line to the next: an arrow of
 * a length, in pixels, with what it measures written over it, level with the
 * boxes it runs between; back where it runs backwards, and red where it is
 * wrong. Where a narrow screen leaves no room it grows shorter, but never
 * shorter than its words.
 */
export function Stretch({
	length,
	back = false,
	wrong = false,
	children,
}: {
	length: number;
	back?: boolean;
	wrong?: boolean;
	children?: ComponentChildren;
}) {
	const labelled = children !== undefined;
	return (
		<span
			class="s-sk-stretch"
			data-back={back ? "" : undefined}
			data-wrong={wrong ? "" : undefined}
			data-labelled={labelled ? "" : undefined}
			style={{ "--s-length": `${length}px` }}
		>
			{labelled && <span class="s-sk-stretch-label">{children}</span>}
			<span class="s-sk-stretch-arrow" />
		</span>
	);
}

/**
 * Round is parts of a drawing set evenly round a circle of a size, in
 * pixels, the first at its top and the rest clockwise: on a table drawn in
 * its middle, or on the line of the circle itself; with what is written in
 * its middle, if anything.
 */
export function Round({
	size,
	track = "table",
	middle,
	children,
}: {
	size: number;
	track?: "table" | "line";
	middle?: ComponentChildren;
	children: readonly ComponentChildren[];
}) {
	const count = children.length;
	return (
		<span
			class="s-sk-round"
			data-track={track}
			style={{ "--s-size": `${size}px` }}
		>
			{middle !== undefined && <span class="s-sk-round-middle">{middle}</span>}
			{children.map((child, at) => {
				const turn = (at / count) * 2 * Math.PI;
				return (
					<span
						key={turn}
						class="s-sk-seat"
						style={{
							"--s-x": `${(50 + 42 * Math.sin(turn)).toFixed(2)}%`,
							"--s-y": `${(50 - 42 * Math.cos(turn)).toFixed(2)}%`,
						}}
					>
						{child}
					</span>
				);
			})}
		</span>
	);
}

/** Chip is a short word on a pill of a tone, such as a weekday or a yes. */
export function Chip({
	tone = "plain",
	children,
}: {
	tone?: Tone;
	children: ComponentChildren;
}) {
	return (
		<span class="s-sk-chip" data-tone={tone}>
			{children}
		</span>
	);
}

/** Key is one entry of a drawing's key: what a sample of it means. */
export function Key({
	sample,
	children,
}: {
	sample: ComponentChildren;
	children: ComponentChildren;
}) {
	return (
		<span class="s-sk-key">
			{sample}
			<span>{children}</span>
		</span>
	);
}
