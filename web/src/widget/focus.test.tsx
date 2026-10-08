import { type ComponentChildren, render } from "preact";
import { useRef } from "preact/hooks";
import { act } from "preact/test-utils";
import { afterEach, describe, expect, test, vi } from "vitest";
import { CardIsTheDocument, useFocusKeptOnTheCard } from "./focus";

const root = document.createElement("div");
document.body.append(root);

afterEach(() => {
	act(() => render(null, root));
	vi.restoreAllMocks();
});

// Card is a card whose one thing left to do, once done is true, is the
// button it gives the focus to.
function Card({ done }: { done: boolean }) {
	const next = useRef<HTMLButtonElement>(null);
	useFocusKeptOnTheCard(done, next);
	return (
		<button type="button" ref={next}>
			Another task
		</button>
	);
}

// draw draws the card, done or not, inside around.
function draw(
	done: boolean,
	around: (card: ComponentChildren) => ComponentChildren = (card) => card,
): void {
	act(() => render(around(<Card done={done} />), root));
}

describe("a card whose focus is on nothing once its answer is in", () => {
	test("takes the focus back where it is the whole document, as in a chat", () => {
		draw(false);
		expect(document.activeElement).toBe(document.body);

		draw(true);

		expect(document.activeElement?.textContent).toBe("Another task");
	});

	// The card's frame holds no focus while the adult types in the chat, and
	// an answer that comes meanwhile does not take the focus from the chat.
	test("leaves it where it is when the card's document does not hold the focus", () => {
		vi.spyOn(document, "hasFocus").mockReturnValue(false);
		draw(false);

		draw(true);

		expect(document.activeElement).toBe(document.body);
	});

	// On a page a focus on nothing is where the reader left it: the card is
	// one part of the page, and the reader may be reading another.
	test("leaves it where it is when the card is a part of a page", () => {
		const onAPage = (card: ComponentChildren) => (
			<CardIsTheDocument.Provider value={false}>
				{card}
			</CardIsTheDocument.Provider>
		);
		draw(false, onAPage);

		draw(true, onAPage);

		expect(document.activeElement).toBe(document.body);
	});
});
