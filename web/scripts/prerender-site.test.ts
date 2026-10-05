// @vitest-environment node
import { createHash } from "node:crypto";
import {
	mkdir,
	mkdtemp,
	readdir,
	readFile,
	rm,
	writeFile,
} from "node:fs/promises";
import { createRequire } from "node:module";
import { tmpdir } from "node:os";
import { dirname, join, relative } from "node:path";
import { gunzipSync } from "node:zlib";
import { type DefaultTreeAdapterTypes, parse } from "parse5";
import { afterAll, beforeAll, describe, expect, test } from "vitest";
import catalog from "../../content/catalogs/topics.json";
import data from "../../site/data.json";
import fixture from "../../site/research/testdata/research.json";
import widgetEnglish from "../locales/en.json";
import widgetRussian from "../locales/ru.json";
import { byCodeUnits } from "../src/i18n/order.ts";
import { photoDirectory, photoPath, photos } from "../src/site/brand.ts";
import english from "../src/site/locales/en.json";
import russian from "../src/site/locales/ru.json";
import { photoDirectory as checkedPhotoDirectory } from "../tools/sitecheck/main.ts";
import {
	buildSite,
	givenTwice,
	keptPhotoOf,
	main,
	paperFiles,
	readSources,
	sharingPictures,
} from "./prerender-site.ts";

const repository = join(import.meta.dirname, "..", "..");

// font is the directory of the font package the site sets its text in.
const font = dirname(
	createRequire(import.meta.url).resolve(
		"@fontsource-variable/onest/package.json",
	),
);

// fontFiles are the subsets of the font the site serves, as the stylesheet
// the build makes names them.
const fontFiles = [
	"onest-cyrillic-wght-normal.woff2",
	"onest-latin-wght-normal.woff2",
];

// scripted are the pages that may run a script: the home pages, whose demo
// brings their cards alive.
const scripted: readonly string[] = ["en/index.html", "ru/index.html"];

// carded are the pages that draw a card of the widget, and load its styles.
const carded: readonly string[] = [
	"en/index.html",
	"en/topics/index.html",
	"en/why/index.html",
	"ru/index.html",
	"ru/topics/index.html",
	"ru/why/index.html",
];

// framed are the documents of the site a page frames rather than the pages
// themselves: the coach's prototype is an application of its own, the export
// of the tool it was drawn in, which no rule of a page holds.
const framed: readonly string[] = ["assets/coach-prototype.html"];

// adding is the word of every button that asks a reader to add MathTrail, by
// the language of its page.
const adding: Readonly<Record<string, string>> = {
	en: english["nav.add"],
	ru: russian["nav.add"],
};

// pagesIn are the paths of the pages below dir: every HTML file but a document
// a page frames.
async function pagesIn(dir: string): Promise<string[]> {
	return (await filesIn(dir)).filter(
		(file) => file.endsWith(".html") && !framed.includes(file),
	);
}

// localePages are the paths of the pages of every language below dir: all
// but the apex, which has no language of its own.
async function localePages(dir: string): Promise<string[]> {
	return (await pagesIn(dir)).filter((file) => file.includes("/"));
}

// framesIn are the addresses of the documents a page frames, in the order it
// frames them: parsed rather than matched, as the page's links are.
function framesIn(html: string): string[] {
	const frames: string[] = [];
	const walk = (node: DefaultTreeAdapterTypes.Node): void => {
		if (!("childNodes" in node)) {
			return;
		}
		if ("tagName" in node && node.tagName === "iframe") {
			frames.push(node.attrs.find((attr) => attr.name === "src")?.value ?? "");
		}
		for (const child of node.childNodes) {
			walk(child);
		}
	};
	walk(parse(html));
	return frames;
}

// Link is one anchor of a page, as a browser reads it.
type Link = { className: string; href: string; text: string };

// linksIn are the HTML anchors of a page, in the order it has them, each with
// the text it shows: parsed rather than matched, so that markup inside an
// anchor is read as markup. A template's content is not walked: it is not part
// of the page until a script puts it there.
function linksIn(html: string): Link[] {
	const links: Link[] = [];
	const walk = (node: DefaultTreeAdapterTypes.Node): void => {
		if (!("childNodes" in node)) {
			return;
		}
		if (
			"tagName" in node &&
			node.tagName === "a" &&
			node.namespaceURI === "http://www.w3.org/1999/xhtml"
		) {
			const value = (name: string) =>
				node.attrs.find((attr) => attr.name === name)?.value ?? "";
			links.push({
				className: value("class"),
				href: value("href"),
				text: textIn(node),
			});
		}
		for (const child of node.childNodes) {
			walk(child);
		}
	};
	walk(parse(html));
	return links;
}

// Image is one picture a page shows, by the attributes it is drawn with.
type Image = { src: string; width: string; height: string; alt: string };

// imagesIn are the pictures a page shows, in the order it has them: parsed
// rather than matched, as the page's links are.
function imagesIn(html: string): Image[] {
	const images: Image[] = [];
	const walk = (node: DefaultTreeAdapterTypes.Node): void => {
		if (!("childNodes" in node)) {
			return;
		}
		if ("tagName" in node && node.tagName === "img") {
			const value = (name: string) =>
				node.attrs.find((attr) => attr.name === name)?.value ?? "";
			images.push({
				src: value("src"),
				width: value("width"),
				height: value("height"),
				alt: value("alt"),
			});
		}
		for (const child of node.childNodes) {
			walk(child);
		}
	};
	walk(parse(html));
	return images;
}

// textIn is all the text below node.
function textIn(node: DefaultTreeAdapterTypes.Node): string {
	if (node.nodeName === "#text" && "value" in node) {
		return node.value;
	}
	return "childNodes" in node ? node.childNodes.map(textIn).join("") : "";
}

// urlsIn are the addresses every url() of a stylesheet names, sorted.
function urlsIn(style: string): string[] {
	return [...style.matchAll(/url\(\s*["']?([^"')]*)["']?\s*\)/g)]
		.map(([, address]) => address ?? "")
		.sort(byCodeUnits);
}

// filesIn are the paths of every file below dir, relative to it, sorted.
async function filesIn(dir: string): Promise<string[]> {
	const entries = await readdir(dir, { recursive: true, withFileTypes: true });
	return entries
		.filter((entry) => entry.isFile())
		.map((entry) => relative(dir, join(entry.parentPath, entry.name)))
		.sort(byCodeUnits);
}

// research is the file of the numbers of the page Research every build of
// these tests is given: the bench's fixture, with no file of the paper to
// ship, since none is kept beside it.
let research = "";

beforeAll(async () => {
	const dir = await mkdtemp(join(tmpdir(), "research-"));
	research = join(dir, "research.json");
	await writeFile(
		research,
		JSON.stringify({ ...fixture, paper: { ...fixture.paper, files: [] } }),
	);
});

afterAll(async () => {
	await rm(dirname(research), { recursive: true, force: true });
});

// command runs the build's command line, and is what it said and the code it
// ended with.
async function command(args: string[]) {
	const said: string[] = [];
	const code = await main(
		args,
		(line) => said.push(line),
		(line) => said.push(line),
	);
	return { code, said };
}

describe("the site built from this repository", () => {
	let out = "";

	beforeAll(async () => {
		out = await mkdtemp(join(tmpdir(), "site-"));
		// An older build, with a page this one has no text for.
		await writeFile(join(out, ".nojekyll"), "");
		await mkdir(join(out, "en", "gone"), { recursive: true });
		await writeFile(join(out, "en", "gone", "index.html"), "old");
		expect(
			await command([
				"--base",
				"https://mathtrail.app",
				"--out",
				out,
				"--research",
				research,
			]),
		).toEqual({ code: 0, said: [`site: built into ${out}`] });
	}, 60_000);

	afterAll(async () => {
		await rm(out, { recursive: true, force: true });
	});

	test("is its pages, its stylesheets, its font, its mark, its sharing pictures, the family's photographs, the coach's prototype and the host's files, and nothing older", async () => {
		expect(await filesIn(out)).toEqual([
			".nojekyll",
			"CNAME",
			"assets/card.css",
			"assets/coach-prototype.html",
			"assets/demo.js",
			"assets/favicon.svg",
			"assets/noto-license.txt",
			"assets/og-en.png",
			"assets/og-ru.png",
			"assets/onest-cyrillic-wght-normal.woff2",
			"assets/onest-latin-wght-normal.woff2",
			"assets/onest-license.txt",
			"assets/photos/dad.webp",
			"assets/photos/family.webp",
			"assets/photos/mum.webp",
			"assets/photos/older-son.webp",
			"assets/photos/younger-son.webp",
			"assets/style.css",
			"assets/tokens.css",
			"en/about/index.html",
			"en/coach/index.html",
			"en/index.html",
			"en/privacy/index.html",
			"en/research/index.html",
			"en/techniques/index.html",
			"en/terms/index.html",
			"en/topics/arithmetic-with-a-trick/index.html",
			"en/topics/calendar-and-age/index.html",
			"en/topics/clocks/index.html",
			"en/topics/divisibility-and-remainders/index.html",
			"en/topics/enumeration/index.html",
			"en/topics/figures-on-a-grid/index.html",
			"en/topics/gaps-and-boundaries/index.html",
			"en/topics/index.html",
			"en/topics/knights-and-liars/index.html",
			"en/topics/ordering/index.html",
			"en/topics/overlapping-groups/index.html",
			"en/topics/parity-and-alternation/index.html",
			"en/topics/parts-and-shares/index.html",
			"en/topics/percentages/index.html",
			"en/topics/pigeonhole-principle/index.html",
			"en/topics/ratios-and-sharing/index.html",
			"en/topics/weighing-and-pouring/index.html",
			"en/topics/winning-strategy/index.html",
			"en/why/index.html",
			"index.html",
			"robots.txt",
			"ru/about/index.html",
			"ru/coach/index.html",
			"ru/index.html",
			"ru/privacy/index.html",
			"ru/research/index.html",
			"ru/techniques/index.html",
			"ru/terms/index.html",
			"ru/topics/arithmetic-with-a-trick/index.html",
			"ru/topics/calendar-and-age/index.html",
			"ru/topics/clocks/index.html",
			"ru/topics/divisibility-and-remainders/index.html",
			"ru/topics/enumeration/index.html",
			"ru/topics/figures-on-a-grid/index.html",
			"ru/topics/gaps-and-boundaries/index.html",
			"ru/topics/index.html",
			"ru/topics/knights-and-liars/index.html",
			"ru/topics/ordering/index.html",
			"ru/topics/overlapping-groups/index.html",
			"ru/topics/parity-and-alternation/index.html",
			"ru/topics/parts-and-shares/index.html",
			"ru/topics/percentages/index.html",
			"ru/topics/pigeonhole-principle/index.html",
			"ru/topics/ratios-and-sharing/index.html",
			"ru/topics/weighing-and-pouring/index.html",
			"ru/topics/winning-strategy/index.html",
			"ru/why/index.html",
			"sitemap.xml",
		]);
	});

	test("ships the design's tokens and the mark as the very files they are", async () => {
		expect(await readFile(join(out, "assets", "tokens.css"))).toEqual(
			await readFile(join(repository, "internal", "widget", "tokens.css")),
		);
		expect(await readFile(join(out, "assets", "favicon.svg"))).toEqual(
			await readFile(join(repository, "site", "assets", "favicon.svg")),
		);
	});

	test("ships the font's subsets as its package has them, and its licence beside them", async () => {
		for (const file of fontFiles) {
			expect(await readFile(join(out, "assets", file))).toEqual(
				await readFile(join(font, "files", file)),
			);
		}
		expect(await readFile(join(out, "assets", "onest-license.txt"))).toEqual(
			await readFile(join(font, "LICENSE")),
		);
	});

	test("ships the coach's prototype as the very file the site keeps, and the licence of its fonts beside it", async () => {
		for (const file of [...framed, "assets/noto-license.txt"]) {
			const shipped = await readFile(join(out, file));
			const kept = await readFile(join(repository, "site", file));
			expect(shipped.equals(kept), file).toBe(true);
		}
	});

	// The prototype is a design tool's export, and a new export can bring back
	// what was taken out of it, which only a reader sees. It is held to the hash
	// it had when it was last read through, written here by hand, so that no
	// export ships unread.
	test("keeps the coach's prototype the one that was last read through", async () => {
		const kept = await readFile(
			join(repository, "site", "assets", "coach-prototype.html"),
		);

		expect(createHash("sha256").update(kept).digest("hex")).toBe(
			"cd07c9658db2f82fb6103c47e327b9cf27ef3f87a5904d46b3f6278a4bd0666d",
		);
	});

	test("serves no HTML but its pages and what they frame, and frames the coach's prototype on the coach's page alone", async () => {
		const documents = (await filesIn(out)).filter((file) =>
			file.endsWith(".html"),
		);
		expect(
			documents.filter(
				(file) => file !== "index.html" && !file.endsWith("/index.html"),
			),
		).toEqual(framed);

		const framing: Record<string, string[]> = {};
		for (const page of await pagesIn(out)) {
			const frames = framesIn(await readFile(join(out, page), "utf8"));
			if (frames.length > 0) {
				framing[page] = frames;
			}
		}
		expect(framing).toEqual({
			"en/coach/index.html": ["/assets/coach-prototype.html"],
			"ru/coach/index.html": ["/assets/coach-prototype.html"],
		});
	});

	// The prototype unpacks itself in the browser from data it carries, much of
	// it compressed, so the checker of the site, which reads markup alone, sees
	// none of the addresses inside it. The site promises to load nothing from
	// another origin, so every address the prototype names, in its page and in
	// every text it carries, must be one it never loads from: a namespace of
	// XML, the page a message of React's errors points to, or the address of a
	// copy it carries, which its runtime asks for by that address and is given
	// the copy. Its runtime fetches more only for an import: Babel from another
	// origin to compile a component imported by address, and a file of the
	// prototype's own for a component imported by name. It imports neither.
	test("holds the coach's prototype to the site's own origin: every address it names is one it never loads from", async () => {
		const prototype = await readFile(
			join(repository, "site", "assets", "coach-prototype.html"),
			"utf8",
		);
		const block = (type: string): unknown => {
			const open = `<script type="__bundler/${type}">`;
			const start = prototype.indexOf(open) + open.length;
			return JSON.parse(
				prototype.slice(start, prototype.indexOf("</script>", start)),
			);
		};
		const manifest = block("manifest") as Record<
			string,
			{ mime: string; compressed?: boolean; data: string }
		>;
		const copies = block("ext_resources") as { id: string; uuid: string }[];
		const template = block("template") as string;
		const carried = Object.values(manifest)
			.filter(({ mime }) => /^(?:text\/|.*(?:javascript|json|xml))/.test(mime))
			.map(({ data, compressed }) => {
				const bytes = Buffer.from(data, "base64");
				return (compressed ? gunzipSync(bytes) : bytes).toString("utf8");
			});
		const babel = "https://unpkg.com/@babel/standalone@7.29.0/babel.min.js";
		const neverLoaded = (address: string) =>
			address.startsWith("http://www.w3.org/") ||
			address.startsWith("https://reactjs.org/docs/error-decoder.html") ||
			address === babel ||
			copies.some(({ id }) => id === address);

		expect(carried).not.toEqual([]);
		expect(copies).not.toEqual([]);
		expect(copies.filter(({ uuid }) => !Object.hasOwn(manifest, uuid))).toEqual(
			[],
		);
		for (const text of [prototype, ...carried]) {
			const named = [
				...text.matchAll(/[a-z][a-z0-9+.-]*:\/\/[a-z0-9][^\s"'`\\<>)]*/gi),
			].map(([address]) => address);
			expect(named.filter((address) => !neverLoaded(address))).toEqual([]);
			expect(text).not.toMatch(/(?:src|href)=\\?["']\/\//);
			expect(text).not.toContain("url(//");
		}
		expect(template).not.toMatch(/<(?:x|dc)-import\b/i);
	});

	test("builds the site's own stylesheet, with no copy of the tokens inside it", async () => {
		const style = await readFile(join(out, "assets", "style.css"), "utf8");

		expect(style).toContain("--s-page:");
		expect(style).not.toContain("--surface:");
		expect(style).not.toMatch(/@import|https?:|data:/);
	});

	test("builds a stylesheet that loads the site's own font files and nothing else", async () => {
		const style = await readFile(join(out, "assets", "style.css"), "utf8");

		expect(urlsIn(style)).toEqual(fontFiles.map((file) => `/assets/${file}`));
	});

	test("loads on every page of a language the tokens, then the styles, and a script only on a page allowed one", async () => {
		const pages = await localePages(out);
		expect(pages).not.toEqual([]);
		for (const page of pages) {
			const html = await readFile(join(out, page), "utf8");

			expect(
				[...html.matchAll(/<link rel="stylesheet" href="([^"]+)"/g)].map(
					([, href]) => href,
				),
			).toEqual([
				"/assets/tokens.css",
				"/assets/style.css",
				...(carded.includes(page) ? ["/assets/card.css"] : []),
			]);
			expect(html.includes("<script")).toBe(scripted.includes(page));
			expect(
				[...html.matchAll(/<script type="module" src="([^"]+)"/g)].map(
					([, src]) => src,
				),
			).toEqual(scripted.includes(page) ? ["/assets/demo.js"] : []);
			expect(html.includes('class="mt mt-widget')).toBe(carded.includes(page));
		}
	});

	test("sends a reader of the bare domain on to the English front page, loading nothing first", async () => {
		const apex = await readFile(join(out, "index.html"), "utf8");

		expect(apex).toContain(
			'<meta http-equiv="refresh" content="0; url=/en/"/>',
		);
		expect(apex).toContain(
			'<link rel="canonical" href="https://mathtrail.app/en/"/>',
		);
		expect(apex).not.toMatch(/rel="stylesheet"|<script|<style/);
	});

	test("carries on the home page what its demo needs: the lesson's task, the widget's words in the page's language alone, and an answer for every option", async () => {
		for (const [locale, words] of [
			["en", widgetEnglish],
			["ru", widgetRussian],
		] as const) {
			const html = await readFile(join(out, locale, "index.html"), "utf8");
			const [, carried] =
				/<script type="application\/json" data-demo>([^<]*)<\/script>/.exec(
					html,
				) ?? [];
			const demo = JSON.parse(carried ?? "null");

			expect(demo.locale).toBe(locale);
			expect(demo.words).toEqual(words);
			expect(demo.handed.task.id).toBe("site_home");
			expect(Object.keys(demo.results)).toEqual(["A", "B", "C", "D", "E"]);
		}
	});

	// A script is weighed whole, as one file: the list of the files the build
	// makes holds that, and the site's checker refuses a script that imports.
	// The demo carries none of the widget's dictionaries, which the page
	// carries in its own language, and none of the libraries a card in a chat
	// reads the service and the host with, since the page answers for both.
	// The topics' names it does carry, in the catalog the card's choice of a
	// topic reads its grades from.
	test("builds the demo into one file, with no dictionary of the widget, no zod and nothing of a host's library", async () => {
		const demo = await readFile(join(out, "assets", "demo.js"), "utf8");
		const named = new Set(catalog.map((topic) => topic.name));
		const long = [widgetEnglish, widgetRussian].flatMap((words) =>
			Object.values(words).filter(
				(said): said is string =>
					typeof said === "string" && said.length > 24 && !named.has(said),
			),
		);

		expect(demo).not.toContain("_zod");
		expect(demo).not.toContain("ui/initialize");
		expect(long.filter((said) => demo.includes(said))).toEqual([]);
	});

	test("names on every page the sharing picture of its language, one the build serves at that address", async () => {
		const pages = await pagesIn(out);
		for (const page of pages) {
			const html = await readFile(join(out, page), "utf8");
			const locale = page.includes("/")
				? page.slice(0, page.indexOf("/"))
				: "en";
			const named = [
				...html.matchAll(/<meta property="og:image" content="([^"]+)"/g),
			].map(([, address]) => address ?? "");

			expect(named, page).toEqual([
				`https://mathtrail.app/assets/og-${locale}.png`,
			]);
			const served = await readFile(
				join(out, new URL(named[0] ?? "").pathname),
			);
			const kept = await readFile(
				join(repository, "site", "assets", `og-${locale}.png`),
			);
			expect(served.equals(kept), page).toBe(true);
			expect(html, page).toContain(
				'<meta name="twitter:card" content="summary_large_image"/>',
			);
			expect(
				[
					...html.matchAll(/<meta property="og:image:alt" content="([^"]+)"/g),
				].map(([, alt]) => alt),
				page,
			).toEqual([
				locale === "ru" ? russian["share.picture"] : english["share.picture"],
			]);
		}
	});

	test("builds the cards' stylesheet from the widget's own, loading nothing", async () => {
		const card = await readFile(join(out, "assets", "card.css"), "utf8");

		expect(card).toContain(".mt-widget");
		expect(card).not.toMatch(/url\(|@import|https?:|data:/);
	});

	test("shows on the page of the topics every topic of the catalog, and every link between them", async () => {
		for (const locale of ["en", "ru"]) {
			const html = await readFile(
				join(out, locale, "topics", "index.html"),
				"utf8",
			);
			const cards = [
				...html.matchAll(/<article id="([^"]+)" class="s-topic"/g),
			];
			const lines = [
				...html.matchAll(/<g id="line-[^"]+" class="s-map-line"/g),
			];

			expect(cards.map(([, slug]) => slug ?? "").sort(byCodeUnits)).toEqual(
				catalog.map((topic) => topic.slug).sort(byCodeUnits),
			);
			expect(lines).toHaveLength(
				catalog.reduce((sum, topic) => sum + topic.builds_on.length, 0),
			);
		}
	});

	test("draws the page of every topic the catalog publishes, which the page of the topics leads to", async () => {
		const published = catalog.filter((topic) => topic.site_page);
		expect(published).not.toEqual([]);
		for (const locale of ["en", "ru"]) {
			const topics = await readFile(
				join(out, locale, "topics", "index.html"),
				"utf8",
			);
			for (const { slug } of published) {
				const html = await readFile(
					join(out, locale, "topics", slug, "index.html"),
					"utf8",
				);

				expect(topics).toContain(`href="/${locale}/topics/${slug}/"`);
				expect(html).toContain(`<section id="traps"`);
				expect(html).toContain(`<section id="home"`);
			}
		}
	});

	test("gives on the privacy policy's page one address to write to, the policy's and the footer's", async () => {
		for (const locale of ["en", "ru"]) {
			const html = await readFile(
				join(out, locale, "privacy", "index.html"),
				"utf8",
			);
			const written = [...html.matchAll(/href="mailto:([^"]+)"/g)].map(
				([, address]) => address,
			);

			expect(written.length).toBeGreaterThan(1);
			expect(new Set(written).size).toBe(1);
		}
	});

	test("closes the menu with the research and the page about who makes it, and names both in the footer before the documents, under the menu's words, in every language", async () => {
		const locales = (await readdir(out, { withFileTypes: true }))
			.filter((entry) => entry.isDirectory() && entry.name !== "assets")
			.map((entry) => entry.name);
		expect(locales).toEqual(expect.arrayContaining(["en", "ru"]));
		for (const locale of locales) {
			const html = await readFile(join(out, locale, "index.html"), "utf8");
			const links = (list: string) => {
				const nav = html.match(
					new RegExp(`<nav class="${list}"[^>]*>(.*?)</nav>`, "s"),
				);
				if (nav === null) {
					throw new Error(`${locale}/index.html has no ${list}`);
				}
				return [
					...(nav[1] ?? "").matchAll(/<a href="([^"]+)"[^>]*>([^<]*)<\/a>/g),
				].map(([, href, label]) => ({ href, label }));
			};
			const menu = links("s-navlinks");
			const footer = links("s-footlinks");

			expect(menu.slice(-2).map(({ href }) => href)).toEqual([
				`/${locale}/research/`,
				`/${locale}/about/`,
			]);
			expect(footer.slice(0, 4).map(({ href }) => href)).toEqual([
				`/${locale}/research/`,
				`/${locale}/about/`,
				`/${locale}/privacy/`,
				`/${locale}/terms/`,
			]);
			expect(footer.slice(0, 2).map(({ label }) => label)).toEqual(
				menu.slice(-2).map(({ label }) => label),
			);
		}
	});

	test("opens the menu with the coach's page and names no section of the home page, on every page of every language", async () => {
		const pages = await localePages(out);
		expect(pages).not.toEqual([]);
		for (const page of pages) {
			const locale = page.slice(0, page.indexOf("/"));
			const html = await readFile(join(out, page), "utf8");
			const menu =
				html.match(/<nav class="s-navlinks"[^>]*>(.*?)<\/nav>/s)?.[1] ?? "";
			const links = [...menu.matchAll(/<a href="([^"]+)"/g)].map(
				([, href]) => href,
			);

			expect(links[0], page).toBe(`/${locale}/coach/`);
			expect(
				links.filter((href) => href?.includes("#")),
				page,
			).toEqual([]);
		}
	});

	test("leads every button that asks to add it, the header's first, to the home page's section on connecting, which the home page has", async () => {
		const pages = await localePages(out);
		for (const page of pages) {
			const locale = page.slice(0, page.indexOf("/"));
			const html = await readFile(join(out, page), "utf8");
			const word = adding[locale] ?? "";
			const buttons = linksIn(html).filter((link) => link.text.trim() === word);

			expect(buttons[0]?.className, page).toContain("s-nav-action");
			expect(
				buttons.map((link) => link.href),
				page,
			).toEqual(buttons.map(() => `/${locale}/#connect`));
		}
		for (const locale of Object.keys(adding)) {
			const home = await readFile(join(out, locale, "index.html"), "utf8");

			expect(home).toContain(`<section id="lesson"`);
			expect(home).toContain(`<section id="connect"`);
		}
	});

	test("writes, on the page about who makes it, to the address the privacy policy names", async () => {
		for (const locale of ["en", "ru"]) {
			const addresses = async (page: string) =>
				[
					...(
						await readFile(join(out, locale, page, "index.html"), "utf8")
					).matchAll(/href="mailto:([^"]+)"/g),
				].map(([, address]) => address);
			const policy = await addresses("privacy");
			const about = await addresses("about");

			expect(about.length).toBeGreaterThan(1);
			expect(new Set([...policy, ...about]).size).toBe(1);
		}
	});

	test("ships the family's photographs as the very files the site keeps", async () => {
		for (const photo of photos) {
			const shipped = await readFile(join(out, photoPath(photo.name)));
			const kept = await readFile(keptPhotoOf(photo));
			expect(shipped.equals(kept), photo.name).toBe(true);
		}
	});

	test("shows on the page about who makes it, in every language, each of the family's photographs at its size, saying what it shows", async () => {
		for (const locale of ["en", "ru"]) {
			const shown = imagesIn(
				await readFile(join(out, locale, "about", "index.html"), "utf8"),
			).filter(({ src }) => src.startsWith(photoDirectory));

			expect(
				shown.map(({ src, width, height }) => [src, width, height]),
				locale,
			).toEqual(
				photos.map(({ name, width, height }) => [
					photoPath(name),
					String(width),
					String(height),
				]),
			);
			for (const { src, alt } of shown) {
				expect(alt.trim(), `${locale} ${src}`).not.toBe("");
			}
			expect(new Set(shown.map(({ alt }) => alt)).size, locale).toBe(
				shown.length,
			);
		}
	});

	// The checker of the site weighs a page without the photographs it shows,
	// and knows where they are only by its own word: were the site to keep
	// them elsewhere, every page that shows one would weigh them after all.
	test("keeps its photographs where its checker weighs them apart from the page", () => {
		expect(checkedPhotoDirectory).toBe(photoDirectory);
	});

	test("shows on the page of the techniques every technique of the data, once among the links at its top and once as its card", async () => {
		const ids = (data.techniques?.groups ?? []).flatMap((group) =>
			group.techniques.map((technique) => technique.id),
		);
		expect(ids).not.toEqual([]);
		for (const locale of ["en", "ru"]) {
			const html = await readFile(
				join(out, locale, "techniques", "index.html"),
				"utf8",
			);
			const top = html.slice(
				html.indexOf('class="s-panel s-overview"'),
				html.indexOf("</nav>", html.indexOf('class="s-panel s-overview"')),
			);

			expect([...top.matchAll(/href="#([^"]+)"/g)].map(([, id]) => id)).toEqual(
				ids,
			);
			expect(
				[...html.matchAll(/<article id="([^"]+)" class="s-technique"/g)].map(
					([, id]) => id,
				),
			).toEqual(ids);
		}
	});

	test("leads every link within a page to a part the page has, and gives no two parts of a page one name", async () => {
		const pages = await pagesIn(out);
		expect(pages).not.toEqual([]);
		for (const page of pages) {
			const html = await readFile(join(out, page), "utf8");
			const ids = [...html.matchAll(/\sid="([^"]+)"/g)].map(([, id]) => id);
			const within = [...html.matchAll(/href="#([^"]+)"/g)].map(([, id]) => id);

			expect(new Set(ids).size, page).toBe(ids.length);
			expect(
				within.filter((id) => !ids.includes(id)),
				page,
			).toEqual([]);
		}
	});

	test("is not touched by a build that cannot be made", async () => {
		const before = await filesIn(out);

		expect(
			await command([
				"--base",
				"https://mathtrail.app/",
				"--out",
				out,
				"--research",
				research,
			]),
		).toEqual({
			code: 1,
			said: ['site: the base URL "https://mathtrail.app/" ends in a slash'],
		});
		expect(await filesIn(out)).toEqual(before);
	});

	test("is left where it was by a build whose text cannot be read", async () => {
		const before = await filesIn(out);
		const broken = await mkdtemp(join(tmpdir(), "texts-"));
		try {
			await mkdir(join(broken, "en"));
			await writeFile(join(broken, "en", "index.md"), "# no front matter\n");

			await expect(
				buildSite({
					base: "https://mathtrail.app",
					out,
					content: broken,
					research,
				}),
			).rejects.toThrow("en/index.md: front matter must open");
			expect(await filesIn(out)).toEqual(before);
		} finally {
			await rm(broken, { recursive: true, force: true });
		}
	});
});

describe("a directory that holds something else", () => {
	test("is not replaced by a build of the site", async () => {
		const elsewhere = await mkdtemp(join(tmpdir(), "elsewhere-"));
		try {
			await writeFile(join(elsewhere, "notes.txt"), "keep");

			const { code, said } = await command([
				"--base",
				"https://mathtrail.app",
				"--out",
				elsewhere,
				"--research",
				research,
			]);

			expect(code).toBe(1);
			expect(said).toEqual([
				expect.stringContaining(
					"holds something other than a build of the site",
				),
			]);
			expect(await filesIn(elsewhere)).toEqual(["notes.txt"]);
		} finally {
			await rm(elsewhere, { recursive: true, force: true });
		}
	}, 60_000);

	// Another static site carries .nojekyll too; what tells it from a build of
	// this one is everything at its top that this build never writes.
	test("is not replaced when it is another site's, kept with its history", async () => {
		const other = await mkdtemp(join(tmpdir(), "other-site-"));
		try {
			await writeFile(join(other, ".nojekyll"), "");
			await writeFile(join(other, "index.html"), "theirs");
			await mkdir(join(other, ".git"));
			await writeFile(join(other, ".git", "HEAD"), "ref: refs/heads/main\n");

			const { code, said } = await command([
				"--base",
				"https://mathtrail.app",
				"--out",
				other,
				"--research",
				research,
			]);

			expect(code).toBe(1);
			expect(said).toEqual([
				expect.stringContaining(
					"holds something other than a build of the site",
				),
			]);
			expect(await filesIn(other)).toEqual([
				".git/HEAD",
				".nojekyll",
				"index.html",
			]);
		} finally {
			await rm(other, { recursive: true, force: true });
		}
	}, 60_000);
});

describe("a directory that does not exist yet", () => {
	test("is the build's to make, and is not made by a build that fails", async () => {
		const parent = await mkdtemp(join(tmpdir(), "parent-"));
		try {
			const fresh = join(parent, "dist");

			expect(
				await command([
					"--base",
					"https://mathtrail.app/",
					"--out",
					fresh,
					"--research",
					research,
				]),
			).toEqual({
				code: 1,
				said: ['site: the base URL "https://mathtrail.app/" ends in a slash'],
			});
			expect(await readdir(parent)).toEqual([]);
		} finally {
			await rm(parent, { recursive: true, force: true });
		}
	});

	test("is made by a build that succeeds", async () => {
		const parent = await mkdtemp(join(tmpdir(), "parent-"));
		try {
			const fresh = join(parent, "dist");

			await buildSite({ base: "https://mathtrail.app", out: fresh, research });
			expect(await readdir(fresh)).toContain(".nojekyll");
		} finally {
			await rm(parent, { recursive: true, force: true });
		}
	}, 60_000);
});

describe("the numbers of the page Research", () => {
	test("are what a build cannot do without, and the build names the recipe that makes them", async () => {
		const parent = await mkdtemp(join(tmpdir(), "parent-"));
		try {
			await expect(
				buildSite({
					base: "https://mathtrail.app",
					out: join(parent, "dist"),
					research: join(parent, "research.json"),
				}),
			).rejects.toThrow(
				"the numbers of the page Research, is not there: just research-data makes it",
			);
			expect(await readdir(parent)).toEqual([]);
		} finally {
			await rm(parent, { recursive: true, force: true });
		}
	});

	test("are refused when a run cut them short, and the build names the recipe that makes them again", async () => {
		const parent = await mkdtemp(join(tmpdir(), "parent-"));
		try {
			const cut = join(parent, "research.json");
			await writeFile(cut, JSON.stringify(fixture).slice(0, 100));

			await expect(
				buildSite({
					base: "https://mathtrail.app",
					out: join(parent, "dist"),
					research: cut,
				}),
			).rejects.toThrow(
				`${cut}, the numbers of the page Research, is not whole: just research-data makes it again`,
			);
		} finally {
			await rm(parent, { recursive: true, force: true });
		}
	});
});

describe("the paper the page Research offers", () => {
	let dir = "";
	const pdf = Buffer.from("%PDF-1.7 a paper\n");
	const facts = {
		lang: "en",
		path: "/assets/paper-a.en.pdf",
		pages: 1,
		bytes: pdf.length,
		sha256: createHash("sha256").update(pdf).digest("hex"),
	};

	beforeAll(async () => {
		dir = await mkdtemp(join(tmpdir(), "paper-"));
		await writeFile(join(dir, "paper-a.en.pdf"), pdf);
	});

	afterAll(async () => {
		await rm(dir, { recursive: true, force: true });
	});

	test("is shipped as the very file kept beside the numbers, where they say the site serves it", async () => {
		expect(await paperFiles(dir, [facts])).toEqual([
			{ path: "assets/paper-a.en.pdf", data: pdf },
		]);
	});

	test("is not shipped when the numbers name no file of it", async () => {
		expect(await paperFiles(dir, [])).toEqual([]);
	});

	test("is copied into the site a build makes, at the address its page links", async () => {
		const parent = await mkdtemp(join(tmpdir(), "parent-"));
		try {
			const numbers = join(dir, "research.json");
			await writeFile(
				numbers,
				JSON.stringify({
					...fixture,
					paper: { ...fixture.paper, files: [facts] },
				}),
			);
			const out = join(parent, "dist");

			await buildSite({
				base: "https://mathtrail.app",
				out,
				research: numbers,
			});
			expect(await readFile(join(out, facts.path))).toEqual(pdf);
			expect(
				await readFile(join(out, "en", "research", "index.html"), "utf8"),
			).toContain(`href="${facts.path}"`);
		} finally {
			await rm(parent, { recursive: true, force: true });
		}
	}, 60_000);

	test.each([
		[
			"is not beside the numbers",
			{ path: "/assets/paper-b.en.pdf" },
			"paper-b.en.pdf, which the page Research offers, is not beside its numbers",
		],
		[
			"is of another size",
			{ bytes: pdf.length + 1 },
			"is not the file they vouch for",
		],
		[
			"is of another hash",
			{ sha256: "0".repeat(64) },
			"is not the file they vouch for",
		],
	])("is refused when the file %s", async (_, changed, said) => {
		await expect(paperFiles(dir, [{ ...facts, ...changed }])).rejects.toThrow(
			said,
		);
	});
});

describe("the pictures a shared link shows", () => {
	test("are those of the languages that have one, and a language with none yet still builds", async () => {
		const pictures = await sharingPictures(["en", "xx"]);

		expect(pictures.map(({ path }) => path)).toEqual(["assets/og-en.png"]);
	});
});

describe("the files of a build", () => {
	test("are each written to a path of their own", () => {
		expect(
			givenTwice([{ path: "index.html" }, { path: "assets/style.css" }]),
		).toBeUndefined();
	});

	test("name the path two of them would both be written to", () => {
		expect(
			givenTwice([
				{ path: "assets/favicon.svg" },
				{ path: "index.html" },
				{ path: "assets/favicon.svg" },
			]),
		).toBe("assets/favicon.svg");
	});
});

describe("the build's command line", () => {
	test.each([
		["no origin", ["--out", "dist"]],
		["no directory", ["--base", "https://mathtrail.app"]],
		["an option the build does not know", ["--bse", "https://mathtrail.app"]],
	])("is refused, with its usage, when it names %s", async (_, args) => {
		expect(await command(args)).toEqual({
			code: 2,
			said: [expect.stringMatching(/^usage: /)],
		});
	});
});

// The stylesheet is read here, as it is written, rather than imported into a
// test that runs in a browser's stead: there a stylesheet arrives empty, and a
// rule about what it must not say would hold for nothing.
describe("the site's stylesheet, as it is written", () => {
	let style = "";

	beforeAll(async () => {
		style = await readFile(
			join(import.meta.dirname, "..", "src", "site", "style.css"),
			"utf8",
		);
	});

	test("is the site's own", () => {
		expect(style).toContain(".s-doc");
	});

	test.each([
		[
			"a side named left or right",
			/\b(margin|padding|border|inset)-(left|right)\b/,
		],
		["a position from the left or the right", /(^|[\s;{])(left|right)\s*:/m],
		["text set to one side", /text-align\s*:\s*(left|right)/],
		["a float", /\bfloat\s*:/],
	])("lays a page out the same way in either direction: no %s", (_, rule) => {
		expect(style).not.toMatch(rule);
	});

	test.each([
		["another stylesheet imported", /@import/],
		["an address of another origin", /https?:/],
	])("loads nothing from anywhere: no %s", (_, rule) => {
		expect(style).not.toMatch(rule);
	});

	test("loads by url() the font's subsets alone, from the font's package", () => {
		expect(urlsIn(style)).toEqual(
			fontFiles.map((file) => `@fontsource-variable/onest/files/${file}`),
		);
	});

	test("declares each subset of the font with the characters its package gives it", async () => {
		const ranges = JSON.parse(
			await readFile(join(font, "unicode.json"), "utf8"),
		) as Record<string, string>;
		const declared = [...style.matchAll(/@font-face\s*{([^}]*)}/g)].map(
			([, rules = ""]) => ({
				subset: /onest-([a-z-]+)-wght-normal/.exec(rules)?.[1],
				range: /unicode-range:([^;]*);/.exec(rules)?.[1]?.replace(/\s+/g, ""),
			}),
		);

		expect(declared).toEqual([
			{ subset: "latin", range: ranges.latin },
			{ subset: "cyrillic", range: ranges.cyrillic },
		]);
	});

	test("sets its text in its own font, the system's behind it", () => {
		expect(style).toMatch(/--s-font:\s*"Onest",\s*var\(--font-sans\);/);
		expect(style).toMatch(/html\s*{[^}]*font-family:\s*var\(--s-font\);/);
	});

	test("has one look, whatever the reader's system prefers", () => {
		expect(style).not.toContain("prefers-color-scheme");
	});

	test("takes the design's tokens from their own file, not from a copy", () => {
		expect(style).not.toMatch(
			/--(surface|text|accent|page|space-\d+|font-sans)\s*:/,
		);
	});
});

describe("the site's texts", () => {
	let dir = "";

	beforeAll(async () => {
		dir = await mkdtemp(join(tmpdir(), "texts-"));
		await mkdir(join(dir, "en", "topics"), { recursive: true });
		await writeFile(join(dir, "en", "index.md"), "home");
		await writeFile(join(dir, "en", "why.yaml"), "title: Why");
		await writeFile(join(dir, "en", "topics", "sample.yaml"), "title: Sample");
		await writeFile(join(dir, "en", "notes.txt"), "not a text");
		await mkdir(join(dir, "en", "drafts"));
		await writeFile(join(dir, "README.md"), "not a locale");
	});

	afterAll(async () => {
		await rm(dir, { recursive: true, force: true });
	});

	test("are the .md and .yaml files below each locale's directory, by their path", async () => {
		const read = await readSources(dir);

		expect([...read.keys()]).toEqual(["en"]);
		expect(
			[...(read.get("en") ?? [])].sort(([a], [b]) => byCodeUnits(a, b)),
		).toEqual([
			["index.md", "home"],
			["topics/sample.yaml", "title: Sample"],
			["why.yaml", "title: Why"],
		]);
	});
});
