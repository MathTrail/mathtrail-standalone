import type { ComponentChildren } from "preact";
import { type Linking, PageLink } from "./links";

/**
 * PartTone is the colour a part of a review is headed in: the green of what
 * goes well, the coral of what to work on, the accent of what to do next, or
 * the muted ink of what cannot be said yet.
 */
export type PartTone = "correct" | "wrong" | "accent" | "muted";

/**
 * ReviewPart is a part of a review under a heading in its tone: the topics
 * that go well, the ones to work on, the ones too early to judge, the mistakes
 * that repeat, or the steps to take.
 */
export function ReviewPart({
	tone,
	label,
	children,
}: {
	tone: PartTone;
	label: string;
	children: ComponentChildren;
}) {
	return (
		<section class="mt-review-part">
			<h3 class="mt-review-label" data-tone={tone}>
				{label}
			</h3>
			{children}
		</section>
	);
}

/**
 * JudgedRow is a topic a review names: its name, and why, when there is a why
 * to say. Its id tells it apart from the others.
 */
export type JudgedRow = { id: string; name: string; line?: string };

/**
 * JudgedList is the topics a review names on one side, each after a dot in
 * the side's colour — a picture of the heading over the list, hidden from a
 * screen reader — with its name and, under it, why.
 */
export function JudgedList({
	tone,
	rows,
}: {
	tone: "correct" | "wrong";
	rows: readonly JudgedRow[];
}) {
	return (
		<ul class="mt-judged">
			{rows.map((row) => (
				<li key={row.id} class="mt-judged-row">
					<span
						class="mt-dot mt-judged-dot"
						data-tone={tone}
						aria-hidden="true"
					/>
					<span class="mt-review-text">
						<span class="mt-judged-name">{row.name}</span>
						{row.line !== undefined && (
							<span class="mt-review-note">{row.line}</span>
						)}
					</span>
				</li>
			))}
		</ul>
	);
}

/**
 * NamedLine is a list of names on a line of their own, as the card's language
 * writes a list, with a note under it of what they have in common.
 */
export function NamedLine({ names, note }: { names: string; note: string }) {
	return (
		<p class="mt-review-text">
			<span class="mt-judged-name">{names}</span>
			<span class="mt-review-note">{note}</span>
		</p>
	);
}

/**
 * AdviceStep is a step a review advises: the topic it is for, when it is for
 * one, with the address of the part of its page the step is about, when it
 * has one to link to, and what to do. Its id tells it apart from the others,
 * since two steps may say the same.
 */
export type AdviceStep = {
	id: string;
	topic?: string;
	href?: string;
	text: string;
};

/**
 * AdviceSteps are the steps a review advises, numbered in the order to take
 * them: each the topic it is for over what to do — a link to the topic's page
 * where the screen links pages at all —, or what to do alone when it holds for
 * every topic.
 */
export function AdviceSteps({
	steps,
	linking,
}: {
	steps: readonly AdviceStep[];
	linking?: Linking;
}) {
	return (
		<ol class="mt-advice">
			{steps.map((step, at) => (
				<li key={step.id} class="mt-advice-step">
					<span class="mt-step-num">{at + 1}</span>
					<span class="mt-review-text">
						{step.topic !== undefined && (
							<span class="mt-advice-topic">
								{step.href !== undefined && linking !== undefined ? (
									<PageLink
										label={step.topic}
										href={step.href}
										linking={linking}
									/>
								) : (
									step.topic
								)}
							</span>
						)}
						<span>{step.text}</span>
					</span>
				</li>
			))}
		</ol>
	);
}
