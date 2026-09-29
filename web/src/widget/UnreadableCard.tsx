import { Verdict } from "../design/blocks";
import { CardHeader, CardRoot } from "./CardRoot";
import { useWords } from "./words";

/**
 * UnreadableCard is the card of a payload the widget cannot draw: one that
 * names no screen a card shows, or does not read as the one it names. Every
 * tool says the same in words for the chat, so the card says as much rather
 * than show what arrived.
 */
export function UnreadableCard() {
	const words = useWords();
	return (
		<CardRoot>
			{(wide) => (
				<article aria-label={words.text("app.name")}>
					<CardHeader grade={undefined} wide={wide} />
					<div class="mt-body">
						<Verdict>{words.text("card.unreadable")}</Verdict>
					</div>
				</article>
			)}
		</CardRoot>
	);
}
