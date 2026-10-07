import type { CallToolResult } from "@modelcontextprotocol/client";
import * as z from "zod";
import { dontKnow, letters } from "./choices";

const letter = z.enum(letters);

const child = z.object({
	pseudonym: z.string(),
	grade: z.number().int(),
	ui_language: z.string().nullable(),
});

/**
 * Child is whose card it is: the name the child goes by, the grade, and the
 * language the parent chose for the lessons, or null when they chose none.
 */
export type Child = z.infer<typeof child>;

// site is the site the topics' pages are on: its address and the languages it
// is written in, each page in every one. A progress from before the links has
// none, and one that does not read is read as none: the card links nothing.
const site = z
	.object({ url: z.string(), languages: z.array(z.string()) })
	.optional()
	.catch(undefined);

/**
 * Site is the site the topics' pages are on, as a progress names it: the
 * address it gives, which a card checks before it links anything there, and
 * the languages it is written in.
 */
export type Site = NonNullable<z.infer<typeof site>>;

// topicChoice is what the card of a task offers to keep the lessons to a
// topic: the topic chosen, or null while the rule chooses; the topics the
// review suggests; and the site whose page of topics the groups link to. A
// task without it — in the trial series, of an earlier release, or on a page
// of the site — offers no choice, and one that does not read is read so too.
const topicChoice = z
	.object({
		chosen: z.string().nullable(),
		recommended: z.array(z.string()),
		site,
	})
	.optional()
	.catch(undefined);

/**
 * TopicChoice is what the card of a task offers to keep the lessons to a
 * topic: the topic chosen, or null while the rule chooses; the topics the
 * review suggests; and the site whose page of topics the groups link to.
 */
export type TopicChoice = NonNullable<z.infer<typeof topicChoice>>;

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
	// The lesson's language, which the card's words are in; a card from before
	// it travelled with the task has none.
	language: z.string().optional(),
	topic_choice: topicChoice,
});

/**
 * HandedTask is a task as it is handed to the child's card: what the child
 * may see of it, whose card it is, and what it offers to keep the lessons to.
 * It holds nothing that gives the answer away.
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
	// A card sends a letter alone, but the service tells an answer recorded
	// before as it was recorded: "I don't know" said in the chat comes back to
	// the option pressed after it.
	choice: z.union([letter, z.literal(dontKnow)]),
	correct: z.boolean(),
	correct_answer: letter,
	trap: z
		.object({
			id: z.string(),
			text: z.string(),
			// A result from before mistakes were marked as repeating says nothing
			// of it, and is read as one that does not.
			repeated: z.boolean().default(false),
		})
		.nullable(),
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

const coming = z.object({
	screen: z.literal("coming"),
	request_id: z.string().min(1),
	child,
	// The lesson's language, which the task is written in and the card's
	// words are in while it waits.
	language: z.string().optional(),
});

/**
 * Coming is a task on its way, as the card that waits for it is handed it:
 * the request whose task it asks after, and whose card it is.
 */
export type Coming = { requestId: string; child: Child };

/**
 * TaskStatus is how the task a card waits for stands, as the service last
 * said: still being written, with how many tries the checks have turned down;
 * on the card, as the card shows it; or not coming. A question that brought no
 * answer worth reading — the service failed, the call was lost or held back,
 * or what came does not read — leaves it unknown.
 */
export type TaskStatus =
	| { kind: "writing"; refused: number }
	| { kind: "task"; handed: HandedTask }
	| { kind: "over" }
	| { kind: "unknown" };

const writing = z.object({
	screen: z.literal("coming"),
	refused: z.number().int().nonnegative().default(0),
});

/**
 * readTaskStatus is how the task a card waits for stands, read from what the
 * service answered the card's question with. A profile gone since the card was
 * drawn has no task coming either.
 */
export function readTaskStatus(result: CallToolResult): TaskStatus {
	if (result.isError === true) {
		return { kind: "unknown" };
	}
	const payload = result.structuredContent;
	switch (named.safeParse(payload).data?.screen) {
		case "coming": {
			const read = writing.safeParse(payload);
			return read.success
				? { kind: "writing", refused: read.data.refused }
				: { kind: "unknown" };
		}
		case "task": {
			const handed = readHandedTask(payload);
			return handed === undefined
				? { kind: "unknown" }
				: { kind: "task", handed };
		}
		case "waiting":
		case "first_run":
			return { kind: "over" };
		default:
			return { kind: "unknown" };
	}
}

const waiting = z.object({
	screen: z.literal("waiting"),
	status: z.string().optional(),
	code: z.string().optional(),
	child: child.nullish(),
	language: z.string().optional(),
});

/**
 * Waiting is a card a task did not come to, as a tool's payload draws it: the
 * model's attempt failed its checks and the next one is to come, the request
 * it was for is over or was never opened, its last attempt failed too, or the
 * day has no room for another. A card a task was asked for knows whose it is;
 * one refused for the day by an earlier service may not.
 */
export type Waiting =
	| { kind: "refused" | "stale" | "exhausted"; child: Child }
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
	if (code === "attempts_exhausted") {
		return { kind: "exhausted", child: whose };
	}
	// A task asked for with arguments no request could be opened from had no
	// try to fail, and a task written ahead and kept comes on the card next_task
	// draws when the child asks for it: no task comes to this card, as to one
	// whose request is over.
	if (
		status === "stale" ||
		code === "invalid_arguments" ||
		code === "task_kept"
	) {
		return { kind: "stale", child: whose };
	}
	return { kind: "refused", child: whose };
}

const details = child.extend({
	interests: z.array(z.string()),
	excluded_skills: z.array(z.string()),
	// A profile from a service from before the country was kept names none.
	country: z.string().nullable().catch(null),
	region: z.string().nullable().catch(null),
	// One from before the country of the sign-in could be left out counts it.
	signin_country_off: z.boolean().catch(false),
});

/**
 * Details are the child's profile as a card shows it: who the child is, what
 * the tasks may be dressed in, what the child has not met at school yet, the
 * language of the lessons, the country and the region the family lives in, by
 * their codes, or null for none given, and whether the country the parent
 * signs in from is left out of what is counted. The parent's notes are never
 * among them.
 */
export type Details = z.infer<typeof details>;

/**
 * Problem is a field a change from the form broke a rule of: the field, as
 * the service names it, and the code of the rule, which the card says in its
 * own words. A code the card has no words for is said in words that fit any.
 */
export type Problem = { field: string; code: string };

const edited = z.object({
	status: z.string().optional(),
	code: z.string().optional(),
	problems: z
		.array(z.object({ field: z.string(), code: z.string().catch("") }))
		.default([]),
	changed: z.boolean().default(false),
	profile: details.nullish(),
});

/**
 * EditOutcome is how a change sent from the form ended: saved, with the details
 * as they now stand and, when something changed, the words for the model;
 * refused, field by field; refused because the profile is not there any
 * more; or not saved at all, and worth sending again.
 */
export type EditOutcome =
	| { kind: "saved"; details: Details; told: string | undefined }
	| { kind: "refused"; problems: readonly Problem[] }
	| { kind: "gone" }
	| { kind: "failed" };

/**
 * readEdited is the outcome of a change sent from the form, read from the
 * result the service returned. A failure of the service and anything that does
 * not read are a change not saved.
 */
export function readEdited(result: CallToolResult): EditOutcome {
	if (result.isError === true) {
		return { kind: "failed" };
	}
	const read = edited.safeParse(result.structuredContent);
	if (!read.success) {
		return { kind: "failed" };
	}
	const { status, code, problems, changed, profile } = read.data;
	if (status === "stale" && code === "stale_profile") {
		return { kind: "gone" };
	}
	if (status === "rejected" && problems.length > 0) {
		return { kind: "refused", problems };
	}
	if (status !== undefined || profile === null || profile === undefined) {
		return { kind: "failed" };
	}
	return {
		kind: "saved",
		details: profile,
		told: changed ? wordsOf(result) : undefined,
	};
}

// wordsOf is the text a result gives the model, or undefined when it gives
// none.
function wordsOf(result: CallToolResult): string | undefined {
	for (const block of result.content) {
		if (block.type === "text" && block.text !== "") {
			return block.text;
		}
	}
	return undefined;
}

const trial = z.object({ answered: z.number().int(), of: z.number().int() });

const recommendation = z.object({
	topic: z.string(),
	goal: z.string(),
	// A progress from before the lessons could be kept to a topic says nothing
	// of it, and one that does not read is read as a topic the rule chose.
	chosen: z.boolean().optional().catch(undefined),
});

/**
 * Recommendation is what the rule would set next: a topic, whether it is
 * worked over again after a mistake or is new ground, and whether it is the
 * topic someone chose to keep the lessons to.
 */
export type Recommendation = z.infer<typeof recommendation>;

const location = z.object({
	folder: z.string(),
	file: z.string(),
	others: z.array(z.object({ file: z.string() })),
});

/**
 * Location is where the profile's file is, as a card names it: the folder and
 * the file, and the other files that hold a profile too. A card opens none of
 * them.
 */
export type Location = z.infer<typeof location>;

// share is how far through a rank a rating has come, as a whole percent.
const share = z.number().int().min(0).max(100);

// move is how a rank moved over a while: where it stood before, as it was
// drawn, and the word for the move. The word is read as any text, so that one
// a later release adds draws no move rather than no card; the rating before,
// which the card does not draw, is left unread.
const move = z.object({
	rank: z.number().int().positive().optional(),
	share: share.optional(),
	moved: z.string(),
});

// moves are how a rank moved since the last task and over the week, each null
// where the while cannot be told. A while that does not read is read as one
// that cannot be told, and moves that do not read as none at all: a move is a
// part of the progress a card can do without, never a reason to show none.
const moves = z
	.object({
		last_task: z.nullable(move).catch(null),
		week: z.nullable(move).catch(null),
	})
	.optional()
	.catch(undefined);

/**
 * Move is how a rank moved over a while, as a card reads it: where it stood
 * before, and the word for the move.
 */
export type Move = z.infer<typeof move>;

/**
 * Moves are how a rank moved since the last task and over the week, each null
 * where the while cannot be told.
 */
export type Moves = NonNullable<z.infer<typeof moves>>;

// judged is a topic the review names: the topic, why it is named, by codes,
// the trap its answers keep falling for, and whether a topic to develop has
// begun to move. A code is read as any text, so that one a later release adds
// leaves the card no reason it can say rather than no review.
const judged = z.object({
	topic: z.string(),
	reasons: z.array(z.string()),
	trap: z.string().optional(),
	moving: z.boolean().optional(),
});

// step is a step the review advises: what it advises, by a code, the topic it
// is for — none for the step that holds for every topic —, the trap whose
// advice it is, and the strong topic a topic to begin builds on.
const step = z.object({
	kind: z.string(),
	topic: z.string().optional(),
	trap: z.string().optional(),
	base: z.string().optional(),
});

// review is the review of the topics for the adult. A progress of the trial
// series has none, and neither has one from before the review; one that does
// not read is read as none, and the rest of the progress drawn all the same.
const review = z
	.object({
		strong: z.array(judged),
		develop: z.array(judged),
		early: z.array(z.string()),
		steps: z.array(step),
	})
	.optional()
	.catch(undefined);

/**
 * Judged is a topic the review names: the topic, why, by codes, the trap its
 * answers keep falling for, and whether it has begun to move.
 */
export type Judged = z.infer<typeof judged>;

/**
 * ReviewStep is a step the review advises: its kind, the topic it is for, if
 * it is for one, the trap whose advice it is, and the strong topic a topic to
 * begin builds on.
 */
export type ReviewStep = z.infer<typeof step>;

/**
 * Review is the review of the topics for the adult: the strong topics, the
 * ones to develop, the ones too early to judge, and the steps to take.
 */
export type Review = NonNullable<z.infer<typeof review>>;

const progressReport = z.object({
	screen: z.literal("progress"),
	profile: details,
	trial: trial.nullable(),
	// The overall rating's number is the service's to send and the model's to
	// say: the card draws the rank alone, and needs nothing it does not draw.
	overall: z
		.object({
			rank: z.number().int(),
			ranks: z.number().int(),
			// A progress from before the share has none: its course is drawn
			// with the step under way empty, as an earlier chat's card still
			// is when it is drawn again.
			share: share.optional(),
			// The grade levels, each with the ranks its grades are matched
			// with, which the card marks under the course for the parent. A
			// progress from before the marks has none, and marks that do not
			// read are read as none: the course is drawn unmarked.
			grades: z
				.array(
					z.object({
						grade_level: z.string().regex(/^\d+-\d+$/),
						first_rank: z.number().int().positive(),
						last_rank: z.number().int().positive(),
					}),
				)
				.optional()
				.catch(undefined),
			// A progress from before the moves, or one with nothing to tell of
			// them, has none, and the card draws no move.
			change: moves,
		})
		.nullable(),
	topics: z.array(
		z.object({
			topic: z.string(),
			rating: z.number().int().nullable(),
			// A progress from before the topics had ranks has none of these,
			// and its topics are drawn with an empty course and no word of
			// where they stand. How a rank compares is read as any text, so
			// that one a later release adds draws no word rather than no card.
			rank: z.number().int().positive().nullish(),
			share: share.nullish(),
			compared: z.string().nullish(),
			answers: z.number().int().nonnegative(),
			mastered: z.boolean(),
			skipped: z.number().int(),
			change: moves,
			// A progress from before the topics' pages has neither, and one
			// that does not read is read as none: the topic is drawn with no
			// link to its page.
			slug: z.string().optional().catch(undefined),
			site_page: z.boolean().optional().catch(undefined),
		}),
	),
	recent: z.array(
		z.object({
			topic: z.string(),
			correct: z.boolean().nullable(),
			skipped: z.boolean(),
		}),
	),
	// A progress from before the total of the skips has none, and the card
	// adds its topics' counts up as it did then.
	skipped: z.number().int().nonnegative().optional(),
	// A progress from before the map of mistakes has none, and is read as one
	// where nothing repeats: a card of an earlier chat, drawn again, still shows.
	mistakes: z
		.array(z.object({ trap: z.string(), times: z.number().int().positive() }))
		.default([]),
	review,
	recommendation: recommendation.nullable(),
	// Kept nowhere a person could open, or from before the progress said where
	// the file is, a progress has no location, and the card no parent's data.
	location: location.optional(),
	site,
});

/**
 * ProgressReport is where the child stands, as the progress screen shows it:
 * the overall rating with its rank, how far through it, and the grades the
 * ranks are matched with — or, while the trial series runs, how far the
 * series has got — the topics met or within reach, each with a rank of its
 * own, the latest answers and how many tasks were left without one, the
 * mistakes that keep coming back, the review of the topics once the trial
 * series is over, what comes next, the child's profile, and where its file is.
 */
export type ProgressReport = z.infer<typeof progressReport>;

const profileCard = z.object({
	screen: z.literal("profile"),
	status: z.string().optional(),
	profile: details,
	location: location.optional(),
});

/**
 * ProfileReport is the child's profile as its own card shows it to the parent:
 * the details; where their file is, when the service keeps it somewhere a
 * person can open, and the other files that hold a profile too; and whether
 * the change just asked for was refused, the profile shown as it stays.
 */
export type ProfileReport = {
	details: Details;
	location: Location | undefined;
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
	| { screen: "coming"; coming: Coming }
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
		case "coming": {
			const read = coming.safeParse(payload);
			return read.success
				? {
						screen: "coming",
						coming: { requestId: read.data.request_id, child: read.data.child },
					}
				: undefined;
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
