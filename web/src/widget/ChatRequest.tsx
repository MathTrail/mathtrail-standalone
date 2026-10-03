import { useRef, useState } from "preact/hooks";
import type { Host } from "./bridge";
import { type Key, useWords } from "./words";

/**
 * Request is where something a card asked the chat for stands: not asked
 * yet, on its way, taken by the chat, or refused by it and free to be asked
 * again.
 */
export type Request = "idle" | "sending" | "sent" | "lost";

/**
 * useChatRequest is something a card asks the chat for, in words the child's
 * own message would use — only the model can write the next task — and where
 * the ask stands. A second press while one is on its way is the same ask.
 */
export function useChatRequest(host: Host): {
	state: Request;
	send: (words: string) => void;
} {
	const [state, setState] = useState<Request>("idle");
	const sending = useRef(false);
	function send(words: string) {
		if (sending.current) {
			return;
		}
		sending.current = true;
		setState("sending");
		host.sendMessage(words).then(
			() => {
				sending.current = false;
				setState("sent");
			},
			(error: unknown) => {
				console.error("widget: the ask did not reach the chat", error);
				sending.current = false;
				setState("lost");
			},
		);
	}
	return { state, send };
}

/**
 * RequestNote says where an ask stands once the chat has answered it: taken,
 * in the words given for it, or not, to be asked again. A host that takes a
 * message may still hold it for the person to send, so an ask whose result
 * the card can name says that rather than that it was sent. It is on the page
 * before it says anything, so that a screen reader hears it when it does.
 */
export function RequestNote({ state, taken }: { state: Request; taken: Key }) {
	const words = useWords();
	let note = "";
	if (state === "sent") {
		note = words.text(taken);
	} else if (state === "lost") {
		note = words.text("chat.not_sent");
	}
	return (
		<p class="mt-action-note" aria-live="polite">
			{note}
		</p>
	);
}
