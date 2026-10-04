import { StatList, type StatRow } from "../design/progress";
import {
	type AdviceStep,
	AdviceSteps,
	JudgedList,
	type JudgedRow,
	NamedLine,
	ReviewPart,
} from "../design/review";
import type { Words } from "../i18n/words";
import { listed, topicName, trapAdvice, trapName } from "./names";
import type { Judged, Review, ReviewStep } from "./payload";
import type { Key } from "./words";

/**
 * ReviewSaid is the review of the topics as the card says it: the strong
 * topics and the ones to develop, each with why, the ones too early to judge,
 * the mistakes that repeat, and the steps to take.
 */
export type ReviewSaid = {
	strong: JudgedRow[];
	develop: JudgedRow[];
	early: string[];
	mistakes: StatRow[];
	steps: AdviceStep[];
};

/**
 * reviewSaid is the review in the card's words, with the mistakes that repeat
 * as their lines. What the card has no words for — a reason, a kind of step or
 * the advice for a mistake that a later release adds — is left out rather than
 * said wrong; a topic named for no reason the card can say keeps its name.
 */
export function reviewSaid(
	words: Words<Key>,
	review: Review,
	mistakes: StatRow[],
): ReviewSaid {
	return {
		strong: review.strong.map((topic) => judgedRow(words, topic)),
		develop: review.develop.map((topic) => judgedRow(words, topic)),
		early: review.early.map((topic) => topicName(words, topic)),
		mistakes,
		steps: review.steps.flatMap((step, at) => adviceStep(words, step, at)),
	};
}

/** saysNothing says whether a review said in the card's words is empty. */
export function saysNothing(said: ReviewSaid): boolean {
	return (
		said.strong.length === 0 &&
		said.develop.length === 0 &&
		said.early.length === 0 &&
		said.mistakes.length === 0 &&
		said.steps.length === 0
	);
}

/**
 * ReviewParts are the parts of the review, in the order the adult reads them:
 * what goes well, what to develop, what is too early to judge, the mistakes
 * that repeat, and what to do next. A part with nothing in it is not drawn.
 */
export function ReviewParts({
	words,
	said,
}: {
	words: Words<Key>;
	said: ReviewSaid;
}) {
	return (
		<>
			{said.strong.length > 0 && (
				<ReviewPart tone="correct" label={words.text("review.strong")}>
					<JudgedList tone="correct" rows={said.strong} />
				</ReviewPart>
			)}
			{said.develop.length > 0 && (
				<ReviewPart tone="wrong" label={words.text("review.develop")}>
					<JudgedList tone="wrong" rows={said.develop} />
				</ReviewPart>
			)}
			{said.early.length > 0 && (
				<ReviewPart tone="muted" label={words.text("review.early")}>
					<NamedLine
						names={listed(words, said.early)}
						note={words.text("review.early_note")}
					/>
				</ReviewPart>
			)}
			{said.mistakes.length > 0 && (
				<ReviewPart tone="wrong" label={words.text("progress.mistakes")}>
					<StatList rows={said.mistakes} framed />
				</ReviewPart>
			)}
			{said.steps.length > 0 && (
				<ReviewPart tone="accent" label={words.text("review.next")}>
					<AdviceSteps steps={said.steps} />
				</ReviewPart>
			)}
		</>
	);
}

// The words of each reason a topic is named for, but the trap's, whose words
// name the trap. A map, since a code is any text the service sends.
const reasonWords: ReadonlyMap<string, Key> = new Map([
	["mastered", "review.reason_mastered"],
	["high", "review.reason_high"],
	["rose", "review.reason_rose"],
	["low", "review.reason_low"],
	["failures", "review.reason_failures"],
	["hints", "review.reason_hints"],
	["fell", "review.reason_fell"],
]);

// judgedRow is a topic the review names, as the card says it: its name, and
// why — a sentence for each reason, in the order the service gives them, and
// one that a move has begun for a topic whose last answer was right.
function judgedRow(words: Words<Key>, topic: Judged): JudgedRow {
	const why = topic.reasons.flatMap((reason) =>
		reasonSaid(words, reason, topic.trap),
	);
	if (topic.moving === true) {
		why.push(words.text("review.moving"));
	}
	return {
		id: topic.topic,
		name: topicName(words, topic.topic),
		line: why.length > 0 ? why.join(" ") : undefined,
	};
}

// reasonSaid is a reason in a sentence of the card's language: the trap's with
// the mistake named, and none for a reason the card has no words for.
function reasonSaid(
	words: Words<Key>,
	reason: string,
	trap: string | undefined,
): string[] {
	if (reason === "trap") {
		return trap === undefined
			? []
			: [words.text("review.reason_trap", { trap: trapName(words, trap) })];
	}
	const key = reasonWords.get(reason);
	return key === undefined ? [] : [words.text(key)];
}

// The words of each kind of step that is no trap's advice.
const stepWords: ReadonlyMap<string, Key> = new Map([
	["rhythm", "review.step_rhythm"],
	["practice", "review.step_practice"],
	["unaided", "review.step_unaided"],
]);

// adviceStep is a step the review advises, as the card says it — the topic it
// is for, and what to do — or none when the card has no words for what it
// advises.
function adviceStep(
	words: Words<Key>,
	step: ReviewStep,
	at: number,
): AdviceStep[] {
	const text = stepSaid(words, step);
	if (text === undefined) {
		return [];
	}
	return [
		{
			id: String(at),
			topic:
				step.topic === undefined ? undefined : topicName(words, step.topic),
			text,
		},
	];
}

// stepSaid is what a step advises, in the card's words: a trap's advice, or
// the words of the step's kind.
function stepSaid(words: Words<Key>, step: ReviewStep): string | undefined {
	if (step.kind === "trap") {
		return step.trap === undefined ? undefined : trapAdvice(words, step.trap);
	}
	const key = stepWords.get(step.kind);
	return key === undefined ? undefined : words.text(key);
}
