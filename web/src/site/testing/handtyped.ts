// A page that shows numbers it takes from data shows none it does not take
// from there: a number typed into its words, or into its code, would stay the
// same when the data moved, and say what the data no longer does. handTyped
// finds such a number on a page drawn: any run of digits, in any script, in
// what a reader reads.

/**
 * HandTypedRules are what a page may show with digits that are not its
 * data's: names that carry digits of their own, and the values of its data
 * that a commit's code or a date stands for.
 */
export type HandTypedRules = {
	/** names are names with digits of their own, such as a cipher's. */
	readonly names: readonly string[];
	/** values are the commits and the dates of the page's data, as it holds them. */
	readonly values: ReadonlySet<string>;
};

// digit is a digit of any script: the languages a page is drawn in write
// their numbers in their own.
const digit = /\p{Nd}/u;

// spoken are the attributes a reader is read or shown in place of a text.
const spoken = ["title", "aria-label", "alt"];

// described are the descriptions of a page its head carries for a reader, a
// search engine or a link shared.
const described = [
	'meta[name="description"]',
	'meta[property="og:title"]',
	'meta[property="og:description"]',
	'meta[property="og:image:alt"]',
];

/**
 * handTyped is every text of page that holds a digit it did not take from its
 * data, each with where it was found: the page's title and the descriptions
 * its head carries, and the text and the spoken attributes of its own part,
 * the main. A number written as data, one marked as the text's own choice, a
 * commit in code and a date in time that are values of the data, and the
 * names of rules are let through.
 */
export function handTyped(page: Document, rules: HandTypedRules): string[] {
	const found: string[] = [];
	const check = (where: string, text: string) => {
		if (digit.test(withoutNames(text, rules.names))) {
			found.push(`${where}: ${text}`);
		}
	};
	check("title", page.title);
	for (const selector of described) {
		check(
			selector,
			page.querySelector(selector)?.getAttribute("content") ?? "",
		);
	}
	const main = page.querySelector("main");
	if (main !== null) {
		walk(main, rules, check);
	}
	return found;
}

// walk checks an element and everything below it, but for what a page may
// show with digits.
function walk(
	element: Element,
	rules: HandTypedRules,
	check: (where: string, text: string) => void,
): void {
	const name = element.tagName.toLowerCase();
	if (
		name === "script" ||
		name === "style" ||
		name === "data" ||
		element.hasAttribute("data-given")
	) {
		return;
	}
	if (name === "code" || name === "time") {
		if (!isDataValue(element, rules.values)) {
			check(name, element.textContent ?? "");
		}
		return;
	}
	for (const attribute of spoken) {
		const value = element.getAttribute(attribute);
		if (value !== null) {
			check(`${name}[${attribute}]`, value);
		}
	}
	for (const child of element.childNodes) {
		if (child.nodeType === child.TEXT_NODE) {
			check(name, child.textContent ?? "");
		} else if (child.nodeType === child.ELEMENT_NODE) {
			walk(child as Element, rules, check);
		}
	}
}

// isDataValue says whether a commit's code or a date stands for a value of the
// data: a code that begins a commit of the data, as one is cut to be read, or
// a date whose datetime is a date of the data.
function isDataValue(element: Element, values: ReadonlySet<string>): boolean {
	if (element.tagName.toLowerCase() === "time") {
		return values.has(element.getAttribute("datetime") ?? "");
	}
	const code = element.textContent ?? "";
	return (
		/^[0-9a-f]{7,}$/.test(code) &&
		[...values].some((value) => value.startsWith(code))
	);
}

// withoutNames is text with every name of rules taken out of it.
function withoutNames(text: string, names: readonly string[]): string {
	return names.reduce((left, name) => left.split(name).join(""), text);
}
