import { frontPage, type Texts } from "./content";

/**
 * parseBase checks that raw is the bare origin every address on the site is
 * built from — a scheme and a host, with no path and no trailing slash —
 * written exactly as a browser writes an origin, and returns it as a URL.
 * Every address is built from raw itself, so a base a parser would quietly
 * mend — a stray "?", a host in capitals, a name and a password — is refused
 * rather than carried into every page.
 */
export function parseBase(raw: string): URL {
	if (raw === "") {
		throw new Error("the base URL is empty");
	}
	if (raw.endsWith("/")) {
		throw new Error(`the base URL "${raw}" ends in a slash`);
	}
	let parsed: URL;
	try {
		parsed = new URL(raw);
	} catch {
		throw new Error(`the base URL "${raw}" is not an address`);
	}
	// An http or https URL always has a host: one that names none is not an
	// address at all, and is refused above.
	if (parsed.protocol !== "https:" && parsed.protocol !== "http:") {
		throw new Error(`the base URL "${raw}" is neither https nor http`);
	}
	if (parsed.origin !== raw) {
		throw new Error(
			`the base URL "${raw}" carries more than a scheme and a host, or writes them otherwise than ${parsed.origin}`,
		);
	}
	return parsed;
}

/**
 * address is where a page is served. Every page is a directory, so an address
 * never carries a file extension and never has to change when the builder
 * behind it does.
 */
export function address(locale: string, name: string): string {
	return name === frontPage ? `/${locale}/` : `/${locale}/${name}/`;
}

/**
 * sectionAddress is where the section with the id anchor is, on the page name
 * of locale.
 */
export function sectionAddress(
	locale: string,
	name: string,
	anchor: string,
): string {
	return `${address(locale, name)}#${anchor}`;
}

/** outputPath is the file that serves an address, below the site's root. */
export function outputPath(served: string): string {
	return `${served.slice(1)}index.html`;
}

/**
 * Alternate is one hreflang link: a translation of a page, or the x-default a
 * reader with no matching language is sent to.
 */
export type Alternate = { readonly hreflang: string; readonly url: string };

/**
 * alternatesOf lists every translation of the page name — every locale has
 * every page — plus the x-default a reader with no matching language is sent
 * to: the reference locale's version, the copy every other is translated from.
 * A front page is no exception, since the apex only sends its reader on to the
 * reference locale's.
 */
export function alternatesOf(
	texts: Pick<Texts, "locales">,
	base: string,
	reference: string,
	name: string,
): Alternate[] {
	return [
		...texts.locales.map((locale) => ({
			hreflang: locale,
			url: base + address(locale, name),
		})),
		{ hreflang: "x-default", url: base + address(reference, name) },
	];
}
