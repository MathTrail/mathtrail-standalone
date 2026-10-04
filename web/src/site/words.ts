import { createContext } from "preact";
import { useContext } from "preact/hooks";
import { fallbackLocale } from "../i18n/lookup";
import {
	type Dictionary,
	dictionariesByTag,
	openWords,
	type Words,
} from "../i18n/words";
import type english from "./locales/en.json";

/**
 * SiteKey names a text of the site: a key of its English words, which the
 * words of every other language it speaks have too. The site's words are its
 * own — the widget's dictionaries are a card's, and the site says other things.
 */
export type SiteKey = keyof typeof english;

/**
 * siteDictionaries are the site's words in every language it speaks, by the
 * tag their file is named after.
 */
export const siteDictionaries: ReadonlyMap<string, Dictionary> =
	dictionariesByTag(
		import.meta.glob<Dictionary>("./locales/*.json", {
			eager: true,
			import: "default",
		}),
	);

/**
 * SiteWords are the words of the page being drawn. Every page is drawn inside
 * a provider of its own language's words; English is what a component drawn
 * outside one speaks.
 */
export const SiteWords = createContext<Words<SiteKey>>(
	openWords(fallbackLocale, siteDictionaries),
);

/** useSiteWords are the words of the page being drawn. */
export function useSiteWords(): Words<SiteKey> {
	return useContext(SiteWords);
}

/**
 * groupKey is the site's word for the name of a group of topics, which the
 * page of the topics says, and so does the page of every topic in the group.
 */
export function groupKey(group: { readonly id: string }): SiteKey {
	return `group.${group.id}` as SiteKey;
}

/**
 * gradesText is how words say a span of grades, "Grades 3–6", its numbers
 * written the way their language writes numbers.
 */
export function gradesText(
	words: Words<SiteKey>,
	grades: readonly [number, number],
): string {
	return words.text("topics.grades", {
		range: gradesRange(words.locale, grades),
	});
}

/**
 * gradesRange is a span of grades as its numbers alone, "3–6", written the
 * way the language of locale writes numbers: for a sentence that says the
 * grades in words of its own.
 */
export function gradesRange(
	locale: string,
	[first, last]: readonly [number, number],
): string {
	const numbers = new Intl.NumberFormat(locale);
	return `${numbers.format(first)}–${numbers.format(last)}`;
}
