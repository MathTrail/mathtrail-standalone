import { Icon } from "../design/icons";
import { markPath, siteName } from "./brand";
import { useSiteWords } from "./words";

/**
 * MenuLink is one entry of the menu as a page draws it: where it leads, under
 * which word, how it stands to the page being read — that page itself, the
 * page that one lies under, as the page of the topics is to a topic's own, or
 * neither — and whether its page is still in the making.
 */
export type MenuLink = {
	readonly href: string;
	readonly label: string;
	readonly current: "page" | "parent" | undefined;
	readonly wip: boolean;
};

/**
 * MenuLinkGroup is entries of the menu a page draws held together, under the
 * word a screen reader says for them.
 */
export type MenuLinkGroup = {
	readonly label: string;
	readonly links: readonly MenuLink[];
};

/** MenuEntry is one entry of the menu as a page draws it, or a group of them. */
export type MenuEntry = MenuLink | MenuLinkGroup;

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

// navDisclosures names the header's disclosures as one group, the language
// picker and the folded menu, so that a browser closes the one it has open
// when the other opens.
const navDisclosures = "s-nav";

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
	menu?: readonly MenuEntry[];
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
						{languages.length > 0 && <LanguagePicker languages={languages} />}
						{action !== undefined && (
							<a class="s-btn s-btn-filled s-nav-action" href={action.href}>
								{action.label}
							</a>
						)}
						{hasMenu && (
							<details class="s-menu" name={navDisclosures}>
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

// Menu is the menu's links, the page being read marked and a group's links
// held together, wherever the header draws them.
function Menu({
	class: name,
	links,
}: {
	class: string;
	links: readonly MenuEntry[];
}) {
	const words = useSiteWords();
	return (
		<nav class={name} aria-label={words.text("nav.pages")}>
			{links.map((entry) =>
				"links" in entry ? (
					// biome-ignore lint/a11y/useSemanticElements: a fieldset groups a form's controls and draws its legend above them, and these are links in the menu's row
					<span
						key={entry.label}
						class="s-nav-group"
						role="group"
						aria-label={entry.label}
					>
						{entry.links.map((link) => (
							<MenuAnchor key={link.href} link={link} />
						))}
					</span>
				) : (
					<MenuAnchor key={entry.href} link={entry} />
				),
			)}
		</nav>
	);
}

// MenuAnchor is one link of the menu, marked when it is the page being read,
// and marked as the menu's current entry, not as the page, when the page
// being read lies under it. The link to a page still in the making says so in
// a tag beside its word, a word of its own in the link's text, which a screen
// reader reads with it.
function MenuAnchor({ link }: { link: MenuLink }) {
	const words = useSiteWords();
	const { href, label, current, wip } = link;
	return (
		<a
			href={href}
			class={wip ? "s-nav-wip" : undefined}
			aria-current={current === "parent" ? "true" : current}
		>
			{label}
			{wip && (
				<>
					{" "}
					<span class="s-nav-tag">{words.text("nav.beta")}</span>
				</>
			)}
		</a>
	);
}

// LanguagePicker offers the page in each language of the site: a pill that
// names the language being read by its code, and under it, opened with no
// script, the languages by their own names, the one being read ticked.
function LanguagePicker({ languages }: { languages: readonly LanguageLink[] }) {
	const words = useSiteWords();
	const read = languages.find((language) => language.current);
	return (
		<details class="s-lang" name={navDisclosures}>
			<summary class="s-lang-pill">
				<Icon name="globe" size={17} />
				<span class="s-hidden">{words.text("nav.language")} </span>
				{read === undefined ? undefined : shortName(read.locale)}
				<Icon name="chevron-down" size={14} className="s-lang-chevron" />
			</summary>
			<nav class="s-lang-list" aria-label={words.text("nav.language")}>
				{inPickerOrder(languages).map(({ locale, href, name, current }) => (
					<a
						key={locale}
						href={href}
						hreflang={locale}
						lang={locale}
						aria-current={current ? "page" : undefined}
					>
						<span dir="auto">{name}</span>
						{current && <Icon name="check" size={14} className="s-lang-tick" />}
					</a>
				))}
			</nav>
		</details>
	);
}

// letters orders the languages' names as the letters of every script are
// ordered: the Latin ones first, then the Cyrillic, the Arabic, the Devanagari
// and the Han, and each script by its own letters.
const letters = new Intl.Collator("en");

/**
 * inPickerOrder are the languages in the order the picker lists them: by
 * their own names, in one order on every page, so that a reader finds a
 * language where they found it before.
 */
export function inPickerOrder(
	languages: readonly LanguageLink[],
): LanguageLink[] {
	return [...languages].sort((one, other) =>
		letters.compare(one.name, other.name),
	);
}

// shortName is how the picker names the language being read: its subtag, as
// a code — EN, RU.
function shortName(locale: string): string {
	return (locale.split("-", 1)[0] ?? locale).toUpperCase();
}
