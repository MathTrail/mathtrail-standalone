import { Lexer, Marked, type Token } from "marked";

/**
 * commonMark reads the site's texts as CommonMark alone, the Markdown they are
 * written in: GitHub's additions — tables, struck-through text, bare addresses
 * turned into links — would change what a text that happens to look like one
 * of them says. The documents are read and drawn by it, and a page's words are
 * read with its options, so that what the build checks is what a page shows.
 */
export const commonMark = new Marked({ gfm: false, async: false });

/**
 * inlineTokens are the tokens a text of a page's words is drawn from: the
 * page draws a text from them, and the build reads a title's markup and a
 * text's emphasis from the same tokens.
 */
export function inlineTokens(text: string): Token[] {
	return Lexer.lexInline(text, commonMark.defaults);
}

/**
 * emphasisOf says what tokens stress as a page draws them: how many emphases
 * and strong emphases, and how many of emphasis's signs, * and _, show as they
 * are written — the star in 38*6, the underscore of submit_task, or the signs
 * of an emphasis CommonMark cannot read, as one closed after a full stop and
 * before a letter: **矛盾。**次に. A text that holds tokens of its own, as a
 * list's item does, is read in them.
 */
export function emphasisOf(tokens: Token[]): string {
	let emphases = 0;
	let strong = 0;
	let signs = 0;
	commonMark.walkTokens(tokens, (token) => {
		if (token.type === "em") {
			emphases += 1;
		} else if (token.type === "strong") {
			strong += 1;
		} else if (
			(token.type === "text" && token.tokens === undefined) ||
			token.type === "escape"
		) {
			signs += token.raw.replaceAll(/[^*_]/g, "").length;
		}
	});
	return `emphasis ${emphases}, strong emphasis ${strong}, emphasis signs shown as written ${signs}`;
}
