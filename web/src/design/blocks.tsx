import type { ComponentChildren } from "preact";
import { classes } from "./classes";
import type { Said } from "./controls";
import { Icon } from "./icons";

/**
 * Diagram is a task's text drawing: a grid of characters whose meaning is
 * where each one stands, so it is laid out left to right in every language and
 * never wrapped. One wider than the card scrolls sideways inside its own frame,
 * which a keyboard can reach to scroll. A screen reader is told it is a
 * drawing rather than read it character by character: the wording of a task
 * carries every fact its drawing shows.
 */
export function Diagram({
	drawing,
	label,
}: {
	drawing: string;
	label: string;
}) {
	return (
		<pre
			dir="ltr"
			class="mt-diagram"
			role="img"
			aria-label={label}
			// biome-ignore lint/a11y/noNoninteractiveTabindex: a drawing wider than the card scrolls inside its frame, and a keyboard has to reach it to scroll it
			tabIndex={0}
		>
			{drawing.replace(/\n$/, "")}
		</pre>
	);
}

/** NoteTone is what a note carries: a plain line, the hint, or the trap. */
export type NoteTone = "plain" | "hint" | "trap";

/**
 * Note is a short labelled block: the hint, set aside from the task, the trap
 * behind a wrong answer, or a plain remark. Its words are in the language
 * said, its label in the card's.
 */
export function Note({
	tone = "plain",
	label,
	said = {},
	children,
}: {
	tone?: NoteTone;
	label: string;
	said?: Said;
	children: ComponentChildren;
}) {
	const Element = tone === "hint" ? "aside" : "div";
	return (
		<Element class={classes("mt-note", `mt-note-${tone}`)}>
			<div class="mt-note-label">
				{tone !== "plain" && <Icon name={tone} size={16} />}
				<span>{label}</span>
			</div>
			<p lang={said.lang} dir={said.dir}>
				{children}
			</p>
		</Element>
	);
}

/**
 * VerdictTone is how an answer went: right, wrong, or told without a verdict,
 * as the solution of a task the child did not know is.
 */
export type VerdictTone = "correct" | "wrong" | "none";

/** Verdict is the line that says how the answer went, with its mark. */
export function Verdict({
	tone = "none",
	children,
}: {
	tone?: VerdictTone;
	children: ComponentChildren;
}) {
	return (
		<p class="mt-verdict-line">
			{tone !== "none" && <Icon name={`verdict-${tone}`} size={20} />}
			<span>{children}</span>
		</p>
	);
}

/**
 * SolutionSteps is a solution told step by step, each step numbered: the
 * steps in the language said, the label in the card's.
 */
export function SolutionSteps({
	label,
	steps,
	said = {},
}: {
	label: string;
	steps: readonly string[];
	said?: Said;
}) {
	return (
		<div class="mt-steps">
			<h3 class="mt-section-label">{label}</h3>
			<ol lang={said.lang} dir={said.dir}>
				{numbered(steps).map(({ number, step }) => (
					<li key={number}>
						<span class="mt-step-num">{number}</span> <span>{step}</span>
					</li>
				))}
			</ol>
		</div>
	);
}

// numbered are the steps with the numbers the list shows them under, counted
// from one; a step's number is also what tells it apart, since two steps may
// say the same.
function numbered(steps: readonly string[]) {
	return steps.map((step, at) => ({ number: at + 1, step }));
}
