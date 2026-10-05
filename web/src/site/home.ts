import { z } from "zod";
import { type Choice, dontKnow, type Letter, letters } from "../widget/choices";
import type { AnswerResult, HandedTask } from "../widget/payload";
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
 * grade its task is set in, its drawing — numbers and lines, which read alike
 * in every language — its five options and the right one, the catalog's trap
 * behind each wrong option, the wrong option the lesson's steps pick, and the
 * rating in the topic before an answer, after a wrong one and after a right
 * one. And the three traps it names as examples of what a wrong option is
 * tied to.
 */
export const homeFile = z.object({
	card: z.object({
		topic: z.string(),
		grade: z.number().int(),
		drawing: z.string(),
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
});

type HomeFile = z.infer<typeof homeFile>;

/** HomeCard is the card of the home page's lesson, but for its words. */
export type HomeCard = HomeFile["card"];

/**
 * Home is what the home page draws from the site's data: the card of its
 * lesson, and the three traps it names as examples, in the order it names
 * them.
 */
export type Home = {
	readonly card: HomeCard;
	readonly traps: readonly [string, string, string];
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

/**
 * readHome reads what the home page takes from the site's data, beside the
 * catalog the card speaks of. Every wrong option has a trap of the catalog
 * behind it, as every option of a task the service hands out does, and the
 * right one has none: a card that leaves a wrong option without one, names
 * one the catalog lacks, or gives the right option one is refused. So is a
 * card the catalog would not have set or the widget could not draw, one whose
 * steps pick the right option, and a trap named as an example that the catalog
 * does not have.
 */
export function readHome(catalog: CardCatalog, file: HomeFile): Home {
	const { card } = file;
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
	return { card, traps: file.traps };
}

/**
 * homeTaskOf is the lesson's task with the words said, as the widget reads a
 * task handed out: the card of the steps before an answer.
 */
export function homeTaskOf(card: HomeCard, said: HomeWords): HandedTask {
	return handedOf(factsOf(card), said, homeTask, homeCard);
}

/**
 * homeAnswerOf is the lesson's task with the words said, answered with the
 * option the steps pick, as the widget reads a wrong answer recorded.
 */
export function homeAnswerOf(card: HomeCard, said: HomeWords): WrongAnswer {
	const words: CardWords = { ...said, trap: trapWords(said, card.choice) };
	return answerOf(factsOf(card), words, homeTask, homeCard);
}

/**
 * HomeResults are what the service would record of every answer the card on
 * the first screen can be given, by the choice: each option, and "I don't
 * know".
 */
export type HomeResults = Readonly<Record<Choice, AnswerResult>>;

/**
 * homeResultsOf is what the service would record of every answer to the
 * lesson's task, with the words said, as the widget reads each: the trap
 * behind a wrong option and its words, the solution, and the rating after a
 * right answer or a wrong one. "I don't know" is a wrong answer that names no
 * trap.
 */
export function homeResultsOf(card: HomeCard, said: HomeWords): HomeResults {
	const resultOf = (choice: Choice): AnswerResult => {
		const right = choice === card.correct;
		const trap =
			right || choice === dontKnow
				? null
				: {
						id: card.traps[choice] ?? "",
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
	return {
		A: resultOf("A"),
		B: resultOf("B"),
		C: resultOf("C"),
		D: resultOf("D"),
		E: resultOf("E"),
		[dontKnow]: resultOf(dontKnow),
	};
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
		drawing: card.drawing,
		options: card.options,
		choice: card.choice,
		correct: card.correct,
		trap: card.traps[card.choice] ?? "",
		rating: { before: card.rating.before, after: card.rating.wrong },
	};
}
