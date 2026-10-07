import { Lexer, Marked } from "marked";
import { byCodeUnits } from "../i18n/order";
import { placeholder } from "../i18n/words";
import {
	type PageWords,
	pageWordsDisagreements,
	parsePageWords,
} from "./pagewords";

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
 * PageText is one page's text in one language, read: a document, its head's
 * fields and its body as HTML, or the words of a page a component draws.
 */
export type PageText =
	| {
			readonly kind: "document";
			readonly title: string;
			readonly description: string;
			readonly html: string;
	  }
	| { readonly kind: "words"; readonly words: PageWords };

/** Summary is what a page says of itself: its title and its description. */
export type Summary = { readonly title: string; readonly description: string };

/**
 * Sources are the site's texts as they are written: by locale, then by the
 * file's path below the locale's directory — index.md, topics/sample.yaml —
 * the whole of each file.
 */
export type Sources = ReadonlyMap<string, ReadonlyMap<string, string>>;

/**
 * Texts are every text the site has, read, and the two orders a build walks
 * them in: the locales, and the names of the pages, both sorted, so that two
 * builds of the same texts are the same files. Every locale has every page.
 * front is what the reference locale's front page says of itself, which the
 * page at the address it moved from is made of.
 */
export type Texts = {
	readonly locales: readonly string[];
	readonly names: readonly string[];
	readonly pages: ReadonlyMap<string, ReadonlyMap<string, PageText>>;
	readonly front: Summary;
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

/**
 * characterReference matches a reference such as &#106; or &colon;, which the
 * parser leaves in an address as it is and a browser reads as the character
 * it stands for: an address written with one is not the address it shows.
 */
export const characterReference = /&(#\d+|#x[\da-f]+|[a-z][a-z\d]*);/i;

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

// pageName matches what a page may be called: lowercase words joined by
// dashes, perhaps below a directory of the same kind — privacy,
// topics/knights-and-liars — since the name is the page's address.
const pageName = /^[a-z0-9]+(?:-[a-z0-9]+)*(?:\/[a-z0-9]+(?:-[a-z0-9]+)*)*$/;

/**
 * readTexts reads every source, reference being the locale every other is
 * measured against. A source that cannot be read stops the build with the
 * file it came from, and so do a site with no locale, a locale with no text,
 * a page two files of one locale both write, a page some locale lacks, words
 * that say other things than the English ones, and a reference locale with no
 * front page: the address the product is listed under is built from it.
 */
export function readTexts(sources: Sources, reference: string): Texts {
	const pages = new Map<string, ReadonlyMap<string, PageText>>();
	for (const [locale, files] of sources) {
		if (files.size === 0) {
			throw new Error(`locale "${locale}" holds no text`);
		}
		const read = new Map<string, PageText>();
		for (const [file, source] of files) {
			const [name, text] = pageText(`${locale}/${file}`, source);
			if (read.has(name)) {
				throw new Error(
					`${locale}/${file}: the page ${name} has a text already, and a page has one`,
				);
			}
			read.set(name, text);
		}
		pages.set(locale, read);
	}

	if (pages.size === 0) {
		throw new Error("the content holds no locale");
	}
	const locales = [...pages.keys()].sort(byCodeUnits);
	const names = [
		...new Set([...pages.values()].flatMap((read) => [...read.keys()])),
	].sort(byCodeUnits);
	const texts = { locales, names, pages };
	checkEveryLanguage(texts);
	const front = pages.get(reference)?.get(frontPage);
	if (front === undefined) {
		throw new Error(
			`the reference locale "${reference}" has no ${frontPage} page, the text its front page and the page at the address it moved from are made of`,
		);
	}
	checkWords(texts, reference);
	return { ...texts, front: summaryOf(front) };
}

/**
 * textOf is the text of the page name in locale, which a site read by
 * readTexts always has.
 */
export function textOf(
	texts: Pick<Texts, "pages">,
	locale: string,
	name: string,
): PageText {
	const text = texts.pages.get(locale)?.get(name);
	if (text === undefined) {
		throw new Error(`the site has no page ${name} in ${locale}`);
	}
	return text;
}

/** summaryOf is what a page's text says of the page: its title and description. */
export function summaryOf(text: PageText): Summary {
	return text.kind === "document"
		? { title: text.title, description: text.description }
		: {
				title: text.words.get("title") ?? "",
				description: text.words.get("description") ?? "",
			};
}

// checkEveryLanguage refuses a page some locale lacks, or one that is a
// document in one language and a component's words in another: a reader who
// switches language lands on the same page, and the checker of a built site
// only finds a page a translation lacks.
function checkEveryLanguage(texts: Omit<Texts, "front">): void {
	for (const name of texts.names) {
		const lacking = texts.locales.filter(
			(locale) => !texts.pages.get(locale)?.has(name),
		);
		if (lacking.length > 0) {
			throw new Error(
				`the page ${name} is missing in ${lacking.join(", ")}: a page is in every language of the site or in none`,
			);
		}
		const kinds = new Set(
			texts.locales.map((locale) => textOf(texts, locale, name).kind),
		);
		if (kinds.size > 1) {
			throw new Error(
				`the page ${name} is a document in one language and a component's words in another`,
			);
		}
	}
}

// checkWords holds every language's words for a page to the reference
// locale's, the reference's own among them.
function checkWords(texts: Omit<Texts, "front">, reference: string): void {
	for (const name of texts.names) {
		const english = textOf(texts, reference, name);
		if (english.kind !== "words") {
			continue;
		}
		for (const locale of texts.locales) {
			const text = textOf(texts, locale, name);
			const found =
				text.kind === "words"
					? pageWordsDisagreements(english.words, text.words, locale)
					: [];
			if (found.length > 0) {
				throw new Error(
					`${locale}/${name}.yaml: the words disagree with the English: ${found.join("; ")}`,
				);
			}
		}
	}
}

// pageText reads one file of the site's texts, path being where it is below
// the texts' directory, and says which file it was when it cannot: a .md file
// is a document, and a .yaml file the words of a page a component draws.
function pageText(path: string, source: string): [string, PageText] {
	const extension = /\.(md|yaml)$/.exec(path);
	const name = path.slice(path.indexOf("/") + 1, extension?.index);
	if (extension === null || !pageName.test(name)) {
		throw new Error(
			`${path}: a text is a .md or a .yaml file, named in lowercase words joined by dashes, as its page's address is`,
		);
	}
	try {
		return [
			name,
			extension[1] === "md" ? documentText(source) : wordsText(source),
		];
	} catch (error) {
		const said = error instanceof Error ? error.message : "it cannot be read";
		throw new Error(`${path}: ${said}`, { cause: error });
	}
}

// documentText reads a document: its front matter, and its body as HTML.
function documentText(source: string): PageText {
	const { title, description, body } = parseDocument(source);
	return { kind: "document", title, description, html: bodyHTML(body) };
}

// wordsText reads a page's words, which give its title and its description as
// they are written: the page's head, the page an address that moved serves
// and the footer show them with no slot filled and no markup read, so a slot
// or an emphasis would show as its signs.
function wordsText(source: string): PageText {
	const words = parsePageWords(source);
	for (const key of ["title", "description"]) {
		const text = words.get(key);
		if (text === undefined) {
			throw new Error(`a page's words give its ${key}`);
		}
		if (new RegExp(placeholder.source).test(text)) {
			throw new Error(`a page's ${key} is said as it is written, with no slot`);
		}
		if (
			characterReference.test(text) ||
			Lexer.lexInline(text, { gfm: false }).some(({ type }) => type !== "text")
		) {
			throw new Error(
				`a page's ${key} is said as it is written, with no markup`,
			);
		}
	}
	return { kind: "words", words };
}
