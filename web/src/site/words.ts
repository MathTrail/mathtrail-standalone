import { createContext } from "preact";
import { useContext } from "preact/hooks";
import { fallbackLocale, listFormat } from "../i18n/lookup";
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

/**
 * listMarks are the marks between names, as the language of locale sets a
 * list of them apart: ", " in English, " y " before the last name in Spanish,
 * "、" in Japanese and Chinese. The mark at a place stands before the name at
 * that place, and the first name has none. The names are given as they are
 * written, since a language may change its mark by the name after it, as
 * Spanish writes "e" rather than "y" before Historias.
 */
export function listMarks(locale: string, names: readonly string[]): string[] {
	const marks: string[] = [];
	let mark = "";
	for (const part of listFormat(locale).formatToParts(names)) {
		if (part.type === "literal") {
			mark += part.value;
		} else {
			marks.push(mark);
			mark = "";
		}
	}
	return marks;
}

// unspacedEnd and unspacedStart find writing that sets no space between its
// words where it meets other text: the Han of Chinese and Japanese, Japanese
// kana and their full-width marks, which carry their own room.
const unspacedEnd = /[\p{sc=Han}\u3000-\u30ff\uff01-\uff60]$/u;
const unspacedStart = /^[\p{sc=Han}\u3000-\u30ff\uff01-\uff60]/u;

/**
 * wordSpace is the space between two texts a page sets side by side, such as
 * a label and what it labels or a count and its noun, read as they are
 * written, the signs of an emphasis at their edges aside: none where they
 * meet in Chinese or Japanese writing, «12問», «适用场合：遇到…», and a space
 * everywhere else, "12 tasks".
 */
export function wordSpace(before: string, after: string): string {
	return unspacedEnd.test(emphasisLeftOffEnd(before)) ||
		unspacedStart.test(emphasisLeftOffStart(after))
		? ""
		: " ";
}

// emphasisLeftOffEnd is a text with the signs of an emphasis at its end left
// off, walked back from the end rather than matched: a pattern anchored at the
// end tries each sign of a run as where the run starts, and reads a long run
// that stops short of the end over and over.
function emphasisLeftOffEnd(text: string): string {
	let end = text.length;
	while (end > 0 && emphasisSign(text[end - 1])) {
		end -= 1;
	}
	return text.slice(0, end);
}

// emphasisLeftOffStart is a text with the signs of an emphasis at its start
// left off.
function emphasisLeftOffStart(text: string): string {
	let start = 0;
	while (start < text.length && emphasisSign(text[start])) {
		start += 1;
	}
	return text.slice(start);
}

// emphasisSign reports whether a character is one Markdown marks an emphasis with.
function emphasisSign(character: string | undefined): boolean {
	return character === "*" || character === "_";
}
