import type { Ref } from "preact";
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
 * useChatRequest is something a card asks the chat for, in words the adult's
 * own message would use — only the model can write the next task — and where
 * the ask stands. A second press while one is on its way is the same ask.
 * busy says, at the moment it is asked, whether one is on its way. An ask
 * first asks stillOn whether the card may still ask, and one it turns down
 * goes nowhere, the card left as it was before the press; stillOn with
 * nothing to ask, or a check that fails, lets it go at once. sendChecked asks
 * on behalf of a card that has just been asked that, after a choice of the
 * topic.
 */
export function useChatRequest(
	host: Host,
	stillOn: () => Promise<boolean> | undefined,
): {
	state: Request;
	send: (words: string) => void;
	sendChecked: (words: string) => void;
	busy: () => boolean;
} {
	const [state, setState] = useState<Request>("idle");
	const sending = useRef(false);
	function deliver(words: string) {
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
	// begin locks the ask as it is pressed, the check included: an answer, a
	// choice or a second ask then would race it. It is false for a second
	// press while the first is on its way.
	function begin(): boolean {
		if (sending.current) {
			return false;
		}
		sending.current = true;
		setState("sending");
		return true;
	}
	function send(words: string) {
		if (!begin()) {
			return;
		}
		const asked = stillOn();
		if (asked === undefined) {
			deliver(words);
			return;
		}
		asked.then(
			(on) => {
				if (on) {
					deliver(words);
					return;
				}
				sending.current = false;
				setState("idle");
			},
			() => deliver(words),
		);
	}
	function sendChecked(words: string) {
		if (begin()) {
			deliver(words);
		}
	}
	return { state, send, sendChecked, busy: () => sending.current };
}

/**
 * useModelLines is how a card tells the model a line. The host keeps one line
 * and reads it with the next message, so a later line carries the earlier ones
 * along. A message the chat takes is the card's last word, and one it refuses
 * carried nothing, so no message lets a line go.
 */
export function useModelLines(host: Host): (line: string) => Promise<void> {
	const held = useRef<string | undefined>(undefined);
	return (line) => {
		held.current =
			held.current === undefined ? line : `${held.current}\n\n${line}`;
		return host.tellModel(held.current);
	};
}

/**
 * lineWait is how long a card waits for the host to take the model's line
 * before the card's message goes, in milliseconds: the model reads the line
 * with the message, and a host that keeps no line, or never says so, still
 * takes the message.
 */
export const lineWait = 500;

/** within is promise, waited for no longer than ms. */
export function within(ms: number, promise: Promise<void>): Promise<void> {
	return Promise.race([
		promise,
		new Promise<void>((resolve) => {
			setTimeout(resolve, ms);
		}),
	]);
}

/**
 * RequestNote says where an ask stands once the chat has answered it: taken,
 * in the words given for it, or not, to be asked again; or, on a card closed
 * because its task is no longer the one on the child's card, that it is
 * closed and where the newest task is. A host that takes a message may still
 * hold it for the person to send, so an ask whose result the card can name
 * says that rather than that it was sent. It is on the page before it says
 * anything, so that a screen reader hears it when it does, and it takes the
 * focus, through noteRef, once the ask is taken or the card closed, and the
 * button pressed for it gone.
 */
export function RequestNote({
	state,
	taken,
	noteRef,
	closed = false,
}: {
	state: Request;
	taken: Key;
	noteRef?: Ref<HTMLParagraphElement>;
	closed?: boolean;
}) {
	const words = useWords();
	let note = "";
	if (closed) {
		note = words.text("task.closed");
	} else if (state === "sent") {
		note = words.text(taken);
	} else if (state === "lost") {
		note = words.text("chat.not_sent");
	}
	return (
		<p
			class="mt-action-note"
			aria-live="polite"
			ref={noteRef}
			tabIndex={closed || state === "sent" ? -1 : undefined}
		>
			{note}
		</p>
	);
}
