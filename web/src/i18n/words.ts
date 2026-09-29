import { directionOf, fallbacksOf } from "./lookup";

/**
 * Wording is what a dictionary says for a key: a text, or, for words that
 * change with a number, a text for each plural category of the language —
 * "1 time", "3 times".
 */
export type Wording =
	| string
	| Readonly<Partial<Record<Intl.LDMLPluralRule, string>>>;

/** Dictionary is one language's words by key, keys in dot notation. */
export type Dictionary = Readonly<Record<string, Wording>>;

/**
 * Slots fill a wording's placeholders, `{name}`, by name. A number is written
 * the way the wording's language writes numbers, and the one named `count`
 * also chooses which of a plural wording's texts is said.
 */
export type Slots = Readonly<Record<string, string | number>>;

/** Words are what one language says, by key. */
export type Words<Key extends string> = {
	/** locale is the tag of the dictionary the words are read from. */
	readonly locale: string;
	/** dir is the direction the words run in. */
	readonly dir: "ltr" | "rtl";
	/** text is what the words say for key, with its slots filled. */
	text(key: Key, slots?: Slots): string;
};

/**
 * placeholder is a slot as a wording names it: its name in lowercase letters,
 * in braces — `{count}`.
 */
export const placeholder = /\{([a-z]+)\}/g;

// source is a dictionary a wording may come from, with its language's way of
// counting and of writing a number.
type Source = {
	dictionary: Dictionary;
	plurals: Intl.PluralRules;
	numbers: Intl.NumberFormat;
};

/**
 * openWords are the words of the dictionary tagged locale. A key it lacks is
 * looked for as a language is — under its shorter tag, then in English — and
 * said in that language. A development build throws instead: a gap in a
 * dictionary is a bug to mend, not a fallback to lean on. So it does for a
 * slot a wording names and the caller leaves empty, which the page would
 * otherwise show as it is written, and for a plural wording given no count,
 * which is otherwise said as its `other` text.
 */
export function openWords<Key extends string>(
	locale: string,
	dictionaries: ReadonlyMap<string, Dictionary>,
): Words<Key> {
	const sources: Source[] = [];
	for (const tag of fallbacksOf(locale)) {
		const dictionary = dictionaries.get(tag);
		if (dictionary !== undefined) {
			sources.push({
				dictionary,
				plurals: new Intl.PluralRules(tag),
				numbers: new Intl.NumberFormat(tag),
			});
		}
	}

	return {
		locale,
		dir: directionOf(locale),
		text(key, slots = {}) {
			const own = dictionaries.get(locale);
			if (own === undefined || !Object.hasOwn(own, key)) {
				failInDevelopment(`the words in ${locale} have nothing for ${key}`);
			}
			const source = sources.find(({ dictionary }) =>
				Object.hasOwn(dictionary, key),
			);
			if (source === undefined) {
				// Not even English has it: the key at least says where to look.
				return key;
			}
			return fill(source, said(source, key, slots.count), slots);
		},
	};
}

// said is the text a source says for key: its only one, or the one for
// count's plural category.
function said(
	{ dictionary, plurals }: Source,
	key: string,
	count: Slots[string] | undefined,
): string {
	const wording = dictionary[key] ?? "";
	if (typeof wording === "string") {
		return wording;
	}
	if (typeof count !== "number") {
		failInDevelopment(`${key} changes with a number, and no count was given`);
		return wording.other ?? "";
	}
	return wording[plurals.select(count)] ?? wording.other ?? "";
}

// fill is text with its placeholders replaced by the slots given, a number
// written the way the source's language writes it.
function fill({ numbers }: Source, text: string, slots: Slots): string {
	return text.replace(placeholder, (written, name: string) => {
		const value = Object.hasOwn(slots, name) ? slots[name] : undefined;
		if (value === undefined) {
			failInDevelopment(`"${text}" names {${name}}, and it was not given`);
			return written;
		}
		return typeof value === "number" ? numbers.format(value) : value;
	});
}

// failInDevelopment throws in a development build, where a mistake in the
// words should stop whoever made it, and does nothing in the build that
// ships, where a card that says a little less beats a card that breaks.
function failInDevelopment(what: string): void {
	if (import.meta.env.DEV) {
		throw new Error(`words: ${what}`);
	}
}
