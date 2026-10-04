import type { VNode } from "preact";
import { renderToString } from "preact-render-to-string";
import { disagreements } from "../i18n/dictionaries";
import { fallbackLocale } from "../i18n/lookup";
import { byCodeUnits } from "../i18n/order";
import { type Dictionary, openWords, type Words } from "../i18n/words";
import { ApexPage } from "./ApexPage";
import { address, alternatesOf, outputPath, parseBase } from "./addresses";
import {
	frontPage,
	readTexts,
	type Sources,
	type Summary,
	summaryOf,
	type Texts,
	textOf,
} from "./content";
import { DocumentPage } from "./DocumentPage";
import { type SiteData, siteData } from "./data";
import type { FooterLink } from "./Footer";
import { type Frame, siteFrame } from "./frame";
import type { Head } from "./Layout";
import { cname, robots, sitemap } from "./metadata";
import { type Page, sitePages } from "./pages";
import { openReader } from "./reader";
import { type PageFrame, SitePage } from "./SitePage";
import { type SiteKey, SiteWords, siteDictionaries } from "./words";

/**
 * SiteFile is one file of the built site: where it goes below the site's root,
 * and what it says.
 */
export type SiteFile = { readonly path: string; readonly data: string };

/**
 * renderSite draws every file the site's texts and words make: every page of
 * every locale, the apex, and the files a crawler and the host read. It writes
 * nothing. The files come back sorted by path, so that two builds of the same
 * texts are the same bytes, and anything wrong with a text, a dictionary, the
 * frame, a page's component or the site's data stops the build before a
 * single file is written. The stylesheets and the mark are not among them:
 * they are built and copied beside these.
 *
 * base is the origin the site is published on. The dictionaries, the pages a
 * component draws, the frame and the data are the site's own unless a test
 * hands it others.
 */
export function renderSite({
	base,
	sources,
	dictionaries = siteDictionaries,
	pages = sitePages,
	frame = siteFrame,
	data = siteData(),
}: {
	base: string;
	sources: Sources;
	dictionaries?: ReadonlyMap<string, Dictionary>;
	pages?: ReadonlyMap<string, Page>;
	frame?: Frame;
	data?: SiteData;
}): SiteFile[] {
	const origin = parseBase(base);
	const texts = readTexts(sources, fallbackLocale);
	checkLanguages(dictionaries, texts.locales);
	checkFrame(frame, texts.names);
	checkTopicPages(data, texts.names);
	const opened = new Map<string, Words<SiteKey>>();
	const site: Site = {
		texts,
		base,
		frame,
		pages,
		data,
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
		...texts.locales.flatMap((locale) =>
			texts.names.map((name) => ({
				path: outputPath(address(locale, name)),
				data: drawPage(site, locale, name),
			})),
		),
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
// are published on, the frame, the pages a component draws, the site's data,
// and the words of each locale, which every locale with texts has, opened
// once for the whole build.
type Site = {
	readonly texts: Texts;
	readonly base: string;
	readonly frame: Frame;
	readonly pages: ReadonlyMap<string, Page>;
	readonly data: SiteData;
	readonly wordsOf: (locale: string) => Words<SiteKey>;
};

// drawPage draws the page name of a locale: a document's HTML set in the
// frame, or the page's component, which reads the page's words. A component
// that leaves a text of its words unshown stops the build: the text would be
// translated and never read.
function drawPage(site: Site, locale: string, name: string): string {
	const text = textOf(site.texts, locale, name);
	const words = site.wordsOf(locale);
	const frame = frameOf(site, locale, name);
	if (text.kind === "document") {
		return page(
			words,
			<DocumentPage
				head={headOf(site, locale, name, text)}
				frame={frame}
				html={text.html}
			/>,
		);
	}
	const file = `${locale}/${name}.yaml`;
	const { page: reader, unread } = openReader(text.words, file, locale);
	const { draw: Draw, card, style } = pageOf(site.pages, name);
	const drawn = page(
		words,
		<SitePage
			head={headOf(site, locale, name, {
				title: reader.plain("title"),
				description: reader.plain("description"),
			})}
			frame={frame}
			card={card}
			style={style?.(site.data)}
		>
			<Draw page={reader} data={site.data} />
		</SitePage>,
	);
	const unshown = unread();
	if (unshown.length > 0) {
		throw new Error(`${file}: the page never shows ${unshown.join(", ")}`);
	}
	return drawn;
}

// headOf is what the page name of a locale tells a browser, a search engine
// and a chat about itself.
function headOf(
	site: Site,
	locale: string,
	name: string,
	{ title, description }: Summary,
): Head {
	return {
		lang: locale,
		dir: site.wordsOf(locale).dir,
		title,
		description,
		canonical: site.base + address(locale, name),
		alternates: alternatesOf(site.texts, site.base, fallbackLocale, name),
	};
}

// frameOf is the frame the page name of a locale is set in: its menu, the
// page itself marked, every language of the site to switch to, and the pages
// the footer leads to.
function frameOf(site: Site, locale: string, name: string): PageFrame {
	const words = site.wordsOf(locale);
	return {
		home: address(locale, frontPage),
		menu: site.frame.menu.map(({ page, anchor, label }) => ({
			href:
				anchor === undefined
					? address(locale, page)
					: `${address(locale, page)}#${anchor}`,
			label: words.text(label),
			current: anchor === undefined && page === name,
		})),
		languages: site.texts.locales.map((other) => ({
			locale: other,
			href: address(other, name),
			name: site.wordsOf(other).text("language.name"),
			current: other === locale,
		})),
		footer: footerOf(site, locale),
		translated: locale !== fallbackLocale,
	};
}

// footerOf are the pages the footer of a locale's page leads to, each under
// its own title.
function footerOf(site: Site, locale: string): FooterLink[] {
	return site.frame.footer.map((name) => ({
		href: address(locale, name),
		label: summaryOf(textOf(site.texts, locale, name)).title,
	}));
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
			choices={texts.locales.map((locale) => {
				const own = site.wordsOf(locale);
				return {
					locale,
					href: address(locale, frontPage),
					name: own.text("language.name"),
					dir: own.dir,
				};
			})}
			footer={footerOf(site, fallbackLocale)}
		/>,
	);
}

// checkLanguages holds the site's languages to its dictionaries. A locale
// with texts has words that say what the English ones say, and a locale with
// words has texts: the languages of the site are those of its dictionaries,
// and the card links a topic's page in its own language when the site has it.
function checkLanguages(
	dictionaries: ReadonlyMap<string, Dictionary>,
	locales: readonly string[],
): void {
	for (const locale of locales) {
		checkWords(dictionaries, locale);
	}
	for (const tag of dictionaries.keys()) {
		if (!locales.includes(tag)) {
			throw new Error(
				`the site has words in ${tag} — its dictionary ${tag}.json — and no texts: every page is in every language the site speaks`,
			);
		}
	}
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

// checkFrame refuses a menu or a footer that leads to a page the site does not
// have: a page joins them with the task that publishes it.
function checkFrame(frame: Frame, names: readonly string[]): void {
	for (const { page } of frame.menu) {
		if (!names.includes(page)) {
			throw new Error(
				`the menu leads to the page ${page}, which the site does not have`,
			);
		}
	}
	for (const page of frame.footer) {
		if (!names.includes(page)) {
			throw new Error(
				`the footer leads to the page ${page}, which the site does not have`,
			);
		}
	}
}

// checkTopicPages holds the topics' pages to the catalog: a topic's page,
// topics/<slug>, is published when the catalog says so, and is there whenever
// it does, since the card links it from then on.
function checkTopicPages(data: SiteData, names: readonly string[]): void {
	const published = new Set(
		data.topics.all
			.filter((topic) => topic.sitePage)
			.map((topic) => `topics/${topic.slug}`),
	);
	for (const name of names) {
		if (name.startsWith("topics/") && !published.has(name)) {
			throw new Error(
				`the site has the page ${name}, and no topic of the catalog has its page published there`,
			);
		}
	}
	for (const name of published) {
		if (!names.includes(name)) {
			throw new Error(
				`the catalog has the page ${name} published, and the site does not have it`,
			);
		}
	}
}

// pageOf is the page a component draws by the name name. Words that no
// component draws stop the build: they would be translated into every language
// and published in none.
function pageOf(pages: ReadonlyMap<string, Page>, name: string): Page {
	const found = pages.get(name);
	if (found === undefined) {
		throw new Error(
			`the site has words for the page ${name}, and no component draws it`,
		);
	}
	return found;
}

// page is a whole HTML document: the page drawn in its language's words,
// after the doctype that keeps a browser out of its quirks mode.
function page(words: Words<SiteKey>, content: VNode): string {
	return `<!DOCTYPE html>\n${renderToString(
		<SiteWords.Provider value={words}>{content}</SiteWords.Provider>,
	)}\n`;
}
