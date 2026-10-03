import type { CallToolResult } from "@modelcontextprotocol/client";
import type { McpUiHostContext } from "@modelcontextprotocol/ext-apps";
import type { CallStage } from "../widget/bridge";
import type { AnswerResult } from "../widget/payload";
import {
	answered,
	atTheTop,
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
	profileRead,
	profileRefused,
	progress,
	refused,
	staleAnswer,
	staleWait,
	standing,
	standingBefore,
} from "../widget/testing/lesson";

/**
 * Scene is one state of a lesson on a card: the payload the card is drawn
 * from — a task handed to it, or a card a task did not come to — or, for a
 * task caught being handed in, how far its call has got; what the service
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
		{ name: "another task asked", payload: handed, play: button(2) },
		{
			name: "another task, the ask not sent",
			payload: handed,
			refuseMessages: true,
			play: button(2),
		},
		{ name: "a task being written", caught: "started" },
		{ name: "a task being checked", caught: "running" },
		{ name: "a task not finished", caught: "cancelled" },
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
		{ name: "progress, the model's card", payload: standing },
		{ name: "progress in the trial series", payload: inTrial },
		{ name: "progress, long texts", payload: longProgress },
		{ name: "progress at the highest rank", payload: atTheTop },
		{ name: "progress from an earlier release", payload: standingBefore },
		{
			name: "profile",
			payload: {
				...profileRead,
				profile: { ...profileRead.profile, ui_language: "pt-BR" },
			},
		},
		{ name: "profile, a change refused", payload: profileRefused },
		{ name: "first sign-in", payload: firstRun },
		{
			name: "first sign-in, ticked and asked",
			payload: firstRun,
			play: askForProfile,
		},
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
	overall: { rating: 2700, rank: 10, ranks: 11, share: 23 },
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
	const rank = (at % 11) + 1;
	const share = (at * 37) % 100;
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
	};
}

// askForProfile ticks the adult's statement and asks for the profile.
function askForProfile(card: Document) {
	card.querySelector<HTMLElement>(".mt-check span")?.click();
	// The button is switched on once the card has taken the tick in.
	setTimeout(() => {
		card.querySelector<HTMLElement>(".mt-btn-primary")?.click();
	});
}
