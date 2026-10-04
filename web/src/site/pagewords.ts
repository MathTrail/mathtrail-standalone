import {
	isAlias,
	isMap,
	isScalar,
	isSeq,
	parseAllDocuments,
	visit,
	type YAMLMap,
	type YAMLSeq,
} from "yaml";
import { disagreements } from "../i18n/dictionaries";

/**
 * PageWords are the words of one page in one language, as its YAML file
 * writes them: every text under its key, sections and lists laid out into keys
 * in dot notation — hero.title, examples.1.question — in the order the file
 * writes them.
 */
export type PageWords = ReadonlyMap<string, string>;

// keyName is what a key of a page's file may be: lowercase letters, digits,
// dashes and underscores, beginning with a letter. A dot would read as two
// keys, and a list numbers its items itself.
const keyName = /^[a-z][a-z0-9_-]*$/;

/**
 * parsePageWords reads a page's YAML file. Every value is text — the failsafe
 * schema reads no numbers, dates or booleans — and the file is one document of
 * sections, lists and texts, with nothing that stands for something written
 * elsewhere: an anchor, an alias or a tag is refused, and so is a key written
 * twice and a key no page could name, the merge key among them. YAML reads a
 * text that begins with a star and no quote as an alias, which the parser
 * leaves standing even with no anchor to repeat, so the alias is named with the
 * way out: quotes. Comments are
 * notes for whoever translates, and are left out. A block of text loses the
 * line break YAML keeps at its end.
 */
export function parsePageWords(source: string): PageWords {
	const documents = parseAllDocuments(source, {
		schema: "failsafe",
		uniqueKeys: true,
	});
	const [document] = documents;
	if (documents.length !== 1 || document === undefined) {
		throw new Error(
			`the file holds ${documents.length} documents, and a page's words are one`,
		);
	}
	const [error] = document.errors;
	if (error !== undefined) {
		throw new Error(firstLine(error.message));
	}
	visit(document, {
		Node(_, node) {
			if (isAlias(node)) {
				throw new Error(
					`the file reads *${node.source} as an alias, and a page's words write every text where it is read: a text that begins with * is written in quotes`,
				);
			}
			if (node.anchor !== undefined) {
				throw new Error(
					`the file names &${node.anchor}, and a page's words write every text where it is read`,
				);
			}
			if (node.tag !== undefined) {
				throw new Error(
					`the file tags a value ${node.tag}, and every value of a page's words is text`,
				);
			}
		},
	});
	const [warning] = document.warnings;
	if (warning !== undefined) {
		throw new Error(firstLine(warning.message));
	}
	if (!isMap(document.contents)) {
		throw new Error("the file is no set of sections and texts");
	}
	const words = new Map<string, string>();
	layOut(document.contents, "", words);
	return words;
}

// layOut adds what node holds to words under key: a text as it is, a section
// key by key, a list item by item from 1.
function layOut(node: unknown, key: string, words: Map<string, string>): void {
	if (isScalar(node)) {
		words.set(key, String(node.value).replace(/\n$/, ""));
	} else if (isMap(node)) {
		laySectionOut(node, key, words);
	} else if (isSeq(node)) {
		layListOut(node, key, words);
	} else {
		// A key written with nothing after it: its text is empty, which the
		// comparison of a page's words refuses by name.
		words.set(key, "");
	}
}

// laySectionOut adds what a section holds to words, key by key under its own
// key, refusing a section with nothing in it and a key no page may have.
function laySectionOut(
	section: YAMLMap,
	key: string,
	words: Map<string, string>,
): void {
	if (section.items.length === 0) {
		throw new Error(`${key} is an empty section`);
	}
	for (const { key: written, value } of section.items) {
		const name = isScalar(written) ? String(written.value) : "";
		if (!keyName.test(name)) {
			throw new Error(
				`${JSON.stringify(name)} is no key a page's words may have: a key is lowercase letters, digits, dashes and underscores, and begins with a letter`,
			);
		}
		layOut(value, key === "" ? name : `${key}.${name}`, words);
	}
}

// layListOut adds what a list holds to words, item by item from 1 under its
// own key, refusing a list with no item.
function layListOut(
	list: YAMLSeq,
	key: string,
	words: Map<string, string>,
): void {
	if (list.items.length === 0) {
		throw new Error(`${key} is an empty list`);
	}
	list.items.forEach((item, at) => {
		layOut(item, `${key}.${at + 1}`, words);
	});
}

// firstLine is what a parser's message says before the excerpt of the file it
// quotes beneath.
function firstLine(message: string): string {
	return (message.split("\n", 1)[0] ?? message).replace(/:$/, "");
}

/**
 * pageWordsDisagreements are the ways one language's words for a page, tagged
 * tag, say other things than the English ones, as the site's dictionaries are
 * held to them — a key one of them lacks, other slots, an empty text, a mark a
 * reader cannot see — and, once the keys agree, a key out of the English
 * order: the two files are read side by side, line by line. The English words
 * are held to themselves, for an empty text or a hidden mark of their own.
 */
export function pageWordsDisagreements(
	english: PageWords,
	words: PageWords,
	tag: string,
): string[] {
	const found = disagreements(
		Object.fromEntries(english),
		Object.fromEntries(words),
		tag,
	);
	if (found.length > 0) {
		return found;
	}
	const theirs = [...words.keys()];
	const ours = [...english.keys()];
	const at = ours.findIndex((key, index) => theirs[index] !== key);
	return at < 0
		? []
		: [`${theirs[at]} comes where the English has ${ours[at]}`];
}
