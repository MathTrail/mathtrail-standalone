import type { HandedTask, TaskStatus } from "./payload";
import { moments } from "./waiting";

/**
 * Wait is what a card that waits for a task has been told of it: whether it
 * has heard the task is being written, how many tries the checks have turned
 * down, how many answers in a row have said no task is coming, and the task,
 * once it is on the card. It also counts the answers it has had, so that each
 * answer sets the next question going, and its news, so that the clock that
 * says the wait is long starts again with each; and it keeps how many
 * questions in a row went unanswered, whether the wait has gone long, and
 * whether questions held for news went unanswered twice in a row, after which
 * the card asks for none to be held.
 */
export type Wait = {
	seenWriting: boolean;
	refused: number;
	notComing: number;
	task: HandedTask | undefined;
	answers: number;
	news: number;
	unanswered: number;
	slow: boolean;
	unheld: boolean;
};

/** waitStart is a card that has asked nothing yet. */
export const waitStart: Wait = {
	seenWriting: false,
	refused: 0,
	notComing: 0,
	task: undefined,
	answers: 0,
	news: 0,
	unanswered: 0,
	slow: false,
	unheld: false,
};

/**
 * heldOn is what a card's next question says it has heard, so that the
 * service may hold it until there is news: the tries turned down, once the
 * card has heard the task is being written and nothing since has said none is
 * coming. Nothing, so that the question is answered at once, before that, and
 * for good once two questions in a row went unanswered while held for news —
 * a host may cut a call that takes a while, and a card asking as it did before
 * still finds its task. One such question alone may have met a moment's
 * failure, and is no reason to give the holding up.
 */
export function heldOn(wait: Wait): number | undefined {
	return wait.seenWriting && wait.notComing === 0 && !wait.unheld
		? wait.refused
		: undefined;
}

/**
 * WaitEvent is what moves a wait on: an answer to the card's question, or the
 * clock of the news it names running out with no other news since.
 */
export type WaitEvent =
	| { type: "answered"; status: TaskStatus }
	| { type: "slowed"; news: number };

// confirmed is how many answers in a row it takes to settle that no task is
// coming: a read straight after a write may bring the file as it was before
// it, and one more question tells that from a request that is over.
const confirmed = 2;

// unansweredMost is how many questions in a row may go unanswered before the
// card stops asking: five minutes at the slow pace, long past the time a task
// takes. A service the card cannot reach is asked no more, and the card goes
// on saying the wait is long and the task can be asked for in the chat.
const unansweredMost = 20;

/**
 * waitAfter is the wait once event has happened. A wait that is settled — its
 * task on the card, no task coming, or the card no longer asking — moves no
 * more.
 */
export function waitAfter(wait: Wait, event: WaitEvent): Wait {
	if (isSettled(wait)) {
		return wait;
	}
	if (event.type === "slowed") {
		return event.news === wait.news ? { ...wait, slow: true } : wait;
	}
	const asked = { ...wait, answers: wait.answers + 1, unanswered: 0 };
	const { status } = event;
	switch (status.kind) {
		case "task":
			return { ...asked, task: status.handed, notComing: 0 };
		case "over":
			return { ...asked, notComing: wait.notComing + 1 };
		case "unknown": {
			// The question that went unanswered said what heldOn says of the
			// wait before its answer: one question is asked at a time, and only
			// the clock of the news, which heldOn reads nothing of, moves the
			// wait in between.
			const unanswered = wait.unanswered + 1;
			return {
				...asked,
				unanswered,
				slow: wait.slow || unanswered >= unansweredMost,
				unheld: wait.unheld || (heldOn(wait) !== undefined && unanswered >= 2),
			};
		}
		case "writing": {
			// News is the task heard of for the first time, or heard of again
			// after a word that it was not coming, or another try turned down.
			const news =
				!wait.seenWriting ||
				wait.notComing > 0 ||
				status.refused > wait.refused;
			return {
				...asked,
				seenWriting: true,
				refused: Math.max(wait.refused, status.refused),
				notComing: 0,
				news: news ? wait.news + 1 : wait.news,
				slow: news ? false : wait.slow,
			};
		}
	}
}

/**
 * isSettled says whether there is nothing left to ask: the task is on the
 * card, enough answers in a row said none is coming, or too many questions in
 * a row went unanswered.
 */
export function isSettled(wait: Wait): boolean {
	return (
		wait.task !== undefined ||
		wait.notComing >= confirmed ||
		wait.unanswered >= unansweredMost
	);
}

/**
 * shows is what the card shows of a wait that is not settled into a task: the
 * task being written, or that none is coming — said after the first answer
 * that says so, and taken back if the next says otherwise.
 */
export function shows(wait: Wait): "writing" | "not coming" {
	return wait.notComing > 0 ? "not coming" : "writing";
}

/**
 * firstAskIn is how long the card that waits for the request given puts its
 * first question off: a moment of each request's own, under moments.spread,
 * so that the cards of a chat drawn again ask one after another rather than
 * all at once.
 */
export function firstAskIn(requestId: string): number {
	let moment = 0;
	for (const char of requestId) {
		moment = (moment * 31 + (char.codePointAt(0) ?? 0)) % moments.spread;
	}
	return moment;
}

/**
 * askIn is how long a card waits before it asks again, given how many
 * milliseconds ago the question before started, and nothing once the wait is
 * settled. It asks often while the task is being written, and seldom once the
 * wait has gone long or a question went unanswered — the service failed, or
 * held the card back —, a seldom question put off by the card's own moment as
 * well, so that cards held back together do not ask again together. The pace
 * runs from the start of each question that was answered: one the service
 * held for news is followed at once, one answered at once waits out the rest,
 * and however the service answers the card asks no more often than its pace.
 * After a question that went unanswered the pause runs from its end instead:
 * a service that fails slowly is not asked again at once, and cards held back
 * together stay apart.
 */
export function askIn(wait: Wait, own = 0, since = 0): number | undefined {
	if (isSettled(wait)) {
		return undefined;
	}
	if (wait.unanswered > 0) {
		return moments.askSlowly + own;
	}
	const pace = wait.slow ? moments.askSlowly + own : moments.ask;
	return Math.max(0, pace - Math.max(0, since));
}
