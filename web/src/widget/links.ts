import type { Site } from "./payload";

/**
 * Anchor is the part of a topic's page a link leads to: its top, where the
 * mistakes made in the topic are told, or how to help at home.
 */
export type Anchor = "" | "#traps" | "#home";

/**
 * everyPageLanguage is the language every page of the site is written in, the
 * one a card in a language the site does not speak links to.
 */
export const everyPageLanguage = "en";

// slugPattern is the form of a topic's slug in the catalog: words of small
// letters and digits joined by hyphens, nothing a path could be turned with.
const slugPattern = /^[a-z0-9]+(-[a-z0-9]+)*$/;

/**
 * PageOf is what a card knows of a topic's page: its slug, and whether the
 * page is published.
 */
export type PageOf = { slug?: string; site_page?: boolean };

/**
 * pageAddress is the address of a topic's page on site, in language when the
 * site is written in it and in English otherwise, at anchor — or undefined
 * when the card links nowhere: no site, a site that is not an https origin and
 * nothing more, a slug of another form, or a page not published. The card
 * builds the address from the site's origin, its language and the slug, and
 * never takes an address from the payload whole; nothing of the child is in
 * it.
 */
export function pageAddress(
	site: Site | undefined,
	language: string,
	page: PageOf,
	anchor: Anchor,
): string | undefined {
	if (
		site === undefined ||
		page.site_page !== true ||
		page.slug === undefined ||
		!slugPattern.test(page.slug)
	) {
		return undefined;
	}
	const origin = httpsOrigin(site.url);
	if (origin === undefined) {
		return undefined;
	}
	const written = site.languages.includes(language)
		? language
		: everyPageLanguage;
	return `${origin}/${written}/topics/${page.slug}/${anchor}`;
}

// httpsOrigin is address when it is the origin of an https site and nothing
// more — no path, query, fragment or credentials, written as a browser writes
// an origin — and undefined otherwise.
function httpsOrigin(address: string): string | undefined {
	let parsed: URL;
	try {
		parsed = new URL(address);
	} catch {
		return undefined;
	}
	return parsed.protocol === "https:" && parsed.origin === address
		? address
		: undefined;
}
