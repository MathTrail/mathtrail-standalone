import type { TaskStatus } from "./payload";
import { type Service, useService } from "./service";

/**
 * questionLongest is how long a card waits for the answer to one question of
 * whether its task is still the one on the child's card, in milliseconds.
 * Past it the card goes on as it did before it asked, and asks nothing more: a
 * service slow to answer must not keep the next task back. A card whose task
 * is on the child's card asks once, so the next task waits at most this long;
 * one told its task is set aside asks twice, askAgainAfter apart, six seconds
 * at the most.
 */
export const questionLongest = 2_000;

/**
 * askAgainAfter is how long a card waits, in milliseconds, before it asks a
 * second time whether its task is still on the child's card, once the first
 * answer says it is not. A read straight after a write may bring the file as
 * it was before it: the store reads again a moment after a read that lagged
 * behind a write of its own instance, but a read on another instance has no
 * write of its own to measure by.
 */
export const askAgainAfter = 2_000;

// taskMark and requestMark begin the ids of a task and of the request it was
// written for, which are otherwise one.
const taskMark = "tsk_";
const requestMark = "req_";

/**
 * requestIdOf is the id of the request a task was written for, which the
 * service names the task after; undefined for a task whose id is of another
 * shape, whose request the card cannot name.
 */
export function requestIdOf(taskId: string): string | undefined {
	return taskId.startsWith(taskMark) && taskId.length > taskMark.length
		? requestMark + taskId.slice(taskMark.length)
		: undefined;
}

/**
 * stillOnTheCard says whether a card may still ask for another task or save a
 * topic: from a card drawn again higher up the chat, either would set aside
 * the task of a newer card that waits for the child's answer. It is false only
 * when the service says twice, askAgainAfter apart, that the card's task is no
 * longer on the child's card and another one there waits for its answer.
 * Anything else lets the card go on: its task on the card; no task waiting,
 * which an ask skips nothing of; a profile gone, which the card it asks for
 * meets in words of its own; and an answer that failed or did not come within
 * questionLongest. A task whose request the card cannot name is asked of
 * nobody: it is undefined, so that the card goes on at once, as it did before
 * it asked.
 */
export function stillOnTheCard(
	service: Service,
	taskId: string,
): Promise<boolean> | undefined {
	const requestId = requestIdOf(taskId);
	return requestId === undefined
		? undefined
		: notToldSetAside(service, requestId);
}

/**
 * useStillOn is a card's question, before it asks for another task or saves a
 * topic, whether its task is still the one on the child's card, as
 * stillOnTheCard asks it; once the answer is that it is not, onClosed closes
 * the card.
 */
export function useStillOn(
	taskId: string,
	onClosed: () => void,
): () => Promise<boolean> | undefined {
	const service = useService();
	return () =>
		stillOnTheCard(service, taskId)
			?.then((on) => {
				if (!on) {
					onClosed();
				}
				return on;
			})
			.catch(() => true);
}

// notToldSetAside is false once the service has said twice, askAgainAfter
// apart, that the task of the request has been set aside for another that
// waits for its answer, and true at its first word of anything else.
async function notToldSetAside(
	service: Service,
	requestId: string,
): Promise<boolean> {
	for (let told = 0; told < 2; told++) {
		if (told > 0) {
			await new Promise((resolve) => setTimeout(resolve, askAgainAfter));
		}
		const status = await answerTo(service, requestId);
		if (status.kind !== "over" || status.anotherWaits !== true) {
			return true;
		}
	}
	return false;
}

// unknown is how a task stands for a card that heard nothing worth reading.
const unknown: TaskStatus = { kind: "unknown" };

// answerTo is the service's answer to one question of how the request stands,
// unknown when it fails or does not come within questionLongest.
async function answerTo(
	service: Service,
	requestId: string,
): Promise<TaskStatus> {
	let timer: ReturnType<typeof setTimeout> | undefined;
	const late = new Promise<TaskStatus>((resolve) => {
		timer = setTimeout(() => resolve(unknown), questionLongest);
	});
	try {
		return await Promise.race([
			Promise.resolve()
				.then(() => service.taskStatus(requestId))
				.catch(() => unknown),
			late,
		]);
	} finally {
		clearTimeout(timer);
	}
}
