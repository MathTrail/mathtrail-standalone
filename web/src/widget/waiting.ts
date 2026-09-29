import type { StepStatus } from "../design/blocks";
import type { Key } from "./words";

/**
 * Wait is where a card waiting for the next task stands: which round of
 * asking it is in — a card turned to the wait by the child's own ask starts
 * at the first, one drawn while the model is already at work at none — and
 * whether that round's ask was lost on its way to the chat.
 */
export type Wait = { round: number; lost: boolean };

/** waitStart is a card as it starts to wait, with nothing asked from it yet. */
export const waitStart: Wait = { round: 0, lost: false };

/** WaitEvent is something that happens to a card's asking for the next task. */
export type WaitEvent =
	| { type: "asked"; round: number }
	| { type: "lost"; round: number };

/**
 * waitAfter is the wait after event: a new ask starts its round afresh, and
 * the chat's refusal of one makes its round late at once. The refusal of an
 * ask made before the one the card waits on now changes nothing.
 */
export function waitAfter(wait: Wait, event: WaitEvent): Wait {
	switch (event.type) {
		case "asked":
			return { round: event.round, lost: false };
		case "lost":
			return event.round === wait.round && !wait.lost
				? { ...wait, lost: true }
				: wait;
	}
}

/**
 * moments are the seconds of a round at which what the card shows changes:
 * each step after the first taken up, the warm-up offered, and the round
 * given up on. The card cannot see the work, so the steps are the usual course
 * of one, and move by the clock; the median task took 69 seconds to write.
 */
export const moments = {
	writing: 5,
	warmUp: 30,
	answers: 35,
	readability: 50,
	deadline: 120,
} as const;

// The usual course of a task, in the order it takes them, each with the second
// of a round it is taken up at. The last, the task ready, is never taken up: a
// card cannot see a task arrive, and the new one says so by appearing.
const course: readonly { key: Key; takenUpAt: number }[] = [
	{ key: "waiting.step.topic", takenUpAt: 0 },
	{ key: "waiting.step.writing", takenUpAt: moments.writing },
	{ key: "waiting.step.answers", takenUpAt: moments.answers },
	{ key: "waiting.step.readability", takenUpAt: moments.readability },
	{ key: "waiting.step.ready", takenUpAt: Number.POSITIVE_INFINITY },
];

/** CourseStep is a step of the usual course of a task, and how it stands. */
export type CourseStep = { key: Key; status: StepStatus };

/**
 * stepsAt is the usual course of a task once a round has lasted seconds: the
 * step taken up last is under way, those before it done, the rest to come.
 */
export function stepsAt(seconds: number): CourseStep[] {
	const underWay =
		course.filter((step) => seconds >= step.takenUpAt).length - 1;
	return course.map(({ key }, at) => ({ key, status: statusOf(at, underWay) }));
}

// statusOf is how the step at stands while the one at underWay is under way.
function statusOf(at: number, underWay: number): StepStatus {
	if (at < underWay) {
		return "done";
	}
	return at === underWay ? "active" : "waiting";
}

/**
 * isLate says whether a round is given up on: its ask never reached the chat,
 * or no new card came in the time a task takes and well over.
 */
export function isLate(wait: Wait, seconds: number): boolean {
	return wait.lost || seconds >= moments.deadline;
}

/**
 * warmUps are what the child may do while the task is written: short things to
 * think about or do, with nothing to answer and nothing to check.
 */
export const warmUps = [
	"waiting.warmup.1",
	"waiting.warmup.2",
	"waiting.warmup.3",
	"waiting.warmup.4",
	"waiting.warmup.5",
	"waiting.warmup.6",
	"waiting.warmup.7",
	"waiting.warmup.8",
	"waiting.warmup.9",
] as const satisfies readonly Key[];

/** warmUpAt is the warm-up shown after turns turns, the list going round. */
export function warmUpAt(turns: number): Key {
	return warmUps[turns % warmUps.length] ?? warmUps[0];
}
