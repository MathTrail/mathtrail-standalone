/**
 * rule is one rule of style a page's head carries, written when the site is
 * built: its selectors and its declarations, or nothing when it has no
 * selector, which would leave the declarations standing alone.
 */
export function rule(
	selectors: readonly string[],
	declarations: string,
): string {
	return selectors.length === 0
		? ""
		: `${selectors.join(",\n")}{${declarations}}`;
}
