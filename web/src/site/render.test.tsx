import { Window } from "happy-dom";
import { afterAll, describe, expect, test } from "vitest";
import type { Dictionary } from "../i18n/words";
import type { Frame } from "./frame";
import type { PageComponent, PageProps } from "./pages";
import { renderSite, type SiteFile } from "./render";
import { type SiteKey, siteDictionaries } from "./words";

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

// text is the source of a document: its front matter, then its Markdown.
function text(title: string, description: string, body: string): string {
	return `---\ntitle: ${title}\ndescription: ${description}\n---\n\n${body}`;
}

// A small site in the shape of the real one, in English and Russian: the
// front page and the two documents, a page a component draws, and a topic's
// page below the topics.
const sources = new Map([
	[
		"en",
		new Map([
			[
				"index.md",
				text(
					"MathTrail",
					"Olympiad maths in your chat.",
					"# MathTrail\n\nRead the [privacy policy](privacy/).\n",
				),
			],
			[
				"privacy.md",
				text(
					"Privacy policy",
					"What is kept.",
					"# Privacy policy\n\nNothing.\n",
				),
			],
			[
				"terms.md",
				text("Terms of use", "The rules.", "# Terms of use\n\nPlay fair.\n"),
			],
			[
				"guide.yaml",
				[
					"title: Guide",
					"description: How to help at home.",
					'lead: "**{count}** topics, and {topics} to read."',
					"topics: the topics",
					"tips:",
					"  - Ask *why*.",
					"  - Wait.",
				].join("\n"),
			],
			[
				"topics/sample.yaml",
				"title: Sample\ndescription: A topic.\ndrawing: |\n  * *\n  |--3--|\n",
			],
		]),
	],
	[
		"ru",
		new Map([
			[
				"index.md",
				text("MathTrail", "Олимпиадная математика в чате.", "# MathTrail\n"),
			],
			[
				"privacy.md",
				text(
					"Политика приватности",
					"Что хранится.",
					"# Политика приватности\n",
				),
			],
			["terms.md", text("Условия", "Правила.", "# Условия\n")],
			[
				"guide.yaml",
				[
					"title: Гид",
					"description: Как помочь дома.",
					'lead: "**{count}** тем, и {topics} — почитать."',
					"topics: темы",
					"tips:",
					"  - Спросите *почему*.",
					"  - Подождите.",
				].join("\n"),
			],
			[
				"topics/sample.yaml",
				"title: Пример\ndescription: Тема.\ndrawing: |\n  * *\n  |--3--|\n",
			],
		]),
	],
]);

// Guide draws the guide's words: a lead with its emphasis and slots, the
// link among them, and the tips as a list.
function Guide({ page }: PageProps) {
	return (
		<section class="s-wrap">
			<p class="lead">
				{page.text("lead", {
					count: 17,
					topics: <a href="../topics/sample/">{page.plain("topics")}</a>,
				})}
			</p>
			<ul class="tips">
				{page.list("tips").map((key) => (
					<li key={key}>{page.text(key)}</li>
				))}
			</ul>
		</section>
	);
}

// Sample draws a topic's text drawing as it is written.
function Sample({ page }: PageProps) {
	return <pre dir="ltr">{page.plain("drawing")}</pre>;
}

const components = new Map<string, PageComponent>([
	["guide", Guide],
	["topics/sample", Sample],
]);

// The site's words, with the two the test's menu needs in both languages.
const dictionaries = new Map([
	...siteDictionaries,
	[
		"en",
		{
			...siteDictionaries.get("en"),
			"nav.guide": "Guide",
			"nav.connect": "Connect",
		},
	],
	[
		"ru",
		{
			...siteDictionaries.get("ru"),
			"nav.guide": "Гид",
			"nav.connect": "Подключить",
		},
	],
]);

// A menu of a page and a section of the front page, and the footer's
// documents.
const frame: Frame = {
	menu: [
		{ page: "guide", label: "nav.guide" as SiteKey },
		{ page: "index", anchor: "connect", label: "nav.connect" as SiteKey },
	],
	footer: ["privacy", "terms"],
};

const site = { base, sources, dictionaries, components, frame };
const files = renderSite(site);

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
	"en/guide/index.html",
	"en/index.html",
	"en/privacy/index.html",
	"en/terms/index.html",
	"en/topics/sample/index.html",
	"index.html",
	"ru/guide/index.html",
	"ru/index.html",
	"ru/privacy/index.html",
	"ru/terms/index.html",
	"ru/topics/sample/index.html",
];

describe("the site's files", () => {
	test("are a page for every text of every locale, the apex, and what a crawler and the host read", () => {
		expect(files.map(({ path }) => path)).toEqual([
			".nojekyll",
			"CNAME",
			...pages.slice(0, 6),
			"robots.txt",
			...pages.slice(6),
			"sitemap.xml",
		]);
	});

	test("are the same bytes every time they are built", () => {
		expect(renderSite(site)).toEqual(files);
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

	test.each(pages)("%s loads nothing from another origin by itself", (path) => {
		const doc = page(files, path);
		const loaded = [
			...attributes(doc, "link[rel=stylesheet], link[rel=icon]", "href"),
			...attributes(doc, "[src]", "src"),
		];

		expect(loaded.filter((address) => !address.startsWith("/"))).toEqual([]);
	});
});

describe("a locale's document", () => {
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
		expect(attributes(privacy, "footer .s-brand", "href")).toEqual(["/ru/"]);
	});

	test("offers every language of the site, the one being read marked", () => {
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

	test("closes with the frame's documents under their titles, the code and the address to write to", () => {
		expect(texts(privacy, ".s-footlinks a")).toEqual([
			"Политика приватности",
			"Условия",
			"GitHub",
			"altedtech.info@gmail.com",
		]);
		expect(attributes(privacy, ".s-footlinks a", "href")).toEqual([
			"/ru/privacy/",
			"/ru/terms/",
			"https://github.com/MathTrail/mathtrail-standalone",
			"mailto:altedtech.info@gmail.com",
		]);
		expect(attributes(privacy, ".s-footlinks", "aria-label")).toEqual([
			"Ссылки",
		]);
	});

	test("says that the English text holds, when it is a translation", () => {
		const legal = (path: string) => texts(page(files, path), ".s-legal");

		expect(legal("ru/privacy/index.html")).toEqual([
			"Это перевод. При расхождении действует английская версия.",
		]);
		expect(legal("en/privacy/index.html")).toEqual([]);
	});
});

describe("a page a component draws", () => {
	const guide = page(files, "ru/guide/index.html");

	test("says what it is from its words, at its own address", () => {
		expect(guide.documentElement.lang).toBe("ru");
		expect(guide.title).toBe("Гид");
		expect(attributes(guide, 'meta[name="description"]', "content")).toEqual([
			"Как помочь дома.",
		]);
		expect(attributes(guide, 'link[rel="canonical"]', "href")).toEqual([
			"https://example.test/ru/guide/",
		]);
		expect(attributes(guide, 'link[rel="alternate"]', "href")).toEqual([
			"https://example.test/en/guide/",
			"https://example.test/ru/guide/",
			"https://example.test/en/guide/",
		]);
	});

	test("draws its words in the frame, emphasis, slots and lists", () => {
		expect(guide.querySelector("main .lead")?.innerHTML).toBe(
			'<strong>17</strong> тем, и <a href="../topics/sample/">темы</a> — почитать.',
		);
		expect(
			[...guide.querySelectorAll("main .tips li")].map(
				(item) => item.innerHTML,
			),
		).toEqual(["Спросите <em>почему</em>.", "Подождите."]);
		expect(guide.querySelectorAll("header")).toHaveLength(1);
		expect(guide.querySelectorAll("footer")).toHaveLength(1);
	});

	test("lives below another page when its words do, drawn as they are written", () => {
		const sample = page(files, "en/topics/sample/index.html");

		expect(attributes(sample, 'link[rel="canonical"]', "href")).toEqual([
			"https://example.test/en/topics/sample/",
		]);
		expect(attributes(sample, ".s-seg a", "href")).toEqual([
			"/en/topics/sample/",
			"/ru/topics/sample/",
		]);
		expect(sample.querySelector("main pre")?.textContent).toBe("* *\n|--3--|");
	});
});

describe("the menu", () => {
	const guide = page(files, "ru/guide/index.html");

	test("leads to the frame's pages and sections, in the page's language", () => {
		expect(texts(guide, ".s-navlinks a")).toEqual(["Гид", "Подключить"]);
		expect(attributes(guide, ".s-navlinks a", "href")).toEqual([
			"/ru/guide/",
			"/ru/#connect",
		]);
		expect(attributes(guide, ".s-navlinks", "aria-label")).toEqual([
			"Разделы сайта",
		]);
	});

	test("marks the page being read, and never a section", () => {
		expect(attributes(guide, ".s-navlinks a", "aria-current")).toEqual([
			"page",
			"",
		]);
		expect(
			attributes(page(files, "ru/index.html"), ".s-navlinks a", "aria-current"),
		).toEqual(["", ""]);
	});

	test("opens behind a button with no script, holding the same links", () => {
		const menu = guide.querySelector("header details.s-menu");

		expect(menu?.querySelector("summary .s-hidden")?.textContent).toBe("Меню");
		expect(
			menu?.querySelector("summary svg")?.getAttribute("aria-hidden"),
		).toBe("true");
		expect(attributes(guide, ".s-menu-list a", "href")).toEqual(
			attributes(guide, ".s-navlinks a", "href"),
		);
		expect(attributes(guide, ".s-menu-list a", "aria-current")).toEqual(
			attributes(guide, ".s-navlinks a", "aria-current"),
		);
	});

	test("is not drawn at all when the frame has no entry", () => {
		const bare = renderSite({ ...site, frame: { ...frame, menu: [] } });
		const doc = page(bare, "ru/guide/index.html");

		expect(doc.querySelector(".s-navlinks")).toBeNull();
		expect(doc.querySelector("details")).toBeNull();
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

	test("offers no switch and no menu, since it is every language's doorway", () => {
		expect(apex.querySelector(".s-seg")).toBeNull();
		expect(apex.querySelector(".s-navlinks")).toBeNull();
		expect(apex.querySelector("details")).toBeNull();
	});

	test("closes with the English documents", () => {
		expect(attributes(apex, ".s-footlinks a", "href")).toEqual([
			"/en/privacy/",
			"/en/terms/",
			"https://github.com/MathTrail/mathtrail-standalone",
			"mailto:altedtech.info@gmail.com",
		]);
	});
});

describe("a language written right to left", () => {
	// Arabic words for the site, made from the English ones so that every key
	// is there, and Arabic texts made from the Russian ones.
	const arabic: Dictionary = Object.fromEntries(
		[...Object.entries(dictionaries.get("en") ?? {})].map(([key]) => [
			key,
			key === "language.name" ? "العربية" : `نص ${key}`,
		]),
	);
	const built = renderSite({
		...site,
		sources: new Map([...sources, ["ar", sources.get("ru") ?? new Map()]]),
		dictionaries: new Map([...dictionaries, ["ar", arabic]]),
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

// Misread draws the guide reading a key its words do not have.
function Misread({ page: words }: PageProps) {
	return <p>{words.text("nope")}</p>;
}

// Skimmed draws the guide and leaves its tips unshown.
function Skimmed({ page: words }: PageProps) {
	return (
		<p>
			{words.text("lead", { count: 1, topics: "x" })}
			{words.plain("topics")}
		</p>
	);
}

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
				sources: new Map([...sources, ["fr", sources.get("en") ?? new Map()]]),
			},
			'locale "fr" has texts and no words',
		],
		[
			"a locale with words and no texts",
			{
				dictionaries: new Map([
					...dictionaries,
					["de", dictionaries.get("en") ?? {}],
				]),
			},
			"the site has words in de — its dictionary de.json — and no texts",
		],
		[
			"words that lack one of the English ones",
			{
				dictionaries: new Map([
					...dictionaries,
					[
						"ru",
						Object.fromEntries(
							Object.entries(dictionaries.get("ru") ?? {}).filter(
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
					...dictionaries,
					["ru", { ...dictionaries.get("ru"), "footer.extra": "Ещё" }],
				]),
			},
			"footer.extra is not among the English words",
		],
		[
			"a menu entry whose page the site does not have",
			{
				frame: {
					...frame,
					menu: [{ page: "why", label: "nav.guide" as SiteKey }],
				},
			},
			"the menu leads to the page why, which the site does not have",
		],
		[
			"a footer that leads to a page the site does not have",
			{ frame: { ...frame, footer: ["privacy", "research"] } },
			"the footer leads to the page research, which the site does not have",
		],
		[
			"words no component draws",
			{ components: new Map([["guide", Guide]]) },
			"the site has words for the page topics/sample, and no component draws it",
		],
		[
			"a page that reads a key its words do not have",
			{ components: new Map([...components, ["guide", Misread]]) },
			"en/guide.yaml: the page reads nope, which the file does not have",
		],
		[
			"a page that never shows some of its words",
			{ components: new Map([...components, ["guide", Skimmed]]) },
			"en/guide.yaml: the page never shows tips.1, tips.2",
		],
		[
			"a text that cannot be read",
			{
				sources: new Map([
					...sources,
					[
						"ru",
						new Map([
							...(sources.get("ru") ?? []),
							["index.md", "# MathTrail\n"],
						]),
					],
				]),
			},
			"ru/index.md: front matter must open",
		],
	])("is refused for %s", (_, change, want) => {
		expect(() => renderSite({ ...site, ...change })).toThrow(want);
	});
});
