import { renderToString } from "preact-render-to-string";
import { describe, expect, test } from "vitest";
import { openWords } from "../i18n/words";
import { Divide, Question } from "./Sketch";
import { SiteWords, siteDictionaries } from "./words";

describe("the sign of division in a drawing", () => {
	test.each([
		["English", "en", ":"],
		["Russian", "ru", ":"],
		["Japanese", "ja", "÷"],
		["Chinese", "zh-Hans", "÷"],
	])("is the one %s writes", (_, locale, sign) => {
		expect(
			renderToString(
				<SiteWords.Provider value={openWords(locale, siteDictionaries)}>
					<Divide />
				</SiteWords.Provider>,
			),
		).toBe(sign);
	});
});

describe("the question mark of a drawing", () => {
	test.each([
		["English", "en", "?"],
		["Japanese", "ja", "?"],
		["Arabic", "ar", "؟"],
	])("is the one %s writes", (_, locale, sign) => {
		expect(
			renderToString(
				<SiteWords.Provider value={openWords(locale, siteDictionaries)}>
					<Question />
				</SiteWords.Provider>,
			),
		).toBe(sign);
	});
});
