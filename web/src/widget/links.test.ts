import { describe, expect, test } from "vitest";
import topics from "../../../content/catalogs/topics.json";
import { published } from "../../tools/sitecheck/published";
import { type Anchor, groupAddress, pageAddress } from "./links";
import { topicGroups } from "./topicGroups";

const site = { url: "https://mathtrail.app", languages: ["en", "ru"] };
const gaps = { slug: "gaps-and-boundaries", site_page: true };

describe("the address of a topic's page", () => {
	test("is the site, the card's language and the topic's slug, at the part asked for", () => {
		expect(pageAddress(site, "ru", gaps, "#traps")).toBe(
			"https://mathtrail.app/ru/topics/gaps-and-boundaries/#traps",
		);
		expect(pageAddress(site, "en", gaps, "")).toBe(
			"https://mathtrail.app/en/topics/gaps-and-boundaries/",
		);
	});

	test("is in English for a card in a language the site is not written in", () => {
		expect(pageAddress(site, "fr", gaps, "#home")).toBe(
			"https://mathtrail.app/en/topics/gaps-and-boundaries/#home",
		);
	});

	test.each([
		["no site", undefined, gaps],
		["a page not published", site, { slug: "gaps-and-boundaries" }],
		["a page published, with no slug", site, { site_page: true }],
		[
			"a slug that would turn the path",
			site,
			{ slug: "../privacy", site_page: true },
		],
		[
			"a slug with more than words",
			site,
			{ slug: "gaps?child=Comet", site_page: true },
		],
		["a site with no https", { ...site, url: "http://mathtrail.app" }, gaps],
		["a site with a path", { ...site, url: "https://mathtrail.app/en" }, gaps],
		[
			"a site with a query",
			{ ...site, url: "https://mathtrail.app?child=Comet" },
			gaps,
		],
		[
			"a site with credentials",
			{ ...site, url: "https://parent@mathtrail.app" },
			gaps,
		],
		[
			"a site written otherwise than its origin",
			{ ...site, url: "https://MathTrail.app" },
			gaps,
		],
		["no site at all, only a word", { ...site, url: "mathtrail" }, gaps],
	])("is none for %s", (_, given, page) => {
		expect(pageAddress(given, "en", page, "")).toBeUndefined();
	});

	// Every address the card can build for the catalog is one the site has
	// handed out and keeps: a link from a card that leads nowhere is a promise
	// broken in front of a parent. A topic whose page is not out yet has none.
	test("is one the site keeps, for every topic of the catalog whose page is out, in every language and at every part", () => {
		const anchors: Anchor[] = ["", "#traps", "#home"];
		for (const topic of topics) {
			for (const language of site.languages) {
				for (const anchor of anchors) {
					const address = pageAddress(site, language, topic, anchor);
					const where = `${topic.id} ${language} ${anchor}`;
					if (!topic.site_page) {
						expect(address, where).toBeUndefined();
						continue;
					}
					expect(address, where).toBeDefined();
					expect(published).toContain(
						address?.slice("https://mathtrail.app".length),
					);
				}
			}
		}
	});
});

describe("the address of a group of topics", () => {
	test("is the site's page of topics in the card's language, at the group's part", () => {
		expect(groupAddress(site, "ru", "time")).toBe(
			"https://mathtrail.app/ru/topics/#time",
		);
		expect(groupAddress(site, "fr", "games")).toBe(
			"https://mathtrail.app/en/topics/#games",
		);
	});

	test.each([
		["no site", undefined, "time"],
		["a group that would turn the address", site, "time?child=Comet"],
		["a group that names a path", site, "../privacy"],
		[
			"a site with a path",
			{ ...site, url: "https://mathtrail.app/en" },
			"time",
		],
		["a site with no https", { ...site, url: "http://mathtrail.app" }, "time"],
	])("is none for %s", (_, given, group) => {
		expect(groupAddress(given, "en", group)).toBeUndefined();
	});

	test("is one the site keeps, for every group a card offers, in every language", () => {
		for (const group of topicGroups) {
			for (const language of site.languages) {
				const address = groupAddress(site, language, group.id);
				expect(published, `${group.id} ${language}`).toContain(
					address?.slice("https://mathtrail.app".length),
				);
			}
		}
	});
});
