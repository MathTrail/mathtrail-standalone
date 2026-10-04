import type { CallToolResult } from "@modelcontextprotocol/client";
import type { McpUiHostContext } from "@modelcontextprotocol/ext-apps";
import type { CallStage } from "../widget/bridge";
import { rankCount } from "../widget/names";
import type { AnswerResult } from "../widget/payload";
import {
	answered,
	atTheTop,
	editGone,
	editRefused,
	editSaved,
	exhausted,
	failure,
	fence,
	fenceInArabic,
	fenceInRussian,
	fenceSolution,
	fenceSolutionInArabic,
	fenceSolutionInRussian,
	firstRun,
	firstRunRefused,
	inTrial,
	limited,
	longTexts,
	moving,
	notComing,
	profileRead,
	profileRefused,
	progress,
	refused,
	staleAnswer,
	staleWait,
	standing,
	standingBefore,
	writing,
} from "../widget/testing/lesson";

/**
 * Scene is one state of a lesson on a card: the payload the card is drawn
 * from — a task handed to it, a task on its way, or a card a task did not come
 * to — or, for a task caught being asked for, how far its call has got; what the service
 * answers the card's calls with, whether the host takes its messages, what
 * the host keeps of the screen's edges, and what the child does on the card
 * to reach the state.
 */
export type Scene = {
	name: string;
	answers?: (tool: string) => Promise<CallToolResult>;
	refuseMessages?: boolean;
	insets?: McpUiHostContext["safeAreaInsets"];
	play?: (card: Document) => void;
} & (
	| { payload: object; caught?: undefined }
	| { caught: CallStage; payload?: undefined }
);

// Fence is the fence in one language: the task as it is handed to a card,
// the trap behind B and D, and the solution.
type Fence = {
	handed: typeof fence;
	gaps: string;
	ends: string;
	solution: string;
};

const fenceInEnglish: Fence = {
	handed: fence,
	gaps: "Counted the gaps instead of the posts.",
	ends: "Counted one end twice.",
	solution: fenceSolution,
};

// The fence in the languages it is written in. A card in any other language
// is handed the English one, as a child whose cards speak one language may be
// given a task in another.
const fences: Readonly<Record<string, Fence>> = {
	en: fenceInEnglish,
	ru: {
		handed: fenceInRussian,
		gaps: "Посчитаны промежутки, а не столбы.",
		ends: "Один из концов посчитан дважды.",
		solution: fenceSolutionInRussian,
	},
	ar: {
		handed: fenceInArabic,
		gaps: "عُدّت المسافات بدلًا من الأعمدة.",
		ends: "عُدّ أحد الطرفين مرتين.",
		solution: fenceSolutionInArabic,
	},
};

// A reply that never comes, for a card caught while it checks an answer.
const never = new Promise<CallToolResult>(() => {});

/**
 * scenesIn are every scene of a lesson, on a card whose words are in language:
 * the task and its result in that language when the fence is written in it,
 * and in English otherwise.
 */
export function scenesIn(language: string): Scene[] {
	const words = fences[language] ?? fenceInEnglish;
	const handed = words.handed;
	const result = (fields: Partial<AnswerResult> = {}) =>
		Promise.resolve(
			answered({
				trap: { id: "fence_gaps", text: words.gaps, repeated: false },
				solution: words.solution,
				...fields,
			}),
		);
	const service =
		(fields: Partial<AnswerResult> = {}) =>
		(tool: string) =>
			tool === "read_progress" ? Promise.resolve(progress) : result(fields);
	return [
		{ name: "task", payload: handed },
		{
			name: "selected (checking)",
			payload: handed,
			answers: () => never,
			play: option("B"),
		},
		{ name: "hint", payload: handed, play: button(1) },
		{ name: "wrong", payload: handed, answers: service(), play: option("B") },
		{
			name: "wrong, a mistake made before",
			payload: handed,
			answers: service({
				trap: { id: "fence_gaps", text: words.gaps, repeated: true },
			}),
			play: option("B"),
		},
		{
			name: "right",
			payload: handed,
			answers: service({
				choice: "C",
				correct: true,
				trap: null,
				rating: { before: 1502, after: 1519 },
			}),
			play: option("C"),
		},
		{
			name: "I don't know",
			payload: handed,
			answers: service({
				choice: "?",
				trap: null,
				rating: { before: 1502, after: 1488 },
			}),
			play: button(0),
		},
		{
			name: "trial series, 3 of 5",
			payload: handed,
			answers: service({ rating: null, trial: { answered: 3, of: 5 } }),
			play: option("B"),
		},
		{
			name: "trial series, 5 of 5",
			payload: handed,
			answers: service({ rating: null, trial: { answered: 5, of: 5 } }),
			play: option("B"),
		},
		{
			name: "answered before, with D",
			payload: handed,
			answers: service({
				choice: "D",
				trap: { id: "fence_ends", text: words.ends, repeated: false },
				already_answered: true,
			}),
			play: option("B"),
		},
		{
			name: "closed",
			payload: handed,
			answers: () => Promise.resolve(staleAnswer),
			play: option("B"),
		},
		{
			name: "not recorded",
			payload: handed,
			answers: () => Promise.resolve(failure),
			play: option("B"),
		},
		{
			name: "handed out again",
			payload: { ...handed, status: "stale", code: "stale_request" },
		},
		{ name: "progress", payload: handed, answers: service(), play: topLine },
		{
			name: "progress over a task, every section open",
			payload: handed,
			answers: service(),
			play: progressUnfolded,
		},
		{ name: "another task asked", payload: handed, play: button(2) },
		{
			name: "another task, the ask not sent",
			payload: handed,
			refuseMessages: true,
			play: button(2),
		},
		{ name: "a task being asked for", caught: "started" },
		{ name: "a task asked for, then stopped", caught: "cancelled" },
		{
			name: "a task being written",
			payload: comingFor(handed),
			answers: () => Promise.resolve(writing()),
		},
		{
			name: "a try turned down, a new one being written",
			payload: comingFor(handed),
			answers: () => Promise.resolve(writing(1)),
		},
		{
			name: "the task come to the card that waited",
			payload: comingFor(handed),
			answers: () =>
				Promise.resolve({ content: [], structuredContent: { ...handed } }),
		},
		{
			name: "a card drawn again, its task no longer here",
			payload: comingFor(handed),
			answers: () => Promise.resolve(notComing),
		},
		{
			name: "a try that did not pass",
			payload: { ...refused, child: handed.child },
		},
		{
			name: "a request that is over",
			payload: { ...staleWait, child: handed.child },
		},
		{
			name: "attempts exhausted",
			payload: { ...exhausted, child: handed.child },
		},
		{ name: "limit reached", payload: limited },
		{ name: "progress, the model's card", payload: moving, play: everyFold },
		{
			name: "progress since the last task",
			payload: moving,
			play: inTurn(pickedPeriod("last_task"), everyFold),
		},
		{
			name: "progress, a step back over the week",
			payload: withOverallWeek({
				rating: 1590,
				rank: 3,
				share: 55,
				moved: "back",
			}),
			play: everyFold,
		},
		{
			name: "progress, a rank down over the week",
			payload: withOverallWeek({
				rating: 1670,
				rank: 4,
				share: 3,
				moved: "rank_down",
			}),
			play: everyFold,
		},
		{
			name: "progress, nothing moved over the week",
			payload: stillOverTheWeek,
			play: everyFold,
		},
		{
			name: "progress, a topic new to the week",
			payload: withNewTopic,
			play: everyFold,
		},
		{
			name: "progress, a week that cannot be told yet",
			payload: weekUntold,
			play: everyFold,
		},
		{
			name: "progress, a week that cannot be told yet, chosen",
			payload: weekUntold,
			play: inTurn(pickedPeriod("week"), everyFold),
		},
		{ name: "progress in the trial series", payload: inTrial, play: everyFold },
		{ name: "progress, long texts", payload: longProgress, play: everyFold },
		{
			name: "progress at the highest rank",
			payload: atTheTop,
			play: everyFold,
		},
		{
			name: "progress from an earlier release",
			payload: standingBefore,
			play: everyFold,
		},
		{
			name: "profile",
			payload: {
				...profileRead,
				profile: { ...profileRead.profile, ui_language: "pt-BR" },
			},
		},
		{ name: "profile, a change refused", payload: profileRefused },
		{ name: "profile form", payload: standing, play: inTheForm() },
		{
			name: "profile form, long texts",
			payload: longProgress,
			play: inTheForm(),
		},
		{
			name: "profile form, refused field by field",
			payload: standing,
			answers: () => Promise.resolve(editRefused),
			play: inTheForm(renamed, saved),
		},
		{
			name: "profile form, saving",
			payload: standing,
			answers: () => never,
			play: inTheForm(renamed, saved),
		},
		{
			name: "profile form, not saved",
			payload: standing,
			answers: () => Promise.resolve(failure),
			play: inTheForm(renamed, saved),
		},
		{
			name: "profile saved, the language changed",
			payload: standing,
			answers: () => Promise.resolve(editSaved({ ui_language: "fr" })),
			play: inTheForm(inFrench, saved),
		},
		{
			name: "profile gone",
			payload: standing,
			answers: () => Promise.resolve(editGone),
			play: inTheForm(renamed, saved),
		},
		{ name: "first sign-in", payload: firstRun },
		{ name: "first sign-in, a profile refused", payload: firstRunRefused },
		{ name: "a card it cannot show", payload: { screen: "result" } },
		{ name: "long texts", payload: longTexts },
		{
			name: "room kept at the edges",
			payload: handed,
			insets: { top: 24, right: 0, bottom: 34, left: 0 },
		},
	];
}

// option presses the option with letter.
function option(letter: string) {
	return (card: Document) => {
		for (const row of card.querySelectorAll<HTMLElement>(".mt-option")) {
			if (row.querySelector(".mt-option-letter")?.textContent === letter) {
				row.click();
			}
		}
	};
}

// button presses the card's button at place: 0 "I don't know", 1 the hint,
// 2 another task. They are found by place because their words change with
// the language.
function button(place: number) {
	return (card: Document) => {
		card.querySelectorAll<HTMLElement>(".mt-btns .mt-btn")[place]?.click();
	};
}

// topLine presses the line at the top of the card.
function topLine(card: Document) {
	card.querySelector<HTMLElement>(".mt-bar")?.click();
}

// longProgress is the progress at every limit a card has to fit at its
// narrowest: a pseudonym as long as a profile allows, every topic of the
// catalog listed — every rank among them, each step of the ramp, and topics
// not met yet — the next rank the one with the longest name, as many
// interests, as long, as a profile holds, and the profile's file named at
// length beside other files that hold one.
const longProgress = {
	...standing,
	profile: {
		...standing.profile,
		pseudonym: "SuperCometTheGreatExplorer2026XY",
		interests: Array.from({ length: 10 }, (_, at) =>
			`${at + 1} a long interest of forty characters!`.slice(0, 40),
		),
		excluded_skills: [
			"division_with_remainder",
			"fractions_arithmetic",
			"order_of_operations",
		],
	},
	topics: [
		"logic.ordering",
		"logic.knights_liars",
		"combinatorics.enumeration",
		"counting.gaps",
		"time.clocks",
		"time.calendar",
		"pigeonhole.basic",
		"parity.alternation",
		"arithmetic.tricks",
		"algorithms.weighing_pouring",
		"fractions.parts",
		"percent.basic",
		"ratio.sharing",
		"geometry.grid",
		"number.divisibility",
		"logic.sets",
		"games.strategy",
	].map((topic, at) => rankedAt(topic, at)),
	overall: {
		rating: 2700,
		rank: 10,
		ranks: 11,
		share: 23,
		change: {
			last_task: { rating: 2690, rank: 10, share: 17, moved: "forward" },
			week: { rating: 2640, rank: 9, share: 86, moved: "rank_up" },
		},
	},
	skipped: 8,
	// The window's twenty answers shared by the mistakes with the longest
	// names: one behind eight of them, the rest behind two each.
	mistakes: [
		{ trap: "ratio_total_confusion", times: 8 },
		{ trap: "number_from_text", times: 2 },
		{ trap: "best_case_not_worst", times: 2 },
		{ trap: "reversed_relation", times: 2 },
		{ trap: "off_by_one", times: 2 },
		{ trap: "first_move_assumed", times: 2 },
		{ trap: "percent_wrong_base", times: 2 },
	],
	location: {
		folder: "MathTrail, the olympiad maths of the whole family",
		file: "mathtrail-profile set aside 2026-10-03.json",
		link: "https://drive.google.com/file/d/profile/view",
		others: [
			{ file: "mathtrail-profile set aside 2026-09-12.json" },
			{ file: "mathtrail-profile (1).json" },
			{ file: "mathtrail-profile (2).json" },
		],
	},
};

// rankedAt is the topic at place at of the long progress: the first fifteen
// each a rank in turn, from the first to the last and again, the overall
// rank's tenth ahead of, even with or behind them, and the last two not met
// yet.
function rankedAt(topic: string, at: number) {
	if (at >= 15) {
		return {
			topic,
			rating: null,
			rank: null,
			share: null,
			compared: null,
			answers: 0,
			correct: 0,
			mastered: false,
			skipped: 0,
		};
	}
	const rank = (at % rankCount) + 1;
	// The highest rank is drawn with the whole way behind it, as the service
	// sends it.
	const share = rank === rankCount ? 100 : (at * 37) % 100;
	let compared = "even";
	if (rank > 10) {
		compared = "ahead";
	} else if (rank < 10) {
		compared = "behind";
	}
	return {
		topic,
		rating: 1168 + (rank - 1) * 166 + Math.floor(share * 1.66),
		rank,
		share,
		compared,
		answers: 5,
		correct: 4,
		mastered: at % 3 === 0,
		skipped: at % 5 === 0 ? 2 : 0,
		change: {
			last_task: movedFrom(rank, share, [0.05, 0, -0.05][at % 3] ?? 0),
			week:
				at % 7 === 6
					? { moved: "new" }
					: movedFrom(rank, share, [0.6, 0.2, -0.2, -0.7, 0, 1.3][at % 6] ?? 0),
		},
	};
}

// movedFrom is a move to rank and share from where a rating stood by ranks
// lower — a step back where by is below nothing — told the way the service
// tells it: by the rank, and then by the share within one, the highest rank
// standing at the whole way through it, where no move is drawn. It counts in
// whole percents of a rank, so that no move is made of what a fraction leaves
// over.
function movedFrom(rank: number, share: number, by: number) {
	const percents = (rank - 1) * 100 + share - Math.round(by * 100);
	const held = Math.min(Math.max(percents, 0), rankCount * 100 - 1);
	const before = { rank: Math.floor(held / 100) + 1, share: held % 100 };
	if (before.rank === rankCount) {
		before.share = 100;
	}
	let moved = "same";
	if (before.rank !== rank) {
		moved = before.rank < rank ? "rank_up" : "rank_down";
	} else if (before.share !== share && rank < rankCount) {
		moved = before.share < share ? "forward" : "back";
	}
	return { ...before, moved };
}

// withOverallWeek is how the ranks moved, with the overall rank's week told
// as week says.
function withOverallWeek(week: object) {
	return {
		...moving,
		overall: {
			...moving.overall,
			change: { ...moving.overall.change, week },
		},
	};
}

// stillOverTheWeek is how the ranks moved, with nothing moved over the week:
// the overall rank and every topic stood where they stand.
const stillOverTheWeek = {
	...moving,
	overall: {
		...moving.overall,
		change: {
			...moving.overall.change,
			week: { rating: 1573, rank: 3, share: 43, moved: "same" },
		},
	},
	topics: moving.topics.map((topic) =>
		topic.rank === null
			? topic
			: {
					...topic,
					change: {
						last_task: { rank: topic.rank, share: topic.share, moved: "same" },
						week: { rank: topic.rank, share: topic.share, moved: "same" },
					},
				},
	),
};

// withNewTopic is how the ranks moved, with Clocks first answered over the
// week, and since the last task standing where it stands.
const withNewTopic = {
	...moving,
	topics: [
		...moving.topics,
		{
			topic: "time.clocks",
			rating: 1400,
			rank: 2,
			share: 40,
			compared: "behind",
			answers: 1,
			correct: 1,
			mastered: false,
			skipped: 0,
			change: {
				last_task: { rank: 2, share: 40, moved: "same" },
				week: { moved: "new" },
			},
		},
	],
};

// weekUntold is how the ranks moved, with the week one the history cannot
// tell yet — answers kept before the service kept where they started — and
// the last task told.
const weekUntold = {
	...moving,
	overall: {
		...moving.overall,
		change: { ...moving.overall.change, week: null },
	},
	topics: moving.topics.map((topic) =>
		topic.change === undefined
			? topic
			: { ...topic, change: { ...topic.change, week: null } },
	),
};

// pickedPeriod picks the while the moves are drawn over on the progress's
// switch.
function pickedPeriod(period: "last_task" | "week") {
	return (card: Document) => {
		card
			.querySelector<HTMLInputElement>(`.mt-switch-input[value="${period}"]`)
			?.click();
	};
}

// inTurn does each step in turn, each once the card has taken what came
// before it in.
function inTurn(...steps: ((card: Document) => void)[]) {
	return (card: Document) => {
		steps.forEach((step, at) => {
			setTimeout(() => step(card), (at + 1) * 100);
		});
	};
}

// comingFor is the card a task asked for comes to, for the child the task
// given is for. It names no language of the lesson, so that its words are in
// the language the preview shows, as the words of a task's card are.
function comingFor(handed: typeof fence) {
	return {
		screen: "coming",
		request_id: "req_preview",
		child: handed.child,
		last_answer: null,
	};
}

// everyFold opens every section of the progress still folded, as a person
// opening each in turn does: the sections it opened before stay open.
function everyFold(card: Document) {
	for (const title of card.querySelectorAll<HTMLElement>(
		'.mt-fold-button[aria-expanded="false"]',
	)) {
		title.click();
	}
}

// progressUnfolded opens the progress over the task from the line at its top
// and, once the progress is drawn, every section of it.
function progressUnfolded(card: Document) {
	topLine(card);
	setTimeout(() => everyFold(card), 100);
}

// inTheForm opens the profile's section of the progress and the form in it,
// and does each step on the form in turn, each once the card has taken the one
// before in. The section is found from the button that opens the form, so
// that it is found in any language.
function inTheForm(...steps: ((card: Document) => void)[]) {
	return (card: Document) => {
		const edit = card.querySelector<HTMLElement>(".mt-fields-head .mt-btn");
		edit
			?.closest(".mt-fold")
			?.querySelector<HTMLElement>('.mt-fold-button[aria-expanded="false"]')
			?.click();
		edit?.click();
		inTurn(...steps)(card);
	};
}

// renamed types another pseudonym into the form, as a person does.
function renamed(card: Document) {
	const box = card.querySelector<HTMLInputElement>(".mt-form .mt-input");
	if (box !== null) {
		box.value = "Nova";
		box.dispatchEvent(new Event("input", { bubbles: true }));
	}
}

// inFrench chooses French for the lessons.
function inFrench(card: Document) {
	const lists = card.querySelectorAll<HTMLSelectElement>(".mt-form select");
	const language = lists[lists.length - 1];
	if (language !== undefined) {
		language.value = "fr";
		language.dispatchEvent(new Event("change", { bubbles: true }));
	}
}

// saved presses the form's button that saves it.
function saved(card: Document) {
	card.querySelector<HTMLElement>(".mt-form .mt-btn-primary")?.click();
}
