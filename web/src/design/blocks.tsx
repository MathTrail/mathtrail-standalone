import type { ComponentChildren } from "preact";
import { classes } from "./classes";
import type { Said } from "./controls";
import { Icon } from "./icons";
import { useScopedId } from "./ids";

/** NoteTone is what a note carries: a plain line, the hint, or the trap. */
export type NoteTone = "plain" | "hint" | "trap";

/**
 * Note is a short labelled block: the hint, set aside from the task, the trap
 * behind a wrong answer, or a plain remark. Its words are in the language
 * said; its label, and the detail under the words when there is one, in the
 * card's.
 */
export function Note({
	tone = "plain",
	label,
	said = {},
	detail,
	children,
}: {
	tone?: NoteTone;
	label: string;
	said?: Said;
	detail?: string;
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
			{detail !== undefined && <p class="mt-note-detail">{detail}</p>}
		</Element>
	);
}

/**
 * VerdictTone is how an answer went: right, wrong, or told without a verdict,
 * as the solution of a task the child did not know is.
 */
export type VerdictTone = "correct" | "wrong" | "none";

/**
 * Verdict is the line that says how something went, with its mark when it is
 * an answer's, and the detail under it when there is more to say.
 */
export function Verdict({
	tone = "none",
	detail,
	children,
}: {
	tone?: VerdictTone;
	detail?: string;
	children: ComponentChildren;
}) {
	const line = (
		<p class="mt-verdict-line">
			{tone !== "none" && <Icon name={`verdict-${tone}`} size={20} />}
			<span>{children}</span>
		</p>
	);
	if (detail === undefined) {
		return line;
	}
	return (
		<div class="mt-verdict">
			{line}
			<p class="mt-verdict-detail">{detail}</p>
		</div>
	);
}

/**
 * Fold is a part of a card that folds away under its title. The title, a
 * short summary of what the part holds and an arrow are one button, which
 * opens the part and folds it again, and says which it is; the summary goes
 * to a line of its own where the two do not fit. Folded, what the part holds
 * stays on the page, hidden, so that whatever is half done in it — a form
 * being filled in — is there as it was when the part opens again. Whether it
 * is open is not the fold's to remember: whoever draws it says.
 */
export function Fold({
	title,
	summary,
	open,
	onToggle,
	children,
}: {
	title: string;
	summary?: ComponentChildren;
	open: boolean;
	onToggle: () => void;
	children: ComponentChildren;
}) {
	const body = useScopedId();
	return (
		<div class="mt-fold">
			<h2 class="mt-fold-head">
				<button
					type="button"
					class="mt-fold-button"
					aria-expanded={open}
					aria-controls={body}
					onClick={onToggle}
				>
					<span class="mt-fold-text">
						<span class="mt-fold-title">{title}</span>
						{summary !== undefined && (
							<>
								{" "}
								<span class="mt-fold-summary">{summary}</span>
							</>
						)}
					</span>
					<Icon name="chevron-right" size={16} className="mt-chevron" />
				</button>
			</h2>
			<div id={body} class="mt-fold-body" hidden={!open}>
				{children}
			</div>
		</div>
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

/** StepStatus is how a step of a course stands: done, under way, or to come. */
export type StepStatus = "done" | "active" | "waiting";

/** Step is a step of a course being followed, and how it stands. */
export type Step = { label: string; status: StepStatus };

/**
 * GeneratingSteps is a course followed step by step under its title: each
 * step done, under way or to come, with its mark. A screen reader is told how
 * each step stands, in the words given for it, and hears the list as it moves.
 */
export function GeneratingSteps({
	title,
	steps,
	statusLabels,
}: {
	title: string;
	steps: readonly Step[];
	statusLabels: Record<StepStatus, string>;
}) {
	return (
		<div class="mt-gen">
			<p class="mt-gen-title">{title}</p>
			<ol aria-live="polite">
				{steps.map((step) => (
					<li key={step.label} data-status={step.status}>
						<span class="mt-gen-icon">{stepMark(step.status)}</span>
						<span class="mt-gen-label">
							<span class="mt-vh">{statusLabels[step.status]} </span>
							{step.label}
						</span>
					</li>
				))}
			</ol>
		</div>
	);
}

// stepMark is the mark of a step as it stands: ticked when done, turning in
// the strongest ink while under way, an empty ring while to come.
function stepMark(status: StepStatus) {
	switch (status) {
		case "done":
			return <Icon name="step-done" size={20} />;
		case "active":
			return (
				<span class="mt-gen-turning">
					<Icon name="spinner" size={20} />
				</span>
			);
		case "waiting":
			return <Icon name="step-waiting" size={20} />;
	}
}
