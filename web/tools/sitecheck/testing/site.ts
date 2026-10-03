// Sites for the checker's tests: directories of pages that break none of its
// rules, so that a test can break exactly one of them and see only that.

import { mkdir, writeFile } from "node:fs/promises";
import { dirname, join } from "node:path";

/** base is the origin the sites of the tests are published on. */
export const base = "https://example.test";

/**
 * pageAt is a page at address, among a site whose pages are at addresses, that
 * breaks no rule: it names its language, its direction, its title, its
 * description, its own address and every translation the site has of it.
 */
export function pageAt(address: string, addresses: readonly string[]): string {
	const name = nameOf(address);
	const alternates = addresses
		.filter((other) => other !== "/" && nameOf(other) === name)
		.map((other) => `${localeOf(other)} ${other}`);
	alternates.push(`x-default ${name === "" ? "/" : `/en/${name}/`}`);
	return [
		`<!DOCTYPE html><html lang="${localeOf(address) || "en"}" dir="ltr"><head>`,
		`<title>Title</title><meta name="description" content="Description">`,
		`<link rel="canonical" href="${base}${address}">`,
		...alternates.map((alternate) => {
			const [hreflang, href] = alternate.split(" ");
			return `<link rel="alternate" hreflang="${hreflang}" href="${base}${href}">`;
		}),
		`<link rel="stylesheet" href="/assets/style.css">`,
		`</head><body><a href="https://github.com/example">Source</a></body></html>`,
	].join("");
}

/** siteAt is every file of a site that breaks no rule and has addresses. */
export function siteAt(addresses: readonly string[]): Record<string, string> {
	const files: Record<string, string> = {
		"assets/style.css": "body{color:#000}",
	};
	for (const address of addresses) {
		files[`${address.slice(1)}index.html`] = pageAt(address, addresses);
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
