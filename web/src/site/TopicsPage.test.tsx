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

// drawings are the drawings of the three topics' cards: alpha's a knight and a
// liar, beta's a clock's face, and gamma's three people lined up by height.
const drawings = {
	"logic.alpha": { card: { markup: "knight-and-liar" } },
	"logic.beta": { card: { markup: "clock-face" } },
	"logic.gamma": { card: { markup: "lined-up" } },
};

// dataWith is the site's data over a catalog of three topics in three layers,
// with these drawings: beta builds on alpha, gamma on both, its link from
// alpha skipping a column. Only beta's page is published.
const dataWith = (drawn: unknown) =>
	readSiteData(
		{
			traps,
			tasks: [],
			topics: [
				{
					id: "logic.alpha",
					slug: "alpha",
					grade_levels: ["1-2", "3-4"],
					builds_on: [],
					site_page: false,
				},
				{
					id: "logic.beta",
					slug: "beta",
					grade_levels: ["3-4"],
					builds_on: ["logic.alpha"],
					site_page: true,
				},
				{
					id: "logic.gamma",
					slug: "gamma",
					grade_levels: ["5-6"],
					builds_on: ["logic.alpha", "logic.beta"],
					site_page: false,
				},
			],
		},
		{
			groups: [
				{ id: "logic", topics: ["logic.alpha", "logic.beta"] },
				{ id: "counting", topics: ["logic.gamma"] },
			],
			drawings: drawn,
			examples: {},
			progress: file.progress,
		},
	);

const data = dataWith(drawings);

// words are the page's words for that catalog, the same in each language.
const words = [
	"title: Topics",
	"description: Every topic.",
	"hero:",
	"  title: What tasks your child solves",
	"  lead: Every task belongs to one topic.",
	"  child: Comet",
	"  caption: The progress, as an example.",
	"map:",
	"  title: How the topics link",
	"  lead: Some topics stand on others.",
	"  layers:",
	...[1, 2, 3].flatMap((step) => [
		`    - step: Step ${step}`,
		`      name: Layer ${step}`,
		`      lead: The layer ${step}.`,
	]),
	"  hint: Point at a topic.",
	'  say: "**{topic}** · first: {before} · opens: {after}"',
	"  foundation: nothing",
	"  summit: nothing",
	"  legend:",
	"    first: First",
	"    topic: Topic",
	"    then: Then",
	"groups:",
	"  label: Groups of topics",
	"  logic: Reasoning.",
	"  counting: Counting.",
	"topics:",
	"  alpha:",
	"    phrase: The topic alpha.",
	"  beta:",
	"    phrase: The topic beta.",
	"  gamma:",
	"    phrase: The topic gamma.",
].join("\n");

// document is a document's text under its title.
const document = (title: string) =>
	`---\ntitle: ${title}\ndescription: D\n---\n\n# ${title}\n`;

const sources = new Map(
	["en", "ru"].map((locale) => [
		locale,
		new Map([
			["index.md", document("Home")],
			["privacy.md", document("Privacy")],
			["terms.md", document("Terms")],
			["topics.yaml", words],
			["topics/beta.yaml", "title: Beta\ndescription: B\n"],
		]),
	]),
);

const pages = new Map<string, Page>([
	["topics", sitePages(data).get("topics") ?? { draw: () => <p /> }],
	["topics/beta", { draw: () => <p>Beta</p> }],
]);

// A menu of the one page this site has.
const frame: Frame = {
	menu: [{ page: "topics", label: "nav.topics" }],
	footer: ["privacy", "terms"],
};

const files = renderSite({
	base: "https://example.test",
	sources,
	pages,
	frame,
	data,
});

// pageAt is the built page at path, read as a document.
function pageAt(path: string) {
	const file = files.find((candidate) => candidate.path === path);
	if (file === undefined) {
		throw new Error(`the site has no ${path}`);
	}
	return new browser.DOMParser().parseFromString(file.data, "text/html");
}

// topics is the page of the topics in English, read as a document.
const topics = pageAt("en/topics/index.html");

// all are the values of name on every element selector finds.
const all = (selector: string, name: string) =>
	[...topics.querySelectorAll(selector)].map(
		(element) => element.getAttribute(name) ?? "",
	);

describe("the page of the topics", () => {
	test("shows exactly the catalog's topics, as cards in their groups", () => {
		expect(all(".s-group", "id")).toEqual(["logic", "counting"]);
		expect(all(".s-topic", "id")).toEqual(["alpha", "beta", "gamma"]);
		expect(all(".s-pills a", "href")).toEqual(["#logic", "#counting"]);
	});

	test("shows exactly the catalog's topics on the map, each leading to its card", () => {
		expect(all(".s-map-node", "id")).toEqual([
			"map-alpha",
			"map-beta",
			"map-gamma",
		]);
		expect(all(".s-map-node", "href")).toEqual(["#alpha", "#beta", "#gamma"]);
	});

	test("draws a line on the map for every link of the catalog", () => {
		expect(all(".s-map-line", "id")).toEqual([
			"line-alpha-beta",
			"line-alpha-gamma",
			"line-beta-gamma",
		]);
	});

	test("links a topic to its page when, and only when, the page is published", () => {
		const linked = [...topics.querySelectorAll(".s-topic")].map((card) =>
			card.querySelector(".s-topic-foot a")?.getAttribute("href"),
		);

		expect(linked).toEqual([undefined, "/en/topics/beta/", undefined]);
		expect(
			[...topics.querySelectorAll(".s-soon")].map((soon) => soon.textContent),
		).toEqual(["Soon", "Soon"]);
	});

	test("says in words what each topic builds on and opens, and its grades", () => {
		const said = [...topics.querySelectorAll(".s-topic")].map((card) =>
			[
				...card.querySelectorAll(
					".s-topic-links, .s-topic-foot span:first-child",
				),
			].map((line) => line.textContent?.split(":")[0]),
		);

		expect(said).toEqual([
			["Opens", "Grades 1–4"],
			["Builds on", "Opens", "Grades 3–4"],
			["Builds on", "Grades 5–6"],
		]);
	});

	test("draws each topic's drawing of the site's own, which a screen reader passes over", () => {
		const drawn = [...topics.querySelectorAll(".s-topic")].map(
			(card) => card.firstElementChild,
		);

		expect(drawn.map((box) => box?.getAttribute("aria-hidden"))).toEqual([
			"true",
			"true",
			"true",
		]);
		expect(
			[...(drawn[0]?.querySelectorAll(".s-thumb-islander") ?? [])].map(
				(islander) => islander.getAttribute("data-role"),
			),
		).toEqual(["knight", "liar"]);
		expect(drawn[1]?.querySelector(".s-thumb")).not.toBeNull();
		expect(drawn[2]?.querySelectorAll(".s-thumb-person")).toHaveLength(3);
	});

	test("is refused when the data gives a topic's card no drawing", () => {
		const { "logic.gamma": _, ...fewer } = drawings;

		expect(() =>
			renderSite({
				base: "https://example.test",
				sources,
				pages,
				frame,
				data: dataWith(fewer),
			}),
		).toThrow("site/data.json gives the card of logic.gamma no drawing");
	});

	test("shows the widget's own progress card, inert, and loads its stylesheet", () => {
		expect(topics.querySelector(".s-card[inert] .mt-widget")).not.toBeNull();
		expect(all('link[rel="stylesheet"]', "href")).toContain("/assets/card.css");
	});

	test("carries in its head the rules that light the map, from the catalog's links", () => {
		const rules = topics.querySelector("head style")?.textContent ?? "";

		expect(rules).toContain(
			".s-map:has(#map-gamma:is(:hover,:focus-visible)) #map-alpha",
		);
		expect(rules).toContain(
			".s-map:has(#map-alpha:is(:hover,:focus-visible)) #line-beta-gamma",
		);
	});

	test("reads in full with no script", () => {
		expect(topics.querySelector("script")).toBeNull();
	});

	test("is the menu's entry, marked as the page being read", () => {
		expect(all(".s-navlinks a", "href")).toEqual(["/en/topics/"]);
		expect(all(".s-navlinks a", "aria-current")).toEqual(["page"]);
	});

	test("stays the menu's marked entry on a topic's page, as the page that one lies under", () => {
		const beta = pageAt("en/topics/beta/index.html");
		const entry = beta.querySelector(".s-navlinks a");

		expect(entry?.getAttribute("aria-current")).toBe("true");
	});
});
