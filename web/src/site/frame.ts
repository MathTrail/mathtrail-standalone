import { frontPage } from "./content";
import { connectSection } from "./home";
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
	/** wip marks a page still in the making, which the menu says it is. */
	readonly wip?: boolean;
};

/**
 * MenuGroup is entries of the menu held together under a word a screen reader
 * says, such as the technical pages, which are not written for parents.
 */
export type MenuGroup = {
	readonly label: SiteKey;
	readonly items: readonly MenuItem[];
};

/**
 * Frame is what every page of the site is set in besides its own words: the
 * header's menu, its entries one by one or held together in a group, and the
 * one thing the header asks a reader to do, when it asks one, and the pages
 * the footer leads to before the code and the address that answers
 * questions, each under its own title.
 */
export type Frame = {
	readonly menu: readonly (MenuItem | MenuGroup)[];
	/** action is a button rather than an entry: no tag says it is in beta. */
	readonly action?: Omit<MenuItem, "wip">;
	readonly footer: readonly string[];
};

/** menuItems are the menu's entries, those of its groups among them, in order. */
export function menuItems(menu: Frame["menu"]): MenuItem[] {
	return menu.flatMap((entry) => ("items" in entry ? entry.items : [entry]));
}

/**
 * siteFrame is the site's own. A page joins the menu with the task that
 * publishes it, and the build refuses an entry whose page is not there, so the
 * menu never leads nowhere. The menu names pages, never the home page's
 * sections, which the home page leads through itself: the page of the coach,
 * a product of its own still in the making and marked so, opens it; the
 * technical pages, the numbers behind the product and how the service works,
 * held together as pages not written for parents, and the page about who
 * makes it close it. The header asks a reader to add MathTrail to Claude,
 * which the home page's section on connecting tells how to do. The footer
 * names the page of the numbers, the page about who makes the product, the
 * help and the documents rather than every page in turn, which seventeen
 * topics would bury them under.
 */
export const siteFrame: Frame = {
	menu: [
		{ page: "coach", label: "nav.coach", wip: true },
		{ page: "why", label: "nav.why" },
		{ page: "topics", label: "nav.topics" },
		{ page: "techniques", label: "nav.techniques" },
		{
			label: "nav.technical",
			items: [
				{ page: "research", label: "nav.research" },
				{ page: "service", label: "nav.service" },
			],
		},
		{ page: "about", label: "nav.about" },
	],
	action: { page: frontPage, anchor: connectSection, label: "nav.add" },
	footer: ["research", "about", "help", "privacy", "terms"],
};
