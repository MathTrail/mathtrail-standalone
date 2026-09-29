import type { RefObject } from "preact";
import { useLayoutEffect, useState } from "preact/hooks";

/**
 * useWide says whether an element is at least as wide as a wide card, the
 * width the design's tokens give as --widget-wide. The element is measured
 * itself, not the window, so that a card follows the room it is given — in a
 * chat's frame as on a page — and the width is read from the tokens rather
 * than written here a second time. Where there is nothing to measure with, the
 * card is narrow: the layout that fits everywhere.
 */
export function useWide(element: RefObject<HTMLElement>): boolean {
	const [wide, setWide] = useState(false);
	useLayoutEffect(() => {
		const target = element.current;
		if (target === null || typeof ResizeObserver === "undefined") {
			return;
		}
		const from = Number.parseFloat(
			getComputedStyle(target).getPropertyValue("--widget-wide"),
		);
		if (!Number.isFinite(from)) {
			return;
		}
		// Measured once before the first paint, so that a wide card is not drawn
		// narrow first and then laid out again.
		setWide(target.getBoundingClientRect().width >= from);
		const observer = new ResizeObserver(([entry]) => {
			if (entry !== undefined) {
				setWide(widthOf(entry) >= from);
			}
		});
		observer.observe(target);
		return () => observer.disconnect();
	}, [element]);
	return wide;
}

// widthOf is how wide an observed element is along its line, border included,
// which is how the tokens measure a card. A browser that reports no border box
// gives the content box instead, a few pixels short.
function widthOf(entry: ResizeObserverEntry): number {
	return entry.borderBoxSize?.[0]?.inlineSize ?? entry.contentRect.width;
}
