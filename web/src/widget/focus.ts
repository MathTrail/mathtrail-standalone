import { useEffect } from "preact/hooks";

/**
 * useFocusKeptOnTheCard gives the focus to target — or, when there is none on
 * the card, to fallback — when shown comes true and the focus the card had has
 * been lost: the button pressed has gone or been switched off. A focus still
 * on something is left where it is, a focus elsewhere in the chat is never
 * taken, and the page is not scrolled.
 */
export function useFocusKeptOnTheCard(
	shown: boolean,
	target: { current: HTMLElement | null },
	fallback?: { current: HTMLElement | null },
): void {
	useEffect(() => {
		if (shown && focusIsLost()) {
			(target.current ?? fallback?.current)?.focus({ preventScroll: true });
		}
	}, [shown, target, fallback]);
}

// focusIsLost says whether the card holds the focus on nothing a child could
// act on: the page itself, or an element that has left the page or been
// switched off. A card the focus is not in has lost nothing: the focus is
// where the child put it, in the chat or another card.
function focusIsLost(): boolean {
	if (!document.hasFocus()) {
		return false;
	}
	const focused = document.activeElement;
	return (
		focused === null ||
		focused === document.body ||
		!focused.isConnected ||
		(focused instanceof HTMLButtonElement && focused.disabled)
	);
}
