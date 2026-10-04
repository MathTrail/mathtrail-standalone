import type { ComponentChildren } from "preact";
import type { PageReader } from "./reader";
import { useSiteWords } from "./words";

/**
 * Solution is how an example a page works through is solved: the steps under
 * at, numbered, beside its drawing when it has one, then its answer. A
 * topic's page and the page of the techniques write a solution the same way.
 */
export function Solution({
	page,
	at,
	drawing,
}: {
	page: PageReader;
	at: string;
	drawing?: ComponentChildren;
}) {
	const words = useSiteWords();
	const numbers = new Intl.NumberFormat(page.locale);
	return (
		<>
			<div class="s-example-work">
				<ol class="s-example-steps">
					{page.list(`${at}.steps`).map((key, step) => (
						<li key={key} class="s-step">
							<span class="s-number" aria-hidden="true">
								{numbers.format(step + 1)}
							</span>
							<span>{page.text(key)}</span>
						</li>
					))}
				</ol>
				{drawing}
			</div>
			<p class="s-example-answer">
				<span class="s-example-answer-label">{words.text("topic.answer")}</span>{" "}
				<span>{page.text(`${at}.answer`)}</span>
			</p>
		</>
	);
}
