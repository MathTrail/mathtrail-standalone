import type { ComponentChildren } from "preact";
import type { Words } from "../i18n/words";
import type { Fill, PageReader } from "./reader";
import { doiAddress, type Source } from "./why";
import { type SiteKey, useSiteWords } from "./words";

/**
 * Honest is a note of the page "Why" that says plainly what the page does not
 * promise, under a shield.
 */
export function Honest({
	page,
	at,
	slots,
}: {
	page: PageReader;
	at: string;
	slots?: Readonly<Record<string, Fill>>;
}) {
	return (
		<aside class="s-note s-why-honest">
			<span class="s-why-badge">
				<Stroke size={22}>
					<path d="M12 3l7 3v6c0 4.5-3 7.5-7 9-4-1.5-7-4.5-7-9V6z" />
					<path d="M9 12l2 2 4-4" />
				</Stroke>
			</span>
			<div class="s-why-honest-text">
				<p class="s-note-title">{page.text("honest")}</p>
				<p class="s-note-text">{page.text(at, slots)}</p>
			</div>
		</aside>
	);
}

/**
 * Stroke is a mark of the page drawn in strokes of the colour around it on a
 * 24-unit square, as decoration beside words that say what it means.
 */
export function Stroke({
	size,
	children,
}: {
	size: number;
	children: ComponentChildren;
}) {
	return (
		<svg
			viewBox="0 0 24 24"
			width={size}
			height={size}
			fill="none"
			stroke="currentColor"
			stroke-width="1.8"
			stroke-linecap="round"
			stroke-linejoin="round"
			aria-hidden="true"
		>
			{children}
		</svg>
	);
}

/**
 * Cite is a work as a sentence cites it, its first author and its year, linked
 * to the work itself.
 */
export function Cite({ source }: { source: Source }) {
	const words = useSiteWords();
	return <a href={doiAddress(source)}>{citeText(words, source)}</a>;
}

/**
 * citeText is how words cite source: by its one or two authors, or by the
 * first of more, and its year, as the language writes a citation.
 */
export function citeText(words: Words<SiteKey>, source: Source): string {
	const [first = "", second = ""] = source.authors.map(familyOf);
	const year = String(source.year);
	switch (source.authors.length) {
		case 1:
			return words.text("cite.one", { first, year });
		case 2:
			return words.text("cite.two", { first, second, year });
		default:
			return words.text("cite.many", { first, year });
	}
}

// familyOf is the family name of an author written "Family, I.".
function familyOf(author: string): string {
	return author.split(",", 1)[0] ?? author;
}
