import { describe, expect, test } from "vitest";
import { disagreements } from "../i18n/dictionaries";
import { siteDictionaries } from "./words";

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
