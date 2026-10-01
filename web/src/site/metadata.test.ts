import { describe, expect, test } from "vitest";
import { readTexts } from "./content";
import { cname, robots, sitemap } from "./metadata";

const base = "https://example.test";

// The pages of a small site: English has a front page and a privacy policy,
// Russian a front page alone.
const texts = readTexts(
	new Map([
		[
			"en",
			new Map([
				["index", "---\ntitle: Home\ndescription: D\n---\n"],
				["privacy", "---\ntitle: Privacy\ndescription: D\n---\n"],
			]),
		],
		["ru", new Map([["index", "---\ntitle: Главная\ndescription: D\n---\n"]])],
	]),
	"en",
);

describe("the files a crawler and the host read", () => {
	test("list every address with its translations in the sitemap, the apex first", () => {
		expect(sitemap(texts, base, "en")).toBe(
			[
				'<?xml version="1.0" encoding="UTF-8"?>',
				'<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9" xmlns:xhtml="http://www.w3.org/1999/xhtml">',
				"  <url>",
				"    <loc>https://example.test/</loc>",
				'    <xhtml:link rel="alternate" hreflang="en" href="https://example.test/en/"/>',
				'    <xhtml:link rel="alternate" hreflang="ru" href="https://example.test/ru/"/>',
				'    <xhtml:link rel="alternate" hreflang="x-default" href="https://example.test/"/>',
				"  </url>",
				"  <url>",
				"    <loc>https://example.test/en/</loc>",
				'    <xhtml:link rel="alternate" hreflang="en" href="https://example.test/en/"/>',
				'    <xhtml:link rel="alternate" hreflang="ru" href="https://example.test/ru/"/>',
				'    <xhtml:link rel="alternate" hreflang="x-default" href="https://example.test/"/>',
				"  </url>",
				"  <url>",
				"    <loc>https://example.test/en/privacy/</loc>",
				'    <xhtml:link rel="alternate" hreflang="en" href="https://example.test/en/privacy/"/>',
				'    <xhtml:link rel="alternate" hreflang="x-default" href="https://example.test/en/privacy/"/>',
				"  </url>",
				"  <url>",
				"    <loc>https://example.test/ru/</loc>",
				'    <xhtml:link rel="alternate" hreflang="en" href="https://example.test/en/"/>',
				'    <xhtml:link rel="alternate" hreflang="ru" href="https://example.test/ru/"/>',
				'    <xhtml:link rel="alternate" hreflang="x-default" href="https://example.test/"/>',
				"  </url>",
				"</urlset>",
				"",
			].join("\n"),
		);
	});

	test("let every crawler in and point it at the sitemap", () => {
		expect(robots(base)).toBe(
			"User-agent: *\nAllow: /\n\nSitemap: https://example.test/sitemap.xml\n",
		);
	});

	test("claim the domain the site is published on", () => {
		expect(cname(new URL("https://mathtrail.app"))).toBe("mathtrail.app\n");
		expect(cname(new URL("http://localhost:8081"))).toBe("localhost:8081\n");
	});
});
