import type { RefObject } from "preact";
import { useLayoutEffect, useState } from "preact/hooks";

/**
 * useWide says whether an element is wide enough for a wide card: at least
 * the width the design's tokens give as --widget-wide, or, once it is wide,
 * short of it by no more than the scrollbar its page shows. A card laid out
 * wide can be taller than laid out narrow, and a frame sized to the card a
 * moment late then shows a scrollbar, which takes its width from the card:
 * were the card to go narrow then, it would grow shorter, lose the scrollbar,
 * go wide again, and swing between the two for as long as it is shown. The
 * element is measured itself, not the window, so that a card follows the room
 * it is given — in a chat's frame as on a page — and the width is read from
 * the tokens rather than written here a second time. Where there is nothing
 * to measure with, the card is narrow: the layout that fits everywhere.
 */
export function useWide(element: RefObject<HTMLElement | null>): boolean {
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
				const width = widthOf(entry);
				const room = width + scrollbarOf(target.ownerDocument) + rounding;
				setWide((was) => width >= from || (was && room >= from));
			}
		});
		observer.observe(target);
		return () => observer.disconnect();
	}, [element]);
	return wide;
}

// rounding is the pixel a page's widths may be off by: a window measures them
// in whole pixels, and a card's own width is not one.
const rounding = 1;

// widthOf is how wide an observed element is along its line, border included,
// which is how the tokens measure a card. A browser that reports no border box
// gives the content box instead, a few pixels short.
function widthOf(entry: ResizeObserverEntry): number {
	return entry.borderBoxSize?.[0]?.inlineSize ?? entry.contentRect.width;
}

// scrollbarOf is how wide the scrollbar a page shows down its side is: none
// where the page fits its window or is not laid out at all, and none where
// the browser draws scrollbars over a page rather than beside it.
function scrollbarOf(page: Document): number {
	const view = page.defaultView;
	const inside = page.documentElement.clientWidth;
	return view === null || inside === 0
		? 0
		: Math.max(0, view.innerWidth - inside);
}
