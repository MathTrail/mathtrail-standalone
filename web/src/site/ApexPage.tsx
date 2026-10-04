import { siteName } from "./brand";
import type { FooterLink } from "./Footer";
import type { Head } from "./Layout";
import { SitePage } from "./SitePage";
import { useSiteWords } from "./words";

/**
 * Choice is one language the apex offers: where its front page is, and the
 * language's own name for itself, written the way it runs.
 */
export type Choice = {
	readonly locale: string;
	readonly href: string;
	readonly name: string;
	readonly dir: "ltr" | "rtl";
};

/**
 * ApexPage is the page the bare domain serves. It is a doorway rather than a
 * translation: it names the product, says in one line what it is, and hands
 * the reader the languages it exists in, each in its own name, with no menu
 * and no switch of its own. A redirect would be cheaper, but the apex is the
 * address the product is listed under, and a redirect is a poor thing to list.
 */
export function ApexPage({
	head,
	choices,
	footer,
}: {
	head: Head;
	choices: readonly Choice[];
	footer: readonly FooterLink[];
}) {
	const words = useSiteWords();
	return (
		<SitePage
			head={head}
			frame={{ home: "/", menu: [], languages: [], footer }}
		>
			<div class="s-wrap s-apex">
				<h1>{siteName}</h1>
				<p class="s-lead">{head.description}</p>
				<nav class="s-choices" aria-label={words.text("nav.language")}>
					{choices.map(({ locale, href, name, dir }) => (
						<a
							key={locale}
							class="s-btn"
							href={href}
							hreflang={locale}
							lang={locale}
							dir={dir}
						>
							{name}
						</a>
					))}
				</nav>
			</div>
		</SitePage>
	);
}
