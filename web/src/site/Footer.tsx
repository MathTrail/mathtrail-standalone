import { Brand } from "./Header";
import { useSiteWords } from "./words";

/** sourceURL is where the product's code and content are published. */
const sourceURL = "https://github.com/MathTrail/mathtrail-standalone";

/** DocumentLink is one page a footer leads to, under the page's own title. */
export type DocumentLink = { readonly href: string; readonly label: string };

/**
 * Footer closes a page: the documents of its language and the code, and — on
 * a page that is a translation — that the English text is the one that holds
 * where the two differ.
 */
export function Footer({
	home,
	documents,
	translated,
}: {
	home: string;
	documents: readonly DocumentLink[];
	translated: boolean;
}) {
	const words = useSiteWords();
	return (
		<footer class="s-footer">
			<div class="s-wrap">
				<div class="s-foot">
					<Brand home={home} small />
					<nav class="s-footlinks" aria-label={words.text("footer.documents")}>
						{documents.map(({ href, label }) => (
							<a key={href} href={href}>
								{label}
							</a>
						))}
						<a href={sourceURL}>{words.text("footer.source")}</a>
					</nav>
				</div>
				{translated && (
					<p class="s-legal">{words.text("footer.translation")}</p>
				)}
			</div>
		</footer>
	);
}
