// One built page of the site, read from its HTML: where it sits, what its head
// promises, and every address it points at.

import { posix } from "node:path";
import { type DefaultTreeAdapterTypes, parse } from "parse5";

type Node = DefaultTreeAdapterTypes.Node;
type Element = DefaultTreeAdapterTypes.Element;

/** indexFile is the file every address of the site is served from. */
export const indexFile = "index.html";

const htmlNamespace = "http://www.w3.org/1999/xhtml";

/** Reference is one address a page points at. */
export type Reference = {
	element: string;
	attribute: string;
	value: string;
	/**
	 * subresource is true when the browser loads the address without being
	 * asked, which is what makes another origin a leak rather than a link.
	 */
	subresource: boolean;
};

/** Page is one parsed HTML file of the site. */
export type Page = {
	/** file is the path below the site's root, such as "en/privacy/index.html". */
	file: string;
	/** address is the path the file is served at, such as "/en/privacy/". */
	address: string;
	/**
	 * locale is the language the page is of: the first segment of its path, or
	 * the reference locale for the bare domain's page, which is that locale's
	 * front page.
	 */
	locale: string;
	/** name is the page within its locale, empty for the locale's front page. */
	name: string;
	lang: string;
	dir: string;
	title: string;
	description: string;
	/**
	 * images are the pictures a shared link to the page shows, as its head
	 * names them.
	 */
	images: string[];
	canonical: string;
	/** alternates are the translations the page names, by their hreflang. */
	alternates: Map<string, string>;
	/**
	 * refresh is what the page's first refresh that a browser acts on does — a
	 * browser acts on no other: the address it sends the reader on to at once,
	 * or empty when it reloads the page or waits first. It is undefined when the
	 * page has no refresh a browser acts on.
	 */
	refresh: string | undefined;
	references: Reference[];
	/** ids are the anchors a link may lead to within the page. */
	ids: Set<string>;
};

// subresourceRels are the link relations a browser acts on by itself. A
// canonical or an alternate is a statement about the page, not a file to fetch.
const subresourceRels = new Set([
	"stylesheet",
	"icon",
	"shortcut icon",
	"apple-touch-icon",
	"manifest",
	"preload",
	"modulepreload",
	"prefetch",
]);

// srcElements are the elements whose src the browser fetches on its own.
const srcElements = new Set([
	"script",
	"img",
	"iframe",
	"source",
	"video",
	"audio",
	"embed",
	"track",
	"object",
]);

/**
 * locate works out where a file sits in the site from its path alone: the
 * first segment is the locale, and the rest up to the index file is the page.
 * The bare domain's file is the front page of reference, the locale every
 * other is translated from.
 */
export function locate(
	file: string,
	reference: string,
): {
	locale: string;
	name: string;
	address: string;
} {
	const trimmed = file
		.slice(0, file.length - indexFile.length)
		.replace(/\/$/, "");
	if (trimmed === "") {
		return { locale: reference, name: "", address: "/" };
	}
	const slash = trimmed.indexOf("/");
	if (slash < 0) {
		return { locale: trimmed, name: "", address: `/${trimmed}/` };
	}
	const locale = trimmed.slice(0, slash);
	const name = trimmed.slice(slash + 1);
	return { locale, name, address: `/${locale}/${name}/` };
}

/**
 * isExternal reports whether a reference leaves the site. A scheme or a
 * protocol-relative prefix is the whole test: everything else is the site's own.
 * Only a letter may open a scheme, which keeps a bare colon in a file name from
 * reading as one.
 */
export function isExternal(value: string): boolean {
	if (value.startsWith("//")) {
		return true;
	}
	const colon = value.indexOf(":");
	if (colon <= 0) {
		return false;
	}
	return /^[A-Za-z][A-Za-z0-9+.-]*$/.test(value.slice(0, colon));
}

/**
 * resolveReference returns the file a reference asks for, relative to the site's
 * root, or an empty string when there is nothing to look for: an anchor within
 * the page, or an empty address. The fragment and the query are cut off first,
 * so that an anchor never reads as part of a file's name; and an address with
 * no extension asks for the index of a directory, since every page of the site
 * is one.
 */
export function resolveReference(value: string, fromFile: string): string {
	const cut = value.search(/[#?]/);
	const asked = cut < 0 ? value : value.slice(0, cut);
	if (asked === "") {
		return "";
	}
	const target = asked.startsWith("/")
		? cleaned(asked.slice(1))
		: cleaned(posix.join(posix.dirname(fromFile), asked));
	return asked.endsWith("/") || extension(target) === ""
		? cleaned(posix.join(target, indexFile))
		: target;
}

// cleaned is a path with its dots resolved and no trailing slash, the way a
// file below the site's root is named.
function cleaned(path: string): string {
	const normal = posix.normalize(path === "" ? "." : path);
	return normal.length > 1 && normal.endsWith("/")
		? normal.slice(0, -1)
		: normal;
}

// extension is the suffix of a path's last element from its last dot, empty
// when the element has no dot.
function extension(path: string): string {
	const last = path.slice(path.lastIndexOf("/") + 1);
	const dot = last.lastIndexOf(".");
	return dot < 0 ? "" : last.slice(dot);
}

/**
 * parsePage reads one built HTML file of the site, whose bare domain serves
 * the front page of the reference locale.
 */
export function parsePage(file: string, html: string, reference: string): Page {
	const page: Page = {
		file,
		...locate(file, reference),
		lang: "",
		dir: "",
		title: "",
		description: "",
		images: [],
		canonical: "",
		alternates: new Map(),
		refresh: undefined,
		references: [],
		ids: new Set(),
	};
	walk(parse(html), page);
	return page;
}

// walk reads the head and collects every reference in the document, the content
// of a template included.
function walk(node: Node, page: Page, inTemplate = false): void {
	if ("tagName" in node) {
		readElement(node, page, inTemplate);
	}
	if ("content" in node && node.content !== undefined) {
		walk(node.content, page, true);
	}
	if ("childNodes" in node) {
		for (const child of node.childNodes) {
			walk(child, page, inTemplate);
		}
	}
}

// attribute is the value of an element's attribute, empty when it has none.
function attribute(element: Element, name: string): string {
	return element.attrs.find((a) => a.name === name)?.value ?? "";
}

// readElement takes from one element whatever the checks need from it. What a
// template holds is not part of the page until a script puts it there: it gives
// the page neither an anchor nor anything of its head, while the files it names
// still count, since the page may yet load them.
function readElement(element: Element, page: Page, inTemplate: boolean): void {
	if (!inTemplate) {
		readOwn(element, page);
	}
	if (element.tagName === "a") {
		addReference(element, "href", false, page);
	} else if (element.tagName === "link") {
		readLink(element, page, inTemplate);
	}
	if (srcElements.has(element.tagName)) {
		addReference(element, "src", true, page);
	}
}

// readOwn takes what only the page's own elements give it: an anchor, its
// language and direction, its title, its description, the pictures a shared
// link to it shows, and what its refresh does.
function readOwn(element: Element, page: Page): void {
	const id = attribute(element, "id");
	if (id !== "") {
		page.ids.add(id);
	}
	switch (element.tagName) {
		case "html":
			page.lang = attribute(element, "lang");
			page.dir = attribute(element, "dir");
			break;
		case "title":
			// The page's title is the first title of the document's own: one inside
			// an SVG names the drawing, not the page.
			if (page.title === "" && element.namespaceURI === htmlNamespace) {
				page.title = textOf(element).trim();
			}
			break;
		case "meta":
			if (attribute(element, "name").toLowerCase() === "description") {
				page.description = attribute(element, "content").trim();
			} else if (attribute(element, "property").toLowerCase() === "og:image") {
				page.images.push(attribute(element, "content").trim());
			} else if (
				attribute(element, "http-equiv").toLowerCase() === "refresh" &&
				page.refresh === undefined
			) {
				page.refresh = refreshOf(attribute(element, "content"));
			}
			break;
	}
}

// htmlSpaces are the characters HTML skips as white space: tab, line feed, form
// feed, carriage return and space, and no other.
const htmlSpaces = new Set(
	[9, 10, 12, 13, 32].map((code) => String.fromCodePoint(code)),
);

// refreshOf is what a refresh with content does, read the way a browser reads
// it: the address it sends the reader on to at once, as /en/ for "0; url=/en/",
// or empty when it only reloads the page or waits first, since a page seen
// before it leaves is a page to judge as it is. It is undefined when a browser
// would not act on the content at all, as on "url=/en/" with no time, and then
// a later refresh of the page may still count. A browser counts whole seconds
// alone, so "0.5" is at once.
function refreshOf(content: string): string | undefined {
	const input = withoutLeadingSpace(content);
	const seconds = input.slice(0, input.search(/\D|$/));
	if (seconds === "" && !input.startsWith(".")) {
		return undefined;
	}
	let rest = input.slice(input.search(/[^\d.]|$/));
	const separated =
		rest === "" ||
		rest.startsWith(";") ||
		rest.startsWith(",") ||
		htmlSpaces.has(rest.charAt(0));
	if (!separated) {
		return undefined;
	}
	if (Number(seconds) !== 0) {
		return "";
	}
	rest = withoutLeadingSpace(rest);
	if (rest.startsWith(";") || rest.startsWith(",")) {
		rest = withoutLeadingSpace(rest.slice(1));
	}
	if (rest.slice(0, 3).toLowerCase() === "url") {
		const named = withoutLeadingSpace(rest.slice(3));
		if (named.startsWith("=")) {
			rest = withoutLeadingSpace(named.slice(1));
		}
	}
	const quote = rest.charAt(0);
	if (quote === '"' || quote === "'") {
		const end = rest.indexOf(quote, 1);
		rest = rest.slice(1, end < 0 ? undefined : end);
	}
	return rest;
}

// withoutLeadingSpace is text without the white space HTML skips at its start.
function withoutLeadingSpace(text: string): string {
	let at = 0;
	while (htmlSpaces.has(text.charAt(at))) {
		at++;
	}
	return text.slice(at);
}

// textOf is the text an element holds directly.
function textOf(element: Element): string {
	return element.childNodes
		.map((child) => ("value" in child ? child.value : ""))
		.join("");
}

// readLink sorts a link element into the three things it can be: a statement of
// the page's own address, a pointer at a translation, or a file to fetch. A
// statement inside a template says nothing about the page.
function readLink(element: Element, page: Page, inTemplate: boolean): void {
	const rel = attribute(element, "rel").trim().toLowerCase();
	if (rel === "canonical") {
		if (!inTemplate) {
			page.canonical = attribute(element, "href");
		}
	} else if (rel === "alternate") {
		const hreflang = attribute(element, "hreflang");
		if (hreflang !== "" && !inTemplate) {
			page.alternates.set(hreflang, attribute(element, "href"));
		}
	} else {
		addReference(element, "href", subresourceRels.has(rel), page);
	}
}

// addReference records the address an element's attribute points at.
function addReference(
	element: Element,
	name: string,
	subresource: boolean,
	page: Page,
): void {
	const value = attribute(element, name);
	if (value !== "") {
		page.references.push({
			element: element.tagName,
			attribute: name,
			value,
			subresource,
		});
	}
}
