// Sites for the checker's tests: directories of pages that break none of its
// rules, so that a test can break exactly one of them and see only that.

import { mkdir, writeFile } from "node:fs/promises";
import { dirname, join } from "node:path";

/** base is the origin the sites of the tests are published on. */
export const base = "https://example.test";

/**
 * pageAt is a page at address, among a site whose pages are at addresses, that
 * breaks no rule: it names its language, its direction, its title, its
 * description, its own address, every translation the site has of it and the
 * English one for a reader with none, and holds an element with each of ids.
 */
export function pageAt(
	address: string,
	addresses: readonly string[],
	ids: readonly string[] = [],
): string {
	const name = nameOf(address);
	const alternates = addresses
		.filter((other) => other !== "/" && nameOf(other) === name)
		.map((other) => `${localeOf(other)} ${other}`);
	alternates.push(name === "" ? "x-default /en/" : `x-default /en/${name}/`);
	return [
		`<!DOCTYPE html><html lang="${localeOf(address) || "en"}" dir="ltr"><head>`,
		`<title>Title</title><meta name="description" content="Description">`,
		`<link rel="canonical" href="${base}${address}">`,
		...alternates.map((alternate) => {
			const [hreflang, href] = alternate.split(" ");
			return `<link rel="alternate" hreflang="${hreflang}" href="${base}${href}">`;
		}),
		`<link rel="stylesheet" href="/assets/style.css">`,
		`</head><body><a href="https://github.com/example">Source</a>`,
		...ids.map((id) => `<span id="${id}"></span>`),
		"</body></html>",
	].join("");
}

/**
 * redirectAt is a page that breaks no rule and sends its reader on at once to
 * the page at to: it names its language, its direction, its title and its
 * description, and the page it sends them to as its own address, and loads
 * nothing.
 */
export function redirectAt(to: string): string {
	return [
		`<!DOCTYPE html><html lang="en" dir="ltr"><head>`,
		`<meta http-equiv="refresh" content="0; url=${to}">`,
		`<title>Title</title><meta name="description" content="Description">`,
		`<link rel="canonical" href="${base}${to}">`,
		`</head><body><a href="${to}">Title</a></body></html>`,
	].join("");
}

/**
 * siteAt is every file of a site that breaks no rule and has addresses: a page
 * at each of them, which holds the anchor of every address on it, but for the
 * apex, which sends its reader on to the English front page.
 */
export function siteAt(addresses: readonly string[]): Record<string, string> {
	const files: Record<string, string> = {
		"assets/style.css": "body{color:#000}",
	};
	const pages = addresses.filter((address) => !address.includes("#"));
	for (const address of pages) {
		const ids = addresses
			.filter((other) => other.startsWith(`${address}#`))
			.map((other) => other.slice(address.length + 1));
		files[`${address.slice(1)}index.html`] =
			address === "/" ? redirectAt("/en/") : pageAt(address, pages, ids);
	}
	return files;
}

/** writeSite writes files below dir, each at its path. */
export async function writeSite(
	dir: string,
	files: Record<string, string>,
): Promise<void> {
	for (const [path, text] of Object.entries(files)) {
		await mkdir(dirname(join(dir, path)), { recursive: true });
		await writeFile(join(dir, path), text);
	}
}

// localeOf is the first segment of an address, empty for the apex.
function localeOf(address: string): string {
	return address.split("/")[1] ?? "";
}

// nameOf is the page an address names within its locale, empty for a front
// page and for the apex.
function nameOf(address: string): string {
	return address.split("/").slice(2, -1).join("/");
}
