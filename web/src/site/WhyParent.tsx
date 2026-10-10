import type { ComponentChildren } from "preact";
import type { PageReader } from "./reader";
import { Honest, Stroke } from "./WhyParts";

// needs are the marks of what a parent needs, in the order the words list
// them: a heart for the child's curiosity, a compass for what it is for, a
// clock for the parent's time.
const needs = ["heart", "compass", "clock"] as const;

// talks are the moments of a lesson the page tells a parent how to answer, by
// the names their words have, each with the colour and the mark of its tag:
// the child asks what for, gets it wrong, gets it right.
const talks = [
	["why", "question"],
	["wrong", "warning"],
	["right", "star"],
] as const;

/**
 * Parent is what a parent needs to prepare a child, up to the last grade
 * MathTrail is for: a note on what MathTrail does not replace; what the parent
 * does not need, crossed out, beside what MathTrail takes on, ticked; what it
 * takes from the parent, each under its mark and number; and how to answer a
 * child at three moments of a lesson, as a few lines of a talk.
 */
export function Parent({ page, last }: { page: PageReader; last: number }) {
	const numbers = new Intl.NumberFormat(page.locale);
	return (
		<section class="s-wrap s-section">
			<div class="s-why-parent">
				<div class="s-why-parent-part">
					<h2 class="s-title">{page.text("parent.title")}</h2>
					<Honest page={page} at="parent.note" slots={{ last }} />
					<div class="s-why-lists">
						<div class="s-why-list s-why-unneeded">
							<h3>{page.text("parent.unneeded.title")}</h3>
							<ul>
								{page.list("parent.unneeded.items").map((key) => (
									<li key={key}>
										<span class="s-why-mark-no" aria-hidden="true">
											✕
										</span>
										<s>{page.text(key)}</s>
									</li>
								))}
							</ul>
						</div>
						<div class="s-why-list s-why-taken">
							<h3>{page.text("parent.taken.title")}</h3>
							<ul>
								{page.list("parent.taken.items").map((key) => (
									<li key={key}>
										<span class="s-why-mark-yes" aria-hidden="true">
											✓
										</span>
										{page.text(key)}
									</li>
								))}
							</ul>
						</div>
					</div>
				</div>
				<div class="s-why-parent-part">
					<h3 class="s-label">{page.text("parent.needed.title")}</h3>
					<ol class="s-why-needs">
						{page.list("parent.needed.items").map((key, at) => (
							<li key={key} class="s-why-need">
								<span class="s-why-need-mark">
									<NeedMark name={needs[at] ?? "heart"} />
									<span class="s-why-need-number" aria-hidden="true">
										{numbers.format(at + 1)}
									</span>
								</span>
								<div>
									<h4>{page.text(`${key}.title`)}</h4>
									<p class="s-tile-text">{page.text(`${key}.text`)}</p>
								</div>
							</li>
						))}
					</ol>
				</div>
			</div>
			<div class="s-why-explain">
				<h3>{page.text("parent.explain.title")}</h3>
				<p class="s-why-explain-lead">{page.text("parent.explain.lead")}</p>
				<ul class="s-why-talks">
					{talks.map(([talk, mark]) => (
						<Talk key={talk} page={page} talk={talk} mark={mark} />
					))}
				</ul>
			</div>
		</section>
	);
}

// Talk is one moment of a lesson: its tag, then the child's line and the
// parent's answer, each after the letter of who says it, which stands at the
// start of the child's line and at the end of the parent's, as a chat sets
// two people's messages.
function Talk({
	page,
	talk,
	mark,
}: {
	page: PageReader;
	talk: string;
	mark: (typeof talks)[number][1];
}) {
	const at = `parent.explain.talks.${talk}`;
	return (
		<li class="s-why-talk">
			<span class="s-why-moment" data-moment={talk}>
				<Stroke size={16}>{momentMarks[mark]()}</Stroke>
				{page.text(`${at}.moment`)}
			</span>
			<p class="s-why-line s-why-line-child">
				<span class="s-why-who">{page.text("parent.explain.child")}</span>
				<span class="s-why-said">{page.text(`${at}.child`)}</span>
			</p>
			<p class="s-why-line s-why-line-you">
				<span class="s-why-who">{page.text("parent.explain.you")}</span>
				<span class="s-why-said">{page.text(`${at}.you`)}</span>
			</p>
		</li>
	);
}

// momentMarks are the marks of the moments' tags: a question, a warning sign
// and a star.
const momentMarks: Readonly<
	Record<(typeof talks)[number][1], () => ComponentChildren>
> = {
	question: () => (
		<>
			<circle cx="12" cy="12" r="9" />
			<path d="M9.5 9a2.5 2.5 0 1 1 3.5 2.3c-.7.3-1 .9-1 1.7M12 17h.01" />
		</>
	),
	warning: () => (
		<>
			<path d="M12 3 2.5 20h19z" />
			<path d="M12 10v4M12 17h.01" />
		</>
	),
	star: () => (
		<path d="m12 3 2.7 5.6 6.1.9-4.4 4.3 1 6.1L12 17l-5.4 2.9 1-6.1L3.2 9.5l6.1-.9z" />
	),
};

// NeedMark is the mark of a thing a parent needs: a heart, a compass or a
// clock.
function NeedMark({ name }: { name: (typeof needs)[number] }) {
	switch (name) {
		case "heart":
			return (
				<Stroke size={22}>
					<path d="M19 14c1.5-1.5 3-3.3 3-5.5A5.5 5.5 0 0 0 16.5 3c-1.8 0-3 .5-4.5 2-1.5-1.5-2.7-2-4.5-2A5.5 5.5 0 0 0 2 8.5c0 2.3 1.5 4 3 5.5l7 7z" />
				</Stroke>
			);
		case "compass":
			return (
				<Stroke size={22}>
					<circle cx="12" cy="12" r="9" />
					<path d="m15.5 8.5-2 5-5 2 2-5z" />
				</Stroke>
			);
		case "clock":
			return (
				<Stroke size={22}>
					<circle cx="12" cy="12" r="9" />
					<path d="M12 7v5l3 2" />
				</Stroke>
			);
	}
}
