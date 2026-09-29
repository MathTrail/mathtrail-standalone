import type { OptionState } from "../design/controls";
import {
	type AnswerOutcome,
	type AnswerResult,
	type Choice,
	dontKnow,
	type Letter,
} from "./payload";

/**
 * Answer is where the child's answer to the task stands: not given yet; sent
 * and being checked; recorded, with what the service said of it; refused,
 * because the task is no longer the one being solved; or lost on its way, and
 * free to be given again.
 */
export type Answer =
	| { state: "open" }
	| { state: "checking"; choice: Choice }
	| { state: "answered"; result: AnswerResult }
	| { state: "closed" }
	| { state: "failed" };

/** Question is a question the child typed into the card, and whether it got to the chat. */
export type Question = {
	id: number;
	words: string;
	state: "sending" | "sent" | "lost";
};

/**
 * Progress is the progress shown over the task, as it is read: being read,
 * read, or not read. Each opening is numbered, so that the reply to one closed
 * since is not taken for the reply to the one open now.
 */
export type Progress =
	| { state: "reading"; opening: number }
	| { state: "read"; opening: number; payload: unknown }
	| { state: "failed"; opening: number };

/**
 * Lesson is everything a task card knows of the lesson on it: whether it
 * still shows the task or waits for the next one, the hint, the answer, the
 * questions asked, the words being typed, and the progress when it is open
 * over the task.
 */
export type Lesson = {
	stage: "task" | "waiting";
	hint: { open: boolean; used: boolean };
	answer: Answer;
	questions: readonly Question[];
	draft: string;
	progress: Progress | undefined;
};

/** lessonStart is a card as a task arrives on it. */
export const lessonStart: Lesson = {
	stage: "task",
	hint: { open: false, used: false },
	answer: { state: "open" },
	questions: [],
	draft: "",
	progress: undefined,
};

/** LessonEvent is something that happens on a task card. */
export type LessonEvent =
	| { type: "picked"; choice: Choice }
	| { type: "told"; outcome: AnswerOutcome }
	| { type: "hint toggled" }
	| { type: "typed"; words: string }
	| { type: "asked"; id: number; words: string }
	| { type: "question sent"; id: number }
	| { type: "question lost"; id: number }
	| { type: "another asked" }
	| { type: "progress opened"; opening: number }
	| { type: "progress read"; opening: number; payload: unknown }
	| { type: "progress failed"; opening: number }
	| { type: "back to task" };

/**
 * next is the lesson after event. An event that cannot happen where the
 * lesson stands — a second answer while the first is being checked, a reply
 * for a progress already closed — leaves it as it is.
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
				? {
						...lesson,
						hint: {
							open: !lesson.hint.open,
							used: lesson.hint.used || !lesson.hint.open,
						},
					}
				: lesson;
		case "typed":
			return { ...lesson, draft: event.words };
		case "asked":
			return {
				...lesson,
				questions: [
					...lesson.questions,
					{ id: event.id, words: event.words, state: "sending" },
				],
				draft: "",
			};
		case "question sent":
			return questionNow(lesson, event.id, "sent");
		case "question lost":
			return questionNow(lesson, event.id, "lost");
		case "another asked":
			return lesson.stage === "task" && lesson.answer.state !== "checking"
				? { ...lesson, stage: "waiting" }
				: lesson;
		case "progress opened":
			return lesson.progress === undefined
				? {
						...lesson,
						progress: { state: "reading", opening: event.opening },
					}
				: lesson;
		case "progress read":
			return isRead(lesson, event.opening)
				? {
						...lesson,
						progress: {
							state: "read",
							opening: event.opening,
							payload: event.payload,
						},
					}
				: lesson;
		case "progress failed":
			return isRead(lesson, event.opening)
				? {
						...lesson,
						progress: { state: "failed", opening: event.opening },
					}
				: lesson;
		case "back to task":
			return { ...lesson, progress: undefined };
	}
}

/**
 * canAnswer says whether an answer can be given now: on the task, with the
 * progress closed, and no answer recorded, refused or already on its way.
 */
export function canAnswer(lesson: Lesson): boolean {
	return (
		lesson.stage === "task" &&
		lesson.progress === undefined &&
		(lesson.answer.state === "open" || lesson.answer.state === "failed")
	);
}

// isRead says whether the progress opened as opening is the one being read.
function isRead(lesson: Lesson, opening: number): boolean {
	return (
		lesson.progress?.state === "reading" && lesson.progress.opening === opening
	);
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

// questionNow is the lesson with the question id in state. A question that
// did not reach the chat gives its words back to an empty field, to be sent
// again or changed; words typed since then are not overwritten.
function questionNow(
	lesson: Lesson,
	id: number,
	state: "sent" | "lost",
): Lesson {
	const question = lesson.questions.find((asked) => asked.id === id);
	if (question === undefined || question.state !== "sending") {
		return lesson;
	}
	return {
		...lesson,
		questions: lesson.questions.map((asked) =>
			asked.id === id ? { ...asked, state } : asked,
		),
		draft:
			state === "lost" && lesson.draft === "" ? question.words : lesson.draft,
	};
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
 */
export function modelLineOf(result: AnswerResult): string {
	const task = `Task ${result.task_id} has its answer recorded`;
	if (result.choice === dontKnow) {
		return `${task}: "I don't know", which counts as a wrong answer; the right option is ${result.correct_answer}. The card shows the solution.`;
	}
	if (result.correct) {
		return `${task}: ${result.choice}, which is right. The card shows the solution.`;
	}
	const shown =
		result.trap === null ? "the solution" : "the trap and the solution";
	return `${task}: ${result.choice}, which is wrong; the right option is ${result.correct_answer}. The card shows ${shown}.`;
}
