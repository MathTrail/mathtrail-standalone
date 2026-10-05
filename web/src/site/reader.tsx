import { Lexer, type Token } from "marked";
import type { ComponentChildren, VNode } from "preact";
import { placeholder } from "../i18n/words";
import { characterReference } from "./content";
import type { PageWords } from "./pagewords";

/**
 * Fill is what fills one slot of a page's text: words, a number written the
 * way the page's language writes numbers, or an element, such as a link.
 */
export type Fill = string | number | VNode;

/**
 * PageReader is how a page's component reads its words. A key the file does
 * not have stops the build, and so, once the page is drawn, does a text of the
 * file the page never read: the words and the page are held to each other both
 * ways.
 */
export type PageReader = {
	/** locale is the language the page is drawn in. */
	readonly locale: string;
	/**
	 * text is the text under key with its emphasis drawn — *…* and **…**, the
	 * only markup a page's text may hold — and its slots filled.
	 */
	text(key: string, slots?: Readonly<Record<string, Fill>>): ComponentChildren;
	/**
	 * plain is the text under key as it is written, its slots filled: for the
	 * page's head, and for a text drawing, whose stars are no emphasis.
	 */
	plain(key: string, slots?: Readonly<Record<string, string | number>>): string;
	/** list is the keys of the items of the list under key, in order. */
	list(key: string): string[];
	/**
	 * has reports whether the file holds key, as a text or as a section of
	 * texts, and reads nothing: for a part some pages of one kind have and
	 * others do not, which the page still has to read to show.
	 */
	has(key: string): boolean;
	/**
	 * leaveOut counts the text under key, or every text of the section under
	 * key, as read and draws nothing: for words the page shows for some of its
	 * data and not for the rest, such as a button to a file the site does not
	 * ship yet. The file is still held to the key, so the words the page would
	 * show once the data changes are there.
	 */
	leaveOut(key: string): void;
};

/**
 * openReader opens the words of the page file names, in locale's language,
 * and unread says which of them the page has not read so far.
 */
export function openReader(
	words: PageWords,
	file: string,
	locale: string,
): { page: PageReader; unread: () => string[] } {
	const read = new Set<string>();
	const numbers = new Intl.NumberFormat(locale);
	const take = (key: string): string => {
		const text = words.get(key);
		if (text === undefined) {
			throw new Error(
				`${file}: the page reads ${key}, which the file does not have`,
			);
		}
		read.add(key);
		return text;
	};
	const written = (value: string | number): string =>
		typeof value === "number" ? numbers.format(value) : value;

	const page: PageReader = {
		locale,
		text(key, slots = {}) {
			const where = `${file}: ${key}`;
			return Lexer.lexInline(take(key), { gfm: false }).map((token) =>
				drawn(
					token,
					(name) => {
						const value = slotOf(slots, name, where);
						return typeof value === "object" ? value : written(value);
					},
					where,
				),
			);
		},
		plain(key, slots = {}) {
			const where = `${file}: ${key}`;
			return take(key).replace(placeholder, (_, name: string) =>
				written(slotOf(slots, name, where)),
			);
		},
		list(key) {
			const prefix = `${key}.`;
			const items = [
				...new Set(
					[...words.keys()]
						.filter((name) => name.startsWith(prefix))
						.map((name) => name.slice(prefix.length).split(".", 1)[0]),
				),
			];
			if (items.length === 0) {
				throw new Error(
					`${file}: the page reads the list ${key}, which the file does not have`,
				);
			}
			if (items.some((item, at) => item !== String(at + 1))) {
				throw new Error(
					`${file}: ${key} is a section, and the page reads it as a list`,
				);
			}
			return items.map((item) => prefix + item);
		},
		has(key) {
			const prefix = `${key}.`;
			return (
				words.has(key) ||
				[...words.keys()].some((name) => name.startsWith(prefix))
			);
		},
		leaveOut(key) {
			const prefix = `${key}.`;
			const section = [...words.keys()].filter((name) =>
				name.startsWith(prefix),
			);
			if (section.length === 0) {
				take(key);
			}
			for (const name of section) {
				read.add(name);
			}
		},
	};
	return {
		page,
		unread: () => [...words.keys()].filter((key) => !read.has(key)),
	};
}

// slotOf is what fills the slot name, and stops the build when the page gives
// it nothing: the page would otherwise show the placeholder as it is written.
function slotOf<Value>(
	slots: Readonly<Record<string, Value>>,
	name: string,
	where: string,
): Value {
	const value = Object.hasOwn(slots, name) ? slots[name] : undefined;
	if (value === undefined) {
		throw new Error(`${where} names {${name}}, and the page gives it nothing`);
	}
	return value;
}

// drawn is one token of a text: words with their slots filled, or emphasis
// around more of them. Anything else — a link, an image, markup, code, a
// break — is refused: links and numbers are slots the page fills, and the
// rest has no place in a page's words.
function drawn(
	token: Token,
	fill: (name: string) => string | VNode,
	where: string,
): ComponentChildren {
	switch (token.type) {
		case "text":
			return filled(token.text, fill, where);
		case "escape":
			return token.text;
		case "em":
			return <em>{inner(token.tokens, fill, where)}</em>;
		case "strong":
			return <strong>{inner(token.tokens, fill, where)}</strong>;
		default:
			throw new Error(
				`${where} holds ${JSON.stringify(token.raw)}, and a page's text may hold emphasis alone: *…* or **…**`,
			);
	}
}

// inner is what an emphasis holds, drawn.
function inner(
	tokens: readonly Token[] | undefined,
	fill: (name: string) => string | VNode,
	where: string,
): ComponentChildren {
	return (tokens ?? []).map((part) => drawn(part, fill, where));
}

// filled is words with their slots filled. A character written as a reference
// — &amp; — is refused: the page would show it as it is written.
function filled(
	text: string,
	fill: (name: string) => string | VNode,
	where: string,
): ComponentChildren {
	if (characterReference.test(text)) {
		throw new Error(
			`${where} writes a character as a reference, which the page would show as it is written`,
		);
	}
	const parts: ComponentChildren[] = [];
	let from = 0;
	for (const match of text.matchAll(placeholder)) {
		parts.push(text.slice(from, match.index), fill(match[1] ?? ""));
		from = match.index + match[0].length;
	}
	parts.push(text.slice(from));
	return parts;
}
