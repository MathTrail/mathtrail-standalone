import { describe, expect, test } from "vitest";
import { chooseLocale, directionOf, fallbacksOf, intlLocales } from "./lookup";

// The two languages the widget speaks today, and a wider set such as its
// dictionaries will make, to see how a tag finds its language among them.
const today = new Set(["en", "ru"]);
const wider = new Set(["en", "es", "pt", "ru", "zh-Hans"]);

describe("the dictionary chosen for a language", () => {
	test.each([
		["ru-RU", "ru"],
		["ru", "ru"],
		["en-GB", "en"],
		["es-MX", "en"],
		["pt-BR", "en"],
		["kk", "en"],
		["be-BY", "en"],
		["xx", "en"],
		["und-RU", "en"],
		["und", "en"],
		["ru-Latn", "en"],
		["not a language", "en"],
		["", "en"],
		[undefined, "en"],
	])("for %s, of en and ru, is %s", (tag, want) => {
		expect(chooseLocale([tag], today)).toBe(want);
	});

	test.each([
		["es-MX", "es"],
		["es-419", "es"],
		["pt-BR", "pt"],
		["PT-br", "pt"],
		["zh-CN", "zh-Hans"],
		["zh-Hans-CN", "zh-Hans"],
		["zh", "zh-Hans"],
		["zh-TW", "en"],
		["ru-RU-u-nu-latn", "ru"],
		["tlh", "en"],
	])("for %s, among more languages, is %s", (tag, want) => {
		expect(chooseLocale([tag], wider)).toBe(want);
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
		expect(chooseLocale(wanted, today)).toBe(want);
	});

	// A bare language stands for the script it is usually written in, and a
	// reader of another script is not given it: Traditional characters are not
	// the Simplified, nor Shahmukhi the Gurmukhi.
	test.each([
		["zh-CN", "zh"],
		["zh-SG", "zh"],
		["zh-TW", "en"],
		["zh-Hant", "en"],
		["sr", "sr"],
		["sr-Latn-RS", "en"],
		["pa", "pa"],
		["pa-Arab", "en"],
	])("for %s, where the bare languages have words, is %s", (tag, want) => {
		expect(chooseLocale([tag], new Set(["en", "zh", "sr", "pa"]))).toBe(want);
	});

	// Norwegian named as a whole is Bokmål, which most of its readers write;
	// Nynorsk is a standard of its own, which has no words of its own here.
	test.each([
		["no", "nb"],
		["no-NO", "nb"],
		["NO-no", "nb"],
		["nb", "nb"],
		["nb-NO", "nb"],
		["nn", "en"],
		["nn-NO", "en"],
	])("for %s, where Bokmål has words, is %s", (tag, want) => {
		expect(chooseLocale([tag], new Set(["en", "nb"]))).toBe(want);
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
