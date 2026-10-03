import english from "../../locales/en.json";
import type { Words } from "../i18n/words";
import { isKey, type Key } from "./words";

/**
 * catalogSkills are the skills a parent can leave out of the tasks, by their
 * ids, in the order of the catalog: every skill the card has a name for, its
 * words kept in that order.
 */
export const catalogSkills: readonly string[] = Object.keys(english)
	.filter((key) => key.startsWith("skill."))
	.map((key) => key.slice("skill.".length));

/**
 * topicName is what the card calls a topic of the catalog, in its language;
 * a topic the card has no words for is called by its id.
 */
export function topicName(words: Words<Key>, id: string): string {
	return nameOf(words, `topic.${id}`, id);
}

/**
 * skillName is what the card calls a skill of the catalog, in its language;
 * a skill the card has no words for is called by its id.
 */
export function skillName(words: Words<Key>, id: string): string {
	return nameOf(words, `skill.${id}`, id);
}

/**
 * trapName is what the card calls a mistake of the catalog's, in plain words
 * of its language; a mistake the card has no words for is called by its id.
 */
export function trapName(words: Words<Key>, id: string): string {
	return nameOf(words, `trap.${id}`, id);
}

/**
 * rankCount is how many ranks there are, one name for each in the
 * dictionaries. A topic's course is drawn out of them while the trial series
 * runs, before an overall rating says how many there are.
 */
export const rankCount = 11;

/**
 * rankName is what the card calls a rank, in its language; a rank the card
 * has no words for is called by its number.
 */
export function rankName(words: Words<Key>, rank: number): string {
	return nameOf(words, `rank.${rank}`, String(rank));
}

// nameOf is the text of key, or otherwise when the words have none.
function nameOf(words: Words<Key>, key: string, otherwise: string): string {
	return isKey(key) ? words.text(key) : otherwise;
}

/**
 * listed is a list of names as the card's language writes one, set apart by
 * its own marks: "Space, animals, football".
 */
export function listed(words: Words<Key>, names: readonly string[]): string {
	if (typeof Intl.ListFormat !== "function") {
		// A platform with no list formats of its own gets the commas alone.
		return names.join(", ");
	}
	return new Intl.ListFormat(words.locale, {
		type: "conjunction",
		style: "narrow",
	}).format(names);
}

/**
 * languageName is what the card's language calls the language tag names,
 * written to begin a line; a tag it cannot read is shown as it is.
 */
export function languageName(words: Words<Key>, tag: string): string {
	let name: string;
	try {
		name =
			new Intl.DisplayNames([words.locale], { type: "language" }).of(tag) ??
			tag;
	} catch {
		// A tag the platform cannot read names no language it can say.
		return tag;
	}
	const [first = "", ...rest] = name;
	return first.toLocaleUpperCase(words.locale) + rest.join("");
}
