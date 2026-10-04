// One stylesheet of the site, read for the addresses it loads by itself: the
// files a browser fetches for a page that is never shown them in its HTML.

// comment matches a comment of a stylesheet, whose words load nothing.
const comment = /\/\*[\s\S]*?\*\//g;

// load matches what a stylesheet fetches: an address in url(), quoted or not,
// and the string an @import names without one.
const load =
	/\burl\(\s*(?:"([^"]*)"|'([^']*)'|([^"'()\s]*))\s*\)|@import\s+(?:"([^"]*)"|'([^']*)')/gi;

/**
 * stylesheetReferences are the addresses a stylesheet loads, in the order it
 * names them: every url() and every @import, leaving out what its comments
 * say.
 */
export function stylesheetReferences(text: string): string[] {
	return [...text.replace(comment, "").matchAll(load)].map(
		(match) => match.slice(1).find((address) => address !== undefined) ?? "",
	);
}
