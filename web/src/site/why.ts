import { z } from "zod";
import {
	type AnswerResult,
	type HandedTask,
	type Letter,
	letters,
	readAnswer,
	readHandedTask,
} from "../widget/payload";
import { type CatalogTopic, gradesOf } from "./topics";
import type { CatalogTrap } from "./traps";

// doi is what a work's DOI looks like: 10., the registrant's number, a slash,
// and the name the registrant gave the work.
const doi = /^10\.\d{4,9}\/\S+$/;

// slotName is what a slot of a page's text may be called: lowercase letters.
const slotName = /^[a-z]+$/;

// sourceFile is a work the page cites, as the site's data writes it: its
// authors, each as "Family, I.", its year, its title, where it was published,
// and its DOI, which is where the page links it.
const sourceFile = z.object({
	authors: z.array(z.string().min(1)).min(1),
	year: z.number().int(),
	title: z.string().min(1),
	journal: z.string().min(1),
	volume: z.string().min(1),
	issue: z.string().min(1),
	pages: z.string().min(1),
	doi: z.string().regex(doi),
});

/**
 * whyFile is the shape of what the page "Why" takes from the site's data, as
 * far as no language changes it: the card of a wrong answer it shows; the
 * topics it names as examples of what a problem is about, by the slot of its
 * words each fills; the works its findings come from, in the order it shows
 * them; the work behind what it says of children's apps; and every work it
 * cites, by an id of the site's own. The letters, the trap and the rest of the card are read again
 * by the widget's own readers, which refuse a card the widget could not draw.
 */
export const whyFile = z.object({
	card: z.object({
		topic: z.string(),
		grade: z.number().int(),
		drawing: z.string(),
		options: z.record(z.string(), z.string()),
		choice: z.string(),
		correct: z.string(),
		trap: z.string(),
		rating: z.object({ before: z.number(), after: z.number() }),
	}),
	topics: z.record(z.string().regex(slotName), z.string()),
	findings: z.array(z.string()).min(1),
	apps: z.string(),
	sources: z.record(z.string(), sourceFile),
});

type WhyFile = z.infer<typeof whyFile>;

/** Source is a work the page "Why" cites, under the id the site's data gives it. */
export type Source = z.infer<typeof sourceFile> & { readonly id: string };

/** WhyCard is the card of a wrong answer the page shows, but for its words. */
export type WhyCard = WhyFile["card"];

/**
 * Why is what the page "Why" draws from the site's data: the card, the topics
 * it names as examples by the slot each fills, the works its findings come
 * from in the order it shows them, the work behind what it says of children's
 * apps, and every work it cites, in the order it first cites them.
 */
export type Why = {
	readonly card: WhyCard;
	readonly topics: Readonly<Record<string, string>>;
	readonly findings: readonly Source[];
	readonly apps: Source;
	readonly sources: readonly Source[];
};

/**
 * CardWords are the words of the card on the page "Why", in the page's
 * language: whose card it is, the task, the trap behind the wrong option
 * picked, and the solution.
 */
export type CardWords = {
	readonly language: string;
	readonly child: string;
	readonly question: string;
	readonly trap: string;
	readonly solution: string;
};

/**
 * WrongAnswer is the card on the page "Why" as the widget reads one: the task
 * handed to the child, and the wrong answer the service recorded for it.
 */
export type WrongAnswer = {
	readonly handed: HandedTask;
	readonly result: AnswerResult;
};

/**
 * readWhy reads what the page "Why" takes from the site's data, beside the
 * catalog the card speaks of. A card of a topic or a trap the catalog does not
 * have is refused, and so is one set in a grade its topic is not taught at,
 * one that answers right, and one the widget could not draw. A topic named as
 * an example that the catalog does not have is refused: the page would show
 * its id. A finding of a work the data does not have is refused, and so is a
 * work no part of the page cites: it would be listed among the sources and
 * back nothing.
 */
export function readWhy(
	catalog: {
		readonly topics: readonly CatalogTopic[];
		readonly traps: readonly CatalogTrap[];
	},
	file: WhyFile,
): Why {
	checkCard(catalog, file.card);
	for (const id of Object.values(file.topics)) {
		if (!catalog.topics.some((topic) => topic.id === id)) {
			throw new Error(
				`the page Why names ${id} as an example, a topic the catalog does not have`,
			);
		}
	}
	answerOf(file.card, {
		language: "en",
		child: "Comet",
		question: "?",
		trap: "?",
		solution: "?",
	});
	const sourceOf = (id: string): Source => {
		const work = file.sources[id];
		if (work === undefined) {
			throw new Error(
				`the page Why cites ${id}, which the data's sources do not have`,
			);
		}
		return { ...work, id };
	};
	const cited = [...file.findings, file.apps];
	if (new Set(file.findings).size !== file.findings.length) {
		throw new Error("the page Why shows a finding twice");
	}
	for (const id of Object.keys(file.sources)) {
		if (!cited.includes(id)) {
			throw new Error(`the source ${id} is cited nowhere on the page Why`);
		}
	}
	return {
		card: file.card,
		topics: file.topics,
		findings: file.findings.map(sourceOf),
		apps: sourceOf(file.apps),
		sources: [...new Set(cited)].map(sourceOf),
	};
}

// checkCard refuses a card the catalog would not have set: of a topic or a
// trap it does not have, or in a grade the topic is not taught at. It refuses
// a right answer too, and "I don't know", which names no trap: neither is the
// card the page talks about.
function checkCard(
	catalog: {
		readonly topics: readonly CatalogTopic[];
		readonly traps: readonly CatalogTrap[];
	},
	card: WhyCard,
): void {
	const topic = catalog.topics.find(({ id }) => id === card.topic);
	if (topic === undefined) {
		throw new Error(
			`the card on the page Why is of ${card.topic}, a topic the catalog does not have`,
		);
	}
	const [first, last] = gradesOf(topic.grade_levels);
	if (card.grade < first || card.grade > last) {
		throw new Error(
			`the card on the page Why is set in grade ${card.grade}, which ${card.topic} is not taught in`,
		);
	}
	if (!catalog.traps.some(({ id }) => id === card.trap)) {
		throw new Error(
			`the card on the page Why names the trap ${card.trap}, which the catalog does not have`,
		);
	}
	if (!letters.includes(card.choice as Letter)) {
		throw new Error(
			`the card on the page Why picks ${card.choice}, and the page shows an option picked: a letter A to E`,
		);
	}
	if (card.choice === card.correct) {
		throw new Error(
			"the card on the page Why answers right, and the page shows a wrong answer",
		);
	}
}

// cardTask is the id of the task on the card, which the answer recorded for it
// names: no service handed it out, and nothing shows it.
const cardTask = "site_why";

/**
 * answerOf is card with the words said, as the widget reads one: the task
 * handed to the child, and the wrong answer recorded for it. The task has no
 * hint, which a card no longer shows once its answer is in. A card the
 * widget's own readers refuse stops the build.
 */
export function answerOf(card: WhyCard, said: CardWords): WrongAnswer {
	const handed = readHandedTask({
		screen: "task",
		child: { pseudonym: said.child, grade: card.grade, ui_language: null },
		task: {
			id: cardTask,
			topic: card.topic,
			language: said.language,
			question: said.question,
			drawing: card.drawing,
			options: card.options,
			hint: "",
		},
		language: said.language,
	});
	if (handed === undefined) {
		throw new Error(
			"the card on the page Why is no task the widget can draw: it needs the five options A to E",
		);
	}
	const told = readAnswer(
		{
			content: [],
			structuredContent: {
				screen: "result",
				result: {
					task_id: cardTask,
					topic: card.topic,
					choice: card.choice,
					correct: false,
					correct_answer: card.correct,
					trap: { id: card.trap, text: said.trap, repeated: false },
					solution: said.solution,
					hint_used: false,
					rating: card.rating,
					trial: null,
					already_answered: false,
				},
			},
		},
		cardTask,
	);
	if (told.kind !== "answered") {
		throw new Error(
			"the card on the page Why holds no answer the widget can draw: its choice and its right option are letters A to E",
		);
	}
	return { handed, result: told.result };
}
