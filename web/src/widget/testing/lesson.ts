import type { CallToolResult } from "@modelcontextprotocol/client";
import type { Picture } from "../../design/picture/model";
import type { AnswerResult, HandedTask, TopicChoice } from "../payload";

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
 * fenceRow is the fence's picture: its length and one gap between posts, the
 * row cut short in its middle, as a row is drawn whose posts the child counts,
 * so that the count, which is the answer, cannot be read off it.
 */
export const fenceRow = {
	kind: "row",
	items: [{}, {}, { skip: true }, {}],
	gaps: "3",
	span: "12",
} satisfies Picture;

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
		picture: fenceRow,
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

/**
 * offeredChoice is what the card of a task offers once the trial series is
 * over: the coach chooses the topic, the review suggests Enumeration and
 * Parity and alternation to develop and Pigeonhole principle to begin, and
 * the groups link to the site the progress names.
 */
export const offeredChoice: TopicChoice = {
	chosen: null,
	recommended: [
		"combinatorics.enumeration",
		"parity.alternation",
		"pigeonhole.basic",
	],
	site: { url: "https://mathtrail.app", languages: ["en", "ru"] },
};

/**
 * withTopicChoice is handed with the choice of the topic offered: the one the
 * trial series over offers, with fields given replacing its own.
 */
export function withTopicChoice(
	handed: Handed,
	fields: Partial<TopicChoice> = {},
): Handed {
	return { ...handed, topic_choice: { ...offeredChoice, ...fields } };
}

/** fenceSolution is the solution of the fence, three steps in one text. */
export const fenceSolution =
	"12 ÷ 3 = 4 gaps. A straight fence with posts at both ends has one more post than gaps. 4 + 1 = 5 posts.";

/**
 * fenceSolutionRow is the picture of the fence's solution: all five posts, 3
 * meters apart along 12, the row its task's picture cuts short drawn whole, so
 * that the count the solution comes to is there to be read off it.
 */
export const fenceSolutionRow = {
	kind: "row",
	items: [{}, {}, {}, {}, {}],
	gaps: "3",
	span: "12",
} satisfies Picture;

/** fenceTotal is the equality the fence's solution comes to. */
export const fenceTotal = "12 ÷ 3 + 1 = 5";

/** fenceSolutionInRussian is the Russian one. */
export const fenceSolutionInRussian =
	"12 : 3 = 4 промежутка. У прямого забора со столбами на обоих концах столбов на один больше, чем промежутков. 4 + 1 = 5 столбов.";

/**
 * fenceInArabic is the same task in Arabic, a language written right to
 * left: its words run that way, its picture does not.
 */
export const fenceInArabic: Handed = {
	...fence,
	child: { pseudonym: "مذنّب", grade: 3, ui_language: null },
	task: {
		...fence.task,
		language: "ar",
		question:
			"طول السياج 12 مترًا. تقف الأعمدة كل 3 أمتار، ومنها عمودان عند الطرفين. كم عدد الأعمدة؟",
		hint: "لنبدأ بسياج أقصر: طوله 6 أمتار، وعمود كل 3 أمتار. لنرسمه ونعدّ أعمدته.",
	},
};

/** fenceSolutionInArabic is the Arabic one. */
export const fenceSolutionInArabic =
	"12 ÷ 3 = 4 مسافات. في السياج المستقيم الذي تقف أعمدته عند طرفيه يزيد عدد الأعمدة على عدد المسافات بواحد. 4 + 1 = 5 أعمدة.";

/** flagsPicture is the flags' picture: one flag, in two of the colours the question names. */
export const flagsPicture = {
	kind: "flags",
	colors: { red: "red", blue: "blue" },
	groups: [{ flags: [["red", "blue"]] }],
} satisfies Picture;

/**
 * flagsTask is a task whose pictures paint — flags of two stripes in three
 * colours — as the service hands it to a child's card in English, its
 * picture flagsPicture.
 */
export const flagsTask: Handed = {
	...fence,
	task: {
		id: "task_flags",
		topic: "combinatorics.enumeration",
		language: "en",
		question:
			"A flag has two stripes, one above the other. Each stripe is red, blue or yellow, and the two stripes are of different colours. How many different flags can be made?",
		picture: flagsPicture,
		options: { A: "3", B: "4", C: "5", D: "6", E: "9" },
		hint: "Pick the top stripe first. How many colours are left for the bottom one?",
	},
};

/** flagsSolution is the solution of the flags, a step to each top stripe. */
export const flagsSolution =
	"Let the top stripe be red: the bottom one is blue or yellow, 2 flags. With blue on top, the bottom one is red or yellow: 2 more. With yellow on top, the bottom one is red or blue: 2 more. In all, 2 + 2 + 2 = 6 flags.";

/**
 * flagsSolutionPicture is the picture of the flags' solution: every flag,
 * gathered by its top stripe, each group named by that stripe's colour.
 */
export const flagsSolutionPicture = {
	kind: "flags",
	colors: { red: "red", blue: "blue", yellow: "yellow" },
	groups: [
		{
			color: "red",
			flags: [
				["red", "blue"],
				["red", "yellow"],
			],
		},
		{
			color: "blue",
			flags: [
				["blue", "red"],
				["blue", "yellow"],
			],
		},
		{
			color: "yellow",
			flags: [
				["yellow", "red"],
				["yellow", "blue"],
			],
		},
	],
} satisfies Picture;

/** flagsTotal is the equality the flags' solution comes to. */
export const flagsTotal = "2 + 2 + 2 = 6";

/**
 * resultOf is how an answer to the fence went, as the service tells it: by
 * default the wrong B, whose trap is counting the gaps, with the rating in the
 * topic going down. Fields given replace the default ones.
 */
export function resultOf(fields: Partial<AnswerResult> = {}): AnswerResult {
	return {
		task_id: fence.task.id,
		topic: fence.task.topic,
		choice: "B",
		correct: false,
		correct_answer: "C",
		trap: {
			id: "fence_gaps",
			text: "Counted the gaps instead of the posts.",
			repeated: false,
		},
		solution: fenceSolution,
		hint_used: false,
		rating: { before: 1502, after: 1480 },
		trial: null,
		already_answered: false,
		...fields,
	};
}

/**
 * answered is the result of an answer to the task handed — the fence unless
 * said otherwise — the service recorded, as the card that sent it is told: how
 * the answer went, whole, as the card of how it went shows it, resultOf the
 * fields given, with the choice of the topic given offered.
 */
export function answered(
	fields: Partial<AnswerResult> = {},
	handed: Handed = fence,
	offered?: Partial<TopicChoice>,
): CallToolResult {
	return toolResult(shown(handed, fields, offered));
}

/**
 * answeredByAnEarlierService is the result of an answer to the fence as an
 * earlier service told it to the card that sent it: resultOf the fields given,
 * alone, with nothing of the task, the child or the site beside it.
 */
export function answeredByAnEarlierService(
	fields: Partial<AnswerResult> = {},
): CallToolResult {
	const result = resultOf(fields);
	return toolResult({
		screen: "result",
		last_answer: {
			task_id: result.task_id,
			topic: result.topic,
			correct: result.correct,
		},
		result,
	});
}

/**
 * shown is how an answer to the task handed went, as show_result draws it and
 * as the card that took the answer is told it: whose card it is, the task as
 * it names it in its language, with its topic's page on the site published,
 * how the answer went — resultOf the fields given — and, given offered, the
 * choice of the topic the trial series over offers, with offered replacing
 * its own fields.
 */
export function shown(
	handed: Handed,
	fields: Partial<AnswerResult> = {},
	offered?: Partial<TopicChoice>,
): Record<string, unknown> {
	const { task } = handed;
	const result = resultOf({ task_id: task.id, topic: task.topic, ...fields });
	return {
		screen: "result",
		last_answer: {
			task_id: task.id,
			topic: task.topic,
			correct: result.correct,
		},
		child: handed.child,
		language: task.language,
		task: {
			id: task.id,
			topic: task.topic,
			language: task.language,
			options: task.options,
			slug: "gaps-and-boundaries",
			site_page: true,
		},
		result,
		...(offered === undefined
			? {}
			: { topic_choice: { ...offeredChoice, ...offered } }),
		site: { url: "https://mathtrail.app", languages: ["en", "ru"] },
	};
}

/**
 * nothingShown is the card of how an answer went with nothing to show, as
 * show_result draws it for the task handed: the task has no answer yet, is no
 * longer on the card, or has its answer recorded with its seal lost.
 */
export function nothingShown(
	handed: Handed,
	code: "not_answered" | "stale_task" | "told_no_more",
): object {
	return {
		screen: "result",
		status: code === "not_answered" ? "rejected" : "stale",
		code,
		last_answer: null,
		child: handed.child,
		language: handed.task.language,
	};
}

/**
 * repeatedAnswer is the wrong B again, its trap one the child has fallen for
 * before among the latest answers.
 */
export const repeatedAnswer = answered({
	trap: {
		id: "fence_gaps",
		text: "Counted the gaps instead of the posts.",
		repeated: true,
	},
});

/** rightAnswer is the fence answered with C, which is right. */
export const rightAnswer = answered({
	choice: "C",
	correct: true,
	trap: null,
	rating: { before: 1502, after: 1519 },
});

/**
 * dontKnowAnswer is the fence answered "I don't know" in the chat, a wrong
 * answer with no trap, told again to a card that sends another answer: a card
 * has no button for it, and that is the only way it reaches one.
 */
export const dontKnowAnswer = answered({
	choice: "?",
	trap: null,
	rating: { before: 1502, after: 1488 },
	already_answered: true,
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
	trap: { id: "fence_ends", text: "Counted one end twice.", repeated: false },
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
 * progress — the grades the ranks are matched with marked under the overall
 * course, as the service matches them, four topics met, each with its rank,
 * one ahead of the overall rank, two even with it and one behind it, one of
 * them mastered, and a topic within reach not met yet; the latest answers with
 * one task left without an answer, two mistakes that keep coming back, the
 * review of the topics, the topic worked over again after a mistake next, and
 * where the profile's file is. Every topic's page is published on the site,
 * which is written in English and Russian. The review is the one the service
 * makes of these numbers: Ordering mastered and well above the overall level,
 * risen over the week; Enumeration to develop for the mistake it keeps making,
 * first since the rule sets it next, and Parity and alternation for standing
 * low and falling over the week; Gaps and boundaries met too few times to
 * judge; and a step for each topic to develop, the mistake that repeats most
 * advised among them.
 */
export const standing = {
	screen: "progress",
	last_answer: {
		task_id: "task_fence",
		topic: "combinatorics.enumeration",
		correct: false,
	},
	profile: {
		pseudonym: "Comet",
		grade: 3,
		interests: ["space", "animals", "football"],
		excluded_skills: ["division_with_remainder"],
		ui_language: null,
		country: null,
		region: null,
		signin_country_off: false,
	},
	trial: null,
	overall: {
		rating: 1573,
		rank: 3,
		ranks: 11,
		share: 43,
		grades: [
			{ grade_level: "1-2", first_rank: 1, last_rank: 4 },
			{ grade_level: "3-4", first_rank: 5, last_rank: 7 },
			{ grade_level: "5-6", first_rank: 8, last_rank: 11 },
		],
	},
	topics: [
		{
			topic: "logic.ordering",
			slug: "ordering",
			site_page: true,
			rating: 1712,
			rank: 4,
			share: 27,
			compared: "ahead",
			answers: 6,
			correct: 5,
			mastered: true,
			skipped: 0,
		},
		{
			topic: "combinatorics.enumeration",
			slug: "enumeration",
			site_page: true,
			rating: 1627,
			rank: 3,
			share: 76,
			compared: "even",
			answers: 5,
			correct: 2,
			mastered: false,
			skipped: 1,
		},
		{
			topic: "counting.gaps",
			slug: "gaps-and-boundaries",
			site_page: true,
			rating: 1588,
			rank: 3,
			share: 53,
			compared: "even",
			answers: 3,
			correct: 2,
			mastered: false,
			skipped: 0,
		},
		{
			topic: "parity.alternation",
			slug: "parity-and-alternation",
			site_page: true,
			rating: 1480,
			rank: 2,
			share: 87,
			compared: "behind",
			answers: 5,
			correct: 2,
			mastered: false,
			skipped: 0,
		},
		{
			topic: "pigeonhole.basic",
			slug: "pigeonhole-principle",
			site_page: true,
			rating: null,
			rank: null,
			share: null,
			compared: null,
			answers: 0,
			correct: 0,
			mastered: false,
			skipped: 0,
		},
	],
	recent: [
		{
			topic: "combinatorics.enumeration",
			correct: false,
			skipped: false,
		},
		{
			topic: "combinatorics.enumeration",
			correct: null,
			skipped: true,
		},
		{
			topic: "counting.gaps",
			correct: true,
			skipped: false,
		},
		{
			topic: "logic.ordering",
			correct: true,
			skipped: false,
		},
		{
			topic: "parity.alternation",
			correct: false,
			skipped: false,
		},
		{
			topic: "logic.ordering",
			correct: true,
			skipped: false,
		},
	],
	skipped: 1,
	mistakes: [
		{ trap: "missed_case", times: 3 },
		{ trap: "double_count", times: 2 },
	],
	review: {
		strong: [
			{ topic: "logic.ordering", reasons: ["mastered", "high", "rose"] },
		],
		develop: [
			{
				topic: "combinatorics.enumeration",
				reasons: ["trap"],
				trap: "missed_case",
			},
			{ topic: "parity.alternation", reasons: ["low", "fell"] },
		],
		early: ["counting.gaps"],
		steps: [
			{
				kind: "trap",
				topic: "combinatorics.enumeration",
				trap: "missed_case",
			},
			{ kind: "rhythm", topic: "parity.alternation" },
		],
	},
	recommendation: {
		topic: "combinatorics.enumeration",
		grade_level: "3-4",
		difficulty: 2,
		goal: "reinforce",
	},
	location: {
		folder: "MathTrail",
		file: "mathtrail-profile.json",
		link: "https://drive.google.com/file/d/profile/view",
		others: [],
	},
	site: { url: "https://mathtrail.app", languages: ["en", "ru"] },
};

// The moves of each topic of moving, by its id.
const movesOfTopic: Record<
	string,
	{ last_task: object | null; week: object | null } | undefined
> = {
	"logic.ordering": {
		last_task: { rank: 4, share: 27, moved: "same" },
		week: { rank: 3, share: 90, moved: "rank_up" },
	},
	"combinatorics.enumeration": {
		last_task: { rank: 3, share: 85, moved: "back" },
		week: { rank: 3, share: 60, moved: "forward" },
	},
	"counting.gaps": {
		last_task: { rank: 3, share: 53, moved: "same" },
		week: { rank: 3, share: 53, moved: "same" },
	},
	"parity.alternation": {
		last_task: { rank: 2, share: 87, moved: "same" },
		week: { rank: 3, share: 5, moved: "rank_down" },
	},
};

/**
 * moving is where Comet stands with how the ranks moved: over the week the
 * overall rank rose from the one below, and the last task, a wrong answer in
 * Enumeration, took a little of it back; in the topics, over the week Ordering
 * reached a new rank, Enumeration went forward within its own, Gaps and
 * boundaries stayed where it was and Parity and alternation fell a rank, and
 * since the last task only Enumeration moved, back. Pigeonhole principle, not
 * met yet, has no moves.
 */
export const moving = {
	...standing,
	overall: {
		...standing.overall,
		change: {
			last_task: { rating: 1580, rank: 3, share: 47, moved: "back" },
			week: { rating: 1467, rank: 2, share: 80, moved: "rank_up" },
		},
	},
	topics: standing.topics.map((topic) => ({
		...topic,
		change: movesOfTopic[topic.topic],
	})),
};

/** progress is the progress a card reads when its top line is pressed. */
export const progress = toolResult(standing);

/**
 * progressMoving is the progress a card reads when its top line is pressed,
 * with how the ranks moved.
 */
export const progressMoving = toolResult(moving);

/**
 * standingBefore is where Comet stands, as a progress from before the topics
 * had ranks of their own says it: no share of the rank, no grades marked under
 * it, no rank of a topic, no topic not met yet, no total of the skips and no
 * word of where the file is.
 * A card of an earlier chat, drawn again, is drawn from one.
 */
export const standingBefore = {
	screen: standing.screen,
	last_answer: standing.last_answer,
	profile: standing.profile,
	trial: null,
	overall: { rating: 1573, rank: 3, ranks: 11 },
	topics: standing.topics
		.filter((topic) => topic.answers > 0)
		.map(({ topic, rating, answers, correct, mastered, skipped }) => ({
			topic,
			rating,
			answers,
			correct,
			mastered,
			skipped,
		})),
	recent: standing.recent,
	mistakes: standing.mistakes,
	recommendation: standing.recommendation,
};

/**
 * atTheTop is where a child stands who has reached the highest rank: the
 * whole way of it behind, with no rank left to come.
 */
export const atTheTop = {
	...standing,
	overall: { ...standing.overall, rating: 2905, rank: 11, share: 100 },
};

/**
 * inTrial is the progress of a child three tasks into the trial series: no
 * rating yet, nor a rank, nor a review, no mistake made twice, a topic within
 * reach not met yet, and a new topic next.
 */
export const inTrial = {
	...standing,
	trial: { answered: 3, of: 5 },
	overall: null,
	review: undefined,
	topics: [
		{
			topic: "logic.ordering",
			rating: null,
			rank: null,
			share: null,
			compared: null,
			answers: 1,
			correct: 1,
			mastered: false,
			skipped: 0,
		},
		{
			topic: "counting.gaps",
			rating: null,
			rank: null,
			share: null,
			compared: null,
			answers: 1,
			correct: 0,
			mastered: false,
			skipped: 0,
		},
		{
			topic: "time.clocks",
			rating: null,
			rank: null,
			share: null,
			compared: null,
			answers: 1,
			correct: 1,
			mastered: false,
			skipped: 0,
		},
		{
			topic: "parity.alternation",
			rating: null,
			rank: null,
			share: null,
			compared: null,
			answers: 0,
			correct: 0,
			mastered: false,
			skipped: 0,
		},
	],
	skipped: 0,
	recent: [
		{
			topic: "time.clocks",
			correct: true,
			skipped: false,
		},
		{
			topic: "counting.gaps",
			correct: false,
			skipped: false,
		},
		{
			topic: "logic.ordering",
			correct: true,
			skipped: false,
		},
	],
	mistakes: [],
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
 * editSaved is a change from the form saved: the details as they now stand,
 * and the words the card hands the model.
 */
export function editSaved(
	profile: Record<string, unknown> = {},
): CallToolResult {
	return {
		content: [{ type: "text", text: savedWords }],
		structuredContent: {
			screen: "profile",
			changed: true,
			profile: { ...standing.profile, ...profile },
		},
	};
}

/**
 * topicSaved is a choice of the topic from the card saved: the details as they
 * now stand, the topic among them, and the words the card hands the model.
 */
export function topicSaved(topic: string | null): CallToolResult {
	return {
		content: [{ type: "text", text: topicWords }],
		structuredContent: {
			screen: "profile",
			changed: true,
			profile: { ...standing.profile, lesson_topic: topic },
		},
	};
}

/** topicWords are what the service tells the model of a choice of the topic. */
export const topicWords =
	"On the card, the child or the adult chose the topic of the lessons.";

/** savedWords are what the service tells the model of a change from the form. */
export const savedWords =
	"The adult changed the child's profile with the form on a card.";

/**
 * editRefused is a change from the form refused for every field it shows, each
 * by the code of the rule it broke, the profile handed back as it stays.
 */
export const editRefused = toolResult({
	screen: "profile",
	status: "rejected",
	code: "invalid_profile",
	problems: [
		{ field: "excluded_skills", code: "not_in_catalog", rule: "entry 2" },
		{ field: "grade", code: "out_of_range", rule: "1 to 6" },
		{ field: "interests", code: "entry_length", rule: "1 to 40" },
		{ field: "pseudonym", code: "required", rule: "is required" },
		{ field: "ui_language", code: "not_a_language", rule: "a BCP 47 tag" },
	],
	changed: false,
	profile: standing.profile,
});

/** editGone is a change from the form for a profile no longer there. */
export const editGone = toolResult({
	screen: "first_run",
	status: "stale",
	code: "stale_profile",
	changed: false,
	profile: null,
});

/**
 * longTexts is a task at every limit a card has to fit at its narrowest: a
 * pseudonym as long as a profile allows, a number of fourteen digits and an
 * option of two hundred characters, and a word with no place to break. It has
 * no picture: the pictures at their limits are scenes of their own.
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
		picture: undefined,
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
 * staleWait is the payload of a task handed in for a request that is not the
 * open one, with no task on the card: nothing was checked, and the model asks
 * for a new task.
 */
export const staleWait = {
	screen: "waiting",
	status: "stale",
	code: "stale_request",
	last_answer: null,
	child: fence.child,
	task: null,
} as const;

/**
 * askRefused is the payload of a task asked for with arguments no request
 * could be opened from: the model asks again, and no task comes to this card.
 */
export const askRefused = {
	screen: "waiting",
	status: "rejected",
	code: "invalid_arguments",
	problems: [
		{
			field: "topic",
			code: "not_in_catalog",
			rule: "is not a topic of the catalog",
		},
	],
	last_answer: null,
	child: fence.child,
} as const;

/**
 * limited is the payload of a request refused for the day, as an earlier
 * service sent it: it says whose card it would be nowhere.
 */
export const limited = {
	screen: "waiting",
	status: "limited",
	code: "limit_reached",
	last_answer: null,
} as const;

/**
 * coming is the payload of a task asked for: the card next_task draws, which
 * waits for the task of the request it names.
 */
export const coming = {
	screen: "coming",
	request_id: "req_fence",
	child: fence.child,
	language: "en",
	last_answer: null,
} as const;

/**
 * writing is what the service tells a card of the task it waits for while it
 * is being written, with as many tries turned down as given.
 */
export function writing(refused = 0): CallToolResult {
	return toolResult({
		screen: "coming",
		refused,
		last_answer: null,
		child: fence.child,
		task: null,
		language: "en",
	});
}

/**
 * onTheCard is what the service tells a card once the task it waits for is on
 * the child's card: the fence.
 */
export const onTheCard = toolResult({ ...fence, language: "en" });

/**
 * notComing is what the service tells a card whose request is over with no
 * task of its on the card.
 */
export const notComing = toolResult({
	screen: "waiting",
	code: "stale_request",
	last_answer: null,
	child: fence.child,
	task: null,
});

// toolResult is a tool's result with the payload a card is drawn from.
function toolResult(
	structuredContent: Record<string, unknown>,
): CallToolResult {
	return {
		content: [{ type: "text", text: "The words for the model." }],
		structuredContent,
	};
}
