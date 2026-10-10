import { Window } from "happy-dom";
import { afterAll, describe, expect, test } from "vitest";
import traps from "../../../content/catalogs/traps.json";
import file from "../../../site/data.json";
import type { TopicArt } from "./art/art";
import { Kind } from "./art/pictures";
import { readSiteData } from "./data";
import type { Frame } from "./frame";
import { type Page, sitePages } from "./pages";
import { renderSite } from "./render";
import { topicPage } from "./TopicPage";

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

// drawings are the drawings of the three topics' cards.
const drawings = {
	"logic.ordering": { card: { markup: "lined-up" } },
	"logic.knights_liars": { card: { markup: "knight-and-liar" } },
	"logic.sets": { card: { markup: "two-circles" } },
};

// dataWith is the site's data over a catalog of three topics of the service:
// knights and liars builds on ordering and opens sets; the pages of ordering
// and of knights and liars are published, and knights and liars has these
// examples, and these drawings. Its reference tasks name trusted_statement
// three times, and stopped_early and ignored_condition twice each, which the
// catalog's order puts in that order, and missed_case once.
const dataWith = (examples: unknown, drawn: unknown = drawings) =>
	readSiteData(
		{
			topics: [
				{
					id: "logic.ordering",
					slug: "ordering",
					grade_levels: ["1-2", "3-4"],
					builds_on: [],
					site_page: true,
				},
				{
					id: "logic.knights_liars",
					slug: "knights",
					grade_levels: ["3-4", "5-6"],
					builds_on: ["logic.ordering"],
					site_page: true,
				},
				{
					id: "logic.sets",
					slug: "sets",
					grade_levels: ["5-6"],
					builds_on: ["logic.knights_liars"],
					site_page: false,
				},
			],
			traps,
			tasks: [
				{
					id: "kl-1",
					topic: "logic.knights_liars",
					grade_level: "3-4",
					distractors: {
						B: { trap: "trusted_statement" },
						C: { trap: "stopped_early" },
						D: { trap: "ignored_condition" },
						E: { trap: "missed_case" },
					},
				},
				{
					id: "kl-2",
					topic: "logic.knights_liars",
					grade_level: "3-4",
					distractors: {
						B: { trap: "trusted_statement" },
						C: { trap: "trusted_statement" },
						D: { trap: "ignored_condition" },
						E: { trap: "stopped_early" },
					},
				},
			],
		},
		{
			groups: [
				{
					id: "logic",
					topics: ["logic.ordering", "logic.knights_liars", "logic.sets"],
				},
			],
			drawings: drawn,
			examples: { "logic.knights_liars": examples },
			progress: file.progress,
		},
	);

// examples are the examples of knights and liars.
const examples = [
	{ level: "3-4", solver: "first", answer: "A liar" },
	{ level: "5-6", solver: "second", answer: "2" },
];

const data = dataWith(examples);

// art are the drawings of the page of knights and liars: its first screen's
// alone, which draws no picture as a card draws one and reads none of the
// page's words.
const art: TopicArt = { hero: () => <span>A and B</span> };

// pictured are drawings of the same page whose first screen draws a picture
// of a clock as a card draws one, and which say so.
const pictured: TopicArt = {
	pictures: true,
	hero: ({ page }) => (
		<Kind picture={{ kind: "clock", time: "4:50" }} locale={page.locale} />
	),
};

// words are the page's words for that catalog, the same in each language: its
// first example with a note and its second with none.
const words = [
	"title: Knights",
	"description: About knights.",
	"hero:",
	"  lead: Who is who.",
	"  method: Suppose.",
	"  teaches: Reasoning.",
	"basis:",
	"  title: The rules",
	"  cards:",
	"    - mark: K",
	"      title: Knight",
	"      text: Tells the truth.",
	"idea:",
	"  title: Who says what",
	"  lead: Two sentences.",
	"  table:",
	"    head:",
	"      - Sentence",
	"      - Knight",
	"      - Liar",
	"    rows:",
	"      - - I am a knight",
	"        - Can say it",
	"        - Can say it",
	"  note:",
	"    title: Why",
	"    text: It tells nothing.",
	"method:",
	"  title: How to solve",
	"  lead: Suppose and see.",
	"  steps:",
	"    - Suppose",
	"    - Follow",
	"examples:",
	"  - title: First",
	"    question: Who is who?",
	"    steps:",
	"      - Suppose A is a knight.",
	"      - Then A is a liar.",
	"    answer: A is a liar.",
	"    note:",
	"      title: Why so",
	"      text: Because.",
	"  - title: Second",
	"    question: How many?",
	"    steps:",
	"      - Count them.",
	"    answer: Two.",
	"traps:",
	"  trusted_statement:",
	"    shows: Believed A.",
	"    say: Who said it?",
	"  stopped_early:",
	"    shows: Stopped at A.",
	"    say: And B?",
	"  ignored_condition:",
	"    shows: Skipped a sentence.",
	"    say: Read them all.",
	"home:",
	"  - title: Play",
	"    text: Be a liar.",
	"ask: Give me a task",
].join("\n");

// document is a document's text under its title.
const document = (title: string) =>
	`---\ntitle: ${title}\ndescription: D\n---\n\n# ${title}\n`;

// sourcesWith is the site's texts with these words for the page of knights
// and liars, in English, and in Russian unless russian is given.
const sourcesWith = (english: string, russian = english) =>
	new Map(
		Object.entries({ en: english, ru: russian }).map(([locale, knights]) => [
			locale,
			new Map([
				["index.md", document("Home")],
				["privacy.md", document("Privacy")],
				["terms.md", document("Terms")],
				["topics/ordering.yaml", "title: Ordering\ndescription: O\n"],
				["topics/knights.yaml", knights],
			]),
		]),
	);

// A frame with no menu, since the site drawn here has no page of the topics.
const frame: Frame = { menu: [], footer: ["privacy", "terms"] };

// knightsOf is the topic of knights and liars in the data.
function knightsOf(given: typeof data) {
	const knights = given.topics.all.find(({ slug }) => slug === "knights");
	if (knights === undefined) {
		throw new Error("the test's catalog has no knights and liars");
	}
	return knights;
}

// pagesOf are the site's pages for the data, knights and liars drawn with
// these drawings of its own, and ordering's page by a stand-in.
const pagesOf = (given: typeof data, drawn: TopicArt | undefined) =>
	new Map<string, Page>([
		...sitePages(given),
		["topics/ordering", { draw: () => <p>Ordering</p> }],
		["topics/knights", topicPage(knightsOf(given), drawn)],
	]);

// render is the site built from these words, data and drawings.
const render = (sources = sourcesWith(words), given = data, drawn = art) =>
	renderSite({
		base: "https://example.test",
		sources,
		frame,
		data: given,
		pages: pagesOf(given, drawn),
	});

const files = render();

// pageAt is the built page at path, read as a document, of the site built
// as given, or else of the one built here.
function pageAt(path: string, built = files) {
	return new browser.DOMParser().parseFromString(
		built.find((candidate) => candidate.path === path)?.data ?? "",
		"text/html",
	);
}

const knights = pageAt("en/topics/knights/index.html");

// all are the values of name, or the text, of every element selector finds.
const all = (selector: string, name?: string) =>
	[...knights.querySelectorAll(selector)].map((element) =>
		name === undefined
			? (element.textContent ?? "").trim()
			: (element.getAttribute(name) ?? ""),
	);

describe("a topic's page", () => {
	test("is drawn by the template at the topic's address, in every language", () => {
		expect(files.map(({ path }) => path)).toEqual(
			expect.arrayContaining([
				"en/topics/knights/index.html",
				"ru/topics/knights/index.html",
			]),
		);
		expect(
			pageAt("ru/topics/knights/index.html").querySelector("h1")?.textContent,
		).toBe("Рыцари и лжецы");
	});

	test("leads from the page of the topics through the topic's group", () => {
		expect(all(".s-path a", "href")).toEqual([
			"/en/topics/",
			"/en/topics/#logic",
		]);
		expect(all(".s-path a")).toEqual(["Topics", "Logic"]);
		expect(all('.s-path [aria-current="page"]')).toEqual(["Knights and liars"]);
	});

	test("names the topic as the card does, beside its group and the catalog's grades", () => {
		expect(all("h1")).toEqual(["Knights and liars"]);
		expect(all(".s-chip")).toEqual(["Logic", "Grades 3–6"]);
	});

	test("explains the topic's commonest traps, the commonest first, under the card's names", () => {
		expect(all(".s-trap h3")).toEqual([
			"Believed a statement without checking it",
			"Stopped halfway through the solution",
			"Missed one of the conditions",
		]);
		expect(all(".s-trap .s-tile-text")).toEqual([
			"Believed A.",
			"Who said it?",
			"Stopped at A.",
			"And B?",
			"Skipped a sentence.",
			"Read them all.",
		]);
	});

	test("labels each example with the grades of its level, which the data gives", () => {
		expect(all(".s-example-label")).toEqual([
			"Example 1 · Grades 3–4",
			"Example 2 · Grades 5–6",
		]);
	});

	test("numbers the steps of the move and of each example in the page's own digits", () => {
		expect(all(".s-moves .s-number")).toEqual(["1", "2"]);
		expect(
			[...knights.querySelectorAll(".s-worked-steps")].map((steps) =>
				[...steps.querySelectorAll(".s-worked-step-number")].map(
					(number) => number.textContent,
				),
			),
		).toEqual([["1", "2"], ["1"]]);
	});

	test("draws an example's note only where the example has one, and its answer under every example", () => {
		const worked = [...knights.querySelectorAll(".s-example")];

		expect(
			worked.map((example) => [
				example.querySelector(".s-note-title")?.textContent,
				example.querySelector(".s-answer-pill")?.textContent,
			]),
		).toEqual([
			["Why so", "A is a liar."],
			[undefined, "2Two."],
		]);
	});

	test("draws the key idea's table with a heading for every column and every row", () => {
		expect(all(".s-table thead th")).toEqual(["Sentence", "Knight", "Liar"]);
		expect(all('.s-table tbody th[scope="row"]')).toEqual(["I am a knight"]);
		expect(all(".s-table tbody td")).toEqual(["Can say it", "Can say it"]);
	});

	test("leads to the topics before and after it: to a page that is published, and says the others are coming", () => {
		const related = [...knights.querySelectorAll(".s-related")].map((list) => [
			list.parentElement?.querySelector("h2")?.textContent,
			[...list.querySelectorAll("li")].map((topic) => [
				topic.querySelector(".s-related-name")?.textContent,
				topic.querySelector("a")?.getAttribute("href") ??
					topic.querySelector(".s-soon")?.textContent,
			]),
		]);

		expect(related).toEqual([
			["Learn first", [["Ordering", "/en/topics/ordering/"]]],
			["Opens the way to", [["Overlapping groups", "Soon"]]],
		]);
	});

	test("carries the anchors the card links to, and every anchor its own list of parts leads to", () => {
		const ids = new Set(all("[id]", "id"));

		expect(ids).toContain("traps");
		expect(ids).toContain("home");
		expect(all(".s-contents a", "href")).toEqual([
			"#basics",
			"#idea",
			"#solving",
			"#traps",
			"#home",
		]);
		for (const anchor of all(".s-contents a", "href")) {
			expect(ids).toContain(anchor.slice(1));
		}
	});

	test("asks for a task in the page's words, leads to adding the app on the home page and back to every topic", () => {
		expect(all(".s-ask-phrase")).toEqual(["Give me a task"]);
		expect(all(".s-ask a", "href")).toEqual(["/#connect", "/en/topics/"]);
		expect(all(".s-ask .s-btn-filled")).toEqual(["Add to Claude"]);
	});

	test("draws the topic's own drawing on its first screen, which a screen reader passes over", () => {
		const panel = knights.querySelector(".s-subject-panel");

		expect(panel?.getAttribute("aria-hidden")).toBe("true");
		expect(all(".s-subject-panel .s-sketch-drawn")).toEqual(["A and B"]);
	});

	// The drawings of the page drawn here draw no picture as a card draws one,
	// and the clock's do.
	test("reads in full with no script, and loads the card's stylesheet only where its drawings draw a picture", () => {
		const clock = pageAt(
			"en/topics/knights/index.html",
			render(sourcesWith(words), data, pictured),
		);
		const sheets = (page: typeof knights) =>
			[...page.querySelectorAll('link[rel="stylesheet"]')].map((link) =>
				link.getAttribute("href"),
			);

		expect(knights.querySelector("script")).toBeNull();
		expect(sheets(knights)).not.toContain("/assets/card.css");
		expect(
			clock.querySelector(".s-subject-panel svg.mt-picture"),
		).not.toBeNull();
		expect(sheets(clock)).toContain("/assets/card.css");
	});
});

describe("a topic's page is refused when", () => {
	test.each([
		[
			"its words explain a trap that is not among the topic's commonest",
			words.replace("  ignored_condition:", "  missed_case:"),
			"en/topics/knights.yaml: the page reads traps.ignored_condition.shows, which the file does not have",
		],
		[
			"its words explain one trap more than the commonest",
			words.replace(
				"home:",
				"  missed_case:\n    shows: Missed.\n    say: Again.\nhome:",
			),
			"the page never shows traps.missed_case.shows, traps.missed_case.say",
		],
		[
			"a row of its table has another number of cells than its head",
			words.replace(
				"        - Can say it\n        - Can say it",
				"        - Can say it",
			),
			"idea.table.rows.1 has 2 cells, and the table's head 3",
		],
	])("%s", (_, english, want) => {
		expect(() => render(sourcesWith(english))).toThrow(want);
	});

	// Ordering's page is published and drawn by the template, and no reference
	// task of ordering names a trap: its page would have no trap to explain.
	test("no reference task of its topic names a trap, the commonest of which the page explains", () => {
		expect(() =>
			renderSite({
				base: "https://example.test",
				sources: sourcesWith(words),
				frame,
				data,
				pages: new Map([
					...sitePages(data),
					["topics/knights", topicPage(knightsOf(data), art)],
				]),
			}),
		).toThrow(
			"no reference task of logic.ordering names a trap, and its page explains the most frequent",
		);
	});

	test("its topic has no drawings of its own", () => {
		expect(() =>
			renderSite({
				base: "https://example.test",
				sources: sourcesWith(words),
				frame,
				data,
				pages: pagesOf(data, undefined),
			}),
		).toThrow("the page of logic.knights_liars has no drawings of its own");
	});

	test("its words work through another number of examples than the data gives", () => {
		const fewer = dataWith([{ level: "3-4", solver: "first", answer: "1" }]);

		expect(() => render(sourcesWith(words), fewer)).toThrow(
			"the words work through 2 examples, and site/data.json gives 1",
		);
	});

	test("its words in Russian lack a part the English have", () => {
		expect(() =>
			render(
				sourcesWith(
					words,
					words.replace(
						"    note:\n      title: Why so\n      text: Because.\n",
						"",
					),
				),
			),
		).toThrow("examples.1.note.title");
	});
});
