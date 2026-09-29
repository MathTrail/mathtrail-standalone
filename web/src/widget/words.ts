import { createContext } from "preact";
import { useContext } from "preact/hooks";
import type english from "../../locales/en.json";
import { chooseLocale } from "../i18n/lookup";
import { type Dictionary, openWords, type Words } from "../i18n/words";

/**
 * Key names a text of the widget: a key of its English words, which the words
 * of every other language have too.
 */
export type Key = keyof typeof english;

/**
 * dictionaries are the widget's words in every language it speaks, by the tag
 * their file is named after. The page carries all of them: a card can load
 * nothing, so a language it does not carry is one it can never speak.
 */
export const dictionaries: ReadonlyMap<string, Dictionary> = new Map(
	Object.entries(
		import.meta.glob<Dictionary>("../../locales/*.json", {
			eager: true,
			import: "default",
		}),
	).map(([file, words]) => [tagOf(file), words]),
);

const spoken: ReadonlySet<string> = new Set(dictionaries.keys());

// tagOf is the tag a dictionary's file is named after: ru for locales/ru.json.
function tagOf(file: string): string {
	return file.slice(file.lastIndexOf("/") + 1, -".json".length);
}

/**
 * cardWords are the words a card speaks: in the language the parent chose for
 * the cards when the widget has words for it, in the host's otherwise, and in
 * English when it has words for neither.
 */
export function cardWords(
	chosen: string | undefined,
	host: string | undefined,
): Words<Key> {
	return openWords<Key>(chooseLocale([chosen, host], spoken), dictionaries);
}

/**
 * languageChosenIn is the language the parent chose for the cards, as a
 * payload carries it — with the child's details on a task card, with the
 * profile on the progress and the profile — or undefined when they chose none
 * and the cards follow the chat's.
 */
export function languageChosenIn(payload: unknown): string | undefined {
	for (const holder of ["child", "profile"]) {
		const language = fieldOf(fieldOf(payload, holder), "ui_language");
		if (typeof language === "string" && language !== "") {
			return language;
		}
	}
	return undefined;
}

// fieldOf is a field of an object, or undefined when value is no object or has
// no such field of its own.
function fieldOf(value: unknown, name: string): unknown {
	if (
		typeof value !== "object" ||
		value === null ||
		!Object.hasOwn(value, name)
	) {
		return undefined;
	}
	return (value as Record<string, unknown>)[name];
}

/**
 * ratingText is a rating as chess writes one — 1573, with no separator between
 * the thousands — in the digits of the words' language. Every other number is
 * written the way its language writes it, a separator and all.
 */
export function ratingText(words: Words<Key>, rating: number): string {
	return new Intl.NumberFormat(words.locale, { useGrouping: false }).format(
		rating,
	);
}

/**
 * WordsContext hands a card's words to everything drawn inside it. A component
 * drawn outside any card speaks English.
 */
export const WordsContext = createContext<Words<Key>>(
	cardWords(undefined, undefined),
);

/** useWords are the words of the card a component is drawn in. */
export function useWords(): Words<Key> {
	return useContext(WordsContext);
}
