import { Window } from "happy-dom";
import { afterAll, describe, expect, test } from "vitest";
import topics from "../../../content/catalogs/topics.json";
import traps from "../../../content/catalogs/traps.json";
import file from "../../../site/data.json";
import { contactAddress, issuesURL, sourceURL } from "./brand";
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

// data is the site's data over the service's catalog, with no page of a topic
// published: the page about who makes MathTrail draws none of it.
const data = readSiteData(
	{
		topics: topics.map((topic) => ({ ...topic, site_page: false })),
		traps,
		tasks: [],
	},
	{ groups: file.groups, examples: {}, progress: file.progress },
);

// words are the page's words, the same in each language: two of the family,
// and a sentence that names the address to write to.
const words = [
	"title: About",
	"description: Who makes it.",
	"hero:",
	"  title: Made by one family",
	"  lead: There are two of us.",
	"team:",
	"  title: Who we are",
	"  members:",
	"    - role: Builder",
	"      name: Dad",
	"      text: Writes the code.",
	"    - role: Boss",
	"      name: Mum",
	"      text: Runs everything.",
	"contact:",
	"  title: Found a mistake?",
	"  lead: Write to {address}, or {issue}.",
	"  issue: open an issue",
	"  write: Write to us",
	"  github: GitHub",
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
				["about.yaml", said],
			]),
		]),
	);

// aboutPage is the site's own page about who makes MathTrail.
const aboutPage = sitePages(data).get("about");
if (aboutPage === undefined) {
	throw new Error("the site's pages have no page about who makes MathTrail");
}

const pages = new Map<string, Page>([["about", aboutPage]]);

// A frame shaped like the site's own, over the one page this site has: the
// menu leads to it, and the footer names it before the documents.
const frame: Frame = {
	menu: [{ page: "about", label: "nav.about" }],
	footer: ["about", "privacy", "terms"],
};

const render = (sources = sourcesWith()) =>
	renderSite({ base: "https://example.test", sources, pages, frame, data });

const files = render();

// about is the page in English, read as a document.
const about = new browser.DOMParser().parseFromString(
	files.find(({ path }) => path === "en/about/index.html")?.data ?? "",
	"text/html",
);

// all are the values of name on every element selector finds.
const all = (selector: string, name: string) =>
	[...about.querySelectorAll(selector)].map(
		(element) => element.getAttribute(name) ?? "",
	);

// texts are the words of every element selector finds.
const texts = (selector: string) =>
	[...about.querySelectorAll(selector)].map((element) => element.textContent);

describe("the page about who makes MathTrail", () => {
	test("reads in full with no script, and draws no card and no picture", () => {
		expect(about.querySelector("script")).toBeNull();
		expect(all('link[rel="stylesheet"]', "href")).not.toContain(
			"/assets/card.css",
		);
		expect(about.querySelector("main img")).toBeNull();
	});

	test("shows a card for each of the family, in the words' order: the role, who it is, what they do", () => {
		expect(
			[...about.querySelectorAll(".s-member")].map((card) => [
				card.querySelector(".s-member-role")?.textContent,
				card.querySelector(".s-member-name")?.textContent,
				card.querySelector(".s-tile-text")?.textContent,
			]),
		).toEqual([
			["Builder", "Dad", "Writes the code."],
			["Boss", "Mum", "Runs everything."],
		]);
	});

	test("names each of the family in a heading under one a screen reader alone hears", () => {
		expect(texts(".s-team h2.s-hidden")).toEqual(["Who we are"]);
		expect(all(".s-team", "aria-labelledby")).toEqual(["team"]);
		expect(texts(".s-member h3")).toEqual(["Dad", "Mum"]);
		expect(
			[...about.querySelectorAll(".s-member")].map((card) =>
				[...card.children].map((part) => part.getAttribute("class")),
			)[0],
		).toEqual([
			"s-chip s-chip-group s-member-role",
			"s-member-name",
			"s-tile-text",
		]);
	});

	test("writes to the footer's address, written out and on a button", () => {
		const footer = all(".s-footlinks a", "href");

		expect(all("main a[href^='mailto:']", "href")).toEqual([
			`mailto:${contactAddress}`,
			`mailto:${contactAddress}`,
		]);
		expect(footer).toContain(`mailto:${contactAddress}`);
		expect(texts(".s-contact-copy a")[0]).toBe(contactAddress);
	});

	test("leads to the issues of the code, and to the code the footer leads to", () => {
		expect(all(".s-contact-copy a", "href")).toEqual([
			`mailto:${contactAddress}`,
			issuesURL,
		]);
		expect(texts(".s-contact-copy a")[1]).toBe("open an issue");
		expect(all(".s-contact-actions a", "href")).toEqual([
			`mailto:${contactAddress}`,
			sourceURL,
		]);
		expect(all(".s-footlinks a", "href")).toContain(sourceURL);
	});

	test("is the menu's entry, marked as the page being read", () => {
		expect(all(".s-navlinks a", "href")).toEqual(["/en/about/"]);
		expect(all(".s-navlinks a", "aria-current")).toEqual(["page"]);
	});

	test("stands in the footer under its own title, before the documents", () => {
		expect(all(".s-footlinks a", "href").slice(0, 3)).toEqual([
			"/en/about/",
			"/en/privacy/",
			"/en/terms/",
		]);
		expect(texts(".s-footlinks a")[0]).toBe("About");
	});

	test("is refused when its Russian words lack one the English have", () => {
		const lacking = words.replace("      name: Mum\n", "");

		expect(lacking).not.toBe(words);
		expect(() => render(sourcesWith(lacking))).toThrow("team.members.2.name");
	});
});
