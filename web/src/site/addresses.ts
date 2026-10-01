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

/** outputPath is the file that serves an address, below the site's root. */
export function outputPath(served: string): string {
	return `${served.slice(1)}index.html`;
}

/**
 * Alternate is one hreflang link: a translation of a page, or the x-default a
 * reader with no matching language is sent to.
 */
export type Alternate = { readonly hreflang: string; readonly url: string };

/** localesWith are the locales that have the page name, in order. */
export function localesWith(texts: Texts, name: string): string[] {
	return texts.locales.filter((locale) => texts.pages.get(locale)?.has(name));
}

/**
 * alternatesOf lists every translation of the page name, plus the x-default a
 * reader with no matching language is sent to. For a front page that is the
 * apex, which hands the reader every language; for anything else it is the
 * reference locale's version, the copy every other is translated from.
 */
export function alternatesOf(
	texts: Texts,
	base: string,
	reference: string,
	name: string,
): Alternate[] {
	return [
		...localesWith(texts, name).map((locale) => ({
			hreflang: locale,
			url: base + address(locale, name),
		})),
		{
			hreflang: "x-default",
			url: name === frontPage ? `${base}/` : base + address(reference, name),
		},
	];
}
