/**
 * Linking is how the links of a screen open a page, through the chat the card
 * is drawn in: what opens one, the pages the chat did not open, and the line
 * said over the address of one of those.
 */
export type Linking = {
	open: (href: string) => void;
	refused: ReadonlySet<string>;
	note: string;
};

// middleButton is the button of a mouse that opens a link in a tab of its own,
// as an event numbers it.
const middleButton = 1;

/**
 * PageLink is a name that opens a page through the chat: a link that leaves
 * the card where it is whatever is pressed, since the chat is what opens a
 * page — a press, or the middle button, asks the chat, and the browser's own
 * menu opens the address as it opens any link's, where the chat lets it.
 * Once the chat did not open the page, the address is shown under the name,
 * as text to copy — left to right in every language, broken anywhere to fit,
 * and taken whole at a touch — and the link stays, to try again. The line
 * that says so is in a place of its own under the name from the start, so
 * that a screen reader hears it once it is filled.
 */
export function PageLink({
	label,
	href,
	linking,
}: {
	label: string;
	href: string;
	linking: Linking;
}) {
	return (
		<span class="mt-page-link">
			<a
				class="mt-link"
				href={href}
				target="_blank"
				rel="noopener noreferrer"
				onClick={(event) => {
					event.preventDefault();
					linking.open(href);
				}}
				onAuxClick={(event) => {
					if (event.button === middleButton) {
						event.preventDefault();
						linking.open(href);
					}
				}}
			>
				{label}
			</a>
			<span class="mt-link-refused" aria-live="polite">
				{linking.refused.has(href) && (
					<>
						<span>{linking.note}</span>
						<span class="mt-link-address" dir="ltr">
							{href}
						</span>
					</>
				)}
			</span>
		</span>
	);
}
