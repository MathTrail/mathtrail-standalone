import { z } from "zod";
import { pictureFormat } from "../widget/picture";
import {
	type CardCatalog,
	type CardWords,
	answerOf as cardAnswerOf,
	checkCard,
	type WrongAnswer,
} from "./taskcard";

// doi is what a work's DOI looks like: 10., the registrant's number, a slash,
// and the name the registrant gave the work.
const doi = /^10\.\d{4,9}\/\S+$/;

// slotName is what a slot of a page's text may be called: lowercase letters.
const slotName = /^[a-z]+$/;

// language is a language as an element names it: a tag of BCP 47, such as de.
const language = /^[a-z]{2,3}(?:-[A-Za-z0-9]{2,8})*$/;

// sourceFile is a work the page cites, as the site's data writes it: its
// authors, each as "Family, I.", its year, its title, where it was published —
// with the language of the journal's name when that is not English, as the
// works' own titles all are — and its DOI, which is where the page links it.
const sourceFile = z.object({
	authors: z.array(z.string().min(1)).min(1),
	year: z.number().int(),
	title: z.string().min(1),
	journal: z.string().min(1),
	journal_language: z.string().regex(language).optional(),
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
		picture: pictureFormat,
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

// doiResolver is where a DOI is resolved to the work it names.
const doiResolver = "https://doi.org/";

/**
 * doiAddress is where source's DOI leads. A DOI may hold characters an address
 * reads otherwise — a # would end the path, a ? begin a query — so each part
 * of it is escaped, and the slashes between them kept.
 */
export function doiAddress(source: Pick<Source, "doi">): string {
	return doiResolver + source.doi.split("/").map(encodeURIComponent).join("/");
}

/** WhyCard is the card of a wrong answer the page shows, but for its words. */
export type WhyCard = WhyFile["card"];

/**
 * Why is what the page "Why" draws from the site's data: the card, the topics
 * it names as examples by the slot each fills, the works its findings come
 * from in the order it shows them, and the work behind what it says of
 * children's apps.
 */
export type Why = {
	readonly card: WhyCard;
	readonly topics: Readonly<Record<string, string>>;
	readonly findings: readonly Source[];
	readonly apps: Source;
};

export type { CardWords, WrongAnswer } from "./taskcard";

// whyCard is the card on the page as what the build says of it names it, and
// whyTask the id of the task on it, which the answer recorded for it names.
const whyCard = "the card on the page Why";
const whyTask = "site_why";

/**
 * readWhy reads what the page "Why" takes from the site's data, beside the
 * catalog the card speaks of. A card of a topic or a trap the catalog does not
 * have is refused, and so is one set in a grade its topic is not taught at,
 * one that answers right, and one the widget could not draw. A topic named as
 * an example that the catalog does not have is refused: the page would show
 * its id. A finding of a work the data does not have is refused, and so is a
 * work no part of the page cites: it would back nothing.
 */
export function readWhy(catalog: CardCatalog, file: WhyFile): Why {
	checkCard(catalog, file.card, whyCard);
	for (const id of Object.values(file.topics)) {
		if (!catalog.topics.some((topic) => topic.id === id)) {
			throw new Error(
				`the page Why names ${id} as an example, a topic the catalog does not have`,
			);
		}
	}
	const sourceOf = (id: string): Source => {
		const work = file.sources[id];
		if (work === undefined) {
			throw new Error(
				`the page Why cites ${id}, which the data's sources do not have`,
			);
		}
		return { ...work, id };
	};
	const cited = new Set([...file.findings, file.apps]);
	if (new Set(file.findings).size !== file.findings.length) {
		throw new Error("the page Why shows a finding twice");
	}
	for (const id of Object.keys(file.sources)) {
		if (!cited.has(id)) {
			throw new Error(`the source ${id} is cited nowhere on the page Why`);
		}
	}
	return {
		card: file.card,
		topics: file.topics,
		findings: file.findings.map(sourceOf),
		apps: sourceOf(file.apps),
	};
}

/**
 * answerOf is the card on the page with the words said, as the widget reads
 * one: the task handed to the child, and the wrong answer recorded for it.
 */
export function answerOf(card: WhyCard, said: CardWords): WrongAnswer {
	return cardAnswerOf(card, said, whyTask, whyCard);
}
