import { Window } from "happy-dom";
import { afterAll, describe, expect, test } from "vitest";
import traps from "../../../content/catalogs/traps.json";
import file from "../../../site/data.json";
import { coachPrototypePath } from "./brand";
import { prototypeSection } from "./CoachPage";
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

// data is a catalog of two topics with no page of their own, and three
// techniques in two groups, which the page of the coach counts.
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
				site_page: false,
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

// words are the page's words, the same in each language: two ideas, and two
// things the prototype only pretends to do.
const words = [
	"title: Coach",
	"description: A coach in the making.",
	"hero:",
	"  badge: In development",
	"  title: Where the child thinks",
	"  lead: We are making a coach.",
	"  try: Try the prototype",
	"ideas:",
	"  title: What it is about",
	"  items:",
	"    - title: Four steps",
	"      text: Understand, plan, solve, check.",
	"    - title: A canvas",
	"      text: The drawing is part of the solution.",
	"prototype:",
	"  frame: An interactive prototype",
	"  caption: Try drawing a solution",
	"  open: Open full screen",
	"notes:",
	"  title: This is a prototype",
	"  text: Everything can be clicked.",
	"  imitated:",
	"    title: Only pretended",
	"    items:",
	"      - the voice",
	"      - problems from a photo",
].join("\n");

// document is a document's text under its title.
const document = (title: string) =>
	`---\ntitle: ${title}\ndescription: D\n---\n\n# ${title}\n`;

// sourcesWith are the site's texts, with the page's words in English and in
// Russian, which are these unless others are given.
const sourcesWith = (russian = words) =>
	new Map(
		Object.entries({ en: words, ru: russian }).map(([locale, said]) => [
			locale,
			new Map([
				["index.md", document("Home")],
				["privacy.md", document("Privacy")],
				["terms.md", document("Terms")],
				["coach.yaml", said],
			]),
		]),
	);

// coachPage is the site's own page of the coach.
const coachPage = sitePages(data).get("coach");
if (coachPage === undefined) {
	throw new Error("the site's pages have no page of the coach");
}

const pages = new Map<string, Page>([["coach", coachPage]]);

// A menu of the one page this site has.
const frame: Frame = {
	menu: [{ page: "coach", label: "nav.coach" }],
	footer: ["privacy", "terms"],
};

const render = (sources = sourcesWith()) =>
	renderSite({ base: "https://example.test", sources, pages, frame, data });

const files = render();

// read is the page of the coach in locale, read as a document.
const read = (locale: string) =>
	new browser.DOMParser().parseFromString(
		files.find(({ path }) => path === `${locale}/coach/index.html`)?.data ?? "",
		"text/html",
	);

const coach = read("en");

// all are the values of name on every element selector finds.
const all = (selector: string, name: string) =>
	[...coach.querySelectorAll(selector)].map(
		(element) => element.getAttribute(name) ?? "",
	);

// texts are the words of every element selector finds.
const texts = (selector: string) =>
	[...coach.querySelectorAll(selector)].map((element) => element.textContent);

describe("the page of the coach", () => {
	test("reads in full with no script, and draws no card", () => {
		expect(coach.querySelector("script")).toBeNull();
		expect(all('link[rel="stylesheet"]', "href")).not.toContain(
			"/assets/card.css",
		);
	});

	test("says the coach is in the making, over its heading and what it is", () => {
		expect(texts(".s-coach-hero .s-badge")).toEqual(["In development"]);
		expect(texts(".s-coach-hero h1")).toEqual(["Where the child thinks"]);
		expect(texts(".s-coach-hero .s-lead")).toEqual(["We are making a coach."]);
	});

	test("leads to its prototype further down and to the techniques, as many as the site's data has", () => {
		expect(all(".s-hero-actions a", "href")).toEqual([
			`#${prototypeSection}`,
			"/en/techniques/",
		]);
		expect(texts(".s-hero-actions a")).toEqual([
			"Try the prototype",
			"3 techniques",
		]);
		expect(coach.getElementById(prototypeSection)?.tagName).toBe("SECTION");
		expect(
			[...read("ru").querySelectorAll(".s-hero-actions a")].map(
				(link) => link.textContent,
			)[1],
		).toBe("3 приёма");
	});

	test("shows each idea as a card under a heading a screen reader alone hears", () => {
		expect(texts("h2#ideas.s-hidden")).toEqual(["What it is about"]);
		expect(
			[...coach.querySelectorAll(".s-coach-ideas .s-tile")].map((card) => [
				card.querySelector("h3")?.textContent,
				card.querySelector(".s-tile-text")?.textContent,
			]),
		).toEqual([
			["Four steps", "Understand, plan, solve, check."],
			["A canvas", "The drawing is part of the solution."],
		]);
	});

	test("frames the prototype once a reader comes near it, under a name a screen reader says, and opens it alone in a tab of its own", () => {
		expect(all(`#${prototypeSection} iframe`, "src")).toEqual([
			coachPrototypePath,
		]);
		expect(all(`#${prototypeSection} iframe`, "loading")).toEqual(["lazy"]);
		expect(all(`#${prototypeSection} iframe`, "title")).toEqual([
			"An interactive prototype",
		]);
		expect(texts(".s-prototype-caption span")).toEqual([
			"Try drawing a solution",
		]);
		const open = coach.querySelector(".s-prototype-caption a");
		expect(open?.getAttribute("href")).toBe(coachPrototypePath);
		expect(open?.getAttribute("target")).toBe("_blank");
		expect(open?.getAttribute("rel")).toBe("noopener");
		expect(open?.textContent).toBe("Open full screen");
	});

	test("frames the same prototype in every language", () => {
		expect(
			[...read("ru").querySelectorAll("iframe")].map((frame) =>
				frame.getAttribute("src"),
			),
		).toEqual([coachPrototypePath]);
	});

	test("says what the prototype is, and what it only pretends to do yet", () => {
		expect(texts(".s-coach-notes h2")).toEqual(["This is a prototype"]);
		expect(texts(".s-coach-notes .s-tile-text")).toEqual([
			"Everything can be clicked.",
		]);
		expect(texts(".s-coach-notes h3")).toEqual(["Only pretended"]);
		expect(texts(".s-coach-imitated .s-chip")).toEqual([
			"the voice",
			"problems from a photo",
		]);
	});

	test("is the menu's entry, marked as the page being read", () => {
		expect(all(".s-navlinks a", "href")).toEqual(["/en/coach/"]);
		expect(all(".s-navlinks a", "aria-current")).toEqual(["page"]);
		expect(texts(".s-navlinks a")).toEqual(["Coach"]);
	});

	test("is refused when its Russian words lack one the English have", () => {
		const lacking = words.replace("  open: Open full screen\n", "");

		expect(lacking).not.toBe(words);
		expect(() => render(sourcesWith(lacking))).toThrow("prototype.open");
	});
});
