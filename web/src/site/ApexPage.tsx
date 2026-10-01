import { siteName } from "./brand";
import { type DocumentLink, Footer } from "./Footer";
import { Header } from "./Header";
import { type Head, Layout } from "./Layout";
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
 * the reader the languages it exists in, each in its own name. A redirect
 * would be cheaper, but the apex is the address the product is listed under,
 * and a redirect is a poor thing to list.
 */
export function ApexPage({
	head,
	choices,
	documents,
}: {
	head: Head;
	choices: readonly Choice[];
	documents: readonly DocumentLink[];
}) {
	const words = useSiteWords();
	return (
		<Layout head={head}>
			<Header home="/" />
			<main class="s-main">
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
			</main>
			<Footer home="/" documents={documents} translated={false} />
		</Layout>
	);
}
