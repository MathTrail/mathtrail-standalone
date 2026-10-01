import { describe, expect, test } from "vitest";
import { bodyHTML, byCodeUnits, parseDocument, readTexts } from "./content";

// text is a source with the front matter every page needs.
function text(title: string, body = `# ${title}\n`): string {
	return `---\ntitle: ${title}\ndescription: About ${title}\n---\n\n${body}`;
}

describe("a text's front matter", () => {
	test("gives the title and the description, and keeps the body", () => {
		expect(
			parseDocument(
				"---\ntitle: Home\ndescription: What this is\n---\n\n# Home\n",
			),
		).toEqual({
			title: "Home",
			description: "What this is",
			body: "\n# Home\n",
		});
	});

	test("passes over a blank line between its pairs", () => {
		expect(
			parseDocument("---\ntitle: Home\n\ndescription: D\n---\n"),
		).toMatchObject({ title: "Home", description: "D" });
	});

	test("keeps a colon inside a value", () => {
		expect(
			parseDocument("---\ntitle: MathTrail: a coach\ndescription: D\n---\n")
				.title,
		).toBe("MathTrail: a coach");
	});

	test.each([
		["no front matter at all", "# Home\n", "must open"],
		["front matter never closed", "---\ntitle: Home\n", "never closed"],
		["a line that is not a pair", "---\ntitle Home\n---\n", "not key: value"],
		[
			"a key nobody renders",
			"---\ntitle: Home\ndescription: D\nauthor: Someone\n---\n",
			'key "author"',
		],
		["no title", "---\ndescription: D\n---\n", "no title"],
		["no description", "---\ntitle: Home\n---\n", "no description"],
	])("is refused for %s", (_, source, want) => {
		expect(() => parseDocument(source)).toThrow(want);
	});
});

describe("a text's body", () => {
	test("is CommonMark, with no ids on its headings", () => {
		expect(
			bodyHTML(
				"# One\n\n## Two\n\n### Three\n\nA **strong** word, `code` and [a link](privacy/).\n\n- first\n- second\n",
			),
		).toBe(
			'<h1>One</h1>\n<h2>Two</h2>\n<h3>Three</h3>\n<p>A <strong>strong</strong> word, <code>code</code> and <a href="privacy/">a link</a>.</p>\n<ul>\n<li>first</li>\n<li>second</li>\n</ul>\n',
		);
	});

	test.each([
		["struck-through text", "~~gone~~", "<p>~~gone~~</p>\n"],
		["a bare address", "see www.example.test", "<p>see www.example.test</p>\n"],
		[
			"a table",
			"| a | b |\n|---|---|\n| 1 | 2 |",
			"<p>| a | b |\n|---|---|\n| 1 | 2 |</p>\n",
		],
	])("leaves GitHub's %s as the text it is", (_, body, want) => {
		expect(bodyHTML(body)).toBe(want);
	});

	test.each([
		["a block of markup", "<div>a block</div>\n"],
		["a tag inside a line", "a <b>bold</b> word\n"],
		["a script", "<script>alert(1)</script>\n"],
	])("is refused when it holds %s", (_, body) => {
		expect(() => bodyHTML(body)).toThrow("holds markup");
	});

	test.each([
		["a link that runs a script", "[x](javascript:alert(1))"],
		["one in capitals", "[x](JavaScript:alert(1))"],
		[
			"one written with a reference to a character",
			"[x](&#106;avascript:alert(1))",
		],
		["a colon written as a reference", "[x](javascript&colon;alert(1))"],
		["a scheme split by a space", "[x](<java script:alert(1)>)"],
		[
			"an image carried in its address",
			"![i](data:image/svg+xml;base64,PHN2Zz4=)",
		],
		["a file on the reader's machine", "[x](file:///etc/passwd)"],
	])("is refused when it holds %s", (_, body) => {
		expect(() => bodyHTML(body)).toThrow(
			"may run something rather than lead to a page",
		);
	});

	test.each([
		["a page of the site", "[terms](../terms/)"],
		["another site", "[apps](https://myaccount.google.com/linkedapps)"],
		["an address to write to", "[mail](mailto:help@example.test)"],
	])("leads to %s", (_, body) => {
		expect(bodyHTML(body)).toContain("<a href=");
	});
});

describe("the texts of a site", () => {
	const sources = new Map([
		["ru", new Map([["index", text("Главная")]])],
		[
			"en",
			new Map([
				["terms", text("Terms")],
				["index", text("Home")],
				["privacy", text("Privacy")],
			]),
		],
	]);

	test("are walked in the order of their locales and their names", () => {
		const texts = readTexts(sources, "en");

		expect(texts.locales).toEqual(["en", "ru"]);
		expect(texts.names).toEqual(["index", "privacy", "terms"]);
		expect(texts.pages.get("ru")?.get("index")).toEqual({
			title: "Главная",
			description: "About Главная",
			html: "<h1>Главная</h1>\n",
		});
	});

	test.each([
		["a site with no locale", new Map(), "holds no locale"],
		[
			"a locale with no text",
			new Map([...sources, ["fr", new Map()]]),
			'locale "fr" holds no text',
		],
		[
			"a reference locale with no front page",
			new Map([["en", new Map([["terms", text("Terms")]])]]),
			'reference locale "en" has no index.md',
		],
		[
			"a text that cannot be read, naming its file",
			new Map([...sources, ["ru", new Map([["index", "# Главная\n"]])]]),
			"ru/index.md: front matter must open",
		],
		[
			"a text holding markup, naming its file",
			new Map([
				...sources,
				["ru", new Map([["index", text("Главная", "<p>hi</p>\n")]])],
			]),
			"ru/index.md: the text holds markup",
		],
	])("are refused for %s", (_, broken, want) => {
		expect(() => readTexts(broken, "en")).toThrow(want);
	});
});

describe("the order of a build", () => {
	test("is by code units, wherever the build runs", () => {
		expect(["ru", "en", "zh-Hans", "en-GB", "Z"].sort(byCodeUnits)).toEqual([
			"Z",
			"en",
			"en-GB",
			"ru",
			"zh-Hans",
		]);
		expect(byCodeUnits("en", "en")).toBe(0);
	});
});
