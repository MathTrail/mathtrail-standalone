import type { OptionState } from "../design/controls";
import { dontKnow, type Letter } from "./choices";
import type { AnswerOutcome, AnswerResult } from "./payload";

/**
 * Answer is where the child's answer to the task stands: not given yet; sent
 * and being checked; recorded, with what the service said of it; refused,
 * because the task is no longer the one being solved; or lost on its way, and
 * free to be given again.
 */
export type Answer =
	| { state: "open" }
	| { state: "checking"; choice: Letter }
	| { state: "answered"; result: AnswerResult }
	| { state: "closed" }
	| { state: "failed" };

/**
 * Lesson is everything a task card knows of the lesson on it: the hint, and
 * the answer.
 */
export type Lesson = {
	hint: { open: boolean; used: boolean };
	answer: Answer;
};

/** lessonStart is a card as a task arrives on it. */
export const lessonStart: Lesson = {
	hint: { open: false, used: false },
	answer: { state: "open" },
};

/** LessonEvent is something that happens on a task card. */
export type LessonEvent =
	| { type: "picked"; choice: Letter }
	| { type: "told"; outcome: AnswerOutcome }
	| { type: "hint toggled" };

/**
 * next is the lesson after event. An event that cannot happen where the
 * lesson stands — a second answer while the first is being checked, the hint
 * of a task done with — leaves it as it is.
 */
export function next(lesson: Lesson, event: LessonEvent): Lesson {
	switch (event.type) {
		case "picked":
			return canAnswer(lesson)
				? { ...lesson, answer: { state: "checking", choice: event.choice } }
				: lesson;
		case "told":
			return lesson.answer.state === "checking"
				? { ...lesson, answer: answerAfter(event.outcome) }
				: lesson;
		case "hint toggled":
			return canAnswer(lesson)
				? { ...lesson, hint: toggled(lesson.hint) }
				: lesson;
	}
}

// toggled is the hint shown if it was hidden and hidden if it was shown, used
// for good once it has been shown.
function toggled(hint: Lesson["hint"]): Lesson["hint"] {
	return { open: !hint.open, used: hint.used || !hint.open };
}

/**
 * canAnswer says whether an answer can be given now: with no answer recorded,
 * refused or already on its way.
 */
export function canAnswer(lesson: Lesson): boolean {
	return lesson.answer.state === "open" || lesson.answer.state === "failed";
}

/**
 * isSettled says whether the task is done with: answered, or no longer the
 * one being solved. A settled task takes no answer and shows no hint, and the
 * one thing left to do on it is to ask for another.
 */
export function isSettled(answer: Answer): boolean {
	return answer.state === "answered" || answer.state === "closed";
}

/**
 * isOver says whether the card has something to say of the answer: its
 * result, why it was refused, or that it could not be checked.
 */
export function isOver(answer: Answer): boolean {
	return isSettled(answer) || answer.state === "failed";
}

// answerAfter is the answer an outcome leaves.
function answerAfter(outcome: AnswerOutcome): Answer {
	switch (outcome.kind) {
		case "answered":
			return { state: "answered", result: outcome.result };
		case "closed":
			return { state: "closed" };
		case "failed":
			return { state: "failed" };
	}
}

/**
 * optionStateOf is how the option letter stands: from the service's result
 * once the answer is recorded — the right option, the child's wrong choice,
 * the rest set back — and, before it, being checked when it is the choice on
 * its way.
 */
export function optionStateOf(answer: Answer, letter: Letter): OptionState {
	if (answer.state === "answered") {
		const { correct_answer: right, choice } = answer.result;
		if (letter === right) {
			return "correct";
		}
		return letter === choice ? "wrong" : "muted";
	}
	if (answer.state === "checking" && answer.choice === letter) {
		return "selected";
	}
	return "default";
}

/**
 * modelLineOf is the one line the model is told once the card shows a
 * recorded answer: which task, which choice, and how it went. It says what is
 * recorded rather than what was just done, since the card also shows an
 * answer recorded before — in the chat, in another card — as it was recorded.
 * The model wrote the task, so it holds the options, the traps and the
 * solution already; after an answer, naming the letters gives nothing away.
 * A mistake the child has made before asks for a reminder of it, in the
 * model's own words, since the card keeps none. And whatever the model says
 * next is worded about the step, addressing nobody, since the model talks
 * with the adult and the child hears it as the adult reads it out; nor does it
 * show whether the child is a boy or a girl, which nothing tells the card: he
 * or she shows it, and in a language with grammatical gender so does a
 * past-tense sentence about what the child did.
 */
export function modelLineOf(result: AnswerResult): string {
	return `${recordedLineOf(result)} ${aboutTheStep}`;
}

const aboutTheStep =
	"Word it about the step, addressing nobody, in short sentences that fit the child's grade, so the adult can read it out as it is, and so it does not show whether the child is a boy or a girl: speak of the child by the pseudonym, never as he or she, praise the step, not the child, and keep to the present tense.";

// recordedLineOf is what the card says is recorded, and how it went.
function recordedLineOf(result: AnswerResult): string {
	const task = `Task ${result.task_id} has its answer recorded`;
	if (result.choice === dontKnow) {
		return `${task}: "I don't know", which counts as a wrong answer; the right option is ${result.correct_answer}. The card shows the solution.`;
	}
	if (result.correct) {
		return `${task}: ${result.choice}, which is right. The card shows the solution.`;
	}
	const shown =
		result.trap === null ? "the solution" : "the trap and the solution";
	const line = `${task}: ${result.choice}, which is wrong; the right option is ${result.correct_answer}. The card shows ${shown}.`;
	if (result.trap?.repeated) {
		return `${line} The child has made this mistake before among the latest answers: end your explanation with one short reminder of it, in your own words, that the child can keep in mind next time.`;
	}
	return line;
}
