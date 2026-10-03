import type { StepStatus } from "../design/blocks";
import type { Key } from "./words";

/**
 * moments are the milliseconds a card that waits for a task keeps to: how long
 * it shows nothing before the service has answered the ask, so that an answer
 * that comes at once does not make a wait flash past; how often it asks how
 * the task stands, and how often once the wait has gone long or a question
 * went unanswered; how long the wait may go without news before the card says
 * it is taking long; and how far its first question may be put off, so that
 * the cards of a chat drawn again do not all ask at once. A task usually takes
 * a minute to write, and its checks seconds.
 */
export const moments = {
	shown: 400,
	ask: 4_000,
	askSlowly: 15_000,
	slow: 120_000,
	spread: 1_500,
} as const;

/**
 * Phase is how far a task has got as a card can see it: its topic and
 * difficulty being picked, before the service has answered the ask, or the
 * task being written.
 */
export type Phase = "choosing" | "writing";

// The usual course of a task, in the order it takes them. The checks run
// inside the hand-in, out of the card's sight, and the last step, the task
// ready, is never taken up: a task that is ready takes the course's place.
const course: readonly Key[] = [
	"waiting.step.topic",
	"waiting.step.writing",
	"waiting.step.answers",
	"waiting.step.ready",
];

/** CourseStep is a step of the usual course of a task, and how it stands. */
export type CourseStep = { key: Key; status: StepStatus };

/**
 * stepsOf is the usual course of a task as far as a card can see it: the topic
 * and difficulty being picked, then picked, with the task being written.
 */
export function stepsOf(phase: Phase): CourseStep[] {
	const underWay = phase === "choosing" ? 0 : 1;
	return course.map((key, at) => ({ key, status: statusOf(at, underWay) }));
}

// statusOf is how the step at stands while the one at underWay is under way.
function statusOf(at: number, underWay: number): StepStatus {
	if (at < underWay) {
		return "done";
	}
	return at === underWay ? "active" : "waiting";
}
