// @vitest-environment node
import { describe, expect, test } from "vitest";
import { isExternal, locate, parsePage, resolveReference } from "./page.ts";

describe("resolveReference", () => {
	test.for([
		// An address with a trailing slash is a directory, and its index is the page.
		["/en/privacy/", "en/index.html", "en/privacy/index.html"],
		// So is one without: a page's address need not end in a slash.
		["/en/privacy", "en/index.html", "en/privacy/index.html"],
		// The fragment and the query are cut off before the address is read.
		["/en/privacy/#top", "en/index.html", "en/privacy/index.html"],
		["/en/privacy?from=footer", "en/index.html", "en/privacy/index.html"],
		["/en/#lesson", "ru/index.html", "en/index.html"],
		// An anchor within the page, or an empty address, asks for no file.
		["#lesson", "en/index.html", ""],
		["?q=1", "en/index.html", ""],
		// A file with an extension is that file.
		["/assets/style.css", "en/privacy/index.html", "assets/style.css"],
		// The apex.
		["/", "en/index.html", "index.html"],
		// Relative addresses are read from the page's own directory.
		["privacy/", "en/index.html", "en/privacy/index.html"],
		["../terms/", "en/privacy/index.html", "en/terms/index.html"],
		["../../ru/", "en/privacy/index.html", "ru/index.html"],
		[
			"topics/knights-and-liars/",
			"en/index.html",
			"en/topics/knights-and-liars/index.html",
		],
	])("reads %s on %s as %s", ([value, from, want]) => {
		expect(resolveReference(value ?? "", from ?? "")).toBe(want);
	});
});

describe("isExternal", () => {
	test.for([
		["https://github.com/example", true],
		["http://example.com/", true],
		["//cdn.example.com/x.js", true],
		["mailto:someone@example.com", true],
		["/en/", false],
		["assets/style.css", false],
		["#top", false],
		// Only a letter opens a scheme: a colon after a digit or a slash is part
		// of a name.
		["1st:place", false],
		["a/b:c", false],
		[":", false],
	] as const)("says %s leaves the site: %s", ([value, want]) => {
		expect(isExternal(value)).toBe(want);
	});
});

describe("locate", () => {
	test.for([
		["index.html", { locale: "", name: "", address: "/" }],
		["en/index.html", { locale: "en", name: "", address: "/en/" }],
		[
			"en/privacy/index.html",
			{ locale: "en", name: "privacy", address: "/en/privacy/" },
		],
		[
			"ru/topics/knights-and-liars/index.html",
			{
				locale: "ru",
				name: "topics/knights-and-liars",
				address: "/ru/topics/knights-and-liars/",
			},
		],
	] as const)("places %s", ([file, want]) => {
		expect(locate(file)).toEqual(want);
	});
});

describe("parsePage", () => {
	test("takes nothing of the page's head from a template, and still counts what the template loads", () => {
		const page = parsePage(
			"en/index.html",
			[
				`<!DOCTYPE html><html lang="en" dir="ltr"><head>`,
				`<template><title>Drawn later</title></template>`,
				`<title>Page</title>`,
				`<meta name="description" content="The page.">`,
				`<link rel="canonical" href="https://example.test/en/">`,
				`<link rel="alternate" hreflang="ru" href="https://example.test/ru/">`,
				`</head><body><template>`,
				`<meta name="description" content="Another page.">`,
				`<link rel="canonical" href="https://example.test/elsewhere/">`,
				`<link rel="alternate" hreflang="fr" href="https://example.test/fr/">`,
				`<link rel="stylesheet" href="/assets/demo.css">`,
				`</template></body></html>`,
			].join(""),
		);

		expect(page).toMatchObject({
			title: "Page",
			description: "The page.",
			canonical: "https://example.test/en/",
		});
		expect([...page.alternates]).toEqual([["ru", "https://example.test/ru/"]]);
		expect(page.references.map((ref) => ref.value)).toEqual([
			"/assets/demo.css",
		]);
	});

	test("reads the head and sorts every reference by what the browser does with it", () => {
		const page = parsePage(
			"en/index.html",
			[
				`<!DOCTYPE html><html lang="en" dir="ltr"><head>`,
				`<title> MathTrail </title>`,
				`<meta name="Description" content=" Olympiad maths ">`,
				`<link rel="canonical" href="https://example.test/en/">`,
				`<link rel="alternate" hreflang="ru" href="https://example.test/ru/">`,
				`<link rel="alternate" href="https://example.test/feed.xml">`,
				`<link rel="Stylesheet" href="/assets/style.css">`,
				`<link rel="author" href="/humans.txt">`,
				`</head><body>`,
				`<a href="/ru/">Русский</a><a>no address</a>`,
				`<img src="/assets/mark.svg" alt="">`,
				`<template><div id="drawn-later"><script src="/assets/demo.js"></script></div></template>`,
				`<section id="connect"><h2 id="connect-title">Connect</h2></section>`,
				`<svg><title>A map</title></svg>`,
				`</body></html>`,
			].join(""),
		);

		expect(page).toMatchObject({
			address: "/en/",
			locale: "en",
			name: "",
			lang: "en",
			dir: "ltr",
			title: "MathTrail",
			description: "Olympiad maths",
			canonical: "https://example.test/en/",
		});
		expect([...page.alternates]).toEqual([["ru", "https://example.test/ru/"]]);
		expect([...page.ids]).toEqual(["connect", "connect-title"]);
		expect(
			page.references.map(
				(ref) => `${ref.element} ${ref.value} ${ref.subresource}`,
			),
		).toEqual([
			"link /assets/style.css true",
			"link /humans.txt false",
			"a /ru/ false",
			"img /assets/mark.svg true",
			"script /assets/demo.js true",
		]);
	});
});
