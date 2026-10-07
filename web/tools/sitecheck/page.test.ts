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
		// The bare domain, with an anchor or a query on it as well.
		["/", "en/index.html", "index.html"],
		["/#connect", "en/why/index.html", "index.html"],
		["/?from=footer", "en/index.html", "index.html"],
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
		// The bare domain's is the reference locale's front page.
		["index.html", { locale: "en", name: "", address: "/" }],
		["en/index.html", { locale: "en", name: "", address: "/en/" }],
		["ru/index.html", { locale: "ru", name: "", address: "/ru/" }],
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
		expect(locate(file, "en")).toEqual(want);
	});

	test("places the bare domain's file in the reference locale it is given", () => {
		expect(locate("index.html", "ru")).toEqual({
			locale: "ru",
			name: "",
			address: "/",
		});
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
				`<meta http-equiv="refresh" content="0; url=/elsewhere/">`,
				`<link rel="canonical" href="https://example.test/elsewhere/">`,
				`<link rel="alternate" hreflang="fr" href="https://example.test/fr/">`,
				`<link rel="stylesheet" href="/assets/demo.css">`,
				`</template></body></html>`,
			].join(""),
			"en",
		);

		expect(page).toMatchObject({
			title: "Page",
			description: "The page.",
			canonical: "https://example.test/en/",
			refresh: undefined,
		});
		expect([...page.alternates]).toEqual([["ru", "https://example.test/ru/"]]);
		expect(page.references.map((ref) => ref.value)).toEqual([
			"/assets/demo.css",
		]);
	});

	// noBreak is a space HTML does not skip as white space, as a browser reads a
	// refresh.
	const noBreak = String.fromCodePoint(0xa0);

	test.for([
		["0; url=/en/", "/en/"],
		["0;URL='/en/'", "/en/"],
		[' 0 , url = "/en/" ', "/en/"],
		["0; /en/", "/en/"],
		["0;url='/en/", "/en/"],
		// A browser counts whole seconds alone.
		["0.5; url=https://example.test/en/", "https://example.test/en/"],
		[".5; url=/en/", "/en/"],
		// A refresh that waits first shows its page before it leaves.
		["5; url=/en/", ""],
		// A refresh that names no address only reloads the page.
		["0", ""],
		["5", ""],
		// A browser does not refresh at all without the time, with anything but
		// a separator after it, or past a space that is not HTML's.
		["url=/en/", undefined],
		["; url=/en/", undefined],
		["0x; url=/en/", undefined],
		[`0${noBreak}url=/en/`, undefined],
		[`${noBreak}0; url=/en/`, undefined],
		["", undefined],
	] as const)("reads a refresh of %j as doing %j", ([content, want]) => {
		expect(parsePage("index.html", refreshing(content), "en").refresh).toBe(
			want,
		);
	});

	test("acts on the first refresh a browser acts on, as a browser does", () => {
		expect(
			parsePage("index.html", refreshing("5", "0; url=/en/"), "en").refresh,
		).toBe("");
		expect(
			parsePage("index.html", refreshing("0; url=/en/", "5"), "en").refresh,
		).toBe("/en/");
		expect(
			parsePage("index.html", refreshing("url=/ru/", "0; url=/en/"), "en")
				.refresh,
		).toBe("/en/");
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
			"en",
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

// refreshing is a page whose head holds a refresh with each of contents, in
// order.
function refreshing(...contents: string[]): string {
	const metas = contents.map(
		(content) =>
			`<meta http-equiv="Refresh" content="${content.replaceAll('"', "&quot;")}">`,
	);
	return `<!DOCTYPE html><html><head>${metas.join("")}</head></html>`;
}
