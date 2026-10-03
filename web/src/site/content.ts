import { Marked } from "marked";
import { byCodeUnits } from "../i18n/order";

/**
 * Document is one text of the site: the fields a page's head is made of, and
 * the Markdown body that follows them.
 */
export type Document = {
	readonly title: string;
	readonly description: string;
	readonly body: string;
};

/**
 * PageText is one page's text, read: the fields of its head, and its body as
 * HTML.
 */
export type PageText = {
	readonly title: string;
	readonly description: string;
	readonly html: string;
};

/**
 * Sources are the site's texts as they are written: by locale, then by page
 * name, the whole of each file.
 */
export type Sources = ReadonlyMap<string, ReadonlyMap<string, string>>;

/**
 * Texts are every text the site has, read, and the two orders a build walks
 * them in: the locales, and the names of the pages any of them has, both
 * sorted, so that two builds of the same texts are the same files. front is
 * the reference locale's front page, which the apex is made of.
 */
export type Texts = {
	readonly locales: readonly string[];
	readonly names: readonly string[];
	readonly pages: ReadonlyMap<string, ReadonlyMap<string, PageText>>;
	readonly front: PageText;
};

/**
 * frontPage is the page name that becomes a locale's own front page rather
 * than a directory beneath it.
 */
export const frontPage = "index";

const fence = "---";

/**
 * parseDocument splits a text into its front matter and its body. The front
 * matter is a fenced block of "key: value" lines at the top of the file; it is
 * required, because a page with no title or description cannot be shown
 * correctly by a search engine or a chat.
 */
export function parseDocument(source: string): Document {
	if (!source.startsWith(`${fence}\n`)) {
		throw new Error(`front matter must open with "${fence}" on the first line`);
	}
	const rest = source.slice(fence.length + 1);
	const close = rest.indexOf(`\n${fence}\n`);
	if (close < 0) {
		throw new Error(
			`front matter is never closed by "${fence}" on a line of its own`,
		);
	}

	let title = "";
	let description = "";
	for (const line of rest.slice(0, close).split("\n")) {
		if (line.trim() === "") {
			continue;
		}
		const colon = line.indexOf(":");
		if (colon < 0) {
			throw new Error(
				`front matter line ${JSON.stringify(line)} is not key: value`,
			);
		}
		const key = line.slice(0, colon).trim();
		const value = line.slice(colon + 1).trim();
		switch (key) {
			case "title":
				title = value;
				break;
			case "description":
				description = value;
				break;
			default:
				throw new Error(
					`front matter key ${JSON.stringify(key)} is not one this site understands`,
				);
		}
	}

	if (title === "") {
		throw new Error("front matter has no title");
	}
	if (description === "") {
		throw new Error("front matter has no description");
	}
	return { title, description, body: rest.slice(close + fence.length + 2) };
}

// commonMark reads the texts as CommonMark alone, the Markdown they are
// written in: GitHub's additions — tables, struck-through text, bare addresses
// turned into links — would change what a text that happens to look like one
// of them says.
const commonMark = new Marked({ gfm: false, async: false });

// runningScheme matches an address that runs something, or carries its own
// content, rather than leading to a page.
const runningScheme = /^(javascript|vbscript|file|data):/i;

// characterReference matches a reference such as &#106; or &colon;, which the
// parser leaves in an address as it is and a browser reads as the character
// it stands for: an address written with one is not the address it shows.
const characterReference = /&(#\d+|#x[\da-f]+|[a-z][a-z\d]*);/i;

/**
 * bodyHTML is a document's body as HTML. Markup written into a text is refused
 * rather than carried through, and so is a link or an image whose address
 * runs something rather than leads somewhere: the texts are prose, and either
 * is a mistake the page would otherwise publish as it is.
 */
export function bodyHTML(body: string): string {
	const tokens = commonMark.lexer(body);
	const refused: string[] = [];
	commonMark.walkTokens(tokens, (token) => {
		if (token.type === "html") {
			refused.push(
				`the text holds markup, ${JSON.stringify(token.raw.trim())}, and the site's texts are Markdown alone`,
			);
		}
		if (
			(token.type === "link" || token.type === "image") &&
			mayRunSomething(token.href)
		) {
			refused.push(
				`the text leads to ${JSON.stringify(token.href)}, an address that may run something rather than lead to a page`,
			);
		}
	});
	if (refused.length > 0) {
		throw new Error(refused.join("; "));
	}
	return commonMark.parser(tokens);
}

// mayRunSomething says whether an address could be read by a browser as one
// that runs something: its scheme is such, once the spaces and control
// characters a browser drops are gone, or it hides a character behind a
// reference.
function mayRunSomething(href: string): boolean {
	const visible = [...href].filter((character) => character > " ").join("");
	return runningScheme.test(visible) || characterReference.test(href);
}

/**
 * readTexts reads every source, reference being the locale every other is
 * measured against. A source that cannot be read stops the build with the
 * file it came from, and so do a site with no locale, a locale with no text,
 * and a reference locale with no front page: the address the product is
 * listed under is built from it.
 */
export function readTexts(sources: Sources, reference: string): Texts {
	const pages = new Map<string, ReadonlyMap<string, PageText>>();
	const names = new Set<string>();
	for (const [locale, files] of sources) {
		if (files.size === 0) {
			throw new Error(`locale "${locale}" holds no text`);
		}
		const read = new Map<string, PageText>();
		for (const [name, source] of files) {
			read.set(name, pageText(source, `${locale}/${name}.md`));
			names.add(name);
		}
		pages.set(locale, read);
	}

	if (pages.size === 0) {
		throw new Error("the content holds no locale");
	}
	const front = pages.get(reference)?.get(frontPage);
	if (front === undefined) {
		throw new Error(
			`the reference locale "${reference}" has no ${frontPage}.md, the text its front page and the apex are made of`,
		);
	}
	return {
		locales: [...pages.keys()].sort(byCodeUnits),
		names: [...names].sort(byCodeUnits),
		pages,
		front,
	};
}

// pageText reads the source of the file named file, and says which file it was
// when it cannot.
function pageText(source: string, file: string): PageText {
	try {
		const { title, description, body } = parseDocument(source);
		return { title, description, html: bodyHTML(body) };
	} catch (error) {
		const said = error instanceof Error ? error.message : "it cannot be read";
		throw new Error(`${file}: ${said}`, { cause: error });
	}
}
