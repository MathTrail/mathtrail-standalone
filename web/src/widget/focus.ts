import { createContext } from "preact";
import { useContext, useEffect } from "preact/hooks";

/**
 * CardIsTheDocument says whether a card is the whole document it is drawn in,
 * as it is in a chat's frame, rather than a part of a page. Only a card that
 * is the whole document takes back a focus it has lost. On a page, a focus on
 * nothing is where the reader left it — on the page, or in a browser that
 * gives a pressed button no focus — and from a button gone the browser takes
 * the next step itself.
 */
export const CardIsTheDocument = createContext(true);

/**
 * useFocusKeptOnTheCard gives the focus to target — or, when there is none on
 * the card, to fallback — when shown comes true and the focus the card had has
 * been lost: the button pressed has gone or been switched off. A focus still
 * on something is left where it is, a focus elsewhere in the chat is never
 * taken, and the page is not scrolled. A card that is a part of a page takes
 * nothing back.
 */
export function useFocusKeptOnTheCard(
	shown: boolean,
	target: { current: HTMLElement | null },
	fallback?: { current: HTMLElement | null },
): void {
	const whole = useContext(CardIsTheDocument);
	useEffect(() => {
		if (whole && shown && focusIsLost()) {
			(target.current ?? fallback?.current)?.focus({ preventScroll: true });
		}
	}, [whole, shown, target, fallback]);
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
