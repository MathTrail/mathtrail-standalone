import type { ComponentChildren } from "preact";
import type { PageReader } from "./reader";

/**
 * WorkedSteps are the steps of a worked solution on white, numbered, each
 * beside its picture where it has one. The page's words give the steps under
 * at, and pictureOf the picture of a step, counted from 0, or nothing; name
 * names the frames the pictures stand in. A picture is a drawing a screen
 * reader passes over: the step's words say it. The page of the techniques
 * and a topic's page work their examples through the same way.
 */
export function WorkedSteps({
	page,
	at,
	name,
	pictureOf,
}: {
	page: PageReader;
	at: string;
	name?: string;
	pictureOf: (step: number) => ComponentChildren;
}) {
	const numbers = new Intl.NumberFormat(page.locale);
	return (
		<ol class="s-worked-steps">
			{page.list(at).map((key, step) => {
				const picture = pictureOf(step);
				return (
					<li key={key} class="s-worked-step">
						<p class="s-worked-step-text">
							<span class="s-worked-step-number" aria-hidden="true">
								{numbers.format(step + 1)}
							</span>
							<span>{page.text(key)}</span>
						</p>
						{picture !== null && picture !== undefined && picture !== false && (
							<StepFrame name={name}>{picture}</StepFrame>
						)}
					</li>
				);
			})}
		</ol>
	);
}

/**
 * StepFrame is the pale panel a step's picture stands on, beside the step's
 * words or under them on a narrow screen, its drawing in its middle. A
 * screen reader passes it over. The name it is given, if any, is the
 * stylesheet's to colour a picture by.
 */
export function StepFrame({
	name,
	children,
}: {
	name?: string;
	children: ComponentChildren;
}) {
	return (
		<div class="s-step-picture" data-technique={name} aria-hidden="true">
			<div class="s-step-drawn">{children}</div>
		</div>
	);
}

/**
 * Answer is a worked example's answer on green, behind its badge: the answer
 * in a word or a number, which a screen reader passes over since the answer's
 * words follow it. A worked example whose answer has no short form shows no
 * badge.
 */
export function Answer({
	label,
	badge,
	children,
}: {
	label: ComponentChildren;
	badge?: string;
	children: ComponentChildren;
}) {
	return (
		<p class="s-answer">
			<span class="s-answer-label">{label}</span>{" "}
			<span class="s-answer-pill">
				{badge !== undefined && (
					<span class="s-answer-badge" aria-hidden="true">
						{badge}
					</span>
				)}
				<span>{children}</span>
			</span>
		</p>
	);
}
