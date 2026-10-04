import { isMap, isScalar, isSeq, parseAllDocuments, visit } from "yaml";
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
 * elsewhere: an anchor is refused, and with it every alias, which can only
 * repeat an anchor written before it; so are a tag, a key written twice and a
 * key no page could name, the merge key among them. Comments are
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
		return;
	}
	if (isMap(node)) {
		if (node.items.length === 0) {
			throw new Error(`${key} is an empty section`);
		}
		for (const { key: written, value } of node.items) {
			const name = isScalar(written) ? String(written.value) : "";
			if (!keyName.test(name)) {
				throw new Error(
					`${JSON.stringify(name)} is no key a page's words may have: a key is lowercase letters, digits, dashes and underscores, and begins with a letter`,
				);
			}
			layOut(value, key === "" ? name : `${key}.${name}`, words);
		}
		return;
	}
	if (isSeq(node)) {
		if (node.items.length === 0) {
			throw new Error(`${key} is an empty list`);
		}
		node.items.forEach((item, at) => {
			layOut(item, `${key}.${at + 1}`, words);
		});
		return;
	}
	// A key written with nothing after it: its text is empty, which the
	// comparison of a page's words refuses by name.
	words.set(key, "");
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
