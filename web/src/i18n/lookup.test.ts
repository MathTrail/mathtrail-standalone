import { describe, expect, test } from "vitest";
import { chooseLocale, directionOf, fallbacksOf, intlLocales } from "./lookup";
import vectors from "./testdata/lookup.json";

// A language is found among others whole, then shorter, then in the script it
// is usually written in, but never cut past its script: a bare language
// stands for the script it is usually written in, and a reader of another
// script is not given it — Traditional characters are not the Simplified, nor
// Shahmukhi the Gurmukhi. Norwegian named as a whole is Bokmål, which most of
// its readers write; Nynorsk is a standard of its own. The cases are kept in
// a file of their own, the one rule a language is read by wherever it is
// read.
describe("the dictionary chosen for a language", () => {
	test.each(
		vectors.flatMap(({ among, read }) =>
			Object.entries(read).map(([tag, want]) => [tag, among, want] as const),
		),
	)("for %j among %j is %s", (tag, among, want) => {
		expect(chooseLocale([tag], new Set(among))).toBe(want);
	});

	test("for no language at all is English", () => {
		expect(chooseLocale([undefined], new Set(["en", "ru"]))).toBe("en");
	});

	test.each([
		[["ru", "en-US"], "ru"],
		[["en", "ru-RU"], "en"],
		[[undefined, "ru-RU"], "ru"],
		[["kk", "ru-RU"], "ru"],
		[["und", "ru-RU"], "ru"],
		[["not a language", "ru-RU"], "ru"],
		[["kk", "es-MX"], "en"],
	])("for %j, most wanted first, is %s", (wanted, want) => {
		expect(chooseLocale(wanted, new Set(["en", "ru"]))).toBe(want);
	});
});

describe("the dictionaries a missing word is looked for in", () => {
	test.each([
		["pt-BR", ["pt-BR", "pt", "en"]],
		["zh-Hans", ["zh-Hans", "zh", "en"]],
		["zh-Hant", ["zh-Hant", "en"]],
		["sr-Latn", ["sr-Latn", "en"]],
		["ru", ["ru", "en"]],
		["en", ["en"]],
	])("for %s are %j", (locale, want) => {
		expect(fallbacksOf(locale)).toEqual(want);
	});
});

describe("the locales a language is formatted in", () => {
	test("are its own, then English", () => {
		expect(intlLocales("mai")).toEqual(["mai", "en"]);
	});

	// Klingon, which no platform has data for, is counted, written, listed,
	// named and sorted in English rather than in the machine's own language.
	test("are English for a language the platform has no data for", () => {
		const formats = [
			new Intl.Collator(intlLocales("tlh")),
			new Intl.DisplayNames(intlLocales("tlh"), { type: "language" }),
			new Intl.ListFormat(intlLocales("tlh")),
			new Intl.NumberFormat(intlLocales("tlh")),
			new Intl.PluralRules(intlLocales("tlh")),
		];

		for (const format of formats) {
			expect(format.resolvedOptions().locale).toBe("en");
		}
	});
});

describe("the direction a language runs in", () => {
	test.each(["ar", "ar-EG", "fa", "ur", "he", "iw", "az-Arab", "ff-Adlm"])(
		"is right to left for %s",
		(tag) => {
			expect(directionOf(tag)).toBe("rtl");
		},
	);

	test.each(["en", "ru", "zh-Hans", "az", "hi", "xx", "not a language"])(
		"is left to right for %s",
		(tag) => {
			expect(directionOf(tag)).toBe("ltr");
		},
	);
});
