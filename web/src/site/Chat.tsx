import type { ComponentChildren } from "preact";
import { Icon } from "../design/icons";
import type { PageReader } from "./reader";

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
				<Icon name="chat" size={16} />
				<span>{title}</span>
			</p>
			<div class="s-frame-body">
				<div class="s-message s-message-adult">
					<p class="s-bubble">{ask}</p>
				</div>
				{children}
			</div>
		</div>
	);
}

/**
 * AdultAsks is the adult's question about a card and the model's reply, as
 * messages of the chat beneath the card labelled as an illustration, from the
 * page's words under at: the label, who speaks, the question and the reply.
 * The adult types in the chat, and the child answers on the card.
 */
export function AdultAsks({ page, at }: { page: PageReader; at: string }) {
	return (
		<ChatLines
			label={page.text(`${at}.label`)}
			lines={[
				{
					from: "adult",
					speaker: page.text(`${at}.adult`),
					said: page.text(`${at}.question`),
				},
				{
					from: "model",
					speaker: page.text(`${at}.model`),
					said: page.text(`${at}.reply`),
				},
			]}
		/>
	);
}

// ChatLine is one message of a chat under a card: who says it, which a screen
// reader hears and the eye sees by the side the message stands on, and what
// is said.
type ChatLine = {
	readonly from: "adult" | "model";
	readonly speaker: ComponentChildren;
	readonly said: ComponentChildren;
};

// ChatLines are messages of the chat beneath a card, labelled as an
// illustration: a card never holds the model's words, which are the chat's.
// The adult's message stands at the end of the line and the model's at its
// start, as a chat sets them.
function ChatLines({
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
