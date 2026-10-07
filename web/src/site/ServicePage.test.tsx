import { readFileSync } from "node:fs";
import { dirname, join } from "node:path";
import { fileURLToPath } from "node:url";
import { Window } from "happy-dom";
import { afterAll, describe, expect, test } from "vitest";
import { readSiteData, siteData } from "./data";
import type { Frame } from "./frame";
import { type Page, sitePages } from "./pages";
import { renderSite } from "./render";
import {
	coreParts,
	faceId,
	gridColumn,
	linkKey,
	neighbours,
	parts,
	pickId,
	quickPicks,
	scenarioId,
	scenarios,
	serviceRules,
	stepsId,
	tools,
	wholeMap,
} from "./service";

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

// wordsOf are the page's own words in locale, as the site keeps them.
const wordsOf = (locale: string) =>
	readFileSync(
		join(
			dirname(fileURLToPath(import.meta.url)),
			"..",
			"..",
			"..",
			"site",
			"content",
			locale,
			"service.yaml",
		),
		"utf8",
	);
const english = wordsOf("en");
const russian = wordsOf("ru");

// data is the site's data over a catalog of one topic with no page of its
// own: the page of the service draws none of it.
const data = readSiteData(
	{
		topics: [
			{
				id: "logic.sample",
				slug: "sample",
				grade_levels: ["3-4"],
				builds_on: [],
				site_page: false,
			},
		],
		traps: [],
		tasks: [],
	},
	{
		groups: [{ id: "logic", topics: ["logic.sample"] }],
		examples: {},
		progress: siteData().progress,
	},
);

// document is a document's text under its title.
const document = (title: string) =>
	`---\ntitle: ${title}\ndescription: D\n---\n\n# ${title}\n`;

// sourcesWith are the site's texts, the page's words in English and in
// Russian, which are the site's own unless others are given.
const sourcesWith = (inEnglish = english, inRussian = russian) =>
	new Map(
		Object.entries({ en: inEnglish, ru: inRussian }).map(([locale, words]) => [
			locale,
			new Map([
				["index.md", document("Home")],
				["privacy.md", document("Privacy")],
				["terms.md", document("Terms")],
				["service.yaml", words],
			]),
		]),
	);

// servicePage is the site's own page of the service.
const servicePage = sitePages(data).get("service");
if (servicePage === undefined) {
	throw new Error("the site's pages have no page of the service");
}

const pages = new Map<string, Page>([["service", servicePage]]);

// A menu of the one page this site has, in the group of the technical pages.
const frame: Frame = {
	menu: [
		{
			label: "nav.technical",
			items: [{ page: "service", label: "nav.service" }],
		},
	],
	footer: ["privacy", "terms"],
};

const render = (sources = sourcesWith()) =>
	renderSite({ base: "https://example.test", sources, pages, frame, data });

const files = render();

// read is the page of the service in locale, read as a document.
const read = (locale: string) =>
	new browser.DOMParser().parseFromString(
		files.find(({ path }) => path === `${locale}/service/index.html`)?.data ??
			"",
		"text/html",
	);

const service = read("en");
const russianService = read("ru");

// all are the values of name on every element selector finds.
const all = (selector: string, name: string, doc = service) =>
	[...doc.querySelectorAll(selector)].map(
		(element) => element.getAttribute(name) ?? "",
	);

// styleOf is what the style the page writes on the element selector finds
// first sets property to, as the page writes it.
const styleOf = (selector: string, property: string, doc = service) =>
	declared(doc.querySelector(selector)?.getAttribute("style"), property);

// stylesOf is what the style of every element selector finds sets property
// to.
const stylesOf = (selector: string, property: string, doc = service) =>
	[...doc.querySelectorAll(selector)].map((element) =>
		declared(element.getAttribute("style"), property),
	);

// declared is what a style attribute sets property to, or nothing.
function declared(style: string | null | undefined, property: string): string {
	for (const declaration of (style ?? "").split(";")) {
		const colon = declaration.indexOf(":");
		if (declaration.slice(0, colon).trim() === property) {
			return declaration.slice(colon + 1).trim();
		}
	}
	return "";
}

// texts are the words of every element selector finds.
const texts = (selector: string, doc = service) =>
	[...doc.querySelectorAll(selector)].map(
		(element) => element.textContent ?? "",
	);

describe("the page of the service", () => {
	test("reads in full with no script, and draws no card", () => {
		expect(service.querySelector("script")).toBeNull();
		expect(all('link[rel="stylesheet"]', "href")).not.toContain(
			"/assets/card.css",
		);
	});

	test("carries in its head the rules written from the map and the scenarios", () => {
		expect(texts("head style")).toEqual([serviceRules()]);
	});

	test("says how the service works on its first screen, beside the service in three tiers", () => {
		expect(texts(".s-service-hero h1")).toEqual(["How the service works"]);
		expect(texts(".s-service-chips .s-chip")).toHaveLength(5);
		expect(all(".s-service-tiers", "aria-label")).toEqual([
			"MathTrail in three tiers",
		]);
		expect(service.querySelectorAll(".s-service-tier")).toHaveLength(3);
		expect(texts(".s-service-via")).toEqual(["MCP", "Drive API v3"]);
		expect(texts(".s-service-file")).toEqual([
			"MathTrail/mathtrail-profile.json",
		]);
	});

	test("names the service's core by the names the map gives its parts", () => {
		const names = coreParts.map(
			(part) =>
				service.querySelector(`#${pickId(part)}`)?.getAttribute("aria-label") ??
				"",
		);

		expect(texts(".s-service-core li")).toEqual(names);
		expect(names).not.toContain("");
	});

	test("numbers its parts in the page's order, each under its heading", () => {
		expect(texts(".s-service-number")).toEqual(["01", "02", "03", "04", "05"]);
		expect(
			[...service.querySelectorAll(".s-service-section")].map(
				(section) => section.id,
			),
		).toEqual(["map", "flow", "state", "profile", "seal"]);
		expect(all(".s-service-section", "aria-labelledby")).toEqual([
			"map-title",
			"flow-title",
			"state-title",
			"profile-title",
			"seal-title",
		]);
	});

	test("draws every part of the map once, as a choice named by the part's name and labelled by the part on the map, the whole map chosen at first", () => {
		const picks = [...service.querySelectorAll('input[name="service-part"]')];

		expect(picks.map((pick) => pick.id)).toEqual([
			...parts.map(pickId),
			wholeMap,
		]);
		expect(picks.every((pick) => pick.getAttribute("type") === "radio")).toBe(
			true,
		);
		expect(
			picks
				.filter((pick) => pick.hasAttribute("checked"))
				.map((pick) => pick.id),
		).toEqual([wholeMap]);
		expect(all("#pick-adult", "aria-label")).toEqual(["Adult"]);
		expect(all(".s-service-columns label", "for")).toEqual(parts.map(pickId));
		expect(all(".s-service-panel > input", "id")).toEqual(
			picks.map((pick) => pick.id),
		);
		expect(all(`#${wholeMap}`, "aria-label")).toEqual(["Nothing chosen"]);
	});

	test("lists every tool of the service on the map's part of the tools", () => {
		expect(texts(`#${faceId("tools")} code`)).toEqual([...tools]);
	});

	test("offers a few parts before any is chosen", () => {
		expect(all(".s-service-hint label", "for")).toEqual(quickPicks.map(pickId));
	});

	test("tells each part under the map, with every part it talks to and what passes between them, each a choice of that part", () => {
		for (const part of parts) {
			const about = service.getElementById(`about-${part}`);

			expect(about, part).not.toBeNull();
			expect(
				[...(about?.querySelectorAll(".s-service-neighbours label") ?? [])].map(
					(label) => label.getAttribute("for"),
				),
				part,
			).toEqual(neighbours(part).map(({ part: other }) => pickId(other)));
			expect(
				about?.querySelector(".s-service-reset")?.getAttribute("for"),
			).toBe(wholeMap);
		}
		expect(texts("#about-cimd .s-service-about-name")).toEqual([
			"Client metadata document",
		]);
		expect(texts("#about-cimd .s-service-about-group")).toEqual(["Chat host"]);
		expect(texts("#about-cimd .s-service-passes")).toEqual([
			"fetches the client's document",
		]);
	});

	test("draws every scenario's steps in the data's order, the arrows numbered and the notes not, each between its actors", () => {
		for (const scenario of scenarios) {
			const steps = service.getElementById(stepsId(scenario.id));
			const rows = [...(steps?.querySelectorAll(".s-service-row") ?? [])];
			const arrows = scenario.steps.filter((step) => step.kind !== "note");

			expect(rows, scenario.id).toHaveLength(scenario.steps.length);
			expect(
				rows.map((row) =>
					declared(row.firstElementChild?.getAttribute("style"), "grid-column"),
				),
				scenario.id,
			).toEqual(scenario.steps.map((step) => gridColumn(step)));
			expect(
				[...(steps?.querySelectorAll(".s-service-step") ?? [])].map(
					(number) => number.textContent,
				),
				scenario.id,
			).toEqual(arrows.map((_, at) => String(at + 1)));
			expect(
				steps?.querySelectorAll(".s-service-actor"),
				scenario.id,
			).toHaveLength(scenario.actors.length);
			expect(
				["--actors", "--halves"].map((property) =>
					declared(
						steps?.querySelector(".s-service-diagram")?.getAttribute("style"),
						property,
					),
				),
				scenario.id,
			).toEqual([
				String(scenario.actors.length),
				String(scenario.actors.length * 2),
			]);
		}
	});

	test("tells a screen reader whom each step goes from and to", () => {
		expect(
			texts(`#${stepsId("task")} .s-service-arrow-words .s-hidden`).slice(0, 2),
		).toEqual(["Child → The host's model: ", "The host's model → MathTrail: "]);
	});

	test("chooses the first scenario when the page opens, and offers the others", () => {
		expect(all('input[name="service-scenario"]', "id")).toEqual(
			scenarios.map(({ id }) => scenarioId(id)),
		);
		expect(all('input[name="service-scenario"][checked]', "value")).toEqual([
			"task",
		]);
	});

	test("says the file's size both usually and near its caps, in the page's language", () => {
		expect(texts(".s-service-total .s-service-typical")).toEqual(["≈ 61 KB"]);
		expect(texts(".s-service-total .s-service-cap")).toEqual(["≈ 79 KB"]);
		expect(
			texts(".s-service-blocks tbody tr:last-child td:last-child span"),
		).toEqual(["0.6 KB", " / ", "1.5 KB"]);
		expect(
			texts(
				".s-service-blocks tbody tr:last-child td:last-child span",
				russianService,
			),
		).toEqual(["0,6 КБ", " / ", "1,5 КБ"]);
		expect(all('input[name="service-size"][checked]', "value")).toEqual([
			"typical",
		]);
	});

	test("draws each block's stretch of the bar at both its sizes, and marks the size the file was meant to keep within", () => {
		expect(styleOf(".s-service-segment", "--typical")).toBe("33.75%");
		expect(styleOf(".s-service-segment", "--cap")).toBe("36.25%");
		expect(all(".s-service-segment", "title")[0]).toBe(
			"Fingerprints · 27 / 29 KB",
		);
		expect(styleOf(".s-service-goal", "inset-inline-start")).toBe("80%");
		expect(texts(".s-service-scale span")).toEqual([
			"0",
			"20 KB",
			"40 KB",
			"60 KB",
			"80 KB",
		]);
	});

	test("names the tools that write each block of the file", () => {
		expect(
			texts(".s-service-blocks tbody tr:first-child .s-service-writers > *"),
		).toEqual(["next_task", "submit_task"]);
		expect(
			texts(".s-service-blocks tbody tr:last-child .s-service-writers > *"),
		).toEqual(["save_profile", "edit_profile", "every write"]);
	});

	test("lays the lifetimes on the scale of the logarithm of the time", () => {
		expect(
			stylesOf(".s-service-life-mark", "inset-inline-start").slice(0, 3),
		).toEqual(["32.85%", "58.34%", "73.95%"]);
		expect(
			stylesOf(".s-service-marks span", "inset-inline-start").slice(1, -1),
		).toEqual(["32.85%", "58.34%", "73.95%"]);
		expect(service.querySelectorAll(".s-service-life-mark")).toHaveLength(18);
		expect(stylesOf(".s-service-life-bar", "--lives")).toEqual([
			"0%",
			"18.47%",
			"21.72%",
			"85.63%",
			"94.44%",
			"100%",
		]);
		expect(stylesOf(".s-service-life-more", "--more")).toEqual(["8.81%"]);
	});

	test("takes the task on the card through its states by the service's calls, the one past the checks ticked", () => {
		expect(texts(".s-service-states > code")).toEqual([
			"none",
			"requested",
			"issued",
			"answered",
		]);
		expect(texts(".s-service-forward code")).toEqual([
			"next_task",
			"submit_task ✓",
			"submit_answer",
		]);
		expect(texts(".s-service-turns code")).toHaveLength(7);
	});

	test("names every purpose a seal is cut for", () => {
		expect(
			texts(".s-service-token-legend .s-service-token-purpose dd"),
		).toEqual([
			"c code · a access · r refresh · d client · s state · k consent · t a task's answer",
		]);
	});

	test("gives every element its own id, every choice's label a choice the page has, and every frame a heading it is named by", () => {
		const ids = all("[id]", "id");

		expect(new Set(ids).size).toBe(ids.length);
		for (const target of all("label[for]", "for")) {
			expect(ids, target).toContain(target);
		}
		for (const target of all("[aria-labelledby]", "aria-labelledby")) {
			expect(ids, target).toContain(target);
		}
	});

	test("is the menu's entry in the group of the technical pages, marked as the page being read", () => {
		expect(all(".s-navlinks > .s-nav-group", "role")).toEqual(["group"]);
		expect(all(".s-navlinks > .s-nav-group", "aria-label")).toEqual([
			"Technical pages",
		]);
		expect(all(".s-nav-group a", "href")).toEqual([
			"/en/service/",
			"/en/service/",
		]);
		expect(all(".s-navlinks .s-nav-group a", "aria-current")).toEqual(["page"]);
		expect(texts(".s-navlinks .s-nav-group a")).toEqual(["Service"]);
		expect(texts(".s-navlinks .s-nav-group a", russianService)).toEqual([
			"Сервис",
		]);
	});

	test("is refused when its words lack a step the service has", () => {
		const lacking = english.replace(
			"      - text · drawing · five options · hint — no answer\n",
			"",
		);

		expect(lacking).not.toBe(english);
		expect(() => render(sourcesWith(lacking, lacking))).toThrow(
			"the page reads flow.task.rows.14",
		);
	});

	test("is refused when its words give a step the service lacks", () => {
		const more = english.replace(
			"      - text · drawing · five options · hint — no answer\n",
			"      - text · drawing · five options · hint — no answer\n      - one more\n",
		);

		expect(more).not.toBe(english);
		expect(() => render(sourcesWith(more, more))).toThrow(
			"the page never shows flow.task.rows.15",
		);
	});

	test("is refused when its Russian words lack one the English have", () => {
		const link = linkKey(["authsrv", "cimd"]);
		const lacking = russian.replace(
			`    ${link}: загружает документ клиента\n`,
			"",
		);

		expect(lacking).not.toBe(russian);
		expect(() => render(sourcesWith(english, lacking))).toThrow(
			`map.links.${link}`,
		);
	});
});
