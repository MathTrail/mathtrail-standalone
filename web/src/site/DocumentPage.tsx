import type { Head } from "./Layout";
import { type PageFrame, SitePage } from "./SitePage";

/**
 * DocumentPage is a page that is read: a locale's document — its privacy
 * policy, its terms — set in the site's frame. The text arrives as HTML made
 * from the site's own Markdown, which carries no markup of its own.
 */
export function DocumentPage({
	head,
	frame,
	html,
}: {
	head: Head;
	frame: PageFrame;
	html: string;
}) {
	return (
		<SitePage head={head} frame={frame}>
			<div class="s-wrap">
				<article class="s-doc" dangerouslySetInnerHTML={{ __html: html }} />
			</div>
		</SitePage>
	);
}
