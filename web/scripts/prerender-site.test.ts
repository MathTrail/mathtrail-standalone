// @vitest-environment node
import {
	mkdir,
	mkdtemp,
	readdir,
	readFile,
	rm,
	writeFile,
} from "node:fs/promises";
import { tmpdir } from "node:os";
import { join, relative } from "node:path";
import { afterAll, beforeAll, describe, expect, test } from "vitest";
import { byCodeUnits } from "../src/site/content.ts";
import { buildSite, givenTwice, main, readSources } from "./prerender-site.ts";

const repository = join(import.meta.dirname, "..", "..");

// published are the addresses the site has handed out: the consent screen of
// the sign-in points at two of them, and a listing points at the apex. A build
// that dropped one would break a link somebody else holds, and no check of
// the built site can know what was handed out before it.
const published = [
	"/",
	"/en/",
	"/en/privacy/",
	"/en/terms/",
	"/ru/",
	"/ru/privacy/",
	"/ru/terms/",
];

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

	test("keeps every address the site has published", async () => {
		for (const address of published) {
			await expect(
				readFile(join(out, address, "index.html"), "utf8"),
			).resolves.toMatch(/^<!DOCTYPE html>/);
		}
	});

	test("is its pages, its stylesheets, its mark and the host's files, and nothing older", async () => {
		expect(await filesIn(out)).toEqual([
			".nojekyll",
			"CNAME",
			"assets/favicon.svg",
			"assets/style.css",
			"assets/tokens.css",
			"en/index.html",
			"en/privacy/index.html",
			"en/terms/index.html",
			"index.html",
			"robots.txt",
			"ru/index.html",
			"ru/privacy/index.html",
			"ru/terms/index.html",
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

	test("builds the site's own stylesheet, with no copy of the tokens inside it", async () => {
		const style = await readFile(join(out, "assets", "style.css"), "utf8");

		expect(style).toContain("--s-page:");
		expect(style).not.toContain("--surface:");
		expect(style).not.toMatch(/url\(|@import|https?:|data:/);
	});

	test("loads on every page the tokens, then the styles, and no script", async () => {
		for (const address of published) {
			const html = await readFile(join(out, address, "index.html"), "utf8");

			expect(
				[...html.matchAll(/<link rel="stylesheet" href="([^"]+)"/g)].map(
					([, href]) => href,
				),
			).toEqual(["/assets/tokens.css", "/assets/style.css"]);
			expect(html).not.toContain("<script");
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
	});
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
		["a resource loaded by url()", /url\(/],
		["another stylesheet imported", /@import/],
		["an address of another origin", /https?:/],
	])("loads nothing from anywhere: no %s", (_, rule) => {
		expect(style).not.toMatch(rule);
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
		await mkdir(join(dir, "en"));
		await writeFile(join(dir, "en", "index.md"), "home");
		await writeFile(join(dir, "en", "notes.txt"), "not a page");
		await mkdir(join(dir, "en", "drafts"));
		await writeFile(join(dir, "README.md"), "not a locale");
	});

	afterAll(async () => {
		await rm(dir, { recursive: true, force: true });
	});

	test("are the .md files of each locale's directory, by page name", async () => {
		expect(await readSources(dir)).toEqual(
			new Map([["en", new Map([["index", "home"]])]]),
		);
	});
});
