import { siteName } from "./brand";
import { scriptJSON } from "./scriptjson";

// vocabulary names the words structured data is written in. It is a name, not
// an address anything is fetched from, and schema.org spells it with https.
const vocabulary = "https://schema.org";

/**
 * WebSite is the site's name in a search engine's words, which it shows
 * beside the site's results. A search engine reads it from the front page of
 * the bare domain alone.
 */
export type WebSite = {
	readonly "@context": typeof vocabulary;
	readonly "@type": "WebSite";
	readonly name: string;
	readonly url: string;
};

/**
 * BreadcrumbList is a page's path in a search engine's words: the steps a
 * reader takes to the page, numbered from one, the page itself the last.
 */
export type BreadcrumbList = {
	readonly "@context": typeof vocabulary;
	readonly "@type": "BreadcrumbList";
	readonly itemListElement: readonly ListItem[];
};

/** ListItem is a step of a path: its place, its words and its address. */
type ListItem = {
	readonly "@type": "ListItem";
	readonly position: number;
	readonly name: string;
	readonly item: string;
};

/** StructuredData is what a page says of itself in schema.org's words. */
export type StructuredData = WebSite | BreadcrumbList;

/** Crumb is a step of a page's path: its words and the page it leads to. */
export type Crumb = { readonly name: string; readonly href: string };

/** webSite is the site's name, said by the front page at url. */
export function webSite(url: string): WebSite {
	return { "@context": vocabulary, "@type": "WebSite", name: siteName, url };
}

/**
 * breadcrumbList is the path steps make, each step's address taken on the
 * site published at base.
 */
export function breadcrumbList(
	base: string,
	steps: readonly Crumb[],
): BreadcrumbList {
	return {
		"@context": vocabulary,
		"@type": "BreadcrumbList",
		itemListElement: steps.map(({ name, href }, at) => ({
			"@type": "ListItem",
			position: at + 1,
			name,
			item: base + href,
		})),
	};
}

/**
 * StructuredDataScript is the element a page's head carries its structured
 * data in: a script of JSON, which a browser never runs.
 */
export function StructuredDataScript({ data }: { data: StructuredData }) {
	return (
		<script
			type="application/ld+json"
			dangerouslySetInnerHTML={{ __html: scriptJSON(data) }}
		/>
	);
}
