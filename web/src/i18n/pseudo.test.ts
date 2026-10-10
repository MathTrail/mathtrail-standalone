import { describe, expect, test } from "vitest";
import english from "../../locales/en.json";
import { lengthOf, pseudoLocale, pseudoText, pseudoWords } from "./pseudo";
import { type Dictionary, placeholder } from "./words";

// textsByKey are the texts of every wording of words, by key and form.
function textsByKey(words: Dictionary): [string, string][] {
	return Object.entries(words).flatMap(([key, wording]) =>
		typeof wording === "string"
			? [[key, wording] as [string, string]]
			: Object.entries(wording).map(
					([form, text]) => [`${key}.${form}`, text] as [string, string],
				),
	);
}

describe("the pseudo-language", () => {
	const pseudo = pseudoWords(english, []);
	const texts = new Map(textsByKey(english));

	test("is English by its tag, so it counts and runs as English does", () => {
		const locale = new Intl.Locale(pseudoLocale);

		expect(locale.language).toBe("en");
		expect(locale.baseName).toBe(pseudoLocale);
	});

	test.each(textsByKey(pseudo))(
		"says %s twice and ten letters longer than English, in brackets",
		(key, text) => {
			const said = texts.get(key) ?? "";

			expect(lengthOf(text)).toBeGreaterThanOrEqual(lengthOf(said) * 2);
			expect(lengthOf(text)).toBeGreaterThanOrEqual(lengthOf(said) + 10);
			expect(text.startsWith("[")).toBe(true);
			expect(text.endsWith("]")).toBe(true);
		},
	);

	test("says a key no shorter than the longest text a translation says for it, in any of its forms", () => {
		const spanish = "Una pista que dice mucho más que la inglesa";
		const russian = "{count} раз, и ещё много-много слов после него";
		const said = pseudoWords(
			{
				"task.hint": "Hint",
				"child.times": { one: "{count} time", other: "{count} times" },
			},
			[
				{ "task.hint": spanish },
				{
					"child.times": {
						one: "{count} раз",
						few: "{count} раза",
						many: russian,
						other: "{count} раза",
					},
				},
			],
		);

		// The longest text each of the pseudo-language's texts is held to.
		const longest: Record<string, string> = {
			"task.hint": spanish,
			"child.times.one": russian,
			"child.times.other": russian,
		};
		expect(textsByKey(said).map(([key]) => key)).toEqual(Object.keys(longest));
		for (const [key, text] of textsByKey(said)) {
			expect(lengthOf(text), key).toBeGreaterThanOrEqual(
				lengthOf(longest[key] ?? ""),
			);
		}
	});

	// A dictionary is read from a file, and what it holds in place of a text
	// says nothing: the check of the dictionaries is what names it.
	test("is made whatever a translation holds in place of a text", () => {
		const broken: Dictionary = JSON.parse(
			'{"task.hint": null, "child.grade": 3, "progress.times": {"one": null, "other": 3}}',
		);

		expect(pseudoWords(english, [broken])).toEqual(pseudo);
	});

	test("is left as it was by translations no longer than it", () => {
		const shorter = Object.fromEntries(
			Object.keys(english).map((key) => [key, "Kurz"]),
		);

		expect(pseudoWords(english, [shorter, english])).toEqual(pseudo);
	});

	test.each([
		["a letter with its mark, as a script of India writes it", "कि", 1],
		["an emoji with its skin tone", "👍🏽", 1],
		["a text with a slot, which the slot's words decide", "{count} times", 6],
	])("counts %s in the letters a reader sees", (_, text, letters) => {
		expect(lengthOf(text)).toBe(letters);
	});

	test("keeps every slot, and every plural wording's forms", () => {
		for (const [key, wording] of Object.entries(english)) {
			const said = pseudo[key];
			if (typeof wording === "string") {
				expect(said, key).toBeTypeOf("string");
				continue;
			}
			expect(Object.keys(said ?? {}), key).toEqual(Object.keys(wording));
		}
		for (const [key, text] of textsByKey(pseudo)) {
			const slots = (words: string) =>
				[...words.matchAll(placeholder)].map(([slot]) => slot);
			expect(slots(text), key).toEqual(slots(texts.get(key) ?? ""));
		}
	});

	test("shows no letter of English as it was", () => {
		for (const [key, text] of textsByKey(pseudo)) {
			expect(text.replace(placeholder, ""), key).not.toMatch(/[a-zA-Z]/);
		}
	});

	test.each([
		["Hint", "[Ĥííñţ óóñéé ţŵóó]"],
		["{before} → {after}", "[{before} → {after} óóñéé ţŵóó]"],
		[
			"Rank {rank} of {total} · {name}",
			"[Ŕááñķ {rank} óóƒ {total} · {name} óóñéé ţŵóó]",
		],
	])("stretches %s as a longer language would", (text, want) => {
		expect(pseudoText(text)).toBe(want);
	});
});
