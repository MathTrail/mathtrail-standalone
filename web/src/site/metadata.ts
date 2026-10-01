import { address, alternatesOf } from "./addresses";
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
			if (texts.pages.get(locale)?.has(name)) {
				entries.push(entry(base + address(locale, name), name));
			}
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

/** robots lets every crawler in, and tells it where the sitemap is. */
export function robots(base: string): string {
	return `User-agent: *\nAllow: /\n\nSitemap: ${base}/sitemap.xml\n`;
}

/**
 * cname claims the site's domain. The host serves whatever it is handed, so
 * the domain is claimed by a file rather than by a setting somebody has to
 * remember.
 */
export function cname(origin: URL): string {
	return `${origin.host}\n`;
}
