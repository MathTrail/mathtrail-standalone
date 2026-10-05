// @vitest-environment node
import { mkdtemp, rm, symlink, unlink } from "node:fs/promises";
import { tmpdir } from "node:os";
import { join } from "node:path";
import { afterEach, describe, expect, test } from "vitest";
import { check, type Finding, lineOf, type Options } from "./check.ts";
import { base, pageAt, redirectAt, siteAt, writeSite } from "./testing/site.ts";

const addresses = ["/", "/en/", "/ru/"];

function options(): Options {
	return {
		base,
		referenceLocale: "en",
		maxPageBytes: 4096,
		maxFrameBytes: 4096,
		photos: "/assets/photos/",
		published: addresses,
	};
}

// showing makes the English front page show the picture at src, and gives the
// site a picture past the budget in its photo directory and in another one
// whose name only begins the same.
function showing(src: string) {
	return (files: Record<string, string>): void => {
		files["en/index.html"] = pageAt("/en/", addresses).replace(
			"</body>",
			`<img src="${src}" alt="Us"></body>`,
		);
		files["assets/photos/us.webp"] = "x".repeat(8192);
		files["assets/photos-old/us.webp"] = "x".repeat(8192);
	};
}

// framing makes the page at each of framers frame the document at src.
function framing(
	files: Record<string, string>,
	src: string,
	framers: readonly string[] = ["/en/"],
): void {
	for (const address of framers) {
		files[`${address.slice(1)}index.html`] = pageAt(address, addresses).replace(
			"</body>",
			`<iframe src="${src}" title="Demo"></iframe></body>`,
		);
	}
}

// withPicture is page with the picture a shared link to it shows named in its
// head.
function withPicture(page: string, picture: string): string {
	return page.replace(
		"</head>",
		`<meta property="og:image" content="${picture}"></head>`,
	);
}

const made: string[] = [];

// site writes a site that breaks no rule, changed by change, and returns its
// directory.
async function site(
	change: (files: Record<string, string>) => void = () => {},
): Promise<string> {
	const dir = await mkdtemp(join(tmpdir(), "sitecheck-"));
	made.push(dir);
	const files = siteAt(addresses);
	change(files);
	await writeSite(dir, files);
	return dir;
}

afterEach(async () => {
	await Promise.all(
		made.splice(0).map((dir) => rm(dir, { recursive: true, force: true })),
	);
});

describe("check", () => {
	test("passes a site with nothing wrong", async () => {
		expect(await check(await site(), options())).toEqual([]);
	});

	test("passes a page whose shared link shows a picture the site serves", async () => {
		const dir = await site((files) => {
			files["en/index.html"] = withPicture(
				pageAt("/en/", addresses),
				`${base}/assets/og-en.png`,
			);
			files["assets/og-en.png"] = "png";
		});

		expect(await check(dir, options())).toEqual([]);
	});

	const broken: {
		name: string;
		change?: (files: Record<string, string>) => void;
		adjust?: (options: Options) => Options;
		rule: Finding["rule"];
		contains: string;
	}[] = [
		{
			name: "a link to a page that was never built",
			change: (files) => {
				files["en/index.html"] = pageAt("/en/", addresses).replace(
					"</body>",
					`<a href="/en/privacy/">Privacy</a></body>`,
				);
			},
			rule: "link",
			contains: "en/privacy/index.html",
		},
		{
			name: "a stylesheet loaded from somebody else's domain",
			change: (files) => {
				files["en/index.html"] = pageAt("/en/", addresses).replace(
					`href="/assets/style.css"`,
					`href="https://cdn.example.com/style.css"`,
				);
			},
			rule: "external",
			contains: "another origin",
		},
		{
			name: "a font pulled in from inside the stylesheet",
			change: (files) => {
				files["assets/style.css"] =
					"@import url(https://fonts.example.com/x.css);";
			},
			rule: "external",
			contains: "stylesheet mentions",
		},
		{
			name: "a page with no description",
			change: (files) => {
				files["ru/index.html"] = pageAt("/ru/", addresses).replace(
					`<meta name="description" content="Description">`,
					"",
				);
			},
			rule: "head",
			contains: "no description",
		},
		{
			name: "a page whose only title is a drawing's",
			change: (files) => {
				files["ru/index.html"] = pageAt("/ru/", addresses)
					.replace("<title>Title</title>", "")
					.replace("</body>", "<svg><title>A map</title></svg></body>");
			},
			rule: "head",
			contains: "no title",
		},
		{
			name: "a picture for a shared link the site does not have",
			change: (files) => {
				files["en/index.html"] = withPicture(
					pageAt("/en/", addresses),
					`${base}/assets/og-en.png`,
				);
			},
			rule: "head",
			contains: "is a file the site does not have",
		},
		{
			name: "a picture for a shared link the site does not have, its property in capitals, before one it has",
			change: (files) => {
				files["en/index.html"] = withPicture(
					withPicture(pageAt("/en/", addresses), `${base}/assets/og-en.png`),
					`${base}/assets/style.css`,
				).replace("og:image", "OG:image");
			},
			rule: "head",
			contains: 'og-en.png", is a file the site does not have',
		},
		{
			name: "a picture for a shared link on another origin",
			change: (files) => {
				files["en/index.html"] = withPicture(
					pageAt("/en/", addresses),
					"https://cdn.example.com/og-en.png",
				);
			},
			rule: "head",
			contains: `is not on ${base}`,
		},
		{
			name: "a canonical that points at another page",
			change: (files) => {
				files["ru/index.html"] = pageAt("/ru/", addresses).replace(
					`${base}/ru/">`,
					`${base}/en/">`,
				);
			},
			rule: "head",
			contains: "canonical is",
		},
		{
			name: "a page that sends its reader on and names itself as its canonical",
			change: (files) => {
				files["index.html"] = redirectAt("/en/").replace(
					`href="${base}/en/"`,
					`href="${base}/"`,
				);
			},
			rule: "head",
			contains: `canonical is "${base}/", and the page sends its reader on to "${base}/en/"`,
		},
		{
			name: "a page that sends its reader on to a page the site does not have",
			change: (files) => {
				files["index.html"] = redirectAt("/fr/");
			},
			rule: "head",
			contains: `sends its reader on to "/fr/", which is no page of the site`,
		},
		{
			name: "a page that sends its reader on to a page the site does not have, by its whole address",
			change: (files) => {
				files["index.html"] = redirectAt("/fr/").replace(
					"url=/fr/",
					`url=${base}/fr/`,
				);
			},
			rule: "head",
			contains: `sends its reader on to "${base}/fr/", which is no page of the site`,
		},
		{
			name: "a page that sends its reader on to another origin",
			change: (files) => {
				files["index.html"] = redirectAt("/en/").replace(
					"url=/en/",
					"url=https://other.example/en/",
				);
			},
			rule: "head",
			contains: `sends its reader on to "https://other.example/en/", which is no page of the site`,
		},
		{
			name: "a page that sends its reader on to itself",
			change: (files) => {
				files["index.html"] = redirectAt("/");
			},
			rule: "head",
			contains: `sends its reader on to "/", which is the page itself`,
		},
		{
			name: "a page of a language that sends its reader on",
			change: (files) => {
				files["ru/index.html"] = redirectAt("/en/").replace(
					`lang="en"`,
					`lang="ru"`,
				);
			},
			rule: "head",
			contains: `sends its reader on to "/en/", and only the apex may`,
		},
		{
			name: "an apex that waits before it sends its reader on, judged as the page it is",
			change: (files) => {
				files["index.html"] = redirectAt("/en/").replace(
					`content="0;`,
					`content="5;`,
				);
			},
			rule: "head",
			contains: `canonical is "${base}/en/", and the page is served at "${base}/"`,
		},
		{
			name: "a front page that sends a reader with no matching language to the apex",
			change: (files) => {
				files["ru/index.html"] = pageAt("/ru/", addresses).replace(
					`hreflang="x-default" href="${base}/en/"`,
					`hreflang="x-default" href="${base}/"`,
				);
			},
			rule: "head",
			contains: `alternate for "x-default" is "${base}/", and that page is at "${base}/en/"`,
		},
		{
			name: "a lang that disagrees with the address",
			change: (files) => {
				files["ru/index.html"] = pageAt("/ru/", addresses).replace(
					`lang="ru"`,
					`lang="en"`,
				);
			},
			rule: "head",
			contains: "disagrees with the locale",
		},
		{
			name: "a translation the page never points at",
			change: (files) => {
				files["en/index.html"] = pageAt("/en/", ["/en/"]);
			},
			rule: "head",
			contains: `no alternate for "ru"`,
		},
		{
			name: "a locale that lost a page the reference locale has",
			change: (files) => {
				files["en/privacy/index.html"] = pageAt("/en/privacy/", [
					"/en/privacy/",
				]);
			},
			rule: "translation",
			contains: `locale "ru" is missing`,
		},
		{
			name: "a page that costs the reader more than the budget",
			adjust: (options) => ({ ...options, maxPageBytes: 64 }),
			rule: "weight",
			contains: "the budget is 64",
		},
		{
			name: "a page whose stylesheet loads a font past the budget",
			change: (files) => {
				files["assets/style.css"] =
					"@font-face{font-family:A;src:url(/assets/a.woff2)}";
				files["assets/a.woff2"] = "x".repeat(4096);
			},
			rule: "weight",
			contains: "the budget is 4096",
		},
		{
			name: "a framed document past a budget of its own",
			change: (files) => {
				framing(files, "/assets/demo.html");
				files["assets/demo.html"] = "x".repeat(5000);
			},
			rule: "weight",
			contains:
				"assets/demo.html: weight: a page frames it, and with what it loads it comes to 5000 bytes; the budget of a framed document is 4096",
		},
		{
			name: "a frame of a file the site does not have",
			change: (files) => {
				framing(files, "/assets/gone.html");
			},
			rule: "link",
			contains: `<iframe src="/assets/gone.html"> leads to assets/gone.html`,
		},
		{
			name: "a framed document that loads from another origin",
			change: (files) => {
				framing(files, "/assets/demo.html");
				files["assets/demo.html"] =
					'<script src="https://cdn.example.test/x.js"></script>';
			},
			rule: "external",
			contains: `assets/demo.html: external: <script src="https://cdn.example.test/x.js"> loads from another origin`,
		},
		{
			name: "a file a framed document loads that the site does not have",
			change: (files) => {
				framing(files, "/assets/demo.html");
				files["assets/demo.html"] = '<img src="missing.png" alt="">';
			},
			rule: "link",
			contains: `assets/demo.html: link: <img src="missing.png"> leads to assets/missing.png`,
		},
		{
			name: "a font a stylesheet loads that the site does not have",
			change: (files) => {
				files["assets/style.css"] =
					'@font-face{font-family:A;src:url("fonts/a.woff2")}';
			},
			rule: "link",
			contains: `assets/style.css: link: the stylesheet loads "fonts/a.woff2", which leads to assets/fonts/a.woff2`,
		},
		{
			name: "a page the site serves and its list of published addresses does not hold",
			change: (files) => {
				files["en/why/index.html"] = pageAt("/en/why/", ["/en/why/"]);
			},
			rule: "published",
			contains:
				"the site serves /en/why/, which the list of published addresses does not hold",
		},
		{
			name: "an address the site handed out and no longer has",
			adjust: (options) => ({
				...options,
				published: [...addresses, "/en/privacy/"],
			}),
			rule: "published",
			contains: "/en/privacy/",
		},
		{
			name: "an anchor the site handed out that its page no longer holds",
			adjust: (options) => ({
				...options,
				published: [...addresses, "/en/#connect"],
			}),
			rule: "published",
			contains: `no element with the id "connect"`,
		},
	];

	test.for(broken)(
		"names $name",
		async ({ change, adjust, rule, contains }) => {
			const findings = await check(
				await site(change),
				adjust ? adjust(options()) : options(),
			);

			expect(
				findings.filter(
					(finding) =>
						finding.rule === rule && lineOf(finding).includes(contains),
				),
				`${findings.map(lineOf).join("\n")}`,
			).not.toEqual([]);
		},
	);

	test("keeps an anchor the site handed out while its page holds it", async () => {
		const dir = await site((files) => {
			files["en/index.html"] = pageAt("/en/", addresses).replace(
				"</body>",
				`<section id="connect">Connect</section></body>`,
			);
		});

		expect(
			await check(dir, {
				...options(),
				published: [...addresses, "/en/#connect"],
			}),
		).toEqual([]);
	});

	test("serves a linked file as the file it links to", async () => {
		const dir = await site((files) => {
			files["shared/style.css"] =
				"@import url(https://fonts.example.com/x.css);";
		});
		await unlink(join(dir, "assets", "style.css"));
		await symlink(
			join(dir, "shared", "style.css"),
			join(dir, "assets", "style.css"),
		);

		const findings = await check(dir, options());

		expect(findings.filter((finding) => finding.rule === "link")).toEqual([]);
		expect(findings).toContainEqual({
			path: "assets/style.css",
			rule: "external",
			message: `the stylesheet mentions "https://", so it may load from another origin`,
		});
	});

	test("weighs what a stylesheet loads through another, once each", async () => {
		const dir = await site((files) => {
			files["assets/style.css"] = '@import "more.css";';
			files["assets/more.css"] =
				"@font-face{src:url(a.woff2)}@font-face{src:url(/assets/a.woff2)}";
			// A page of some 500 bytes with this font is past 4096 bytes, and past
			// 6000 only if the font were counted for each of its two addresses.
			files["assets/a.woff2"] = "x".repeat(3600);
		});

		expect(
			(await check(dir, options())).filter(({ rule }) => rule === "weight"),
		).toEqual(
			["en/index.html", "ru/index.html"].map((path) => ({
				path,
				rule: "weight",
				message: expect.stringContaining("and the budget is 4096"),
			})),
		);
		expect(await check(dir, { ...options(), maxPageBytes: 6000 })).toEqual([]);
	});

	test("weighs a framed document apart from the page that frames it", async () => {
		const dir = await site((files) => {
			framing(files, "/assets/demo.html");
			files["assets/demo.html"] = "x".repeat(8000);
		});

		expect(await check(dir, { ...options(), maxFrameBytes: 8192 })).toEqual([]);
	});

	test("weighs a framed document once, however many pages frame it", async () => {
		const dir = await site((files) => {
			framing(files, "/assets/demo.html", ["/en/", "/ru/"]);
			files["assets/demo.html"] = "x".repeat(5000);
		});

		expect(
			(await check(dir, options())).filter(({ rule }) => rule === "weight"),
		).toEqual([
			{
				path: "assets/demo.html",
				rule: "weight",
				message: expect.stringContaining("comes to 5000 bytes"),
			},
		]);
	});

	test("weighs what a framed document loads with it, and what it frames in turn apart from it", async () => {
		const demo = '<img src="pic.png" alt=""><iframe src="inner.html"></iframe>';
		const dir = await site((files) => {
			framing(files, "/assets/demo.html");
			files["assets/demo.html"] = demo;
			files["assets/pic.png"] = "x".repeat(4100);
			files["assets/inner.html"] = "x".repeat(5000);
		});

		expect(
			(await check(dir, options())).filter(({ rule }) => rule === "weight"),
		).toEqual([
			{
				path: "assets/demo.html",
				rule: "weight",
				message: expect.stringContaining(
					`comes to ${demo.length + 4100} bytes`,
				),
			},
			{
				path: "assets/inner.html",
				rule: "weight",
				message: expect.stringContaining("comes to 5000 bytes"),
			},
		]);
	});

	test("weighs a page without the photographs it shows, which are worth their weight", async () => {
		const dir = await site(showing("/assets/photos/us.webp"));

		expect(await check(dir, options())).toEqual([]);
	});

	test("weighs a page with any other picture it shows, and with every one when it is given no photo directory", async () => {
		const weighed = (findings: Finding[]) =>
			findings.map(({ path, rule }) => `${path}: ${rule}`);

		expect(
			weighed(
				await check(
					await site(showing("/assets/photos-old/us.webp")),
					options(),
				),
			),
		).toEqual(["en/index.html: weight"]);
		expect(
			weighed(
				await check(await site(showing("/assets/photos/us.webp")), {
					...options(),
					photos: "",
				}),
			),
		).toEqual(["en/index.html: weight"]);
	});

	test("judges a page that another page frames as a page, not as a frame", async () => {
		const dir = await site((files) => {
			framing(files, "/ru/");
		});

		expect(await check(dir, { ...options(), maxFrameBytes: 64 })).toEqual([]);
	});

	test("reads nothing in a framed file that is no HTML, and weighs it by its size", async () => {
		const picture =
			'<img src="https://cdn.example.test/x.png" alt=""><a href="gone.html">';
		const dir = await site((files) => {
			framing(files, "/assets/picture.png");
			files["assets/picture.png"] = picture;
		});

		expect(await check(dir, options())).toEqual([]);
		expect(await check(dir, { ...options(), maxFrameBytes: 16 })).toEqual([
			{
				path: "assets/picture.png",
				rule: "weight",
				message: expect.stringContaining(`comes to ${picture.length} bytes`),
			},
		]);
	});

	test("leaves an address of another origin in a stylesheet to the stylesheet rule", async () => {
		const dir = await site((files) => {
			files["assets/style.css"] =
				"@font-face{src:url(https://fonts.example.com/a.woff2)}";
		});

		expect(
			(await check(dir, options())).map(({ path, rule }) => `${path}: ${rule}`),
		).toEqual(["assets/style.css: external"]);
	});

	test("names a page the build lost, though the link rule cannot see it", async () => {
		const dir = await site();
		await unlink(join(dir, "ru", "index.html"));

		const findings = await check(dir, options());

		expect(findings).toContainEqual({
			path: "ru/index.html",
			rule: "published",
			message: "the site has published /ru/, and the build no longer has it",
		});
	});

	test("judges a page whatever its markup is, and never throws over it", async () => {
		for (const document of ["<html><body><a href=", "", "\u0000<<<>>>\uffff"]) {
			const dir = await site((files) => {
				files["en/index.html"] = document;
			});

			await expect(check(dir, options())).resolves.toBeInstanceOf(Array);
		}
	});

	test("reports in the order of file, rule and message", async () => {
		const dir = await site((files) => {
			files["ru/index.html"] = pageAt("/ru/", addresses)
				.replace(`<meta name="description" content="Description">`, "")
				.replace("<title>Title</title>", "");
			files["en/index.html"] = pageAt("/en/", addresses).replace(
				"</body>",
				`<a href="/en/gone/">Gone</a></body>`,
			);
		});

		expect((await check(dir, options())).map(lineOf)).toEqual([
			`en/index.html: link: <a href="/en/gone/"> leads to en/gone/index.html, which the site does not have`,
			"ru/index.html: head: the page has no description",
			"ru/index.html: head: the page has no title",
		]);
	});

	const unreadable: {
		name: string;
		files: Record<string, string>;
		options: Options;
		message: string;
	}[] = [
		{
			name: "no base to measure addresses against",
			files: siteAt(addresses),
			options: { ...options(), base: "" },
			message: "base URL is empty",
		},
		{
			name: "no locale for the others to match",
			files: siteAt(addresses),
			options: { ...options(), referenceLocale: "" },
			message: "reference locale is empty",
		},
		{
			name: "an empty directory, which is not a site that passed",
			files: {},
			options: options(),
			message: "the site is empty",
		},
		{
			name: "files and no page among them",
			files: { "assets/style.css": "body{}" },
			options: options(),
			message: "the site holds no page",
		},
		{
			name: "a reference locale that was never built",
			files: siteAt(addresses),
			options: { ...options(), referenceLocale: "de" },
			message: `reference locale "de" has no front page`,
		},
		{
			name: "a base with a slash at its end",
			files: siteAt(addresses),
			options: { ...options(), base: `${base}/` },
			message: `base URL "${base}/" is not an origin`,
		},
		{
			name: "a base with a path",
			files: siteAt(addresses),
			options: { ...options(), base: `${base}/site` },
			message: `base URL "${base}/site" is not an origin`,
		},
		{
			name: "a published address that is no page's",
			files: siteAt(addresses),
			options: { ...options(), published: [...addresses, "/en/privacy"] },
			message: `published address "/en/privacy" is no page's address`,
		},
		{
			name: "a photo directory with no slash at its end, which would take in its namesakes",
			files: siteAt(addresses),
			options: { ...options(), photos: "/assets/photos" },
			message: `photo directory "/assets/photos" is no directory of the site`,
		},
		{
			name: "a photo directory that does not start at the site's root",
			files: siteAt(addresses),
			options: { ...options(), photos: "assets/photos/" },
			message: `photo directory "assets/photos/" is no directory of the site`,
		},
	];

	test.for(unreadable)(
		"refuses to judge $name",
		async ({ files, options, message }) => {
			const dir = await mkdtemp(join(tmpdir(), "sitecheck-"));
			made.push(dir);
			await writeSite(dir, files);

			await expect(check(dir, options)).rejects.toThrow(message);
		},
	);
});

describe("a script the site serves", () => {
	// scripted makes the English front page run the script at /assets/demo.js,
	// which says text.
	const scripted = (text: string) => (files: Record<string, string>) => {
		files["en/index.html"] = pageAt("/en/", addresses).replace(
			"</body>",
			'<script type="module" src="/assets/demo.js"></script></body>',
		);
		files["assets/demo.js"] = text;
		files["assets/chunk.js"] = "";
	};

	test.each([
		["names nothing it loads", 'document.querySelector("main");'],
		["reads where it runs from", "console.log(import.meta.url);"],
		[
			"says words that begin with import",
			'var important="x",imports=[];e.importNode(n);console.log(important,imports);',
		],
	])("passes when it %s", async (_, text) => {
		const dir = await site(scripted(text));

		expect(await check(dir, options())).toEqual([]);
	});

	test.each([
		["imports a module", 'import{a}from"./chunk.js";a();'],
		["imports a module by its name", 'import a from "./chunk.js";a();'],
		["imports a module for what it does", 'import"./chunk.js";'],
		["imports everything of a module", 'import*as c from"./chunk.js";c;'],
		["calls for a module", 'import("./chunk.js").then(()=>{});'],
		["passes a module on", 'export*from"./chunk.js";'],
		["passes names of a module on", 'export{a}from"./chunk.js";'],
	])(
		"is refused when it %s, which the checker does not follow",
		async (_, text) => {
			const dir = await site(scripted(text));

			expect(await check(dir, options())).toEqual([
				{
					path: "assets/demo.js",
					rule: "script",
					message:
						"the script loads another file, which the checker does not follow: a script of the site is one file, weighed whole",
				},
			]);
		},
	);

	test.each([
		["fetch", "fetch(u);"],
		["fetch", "const f=globalThis.fetch;f(u);"],
		["fetch", 'window["fetch"](u);'],
		["XMLHttpRequest", "new XMLHttpRequest;"],
		["sendBeacon", "navigator.sendBeacon(u);"],
		["WebSocket", "new WebSocket(u);"],
		["EventSource", "new EventSource(u);"],
		["new Image", "new Image().src=u;"],
		["serviceWorker", "navigator.serviceWorker.register(u);"],
		["localStorage", "localStorage.setItem(k,v);"],
		["sessionStorage", "sessionStorage.setItem(k,v);"],
		["indexedDB", "indexedDB.open(k);"],
		["cookie", "document.cookie=v;"],
	])("is refused when it names %s, as in %s", async (name, text) => {
		const dir = await site(scripted(text));

		expect((await check(dir, options())).map(lineOf)).toEqual([
			`assets/demo.js: script: the script names ${name}, and a script of the site stores nothing and sends nothing`,
		]);
	});
});
