// @vitest-environment node
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
import { afterAll, beforeAll, describe, expect, test } from "vitest";
import catalog from "../../content/catalogs/topics.json";
import { byCodeUnits } from "../src/i18n/order.ts";
import { buildSite, givenTwice, main, readSources } from "./prerender-site.ts";

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

// scripted are the pages that may run a script: none, until the home page's
// demo comes.
const scripted: readonly string[] = [];

// carded are the pages that draw a card of the widget, and load its styles.
const carded: readonly string[] = [
	"en/topics/index.html",
	"ru/topics/index.html",
];

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
			await command(["--base", "https://mathtrail.app", "--out", out]),
		).toEqual({ code: 0, said: [`site: built into ${out}`] });
	}, 60_000);

	afterAll(async () => {
		await rm(out, { recursive: true, force: true });
	});

	test("is its pages, its stylesheets, its font, its mark and the host's files, and nothing older", async () => {
		expect(await filesIn(out)).toEqual([
			".nojekyll",
			"CNAME",
			"assets/card.css",
			"assets/favicon.svg",
			"assets/onest-cyrillic-wght-normal.woff2",
			"assets/onest-latin-wght-normal.woff2",
			"assets/onest-license.txt",
			"assets/style.css",
			"assets/tokens.css",
			"en/index.html",
			"en/privacy/index.html",
			"en/terms/index.html",
			"en/topics/index.html",
			"index.html",
			"robots.txt",
			"ru/index.html",
			"ru/privacy/index.html",
			"ru/terms/index.html",
			"ru/topics/index.html",
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

	test("loads on every page the tokens, then the styles, and a script only on a page allowed one", async () => {
		const pages = (await filesIn(out)).filter((file) => file.endsWith(".html"));
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
			expect(html.includes('class="mt mt-widget')).toBe(carded.includes(page));
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

	test("is not touched by a build that cannot be made", async () => {
		const before = await filesIn(out);

		expect(
			await command(["--base", "https://mathtrail.app/", "--out", out]),
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
				buildSite({ base: "https://mathtrail.app", out, content: broken }),
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
				await command(["--base", "https://mathtrail.app/", "--out", fresh]),
			).toEqual({
				code: 1,
				said: ['site: the base URL "https://mathtrail.app/" ends in a slash'],
			});
			expect(await readdir(parent)).toEqual([]);
		} finally {
			await rm(parent, { recursive: true, force: true });
		}
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
