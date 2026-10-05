import { afterEach, describe, expect, test, vi } from "vitest";
import { type Dictionary, dictionariesByTag, openWords } from "./words";

// Words in a few languages, the Brazilian ones left unfinished, to see how a
// wording is said, counted and filled, and where a missing one comes from.
const dictionaries = new Map<string, Dictionary>([
	[
		"en",
		{
			greeting: "Hello, {name}",
			distance: "Distance {metres} m",
			times: { one: "{count} time", other: "{count} times" },
			english: "Only in English",
		},
	],
	[
		"pt",
		{
			greeting: "Olá, {name}",
			distance: "Distância {metres} m",
			times: {
				one: "{count} vez",
				many: "{count} de vezes",
				other: "{count} vezes",
			},
		},
	],
	["pt-BR", { greeting: "Oi, {name}" }],
	[
		"ru",
		{
			times: {
				one: "{count} раз",
				few: "{count} раза",
				many: "{count} раз",
				other: "{count} раза",
			},
		},
	],
	["de", { distance: "Entfernung {metres} m" }],
	["fr", { greeting: "Bonjour, {name}" }],
	["ar", { greeting: "مرحبا {name}" }],
]);

afterEach(() => {
	vi.unstubAllEnvs();
});

describe("words", () => {
	test("say a wording with its slots filled", () => {
		const words = openWords("en", dictionaries);

		expect(words.text("greeting", { name: "Otter" })).toBe("Hello, Otter");
	});

	test("write a number the way their language writes it", () => {
		expect(
			openWords("de", dictionaries).text("distance", { metres: 1573 }),
		).toBe("Entfernung 1.573 m");
		expect(
			openWords("en", dictionaries).text("distance", { metres: 1573 }),
		).toBe("Distance 1,573 m");
	});

	test.each([
		[1, "1 раз"],
		[3, "3 раза"],
		[5, "5 раз"],
		[21, "21 раз"],
		[22, "22 раза"],
		[1.5, "1,5 раза"],
	])("count %s in Russian as %s", (count, want) => {
		expect(openWords("ru", dictionaries).text("times", { count })).toBe(want);
	});

	test.each([
		[0, "0 times"],
		[1, "1 time"],
		[2, "2 times"],
	])("count %s in English as %s", (count, want) => {
		expect(openWords("en", dictionaries).text("times", { count })).toBe(want);
	});

	test("name their language and the way it runs", () => {
		const russian = openWords("ru", dictionaries);
		const arabic = openWords("ar", dictionaries);

		expect([russian.locale, russian.dir]).toEqual(["ru", "ltr"]);
		expect([arabic.locale, arabic.dir]).toEqual(["ar", "rtl"]);
	});
});

describe("in a development build, words", () => {
	test("throw for a word their dictionary lacks", () => {
		const words = openWords("pt-BR", dictionaries);

		expect(() => words.text("times", { count: 2 })).toThrow(
			"the words in pt-BR have nothing for times",
		);
	});

	test("throw for a slot left empty", () => {
		const words = openWords("en", dictionaries);

		expect(() => words.text("greeting")).toThrow("{name}");
	});

	test("throw for a plural wording given no count", () => {
		const words = openWords("en", dictionaries);

		expect(() => words.text("times")).toThrow("no count");
	});
});

describe("in the build that ships, words", () => {
	test("take a word their dictionary lacks from its shorter tag, then from English", () => {
		vi.stubEnv("DEV", false);
		const words = openWords("pt-BR", dictionaries);

		expect(words.text("greeting", { name: "Ana" })).toBe("Oi, Ana");
		expect(words.text("times", { count: 2 })).toBe("2 vezes");
		expect(words.text("english")).toBe("Only in English");
	});

	test("count a word taken from another language by that language's rules", () => {
		vi.stubEnv("DEV", false);

		// French counts nothing as one, English as many.
		expect(openWords("fr", dictionaries).text("times", { count: 0 })).toBe(
			"0 times",
		);
	});

	test("say a key no dictionary has as the key", () => {
		vi.stubEnv("DEV", false);

		expect(openWords("en", dictionaries).text("nowhere")).toBe("nowhere");
	});

	test("leave a slot left empty as it is written", () => {
		vi.stubEnv("DEV", false);

		expect(openWords("en", dictionaries).text("greeting")).toBe(
			"Hello, {name}",
		);
	});

	test("say a plural wording given no count by its other text", () => {
		vi.stubEnv("DEV", false);

		expect(openWords("en", dictionaries).text("times")).toBe("{count} times");
	});

	test.each([
		["their own language has", "pt-BR", "greeting", true],
		["a language they fall back on has", "pt-BR", "distance", true],
		["only English has, when English is among them", "ru", "english", true],
		["no language of theirs has", "ru", "nowhere", false],
	])("have a key %s", (_, locale, key, want) => {
		expect(openWords(locale, dictionaries).has(key)).toBe(want);
	});

	test("given one language alone have no key it lacks, whatever English has", () => {
		const russianAlone = new Map([["ru", dictionaries.get("ru") ?? {}]]);

		expect(openWords("ru", russianAlone).has("times")).toBe(true);
		expect(openWords("ru", russianAlone).has("english")).toBe(false);
	});
});

describe("dictionaries a glob found", () => {
	test("are known by the tag each file is named after", () => {
		const en: Dictionary = { "task.hint": "Hint" };
		const zh: Dictionary = { "task.hint": "提示" };

		expect(
			dictionariesByTag({
				"../../locales/en.json": en,
				"./locales/zh-Hans.json": zh,
			}),
		).toEqual(
			new Map([
				["en", en],
				["zh-Hans", zh],
			]),
		);
	});
});
