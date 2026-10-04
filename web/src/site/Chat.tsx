import type { ComponentChildren } from "preact";

/**
 * ChatFrame is a card of the widget as a chat a parent already uses shows it:
 * a frame under the chat's name, the parent's message, and below it the card
 * and whatever the chat says after it.
 */
export function ChatFrame({
	title,
	ask,
	children,
}: {
	title: ComponentChildren;
	ask: ComponentChildren;
	children: ComponentChildren;
}) {
	return (
		<div class="s-frame">
			<p class="s-frame-head">
				<ChatIcon />
				<span>{title}</span>
			</p>
			<div class="s-frame-body">
				<div class="s-message s-message-child">
					<p class="s-bubble">{ask}</p>
				</div>
				{children}
			</div>
		</div>
	);
}

/**
 * ChatLine is one message of a chat under a card: who says it, which a screen
 * reader hears and the eye sees by the side the message stands on, and what
 * is said.
 */
export type ChatLine = {
	readonly from: "child" | "model";
	readonly speaker: ComponentChildren;
	readonly said: ComponentChildren;
};

/**
 * ChatLines are messages of the chat beneath a card, labelled as an
 * illustration: a card never holds the model's words, which are the chat's.
 * The child's message stands at the end of the line and the model's at its
 * start, as a chat sets them.
 */
export function ChatLines({
	label,
	lines,
}: {
	label: ComponentChildren;
	lines: readonly ChatLine[];
}) {
	return (
		<figure class="s-chat">
			<figcaption>{label}</figcaption>
			{lines.map(({ from, speaker, said }, at) => (
				<div key={at} class={`s-message s-message-${from}`}>
					<span class="s-hidden">{speaker}</span>
					<p class="s-bubble">{said}</p>
				</div>
			))}
		</figure>
	);
}

// ChatIcon is a speech bubble, the mark of a chat, in the colour of the words
// beside it.
function ChatIcon() {
	return (
		<svg
			width="16"
			height="16"
			viewBox="0 0 24 24"
			fill="none"
			stroke="currentColor"
			stroke-width="2"
			stroke-linecap="round"
			stroke-linejoin="round"
			aria-hidden="true"
		>
			<path d="M21 12a8 8 0 0 1-11.6 7.1L4 20l1-4.6A8 8 0 1 1 21 12Z" />
		</svg>
	);
}
