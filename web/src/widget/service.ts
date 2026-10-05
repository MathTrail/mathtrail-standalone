import { createContext } from "preact";
import { useContext } from "preact/hooks";
import type { Choice } from "./choices";
import type { AnswerOutcome, EditOutcome, TaskStatus } from "./payload";

/**
 * ProgressRead is how a reading of the child's progress ended: the service's
 * reply, which whoever draws it reads, or none worth reading.
 */
export type ProgressRead =
	| { kind: "read"; payload: unknown }
	| { kind: "failed" };

/**
 * Service is what a card asks of MathTrail's service: to record the child's
 * answer to a task, to say how the task a card waits for stands, to read the
 * child's progress, and to save a change to the child's profile. In a chat
 * each is a call of one of the service's tools through the host, its reply
 * read before the card sees it; a page that shows a card live answers for the
 * service itself, and calls nobody.
 */
export type Service = {
	/**
	 * recordAnswer records choice as the answer to the task taskId, with
	 * whether the hint was open, and says how it ended.
	 */
	recordAnswer(
		taskId: string,
		choice: Choice,
		hintUsed: boolean,
	): Promise<AnswerOutcome>;
	/** taskStatus says how the task of the request requestId stands. */
	taskStatus(requestId: string): Promise<TaskStatus>;
	/** readProgress reads the child's progress as it stands. */
	readProgress(): Promise<ProgressRead>;
	/**
	 * saveEdit saves changes to the child's profile, named as edit_profile
	 * takes them, and says how it ended.
	 */
	saveEdit(changes: Record<string, unknown>): Promise<EditOutcome>;
};

// unreached is the service of a card drawn where nothing answers for the
// service: every question ends as one that never reached it.
const unreached: Service = {
	recordAnswer: () => Promise.resolve({ kind: "failed" }),
	taskStatus: () => Promise.resolve({ kind: "unknown" }),
	readProgress: () => Promise.resolve({ kind: "failed" }),
	saveEdit: () => Promise.resolve({ kind: "failed" }),
};

/**
 * ServiceContext hands the service to every card drawn inside it: the one a
 * chat's host reaches, or the one a page answers for. A card drawn where
 * nothing answers for the service — still, on a page that only shows it —
 * asks one that every question fails to reach.
 */
export const ServiceContext = createContext<Service>(unreached);

/** useService is the service the card a component is drawn in asks. */
export function useService(): Service {
	return useContext(ServiceContext);
}
