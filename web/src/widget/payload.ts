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

const details = child.extend({
	interests: z.array(z.string()),
	excluded_skills: z.array(z.string()),
});

/**
 * Details are the child's profile as a card shows it: who the child is, what
 * the tasks may be dressed in, what the child has not met at school yet, and
 * the language of the cards. The parent's notes are never among them.
 */
export type Details = z.infer<typeof details>;

const trial = z.object({ answered: z.number().int(), of: z.number().int() });

const recommendation = z.object({
	topic: z.string(),
	goal: z.string(),
});

/**
 * Recommendation is what the rule would set next: a topic, and whether it is
 * worked over again after a mistake or is new ground.
 */
export type Recommendation = z.infer<typeof recommendation>;

const progressReport = z.object({
	screen: z.literal("progress"),
	profile: details,
	trial: trial.nullable(),
	overall: z
		.object({
			rating: z.number().int(),
			rank: z.number().int(),
			ranks: z.number().int(),
		})
		.nullable(),
	topics: z.array(
		z.object({
			topic: z.string(),
			rating: z.number().int().nullable(),
			mastered: z.boolean(),
			skipped: z.number().int(),
		}),
	),
	recent: z.array(
		z.object({
			topic: z.string(),
			correct: z.boolean().nullable(),
			skipped: z.boolean(),
		}),
	),
	recommendation: recommendation.nullable(),
});

/**
 * ProgressReport is where the child stands, as the progress screen shows it:
 * the overall rating with its rank — or, while the trial series runs, how far
 * the series has got — the topics met, the latest answers and the tasks left
 * without one, what comes next, and the child's profile.
 */
export type ProgressReport = z.infer<typeof progressReport>;

const profileCard = z.object({
	screen: z.literal("profile"),
	status: z.string().optional(),
	profile: details,
	location: z
		.object({
			folder: z.string(),
			file: z.string(),
			others: z.array(z.object({ file: z.string() })),
		})
		.optional(),
});

/**
 * ProfileReport is the child's profile as its own card shows it to the parent:
 * the details; where their file is, when the service keeps it somewhere a
 * person can open, and the other files that hold a profile too; and whether
 * the change just asked for was refused, the profile shown as it stays.
 */
export type ProfileReport = {
	details: Details;
	location: z.infer<typeof profileCard>["location"];
	refused: boolean;
};

const firstRun = z.object({
	screen: z.literal("first_run"),
	status: z.string().optional(),
});

/**
 * FirstRun is the card of an account with no profile yet, and whether the
 * profile just asked for was refused.
 */
export type FirstRun = { refused: boolean };

/**
 * Screen is what a payload draws, read: each screen a card can show, with
 * what it shows. The payload names its screen; the widget never guesses one
 * from the shape of the data.
 */
export type Screen =
	| { screen: "task"; handed: HandedTask }
	| { screen: "waiting"; waiting: Waiting }
	| { screen: "progress"; report: ProgressReport }
	| { screen: "profile"; profile: ProfileReport }
	| { screen: "first_run"; firstRun: FirstRun };

const named = z.object({ screen: z.string() });

/**
 * readScreen is the screen a tool's payload draws, read as the payload names
 * it, or undefined when it names none a card draws, or does not read as the
 * one it names.
 */
export function readScreen(payload: unknown): Screen | undefined {
	switch (named.safeParse(payload).data?.screen) {
		case "task": {
			const handed = readHandedTask(payload);
			return handed === undefined ? undefined : { screen: "task", handed };
		}
		case "waiting": {
			const waiting = readWaiting(payload);
			return waiting === undefined ? undefined : { screen: "waiting", waiting };
		}
		case "progress": {
			const report = progressReport.safeParse(payload);
			return report.success
				? { screen: "progress", report: report.data }
				: undefined;
		}
		case "profile":
			return readProfile(payload);
		case "first_run": {
			const first = firstRun.safeParse(payload);
			return first.success
				? {
						screen: "first_run",
						firstRun: { refused: first.data.status === "rejected" },
					}
				: undefined;
		}
		default:
			return undefined;
	}
}

// readProfile is the profile's card a payload draws, or undefined when it
// does not read as one.
function readProfile(payload: unknown): Screen | undefined {
	const read = profileCard.safeParse(payload);
	if (!read.success) {
		return undefined;
	}
	const { profile, location, status } = read.data;
	return {
		screen: "profile",
		profile: { details: profile, location, refused: status === "rejected" },
	};
}
