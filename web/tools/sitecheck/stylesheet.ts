// One stylesheet of the site, read for the addresses it loads by itself: the
// files a browser fetches for a page that is never shown them in its HTML.

// comment matches a comment of a stylesheet, whose words load nothing.
const comment = /\/\*[\s\S]*?\*\//g;

// load matches what a stylesheet fetches: an address in url(), quoted or bare,
// and the string an @import names without one. A bare address is at least one
// character long: a url() with nothing in it loads nothing, and an empty one
// would let the spaces before it and after it match one run of spaces both
// ways, which takes time growing with the square of its length. A quoted one
// ends at the next quote of its kind, so a quote that never closes costs no
// more. The two forms are one expression, matched in one pass, so that a url()
// written inside an @import's string is never read as a second load.
const load =
	/\burl\(\s*(?:"([^"]*)"|'([^']*)'|([^"'()\s]+))\s*\)|@import\s+(?:"([^"]*)"|'([^']*)')/gi;

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
