import { type Dictionary, placeholder, type Wording } from "./words";

// textsOf are the texts a wording says: its one, or one per plural category.
function textsOf(wording: Wording): unknown[] {
	return typeof wording === "string" ? [wording] : Object.values(wording);
}

// slotsOf are the slots a wording names, in any of its texts, sorted.
function slotsOf(wording: Wording): string[] {
	const names = textsOf(wording).flatMap((text) =>
		typeof text === "string"
			? [...text.matchAll(placeholder)].flatMap(([, name]) => name ?? [])
			: [],
	);
	return [...new Set(names)].sort();
}

// strayBraces says whether a text has a brace that is not part of a
// placeholder — {Grade}, {task_id}, an unclosed {grade — which a page would
// show as it is written and the comparison of slots could not see.
function strayBraces(text: string): boolean {
	return /[{}]/.test(text.replace(placeholder, ""));
}

// hiddenMark says whether a text has a character a reader cannot see, one of
// Unicode's format characters: a mark that turns the way text runs, a space
// or a joiner of no width, a soft hyphen. A card or a page sets which way its
// words run itself, and wraps them itself, and such a character in a wording
// would do either otherwise. The joiners Persian and the scripts of India write their
// words with are part of those words, and not among them.
function hiddenMark(text: string): boolean {
	return /(?![\u200c\u200d])\p{Cf}/u.test(text);
}

/**
 * disagreements are the ways the words of one language, tagged tag, say other
 * things than the English ones: a key one of them lacks, a wording with other
 * slots, a wording plural in one and not in the other, a text that is empty or
 * no text, a brace that is no placeholder, a mark a reader cannot see, and a
 * plural wording that does not name exactly the categories its language
 * counts in — one it lacks is a form of the words a page cannot say, and one
 * its language never chooses is a form nobody reads.
 */
export function disagreements(
	english: Dictionary,
	words: Dictionary,
	tag: string,
): string[] {
	const found = Object.keys(english)
		.filter((key) => !Object.hasOwn(words, key))
		.map((key) => `${key} is missing`);
	const counted = [
		...new Intl.PluralRules(tag).resolvedOptions().pluralCategories,
	].sort();
	for (const [key, wording] of Object.entries(words)) {
		const reference = english[key];
		if (reference === undefined) {
			found.push(`${key} is not among the English words`);
			continue;
		}
		if (slotsOf(wording).join() !== slotsOf(reference).join()) {
			found.push(
				`${key} names the slots [${slotsOf(wording)}], the English [${slotsOf(reference)}]`,
			);
		}
		if ((typeof wording === "string") !== (typeof reference === "string")) {
			found.push(`${key} changes with a number in one language only`);
		}
		if (
			textsOf(wording).some(
				(text) => typeof text !== "string" || text.trim() === "",
			)
		) {
			found.push(`${key} has a text that is empty or no text`);
		}
		if (
			textsOf(wording).some(
				(text) => typeof text === "string" && strayBraces(text),
			)
		) {
			found.push(`${key} has a brace that is no placeholder`);
		}
		if (
			textsOf(wording).some(
				(text) => typeof text === "string" && hiddenMark(text),
			)
		) {
			found.push(`${key} has a mark a reader cannot see`);
		}
		if (typeof wording !== "string") {
			const named = Object.keys(wording).sort();
			if (named.join() !== counted.join()) {
				found.push(
					`${key} is written for [${named}], ${tag} counts in [${counted}]`,
				);
			}
		}
	}
	return found;
}
