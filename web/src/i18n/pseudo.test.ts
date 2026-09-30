import { describe, expect, test } from "vitest";
import english from "../../locales/en.json";
import { pseudoLocale, pseudoText, pseudoWords } from "./pseudo";
import { type Dictionary, placeholder } from "./words";

// literal is a text without its slots, whose length the words filled in
// decide.
const literal = (text: string) => [...text.replace(placeholder, "")].length;

// textsOf are the texts of every wording of words, by key and form.
function textsOf(words: Dictionary): [string, string][] {
	return Object.entries(words).flatMap(([key, wording]) =>
		typeof wording === "string"
			? [[key, wording] as [string, string]]
			: Object.entries(wording).map(
					([form, text]) => [`${key}.${form}`, text] as [string, string],
				),
	);
}

describe("the pseudo-language", () => {
	const pseudo = pseudoWords(english);
	const texts = new Map(textsOf(english));

	test("is English by its tag, so it counts and runs as English does", () => {
		const locale = new Intl.Locale(pseudoLocale);

		expect(locale.language).toBe("en");
		expect(locale.baseName).toBe(pseudoLocale);
	});

	test.each(textsOf(pseudo))(
		"says %s twice and ten characters longer than English, in brackets",
		(key, text) => {
			const said = texts.get(key) ?? "";

			expect(literal(text)).toBeGreaterThanOrEqual(literal(said) * 2);
			expect(literal(text)).toBeGreaterThanOrEqual(literal(said) + 10);
			expect(text.startsWith("[")).toBe(true);
			expect(text.endsWith("]")).toBe(true);
		},
	);

	test("keeps every slot, and every plural wording's forms", () => {
		for (const [key, wording] of Object.entries(english)) {
			const said = pseudo[key];
			if (typeof wording === "string") {
				expect(said, key).toBeTypeOf("string");
				continue;
			}
			expect(Object.keys(said ?? {}), key).toEqual(Object.keys(wording));
		}
		for (const [key, text] of textsOf(pseudo)) {
			const slots = (words: string) =>
				[...words.matchAll(placeholder)].map(([slot]) => slot);
			expect(slots(text), key).toEqual(slots(texts.get(key) ?? ""));
		}
	});

	test("shows no letter of English as it was", () => {
		for (const [key, text] of textsOf(pseudo)) {
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
