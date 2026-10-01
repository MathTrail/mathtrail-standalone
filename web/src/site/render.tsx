import type { VNode } from "preact";
import { renderToString } from "preact-render-to-string";
import { disagreements } from "../i18n/dictionaries";
import { fallbackLocale } from "../i18n/lookup";
import { type Dictionary, openWords, type Words } from "../i18n/words";
import { ApexPage } from "./ApexPage";
import {
	address,
	alternatesOf,
	localesWith,
	outputPath,
	parseBase,
} from "./addresses";
import {
	byCodeUnits,
	frontPage,
	readTexts,
	type Sources,
	type Texts,
} from "./content";
import { DocumentPage } from "./DocumentPage";
import type { DocumentLink } from "./Footer";
import { cname, robots, sitemap } from "./metadata";
import { type SiteKey, SiteWords, siteDictionaries } from "./words";

/**
 * SiteFile is one file of the built site: where it goes below the site's root,
 * and what it says.
 */
export type SiteFile = { readonly path: string; readonly data: string };

/**
 * renderSite draws every file the site's texts and words make: a page for each
 * text of each locale, the apex, and the files a crawler and the host read. It
 * writes nothing. The files come back sorted by path, so that two builds of
 * the same texts are the same bytes, and anything wrong with a text or a
 * dictionary stops the build before a single file is written. The stylesheets
 * and the mark are not among them: they are built and copied beside these.
 *
 * base is the origin the site is published on. The dictionaries are the
 * site's own unless a test hands it others.
 */
export function renderSite({
	base,
	sources,
	dictionaries = siteDictionaries,
}: {
	base: string;
	sources: Sources;
	dictionaries?: ReadonlyMap<string, Dictionary>;
}): SiteFile[] {
	const origin = parseBase(base);
	const texts = readTexts(sources, fallbackLocale);
	for (const locale of texts.locales) {
		checkWords(dictionaries, locale);
	}
	const opened = new Map<string, Words<SiteKey>>();
	const site: Site = {
		texts,
		base,
		wordsOf(locale) {
			let words = opened.get(locale);
			if (words === undefined) {
				words = openWords<SiteKey>(locale, dictionaries);
				opened.set(locale, words);
			}
			return words;
		},
	};

	const files: SiteFile[] = [
		...texts.locales.flatMap((locale) => pagesOf(site, locale)),
		{ path: "index.html", data: apex(site) },
		{ path: "sitemap.xml", data: sitemap(texts, base, fallbackLocale) },
		{ path: "robots.txt", data: robots(base) },
		{ path: "CNAME", data: cname(origin) },
		// The host builds nothing of its own from a site that carries this
		// file; it costs one empty file and removes the question.
		{ path: ".nojekyll", data: "" },
	];
	return files.sort((a, b) => byCodeUnits(a.path, b.path));
}

// Site is what every page of a build is drawn from: the texts, the origin they
// are published on, and the words of each locale, which every locale with
// texts has, opened once for the whole build.
type Site = {
	readonly texts: Texts;
	readonly base: string;
	readonly wordsOf: (locale: string) => Words<SiteKey>;
};

// pagesOf draws every page a locale has.
function pagesOf(site: Site, locale: string): SiteFile[] {
	const { texts, base } = site;
	const words = site.wordsOf(locale);
	const home = address(locale, frontPage);
	return [...(texts.pages.get(locale) ?? [])].map(([name, text]) => {
		const served = address(locale, name);
		return {
			path: outputPath(served),
			data: page(
				words,
				<DocumentPage
					head={{
						lang: locale,
						dir: words.dir,
						title: text.title,
						description: text.description,
						canonical: base + served,
						alternates: alternatesOf(texts, base, fallbackLocale, name),
					}}
					home={home}
					languages={localesWith(texts, name).map((other) => ({
						locale: other,
						href: address(other, name),
						name: site.wordsOf(other).text("language.name"),
						current: other === locale,
					}))}
					documents={documentsOf(texts, locale)}
					translated={locale !== fallbackLocale}
					html={text.html}
				/>,
			),
		};
	});
}

// apex draws the page the bare domain serves, in the reference locale: its
// title and its one line are the reference front page's.
function apex(site: Site): string {
	const { texts, base } = site;
	const words = site.wordsOf(fallbackLocale);
	return page(
		words,
		<ApexPage
			head={{
				lang: fallbackLocale,
				dir: words.dir,
				title: texts.front.title,
				description: texts.front.description,
				canonical: `${base}/`,
				alternates: alternatesOf(texts, base, fallbackLocale, frontPage),
			}}
			choices={localesWith(texts, frontPage).map((locale) => {
				const own = site.wordsOf(locale);
				return {
					locale,
					href: address(locale, frontPage),
					name: own.text("language.name"),
					dir: own.dir,
				};
			})}
			documents={documentsOf(texts, fallbackLocale)}
		/>,
	);
}

// checkWords refuses a locale whose words say other things than the English
// ones — a key missing or extra, another slot, a text left empty: a page drawn
// from them would say a sentence in English, or a placeholder as it is
// written, and nothing after the build would notice.
function checkWords(
	dictionaries: ReadonlyMap<string, Dictionary>,
	locale: string,
): void {
	const words = dictionaries.get(locale);
	if (words === undefined) {
		throw new Error(
			`locale "${locale}" has texts and no words: the site's dictionary ${locale}.json is missing`,
		);
	}
	const found = disagreements(
		dictionaries.get(fallbackLocale) ?? {},
		words,
		locale,
	);
	if (found.length > 0) {
		throw new Error(
			`the site's words in ${locale} disagree with the English: ${found.join("; ")}`,
		);
	}
}

// documentsOf are the pages a locale has besides its front page, in the order
// of their names, each under its own title.
function documentsOf(texts: Texts, locale: string): DocumentLink[] {
	const pages = texts.pages.get(locale);
	return texts.names.flatMap((name) => {
		const text = pages?.get(name);
		return name === frontPage || text === undefined
			? []
			: [{ href: address(locale, name), label: text.title }];
	});
}

// page is a whole HTML document: the page drawn in its language's words,
// after the doctype that keeps a browser out of its quirks mode.
function page(words: Words<SiteKey>, content: VNode): string {
	return `<!DOCTYPE html>\n${renderToString(
		<SiteWords.Provider value={words}>{content}</SiteWords.Provider>,
	)}\n`;
}
