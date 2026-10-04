import type { VNode } from "preact";
import type { PageReader } from "./reader";

/** PageProps are what a page's component draws: its words, in the page's language. */
export type PageProps = { readonly page: PageReader };

/**
 * PageComponent draws the part of one page that is its own, between the
 * site's header and its footer.
 */
export type PageComponent = (props: PageProps) => VNode;

/**
 * pageComponents are the pages a component draws, by the name their words'
 * file has in every language — topics for topics.yaml. A file of words that no
 * component draws stops the build.
 */
export const pageComponents: ReadonlyMap<string, PageComponent> = new Map();
