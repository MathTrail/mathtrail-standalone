import { describe, expect, test } from "vitest";
import { address, alternatesOf, outputPath, parseBase } from "./addresses";
import { readTexts } from "./content";

const base = "https://example.test";

// textsOf are texts for the pages each locale is given, by name.
function textsOf(pages: Record<string, string[]>) {
	return readTexts(
		new Map(
			Object.entries(pages).map(([locale, names]) => [
				locale,
				new Map(
					names.map((name) => [
						`${name}.md`,
						`---\ntitle: ${name}\ndescription: ${name}\n---\n`,
					]),
				),
			]),
		),
		"en",
	);
}

describe("the base URL", () => {
	test.each(["https://mathtrail.app", "http://localhost:8081"])(
		"%s is an origin to build on",
		(raw) => {
			expect(parseBase(raw).origin).toBe(raw);
		},
	);

	test.each([
		["empty", "", "is empty"],
		["ending in a slash", "https://mathtrail.app/", "ends in a slash"],
		["not an address", "mathtrail.app", "is not an address"],
		["neither https nor http", "ftp://mathtrail.app", "neither https nor http"],
		["naming no host", "https:", "is not an address"],
		["with a path", "https://mathtrail.app/site", "more than a scheme"],
		["with a query", "https://mathtrail.app?x=1", "more than a scheme"],
		["with a fragment", "https://mathtrail.app#top", "more than a scheme"],
		["with an empty query", "https://mathtrail.app?", "more than a scheme"],
		["with an empty fragment", "https://mathtrail.app#", "more than a scheme"],
		[
			"with a path a parser drops",
			"https://mathtrail.app/.",
			"more than a scheme",
		],
		["with a host in capitals", "https://MathTrail.app", "otherwise than"],
		[
			"with a name and a password",
			"https://u:p@mathtrail.app",
			"more than a scheme",
		],
		[
			"with the scheme's own port",
			"https://mathtrail.app:443",
			"otherwise than",
		],
	])("is refused when %s", (_, raw, want) => {
		expect(() => parseBase(raw)).toThrow(want);
	});
});

describe("a page's address", () => {
	test.each([
		["en", "index", "/en/", "en/index.html"],
		["en", "privacy", "/en/privacy/", "en/privacy/index.html"],
		["zh-Hans", "terms", "/zh-Hans/terms/", "zh-Hans/terms/index.html"],
		[
			"ru",
			"topics/knights-and-liars",
			"/ru/topics/knights-and-liars/",
			"ru/topics/knights-and-liars/index.html",
		],
	])("of %s %s is %s, served from %s", (locale, name, want, file) => {
		expect(address(locale, name)).toBe(want);
		expect(outputPath(want)).toBe(file);
	});
});

describe("a page's translations", () => {
	const texts = textsOf({
		en: ["index", "privacy"],
		ru: ["index", "privacy"],
	});

	test("of a front page send a reader with no match to the reference locale's, not to the apex", () => {
		expect(alternatesOf(texts, base, "en", "index")).toEqual([
			{ hreflang: "en", url: "https://example.test/en/" },
			{ hreflang: "ru", url: "https://example.test/ru/" },
			{ hreflang: "x-default", url: "https://example.test/en/" },
		]);
	});

	test("of any other page send them to the reference locale's", () => {
		expect(alternatesOf(texts, base, "en", "privacy")).toEqual([
			{ hreflang: "en", url: "https://example.test/en/privacy/" },
			{ hreflang: "ru", url: "https://example.test/ru/privacy/" },
			{ hreflang: "x-default", url: "https://example.test/en/privacy/" },
		]);
	});

	test("of a page below another lead to the same page in every language", () => {
		expect(alternatesOf(texts, base, "en", "topics/sample")).toEqual([
			{ hreflang: "en", url: "https://example.test/en/topics/sample/" },
			{ hreflang: "ru", url: "https://example.test/ru/topics/sample/" },
			{ hreflang: "x-default", url: "https://example.test/en/topics/sample/" },
		]);
	});
});
