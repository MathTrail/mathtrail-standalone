// @vitest-environment node
import { mkdtemp, rm, writeFile } from "node:fs/promises";
import { tmpdir } from "node:os";
import { join } from "node:path";
import { afterEach, describe, expect, test } from "vitest";
import { main } from "./main.ts";
import { published } from "./published.ts";
import { base, siteAt, writeSite } from "./testing/site.ts";

const made: string[] = [];

// publishable writes a site with nothing wrong in it, at every address the site
// has published, into a directory of its own.
async function publishable(): Promise<string> {
	const dir = await mkdtemp(join(tmpdir(), "sitecheck-"));
	made.push(dir);
	await writeSite(dir, siteAt(published));
	return dir;
}

// command runs the checker's command line, and is what it said on each stream
// and the code it ended with.
async function command(args: string[]) {
	const out: string[] = [];
	const err: string[] = [];
	const code = await main(
		args,
		(line) => out.push(line),
		(line) => err.push(line),
	);
	return { code, out, err };
}

afterEach(async () => {
	await Promise.all(
		made.splice(0).map((dir) => rm(dir, { recursive: true, force: true })),
	);
});

describe("the checker's command line", () => {
	test("passes a site with nothing wrong", async () => {
		const dir = await publishable();

		expect(await command(["--base", base, "--dir", dir])).toEqual({
			code: 0,
			out: [`${dir} is publishable`],
			err: [],
		});
	});

	test("names every problem before it fails, and counts them", async () => {
		const dir = await publishable();
		await writeFile(
			join(dir, "en", "index.html"),
			`<html><body><a href="/en/missing/">gone</a></body></html>`,
		);

		const { code, out, err } = await command(["--base", base, "--dir", dir]);

		expect(code).toBe(1);
		expect(out).toContain(
			`en/index.html: link: <a href="/en/missing/"> leads to en/missing/index.html, which the site does not have`,
		);
		expect(err).toEqual([`sitecheck: ${out.length} problems in ${dir}`]);
	});

	test("refuses a directory that is not a site", async () => {
		const dir = await mkdtemp(join(tmpdir(), "sitecheck-"));
		made.push(dir);

		expect(await command(["--base", base, "--dir", dir])).toEqual({
			code: 1,
			out: [],
			err: ["sitecheck: the site is empty"],
		});
	});

	test("holds a page to the budget it is given", async () => {
		const dir = await publishable();

		const { code, out } = await command([
			"--base",
			base,
			"--dir",
			dir,
			"--max-page-bytes",
			"64",
		]);

		expect(code).toBe(1);
		expect(out.join("\n")).toContain("the budget is 64");
	});

	test.for([
		["no base", ["--dir", "site"]],
		["no directory", ["--base", base]],
		["an option it does not know", ["--base", base, "--dir", "site", "--fix"]],
		[
			"a budget that is no number of bytes",
			["--base", base, "--dir", "site", "--max-page-bytes", "lots"],
		],
		[
			"an empty budget, which would turn the weight rule off",
			["--base", base, "--dir", "site", "--max-page-bytes", ""],
		],
		[
			"a budget written other than in decimal digits",
			["--base", base, "--dir", "site", "--max-page-bytes", "1e3"],
		],
	] as const)("explains its usage for %s", async ([, args]) => {
		const { code, out, err } = await command([...args]);

		expect(code).toBe(2);
		expect(out).toEqual([]);
		expect(err).toEqual([expect.stringMatching(/^usage: /)]);
	});
});
