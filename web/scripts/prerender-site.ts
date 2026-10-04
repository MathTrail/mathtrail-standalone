// The site's build. It reads the texts in site/content/<locale>/ — the
// documents' Markdown and the pages' words — draws every page with the site's
// components, builds the stylesheet with the font it sets, and writes the whole
// site into the one directory a static host publishes.
//
//	node scripts/prerender-site.ts --base https://mathtrail.app --out ../site/dist

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
import { dirname, join, relative, resolve, sep } from "node:path";
import { parseArgs } from "node:util";
import { build, createServer } from "vite";
import { sharingPicturePath } from "../src/site/brand.ts";
import type { SiteFile } from "../src/site/render.tsx";

const web = join(import.meta.dirname, "..");
const repository = join(web, "..");
const siteConfig = join(web, "vite.config.site.ts");
const require = createRequire(import.meta.url);

// textFile matches the files a locale's texts are: a document's Markdown, or
// the words of a page a component draws.
const textFile = /\.(md|yaml)$/;

/** SiteOptions say where a build of the site goes and what it is made of. */
export type SiteOptions = {
	/** base is the origin the site is published on, with no trailing slash. */
	base: string;
	/** out is the directory the site is written into, replacing what it held. */
	out: string;
	/** content is the directory of the site's texts, one directory per locale. */
	content?: string;
};

/**
 * readSources reads the site's texts below dir: every directory in it is a
 * locale, and every .md and .yaml file below a locale, however deep, is a text,
 * by its path from the locale's directory — topics/knights-and-liars.yaml.
 * Anything else is left where it is.
 */
export async function readSources(
	dir: string,
): Promise<Map<string, Map<string, string>>> {
	const locales = (await readdir(dir, { withFileTypes: true })).filter(
		(entry) => entry.isDirectory(),
	);
	return new Map(
		await Promise.all(
			locales.map(
				async (locale) =>
					[locale.name, await readLocale(join(dir, locale.name))] as const,
			),
		),
	);
}

// readLocale reads the texts of one locale: every text file below dir, by its
// path from dir, written with forward slashes on every system.
async function readLocale(dir: string): Promise<Map<string, string>> {
	const files = (
		await readdir(dir, { recursive: true, withFileTypes: true })
	).filter((file) => file.isFile() && textFile.test(file.name));
	return new Map(
		await Promise.all(
			files.map(async (file) => {
				const path = join(file.parentPath, file.name);
				return [
					relative(dir, path).split(sep).join("/"),
					await readFile(path, "utf8"),
				] as const;
			}),
		),
	);
}

/** Made is one file of a build, ready to be written below the site's root. */
type Made = { readonly path: string; readonly data: string | Buffer };

/**
 * buildSite builds the site into out: the pages its texts make, the
 * stylesheet and the font files it names, the font's licence, the design's
 * tokens, the mark and the pictures a shared link shows. Every file is made
 * before anything is written, so a build that fails at any step leaves the
 * last one where it was.
 */
export async function buildSite({
	base,
	out,
	content = join(repository, "site", "content"),
}: SiteOptions): Promise<void> {
	const sources = await readSources(content);
	const files: Made[] = [
		...(await drawPages(base, sources)),
		...(await sharingPictures([...sources.keys()])),
		...(await buildStyles()),
		// The tokens are the very file the widget and the sign-in pages read,
		// so the site takes its fonts and sizes from where they do; the mark is
		// the site's own.
		{
			path: "assets/tokens.css",
			data: await readFile(
				join(repository, "internal", "widget", "tokens.css"),
			),
		},
		{
			path: "assets/favicon.svg",
			data: await readFile(join(repository, "site", "assets", "favicon.svg")),
		},
		// The font's licence travels with its files, as the licence asks of
		// every copy of the font.
		{
			path: "assets/onest-license.txt",
			data: await readFile(
				require.resolve("@fontsource-variable/onest/LICENSE"),
			),
		},
	];
	const twice = givenTwice(files);
	if (twice !== undefined) {
		throw new Error(`${twice} would be written twice`);
	}
	if (!(await replaceable(out, files))) {
		throw new Error(
			`${out} holds something other than a build of the site, and a build replaces everything in it`,
		);
	}
	await rm(out, { recursive: true, force: true });
	await Promise.all(
		files.map(async ({ path, data }) => {
			const target = join(out, path);
			await mkdir(dirname(target), { recursive: true });
			await writeFile(target, data);
		}),
	);
}

// replaceable says whether out may be replaced by a build that writes files:
// it does not exist yet, it is empty, or it holds an earlier build of the site
// — the .nojekyll every build writes, and nothing at its top that this build
// would not write too. A build deletes everything in out, so anything else —
// the repository itself, another site's folder with its history, a directory
// named by mistake — is not the build's to delete.
async function replaceable(
	out: string,
	files: readonly Made[],
): Promise<boolean> {
	let entries: string[];
	try {
		entries = await readdir(out);
	} catch (error) {
		return error instanceof Error && "code" in error && error.code === "ENOENT";
	}
	const written = new Set(files.map(({ path }) => path.split("/")[0]));
	return (
		entries.length === 0 ||
		(entries.includes(".nojekyll") &&
			entries.every((entry) => written.has(entry)))
	);
}

/**
 * givenTwice is a path two of the files would both be written to, or
 * undefined when every file has a path of its own: one of the two would
 * otherwise quietly take the other's place.
 */
export function givenTwice(
	files: readonly { readonly path: string }[],
): string | undefined {
	const seen = new Set<string>();
	for (const { path } of files) {
		if (seen.has(path)) {
			return path;
		}
		seen.add(path);
	}
	return undefined;
}

/**
 * keptPictureOf is the file the sharing picture of locale is kept in: the
 * site's own assets, under the name its pages serve it by.
 */
export function keptPictureOf(locale: string): string {
	return join(repository, "site", sharingPicturePath(locale).slice(1));
}

// sharingPictures are the pictures a shared link to a page shows, one for each
// language of the site that has one, kept in the site's assets under the name
// they are served by. A language with none yet still builds, since a picture
// is photographed from a site already built; its pages name a picture the site
// does not have, and a check of the site refuses to publish them so.
async function sharingPictures(locales: readonly string[]): Promise<Made[]> {
	const kept: Made[] = [];
	for (const locale of locales) {
		try {
			kept.push({
				path: sharingPicturePath(locale).slice(1),
				data: await readFile(keptPictureOf(locale)),
			});
		} catch (error) {
			if (
				!(error instanceof Error && "code" in error && error.code === "ENOENT")
			) {
				throw error;
			}
		}
	}
	return kept;
}

// buildStyles builds the site's stylesheet into a directory of its own and
// reads back what the build made there.
async function buildStyles(): Promise<Made[]> {
	const dir = await mkdtemp(join(tmpdir(), "site-styles-"));
	try {
		await build({
			configFile: siteConfig,
			logLevel: "warn",
			build: { outDir: dir, emptyOutDir: true },
		});
		const entries = await readdir(dir, {
			recursive: true,
			withFileTypes: true,
		});
		return await Promise.all(
			entries
				.filter((entry) => entry.isFile())
				.map(async (entry) => ({
					path: relative(dir, join(entry.parentPath, entry.name)),
					data: await readFile(join(entry.parentPath, entry.name)),
				})),
		);
	} finally {
		await rm(dir, { recursive: true, force: true });
	}
}

// drawPages draws every page of the site in memory. The components are read
// through Vite, as the tests and the widget's build read them — TSX, the
// dictionaries' JSON, the glob that finds them — by a server that serves
// nothing and watches nothing.
async function drawPages(
	base: string,
	sources: Map<string, Map<string, string>>,
): Promise<SiteFile[]> {
	const server = await createServer({
		configFile: siteConfig,
		logLevel: "warn",
		// Nothing runs in a browser here, so nothing is prepared for one.
		optimizeDeps: { noDiscovery: true },
		server: { middlewareMode: true, ws: false, watch: null },
	});
	try {
		const { renderSite } = (await server.ssrLoadModule(
			"/src/site/render.tsx",
		)) as typeof import("../src/site/render.tsx");
		return renderSite({ base, sources });
	} finally {
		await server.close();
	}
}

const usage =
	"usage: node scripts/prerender-site.ts --base <origin> --out <directory>";

/**
 * main builds the site a command line asks for and returns the command's exit
 * code: 0 when the site is built, 1 when it cannot be, and 2 when the command
 * line does not name an origin and a directory.
 */
export async function main(
	args: string[],
	out: (line: string) => void,
	err: (line: string) => void,
): Promise<number> {
	let values: { base?: string; out?: string };
	try {
		({ values } = parseArgs({
			args,
			options: { base: { type: "string" }, out: { type: "string" } },
		}));
	} catch {
		err(usage);
		return 2;
	}
	if (values.base === undefined || values.out === undefined) {
		err(usage);
		return 2;
	}
	try {
		await buildSite({ base: values.base, out: resolve(values.out) });
	} catch (error) {
		err(
			`site: ${error instanceof Error ? error.message : "it cannot be built"}`,
		);
		return 1;
	}
	out(`site: built into ${values.out}`);
	return 0;
}

if (import.meta.main) {
	// Vite takes a build's mode from NODE_ENV, once for the whole process, and
	// its first call sets it when nothing has: the site is built to be
	// published, so that is said before either call.
	process.env.NODE_ENV ??= "production";
	process.exitCode = await main(
		process.argv.slice(2),
		(line) => console.log(line),
		(line) => console.error(line),
	);
}
