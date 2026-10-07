import { describe, expect, test } from "vitest";
import { readTexts } from "./content";
import { cname, robots, sitemap } from "./metadata";

const base = "https://example.test";

// The pages of a small site, in English and Russian: a front page and a
// topic's page below the topics.
const texts = readTexts(
	new Map(
		["en", "ru"].map((locale) => [
			locale,
			new Map([
				["index.md", "---\ntitle: Home\ndescription: D\n---\n"],
				["topics/sample.yaml", "title: Sample\ndescription: D\n"],
			]),
		]),
	),
	"en",
);

describe("the files a crawler and the host read", () => {
	test("list every page with its translations in the sitemap, the English front page at the bare domain and not at the address it moved from, which only sends its reader on", () => {
		expect(sitemap(texts, base)).toBe(
			[
				'<?xml version="1.0" encoding="UTF-8"?>',
				'<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9" xmlns:xhtml="http://www.w3.org/1999/xhtml">',
				"  <url>",
				"    <loc>https://example.test/</loc>",
				'    <xhtml:link rel="alternate" hreflang="en" href="https://example.test/"/>',
				'    <xhtml:link rel="alternate" hreflang="ru" href="https://example.test/ru/"/>',
				'    <xhtml:link rel="alternate" hreflang="x-default" href="https://example.test/"/>',
				"  </url>",
				"  <url>",
				"    <loc>https://example.test/en/topics/sample/</loc>",
				'    <xhtml:link rel="alternate" hreflang="en" href="https://example.test/en/topics/sample/"/>',
				'    <xhtml:link rel="alternate" hreflang="ru" href="https://example.test/ru/topics/sample/"/>',
				'    <xhtml:link rel="alternate" hreflang="x-default" href="https://example.test/en/topics/sample/"/>',
				"  </url>",
				"  <url>",
				"    <loc>https://example.test/ru/</loc>",
				'    <xhtml:link rel="alternate" hreflang="en" href="https://example.test/"/>',
				'    <xhtml:link rel="alternate" hreflang="ru" href="https://example.test/ru/"/>',
				'    <xhtml:link rel="alternate" hreflang="x-default" href="https://example.test/"/>',
				"  </url>",
				"  <url>",
				"    <loc>https://example.test/ru/topics/sample/</loc>",
				'    <xhtml:link rel="alternate" hreflang="en" href="https://example.test/en/topics/sample/"/>',
				'    <xhtml:link rel="alternate" hreflang="ru" href="https://example.test/ru/topics/sample/"/>',
				'    <xhtml:link rel="alternate" hreflang="x-default" href="https://example.test/en/topics/sample/"/>',
				"  </url>",
				"</urlset>",
				"",
			].join("\n"),
		);
	});

	test("let every crawler in but to the coach's prototype and the photographs, and point it at the sitemap", () => {
		expect(robots(base)).toBe(
			"User-agent: *\nAllow: /\nDisallow: /assets/coach-prototype.html\nDisallow: /assets/photos/\n\nSitemap: https://example.test/sitemap.xml\n",
		);
	});

	test("claim the domain the site is published on", () => {
		expect(cname(new URL("https://mathtrail.app"))).toBe("mathtrail.app\n");
		expect(cname(new URL("http://localhost:8081"))).toBe("localhost:8081\n");
	});
});
