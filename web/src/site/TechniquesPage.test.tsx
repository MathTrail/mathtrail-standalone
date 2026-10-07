import { Window } from "happy-dom";
import { afterAll, describe, expect, test } from "vitest";
import traps from "../../../content/catalogs/traps.json";
import file from "../../../site/data.json";
import { readSiteData } from "./data";
import type { Frame } from "./frame";
import { type Page, sitePages } from "./pages";
import { renderSite } from "./render";

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

// A catalog of two topics, knights and liars with its page published and the
// games without, and three techniques in two groups: one leading to the
// knights, one to the games, and one to no topic.
const data = readSiteData(
	{
		traps,
		tasks: [],
		topics: [
			{
				id: "logic.knights_liars",
				slug: "knights-and-liars",
				grade_levels: ["3-4", "5-6"],
				builds_on: [],
				site_page: true,
			},
			{
				id: "games.strategy",
				slug: "winning-strategy",
				grade_levels: ["5-6"],
				builds_on: [],
				site_page: false,
			},
		],
	},
	{
		groups: [
			{ id: "logic", topics: ["logic.knights_liars"] },
			{ id: "games", topics: ["games.strategy"] },
		],
		examples: {},
		progress: file.progress,
		techniques: {
			groups: [
				{
					id: "see",
					techniques: [
						{
							id: "draw",
							topics: ["logic.knights_liars"],
							example: { level: "3-4", solver: "one", answer: "1" },
						},
					],
				},
				{
					id: "prove",
					techniques: [
						{
							id: "mirror",
							topics: ["games.strategy"],
							example: { level: "5-6", solver: "two", answer: "2" },
						},
						{
							id: "all-the-same",
							topics: [],
							example: { level: "3-4", solver: "three", answer: "3" },
						},
					],
				},
			],
			cues: [["draw"], ["mirror", "all-the-same"]],
		},
	},
);

// technique is the words of a technique: draw alone has a second name, and
// each has the labels of its drawing.
const technique = (id: string, drawing: readonly string[], aka = false) => [
	`  ${id}:`,
	`    name: The ${id}`,
	...(aka ? [`    aka: Also ${id}`] : []),
	`    what: What ${id} is.`,
	`    when: when ${id} helps.`,
	`    task: The task of ${id}?`,
	"    steps:",
	`      - First step of ${id}.`,
	`      - Second step of ${id}.`,
	`    answer: The answer of ${id}.`,
	"    drawing:",
	...drawing.map((label) => `      ${label}: ${label} of ${id}`),
	`    caption: The drawing of ${id}.`,
];

// words are the page's words for those techniques, with as many rows of the
// hint as rows given, the same in each language.
const words = (rows = 2) =>
	[
		"title: Techniques",
		"description: Every technique.",
		"hero:",
		"  title: How problems are solved",
		"  lead: Here are {count}, each worked through.",
		"  advice: Try the problem first.",
		"start:",
		"  title: Where to begin",
		"  lead: Four steps.",
		"  steps:",
		"    - title: Understand",
		"      text: Say it in your own words.",
		"groups:",
		"  see:",
		"    title: See it",
		"    lead: Picture the problem.",
		"  prove:",
		"    title: Prove it",
		"    lead: Explain why.",
		"techniques:",
		...technique(
			"draw",
			["title", "sister", "brother", "without", "count"],
			true,
		),
		...technique("mirror", ["title", "first", "second", "moves"]),
		...technique("all-the-same", [
			"guess",
			"guess-legs",
			"short",
			"real-legs",
			"rabbits",
			"hens",
		]),
		"labels:",
		'  when: "When it helps:"',
		"  task: Problem · {grades}",
		"  show: Show the solution",
		'  topics: "In the topics:"',
		"cues:",
		"  title: Which to try",
		"  lead: A hint.",
		"  if: If the problem has…",
		"  try: …try",
		"  rows:",
		...Array.from(
			{ length: rows },
			(_, row) => `    - Words of row ${row + 1}`,
		),
		"end:",
		"  title: Try them",
		"  lead: In every topic.",
	].join("\n");

// document is a document's text under its title.
const document = (title: string) =>
	`---\ntitle: ${title}\ndescription: D\n---\n\n# ${title}\n`;

// sourcesWith are the site's texts with the page's words given.
const sourcesWith = (techniques: string) =>
	new Map(
		["en", "ru"].map((locale) => [
			locale,
			new Map([
				["index.md", document("Home")],
				["privacy.md", document("Privacy")],
				["terms.md", document("Terms")],
				["techniques.yaml", techniques],
				["topics/knights-and-liars.yaml", "title: Knights\ndescription: K\n"],
			]),
		]),
	);

const pages = new Map<string, Page>([
	["techniques", sitePages(data).get("techniques") ?? { draw: () => <p /> }],
	["topics/knights-and-liars", { draw: () => <p>Knights</p> }],
]);

// A menu of the one page this site has.
const frame: Frame = {
	menu: [{ page: "techniques", label: "nav.techniques" }],
	footer: ["privacy", "terms"],
};

const site = {
	base: "https://example.test",
	sources: sourcesWith(words()),
	pages,
	frame,
	data,
};

const files = renderSite(site);

// techniques is the page of the techniques in English, read as a document.
const techniques = new browser.DOMParser().parseFromString(
	files.find(({ path }) => path === "en/techniques/index.html")?.data ?? "",
	"text/html",
);

// all are the values of name on every element selector finds.
const all = (selector: string, name: string) =>
	[...techniques.querySelectorAll(selector)].map(
		(element) => element.getAttribute(name) ?? "",
	);

// texts are the texts of every element selector finds.
const texts = (selector: string) =>
	[...techniques.querySelectorAll(selector)].map(
		(element) => element.textContent?.trim() ?? "",
	);

// card is the card of the technique called id.
const card = (id: string) => techniques.getElementById(id);

describe("the page of the techniques", () => {
	test("lists every technique by its group on its first screen, numbered in the order of the page", () => {
		expect(texts(".s-overview-title")).toEqual(["See it", "Prove it"]);
		expect(all(".s-overview a", "href")).toEqual([
			"#draw",
			"#mirror",
			"#all-the-same",
		]);
		expect(texts(".s-overview .s-technique-link-number")).toEqual([
			"01",
			"02",
			"03",
		]);
	});

	test("says how many techniques there are in the dictionary's words", () => {
		expect(texts(".s-hero .s-lead")).toEqual([
			"Here are 3 techniques, each worked through.",
		]);
	});

	test("holds each technique in its group's part of the page", () => {
		const parts = [...techniques.querySelectorAll("section")].filter(
			(part) => part.querySelector(".s-technique") !== null,
		);

		expect(
			parts.map((part) =>
				[...part.querySelectorAll(".s-technique")].map((each) => each.id),
			),
		).toEqual([["draw"], ["mirror", "all-the-same"]]);
	});

	test("sets each problem at its example's grades", () => {
		expect(texts(".s-technique-task .s-example-label")).toEqual([
			"Problem · Grades 3–4",
			"Problem · Grades 5–6",
			"Problem · Grades 3–4",
		]);
	});

	test("folds the solution — its steps, its drawing and its answer — and leaves the problem out of the fold", () => {
		const fold = card("draw")?.querySelector("details.s-solution");

		expect(fold?.hasAttribute("open")).toBe(false);
		expect(fold?.querySelector("summary")?.textContent).toBe(
			"Show the solution",
		);
		expect(fold?.querySelectorAll(".s-example-steps li")).toHaveLength(2);
		expect(fold?.querySelector(".s-example-answer")?.textContent).toContain(
			"The answer of draw.",
		);
		expect(fold?.querySelector(".s-art .s-ages")).not.toBeNull();
		expect(
			card("draw")?.querySelector(".s-technique-question")?.closest("details"),
		).toBeNull();
	});

	test("hides a drawing from a screen reader, keeping its caption", () => {
		const figure = card("mirror")?.querySelector(".s-figure");

		expect(figure?.querySelector(".s-art")?.getAttribute("aria-hidden")).toBe(
			"true",
		);
		expect(figure?.querySelector("figcaption")?.textContent).toBe(
			"The drawing of mirror.",
		);
	});

	test("gives a technique its second name only where its words have one", () => {
		expect(texts(".s-technique-aka")).toEqual(["Also draw"]);
	});

	test("links a technique's topic to its page when, and only when, the page is published", () => {
		expect(
			[...(card("draw")?.querySelectorAll(".s-technique-topics a") ?? [])].map(
				(link) => link.getAttribute("href"),
			),
		).toEqual(["/en/topics/knights-and-liars/"]);
		expect(card("mirror")?.querySelector(".s-technique-topics a")).toBeNull();
		expect(
			card("mirror")?.querySelector(".s-technique-topics .s-soon")?.textContent,
		).toBe("Soon");
		expect(
			card("all-the-same")?.querySelector(".s-technique-topics"),
		).toBeNull();
	});

	test("says to a screen reader, as to the eye, that the techniques beside a problem's words are what to try", () => {
		const head = techniques.querySelector(".s-cues-head");

		expect(head?.getAttribute("aria-hidden")).toBeNull();
		expect(head?.textContent).toBe("If the problem has……try");
	});

	test("leads every row of the hint to techniques of the page", () => {
		expect(
			[...techniques.querySelectorAll(".s-cue dd")].map((row) =>
				[...row.querySelectorAll("a")].map((link) => link.getAttribute("href")),
			),
		).toEqual([["#draw"], ["#mirror", "#all-the-same"]]);
	});

	test("names every part of itself once, and every link inside it leads to one", () => {
		const ids = all("[id]", "id");
		const inside = all('a[href^="#"]', "href").map((href) => href.slice(1));

		expect(new Set(ids).size).toBe(ids.length);
		expect(inside.filter((id) => !ids.includes(id))).toEqual([]);
	});

	test("closes with the way to add the app on the home page, and to every topic", () => {
		expect(all(".s-ask .s-btn", "href")).toEqual(["/#connect", "/en/topics/"]);
	});

	test("is refused when the words give the hint another number of rows than the data", () => {
		expect(() =>
			renderSite({ ...site, sources: sourcesWith(words(1)) }),
		).toThrow(
			"the hint has 1 rows in the words of the page, and 2 in site/data.json",
		);
	});

	test("is refused when the data gives it no techniques", () => {
		const bare = { ...data, techniques: undefined };

		expect(() =>
			renderSite({
				...site,
				pages: new Map([
					...pages,
					[
						"techniques",
						sitePages(bare).get("techniques") ?? { draw: () => <p /> },
					],
				]),
				data: bare,
			}),
		).toThrow("site/data.json gives the page of the techniques none");
	});
});
