import type { Token } from "marked";
import type { ComponentChildren, VNode } from "preact";
import { directionOf } from "../i18n/lookup";
import { placeholder } from "../i18n/words";
import { leftToRightSpans, type Span } from "./bidi";
import { characterReference } from "./content";
import { inlineTokens } from "./inline";
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
	 * file is the file the words come from, as a refusal names it:
	 * es/topics/clocks.yaml.
	 */
	readonly file: string;
	/**
	 * text is the text under key with its emphasis drawn — *…* and **…**, the
	 * only markup a page's text may hold — and its slots filled.
	 */
	text(key: string, slots?: Readonly<Record<string, Fill>>): ComponentChildren;
	/**
	 * plain is the text under key as it is written, its slots filled: for the
	 * page's head and an element's attributes, and for a text whose stars are
	 * no emphasis.
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
	const arrange = directionOf(locale) === "rtl" ? keptLeftToRight : inOrder;

	const page: PageReader = {
		locale,
		file,
		text(key, slots = {}) {
			const where = `${file}: ${key}`;
			return drawnText(
				inlineTokens(take(key)),
				(name) => {
					const value = slotOf(slots, name, where);
					return typeof value === "object" ? value : written(value);
				},
				where,
				arrange,
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

// Piece is a part of a text as the page draws it, with the words it says as
// plain text: the text's own words, which a run that reads left to right may
// cut, or a slot's filling or an emphasis, which such a run takes whole or
// not at all. A filling the page draws as an element says nothing a run can
// read, and stands as a mark no run holds; an emphasis keeps the runs of its
// own words itself.
type Piece = {
	readonly drawn: ComponentChildren;
	readonly plain: string;
	readonly kind: "words" | "filling" | "emphasis";
};

// Arrange draws a text from its pieces, as its page's direction wants: in
// their order, or, on a page written right to left, with each run that reads
// left to right, a sum or a piece of code, set apart in an element of that
// direction.
type Arrange = (pieces: readonly Piece[]) => ComponentChildren;

// drawnText is a text's tokens drawn and arranged.
function drawnText(
	tokens: readonly Token[],
	fill: (name: string) => string | VNode,
	where: string,
	arrange: Arrange,
): ComponentChildren {
	return arrange(
		tokens.flatMap((token) => piecesOf(token, fill, where, arrange)),
	);
}

// inOrder is a text drawn from its pieces in their order, as a page written
// left to right wants it.
function inOrder(pieces: readonly Piece[]): ComponentChildren {
	return pieces.map((piece) => piece.drawn);
}

// piecesOf is one token of a text in its pieces: words with their slots
// filled, or emphasis around more of them, which says the words of its own
// pieces. Anything else — a link, an image, markup, code, a break — is
// refused: links and numbers are slots the page fills, and the rest has no
// place in a page's words.
function piecesOf(
	token: Token,
	fill: (name: string) => string | VNode,
	where: string,
	arrange: Arrange,
): Piece[] {
	switch (token.type) {
		case "text":
			return filled(token.text, fill, where);
		case "escape":
			return [{ drawn: token.text, plain: token.text, kind: "words" }];
		case "em":
		case "strong": {
			const inner = (token.tokens ?? []).flatMap((part) =>
				piecesOf(part, fill, where, arrange),
			);
			const Emphasis = token.type;
			return [
				{
					drawn: <Emphasis>{arrange(inner)}</Emphasis>,
					plain: inner.map((piece) => piece.plain).join(""),
					kind: "emphasis",
				},
			];
		}
		default:
			throw new Error(
				`${where} holds ${JSON.stringify(token.raw)}, and a page's text may hold emphasis alone: *…* or **…**`,
			);
	}
}

// keptLeftToRight is a text written right to left drawn from its pieces, each
// stretch of them that reads left to right set apart in an element of that
// direction.
function keptLeftToRight(pieces: readonly Piece[]): ComponentChildren[] {
	const bounds: Span[] = [];
	let at = 0;
	for (const { plain } of pieces) {
		bounds.push([at, at + plain.length]);
		at += plain.length;
	}
	const spans = leftToRightSpans(
		pieces.map(({ plain }) => plain).join(""),
		bounds.filter((_, index) => pieces[index]?.kind !== "words"),
	).filter(
		([first, last]) =>
			!bounds.some(
				([start, end], index) =>
					start === first && end === last && pieces[index]?.kind === "emphasis",
			),
	);
	const cuts = spans.flat();
	const drawnParts: ComponentChildren[] = [];
	let kept: ComponentChildren[] = [];
	let keptIn: Span | undefined;
	const close = () => {
		if (keptIn !== undefined) {
			drawnParts.push(
				<bdi key={keptIn[0]} dir="ltr">
					{kept}
				</bdi>,
			);
		}
		kept = [];
		keptIn = undefined;
	};
	pieces.forEach((piece, index) => {
		const [start, end] = bounds[index] ?? [0, 0];
		const marks =
			piece.kind === "words"
				? [start, ...cuts.filter((cut) => cut > start && cut < end), end]
				: [start, end];
		for (let place = 0; place + 1 < marks.length; place += 1) {
			const from = marks[place] ?? start;
			const to = marks[place + 1] ?? end;
			const part =
				piece.kind === "words" && typeof piece.drawn === "string"
					? piece.drawn.slice(from - start, to - start)
					: piece.drawn;
			const span = spans.find(([first, last]) => first <= from && to <= last);
			if (span !== keptIn) {
				close();
				keptIn = span;
			}
			if (span === undefined) {
				drawnParts.push(part);
			} else {
				kept.push(part);
			}
		}
	});
	close();
	return drawnParts;
}

// filled is words with their slots filled, in their pieces. A character
// written as a reference — &amp; — is refused: the page would show it as it
// is written.
function filled(
	text: string,
	fill: (name: string) => string | VNode,
	where: string,
): Piece[] {
	if (characterReference.test(text)) {
		throw new Error(
			`${where} writes a character as a reference, which the page would show as it is written`,
		);
	}
	const words = (part: string): Piece => ({
		drawn: part,
		plain: part,
		kind: "words",
	});
	const pieces: Piece[] = [];
	let from = 0;
	for (const match of text.matchAll(placeholder)) {
		const value = fill(match[1] ?? "");
		pieces.push(words(text.slice(from, match.index)), {
			drawn: value,
			plain: typeof value === "string" ? value : "\uFFFC",
			kind: "filling",
		});
		from = match.index + match[0].length;
	}
	pieces.push(words(text.slice(from)));
	return pieces;
}
