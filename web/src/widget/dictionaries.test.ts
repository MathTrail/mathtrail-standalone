import { afterEach, describe, expect, test, vi } from "vitest";
import notes from "../../locales/notes/keys.json";
import { disagreements } from "../i18n/dictionaries";
import { lengthOf, pseudoLocale } from "../i18n/pseudo";
import { textsOf } from "../i18n/words";
import { cardWords, dictionaries, lessonLanguages } from "./dictionaries";

describe("the widget's words", () => {
	const english = dictionaries.get("en") ?? {};

	test("are in English, the language every other is written from", () => {
		expect(dictionaries.has("en")).toBe(true);
		expect(Object.keys(english).length).toBeGreaterThan(0);
	});

	test.each([...dictionaries.keys()])(
		"in %s are named by the tag a browser writes",
		(tag) => {
			expect(new Intl.Locale(tag).baseName).toBe(tag);
		},
	);

	test.each([...dictionaries])(
		"in %s say what the English say",
		(tag, words) => {
			expect(disagreements(english, words, tag)).toEqual([]);
		},
	);

	// The adult types in the chat and the child answers on the card, so a card
	// that sends its reader to the chat to ask for a task speaks to the adult,
	// in the form its words for the adult already take where the language
	// tells the two apart: the form the card asks for a new profile in.
	test.each([
		["ar", "يمكن طلب"],
		["bn", "চান"],
		["de", "Sie"],
		["es", "pida"],
		["fa", "بخواهید"],
		["fr", "demandez"],
		["hi", "माँगें"],
		["it", "chieda"],
		["ja", "頼んでください"],
		["ko", "요청해 주세요"],
		["ru", "попросите"],
		["tr", "isteyin"],
		["uk", "попросіть"],
		["ur", "مانگیں"],
		["vi", "yêu cầu"],
		["zh-Hans", "请"],
	])("in %s send the adult to the chat as %s", (tag, form) => {
		const words = dictionaries.get(tag) ?? {};
		for (const key of [
			"profile.gone",
			"waiting.ask_in_chat",
			"waiting.next_below",
			"waiting.slow_detail",
		]) {
			expect(String(words[key]).toLocaleLowerCase(tag), key).toContain(
				form.toLocaleLowerCase(tag),
			);
		}
	});

	test("tell a rank by its number, with no name of its own", () => {
		expect(
			Object.keys(english).filter((key) => key.startsWith("rank.")),
		).toEqual([]);
		expect(cardWords("en", undefined).text("progress.rank", { rank: 3 })).toBe(
			"Rank 3",
		);
		expect(cardWords("ru", undefined).text("progress.rank", { rank: 11 })).toBe(
			"Ранг 11",
		);
	});
});

describe("a card's words", () => {
	test.each([
		["ru", "en-US", "ru"],
		["en", "ru-RU", "en"],
		[undefined, "ru-RU", "ru"],
		["be", "ru-RU", "ru"],
		[undefined, "es-MX", "es"],
		["pt-BR", "ru-RU", "pt"],
		[undefined, "zh-CN", "zh-Hans"],
		[undefined, "so-SO", "en"],
		[undefined, undefined, "en"],
	])(
		"with the language chosen %s and the host's %s are in %s",
		(chosen, host, want) => {
			expect(cardWords(chosen, host).locale).toBe(want);
		},
	);

	test("say what their dictionary says, with the grade in its slot", () => {
		const russian = cardWords("ru", undefined);

		expect(russian.text("task.hint")).toBe("Подсказка");
		expect(russian.text("child.grade", { grade: 3 })).toBe("3 класс");
		expect(
			cardWords(undefined, "en-GB").text("child.grade", { grade: 3 }),
		).toBe("Grade 3");
	});

	test("are in every one of the twenty-two languages of v1", () => {
		expect([...dictionaries.keys()]).toEqual(
			expect.arrayContaining([
				"en",
				"zh-Hans",
				"hi",
				"es",
				"ar",
				"fr",
				"bn",
				"pt",
				"ru",
				"ur",
				"id",
				"de",
				"ja",
				"tr",
				"ko",
				"vi",
				"it",
				"fa",
				"pl",
				"uk",
				"th",
				"nl",
			]),
		);
	});
});

describe("the pseudo-language", () => {
	afterEach(() => {
		vi.unstubAllEnvs();
		vi.unstubAllGlobals();
		vi.resetModules();
	});

	test("is spoken by the preview and the tests", () => {
		expect(dictionaries.has(pseudoLocale)).toBe(true);
	});

	// A card that holds the pseudo-language holds every language's words: it
	// says each text of each key at least as long as any dictionary says one,
	// in letters as a reader sees them.
	test.each([...dictionaries].filter(([tag]) => tag !== pseudoLocale))(
		"says every text no shorter than %s says its key",
		(_, words) => {
			const pseudo = dictionaries.get(pseudoLocale) ?? {};
			const shorter = Object.entries(words).flatMap(([key, wording]) =>
				textsOf(wording).flatMap((text) =>
					textsOf(pseudo[key] ?? {})
						.filter((said) => lengthOf(said) < lengthOf(text))
						.map((said) => `${key}: ${said} is shorter than ${text}`),
				),
			);

			expect(shorter).toEqual([]);
		},
	);

	test("is left out of a build for production, which a host could ask it of", async () => {
		vi.stubEnv("MODE", "production");
		vi.resetModules();

		const { dictionaries: shipped } = await import("./dictionaries");

		expect(shipped.has(pseudoLocale)).toBe(false);
		expect(shipped.has("en")).toBe(true);
	});

	// The build that ships carries the pseudo-language's code and never runs
	// it, so the card's words load where the platform cannot split letters.
	test("asks the platform for nothing in a build for production", async () => {
		vi.stubEnv("MODE", "production");
		vi.stubGlobal(
			"Intl",
			Object.create(Intl, { Segmenter: { value: undefined } }),
		);
		vi.resetModules();

		const { dictionaries: shipped } = await import("./dictionaries");

		expect(shipped.has("en")).toBe(true);
	});
});

describe("the notes for whoever translates the words", () => {
	const english = dictionaries.get("en") ?? {};

	// A key with no note is translated blind, and a note with no key is read
	// for nothing; in English's order, the two files read side by side.
	test("are one for every key of English, in its order", () => {
		expect(Object.keys(notes)).toEqual(Object.keys(english));
	});

	test.each(Object.entries(notes))(
		"say whom %s is for, where it stands, how long it may be and how it speaks",
		(_, note) => {
			expect(Object.keys(note)).toEqual(
				expect.arrayContaining(["reader", "where", "length", "address"]),
			);
			expect(["reader", "where", "length", "address", "mind"]).toEqual(
				expect.arrayContaining(Object.keys(note)),
			);
			expect(["child", "parent", "both"]).toContain(note.reader);
			expect(["label", "line", "sentence", "heard"]).toContain(note.length);
			expect(["informal", "polite", "none"]).toContain(note.address);
			for (const said of Object.values(note)) {
				expect(said.trim()).not.toBe("");
			}
			// The child is spoken to as a child, and the parent as an adult.
			const wrongFor: Readonly<Record<string, string>> = {
				child: "polite",
				parent: "informal",
			};
			expect(note.address).not.toBe(wrongFor[note.reader]);
		},
	);
});

describe("the languages of the lessons", () => {
	// A language a card cannot speak would be one its task is written in and
	// its buttons are not.
	test("are every language the widget is written in, and never the pseudo-language", () => {
		expect([...lessonLanguages].sort()).toEqual(
			[...dictionaries.keys()].filter((tag) => tag !== pseudoLocale).sort(),
		);
	});
});
