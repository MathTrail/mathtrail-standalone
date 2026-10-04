import { describe, expect, test } from "vitest";
import { bodyHTML, parseDocument, readTexts } from "./content";

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
	// words are a page's words with the title and description every page has.
	const words = (title: string, rest = "") =>
		`title: ${title}\ndescription: About ${title}\n${rest}`;
	const sources = new Map([
		[
			"ru",
			new Map([
				["index.md", text("Главная")],
				["privacy.md", text("Приватность")],
				["topics/sample.yaml", words("Пример", "lead: Текст.\n")],
			]),
		],
		[
			"en",
			new Map([
				["topics/sample.yaml", words("Sample", "lead: Text.\n")],
				["index.md", text("Home")],
				["privacy.md", text("Privacy")],
			]),
		],
	]);

	test("are walked in the order of their locales and their names", () => {
		const texts = readTexts(sources, "en");

		expect(texts.locales).toEqual(["en", "ru"]);
		expect(texts.names).toEqual(["index", "privacy", "topics/sample"]);
		expect(texts.front).toEqual({ title: "Home", description: "About Home" });
	});

	test("are a document for a .md file and a page's words for a .yaml one", () => {
		const texts = readTexts(sources, "en");

		expect(texts.pages.get("ru")?.get("index")).toEqual({
			kind: "document",
			title: "Главная",
			description: "About Главная",
			html: "<h1>Главная</h1>\n",
		});
		expect(texts.pages.get("ru")?.get("topics/sample")).toEqual({
			kind: "words",
			words: new Map([
				["title", "Пример"],
				["description", "About Пример"],
				["lead", "Текст."],
			]),
		});
	});

	// replacing is the sources with one locale's texts replaced by files.
	const replacing = (locale: string, files: [string, string][]) =>
		new Map([...sources, [locale, new Map(files)]]);
	// russian are the Russian texts, with the files given written after them.
	const russian = (...changed: [string, string][]): [string, string][] => [
		["index.md", text("Главная")],
		["privacy.md", text("Приватность")],
		["topics/sample.yaml", words("Пример", "lead: Текст.\n")],
		...changed,
	];

	test.each([
		["a site with no locale", new Map(), "holds no locale"],
		[
			"a locale with no text",
			new Map([...sources, ["fr", new Map()]]),
			'locale "fr" holds no text',
		],
		[
			"a reference locale with no front page",
			new Map([["en", new Map([["terms.md", text("Terms")]])]]),
			'reference locale "en" has no index page',
		],
		[
			"a page some locale lacks",
			replacing("ru", [
				["index.md", text("Главная")],
				["topics/sample.yaml", words("Пример", "lead: Текст.\n")],
			]),
			"the page privacy is missing in ru: a page is in every language of the site or in none",
		],
		[
			"a page that is a document in one locale and words in another",
			replacing("ru", [
				["index.md", text("Главная")],
				["privacy.yaml", words("Приватность")],
				["topics/sample.yaml", words("Пример", "lead: Текст.\n")],
			]),
			"the page privacy is a document in one language and a component's words in another",
		],
		[
			"a page two files write",
			replacing("ru", russian(["privacy.yaml", words("Приватность")])),
			"ru/privacy.yaml: the page privacy has a text already",
		],
		[
			"a file neither a document nor words",
			replacing("ru", russian(["notes.txt", "a note"])),
			"ru/notes.txt: a text is a .md or a .yaml file",
		],
		[
			"a page named otherwise than its address",
			replacing("ru", russian(["Topics/Sample.yaml", words("Пример")])),
			"ru/Topics/Sample.yaml: a text is a .md or a .yaml file, named in lowercase words",
		],
		[
			"words that disagree with the English, naming their file",
			replacing("ru", russian(["topics/sample.yaml", words("Пример")])),
			"ru/topics/sample.yaml: the words disagree with the English: lead is missing",
		],
		[
			"words with no title",
			replacing(
				"ru",
				russian(["topics/sample.yaml", "description: D\nlead: L\n"]),
			),
			"ru/topics/sample.yaml: a page's words give its title",
		],
		[
			"a title with a slot",
			replacing(
				"ru",
				russian(["topics/sample.yaml", words("Тема {name}", "lead: Текст.\n")]),
			),
			"ru/topics/sample.yaml: a page's title is said as it is written, with no slot",
		],
		[
			"words that cannot be read, naming their file",
			replacing("ru", russian(["topics/sample.yaml", "title: [A\n"])),
			"ru/topics/sample.yaml: Flow sequence in block collection",
		],
		[
			"a text that cannot be read, naming its file",
			replacing("ru", russian(["index.md", "# Главная\n"])),
			"ru/index.md: front matter must open",
		],
		[
			"a text holding markup, naming its file",
			replacing("ru", russian(["index.md", text("Главная", "<p>hi</p>\n")])),
			"ru/index.md: the text holds markup",
		],
	])("are refused for %s", (_, broken, want) => {
		expect(() => readTexts(broken, "en")).toThrow(want);
	});
});
