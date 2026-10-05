import type { StepStatus } from "../design/blocks";
import type { Key } from "./words";

/**
 * moments are the milliseconds a card that waits for a task keeps to: how long
 * it shows nothing before the service has answered the ask, so that an answer
 * that comes at once does not make a wait flash past; how often it asks how
 * the task stands, and how often once the wait has gone long or a question
 * went unanswered; how long the wait may go without news before the card says
 * it is taking long; how far its first question may be put off, so that the
 * cards of a chat drawn again do not all ask at once; and, once the task has
 * come, how long after ticking off its checks the card ticks the task ready,
 * and how long it holds the course done before the task takes its place. A
 * task usually takes a minute to write, and its checks seconds.
 */
export const moments = {
	shown: 400,
	ask: 4_000,
	askSlowly: 15_000,
	slow: 120_000,
	spread: 1_500,
	beat: 300,
	held: 500,
} as const;

/**
 * Phase is how far a task has got as a card can see it: its topic and
 * difficulty being picked, before the service has answered the ask; the task
 * being written; and, once the task has come, its checks passed, then the
 * task ready. The checks run inside the hand-in, out of the card's sight, so
 * the card claims them only once the task is there: a task reaches the card
 * only after every check has passed.
 */
export type Phase = "choosing" | "writing" | "checked" | "ready";

// The usual course of a task, in the order it takes them.
const course: readonly Key[] = [
	"waiting.step.topic",
	"waiting.step.writing",
	"waiting.step.answers",
	"waiting.step.ready",
];

// taken is how many steps of the course each phase has behind it.
const taken: Record<Phase, number> = {
	choosing: 0,
	writing: 1,
	checked: 3,
	ready: 4,
};

/** CourseStep is a step of the usual course of a task, and how it stands. */
export type CourseStep = { key: Key; status: StepStatus };

/**
 * stepsOf is the usual course of a task at the phase given: the steps behind
 * it done, and the rest to come. While the task is still awaited the first of
 * the rest is under way; once it has come nothing is, since what is left is
 * already over and only waits to be ticked off.
 */
export function stepsOf(phase: Phase): CourseStep[] {
	const done = taken[phase];
	const awaited = phase === "choosing" || phase === "writing";
	return course.map((key, at) => ({
		key,
		status: statusOf(at, done, awaited),
	}));
}

// statusOf is how the step at stands with done steps behind the course, and
// the next one under way while the task is awaited.
function statusOf(at: number, done: number, awaited: boolean): StepStatus {
	if (at < done) {
		return "done";
	}
	return at === done && awaited ? "active" : "waiting";
}
