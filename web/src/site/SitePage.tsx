import type { ComponentChildren } from "preact";
import { Footer, type FooterLink } from "./Footer";
import { Header, type LanguageLink, type MenuLink } from "./Header";
import { type Head, Layout } from "./Layout";

/** PageFrame is what one page is set in besides its own words, as it draws it. */
export type PageFrame = {
	/** home is the front page the mark leads to. */
	readonly home: string;
	readonly menu: readonly MenuLink[];
	readonly languages: readonly LanguageLink[];
	readonly footer: readonly FooterLink[];
	/** translated is true for a page in any language but English. */
	readonly translated: boolean;
};

/**
 * SitePage is a whole page of the site: its head, the header, the page's own
 * part as the main one, and the footer. A document and a page a component
 * draws are set in the same frame; a page that draws a card of the widget
 * says so, and loads the card's stylesheet, and a page may carry rules of
 * style of its own in its head.
 */
export function SitePage({
	head,
	frame,
	card = false,
	style,
	children,
}: {
	head: Head;
	frame: PageFrame;
	card?: boolean;
	style?: string;
	children: ComponentChildren;
}) {
	return (
		<Layout head={head} card={card} style={style}>
			<Header home={frame.home} menu={frame.menu} languages={frame.languages} />
			<main class="s-main">{children}</main>
			<Footer
				home={frame.home}
				links={frame.footer}
				translated={frame.translated}
			/>
		</Layout>
	);
}
