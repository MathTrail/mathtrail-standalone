import { createContext } from "preact";
import { useContext, useId } from "preact/hooks";

/**
 * IdScope is what the ids a card gives its parts begin with. Preact numbers
 * the ids of every root it draws from the start, so a card drawn live on a
 * page whose other cards were drawn when the page was built would take ids
 * those cards already hold. Such a page draws its live card in a scope of its
 * own. A card in a chat is the whole document, and its ids begin with nothing.
 */
export const IdScope = createContext("");

/**
 * useScopedId is an id for a part of a card that no other element of the
 * document holds: the one Preact gives, behind the scope the card is drawn in.
 */
export function useScopedId(): string {
	return useContext(IdScope) + useId();
}
