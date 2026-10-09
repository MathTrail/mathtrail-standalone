import { cardWords } from "../widget/dictionaries";
import { topicName } from "../widget/names";
import type { HomePick } from "./home";
import type { PageReader } from "./reader";
import { written } from "./ResearchNumbers";

/**
 * Pick is how the task on the first screen's card was picked for the child,
 * under the promise: the child's topics, each with a bar of how far the child
 * has come in it, the one picked marked as the current one; and why it was
 * picked — the chance of a right answer, on a bar with the corridor the rule
 * keeps that chance in, and the reasons. Like the card beside it, it is an
 * illustration: its numbers are the site's data, and its topics the
 * catalog's, named as the card names them. The bars are drawings, which a
 * screen reader skips.
 */
export function Pick({
	page,
	pick,
	picked,
}: {
	page: PageReader;
	pick: HomePick;
	picked: string;
}) {
	const card = cardWords(page.locale, undefined);
	const { chance, corridor } = pick;
	return (
		<figure class="s-pick" aria-label={page.plain("hero.pick.label")}>
			<p class="s-pick-title">
				<PickMark />
				{page.text("hero.pick.title")}
			</p>
			<div class="s-pick-grid">
				<div class="s-pick-column">
					<p class="s-pick-head">{page.text("hero.pick.skills")}</p>
					<ul class="s-pick-list">
						{pick.skills.map(({ topic, share }) => (
							<li
								key={topic}
								class="s-pick-skill"
								aria-current={topic === picked ? "true" : undefined}
							>
								{topicName(card, topic)}
								<span class="s-pick-bar" aria-hidden="true">
									<span style={{ inlineSize: along(share) }} />
								</span>
							</li>
						))}
					</ul>
				</div>
				<div class="s-pick-column">
					<p class="s-pick-head">
						{page.text("hero.pick.why")}
						<span class="s-pick-auto">{page.text("hero.pick.auto")}</span>
					</p>
					<span class="s-pick-zone" aria-hidden="true">
						<span
							class="s-pick-corridor"
							style={{
								insetInlineStart: along(corridor.low),
								inlineSize: along(corridor.high - corridor.low),
							}}
						/>
						<span
							class="s-pick-chance"
							style={{ insetInlineStart: along(chance) }}
						/>
					</span>
					<ul class="s-pick-list s-pick-reasons">
						<li>
							{page.text("hero.pick.chance", {
								chance: <strong>{written(page.locale, chance, "share")}</strong>,
							})}
						</li>
						{page.list("hero.pick.reasons").map((key) => (
							<li key={key}>
								<span aria-hidden="true">✓ </span>
								{page.text(key)}
							</li>
						))}
					</ul>
				</div>
			</div>
		</figure>
	);
}

// along is a share as a length along a bar, to a tenth of a per cent: no
// further than a share added up of others can stray.
function along(share: number): string {
	return `${Math.round(share * 1000) / 10}%`;
}

// PickMark is the mark before the block's title: a chip, the rule that does
// the picking, drawn in strokes on a 24-unit square.
function PickMark() {
	return (
		<svg
			class="s-pick-mark"
			viewBox="0 0 24 24"
			width="18"
			height="18"
			fill="none"
			aria-hidden="true"
		>
			<path
				d="M6 6h12v12H6zM9 2v4M15 2v4M9 18v4M15 18v4M2 9h4M2 15h4M18 9h4M18 15h4M10 10h4v4h-4z"
				stroke="currentColor"
				stroke-width="1.8"
				stroke-linecap="round"
				stroke-linejoin="round"
			/>
		</svg>
	);
}
