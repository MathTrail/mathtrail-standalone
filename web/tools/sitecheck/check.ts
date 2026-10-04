// The site's checker. It reads a built site and reports what a browser, a
// crawler or a reviewer would otherwise each find broken separately: a link
// that leads nowhere, a page that quietly fetches from somebody else's domain,
// a head that a search result cannot be built from, a locale that lost a page,
// a page that grew past its budget, an address the site handed out and lost.
//
// It reads the finished directory and nothing of what drew it, so it keeps
// judging the same way whatever builds the site.

import { readdir, readFile, stat } from "node:fs/promises";
import { join, posix, relative, sep } from "node:path";
import {
	indexFile,
	isExternal,
	type Page,
	parsePage,
	resolveReference,
} from "./page.ts";

/** Options are the expectations a built site cannot state about itself. */
export type Options = {
	/**
	 * base is the origin the site is published on, with no trailing slash;
	 * canonical and alternate addresses are measured against it.
	 */
	base: string;
	/** referenceLocale is the locale whose pages every other locale must have. */
	referenceLocale: string;
	/**
	 * maxPageBytes caps a page together with the local files it pulls in; zero
	 * leaves the size unchecked.
	 */
	maxPageBytes: number;
	/**
	 * published are the addresses the site has handed out and must keep: a
	 * page's, or an anchor on one, such as "/en/#connect".
	 */
	published: readonly string[];
};

/** Rule names the rule a finding breaks. */
export type Rule =
	| "external"
	| "head"
	| "link"
	| "published"
	| "translation"
	| "weight";

/** Finding is one thing that is wrong, named where a person can go and fix it. */
export type Finding = {
	/** path is the file the finding is about, relative to the site's root. */
	path: string;
	rule: Rule;
	/** message says what is wrong, in the words the fixer needs. */
	message: string;
};

/** lineOf renders a finding as one line. */
export function lineOf(finding: Finding): string {
	return `${finding.path}: ${finding.rule}: ${finding.message}`;
}

// These are what a stylesheet is searched for, not what the checker loads: a
// stylesheet that names any of them is the finding.
const foreignMarkers = ["http://", "https://", "url(//"];

// pageAddress matches what the list of published addresses may hold: a page's
// address, which ends in a slash, and perhaps one anchor on it.
const pageAddress = /^\/(?:[^#?]*\/)?(?:#[^#?]+)?$/;

/**
 * check judges the built site in dir and returns every finding, sorted by file,
 * rule and message; none means the site is publishable. It throws when the site
 * cannot be judged at all.
 */
export async function check(dir: string, options: Options): Promise<Finding[]> {
	if (options.base === "") {
		throw new Error("base URL is empty");
	}
	if (options.referenceLocale === "") {
		throw new Error("reference locale is empty");
	}
	const misread = options.published.find((p) => !pageAddress.test(p));
	if (misread !== undefined) {
		throw new Error(`published address "${misread}" is no page's address`);
	}

	const files = await readSizes(dir);
	const pages = await readPages(dir, files);
	if (!pages.has(address(options.referenceLocale, ""))) {
		throw new Error(
			`reference locale "${options.referenceLocale}" has no front page`,
		);
	}

	const findings: Finding[] = [];
	for (const page of pages.values()) {
		findings.push(
			...links(page, files),
			...external(page),
			...head(page, pages, options),
			...weight(page, files, options),
		);
	}
	findings.push(
		...translations(pages, options),
		...(await stylesheets(dir, files)),
		...publishedAddresses(files, pages, options),
	);
	return findings.sort(
		(a, b) =>
			byCodeUnits(a.path, b.path) ||
			byCodeUnits(a.rule, b.rule) ||
			byCodeUnits(a.message, b.message),
	);
}

// readSizes lists every file of the site with its size, which is both whether a
// link leads somewhere and what a page weighs.
async function readSizes(dir: string): Promise<Map<string, number>> {
	const entries = await readdir(dir, { recursive: true, withFileTypes: true });
	const sizes = new Map<string, number>();
	for (const entry of entries) {
		if (!entry.isFile()) {
			continue;
		}
		const file = join(entry.parentPath, entry.name);
		sizes.set(
			relative(dir, file).split(sep).join("/"),
			(await stat(file)).size,
		);
	}
	if (sizes.size === 0) {
		throw new Error("the site is empty");
	}
	return sizes;
}

// readPages parses every page of the site, by the address it is served at.
async function readPages(
	dir: string,
	files: Map<string, number>,
): Promise<Map<string, Page>> {
	const pages = new Map<string, Page>();
	for (const file of files.keys()) {
		if (posix.basename(file) !== indexFile) {
			continue;
		}
		const page = parsePage(file, await readFile(join(dir, file), "utf8"));
		pages.set(page.address, page);
	}
	if (pages.size === 0) {
		throw new Error("the site holds no page");
	}
	return pages;
}

// links reports every reference that leads to a file the site does not have.
function links(page: Page, files: Map<string, number>): Finding[] {
	return page.references.flatMap((ref) => {
		if (isExternal(ref.value)) {
			return [];
		}
		const target = resolveReference(ref.value, page.file);
		if (target === "" || files.has(target)) {
			return [];
		}
		return [
			{
				path: page.file,
				rule: "link",
				message: `<${ref.element} ${ref.attribute}="${ref.value}"> leads to ${target}, which the site does not have`,
			},
		];
	});
}

// external reports anything the page fetches from another origin. A link a
// reader may follow is fine; a file the browser loads on its own is not, because
// it hands a third party every visit to the page.
function external(page: Page): Finding[] {
	return page.references
		.filter((ref) => ref.subresource && isExternal(ref.value))
		.map((ref) => ({
			path: page.file,
			rule: "external",
			message: `<${ref.element} ${ref.attribute}="${ref.value}"> loads from another origin`,
		}));
}

// head reports a head that a search result, a shared link or a screen reader
// cannot be built from.
function head(
	page: Page,
	pages: Map<string, Page>,
	options: Options,
): Finding[] {
	const findings: Finding[] = [];
	const report = (message: string) =>
		findings.push({ path: page.file, rule: "head", message });

	if (page.lang === "") {
		report("<html> carries no lang");
	} else if (page.locale !== "" && page.lang !== page.locale) {
		report(
			`<html lang="${page.lang}"> disagrees with the locale "${page.locale}" the page is served under`,
		);
	}
	if (page.dir === "") {
		report("<html> carries no dir");
	}
	if (page.title === "") {
		report("the page has no title");
	}
	if (page.description === "") {
		report("the page has no description");
	}

	const canonical = options.base + page.address;
	if (page.canonical !== canonical) {
		report(
			`canonical is "${page.canonical}", and the page is served at "${canonical}"`,
		);
	}

	for (const [hreflang, url] of expectedAlternates(page, pages, options)) {
		const named = page.alternates.get(hreflang);
		if (named === undefined) {
			report(`no alternate for "${hreflang}", which the site has at "${url}"`);
		} else if (named !== url) {
			report(
				`alternate for "${hreflang}" is "${named}", and that page is at "${url}"`,
			);
		}
	}
	return findings;
}

// expectedAlternates works out which translations a page must point at, from
// the pages the site has, and where its x-default leads.
function expectedAlternates(
	page: Page,
	pages: Map<string, Page>,
	options: Options,
): [string, string][] {
	const alternates: [string, string][] = localesOf(pages)
		.filter((locale) => pages.has(address(locale, page.name)))
		.map((locale) => [locale, options.base + address(locale, page.name)]);
	const fallback =
		page.name === ""
			? `${options.base}/`
			: options.base + address(options.referenceLocale, page.name);
	return [...alternates, ["x-default", fallback]];
}

// translations reports a locale missing a page the reference locale has, which
// is how a language quietly loses its privacy policy.
function translations(pages: Map<string, Page>, options: Options): Finding[] {
	const reference = [...namesOf(pages, options.referenceLocale)].sort(
		byCodeUnits,
	);
	return localesOf(pages)
		.filter((locale) => locale !== options.referenceLocale)
		.flatMap((locale) => {
			const has = namesOf(pages, locale);
			return reference
				.filter((name) => !has.has(name))
				.map((name) => ({
					path: address(locale, name).slice(1) + indexFile,
					rule: "translation",
					message: `locale "${locale}" is missing a page that "${options.referenceLocale}" has`,
				}));
		});
}

// weight reports a page that, with everything it pulls in, costs its reader
// more than the budget allows.
function weight(
	page: Page,
	files: Map<string, number>,
	options: Options,
): Finding[] {
	if (options.maxPageBytes === 0) {
		return [];
	}
	const counted = new Set([page.file]);
	for (const ref of page.references) {
		if (!ref.subresource || isExternal(ref.value)) {
			continue;
		}
		const target = resolveReference(ref.value, page.file);
		if (target !== "") {
			counted.add(target);
		}
	}
	let total = 0;
	for (const file of counted) {
		total += files.get(file) ?? 0;
	}
	if (total <= options.maxPageBytes) {
		return [];
	}
	return [
		{
			path: page.file,
			rule: "weight",
			message: `the page and what it loads come to ${total} bytes, and the budget is ${options.maxPageBytes}`,
		},
	];
}

// stylesheets reports a stylesheet that reaches for another origin. A font or an
// image pulled in from a stylesheet is invisible in the HTML, and it hands a
// third party every visit as surely as a script would.
async function stylesheets(
	dir: string,
	files: Map<string, number>,
): Promise<Finding[]> {
	const findings: Finding[] = [];
	for (const file of files.keys()) {
		if (!file.endsWith(".css")) {
			continue;
		}
		const text = await readFile(join(dir, file), "utf8");
		for (const marker of foreignMarkers) {
			if (text.includes(marker)) {
				findings.push({
					path: file,
					rule: "external",
					message: `the stylesheet mentions "${marker}", so it may load from another origin`,
				});
			}
		}
	}
	return findings;
}

// publishedAddresses reports an address the site handed out that the build no
// longer has, or an anchor its page no longer holds: somebody else holds a link
// to it, and nothing on the site may still point at it for the link rule to
// notice.
function publishedAddresses(
	files: Map<string, number>,
	pages: Map<string, Page>,
	options: Options,
): Finding[] {
	return options.published.flatMap((published): Finding[] => {
		const [address = "", anchor = ""] = published.split("#");
		const file = address.slice(1) + indexFile;
		if (!files.has(file)) {
			return [
				{
					path: file,
					rule: "published",
					message: `the site has published ${published}, and the build no longer has it`,
				},
			];
		}
		if (anchor !== "" && !pages.get(address)?.ids.has(anchor)) {
			return [
				{
					path: file,
					rule: "published",
					message: `the site has published ${published}, and the page has no element with the id "${anchor}"`,
				},
			];
		}
		return [];
	});
}

// localesOf lists the locales the site has, in a stable order.
function localesOf(pages: Map<string, Page>): string[] {
	const locales = new Set<string>();
	for (const page of pages.values()) {
		if (page.locale !== "") {
			locales.add(page.locale);
		}
	}
	return [...locales].sort(byCodeUnits);
}

// namesOf lists the pages one locale has.
function namesOf(pages: Map<string, Page>, locale: string): Set<string> {
	const names = new Set<string>();
	for (const page of pages.values()) {
		if (page.locale === locale) {
			names.add(page.name);
		}
	}
	return names;
}

// address is the path a locale's page is served at.
function address(locale: string, name: string): string {
	return name === "" ? `/${locale}/` : `/${locale}/${name}/`;
}

// byCodeUnits orders strings by their code units, so that a report reads the
// same on every machine.
function byCodeUnits(a: string, b: string): number {
	if (a === b) {
		return 0;
	}
	return a < b ? -1 : 1;
}
