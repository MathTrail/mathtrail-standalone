import { createContext } from "preact";
import { useContext } from "preact/hooks";
import type english from "../../locales/en.json";
import type { Words } from "../i18n/words";

/**
 * Key names a text of the widget: a key of its English words, which the words
 * of every other language have too.
 */
export type Key = keyof typeof english;

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
 * WordsContext hands a card's words to everything drawn inside it. The place
 * a card is drawn in gives them — the widget's page in the language its
 * payload names, a page of the site in its own — and a component drawn
 * outside any card has none: that is a mistake to mend, not a card to show in
 * English.
 */
export const WordsContext = createContext<Words<Key> | undefined>(undefined);

/** useWords are the words of the card a component is drawn in. */
export function useWords(): Words<Key> {
	const words = useContext(WordsContext);
	if (words === undefined) {
		throw new Error(
			"widget: a component of a card was drawn outside any card, with no words to speak",
		);
	}
	return words;
}

/**
 * countText is a whole number as the words' language writes one, in its
 * digits.
 */
export function countText(words: Words<Key>, count: number): string {
	return new Intl.NumberFormat(words.locale).format(count);
}
