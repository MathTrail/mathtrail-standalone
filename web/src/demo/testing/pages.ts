import { readFileSync } from "node:fs";
import { join } from "node:path";
import { Window } from "happy-dom";
import topics from "../../../../content/catalogs/topics.json";
import traps from "../../../../content/catalogs/traps.json";
import file from "../../../../site/data.json";
import { address, outputPath } from "../../site/addresses";
import { frontPage } from "../../site/content";
import { readSiteData } from "../../site/data";
import type { Frame } from "../../site/frame";
import { connectSection } from "../../site/home";
import { type Page, sitePages } from "../../site/pages";
import { renderSite } from "../../site/render";

const repository = join(import.meta.dirname, "..", "..", "..", "..");

// why is the name of the page Why among the site's pages.
const why = "why";

// built are the pages built here, each by the file of its words.
const built = new Map([
	[frontPage, "index.yaml"],
	[why, "why.yaml"],
]);

// documentOf is a document's text under its title: the site builds no page
// without its privacy policy and its terms.
const documentOf = (title: string) =>
	`---\ntitle: ${title}\ndescription: D\n---\n\n# ${title}\n`;

// frame is the site's frame over the pages built here: the menu names pages,
// none of which is built here, and the header's button leads to connecting.
const frame: Frame = {
	menu: [],
	action: { page: frontPage, anchor: connectSection, label: "nav.add" },
	footer: ["privacy", "terms"],
};

// builtPage is the page called name in locale as the site builds it, the text
// of the whole page: the site's own words and data, no topic's page
// published. The page is built alone; the header's button leads to the home
// page, which a document stands in for when the page built is another.
function builtPage(name: string, locale: string): string {
	const data = readSiteData(
		{
			topics: topics.map((topic) => ({ ...topic, site_page: false })),
			traps,
			tasks: [],
		},
		{
			groups: file.groups,
			examples: {},
			progress: file.progress,
			home: file.home,
			why: file.why,
		},
	);
	const drawn = sitePages(data).get(name);
	const words = built.get(name);
	if (drawn === undefined || words === undefined) {
		throw new Error(`the site's pages have no page ${name}`);
	}
	const home: [string, string][] =
		name === frontPage ? [] : [["index.md", documentOf("Home")]];
	const sources = new Map(
		["en", "ru"].map((language) => [
			language,
			new Map([
				[
					words,
					readFileSync(
						join(repository, "site", "content", language, words),
						"utf8",
					),
				],
				...home,
				["privacy.md", documentOf("Privacy")],
				["terms.md", documentOf("Terms")],
			]),
		]),
	);
	const page = renderSite({
		base: "https://example.test",
		sources,
		pages: new Map<string, Page>([[name, drawn]]),
		frame,
		data,
	}).find(({ path }) => path === outputPath(address(locale, name)));
	if (page === undefined) {
		throw new Error(`the site has no page ${name} in ${locale}`);
	}
	return page.data;
}

// openPage puts the body of the page called name in locale into the test's
// document, as a browser holds it once it has read the page and before any
// script has run. The page is read apart, by a browser that loads none of the
// files it names, and its tag for its script is left out: the test runs the
// script from its source.
function openPage(name: string, locale: string): void {
	const reader = new Window({
		settings: {
			disableCSSFileLoading: true,
			disableJavaScriptFileLoading: true,
			handleDisabledFileLoadingAsSuccess: true,
		},
	});
	const page = new reader.DOMParser().parseFromString(
		builtPage(name, locale),
		"text/html",
	);
	page.querySelector('script[type="module"]')?.remove();
	document.documentElement.lang = page.documentElement.lang;
	document.body.innerHTML = page.body.innerHTML;
	void reader.happyDOM.close();
}

/**
 * openHome puts the body of the home page in locale into the test's document,
 * as a browser holds it before the demo has run.
 */
export function openHome(locale = "en"): void {
	openPage(frontPage, locale);
}

/**
 * openWhy puts the body of the page Why in locale into the test's document,
 * as a browser holds it before the pictures of the history have started.
 */
export function openWhy(locale = "en"): void {
	openPage(why, locale);
}
