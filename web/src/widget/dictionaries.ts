import { chooseLocale } from "../i18n/lookup";
import { pseudoLocale, pseudoWords } from "../i18n/pseudo";
import {
	type Dictionary,
	dictionariesByTag,
	openWords,
	type Words,
} from "../i18n/words";
import type { Key } from "./words";

// written are the dictionaries of the languages the widget is written in, by
// the tag their file is named after.
const written: ReadonlyMap<string, Dictionary> = dictionariesByTag(
	import.meta.glob<Dictionary>("../../locales/*.json", {
		eager: true,
		import: "default",
	}),
);

// english are the widget's English words, the ones every other language is
// written from.
const english: Dictionary = written.get("en") ?? {};

/**
 * dictionaries are the widget's words in every language it speaks, by the tag
 * their file is named after. The widget's page carries all of them: a card
 * can load nothing, so a language it does not carry is one it can never
 * speak. A page that draws a card in its own language alone carries only that
 * language's words, and never imports this module. The preview and the
 * widget's tests speak the pseudo-language too — English stretched further
 * than any of the languages stretches it, to see a card hold longer words —
 * and a build for production leaves it out, since a host could name it. It is
 * the build's mode that decides, not the environment it runs in.
 */
export const dictionaries: ReadonlyMap<string, Dictionary> =
	import.meta.env.MODE === "production"
		? written
		: new Map([
				...written,
				[pseudoLocale, pseudoWords(english, written.values())],
			]);

const spoken: ReadonlySet<string> = new Set(dictionaries.keys());

/**
 * lessonLanguages are the languages a parent can choose for the lessons on a
 * card: every one the widget is written in, so that a card of the lessons
 * speaks the language its task is written in. The pseudo-language is no
 * language a lesson is held in.
 */
export const lessonLanguages: readonly string[] = [...written.keys()];

/**
 * catalogSkills are the skills a parent can leave out of the tasks, by their
 * ids, in the order of the catalog: every skill the card has a name for, its
 * words kept in that order.
 */
export const catalogSkills: readonly string[] = Object.keys(english)
	.filter((key) => key.startsWith("skill."))
	.map((key) => key.slice("skill.".length));

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
