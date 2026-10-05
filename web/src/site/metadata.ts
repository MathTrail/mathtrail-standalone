import { address, alternatesOf } from "./addresses";
import { coachPrototypePath, photoDirectory } from "./brand";
import { frontPage, type Texts } from "./content";

// The names the sitemap protocol gives its two vocabularies. They are names,
// not addresses anything is fetched from, and the protocol spells them with
// http.
const sitemapVocabulary = "http://www.sitemaps.org/schemas/sitemap/0.9";
const xhtmlVocabulary = "http://www.w3.org/1999/xhtml";

/**
 * sitemap lists every address the site serves, each with its translations, so
 * that a crawler learns the whole matrix of languages from one file. The apex
 * comes first, then every locale's pages.
 */
export function sitemap(texts: Texts, base: string, reference: string): string {
	const entry = (served: string, name: string): string =>
		[
			"  <url>",
			`    <loc>${served}</loc>`,
			...alternatesOf(texts, base, reference, name).map(
				({ hreflang, url }) =>
					`    <xhtml:link rel="alternate" hreflang="${hreflang}" href="${url}"/>`,
			),
			"  </url>",
		].join("\n");

	const entries = [entry(`${base}/`, frontPage)];
	for (const locale of texts.locales) {
		for (const name of texts.names) {
			entries.push(entry(base + address(locale, name), name));
		}
	}
	return [
		'<?xml version="1.0" encoding="UTF-8"?>',
		`<urlset xmlns="${sitemapVocabulary}" xmlns:xhtml="${xhtmlVocabulary}">`,
		...entries,
		"</urlset>",
		"",
	].join("\n");
}

/**
 * robots lets every crawler in but to the coach's prototype and the site's
 * photographs, and tells it where the sitemap is. The prototype is a document
 * the coach's page frames, not a page: read alone, it would make a search
 * result with no word of what it is, though its address may still be listed
 * from the links to it. The photographs are the family's, children's faces
 * among them: the page that shows them is found by a search, and they are not
 * found apart from it.
 */
export function robots(base: string): string {
	return `User-agent: *\nAllow: /\nDisallow: ${coachPrototypePath}\nDisallow: ${photoDirectory}\n\nSitemap: ${base}/sitemap.xml\n`;
}

/**
 * cname claims the site's domain. The host serves whatever it is handed, so
 * the domain is claimed by a file rather than by a setting somebody has to
 * remember.
 */
export function cname(origin: URL): string {
	return `${origin.host}\n`;
}
