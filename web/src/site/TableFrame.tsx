import type { ComponentChildren } from "preact";

/**
 * TableFrame holds a table that may be wider than a narrow screen, which
 * scrolls it sideways inside the frame. A keyboard can reach the frame to
 * scroll it, and a screen reader names it by the heading labelledBy points to.
 * A page whose frame looks otherwise names a class of its own beside the
 * frame's.
 */
export function TableFrame({
	labelledBy,
	class: look,
	children,
}: {
	labelledBy: string;
	class?: string;
	children: ComponentChildren;
}) {
	return (
		<section
			class={look === undefined ? "s-table-frame" : `s-table-frame ${look}`}
			aria-labelledby={labelledBy}
			// biome-ignore lint/a11y/noNoninteractiveTabindex: a table wider than the screen scrolls inside its frame, and a keyboard has to reach the frame to scroll it
			tabIndex={0}
		>
			{children}
		</section>
	);
}
