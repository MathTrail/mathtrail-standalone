/**
 * scriptJSON is data written as JSON a script element carries as it is: each
 * "<" in it written as its escape, so that no word of the data can end the
 * element or open a comment. A browser never runs such a script; the page's
 * demo and a search engine read it.
 */
export function scriptJSON(data: unknown): string {
	return JSON.stringify(data).replaceAll("<", String.raw`\u003c`);
}
