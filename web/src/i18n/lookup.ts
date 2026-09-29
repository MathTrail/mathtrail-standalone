/**
 * fallbackLocale is the language spoken when no language wanted has words:
 * every dictionary is written from the English one, so English is the one
 * certain to have every word.
 */
export const fallbackLocale = "en";

/**
 * chooseLocale is the tag of the dictionary to speak for the languages wanted,
 * most wanted first: the first of them there are words for, and English when
 * there are words for none. Each is tried whole and then shorter — `pt-BR`,
 * then `pt` — and then with the script its language is usually written in, so
 * that `zh-CN` finds `zh-Hans`; but never shorter than its script, so that
 * `zh-TW` is not answered in the simplified characters a bare `zh` stands
 * for. A language no tag names is never tried: `kk` is not answered in
 * Russian because many who read Kazakh read Russian too, nor `und-RU` because
 * most in Russia read Russian.
 */
export function chooseLocale(
	wanted: readonly (string | undefined)[],
	available: ReadonlySet<string>,
): string {
	for (const tag of wanted) {
		const found = candidatesOf(tag).find((candidate) =>
			available.has(candidate),
		);
		if (found !== undefined) {
			return found;
		}
	}
	return fallbackLocale;
}

/**
 * fallbacksOf are the dictionaries a word missing from locale's is looked for
 * in, locale's own first: its shorter tags written in its script, then
 * English — never a bare language written in another script, as `zh` is for
 * `zh-Hant`.
 */
export function fallbacksOf(locale: string): string[] {
	return [
		...new Set([
			...writtenIn(shorterTags(locale), scriptOf(locale)),
			fallbackLocale,
		]),
	];
}

// The scripts written right to left, among those a language's usual script
// can be.
const rightToLeftScripts = new Set([
	"Adlm",
	"Arab",
	"Hebr",
	"Mand",
	"Mend",
	"Nkoo",
	"Rohg",
	"Samr",
	"Syrc",
	"Thaa",
	"Yezi",
]);

/**
 * directionOf is the direction a language's text runs in: right to left when
 * its script is written that way — the script the tag names, or the one its
 * language is usually written in — and left to right otherwise, and for a tag
 * that names no language. The script decides rather than a list of languages,
 * so a language added later needs no code to be laid out.
 */
export function directionOf(tag: string): "ltr" | "rtl" {
	const script = scriptOf(tag);
	return script !== undefined && rightToLeftScripts.has(script) ? "rtl" : "ltr";
}

// candidatesOf are the tags a wanted language may be found under, in the
// order they are tried, written as a dictionary's name writes them: those
// written in the script the language is wanted in, and none for a tag that
// names no language. Extensions, such as a numbering system, say nothing about
// the words and are left out.
function candidatesOf(tag: string | undefined): string[] {
	const locale = tag === undefined ? undefined : localeOf(tag);
	// An undetermined language is read from its first subtag: newer engines
	// leave the locale's language undefined for it, older ones say "und".
	if (locale === undefined || locale.baseName.split("-")[0] === "und") {
		return [];
	}
	const likely = locale.maximize();
	return writtenIn(
		[
			...new Set([
				...shorterTags(locale.baseName),
				...shorterTags(likely.baseName),
			]),
		],
		likely.script,
	);
}

// writtenIn are the tags among tags written in script, the one they name or
// the one their language is usually written in.
function writtenIn(tags: string[], script: string | undefined): string[] {
	return tags.filter((tag) => scriptOf(tag) === script);
}

// scriptOf is the script a tag names, or the one its language is usually
// written in; undefined for a tag that names no language, or a language whose
// script is not known.
function scriptOf(tag: string): string | undefined {
	return localeOf(tag)?.maximize().script;
}

// shorterTags are a tag and every tag its end can be cut to: zh-Hans-CN,
// zh-Hans, zh.
function shorterTags(tag: string): string[] {
	const subtags = tag.split("-");
	return subtags.map((_, cut) =>
		subtags.slice(0, subtags.length - cut).join("-"),
	);
}

// localeOf is the locale a tag names, or undefined when it is no language tag.
function localeOf(tag: string): Intl.Locale | undefined {
	try {
		return new Intl.Locale(tag);
	} catch {
		return undefined;
	}
}
