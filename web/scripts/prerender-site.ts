// The site's build. It reads the texts in site/content/<locale>/*.md, draws
// every page with the site's components, builds the stylesheet, and writes the
// whole site into the one directory a static host publishes.
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
import { tmpdir } from "node:os";
import { dirname, join, relative, resolve } from "node:path";
import { parseArgs } from "node:util";
import { build, createServer } from "vite";
import type { SiteFile } from "../src/site/render.tsx";

const web = join(import.meta.dirname, "..");
const repository = join(web, "..");
const siteConfig = join(web, "vite.config.site.ts");

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
 * locale, and every .md file in a locale is a page named after the file.
 * Anything else is left where it is.
 */
export async function readSources(
	dir: string,
): Promise<Map<string, Map<string, string>>> {
	const sources = new Map<string, Map<string, string>>();
	for (const locale of await readdir(dir, { withFileTypes: true })) {
		if (!locale.isDirectory()) {
			continue;
		}
		const pages = new Map<string, string>();
		for (const file of await readdir(join(dir, locale.name), {
			withFileTypes: true,
		})) {
			if (file.isFile() && file.name.endsWith(".md")) {
				pages.set(
					file.name.slice(0, -".md".length),
					await readFile(join(dir, locale.name, file.name), "utf8"),
				);
			}
		}
		sources.set(locale.name, pages);
	}
	return sources;
}

/** Made is one file of a build, ready to be written below the site's root. */
type Made = { readonly path: string; readonly data: string | Buffer };

/**
 * buildSite builds the site into out: the pages its texts make, the
 * stylesheet, the design's tokens and the mark. Every file is made before
 * anything is written, so a build that fails at any step leaves the last one
 * where it was.
 */
export async function buildSite({
	base,
	out,
	content = join(repository, "site", "content"),
}: SiteOptions): Promise<void> {
	if (!(await replaceable(out))) {
		throw new Error(
			`${out} holds something other than a build of the site, and a build replaces everything in it`,
		);
	}
	const files: Made[] = [
		...(await drawPages(base, await readSources(content))),
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
	];
	const twice = givenTwice(files);
	if (twice !== undefined) {
		throw new Error(`${twice} would be written twice`);
	}
	await rm(out, { recursive: true, force: true });
	for (const { path, data } of files) {
		const target = join(out, path);
		await mkdir(dirname(target), { recursive: true });
		await writeFile(target, data);
	}
}

// replaceable says whether out may be replaced by a build: it does not exist
// yet, it is empty, or it holds a build of the site, which always carries
// .nojekyll. A build deletes everything in out, so anything else — the
// repository itself, a directory named by mistake — is not the build's to
// delete.
async function replaceable(out: string): Promise<boolean> {
	try {
		const entries = await readdir(out);
		return entries.length === 0 || entries.includes(".nojekyll");
	} catch (error) {
		return error instanceof Error && "code" in error && error.code === "ENOENT";
	}
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
