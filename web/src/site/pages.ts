import type { VNode } from "preact";
import type { SiteData } from "./data";
import type { PageReader } from "./reader";
import { TopicsPage, topicsStyle } from "./TopicsPage";

/**
 * PageProps are what a page's component draws: its words, in the page's
 * language, and the site's data, which no language changes.
 */
export type PageProps = { readonly page: PageReader; readonly data: SiteData };

/**
 * Page is a page a component draws: the component, which draws the part of the
 * page that is its own, between the site's header and its footer; whether it
 * draws a card of the widget, which loads the card's stylesheet; and the rules
 * of style its head carries, written from the site's data when the site is
 * built.
 */
export type Page = {
	readonly draw: (props: PageProps) => VNode;
	readonly card?: boolean;
	readonly style?: (data: SiteData) => string;
};

/**
 * sitePages are the pages a component draws, by the name their words' file
 * has in every language — topics for topics.yaml. A file of words that no
 * component draws stops the build.
 */
export const sitePages: ReadonlyMap<string, Page> = new Map([
	["topics", { draw: TopicsPage, card: true, style: topicsStyle }],
]);
