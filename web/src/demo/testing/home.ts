import { readFileSync } from "node:fs";
import { join } from "node:path";
import { Window } from "happy-dom";
import topics from "../../../../content/catalogs/topics.json";
import traps from "../../../../content/catalogs/traps.json";
import file from "../../../../site/data.json";
import { frontPage } from "../../site/content";
import { readSiteData } from "../../site/data";
import type { Frame } from "../../site/frame";
import { connectSection } from "../../site/home";
import { type Page, sitePages } from "../../site/pages";
import { renderSite } from "../../site/render";

const repository = join(import.meta.dirname, "..", "..", "..", "..");

// documentOf is a document's text under its title: the site builds no page
// without its privacy policy and its terms.
const documentOf = (title: string) =>
	`---\ntitle: ${title}\ndescription: D\n---\n\n# ${title}\n`;

// frame is the site's frame over the one page built here: the menu names
// pages, none of which is built here, and the header's button leads to
// connecting.
const frame: Frame = {
	menu: [],
	action: { page: frontPage, anchor: connectSection, label: "nav.add" },
	footer: ["privacy", "terms"],
};

// builtHome is the home page in locale as the site builds it, the text of the
// whole page: the site's own words and data, no topic's page published.
function builtHome(locale: string): string {
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
		},
	);
	const home = sitePages(data).get(frontPage);
	if (home === undefined) {
		throw new Error("the site's pages have no home page");
	}
	const sources = new Map(
		["en", "ru"].map((language) => [
			language,
			new Map([
				[
					"index.yaml",
					readFileSync(
						join(repository, "site", "content", language, "index.yaml"),
						"utf8",
					),
				],
				["privacy.md", documentOf("Privacy")],
				["terms.md", documentOf("Terms")],
			]),
		]),
	);
	const built = renderSite({
		base: "https://example.test",
		sources,
		pages: new Map<string, Page>([[frontPage, home]]),
		frame,
		data,
	}).find(({ path }) => path === `${locale}/index.html`);
	if (built === undefined) {
		throw new Error(`the site has no home page in ${locale}`);
	}
	return built.data;
}

/**
 * openHome puts the body of the home page in locale into the test's document,
 * as a browser holds it once it has read the page and before any script has
 * run. The page is read apart, by a browser that loads none of the files it
 * names, and its tag for the demo is left out: the test runs the demo from
 * its source.
 */
export function openHome(locale = "en"): void {
	const reader = new Window({
		settings: {
			disableCSSFileLoading: true,
			disableJavaScriptFileLoading: true,
			handleDisabledFileLoadingAsSuccess: true,
		},
	});
	const page = new reader.DOMParser().parseFromString(
		builtHome(locale),
		"text/html",
	);
	page.querySelector('script[type="module"]')?.remove();
	document.documentElement.lang = page.documentElement.lang;
	document.body.innerHTML = page.body.innerHTML;
	void reader.happyDOM.close();
}
