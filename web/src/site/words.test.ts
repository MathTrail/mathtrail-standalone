import { describe, expect, test } from "vitest";
import { disagreements } from "../i18n/dictionaries";
import { listMarks, siteDictionaries, wordSpace } from "./words";

describe("the site's words", () => {
	const english = siteDictionaries.get("en") ?? {};

	test("are in English, the language every other is written from", () => {
		expect(Object.keys(english).length).toBeGreaterThan(0);
	});

	test.each([...siteDictionaries.keys()])(
		"in %s are named by the tag a browser writes",
		(tag) => {
			expect(new Intl.Locale(tag).baseName).toBe(tag);
		},
	);

	test.each([...siteDictionaries])(
		"in %s say what the English say",
		(tag, words) => {
			expect(disagreements(english, words, tag)).toEqual([]);
		},
	);
});

describe("a list of names", () => {
	test.each([
		["in English, of three names", "en", ["A", "B", "C"], ["", ", ", ", "]],
		[
			"in Spanish, of three names",
			"es",
			["Juegos", "Libros", "Cuentos"],
			["", ", ", " y "],
		],
		[
			"in Spanish, before a name of an i sound",
			"es",
			["Juegos", "Historias"],
			["", " e "],
		],
		["in Japanese", "ja", ["『A』", "『B』"], ["", "、"]],
		["in Chinese", "zh-Hans", ["《A》", "《B》"], ["", "、"]],
		["in English, of one name", "en", ["A"], [""]],
	] as const)(
		"is set apart %s as the language sets it apart",
		(_, locale, names, marks) => {
			expect(listMarks(locale, names)).toEqual(marks);
		},
	);
});

describe("two texts side by side", () => {
	test.each([
		["a count and its English noun", "12", "reference tasks", " "],
		["a label and its English text", "When it helps:", "ages", " "],
		["a count and its Japanese noun", "12", "問のお手本問題", ""],
		["a count and its Chinese noun", "12", "道参考题", ""],
		[
			"a label that ends in a full-width colon",
			"使える場面：",
			"「〜より」",
			"",
		],
		["a Chinese label and a Latin word", "适用场合：", "Claude", ""],
		[
			"a Latin label that ends in a middle dot",
			"Example 2 ·",
			"Grades 3–4",
			" ",
		],
		[
			"a strong Japanese label and its text",
			"**使える場面：**",
			"Claude の画面",
			"",
		],
		["a count and a noun that opens with emphasis", "12", "*問*", ""],
		// A long run of emphasis is pinned by the space it gives, not by how
		// long it takes: a clock in a test fails on a loaded machine, and the
		// run is read once by how wordSpace is written.
		[
			"a Japanese word closed by a long run of emphasis and a Latin word",
			`問${"*".repeat(5000)}`,
			"Claude",
			"",
		],
		[
			"a text with a long run of emphasis short of its end and a noun",
			`12${"_".repeat(5000)}x`,
			"tasks",
			" ",
		],
	])("%s are set apart by %j", (_, before, after, space) => {
		expect(wordSpace(before, after)).toBe(space);
	});
});
