import { readFileSync } from "node:fs";
import { dirname, join } from "node:path";
import { fileURLToPath } from "node:url";
import { Window } from "happy-dom";
import { afterAll, describe, expect, test } from "vitest";
import file from "../../../site/data.json";
import fixture from "../../../site/research/testdata/research.json";
import { letters } from "../widget/choices";
import { readSiteData, type SiteData, siteCatalog } from "./data";
import type { Frame } from "./frame";
import { sitePages } from "./pages";
import { researchSections } from "./ResearchPage";
import { secondRun } from "./ResearchTheses";
import { renderSite } from "./render";
import type { Research } from "./research";
import { handTyped } from "./testing/handtyped";

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

// textOf is a file of the repository, by its path from the repository's root:
// the page's own words, and the paper's sources its title is held to.
const textOf = (...path: string[]) =>
	readFileSync(
		join(dirname(fileURLToPath(import.meta.url)), "..", "..", "..", ...path),
		"utf8",
	);
const english = textOf("site", "content", "en", "research.yaml");
const russian = textOf("site", "content", "ru", "research.yaml");
const tex = textOf("research", "paper-a", "main.tex");
const russianDraft = textOf("research", "paper-a", "draft.ru.md");

// names are the names on the page that carry digits of their own.
const names = ["XChaCha20-Poly1305"];

// withPaper is the fixture as the bench writes it, the paper's PDF named;
// withoutPaper is the same with no file of the paper shipped.
const withPaper = dataOf(fixture);
const withoutPaper = dataOf({
	...fixture,
	paper: { ...fixture.paper, files: [] },
});

// dataOf is the site's data with the numbers of the page Research, over the
// service's catalog with no topic's page published: the site of these tests
// has the one page.
function dataOf(research: unknown): SiteData {
	return readSiteData(
		{
			...siteCatalog,
			topics: siteCatalog.topics.map((topic) => ({
				...topic,
				site_page: false,
			})),
		},
		{
			groups: file.groups,
			examples: {},
			progress: file.progress,
			research: file.research,
		},
		research,
	);
}

// document is a document's text under its title.
const document = (title: string) =>
	`---\ntitle: ${title}\ndescription: D\n---\n\n# ${title}\n`;

// sources are the site's texts, the page's own words in both its languages.
const sources = new Map(
	Object.entries({ en: english, ru: russian }).map(([locale, words]) => [
		locale,
		new Map([
			["index.md", document("Home")],
			["privacy.md", document("Privacy")],
			["terms.md", document("Terms")],
			["research.yaml", words],
		]),
	]),
);

// A menu of the one page this site has.
const frame: Frame = {
	menu: [{ page: "research", label: "nav.research" }],
	footer: ["privacy", "terms"],
};

// pagesOf are the pages of the site drawn from data, the page Research in
// every language, read as documents.
function pagesOf(data: SiteData): Map<string, Document> {
	const page = sitePages(data).get("research");
	if (page === undefined) {
		throw new Error("the site draws no page Research");
	}
	const files = renderSite({
		base: "https://example.test",
		sources,
		pages: new Map([["research", page]]),
		frame,
		data,
	});
	return new Map(
		["en", "ru"].map((locale) => [
			locale,
			new browser.DOMParser().parseFromString(
				files.find(({ path }) => path === `${locale}/research/index.html`)
					?.data ?? "",
				"text/html",
			) as unknown as Document,
		]),
	);
}

const pages = pagesOf(withPaper);
const page = pages.get("en") as Document;
const russianPage = pages.get("ru") as Document;
const unshipped = pagesOf(withoutPaper);

// values are the values of every element selector finds on a page.
const values = (on: Document, selector: string) =>
	[...on.querySelectorAll(selector)].map(
		(element) => element.getAttribute("value") ?? "",
	);

// texts are the texts of every element selector finds on a page.
const texts = (on: Document, selector: string) =>
	[...on.querySelectorAll(selector)].map((element) =>
		(element.textContent ?? "").replace(/\s+/g, " ").trim(),
	);

// research is what the page draws, read from the fixture.
const research = withPaper.research as Research;

describe("the page Research", () => {
	test("reads in full with no script", () => {
		for (const each of pages.values()) {
			expect(each.querySelector("script")).toBeNull();
		}
	});

	test("is headed by the paper's title, in English as the paper gives it and in Russian as its Russian draft does", () => {
		const title = /\\title\{([^}]*)\}/.exec(tex)?.[1] ?? "";
		const russianTitle = /^# (.+)$/m.exec(russianDraft)?.[1] ?? "";

		expect(texts(page, "h1")).toEqual([
			title.replace(/\\\\\s*/g, " ").replace(/\s+/g, " "),
		]);
		expect(texts(russianPage, "h1")).toEqual([russianTitle]);
	});

	test("counts its chips and its lead from the data", () => {
		const { product } = research;
		expect(values(page, ".s-chip data")).toEqual(
			[
				product.reference_tasks.total,
				product.checks,
				product.topics,
				product.traps,
				product.grades.first,
				product.grades.last,
			].map(String),
		);
		expect(values(page, ".s-lead data")).toEqual([String(product.checks)]);
	});

	test.each([
		[1, "прошедшие 1 проверку"],
		[2, "прошедшие 2 проверки"],
		[5, "прошедшие 5 проверок"],
		[21, "прошедшие 21 проверку"],
	])(
		"words the checks a task passes as Russian counts them, here %s",
		(checks, said) => {
			const moved = pagesOf(
				dataOf({ ...fixture, product: { ...fixture.product, checks } }),
			).get("ru") as Document;

			expect(texts(moved, ".s-lead").join(" ")).toContain(said);
		},
	);

	test("offers the paper's PDF with its pages when the data has one, the English file in every language", () => {
		const [paper] = research.paper.files;
		for (const each of pages.values()) {
			const buttons = [...each.querySelectorAll('a[type="application/pdf"]')];
			expect(buttons.map((button) => button.getAttribute("href"))).toEqual([
				paper?.path,
				paper?.path,
			]);
			expect(buttons.map((button) => button.getAttribute("hreflang"))).toEqual([
				"en",
				"en",
			]);
			expect(
				values(each, 'a[type="application/pdf"] data').every(
					(value) => value === String(paper?.pages),
				),
			).toBe(true);
		}
	});

	test("offers no PDF while the site ships none, and says when it comes", () => {
		for (const each of unshipped.values()) {
			expect(each.querySelector('a[type="application/pdf"]')).toBeNull();
		}
		expect(
			texts(unshipped.get("en") as Document, "#paper .s-ask-lead"),
		).toEqual([
			"Its methods, tables, protocols, limitations and ethics. The PDF comes once the paper names its authors and its archive.",
		]);
	});

	test("leads to the code from the first screen and from the paper's part", () => {
		const code = [...page.querySelectorAll("main a")].filter(
			(link) =>
				link.getAttribute("href") ===
				"https://github.com/MathTrail/mathtrail-standalone",
		);
		expect(code).toHaveLength(2);
	});

	test("draws the solver's second run from the data's relabelling", () => {
		const { relabelled } = research.product;
		expect(secondRun([3, 4, 5, 6, 12], ["C", "D", "E", "A", "B"])).toEqual([
			6, 12, 3, 4, 5,
		]);
		const rows = [...page.querySelectorAll(".s-research-runs tbody tr")];
		const second = [...(rows[1]?.querySelectorAll("td [data-given]") ?? [])];
		expect(second.map((cell) => cell.textContent)).toEqual(
			secondRun([3, 4, 5, 6, 12], relabelled).map(String),
		);
		expect(rows[1]?.querySelector("td:last-child")?.textContent).toBe(
			relabelled[letters.indexOf("C")],
		);
	});

	test("says the shift, the corridor, the guess and the trial series through the data, in every language", () => {
		const { product } = research;
		for (const each of pages.values()) {
			const theses = [...each.querySelectorAll(".s-research-thesis")];
			expect(values(theses[1] as unknown as Document, "p data")).toContain(
				String(product.relabel),
			);
			expect(values(theses[3] as unknown as Document, "p data")).toEqual(
				expect.arrayContaining(
					[
						product.corridor.middle,
						product.corridor.low,
						product.corridor.high,
						product.trial_answers,
						product.guess,
						product.options,
					].map(String),
				),
			);
		}
	});

	test("pins its drawings left to right in every language", () => {
		for (const selector of [
			".s-research-runs",
			".s-research-axis",
			".s-research-formula",
			".s-research-chart",
		]) {
			const drawn = [...russianPage.querySelectorAll(selector)];
			expect(drawn.length).toBeGreaterThan(0);
			for (const element of drawn) {
				expect(element.getAttribute("dir")).toBe("ltr");
			}
		}
	});

	test("shows every row of the goals in the data's order, each rule's number with its interval and the goal's mark", () => {
		const rows = [
			...page.querySelectorAll(".s-research-goals tbody tr"),
		].filter((row) => row.querySelector('th[scope="row"]') !== null);
		expect(rows).toHaveLength(research.bench.rows.length);
		research.bench.rows.forEach((row, at) => {
			const cells = [...(rows[at]?.querySelectorAll("td") ?? [])];
			const shown = (cell: number) =>
				[...(cells[cell]?.querySelectorAll("data") ?? [])].map((element) =>
					element.getAttribute("value"),
				);
			expect(shown(0).slice(0, 3)).toEqual(
				[
					row.values.earlier.value,
					row.values.earlier.low,
					row.values.earlier.high,
				].map(String),
			);
			expect(shown(1).slice(0, 3)).toEqual(
				[
					row.values.service.value,
					row.values.service.low,
					row.values.service.high,
				].map(String),
			);
			const marks = (cell: number) =>
				[...(cells[cell]?.querySelectorAll(".s-pill") ?? [])].map(
					(pill) => pill.className,
				);
			expect(marks(0)).toEqual(
				row.kind === "goal"
					? [`s-pill s-research-mark s-research-${row.values.earlier.mark}`]
					: [],
			);
			expect(marks(1)).toEqual(
				row.kind === "goal"
					? [`s-pill s-research-mark s-research-${row.values.service.mark}`]
					: [],
			);
		});
	});

	test.each([
		[1.5, "в 1,5 раза больше, чем у сервиса"],
		[2, "в 2 раза больше, чем у сервиса"],
		[5, "в 5 раз больше, чем у сервиса"],
	])(
		"words a goal of %s times the service's as Russian counts it",
		(times, said) => {
			const moved = pagesOf(
				dataOf({
					...fixture,
					bench: {
						...fixture.bench,
						goal_parameters: {
							...fixture.bench.goal_parameters,
							late_times: times,
						},
					},
				}),
			).get("ru") as Document;

			expect(texts(moved, ".s-research-rule")).toContain(said);
		},
	);

	test("names every mark and every ceiling in its legend, whatever marks the rows hold", () => {
		expect(texts(page, ".s-research-legend dt")).toEqual([
			"Reached",
			"On the edge",
			"Not reached",
			"Baseline",
			"The oracle",
			"Perfection",
		]);
	});

	test("hides its bars from a screen reader, which reads the cells", () => {
		const charts = [...page.querySelectorAll(".s-research-chart")];
		expect(charts).toHaveLength(research.bench.rows.length);
		for (const chart of charts) {
			expect(chart.getAttribute("aria-hidden")).toBe("true");
		}
	});

	test("tells the commit, the date and the run its numbers come from", () => {
		expect(texts(page, ".s-research-provenance code")).toEqual([
			research.built_from.commit.slice(0, 12),
			research.paper.commit.slice(0, 12),
		]);
		expect(
			page
				.querySelector(".s-research-provenance time")
				?.getAttribute("datetime"),
		).toBe(research.built_from.date);
		expect(values(page, ".s-research-provenance data")).toEqual(
			expect.arrayContaining(
				[
					research.bench.seed,
					research.bench.children,
					research.bench.answers,
				].map(String),
			),
		);
	});

	test("lists the authors in the data's order, each with the year they died", () => {
		expect(values(page, ".s-research-died data")).toEqual(
			file.research.sources.map(({ died }) => String(died)),
		);
		expect(texts(page, ".s-research-died")).toContain("died 1942");
	});

	test("draws the reference tasks by level as a bar of their counts", () => {
		const parts = [...page.querySelectorAll(".s-research-levels li")];
		expect(
			parts.map((part) => part.getAttribute("style")?.replace(/[\s;]/g, "")),
		).toEqual(research.levels.map(({ count }) => `flex-grow:${count}`));
	});

	test("says in its live block that the numbers are still to come, and shows none", () => {
		const live = page.querySelector(`#${researchSections.live}`);
		expect(live?.querySelector(".s-badge")?.textContent).toBe("Coming");
		expect(live?.querySelectorAll("data").length).toBe(0);
	});

	test("has the anchors other pages lead to", () => {
		for (const id of Object.values(researchSections)) {
			expect(page.getElementById(id), id).not.toBeNull();
		}
	});
});

describe("the numbers on the page Research", () => {
	test("are typed into none of its words, in any language, but in the names that carry their own", () => {
		for (const words of [english, russian]) {
			const typed = words
				.split("\n")
				.filter((line) =>
					/\p{Nd}/u.test(
						names.reduce((left, name) => left.split(name).join(""), line),
					),
				);
			expect(typed).toEqual([]);
		}
	});

	test("show no digit that did not come from the data, in any language", () => {
		const commits = new Set([
			research.built_from.commit,
			research.paper.commit,
			research.built_from.date,
		]);
		for (const [locale, each] of [...pages, ...unshipped]) {
			expect(handTyped(each, { names, values: commits }), locale).toEqual([]);
		}
	});

	test("are each a number of the data", () => {
		const known = new Set<string>();
		const collect = (value: unknown): void => {
			if (typeof value === "number") {
				known.add(String(value));
			} else if (typeof value === "object" && value !== null) {
				Object.values(value).forEach(collect);
			}
		};
		collect(fixture);
		collect(file.research);
		collect(research.levels);
		for (const each of pages.values()) {
			for (const value of values(each, "data")) {
				expect(known.has(value), value).toBe(true);
			}
		}
	});

	test("all move when the data moves", () => {
		const moved = (value: unknown): unknown => {
			if (typeof value === "number") {
				return value + 1;
			}
			if (Array.isArray(value)) {
				return value.map(moved);
			}
			if (typeof value === "object" && value !== null) {
				return Object.fromEntries(
					Object.entries(value).map(([key, inner]) => [key, moved(inner)]),
				);
			}
			return value;
		};
		const before = values(page, "data");
		const after = values(
			pagesOf({ ...withPaper, research: moved(research) as Research }).get(
				"en",
			) as Document,
			"data",
		);
		expect(after).toHaveLength(before.length);
		after.forEach((value, at) => {
			expect(value, `the number ${before[at]}`).not.toBe(before[at]);
		});
	});
});

describe("the check of numbers typed by hand", () => {
	// pageWith is a page whose main holds markup, under a head of its own.
	const pageWith = (
		main: string,
		head = '<title>T</title><meta name="description" content="D">',
	) =>
		new browser.DOMParser().parseFromString(
			`<html><head>${head}</head><body><main>${main}</main></body></html>`,
			"text/html",
		) as unknown as Document;
	const rules = {
		names,
		values: new Set(["0123456789abcdef", "2026-10-05T00:00:00Z"]),
	};

	test.each([
		["a text", "<p>Nine of 9 checks</p>"],
		["an aria-label", '<p aria-label="2 of them">two</p>'],
		["a title", '<abbr title="3 checks">three</abbr>'],
		["an alt", '<img alt="4 options">'],
		["a digit of another script", "<p>٣ ловушки</p>"],
		["a commit that is not the data's", "<code>fedcba9876</code>"],
		["a code too short to name a commit", "<code>01234</code>"],
		[
			"a date that is not the data's",
			'<time datetime="2020-01-01">2020</time>',
		],
	])("finds a digit typed into %s", (_, main) => {
		expect(handTyped(pageWith(main), rules)).not.toEqual([]);
	});

	test.each([
		["its title", "<title>Research 2026</title>"],
		["its description", '<meta name="description" content="17 topics">'],
		[
			"the title a shared link shows",
			'<meta property="og:title" content="Research 2026">',
		],
		[
			"the description a shared link shows",
			'<meta property="og:description" content="17 topics">',
		],
		[
			"the words a shared link gives its picture",
			'<meta property="og:image:alt" content="A card of 5 options">',
		],
	])("finds a digit in %s", (_, head) => {
		expect(handTyped(pageWith("", head), rules)).not.toEqual([]);
	});

	test("lets through a number written as data or marked as given, a name of rules, and a commit and a date of the data", () => {
		const main = [
			'<p><data value="9">9</data> checks, <span data-given="">01</span></p>',
			"<p>XChaCha20-Poly1305</p>",
			"<code>0123456789ab</code>",
			'<time datetime="2026-10-05T00:00:00Z">5 October 2026</time>',
		].join("");
		expect(handTyped(pageWith(main), rules)).toEqual([]);
	});
});

describe("the page Research under the site's stylesheets", () => {
	// screen is a window as wide as a desktop's, under the stylesheets every
	// page of the site loads.
	const screen = new Window({ width: 1280, height: 800 });
	for (const sheet of [
		["internal", "widget", "tokens.css"],
		["web", "src", "site", "style.css"],
	]) {
		const style = screen.document.createElement("style");
		style.textContent = textOf(...sheet);
		screen.document.head.append(style);
	}

	afterAll(async () => {
		await screen.happyDOM.close();
	});

	// drawn is what selector finds on the page, drawn alone on the screen, with
	// the style it is drawn in there.
	const drawn = (selector: string) => {
		screen.document.body.innerHTML =
			page.querySelector(selector)?.outerHTML ?? "";
		const element = screen.document.body.firstElementChild;
		if (element === null) {
			throw new Error(`the page draws nothing ${selector} finds`);
		}
		return element;
	};
	const styleOf = (element: ReturnType<typeof drawn>) =>
		screen.getComputedStyle(element);

	// token is a colour of the site's palette, by its name.
	const token = (name: string) =>
		styleOf(screen.document.documentElement).getPropertyValue(name).trim();

	test.each([
		["reached", "--s-right", "--s-on-right"],
		["not_reached", "--s-trap", "--s-on-trap"],
		["on_the_edge", "--s-wip", "--s-on-wip"],
	])(
		"draws the mark %s in the colours of its kind, over the pill it is drawn as",
		(mark, background, text) => {
			const style = styleOf(drawn(`.s-research-legend .s-research-${mark}`));

			expect(style.backgroundColor).toBe(token(background));
			expect(style.color).toBe(token(text));
			expect(style.borderTopColor).toBe("transparent");
		},
	);

	test("starts the words of its first screen at the top of their row, with no card beside them", () => {
		const hero = drawn(".s-hero");
		const words = hero.querySelector(".s-hero-copy");

		expect(words === null ? "" : styleOf(words).paddingBlockStart).toBe("0");
	});
});
