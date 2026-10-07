import type { ComponentChildren } from "preact";
import type { PageReader } from "./reader";
import { type SectionId, serviceSections, titleId } from "./service";

/**
 * ServiceSection is one numbered part of the page "Service": its number in
 * the page's order, its heading and the lines that open it, then what it
 * draws. Its id is the anchor it is reached by, and the prefix of its words.
 */
export function ServiceSection({
	page,
	id,
	children,
}: {
	page: PageReader;
	id: SectionId;
	children: ComponentChildren;
}) {
	const two = new Intl.NumberFormat(page.locale, { minimumIntegerDigits: 2 });
	const title = titleId(id);
	return (
		<section id={id} class="s-wrap s-service-section" aria-labelledby={title}>
			<p class="s-service-number">
				{two.format(serviceSections.indexOf(id) + 1)}
			</p>
			<h2 id={title}>{page.text(`${id}.title`)}</h2>
			<p class="s-service-lead">{page.text(`${id}.lead`)}</p>
			{children}
		</section>
	);
}
