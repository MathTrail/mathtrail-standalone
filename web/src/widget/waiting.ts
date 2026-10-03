import type { StepStatus } from "../design/blocks";
import type { CallStage } from "./bridge";
import type { Key } from "./words";

/**
 * moments are the milliseconds a card of a task being handed in waits on its
 * own: before it shows the course at all, so that a result that comes with the
 * call does not make a wait flash past; and before it gives up, counted afresh
 * once the task has been handed in whole, so that a slow hand-in is never
 * taken for a task that is not coming. Most of a task's writing is done before
 * its hand-in starts, and its checks take seconds.
 */
export const moments = { shown: 400, givenUp: 120_000 } as const;

// The usual course of a task, in the order it takes them. The last, the task
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
 * stepsOf is the usual course of a task as far as its hand-in has got: the
 * topic chosen — no task is handed in before one is — the task being written
 * while its arguments come, and checked once they have come whole.
 */
export function stepsOf(stage: Exclude<CallStage, "cancelled">): CourseStep[] {
	const underWay = stage === "started" ? 1 : 2;
	return course.map((key, at) => ({ key, status: statusOf(at, underWay) }));
}

// statusOf is how the step at stands while the one at underWay is under way.
function statusOf(at: number, underWay: number): StepStatus {
	if (at < underWay) {
		return "done";
	}
	return at === underWay ? "active" : "waiting";
}
