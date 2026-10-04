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

/**
 * PageLink is a name that opens a page through the chat: a link that leaves
 * the card where it is whatever is pressed, since the chat is what opens a
 * page. Once the chat did not open the page, the address is shown under the
 * name, as text to copy — left to right in every language, broken anywhere
 * to fit, and taken whole at a touch — and the link stays, to try again.
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
			>
				{label}
			</a>
			{linking.refused.has(href) && (
				<span class="mt-link-refused" role="status">
					<span>{linking.note}</span>
					<span class="mt-link-address" dir="ltr">
						{href}
					</span>
				</span>
			)}
		</span>
	);
}
