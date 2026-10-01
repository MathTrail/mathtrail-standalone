import { type DocumentLink, Footer } from "./Footer";
import { Header, type LanguageLink } from "./Header";
import { type Head, Layout } from "./Layout";

/**
 * DocumentPage is a page that is read: a locale's text — its front page, its
 * privacy policy, its terms — between the header and the footer. The text
 * arrives as HTML made from the site's own Markdown, which carries no markup
 * of its own.
 */
export function DocumentPage({
	head,
	home,
	languages,
	documents,
	translated,
	html,
}: {
	head: Head;
	home: string;
	languages: readonly LanguageLink[];
	documents: readonly DocumentLink[];
	translated: boolean;
	html: string;
}) {
	return (
		<Layout head={head}>
			<Header home={home} languages={languages} />
			<main class="s-main">
				<div class="s-wrap">
					<article class="s-doc" dangerouslySetInnerHTML={{ __html: html }} />
				</div>
			</main>
			<Footer home={home} documents={documents} translated={translated} />
		</Layout>
	);
}
