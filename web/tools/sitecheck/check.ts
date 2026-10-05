// The site's checker. It reads a built site and reports what a browser, a
// crawler or a reviewer would otherwise each find broken separately: a link
// that leads nowhere, a page that quietly fetches from somebody else's domain,
// a head that a search result cannot be built from, a locale that lost a page,
// a page, or a document a page frames, that grew past its budget, an address
// the site handed out and lost, a page it serves without having listed it
// among the addresses it keeps, and a script that loads more than itself, or
// sends or keeps something.
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
import { stylesheetReferences } from "./stylesheet.ts";

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
	 * maxPageBytes caps a page together with the local files it pulls in, but
	 * for a document it frames, and the files its stylesheets pull in; zero
	 * leaves the size unchecked.
	 */
	maxPageBytes: number;
	/**
	 * maxFrameBytes caps a document a page frames, together with what it pulls
	 * in. A frame holds an application of its own, which a reader meets in the
	 * page rather than as part of it, so it is weighed apart from the page; zero
	 * leaves the size unchecked.
	 */
	maxFrameBytes: number;
	/**
	 * photos is the directory of the site's photographs, from its root, such
	 * as "/assets/photos/". A photograph is worth its weight, so a page is
	 * weighed without the ones it shows; empty weighs them with the page, as
	 * any other file.
	 */
	photos: string;
	/**
	 * published are the addresses the site has handed out and must keep: a
	 * page's, or an anchor on one, such as "/en/#connect". Every page the site
	 * serves is among them.
	 */
	published: readonly string[];
};

/** Rule names the rule a finding breaks. */
export type Rule =
	| "external"
	| "head"
	| "link"
	| "published"
	| "script"
	| "translation"
	| "weight";

/**
 * Load is one address a stylesheet loads, as it is written, and the file of
 * the site it asks for.
 */
type Load = { value: string; target: string };

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

// directoryAddress matches a directory of the site, from its root: it begins
// and ends with a slash, so that it holds what lies below it and nothing whose
// name merely begins the same.
const directoryAddress = /^\/(?:[^/#?]+\/)+$/;

/**
 * check judges the built site in dir and returns every finding, sorted by file,
 * rule and message; none means the site is publishable. It throws when the site
 * cannot be judged at all.
 */
export async function check(dir: string, options: Options): Promise<Finding[]> {
	if (options.base === "") {
		throw new Error("base URL is empty");
	}
	// Every address the checker expects is the base with a path after it, so a
	// base that is not an origin would turn every page into a finding.
	if (!isOrigin(options.base)) {
		throw new Error(
			`base URL "${options.base}" is not an origin: a scheme and a host, with no path and no slash at the end`,
		);
	}
	if (options.referenceLocale === "") {
		throw new Error("reference locale is empty");
	}
	const misread = options.published.find((p) => !pageAddress.test(p));
	if (misread !== undefined) {
		throw new Error(`published address "${misread}" is no page's address`);
	}
	if (options.photos !== "" && !directoryAddress.test(options.photos)) {
		throw new Error(
			`photo directory "${options.photos}" is no directory of the site: it begins and ends with a slash`,
		);
	}

	const files = await readSizes(dir);
	const [pages, sheets] = await Promise.all([
		readPages(dir, files),
		readStylesheets(dir, files),
	]);
	if (!pages.has(address(options.referenceLocale, ""))) {
		throw new Error(
			`reference locale "${options.referenceLocale}" has no front page`,
		);
	}

	const loads = loadsOf(sheets);
	const frames = await readFrames(dir, files, pages);
	const findings: Finding[] = [];
	for (const page of pages.values()) {
		findings.push(
			...links(page, files),
			...external(page),
			...head(page, pages, files, options),
			...weight(page, files, loads, options),
		);
	}
	// A document a page frames is held to what a page is held to about the
	// files it reaches for, and to a budget of its own; it needs no head.
	for (const frame of frames) {
		findings.push(
			...links(frame, files),
			...external(frame),
			...frameWeight(frame, files, loads, options),
		);
	}
	findings.push(
		...stylesheetLinks(loads, files),
		...translations(pages, options),
		...stylesheets(sheets),
		...scripts(await readScripts(dir, files)),
		...publishedAddresses(files, pages, options),
		...unlisted(pages, options),
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
	// A link to a file serves that file, and weighs what it weighs; a link to a
	// directory or to nothing serves no file at all.
	const sized = await Promise.all(
		entries
			.filter((entry) => entry.isFile() || entry.isSymbolicLink())
			.map(async (entry) => {
				const file = join(entry.parentPath, entry.name);
				const info = await stat(file).catch(() => undefined);
				return info?.isFile()
					? ([relative(dir, file).split(sep).join("/"), info.size] as const)
					: undefined;
			}),
	);
	const sizes = new Map(sized.filter((entry) => entry !== undefined));
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
	const parsed = await Promise.all(
		[...files.keys()]
			.filter((file) => posix.basename(file) === indexFile)
			.map(async (file) =>
				parsePage(file, await readFile(join(dir, file), "utf8")),
			),
	);
	const pages = new Map(parsed.map((page) => [page.address, page]));
	if (pages.size === 0) {
		throw new Error("the site holds no page");
	}
	return pages;
}

// frameElements are the elements whose address is a document of its own,
// which a reader meets inside the page rather than as a part of it.
const frameElements = new Set(["iframe"]);

// framedBy lists the files of the site a document frames.
function framedBy(document: Page): string[] {
	return document.references
		.filter((ref) => frameElements.has(ref.element) && !isExternal(ref.value))
		.map((ref) => resolveReference(ref.value, document.file))
		.filter((target) => target !== "");
}

// markupFile matches the names of the files a frame holds as HTML; any other
// file a page frames, such as a picture, loads nothing by itself.
const markupFile = /\.html?$/i;

// readFrames parses every document a page frames, and every one those frame in
// turn, once each. A page that another frames is judged as a page already, and
// a frame that leads nowhere is the link rule's to name, so neither is read;
// nor is a framed file that is no HTML, which is weighed by its own size.
async function readFrames(
	dir: string,
	files: Map<string, number>,
	pages: Map<string, Page>,
): Promise<Page[]> {
	const pageFiles = new Set([...pages.values()].map((page) => page.file));
	const frames = new Map<string, Page>();
	let framers = [...pages.values()];
	while (framers.length > 0) {
		const unread = [...new Set(framers.flatMap(framedBy))].filter(
			(file) => files.has(file) && !pageFiles.has(file) && !frames.has(file),
		);
		framers = await Promise.all(
			unread.map(async (file) =>
				parsePage(
					file,
					markupFile.test(file) ? await readFile(join(dir, file), "utf8") : "",
				),
			),
		);
		for (const frame of framers) {
			frames.set(frame.file, frame);
		}
	}
	return [...frames.values()];
}

// readStylesheets reads every stylesheet of the site, by its file.
async function readStylesheets(
	dir: string,
	files: Map<string, number>,
): Promise<Map<string, string>> {
	const read = await Promise.all(
		[...files.keys()]
			.filter((file) => file.endsWith(".css"))
			.map(
				async (file) =>
					[file, await readFile(join(dir, file), "utf8")] as const,
			),
	);
	return new Map(read);
}

// readScripts reads every script of the site, by its file.
async function readScripts(
	dir: string,
	files: Map<string, number>,
): Promise<Map<string, string>> {
	const read = await Promise.all(
		[...files.keys()]
			.filter((file) => file.endsWith(".js"))
			.map(
				async (file) =>
					[file, await readFile(join(dir, file), "utf8")] as const,
			),
	);
	return new Map(read);
}

// loadsOf are the files of the site each stylesheet loads, by the stylesheet's
// file. An address of another origin is no file of the site: the stylesheet
// rule names it.
function loadsOf(sheets: Map<string, string>): Map<string, Load[]> {
	return new Map(
		[...sheets].map(([file, text]) => [
			file,
			stylesheetReferences(text)
				.filter((value) => !isExternal(value))
				.map((value) => ({ value, target: resolveReference(value, file) }))
				.filter(({ target }) => target !== ""),
		]),
	);
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

// stylesheetLinks reports every file a stylesheet loads that the site does
// not have: a font that never arrives leaves the page in another typeface, and
// nothing else would say so.
function stylesheetLinks(
	loads: Map<string, Load[]>,
	files: Map<string, number>,
): Finding[] {
	return [...loads].flatMap(([file, wanted]) =>
		wanted
			.filter(({ target }) => !files.has(target))
			.map(({ value, target }) => ({
				path: file,
				rule: "link" as const,
				message: `the stylesheet loads "${value}", which leads to ${target}, and the site does not have it`,
			})),
	);
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
// cannot be built from. The picture a shared link shows is the site's own, a
// file it serves: a link to it that leads nowhere shows a preview with none.
function head(
	page: Page,
	pages: Map<string, Page>,
	files: Map<string, number>,
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
	for (const picture of page.images) {
		const problem = pictureProblem(picture, page, files, options);
		if (problem !== undefined) {
			report(problem);
		}
	}

	const canonical = canonicalProblem(page, pages, options);
	if (canonical !== undefined) {
		report(canonical);
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

// canonicalProblem is what is wrong with the address page names as its own, or
// undefined when nothing is. A page names the address it is served at. The
// apex alone may send its reader on instead, to a page of the site, and then
// stands for that page and names its address: a page of a language that sends
// its reader on is a translation the language has lost.
function canonicalProblem(
	page: Page,
	pages: Map<string, Page>,
	options: Options,
): string | undefined {
	const served = options.base + page.address;
	if (page.refresh === "") {
		return page.canonical === served
			? undefined
			: `canonical is "${page.canonical}", and the page is served at "${served}"`;
	}
	const sends = `the page sends its reader on to "${page.refresh}"`;
	if (page.locale !== "") {
		return `${sends}, and only the apex may: a page of a language is read in it`;
	}
	const target = URL.parse(page.refresh, served);
	if (
		target === null ||
		target.origin !== options.base ||
		!pages.has(target.pathname)
	) {
		return `${sends}, which is no page of the site`;
	}
	return page.canonical === target.href
		? undefined
		: `canonical is "${page.canonical}", and the page sends its reader on to "${target.href}"`;
}

// pictureProblem is what is wrong with picture, one a shared link to page
// shows, or undefined when it is a file the site serves.
function pictureProblem(
	picture: string,
	page: Page,
	files: Map<string, number>,
	options: Options,
): string | undefined {
	const named = `the picture a shared link shows, "${picture}",`;
	if (!picture.startsWith(`${options.base}/`)) {
		return `${named} is not on ${options.base}`;
	}
	const file = resolveReference(picture.slice(options.base.length), page.file);
	return files.has(file)
		? undefined
		: `${named} is a file the site does not have`;
}

// expectedAlternates works out which translations a page must point at, from
// the pages the site has, and where its x-default leads: to the reference
// locale's version of the page. A page that sends its reader on has no
// translations of its own to name: they are the page's it stands for.
function expectedAlternates(
	page: Page,
	pages: Map<string, Page>,
	options: Options,
): [string, string][] {
	if (page.refresh !== "") {
		return [];
	}
	const alternates: [string, string][] = localesOf(pages)
		.filter((locale) => pages.has(address(locale, page.name)))
		.map((locale) => [locale, options.base + address(locale, page.name)]);
	return [
		...alternates,
		["x-default", options.base + address(options.referenceLocale, page.name)],
	];
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

// weight reports a page that, with everything it pulls in but the site's
// photographs, costs its reader more than the budget allows.
function weight(
	page: Page,
	files: Map<string, number>,
	loads: Map<string, Load[]>,
	options: Options,
): Finding[] {
	if (options.maxPageBytes === 0) {
		return [];
	}
	const total = weightOf(page, files, loads, options.photos);
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

// frameWeight reports a document a page frames that, with everything it pulls
// in, costs its reader more than a framed document's budget allows.
function frameWeight(
	frame: Page,
	files: Map<string, number>,
	loads: Map<string, Load[]>,
	options: Options,
): Finding[] {
	if (options.maxFrameBytes === 0) {
		return [];
	}
	const total = weightOf(frame, files, loads, options.photos);
	if (total <= options.maxFrameBytes) {
		return [];
	}
	return [
		{
			path: frame.file,
			rule: "weight",
			message: `a page frames it, and with what it loads it comes to ${total} bytes; the budget of a framed document is ${options.maxFrameBytes}`,
		},
	];
}

// weightOf is what a document costs its reader: its own file, every file it
// loads but for a document it frames, which is weighed on its own, and every
// file its stylesheets load in turn — each font subset they declare, as if all
// were fetched. A photograph below the directory photos is no part of it.
function weightOf(
	document: Page,
	files: Map<string, number>,
	loads: Map<string, Load[]>,
	photos: string,
): number {
	const counted = new Set([document.file]);
	for (const ref of document.references) {
		if (
			!ref.subresource ||
			frameElements.has(ref.element) ||
			isExternal(ref.value)
		) {
			continue;
		}
		const target = resolveReference(ref.value, document.file);
		if (target !== "") {
			counted.add(target);
		}
	}
	// A set visits what is added to it while it is walked, so a stylesheet
	// another one loads is counted, and what it loads, once each.
	for (const file of counted) {
		for (const { target } of loads.get(file) ?? []) {
			counted.add(target);
		}
	}
	let total = 0;
	for (const file of counted) {
		if (photos === "" || !`/${file}`.startsWith(photos)) {
			total += files.get(file) ?? 0;
		}
	}
	return total;
}

// importMarker finds in a script what loads another file as it runs: an
// import written into the module, from a file or for what the file does, an
// import called, and an export passed on from another module. A word that
// begins with "import", and import.meta, load nothing.
const importMarker =
	/\bimport\s*(?:[\w$*{][^;]*?\bfrom\s*["'`]|["'`(])|\bexport\s*(?:\*|\{[^}]*\})[^;]*?\bfrom\s*["'`]/;

// sendingWays are the usual ways a script sends something away or keeps it
// in the reader's browser, each by its name and how it is written. They are
// found by their names, the same as a script written by hand or by a bundler
// says them: a script could still hide one, and a reader of its source is the
// last check.
const sendingWays: readonly (readonly [string, RegExp])[] = [
	["fetch", /\bfetch\b/],
	["XMLHttpRequest", /\bXMLHttpRequest\b/],
	["sendBeacon", /\bsendBeacon\b/],
	["WebSocket", /\bWebSocket\b/],
	["EventSource", /\bEventSource\b/],
	["new Image", /\bnew\s+Image\b/],
	["serviceWorker", /\bserviceWorker\b/],
	["localStorage", /\blocalStorage\b/],
	["sessionStorage", /\bsessionStorage\b/],
	["indexedDB", /\bindexedDB\b/],
	["cookie", /\bcookie\b/],
];

// scripts reports a script that loads another file, which the checker does
// not follow, so that neither its weight nor its address would be checked;
// and a script that names a usual way to send or keep something, which a
// script of the site never does.
function scripts(read: Map<string, string>): Finding[] {
	return [...read].flatMap(([file, text]): Finding[] => [
		...(importMarker.test(text)
			? [
					{
						path: file,
						rule: "script" as const,
						message:
							"the script loads another file, which the checker does not follow: a script of the site is one file, weighed whole",
					},
				]
			: []),
		...sendingWays
			.filter(([, written]) => written.test(text))
			.map(([name]) => ({
				path: file,
				rule: "script" as const,
				message: `the script names ${name}, and a script of the site stores nothing and sends nothing`,
			})),
	]);
}

// stylesheets reports a stylesheet that reaches for another origin. A font or an
// image pulled in from a stylesheet is invisible in the HTML, and it hands a
// third party every visit as surely as a script would.
function stylesheets(sheets: Map<string, string>): Finding[] {
	return [...sheets].flatMap(([file, text]) =>
		foreignMarkers
			.filter((marker) => text.includes(marker))
			.map(
				(marker): Finding => ({
					path: file,
					rule: "external",
					message: `the stylesheet mentions "${marker}", so it may load from another origin`,
				}),
			),
	);
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

// unlisted reports a page the site serves that the list of published
// addresses does not hold. Every page is somebody's link once it is out, and
// only an address on the list is held to staying, so a page joins the list
// with the change that publishes it.
function unlisted(pages: Map<string, Page>, options: Options): Finding[] {
	const listed = new Set(options.published.map((p) => p.split("#")[0]));
	return [...pages.values()]
		.filter((page) => !listed.has(page.address))
		.map((page) => ({
			path: page.file,
			rule: "published",
			message: `the site serves ${page.address}, which the list of published addresses does not hold`,
		}));
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

// isOrigin reports whether base is an origin written as a browser writes one: a
// web scheme and a host, perhaps a port, and nothing after them.
function isOrigin(base: string): boolean {
	try {
		const url = new URL(base);
		return (
			(url.protocol === "https:" || url.protocol === "http:") &&
			url.origin === base
		);
	} catch {
		return false;
	}
}

// byCodeUnits orders strings by their code units, so that a report reads the
// same on every machine.
function byCodeUnits(a: string, b: string): number {
	if (a === b) {
		return 0;
	}
	return a < b ? -1 : 1;
}
