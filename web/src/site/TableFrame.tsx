import type { ComponentChildren } from "preact";

/**
 * TableFrame holds a table that may be wider than a narrow screen, which
 * scrolls it sideways inside the frame. A keyboard can reach the frame to
 * scroll it, and a screen reader names it by the heading labelledBy points to.
 */
export function TableFrame({
	labelledBy,
	children,
}: {
	labelledBy: string;
	children: ComponentChildren;
}) {
	return (
		<section
			class="s-table-frame"
			aria-labelledby={labelledBy}
			// biome-ignore lint/a11y/noNoninteractiveTabindex: a table wider than the screen scrolls inside its frame, and a keyboard has to reach the frame to scroll it
			tabIndex={0}
		>
			{children}
		</section>
	);
}
