import type { CallToolResult } from "@modelcontextprotocol/client";
import type { McpUiHostContext } from "@modelcontextprotocol/ext-apps";
import topics from "../../../content/catalogs/topics.json";
import type { CallStage } from "../widget/bridge";
import { cardWords } from "../widget/dictionaries";
import { topicName } from "../widget/names";
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
	topicSaved,
	withTopicChoice,
	writing,
} from "../widget/testing/lesson";
import { topicGroups } from "../widget/topicGroups";

/**
 * Scene is one state of a lesson on a card: the payload the card is drawn
 * from — a task handed to it, a task on its way, or a card a task did not come
 * to — or, for a task caught being asked for, how far its call has got; what the service
 * answers the card's calls with, told how many calls of the same tool the card
 * made before, whether the host takes its messages, whether
 * it opens pages — it says it does and opens them, unless links says it opens
 * none or refuses each —, what the host keeps of the screen's edges, and what
 * the child does on the card to reach the state.
 */
export type Scene = {
	name: string;
	answers?: (tool: string, before: number) => Promise<CallToolResult>;
	refuseMessages?: boolean;
	links?: "none" | "refuse";
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

// longestTopicIn is the topic the choice offers whose name is the longest in
// language, a character counted as one however many units it takes.
function longestTopicIn(language: string): string {
	const words = cardWords(language, undefined);
	const length = (topic: string) => [...topicName(words, topic)].length;
	const offered = topicGroups.flatMap((group) =>
		group.topics.map((topic) => topic.id),
	);
	return offered.reduce(
		(longest, topic) => (length(topic) > length(longest) ? topic : longest),
		offered[0] ?? "",
	);
}

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
	// offering is the task once the trial series is over, with the choice of
	// the topic it offers.
	const offering = withTopicChoice(handed);
	return [
		{ name: "task", payload: handed },
		{ name: "task after the trial series", payload: offering },
		{
			name: "wrong after the trial series",
			payload: offering,
			answers: service(),
			play: option("B"),
		},
		// The same task with its row of posts cut short, as a row the child
		// counts is drawn, so that the count, which is the answer, cannot be read
		// off the card before the child answers; and the progress with its
		// review alone open, a card short enough to take in at once.
		{
			name: "task after the trial series, its row cut short",
			payload: rowCut(offering),
		},
		{
			name: "hint after the trial series, its row cut short",
			payload: rowCut(offering),
			play: button(0),
		},
		{
			name: "wrong after the trial series, its row cut short",
			payload: rowCut(offering),
			answers: service(),
			play: option("B"),
		},
		{
			name: "progress, the review open",
			payload: reviewedInFull,
			play: reviewAlone,
		},
		{
			name: "topic choice open",
			payload: offering,
			play: inThePanel(),
		},
		{
			name: "topic chosen, its mark above the task",
			payload: withTopicChoice(handed, { chosen: handed.task.topic }),
			play: inThePanel(),
		},
		// The button names the topic chosen with the choice shut: the topic whose
		// name is the longest in the card's language, so that the measure of the
		// layout sees the longest name the button holds.
		{
			name: "topic chosen, the choice shut",
			payload: withTopicChoice(handed, { chosen: longestTopicIn(language) }),
		},
		{
			name: "topic choice, a child in grade 1",
			payload: withTopicChoice({
				...handed,
				child: { ...handed.child, grade: 1 },
			}),
			play: inThePanel(),
		},
		{
			name: "topic chosen, the next task asked for",
			payload: offering,
			answers: () => Promise.resolve(topicSaved("time.clocks")),
			play: inThePanel(topicPicked),
		},
		{
			name: "topic choice, not saved",
			payload: offering,
			answers: () => Promise.resolve(failure),
			play: inThePanel(topicPicked),
		},
		{
			name: "topic choice, a chat that opens no links",
			payload: offering,
			links: "none",
			play: inThePanel(),
		},
		{
			name: "topic choice, a group the chat did not open",
			payload: offering,
			links: "refuse",
			play: inThePanel(groupPressed),
		},
		{
			name: "selected (checking)",
			payload: handed,
			answers: () => never,
			play: option("B"),
		},
		{ name: "hint", payload: handed, play: button(0) },
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
			name: "I don't know, said in the chat",
			payload: handed,
			answers: service({
				choice: "?",
				trap: null,
				rating: { before: 1502, after: 1488 },
				already_answered: true,
			}),
			play: option("B"),
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
		{
			name: "another task asked, the card done with",
			payload: handed,
			play: button(1),
		},
		{
			name: "another task asked once the answer is in, the card done with",
			payload: handed,
			answers: service(),
			play: inTurn(option("B"), onceAnswered(button(0))),
		},
		{
			name: "another task, the ask not sent",
			payload: handed,
			refuseMessages: true,
			play: button(1),
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
			name: "a task written, its course ticked off as it comes",
			payload: comingFor(handed),
			answers: (tool, before) => {
				if (tool !== "read_task") {
					return service()(tool);
				}
				return Promise.resolve(
					before === 0
						? writing()
						: { content: [], structuredContent: { ...handed } },
				);
			},
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
		{
			name: "progress, a chat that opens no links",
			payload: moving,
			links: "none",
			play: everyFold,
		},
		{
			name: "progress, a page the chat did not open",
			payload: longProgress,
			links: "refuse",
			play: inTurn(everyFold, linkPressed("divisibility-and-remainders")),
		},
		{
			name: "progress in the trial series, a mistake made twice",
			payload: trialMistake,
			play: everyFold,
		},
		{
			name: "progress, a review with every part",
			payload: reviewedInFull,
			play: everyFold,
		},
		{
			name: "progress, too early to judge",
			payload: tooEarly,
			play: everyFold,
		},
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
		{
			name: "profile, a family in the United States",
			payload: placed(profileRead, "US", "US-TX"),
		},
		{
			name: "profile, the longest name of a country",
			payload: placed(profileRead, "GS", null),
		},
		{ name: "profile form", payload: standing, play: inTheForm() },
		{
			name: "profile form, a family in the United States",
			payload: placed(standing, "US", "US-TX"),
			play: inTheForm(),
		},
		{
			name: "profile form, the longest name of a country",
			payload: placed(standing, "GS", null),
			play: inTheForm(),
		},
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

// rowCut is a task with its row of posts drawn cut short, its middle left
// out, as a task is drawn whose posts the child counts: the count cannot be
// read off the drawing before the child answers.
function rowCut(payload: typeof fence): typeof fence {
	return {
		...payload,
		task: { ...payload.task, drawing: "|--3--|--3-- ... --3--|\n" },
	};
}

// reviewAlone leaves the review the one section of the progress open, each
// section opened or folded by its title. The review is known by the parts it
// holds, which a folded section keeps drawn, because the titles' words change
// with the language.
function reviewAlone(card: Document) {
	for (const section of card.querySelectorAll(".mt-fold")) {
		const title = section.querySelector<HTMLElement>(".mt-fold-button");
		const review = section.querySelector(".mt-review-part") !== null;
		if ((title?.getAttribute("aria-expanded") === "true") !== review) {
			title?.click();
		}
	}
}

// button presses the card's button at place: 0 the hint, 1 another task; once
// the answer is in, 0 the next task. They are found by place because their
// words change with the language.
function button(place: number) {
	return (card: Document) => {
		card.querySelectorAll<HTMLElement>(".mt-btns .mt-btn")[place]?.click();
	};
}

// inThePanel opens the choice of the topic, and does each step in it in turn,
// each once the card has taken the one before in. The panel opens at once, so
// that the card has changed by the time it is looked at again.
function inThePanel(...steps: ((card: Document) => void)[]) {
	return (card: Document) => {
		card.querySelector<HTMLElement>(".mt-topic-button")?.click();
		inTurn(...steps)(card);
	};
}

// topicPicked chooses the first topic of the panel not chosen already.
function topicPicked(card: Document) {
	card
		.querySelector<HTMLElement>(
			'.mt-topic-rows .mt-topic-option[aria-pressed="false"]',
		)
		?.click();
}

// groupPressed presses the link of the panel's first group of topics.
function groupPressed(card: Document) {
	card.querySelector<HTMLElement>(".mt-topic-row-name a.mt-link")?.click();
}

// topLine presses the line at the top of the card.
function topLine(card: Document) {
	card.querySelector<HTMLElement>(".mt-bar")?.click();
}

// pages are the pages of the catalog's topics on the site, by the topic's id:
// each its slug, and whether it is published.
const pages: ReadonlyMap<string, { slug: string; site_page: boolean }> =
	new Map(
		topics.map((topic) => [
			topic.id,
			{ slug: topic.slug, site_page: topic.site_page },
		]),
	);

// rankCount is how many ranks there are, as the service counts them.
const rankCount = standing.overall.ranks;

// longRank is the overall rank of the long progress: the last of the
// narrowest run of grades marked under the course, whose label is drawn bold
// where it has the least room.
const longRank = standing.overall.grades.reduce(
	(narrowest, run) => {
		const width = run.last_rank - run.first_rank;
		return width < narrowest.width ? { width, last: run.last_rank } : narrowest;
	},
	{ width: Number.POSITIVE_INFINITY, last: 0 },
).last;

// longProgress is the progress at every limit a card has to fit at its
// narrowest: a pseudonym as long as a profile allows, every topic of the
// catalog listed — every rank among them, each step of the ramp, and topics
// not met yet — the rank reached in the narrowest run of grades, a review
// naming as many topics as it can, by the names that run longest across the
// languages, whatever the topics above say of them, with every reason they can
// have at once and the longest advice, as many interests, as long, as a
// profile holds, and the profile's file named at length beside other files
// that hold one. It is a card's limits, not a child's progress.
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
	].map((topic, at) => ({
		...rankedAt(topic, at),
		...pages.get(topic),
	})),
	overall: {
		...standing.overall,
		rating: ratingAt(longRank, 81),
		rank: longRank,
		share: 81,
		change: {
			last_task: {
				rating: ratingAt(longRank, 75),
				rank: longRank,
				share: 75,
				moved: "forward",
			},
			week: {
				rating: ratingAt(longRank - 1, 91),
				rank: longRank - 1,
				share: 91,
				moved: "rank_up",
			},
		},
	},
	skipped: 8,
	review: {
		strong: ["games.strategy", "logic.sets", "algorithms.weighing_pouring"].map(
			(topic) => ({ topic, reasons: ["mastered", "high", "rose"] }),
		),
		develop: [
			{
				topic: "ratio.sharing",
				reasons: ["low", "hints", "trap", "fell"],
				trap: "best_case_not_worst",
				moving: true,
			},
			{
				topic: "parity.alternation",
				reasons: ["low", "hints", "trap", "fell"],
				trap: "number_from_text",
				moving: true,
			},
			{ topic: "geometry.grid", reasons: ["low", "failures"] },
		],
		early: [
			"pigeonhole.basic",
			"number.divisibility",
			"logic.knights_liars",
			"arithmetic.tricks",
		],
		steps: [
			{ kind: "trap", topic: "ratio.sharing", trap: "best_case_not_worst" },
			{
				kind: "trap",
				topic: "parity.alternation",
				trap: "number_from_text",
			},
			{ kind: "practice", topic: "geometry.grid" },
			{ kind: "trap", trap: "ratio_total_confusion" },
			{ kind: "begin", topic: "number.divisibility", base: "games.strategy" },
		],
	},
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

// ratingAt is a rating the service shows at rank, share of the way through
// it: the rank's floor, the first rank's counted from a step below the
// second's, and as much of a step of 166 points again as the share says.
function ratingAt(rank: number, share: number): number {
	return 1168 + (rank - 1) * 166 + Math.floor(share * 1.66);
}

// rankedAt is the topic at place at of the long progress: the first fifteen
// each a rank in turn, from the first to the last and again, the overall
// rank ahead of, even with or behind them, and the last two not met yet.
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
	if (rank > longRank) {
		compared = "ahead";
	} else if (rank < longRank) {
		compared = "behind";
	}
	return {
		topic,
		rating: ratingAt(rank, share),
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

// trialMistake is the progress of the trial series with a mistake made in two
// of its three answers: no review yet, and the mistake in a section of its own.
const trialMistake = {
	...inTrial,
	topics: inTrial.topics.map((topic) =>
		topic.topic === "time.clocks" ? { ...topic, correct: 0 } : topic,
	),
	recent: inTrial.recent.map((entry) =>
		entry.topic === "time.clocks" ? { ...entry, correct: false } : entry,
	),
	mistakes: [{ trap: "off_by_one", times: 2 }],
};

// reviewedInFull is how the ranks moved, with a review of every part a review
// has: Ordering strong for each reason a topic can be, Enumeration to develop
// for the mistake it keeps making, Parity and alternation for standing low and
// falling, Gaps and boundaries for the hint, its last answer right; Clocks met
// too few times to judge; and the steps of a trap, of a rhythm, of the tasks
// tried without the hint, the mistake that repeats as often as the one
// advised, for every topic, and Knights and liars to begin, not met yet and
// built on Ordering.
const reviewedInFull = {
	...moving,
	topics: [
		...moving.topics,
		{
			topic: "time.clocks",
			rating: 1400,
			rank: 2,
			share: 40,
			compared: "behind",
			answers: 2,
			correct: 1,
			mastered: false,
			skipped: 0,
		},
		{
			topic: "logic.knights_liars",
			rating: null,
			rank: null,
			share: null,
			compared: null,
			answers: 0,
			correct: 0,
			mastered: false,
			skipped: 0,
			...pages.get("logic.knights_liars"),
		},
	],
	mistakes: [
		{ trap: "missed_case", times: 3 },
		{ trap: "double_count", times: 3 },
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
			{ topic: "counting.gaps", reasons: ["hints"], moving: true },
		],
		early: ["time.clocks"],
		steps: [
			{
				kind: "trap",
				topic: "combinatorics.enumeration",
				trap: "missed_case",
			},
			{ kind: "rhythm", topic: "parity.alternation" },
			{ kind: "unaided", topic: "counting.gaps" },
			{ kind: "trap", trap: "double_count" },
			{ kind: "begin", topic: "logic.knights_liars", base: "logic.ordering" },
		],
	},
};

// tooEarly is the progress just after the trial series: every topic met too
// few times to judge — fewer answers than mastering one takes, and none
// mastered —, and the mistake that repeats most advised for every topic.
const tooEarly = {
	...standing,
	topics: standing.topics.map((topic) => ({
		...topic,
		answers: Math.min(topic.answers, 2),
		correct: Math.min(topic.correct, 1),
		mastered: false,
	})),
	review: {
		strong: [],
		develop: [],
		early: standing.topics
			.filter((topic) => topic.answers > 0)
			.map((topic) => topic.topic),
		steps: [{ kind: "trap", trap: "missed_case" }],
	},
};

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

// onceAnswered does step once the card shows the result of the answer pressed
// before it, which the service sends back a moment later, and the buttons
// under the task with it. It gives up after ten seconds, the scene left as
// the answer leaves it.
function onceAnswered(step: (card: Document) => void) {
	return (card: Document) => {
		const tried = (left: number) => {
			if (card.querySelector(".mt-verdict-line") !== null) {
				step(card);
			} else if (left > 0) {
				setTimeout(() => tried(left - 1), 50);
			}
		};
		tried(200);
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

// linkPressed presses the link to the page whose slug is slug, as a person
// does; the first of them, where the page is linked twice.
function linkPressed(slug: string) {
	return (card: Document) => {
		for (const link of card.querySelectorAll<HTMLAnchorElement>("a.mt-link")) {
			if (link.getAttribute("href")?.includes(`/topics/${slug}/`)) {
				link.click();
				return;
			}
		}
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

// placed is a payload whose profile says where the family lives: the country,
// by its code — South Georgia and the South Sandwich Islands, GS, has the
// longest name of any in most languages — and the state, or none.
function placed<T extends { profile: object }>(
	payload: T,
	country: string,
	region: string | null,
): T {
	return { ...payload, profile: { ...payload.profile, country, region } };
}

// inFrench chooses French for the lessons, on the one list of the form that
// offers languages: the form's lists are told apart by what they offer, since
// their labels are in whatever language the card speaks. An option's value is
// read from the option, since one whose value is its text has no attribute.
function inFrench(card: Document) {
	const language = [
		...card.querySelectorAll<HTMLSelectElement>(".mt-form select"),
	].find((list) => [...list.options].some((option) => option.value === "fr"));
	if (language !== undefined) {
		language.value = "fr";
		language.dispatchEvent(new Event("change", { bubbles: true }));
	}
}

// saved presses the form's button that saves it.
function saved(card: Document) {
	card.querySelector<HTMLElement>(".mt-form .mt-btn-primary")?.click();
}
