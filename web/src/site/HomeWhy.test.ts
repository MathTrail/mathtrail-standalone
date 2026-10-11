import { describe, expect, test } from "vitest";
import { languagesFirst } from "./HomeWhy";

describe("the languages the tile of tasks that never run out names", () => {
	const named = ["en", "ru", "es", "zh"];

	test.each([
		["en", ["en", "ru", "es", "zh"]],
		["ru", ["ru", "en", "es", "zh"]],
		["es", ["es", "en", "ru", "zh"]],
		["zh-Hans", ["zh-Hans", "en", "ru", "es"]],
		["zh-Hant", ["zh-Hant", "en", "ru", "es", "zh"]],
		["pt", ["pt", "en", "ru", "es", "zh"]],
	])("begin on a page in %s with its own language: %j", (locale, want) => {
		expect(languagesFirst(locale, named)).toEqual(want);
	});
});
