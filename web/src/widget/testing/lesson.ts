import type { CallToolResult } from "@modelcontextprotocol/client";
import type { AnswerResult, HandedTask } from "../payload";

/**
 * Handed is a payload that hands the card a task, with the fields the service
 * sends beside it and the card does not read.
 */
export type Handed = HandedTask & {
	attempt: number;
	last_answer: null;
	status?: "stale";
	code?: "stale_request";
};

/**
 * fence is the task of the approved design — posts along a fence — as the
 * service hands it to a child's card, in English.
 */
export const fence: Handed = {
	screen: "task",
	attempt: 1,
	last_answer: null,
	child: { pseudonym: "Comet", grade: 3, ui_language: null },
	task: {
		id: "task_fence",
		topic: "counting.gaps",
		language: "en",
		question:
			"A fence is 12 meters long. Posts stand every 3 meters, including both ends. How many posts are there?",
		drawing: "|--3--|--3--|--3--|--3--|\n",
		options: { A: "3", B: "4", C: "5", D: "6", E: "12" },
		hint: "Try a smaller fence first: 6 meters long, with a post every 3 meters. Draw it and count the posts.",
	},
};

/**
 * fenceInRussian is the same task in Russian: the same task, so that the
 * results of answering the fence are its results too.
 */
export const fenceInRussian: Handed = {
	...fence,
	child: { pseudonym: "Комета", grade: 3, ui_language: null },
	task: {
		...fence.task,
		language: "ru",
		question:
			"Забор длиной 12 метров. Столбы стоят через каждые 3 метра, включая оба конца. Сколько всего столбов?",
		hint: "Начни с забора поменьше: 6 метров, столб через каждые 3 метра. Нарисуй его и посчитай столбы.",
	},
};

/** fenceSolution is the solution of the fence, three steps in one text. */
export const fenceSolution =
	"12 ÷ 3 = 4 gaps. A straight fence with posts at both ends has one more post than gaps. 4 + 1 = 5 posts.";

/** fenceSolutionInRussian is the Russian one. */
export const fenceSolutionInRussian =
	"12 : 3 = 4 промежутка. У прямого забора со столбами на обоих концах столбов на один больше, чем промежутков. 4 + 1 = 5 столбов.";

/**
 * answered is the result of an answer to the fence the service recorded: by
 * default the wrong B, whose trap is counting the gaps, with the rating in the
 * topic going down. Fields given replace the default ones.
 */
export function answered(fields: Partial<AnswerResult> = {}): CallToolResult {
	const result: AnswerResult = {
		task_id: fence.task.id,
		topic: fence.task.topic,
		choice: "B",
		correct: false,
		correct_answer: "C",
		trap: {
			id: "fence_gaps",
			text: "Counted the gaps instead of the posts.",
		},
		solution: fenceSolution,
		hint_used: false,
		rating: { before: 1502, after: 1480 },
		trial: null,
		already_answered: false,
		...fields,
	};
	return toolResult({
		screen: "result",
		last_answer: {
			task_id: result.task_id,
			topic: result.topic,
			correct: result.correct,
			answered_at: "2026-09-29T12:00:00Z",
		},
		result,
	});
}

/** rightAnswer is the fence answered with C, which is right. */
export const rightAnswer = answered({
	choice: "C",
	correct: true,
	trap: null,
	rating: { before: 1502, after: 1519 },
});

/** dontKnowAnswer is the fence answered "I don't know", a wrong answer with no trap. */
export const dontKnowAnswer = answered({
	choice: "?",
	trap: null,
	rating: { before: 1502, after: 1488 },
});

/** trialAnswer is a wrong answer given as the third task of the trial series. */
export const trialAnswer = answered({
	rating: null,
	trial: { answered: 3, of: 5 },
});

/**
 * toldAgain is the fence answered before, with D, told again to a card that
 * sends another answer.
 */
export const toldAgain = answered({
	choice: "D",
	trap: { id: "fence_ends", text: "Counted one end twice." },
	already_answered: true,
});

/** staleAnswer is the refusal of an answer to a task that is no longer the one being solved. */
export const staleAnswer = toolResult({
	screen: "task",
	status: "stale",
	code: "stale_task",
	last_answer: null,
	result: null,
});

/** failure is a failure of the service: one sentence for the model, and no payload. */
export const failure: CallToolResult = {
	content: [
		{
			type: "text",
			text: "MathTrail could not reach Google Drive. Make the same call again in a minute.",
		},
	],
	isError: true,
};

/**
 * standing is where Comet stands after the trial series: the design's own
 * progress — four topics, one mastered, the latest answers with one task left
 * without an answer, and the topic worked over again after a mistake next.
 */
export const standing = {
	screen: "progress",
	last_answer: {
		task_id: "task_fence",
		topic: "combinatorics.enumeration",
		correct: false,
		answered_at: "2026-09-29T12:00:00Z",
	},
	profile: {
		pseudonym: "Comet",
		grade: 3,
		interests: ["space", "animals", "football"],
		excluded_skills: ["division_with_remainder"],
		ui_language: null,
	},
	trial: null,
	overall: { rating: 1573, rank: 3, ranks: 11 },
	topics: [
		{
			topic: "logic.ordering",
			rating: 1712,
			answers: 6,
			correct: 5,
			mastered: true,
			skipped: 0,
		},
		{
			topic: "combinatorics.enumeration",
			rating: 1627,
			answers: 4,
			correct: 2,
			mastered: false,
			skipped: 1,
		},
		{
			topic: "counting.gaps",
			rating: 1588,
			answers: 3,
			correct: 2,
			mastered: false,
			skipped: 0,
		},
		{
			topic: "parity.alternation",
			rating: 1541,
			answers: 2,
			correct: 1,
			mastered: false,
			skipped: 0,
		},
	],
	recent: [
		{
			topic: "combinatorics.enumeration",
			correct: false,
			skipped: false,
			answered_at: "2026-09-29T12:00:00Z",
		},
		{
			topic: "combinatorics.enumeration",
			correct: null,
			skipped: true,
			answered_at: "2026-09-29T11:50:00Z",
		},
		{
			topic: "counting.gaps",
			correct: true,
			skipped: false,
			answered_at: "2026-09-29T11:40:00Z",
		},
		{
			topic: "logic.ordering",
			correct: true,
			skipped: false,
			answered_at: "2026-09-29T11:30:00Z",
		},
		{
			topic: "parity.alternation",
			correct: false,
			skipped: false,
			answered_at: "2026-09-29T11:20:00Z",
		},
		{
			topic: "logic.ordering",
			correct: true,
			skipped: false,
			answered_at: "2026-09-29T11:10:00Z",
		},
	],
	recommendation: {
		topic: "combinatorics.enumeration",
		grade_level: "3-4",
		difficulty: 2,
		goal: "reinforce",
	},
};

/** progress is the progress a card reads when its top line is pressed. */
export const progress = toolResult(standing);

/**
 * inTrial is the progress of a child three tasks into the trial series: no
 * rating yet, nor a rank, and a new topic next.
 */
export const inTrial = {
	...standing,
	trial: { answered: 3, of: 5 },
	overall: null,
	topics: [
		{
			topic: "logic.ordering",
			rating: null,
			answers: 1,
			correct: 1,
			mastered: false,
			skipped: 0,
		},
		{
			topic: "counting.gaps",
			rating: null,
			answers: 1,
			correct: 0,
			mastered: false,
			skipped: 0,
		},
		{
			topic: "time.clocks",
			rating: null,
			answers: 1,
			correct: 1,
			mastered: false,
			skipped: 0,
		},
	],
	recent: [
		{
			topic: "time.clocks",
			correct: true,
			skipped: false,
			answered_at: "2026-09-29T11:30:00Z",
		},
		{
			topic: "counting.gaps",
			correct: false,
			skipped: false,
			answered_at: "2026-09-29T11:20:00Z",
		},
		{
			topic: "logic.ordering",
			correct: true,
			skipped: false,
			answered_at: "2026-09-29T11:10:00Z",
		},
	],
	recommendation: {
		topic: "parity.alternation",
		grade_level: "3-4",
		difficulty: 1,
		goal: "new_topic",
	},
};

/**
 * profileRead is the profile as the model reads it: the details with the
 * parent's notes — for the model to read back, never for a card — and where
 * the file is, with an older file that holds a profile too.
 */
export const profileRead = {
	screen: "profile",
	last_answer: null,
	profile: {
		...standing.profile,
		ui_language: "ru",
		notes: "Loses heart after two mistakes in a row.",
	},
	trial: null,
	recommendation: standing.recommendation,
	location: {
		folder: "MathTrail",
		file: "mathtrail-profile.json",
		link: "https://drive.google.com/file/d/profile/view",
		others: [
			{
				file: "mathtrail-profile (1).json",
				link: "https://drive.google.com/file/d/older/view",
			},
		],
	},
};

/**
 * profileRefused is a change to the profile refused for a field that broke a
 * rule: the profile is handed back as it stays, with no location.
 */
export const profileRefused = {
	screen: "profile",
	status: "rejected",
	code: "invalid_profile",
	problems: [{ field: "pseudonym", rule: "at most 32 characters, 40 given" }],
	last_answer: null,
	profile: { ...standing.profile, notes: "" },
	trial: null,
	recommendation: standing.recommendation,
};

/** firstRun is an account with no profile yet, as the profile's tool finds it. */
export const firstRun = {
	screen: "first_run",
	last_answer: null,
	profile: null,
	trial: null,
	recommendation: null,
};

/** firstRunRefused is a first profile refused for a field that broke a rule. */
export const firstRunRefused = {
	...firstRun,
	status: "rejected",
	code: "invalid_profile",
	problems: [{ field: "grade", rule: "from 1 to 6, 9 given" }],
};

/**
 * longTexts is a task at every limit a card has to fit at its narrowest: a
 * pseudonym as long as a profile allows, a drawing as wide and as tall as the
 * checks allow, a number of fourteen digits and an option of two hundred
 * characters, and a word with no place to break.
 */
export const longTexts: Handed = {
	...fence,
	child: {
		pseudonym: "SuperCometTheGreatExplorer2026XY",
		grade: 6,
		ui_language: null,
	},
	task: {
		...fence.task,
		id: "task_long",
		question: `Supercalifragilisticexpialidociousandthensomemore ${"and a long question that goes on ".repeat(8)}— how many?`,
		drawing: Array.from({ length: 12 }, (_, row) =>
			row % 2 === 0 ? `+${"-".repeat(28)}+` : `|${` ${row}`.padEnd(28)}|`,
		).join("\n"),
		options: {
			A: "12345678901234",
			B: "An answer of many words. ".repeat(8).trim(),
			C: "5",
			D: "6",
			E: "12",
		},
	},
};

/**
 * refused is the payload of a task the model handed in that failed its checks,
 * with attempts left: the card waits while the model writes it again. What the
 * checks found is for the model, and gives nothing away.
 */
export const refused = {
	screen: "waiting",
	status: "rejected",
	code: "solver_disagrees",
	reasons: [
		{
			code: "solver_disagrees",
			messages: [
				"The solver finds a different right option than the one marked.",
			],
		},
	],
	attempt: 1,
	attempts_left: 2,
	last_answer: null,
	child: fence.child,
	task: null,
} as const;

/**
 * exhausted is the payload of the model's last attempt at a task failing its
 * checks: the request is closed, and the model asks for a new one.
 */
export const exhausted = {
	...refused,
	code: "attempts_exhausted",
	attempt: 3,
	attempts_left: 0,
} as const;

/**
 * limited is the payload of a request refused for the day: it says whose card
 * it would be nowhere, since the tool that refuses reads no child's details.
 */
export const limited = {
	screen: "waiting",
	status: "limited",
	code: "limit_reached",
	last_answer: null,
} as const;

// toolResult is a tool's result with the payload a card is drawn from.
function toolResult(
	structuredContent: Record<string, unknown>,
): CallToolResult {
	return {
		content: [{ type: "text", text: "The words for the model." }],
		structuredContent,
	};
}
