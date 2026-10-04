import { Brand } from "./Header";
import { useSiteWords } from "./words";

/** sourceURL is where the product's code and content are published. */
export const sourceURL = "https://github.com/MathTrail/mathtrail-standalone";

/**
 * contactAddress is where questions about the product and a child's data go:
 * the address the privacy policy names.
 */
export const contactAddress = "altedtech.info@gmail.com";

/** FooterLink is one page a footer leads to, under the page's own title. */
export type FooterLink = { readonly href: string; readonly label: string };

/**
 * Footer closes a page: the pages the site's frame names, the code, and the
 * address that answers questions, written out so that it can be copied with
 * no mail program at hand. It says nothing of which language's text holds:
 * the policy and the terms say that of themselves.
 */
export function Footer({
	home,
	links,
}: {
	home: string;
	links: readonly FooterLink[];
}) {
	const words = useSiteWords();
	return (
		<footer class="s-footer">
			<div class="s-wrap">
				<div class="s-foot">
					<Brand home={home} small />
					<nav class="s-footlinks" aria-label={words.text("footer.links")}>
						{links.map(({ href, label }) => (
							<a key={href} href={href}>
								{label}
							</a>
						))}
						<a href={sourceURL}>{words.text("footer.source")}</a>
						<a href={`mailto:${contactAddress}`}>{contactAddress}</a>
					</nav>
				</div>
			</div>
		</footer>
	);
}
