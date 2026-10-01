import { Window } from "happy-dom";
import { afterAll, describe, expect, test } from "vitest";
import type { Dictionary } from "../i18n/words";
import { renderSite, type SiteFile } from "./render";
import { siteDictionaries } from "./words";

const base = "https://example.test";

// browser reads the built pages and loads nothing they link to: the
// stylesheets live on no server a test can reach, and what is checked here is
// the pages themselves.
const browser = new Window({
	settings: {
		disableCSSFileLoading: true,
		disableJavaScriptFileLoading: true,
		handleDisabledFileLoadingAsSuccess: true,
	},
});

afterAll(async () => {
	await browser.happyDOM.close();
});

// text is the source of a page: its front matter, then its Markdown.
function text(title: string, description: string, body: string): string {
	return `---\ntitle: ${title}\ndescription: ${description}\n---\n\n${body}`;
}

// A small site in the shape of the real one: English has every page, Russian
// lacks the terms.
const sources = new Map([
	[
		"en",
		new Map([
			[
				"index",
				text(
					"MathTrail",
					"Olympiad maths in your chat.",
					"# MathTrail\n\nRead the [privacy policy](privacy/).\n",
				),
			],
			[
				"privacy",
				text(
					"Privacy policy",
					"What is kept.",
					"# Privacy policy\n\nNothing.\n",
				),
			],
			[
				"terms",
				text("Terms of use", "The rules.", "# Terms of use\n\nPlay fair.\n"),
			],
		]),
	],
	[
		"ru",
		new Map([
			[
				"index",
				text("MathTrail", "Олимпиадная математика в чате.", "# MathTrail\n"),
			],
			[
				"privacy",
				text(
					"Политика приватности",
					"Что хранится.",
					"# Политика приватности\n",
				),
			],
		]),
	],
]);

const files = renderSite({ base, sources });

// page is the built file at path, read as a document.
function page(built: readonly SiteFile[], path: string) {
	const file = built.find((candidate) => candidate.path === path);
	if (file === undefined) {
		throw new Error(`the site has no ${path}`);
	}
	return new browser.DOMParser().parseFromString(file.data, "text/html");
}

// Page is a built page, read as a document.
type Page = ReturnType<typeof page>;

// attributes are the values of name on every element selector finds.
function attributes(doc: Page, selector: string, name: string): string[] {
	return [...doc.querySelectorAll(selector)].map(
		(element) => element.getAttribute(name) ?? "",
	);
}

// texts are the text of every element selector finds.
function texts(doc: Page, selector: string): string[] {
	return [...doc.querySelectorAll(selector)].map(
		(element) => element.textContent ?? "",
	);
}

const pages = [
	"en/index.html",
	"en/privacy/index.html",
	"en/terms/index.html",
	"index.html",
	"ru/index.html",
	"ru/privacy/index.html",
];

describe("the site's files", () => {
	test("are a page for every text, the apex, and what a crawler and the host read", () => {
		expect(files.map(({ path }) => path)).toEqual([
			".nojekyll",
			"CNAME",
			...pages.slice(0, 4),
			"robots.txt",
			...pages.slice(4),
			"sitemap.xml",
		]);
	});

	test("are the same bytes every time they are built", () => {
		expect(renderSite({ base, sources })).toEqual(files);
	});
});

describe("every page", () => {
	test.each(pages)(
		"%s opens as a document a browser lays out by the standard",
		(path) => {
			const file = files.find((candidate) => candidate.path === path);

			expect(file?.data.startsWith("<!DOCTYPE html>\n<html")).toBe(true);
		},
	);

	test.each(pages)(
		"%s is light, and loads the tokens before the styles that read them",
		(path) => {
			const doc = page(files, path);

			expect(doc.documentElement.getAttribute("data-theme")).toBe("light");
			expect(attributes(doc, 'meta[name="color-scheme"]', "content")).toEqual([
				"light",
			]);
			expect(attributes(doc, 'link[rel="stylesheet"]', "href")).toEqual([
				"/assets/tokens.css",
				"/assets/style.css",
			]);
			expect(attributes(doc, 'link[rel="icon"]', "href")).toEqual([
				"/assets/favicon.svg",
			]);
		},
	);

	test.each(pages)("%s runs no script", (path) => {
		const file = files.find((candidate) => candidate.path === path);

		expect(file?.data).not.toContain("<script");
	});
});

describe("a locale's page", () => {
	const privacy = page(files, "ru/privacy/index.html");

	test("says what it is, in its language, at its own address", () => {
		expect(privacy.documentElement.lang).toBe("ru");
		expect(privacy.documentElement.dir).toBe("ltr");
		expect(privacy.title).toBe("Политика приватности");
		expect(attributes(privacy, 'meta[name="description"]', "content")).toEqual([
			"Что хранится.",
		]);
		expect(attributes(privacy, 'link[rel="canonical"]', "href")).toEqual([
			"https://example.test/ru/privacy/",
		]);
		expect(attributes(privacy, 'meta[property="og:url"]', "content")).toEqual([
			"https://example.test/ru/privacy/",
		]);
		expect(
			attributes(privacy, 'meta[property="og:site_name"]', "content"),
		).toEqual(["MathTrail"]);
	});

	test("names every translation, and the page for a reader with none", () => {
		expect(attributes(privacy, 'link[rel="alternate"]', "hreflang")).toEqual([
			"en",
			"ru",
			"x-default",
		]);
		expect(attributes(privacy, 'link[rel="alternate"]', "href")).toEqual([
			"https://example.test/en/privacy/",
			"https://example.test/ru/privacy/",
			"https://example.test/en/privacy/",
		]);
	});

	test("carries its text, made from Markdown", () => {
		expect(
			page(files, "en/index.html").querySelector("main article")?.innerHTML,
		).toBe(
			'<h1>MathTrail</h1>\n<p>Read the <a href="privacy/">privacy policy</a>.</p>\n',
		);
	});

	test("leads home to its own language's front page", () => {
		expect(attributes(privacy, "header .s-brand", "href")).toEqual(["/ru/"]);
	});

	test("offers the languages it exists in, the one being read marked", () => {
		expect(attributes(privacy, ".s-seg a", "href")).toEqual([
			"/en/privacy/",
			"/ru/privacy/",
		]);
		expect(texts(privacy, ".s-seg a")).toEqual(["EN", "RU"]);
		expect(attributes(privacy, ".s-seg a", "title")).toEqual([
			"English",
			"Русский",
		]);
		expect(attributes(privacy, ".s-seg a", "lang")).toEqual(["en", "ru"]);
		expect(attributes(privacy, ".s-seg a", "aria-current")).toEqual([
			"",
			"page",
		]);
		expect(attributes(privacy, ".s-seg", "aria-label")).toEqual(["Язык"]);
	});

	test("offers no language that lacks it", () => {
		expect(
			attributes(page(files, "en/terms/index.html"), ".s-seg a", "href"),
		).toEqual(["/en/terms/"]);
	});

	test("closes with the documents of its language, under their titles", () => {
		expect(texts(privacy, ".s-footlinks a")).toEqual([
			"Политика приватности",
			"GitHub",
		]);
		expect(attributes(privacy, ".s-footlinks a", "href")).toEqual([
			"/ru/privacy/",
			"https://github.com/MathTrail/mathtrail-standalone",
		]);
		expect(attributes(privacy, ".s-footlinks", "aria-label")).toEqual([
			"Документы",
		]);
		expect(
			texts(page(files, "en/privacy/index.html"), ".s-footlinks a"),
		).toEqual(["Privacy policy", "Terms of use", "GitHub"]);
	});

	test("says that the English text holds, when it is a translation", () => {
		const legal = (path: string) => texts(page(files, path), ".s-legal");

		expect(legal("ru/privacy/index.html")).toEqual([
			"Это перевод. При расхождении действует английская версия.",
		]);
		expect(legal("en/privacy/index.html")).toEqual([]);
	});
});

describe("the apex", () => {
	const apex = page(files, "index.html");

	test("is the reference locale's front page in what it tells a search engine", () => {
		expect(apex.documentElement.lang).toBe("en");
		expect(apex.title).toBe("MathTrail");
		expect(attributes(apex, 'meta[name="description"]', "content")).toEqual([
			"Olympiad maths in your chat.",
		]);
		expect(attributes(apex, 'link[rel="canonical"]', "href")).toEqual([
			"https://example.test/",
		]);
		expect(attributes(apex, 'link[rel="alternate"]', "href")).toEqual([
			"https://example.test/en/",
			"https://example.test/ru/",
			"https://example.test/",
		]);
	});

	test("hands the reader every language, each in its own name", () => {
		expect(texts(apex, ".s-choices a")).toEqual(["English", "Русский"]);
		expect(attributes(apex, ".s-choices a", "href")).toEqual(["/en/", "/ru/"]);
		expect(attributes(apex, ".s-choices a", "lang")).toEqual(["en", "ru"]);
		expect(texts(apex, ".s-lead")).toEqual(["Olympiad maths in your chat."]);
	});

	test("offers no switch, since it is every language's", () => {
		expect(apex.querySelector(".s-seg")).toBeNull();
	});
});

describe("a language written right to left", () => {
	// Arabic words for the site, made from the English ones so that every key
	// is there.
	const arabic: Dictionary = Object.fromEntries(
		[...Object.entries(siteDictionaries.get("en") ?? {})].map(([key]) => [
			key,
			key === "language.name" ? "العربية" : `نص ${key}`,
		]),
	);
	const built = renderSite({
		base,
		sources: new Map([
			...sources,
			[
				"ar",
				new Map([["index", text("MathTrail", "رياضيات.", "# MathTrail\n")]]),
			],
		]),
		dictionaries: new Map([...siteDictionaries, ["ar", arabic]]),
	});

	test("lays its pages out right to left", () => {
		expect(page(built, "ar/index.html").documentElement.dir).toBe("rtl");
	});

	test("names itself on the apex the way it runs", () => {
		const apex = page(built, "index.html");

		expect(texts(apex, ".s-choices a")).toEqual([
			"العربية",
			"English",
			"Русский",
		]);
		expect(attributes(apex, ".s-choices a", "dir")).toEqual([
			"rtl",
			"ltr",
			"ltr",
		]);
	});
});

describe("a site that cannot be built", () => {
	test.each([
		[
			"a base URL that is no origin",
			{ base: "https://example.test/" },
			"ends in a slash",
		],
		[
			"a locale with texts and no words",
			{
				sources: new Map([
					...sources,
					["fr", new Map([["index", text("A", "B", "")]])],
				]),
			},
			'locale "fr" has texts and no words',
		],
		[
			"words that lack one of the English ones",
			{
				dictionaries: new Map([
					...siteDictionaries,
					[
						"ru",
						Object.fromEntries(
							Object.entries(siteDictionaries.get("ru") ?? {}).filter(
								([key]) => key !== "footer.translation",
							),
						),
					],
				]),
			},
			"footer.translation is missing",
		],
		[
			"words English does not have",
			{
				dictionaries: new Map([
					...siteDictionaries,
					["ru", { ...siteDictionaries.get("ru"), "footer.extra": "Ещё" }],
				]),
			},
			"footer.extra is not among the English words",
		],
		[
			"a text that cannot be read",
			{
				sources: new Map([
					...sources,
					["ru", new Map([["index", "# MathTrail\n"]])],
				]),
			},
			"ru/index.md: front matter must open",
		],
	])("is refused for %s", (_, change, want) => {
		expect(() => renderSite({ base, sources, ...change })).toThrow(want);
	});
});
