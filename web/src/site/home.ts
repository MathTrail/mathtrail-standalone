import { z } from "zod";
import { type Letter, letters } from "../widget/choices";
import type { AnswerResult, HandedTask, TopicChoice } from "../widget/payload";
import { pictureFormat } from "../widget/picture";
import { sectionAddress } from "./addresses";
import { frontPage } from "./content";
import {
	answerOf,
	type CardCatalog,
	type CardFacts,
	type CardWords,
	checkCard,
	handedOf,
	recordedOf,
	type TaskWords,
	type WrongAnswer,
} from "./taskcard";

/**
 * homeFile is the shape of what the home page takes from the site's data, as
 * far as no language changes it. The card of its lesson: the topic and the
 * grade its task is set in, its picture — which holds no words, and reads
 * alike in every language — its five options and the right one, the catalog's trap
 * behind each wrong option, the wrong option the lesson's steps pick, and the
 * rating in the topic before an answer, after a wrong one and after a right
 * one. The three traps it names as examples of what a wrong option is tied
 * to. How the card's task was picked: the child's topics, each with how far
 * the child has come in it, the chance of a right answer and the corridor the
 * rule keeps that chance in. The options of the task a chat writes alone, and
 * those of them that are right, more than one. The languages a task can be
 * written in, a few of many. And the topic whose bases show what builds on
 * what.
 */
export const homeFile = z.object({
	card: z.object({
		topic: z.string(),
		grade: z.number().int(),
		picture: pictureFormat,
		options: z.record(z.string(), z.string()),
		correct: z.string(),
		traps: z.record(z.string(), z.string()),
		choice: z.custom<Letter>(
			(value) => letters.includes(value as Letter),
			"the option the steps pick is a letter A to E",
		),
		rating: z.object({
			before: z.number(),
			wrong: z.number(),
			right: z.number(),
		}),
	}),
	traps: z.tuple([z.string(), z.string(), z.string()]),
	pick: z.object({
		skills: z.array(z.object({ topic: z.string(), share: z.number() })).min(1),
		chance: z.number(),
		corridor: z.object({ low: z.number(), high: z.number() }),
	}),
	alone: z.object({
		options: z.record(z.string(), z.string()),
		right: z.array(z.string()),
	}),
	languages: z.array(z.string()).min(1),
	map: z.string(),
});

type HomeFile = z.infer<typeof homeFile>;

/** HomeCard is the card of the home page's lesson, but for its words. */
export type HomeCard = HomeFile["card"];

/**
 * HomePick is how the task on the card was picked: the child's topics, each
 * with how far the child has come in it as a share of the way, the chance of
 * a right answer, and the corridor the rule keeps that chance in.
 */
export type HomePick = HomeFile["pick"];

/**
 * HomeAlone is the task a chat writes when asked alone: its options by their
 * letters, and the letters of those that are right.
 */
export type HomeAlone = HomeFile["alone"];

/**
 * HomeMap is a topic of the catalog and the two topics it builds on, in the
 * catalog's order: what the home page shows of the map of the topics.
 */
export type HomeMap = {
	readonly topic: string;
	readonly bases: readonly [string, string];
};

/**
 * Home is what the home page draws from the site's data: the card of its
 * lesson; the three traps it names as examples, in the order it names them;
 * how the card's task was picked; the task a chat writes alone; the languages
 * it names, by their tags; and a topic with what it builds on.
 */
export type Home = {
	readonly card: HomeCard;
	readonly traps: readonly [string, string, string];
	readonly pick: HomePick;
	readonly alone: HomeAlone;
	readonly languages: readonly string[];
	readonly map: HomeMap;
};

/**
 * HomeWords are the words of the lesson's task in the page's language: whose
 * card it is, the task, its hint, the trap behind each wrong option by its
 * letter, and the solution.
 */
export type HomeWords = TaskWords & {
	readonly traps: Readonly<Record<string, string>>;
	readonly solution: string;
};

/**
 * lessonSection and connectSection are the ids of the home page's sections on
 * how a lesson goes and on connecting MathTrail, which the header and the
 * other pages lead to.
 */
export const lessonSection = "lesson";
export const connectSection = "connect";

/**
 * connectAddress is where the home page of locale tells how to connect
 * MathTrail: where every button that asks a reader to add it leads.
 */
export function connectAddress(locale: string): string {
	return sectionAddress(locale, frontPage, connectSection);
}

// homeCard is the card on the home page as what the build says of it names it,
// and homeTask the id of the task on it, which an answer recorded for it names.
const homeCard = "the card on the home page";
const homeTask = "site_home";

// coachChooses is the choice of the topic the home page's cards show, as a card
// after the trial series does: the coach chooses, and nothing is suggested.
// The page lets no topic be chosen, so the button is drawn locked.
const coachChooses: TopicChoice = { chosen: null, recommended: [] };

/**
 * readHome reads what the home page takes from the site's data, beside the
 * catalog the card speaks of. Every wrong option has a trap of the catalog
 * behind it, as every option of a task the service hands out does, and the
 * right one has none: a card that leaves a wrong option without one, names
 * one the catalog lacks, or gives the right option one is refused. So is a
 * card the catalog would not have set or the widget could not draw, one whose
 * steps pick the right option, and a trap named as an example that the catalog
 * does not have. So is a pick that leaves out the card's topic or names one
 * the catalog lacks, or whose chance falls outside its corridor; a task of the
 * chat alone with fewer than two right options; and a topic of the map that
 * does not build on two others.
 */
export function readHome(catalog: CardCatalog, file: HomeFile): Home {
	const { card } = file;
	checkTraps(catalog, card);
	const { before, wrong, right } = card.rating;
	if (!(right > before && before > wrong)) {
		throw new Error(
			`${homeCard} moves its rating from ${before} to ${wrong} on a wrong answer and to ${right} on a right one: a right answer raises it, and a wrong one lowers it`,
		);
	}
	for (const trap of file.traps) {
		if (!catalog.traps.some(({ id }) => id === trap)) {
			throw new Error(
				`the home page names ${trap} as an example, a trap the catalog does not have`,
			);
		}
	}
	checkCard(catalog, factsOf(card), homeCard);
	checkPick(catalog, file.pick, card.topic);
	checkAlone(file.alone);
	return {
		card,
		traps: file.traps,
		pick: file.pick,
		alone: file.alone,
		languages: file.languages,
		map: mapOf(catalog, file.map),
	};
}

// checkPick refuses a pick of the card's task that does not show the card's
// topic among the child's, names a topic the catalog does not have, gives a
// share or a chance that is no share, or puts the chance outside a corridor
// that runs the wrong way.
function checkPick(catalog: CardCatalog, pick: HomePick, picked: string): void {
	const share = (value: number) => value >= 0 && value <= 1;
	for (const { topic, share: shown } of pick.skills) {
		if (!catalog.topics.some(({ id }) => id === topic)) {
			throw new Error(
				`the home page shows the child's skill in ${topic}, a topic the catalog does not have`,
			);
		}
		if (!share(shown)) {
			throw new Error(
				`the home page fills ${shown} of the bar of ${topic}, which is no share of it`,
			);
		}
	}
	if (!pick.skills.some(({ topic }) => topic === picked)) {
		throw new Error(
			`the home page picks ${picked}, the topic of its card, and leaves it out of the child's skills`,
		);
	}
	const { low, high } = pick.corridor;
	if (!(share(low) && share(high) && low < high)) {
		throw new Error(
			`the home page's corridor runs from ${low} to ${high}: two shares, the lower first`,
		);
	}
	if (pick.chance < low || pick.chance > high) {
		throw new Error(
			`the home page picks its task at a chance of ${pick.chance}, outside the corridor from ${low} to ${high}`,
		);
	}
}

// checkAlone refuses the task a chat writes alone unless more than one of its
// options is right, which is what the page shows it for, and every right one
// is an option.
function checkAlone(alone: HomeAlone): void {
	if (alone.right.length < 2) {
		throw new Error(
			`the task of the chat alone has ${alone.right.length} right options, and the home page shows it for having more than one`,
		);
	}
	for (const letter of alone.right) {
		if (!Object.hasOwn(alone.options, letter)) {
			throw new Error(
				`the task of the chat alone has ${letter} right, which is no option of it`,
			);
		}
	}
}

// mapOf is topic with the two topics it builds on, in the catalog's order. A
// topic the catalog does not have, or one that builds on other than two, is
// refused: the page shows one base on either side of it.
function mapOf(catalog: CardCatalog, topic: string): HomeMap {
	const found = catalog.topics.find(({ id }) => id === topic);
	if (found === undefined) {
		throw new Error(
			`the home page shows what ${topic} builds on, a topic the catalog does not have`,
		);
	}
	const order = (id: string) =>
		catalog.topics.findIndex((other) => other.id === id);
	const bases = [...found.builds_on].sort((a, b) => order(a) - order(b));
	const [first, second] = bases;
	if (bases.length !== 2 || first === undefined || second === undefined) {
		throw new Error(
			`the home page shows ${topic} between two topics it builds on, and it builds on ${bases.length}`,
		);
	}
	return { topic, bases: [first, second] };
}

// checkTraps refuses card unless a trap of the catalog stands behind each of
// its wrong options and behind nothing else: neither the right option nor a
// letter that is no option.
function checkTraps(catalog: CardCatalog, card: HomeCard): void {
	for (const letter of letters) {
		const trap = card.traps[letter];
		if (letter === card.correct) {
			if (trap !== undefined) {
				throw new Error(
					`${homeCard} ties a trap to ${letter}, its right option`,
				);
			}
		} else if (trap === undefined) {
			throw new Error(
				`${homeCard} leaves its wrong option ${letter} without a trap`,
			);
		} else if (!catalog.traps.some(({ id }) => id === trap)) {
			throw new Error(
				`${homeCard} ties ${letter} to the trap ${trap}, which the catalog does not have`,
			);
		}
	}
	for (const letter of Object.keys(card.traps)) {
		if (!letters.includes(letter as Letter)) {
			throw new Error(
				`${homeCard} ties a trap to ${letter}, which is no option`,
			);
		}
	}
}

/**
 * homeTaskOf is the lesson's task with the words said, as the widget reads a
 * task handed out: the card of the steps before an answer, with the button of
 * the topic, the coach choosing.
 */
export function homeTaskOf(card: HomeCard, said: HomeWords): HandedTask {
	return handedOf(factsOf(card), said, homeTask, homeCard, coachChooses);
}

/**
 * homeAnswerOf is the lesson's task with the words said, answered with the
 * option the steps pick, as the widget reads a wrong answer recorded, with the
 * button of the topic, the coach choosing.
 */
export function homeAnswerOf(card: HomeCard, said: HomeWords): WrongAnswer {
	const words: CardWords = { ...said, trap: trapWords(said, card.choice) };
	return answerOf(factsOf(card), words, homeTask, homeCard, coachChooses);
}

/**
 * HomeResults are what the service would record of every answer the card on
 * the first screen can be given, by the option chosen.
 */
export type HomeResults = Readonly<Record<Letter, AnswerResult>>;

/**
 * homeResultsOf is what the service would record of every answer to the
 * lesson's task, with the words said, as the widget reads each: the trap
 * behind a wrong option and its words, the solution, and the rating after a
 * right answer or a wrong one.
 */
export function homeResultsOf(card: HomeCard, said: HomeWords): HomeResults {
	const resultOf = (choice: Letter): AnswerResult => {
		const right = choice === card.correct;
		const trap = right
			? null
			: {
					id: trapOf(card, choice),
					text: trapWords(said, choice),
					repeated: false,
				};
		return recordedOf(
			{
				task_id: homeTask,
				topic: card.topic,
				choice,
				correct: right,
				correct_answer: card.correct,
				trap,
				solution: said.solution,
				hint_used: false,
				rating: {
					before: card.rating.before,
					after: right ? card.rating.right : card.rating.wrong,
				},
				trial: null,
				already_answered: false,
			},
			homeCard,
		);
	};
	return Object.fromEntries(
		letters.map((letter) => [letter, resultOf(letter)]),
	) as HomeResults;
}

// trapOf is the trap of the catalog the card names behind the wrong option
// letter. readHome refuses a card that leaves one without, and so does this,
// rather than record an answer behind no trap.
function trapOf(card: HomeCard, letter: Letter): string {
	const trap = card.traps[letter];
	if (trap === undefined) {
		throw new Error(
			`${homeCard} leaves its wrong option ${letter} without a trap`,
		);
	}
	return trap;
}

// trapWords are the words of the trap behind the wrong option letter, as the
// page says them. A wrong option the page has no words for stops the build.
function trapWords(said: HomeWords, letter: string): string {
	const words = said.traps[letter];
	if (words === undefined) {
		throw new Error(
			`the home page says nothing of the trap behind its wrong option ${letter}`,
		);
	}
	return words;
}

// factsOf are the facts of the card, with the trap behind the option the steps
// pick.
function factsOf(card: HomeCard): CardFacts {
	return {
		topic: card.topic,
		grade: card.grade,
		picture: card.picture,
		options: card.options,
		choice: card.choice,
		correct: card.correct,
		trap: card.traps[card.choice] ?? "",
		rating: { before: card.rating.before, after: card.rating.wrong },
	};
}
