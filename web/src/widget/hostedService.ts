import type { Host } from "./bridge";
import type { Letter } from "./choices";
import {
	type AnswerOutcome,
	type EditOutcome,
	readAnswer,
	readEdited,
	readTaskStatus,
	type TaskStatus,
} from "./payload";
import type { ProgressRead, Service } from "./service";

/**
 * serviceThrough is the service as a card in a chat reaches it: each question
 * a call of one of the service's tools through host, and its reply read as
 * the card reads what the service says. A call that never reached the
 * service, or whose reply never came back, ends as one the service never
 * answered — an answer not recorded, a task's standing unknown, no progress
 * read, a change not saved — and is worth asking again.
 */
export function serviceThrough(host: Host): Service {
	return {
		recordAnswer: (taskId, choice, hintUsed) =>
			recordAnswer(host, taskId, choice, hintUsed),
		taskStatus: (requestId, refused) => taskStatus(host, requestId, refused),
		readProgress: () => readProgress(host),
		saveEdit: (changes) => saveEdit(host, changes),
	};
}

// recordAnswer sends the child's answer to the service and reads how it went.
// An answer whose reply never came is one the card cannot call recorded; sent
// again, it is either recorded then or told as the service recorded it.
async function recordAnswer(
	host: Host,
	taskId: string,
	choice: Letter,
	hintUsed: boolean,
): Promise<AnswerOutcome> {
	try {
		const result = await host.callTool("submit_answer", {
			task_id: taskId,
			answer: choice,
			hint_used: hintUsed,
		});
		return readAnswer(result, taskId);
	} catch (error: unknown) {
		console.error("widget: the answer did not reach the service", error);
		return { kind: "failed" };
	}
}

// taskStatus asks the service how the task of the request stands, saying the
// tries turned down the card has heard of when it has heard the task is being
// written. A question whose answer never came leaves it unknown, and is asked
// again later.
async function taskStatus(
	host: Host,
	requestId: string,
	refused: number | undefined,
): Promise<TaskStatus> {
	try {
		return readTaskStatus(
			await host.callTool(
				"read_task",
				refused === undefined
					? { request_id: requestId }
					: { request_id: requestId, refused },
			),
		);
	} catch (error: unknown) {
		console.error("widget: how the task stands did not arrive", error);
		return { kind: "unknown" };
	}
}

// readProgress asks the service for the child's progress. A reply that says
// the service failed, and a question whose answer never came, read as none.
async function readProgress(host: Host): Promise<ProgressRead> {
	try {
		const read = await host.callTool("read_progress", {});
		return read.isError === true
			? { kind: "failed" }
			: { kind: "read", payload: read.structuredContent };
	} catch (error: unknown) {
		console.error("widget: the progress did not arrive", error);
		return { kind: "failed" };
	}
}

// saveEdit sends changes to the child's profile to the service and reads how
// it ended. A call that never reached the service, or whose answer never came
// back, is a change not saved.
async function saveEdit(
	host: Host,
	changes: Record<string, unknown>,
): Promise<EditOutcome> {
	try {
		return readEdited(await host.callTool("edit_profile", changes));
	} catch (error: unknown) {
		console.error("widget: the change did not reach the service", error);
		return { kind: "failed" };
	}
}
