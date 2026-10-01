import { markPath, siteName } from "./brand";
import { useSiteWords } from "./words";

/**
 * LanguageLink is one language a page is offered in: where the page is in it,
 * and the language's own name for itself.
 */
export type LanguageLink = {
	readonly locale: string;
	readonly href: string;
	readonly name: string;
	readonly current: boolean;
};

/**
 * Header is the bar on top of a page: the product's mark and name, leading to
 * the front page of the reader's language, and the languages this page is
 * offered in, when there is a choice to offer.
 */
export function Header({
	home,
	languages = [],
}: {
	home: string;
	languages?: readonly LanguageLink[];
}) {
	return (
		<header class="s-nav">
			<div class="s-wrap s-nav-inner">
				<Brand home={home} />
				{languages.length > 0 && <LanguageSwitch languages={languages} />}
			</div>
		</header>
	);
}

/** Brand is the product's mark beside its name, leading home. */
export function Brand({
	home,
	small = false,
}: {
	home: string;
	small?: boolean;
}) {
	const size = small ? 20 : 28;
	return (
		<a class={small ? "s-brand s-brand-sm" : "s-brand"} href={home}>
			<img src={markPath} alt="" width={size} height={size} />
			<span>{siteName}</span>
		</a>
	);
}

// LanguageSwitch offers the page in each language it exists in, by the
// language's code, with its own name for itself on hover. A language that
// lacks the page is not offered: a switch that leads to a missing page is
// worse than one language fewer.
function LanguageSwitch({ languages }: { languages: readonly LanguageLink[] }) {
	const words = useSiteWords();
	return (
		<nav class="s-seg" aria-label={words.text("nav.language")}>
			{languages.map(({ locale, href, name, current }) => (
				<a
					key={locale}
					href={href}
					hreflang={locale}
					lang={locale}
					title={name}
					aria-current={current ? "page" : undefined}
				>
					{shortName(locale)}
				</a>
			))}
		</nav>
	);
}

// shortName is how the switch names a language: its subtag, as a code —
// EN, RU.
function shortName(locale: string): string {
	return (locale.split("-", 1)[0] ?? locale).toUpperCase();
}
