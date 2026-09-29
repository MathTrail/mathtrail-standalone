import type { CallToolResult } from "@modelcontextprotocol/client";
import * as z from "zod";

const letter = z.enum(["A", "B", "C", "D", "E"]);

/** Letter names one of a task's five options. */
export type Letter = z.infer<typeof letter>;

/** letters are the options' letters, in the order the card shows them. */
export const letters: readonly Letter[] = letter.options;

/** dontKnow is the answer "I don't know": a wrong answer that chose no option. */
export const dontKnow = "?";

/** Choice is an answer the child can give: an option, or "I don't know". */
export type Choice = Letter | typeof dontKnow;

const child = z.object({
	pseudonym: z.string(),
	grade: z.number().int(),
	ui_language: z.string().nullable(),
});

/**
 * Child is whose card it is: the name the child goes by, the grade, and the
 * language the parent chose for the cards, or null when they chose none.
 */
export type Child = z.infer<typeof child>;

const handedTask = z.object({
	screen: z.literal("task"),
	child,
	task: z.object({
		id: z.string(),
		topic: z.string(),
		language: z.string(),
		question: z.string(),
		drawing: z.string(),
		options: z.object({
			A: z.string(),
			B: z.string(),
			C: z.string(),
			D: z.string(),
			E: z.string(),
		}),
		hint: z.string(),
	}),
});

/**
 * HandedTask is a task as it is handed to the child's card: what the child
 * may see of it, and whose card it is. It holds nothing that gives the answer
 * away.
 */
export type HandedTask = z.infer<typeof handedTask>;

/**
 * readHandedTask is the task a tool's payload hands the card, or undefined
 * when the payload draws another screen or does not read as a task. A task
 * handed out again — its request already answered with it — is the task the
 * child holds, and reads like any other.
 */
export function readHandedTask(payload: unknown): HandedTask | undefined {
	const read = handedTask.safeParse(payload);
	return read.success ? read.data : undefined;
}

const answerResult = z.object({
	task_id: z.string(),
	topic: z.string(),
	choice: z.union([letter, z.literal(dontKnow)]),
	correct: z.boolean(),
	correct_answer: letter,
	trap: z.object({ id: z.string(), text: z.string() }).nullable(),
	solution: z.string(),
	hint_used: z.boolean(),
	rating: z.object({ before: z.number(), after: z.number() }).nullable(),
	trial: z
		.object({ answered: z.number().int(), of: z.number().int() })
		.nullable(),
	already_answered: z.boolean(),
});

/**
 * AnswerResult is what the service says of a recorded answer: the choice and
 * whether it was right, the right option, the trap behind a wrong option, the
 * solution, and the rating in the topic before and after — or, while the
 * trial series runs, how far it has got.
 */
export type AnswerResult = z.infer<typeof answerResult>;

const recorded = z.object({
	screen: z.literal("result"),
	result: answerResult,
});
const stale = z.object({ status: z.literal("stale") });

/**
 * AnswerOutcome is how an answer sent from the card ended: recorded, with its
 * result; refused because the task is no longer the one being solved; or not
 * recorded at all, and worth sending again.
 */
export type AnswerOutcome =
	| { kind: "answered"; result: AnswerResult }
	| { kind: "closed" }
	| { kind: "failed" };

/**
 * readAnswer is the outcome of an answer to the task taskId, read from the
 * result the service returned. A result for another task, a refusal of the
 * answer itself, a failure of the service and anything that does not read are
 * all an answer not recorded.
 */
export function readAnswer(
	result: CallToolResult,
	taskId: string,
): AnswerOutcome {
	if (result.isError === true) {
		return { kind: "failed" };
	}
	const answered = recorded.safeParse(result.structuredContent);
	if (answered.success && answered.data.result.task_id === taskId) {
		return { kind: "answered", result: answered.data.result };
	}
	if (stale.safeParse(result.structuredContent).success) {
		return { kind: "closed" };
	}
	return { kind: "failed" };
}

const waiting = z.object({
	screen: z.literal("waiting"),
	status: z.string().optional(),
	code: z.string().optional(),
	child: child.nullish(),
});

/**
 * Waiting is a card that waits for the next task, as a tool's payload draws
 * it: while the model writes the task, after the model's last attempt at one
 * failed its checks, or once the day has no room for another. A card that
 * waits for a task knows whose it is; one refused for the day may not.
 */
export type Waiting =
	| { kind: "working" | "exhausted"; child: Child }
	| { kind: "limit"; child: Child | undefined };

/**
 * readWaiting is the wait a tool's payload draws, or undefined when the
 * payload draws another screen, does not read, or has a card wait for a task
 * without saying whose card it is.
 */
export function readWaiting(payload: unknown): Waiting | undefined {
	const read = waiting.safeParse(payload);
	if (!read.success) {
		return undefined;
	}
	const { status, code } = read.data;
	const whose = read.data.child ?? undefined;
	if (status === "limited") {
		return { kind: "limit", child: whose };
	}
	if (whose === undefined) {
		return undefined;
	}
	return {
		kind: code === "attempts_exhausted" ? "exhausted" : "working",
		child: whose,
	};
}
