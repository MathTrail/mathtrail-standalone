import type { SiteKey } from "./words";

/**
 * MenuItem is one entry of the header's menu: a page of the site, or a
 * section of one, under the site's word for it.
 */
export type MenuItem = {
	/** page is the page's name, the same in every language: topics. */
	readonly page: string;
	/** anchor names a section of the page, such as the home page's connect. */
	readonly anchor?: string;
	readonly label: SiteKey;
};

/**
 * Frame is what every page of the site is set in besides its own words: the
 * header's menu, and the pages the footer leads to before the code and the
 * address that answers questions, each under its own title.
 */
export type Frame = {
	readonly menu: readonly MenuItem[];
	readonly footer: readonly string[];
};

/**
 * siteFrame is the site's own. A page joins the menu with the task that
 * publishes it, and the build refuses an entry whose page is not there, so the
 * menu never leads nowhere. The footer names the documents rather than every
 * page in turn, which seventeen topics would bury them under.
 */
export const siteFrame: Frame = {
	menu: [
		{ page: "why", label: "nav.why" },
		{ page: "topics", label: "nav.topics" },
	],
	footer: ["privacy", "terms"],
};
