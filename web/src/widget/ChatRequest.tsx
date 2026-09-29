import { useRef, useState } from "preact/hooks";
import type { Host } from "./bridge";
import { useWords } from "./words";

/**
 * Request is where something a card asked the chat for stands: not asked
 * yet, on its way, taken by the chat, or refused by it and free to be asked
 * again.
 */
export type Request = "idle" | "sending" | "sent" | "lost";

/**
 * useChatRequest is something a card asks the chat for, in words the child's
 * or the adult's own message would use — only the model can change a profile
 * or make one — and where the ask stands. A second press while one is on its
 * way is the same ask.
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
 * or not, to be asked again. It is on the page before it says anything, so
 * that a screen reader hears it when it does.
 */
export function RequestNote({ state }: { state: Request }) {
	const words = useWords();
	let note = "";
	if (state === "sent") {
		note = words.text("chat.sent");
	} else if (state === "lost") {
		note = words.text("chat.not_sent");
	}
	return (
		<p class="mt-action-note" aria-live="polite">
			{note}
		</p>
	);
}
