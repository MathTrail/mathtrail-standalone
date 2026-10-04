import { Icon } from "../design/icons";
import { markPath, siteName } from "./brand";
import { useSiteWords } from "./words";

/**
 * MenuLink is one entry of the menu as a page draws it: where it leads, under
 * which word, and whether it is the page being read.
 */
export type MenuLink = {
	readonly href: string;
	readonly label: string;
	readonly current: boolean;
};

/**
 * ActionLink is the one thing the header asks a reader to do: where it leads,
 * and the words of its button.
 */
export type ActionLink = { readonly href: string; readonly label: string };

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
 * the front page of the reader's language; the site's menu, when it has one;
 * the languages the page is offered in, when there is a choice to offer; and
 * the one thing the header asks a reader to do, as a button, when it asks
 * one. On a narrow screen the menu folds behind a button that opens it with no
 * script: the same links, drawn a second time inside a disclosure, which the
 * stylesheet shows in the menu's place.
 */
export function Header({
	home,
	menu = [],
	action,
	languages = [],
}: {
	home: string;
	menu?: readonly MenuLink[];
	action?: ActionLink;
	languages?: readonly LanguageLink[];
}) {
	const words = useSiteWords();
	const hasMenu = menu.length > 0;
	return (
		<header class="s-nav">
			<div class="s-wrap s-nav-inner">
				<Brand home={home} />
				{hasMenu && <Menu class="s-navlinks" links={menu} />}
				{(hasMenu || languages.length > 0 || action !== undefined) && (
					<div class="s-nav-tools">
						{languages.length > 0 && <LanguageSwitch languages={languages} />}
						{action !== undefined && (
							<a class="s-btn s-btn-filled s-nav-action" href={action.href}>
								{action.label}
							</a>
						)}
						{hasMenu && (
							<details class="s-menu">
								<summary>
									<Icon name="menu" size={24} />
									<span class="s-hidden">{words.text("nav.menu")}</span>
								</summary>
								<Menu class="s-menu-list" links={menu} />
							</details>
						)}
					</div>
				)}
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
	const size = small ? 24 : 28;
	return (
		<a class={small ? "s-brand s-brand-sm" : "s-brand"} href={home}>
			<img src={markPath} alt="" width={size} height={size} />
			<span>{siteName}</span>
		</a>
	);
}

// Menu is the menu's links, the page being read marked, wherever the header
// draws them.
function Menu({
	class: name,
	links,
}: {
	class: string;
	links: readonly MenuLink[];
}) {
	const words = useSiteWords();
	return (
		<nav class={name} aria-label={words.text("nav.pages")}>
			{links.map(({ href, label, current }) => (
				<a key={href} href={href} aria-current={current ? "page" : undefined}>
					{label}
				</a>
			))}
		</nav>
	);
}

// LanguageSwitch offers the page in each language of the site, by the
// language's code, with its own name for itself on hover.
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
