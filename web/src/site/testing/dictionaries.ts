import type { Dictionary } from "../../i18n/words";
import type { Sources } from "../content";
import { siteDictionaries } from "../words";

/**
 * dictionariesOf are the site's dictionaries of the languages sources are
 * written in, and of no other. A test draws a small site in a language or two
 * of its own; the build refuses words in a language that has no texts, so the
 * site a test draws carries the words of its own languages alone, whichever
 * languages the site itself is written in.
 */
export function dictionariesOf(
	sources: Sources,
): ReadonlyMap<string, Dictionary> {
	return new Map(
		[...sources.keys()].flatMap((locale) => {
			const words = siteDictionaries.get(locale);
			return words === undefined ? [] : [[locale, words] as const];
		}),
	);
}
