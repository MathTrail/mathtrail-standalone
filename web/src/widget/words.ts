import { createContext } from "preact";
import { useContext } from "preact/hooks";
import english from "../../locales/en.json";
import { chooseLocale } from "../i18n/lookup";
import { pseudoLocale, pseudoWords } from "../i18n/pseudo";
import {
	type Dictionary,
	dictionariesByTag,
	openWords,
	type Words,
} from "../i18n/words";

/**
 * Key names a text of the widget: a key of its English words, which the words
 * of every other language have too.
 */
export type Key = keyof typeof english;

// written are the dictionaries of the languages the widget is written in, by
// the tag their file is named after.
const written: ReadonlyMap<string, Dictionary> = dictionariesByTag(
	import.meta.glob<Dictionary>("../../locales/*.json", {
		eager: true,
		import: "default",
	}),
);

/**
 * dictionaries are the widget's words in every language it speaks, by the tag
 * their file is named after. The page carries all of them: a card can load
 * nothing, so a language it does not carry is one it can never speak. The
 * preview and the widget's tests speak the pseudo-language too — English
 * stretched as a longer language stretches it, to see a card hold longer
 * words — and a build for production leaves it out, since a host could name
 * it. It is the build's mode that decides, not the environment it runs in.
 */
export const dictionaries: ReadonlyMap<string, Dictionary> =
	import.meta.env.MODE === "production"
		? written
		: new Map([...written, [pseudoLocale, pseudoWords(english)]]);

const spoken: ReadonlySet<string> = new Set(dictionaries.keys());

/**
 * isKey says whether key names a text of the widget: one its English words
 * have, as the words of every other language do.
 */
export function isKey(key: string): key is Key {
	const english = dictionaries.get("en");
	return english !== undefined && Object.hasOwn(english, key);
}

/**
 * cardWords are the words a card speaks: in the language its payload names
 * when the widget has words for it, in the host's otherwise, and in English
 * when it has words for neither.
 */
export function cardWords(
	named: string | undefined,
	host: string | undefined,
): Words<Key> {
	return openWords<Key>(chooseLocale([named, host], spoken), dictionaries);
}

/**
 * languageIn is the language a card speaks, as its payload names it: the
 * lesson's, which a card of a task, or of a wait for one, carries as its
 * `language` — its task's or the awaited request's — and otherwise the one the
 * parent chose for the lessons, with the child's details or with the profile.
 * It is undefined when the payload names none, and the card follows the
 * host's. A task and the words around it are always in one language: buttons
 * in another would leave a child reading two at once. A card from before the
 * lesson's language travelled with it names none, and keeps the parent's
 * choice it was drawn in.
 */
export function languageIn(payload: unknown): string | undefined {
	const named = [
		fieldOf(payload, "language"),
		fieldOf(fieldOf(payload, "child"), "ui_language"),
		fieldOf(fieldOf(payload, "profile"), "ui_language"),
	];
	return named.find(
		(language): language is string =>
			typeof language === "string" && language !== "",
	);
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
 * percentText is a whole percent, from 0 to 100, as the words' language writes
 * one, in its digits and with its own sign.
 */
export function percentText(words: Words<Key>, percent: number): string {
	return new Intl.NumberFormat(words.locale, { style: "percent" }).format(
		percent / 100,
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
