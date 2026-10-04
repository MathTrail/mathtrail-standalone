// @vitest-environment node
import { mkdtemp, rm, unlink } from "node:fs/promises";
import { tmpdir } from "node:os";
import { join } from "node:path";
import { afterEach, describe, expect, test } from "vitest";
import { check, type Finding, lineOf, type Options } from "./check.ts";
import { base, pageAt, siteAt, writeSite } from "./testing/site.ts";

const addresses = ["/", "/en/", "/ru/"];

function options(): Options {
	return {
		base,
		referenceLocale: "en",
		maxPageBytes: 4096,
		published: addresses,
	};
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
			name: "a published address that is no page's",
			files: siteAt(addresses),
			options: { ...options(), published: [...addresses, "/en/privacy"] },
			message: `published address "/en/privacy" is no page's address`,
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
