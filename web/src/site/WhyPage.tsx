import { Diagram } from "../design/picture/diagram";
import { cardWords } from "../widget/dictionaries";
import { topicName } from "../widget/names";
import { address } from "./addresses";
import { AdultAsks } from "./Chat";
import type { PageProps } from "./pages";
import type { PageReader } from "./reader";
import { SourceChip } from "./SourceChip";
import { StaticAnswer } from "./StaticCard";
import { gradesOfAll } from "./topics";
import { answerOf, type Why } from "./why";
import { App } from "./WhyApp";
import { Research } from "./WhyFindings";
import { History } from "./WhyHistory";
import { Parent } from "./WhyParent";
import { NoTime, School } from "./WhySchool";
import { gradesText, useSiteWords } from "./words";

// whyScript is where the page's script is served: one module, which the
// site's build makes of web/src/demo/why.ts, and which moves the pictures of
// the history on by themselves.
const whyScript = "/assets/why.js";

/**
 * WhyPage is the page that tells a parent why olympiad maths is worth a
 * child's time. It opens with a school sum beside an olympiad problem; then
 * how long people have learned to find a way with no example, era by era in
 * pictures; what research has found, each finding beside its work; how a
 * school lesson and an olympiad problem differ, and why school has no time
 * for the second; what a parent needs and does not need to prepare a child;
 * how MathTrail goes about it, beside the widget's own card of a wrong answer
 * and the chat under it; and why it is no ordinary app. The grades are the
 * catalog's, the topics' names and the card's words the widget's, and the
 * works it cites and the history's pictures the site's data. Its script,
 * where it runs, moves the pictures on; without it a chip of an era picks
 * each.
 */
export function WhyPage({ page, data }: PageProps) {
	const why = data.why;
	if (why === undefined) {
		throw new Error(
			"site/data.json has nothing for the page Why: its card and the works it cites",
		);
	}
	const grades = gradesOfAll(data.topics);
	return (
		<>
			<Hero page={page} why={why} grades={grades} />
			<History page={page} eras={why.history} />
			<Research page={page} findings={why.findings} />
			<School page={page} />
			<NoTime page={page} answer={why.card.options[why.card.correct] ?? ""} />
			<Parent page={page} last={grades[1]} />
			<Thinking page={page} why={why} />
			<App page={page} apps={why.apps} />
			<script type="module" src={whyScript} />
		</>
	);
}

// Hero is the first screen: why olympiad maths, the grades MathTrail is for,
// and a school sum beside an olympiad problem on the same numbers, with the
// problem's answer and the trap in it.
function Hero({
	page,
	why,
	grades,
}: {
	page: PageReader;
	why: Why;
	grades: readonly [number, number];
}) {
	const words = useSiteWords();
	const said = cardWords(page.locale, undefined);
	const { card } = why;
	return (
		<section class="s-wrap s-hero">
			<div class="s-hero-copy">
				<h1>{page.text("hero.title")}</h1>
				<p class="s-lead">{page.text("hero.lead", { last: grades[1] })}</p>
				<p class="s-chips">
					<span class="s-chip s-chip-free">{page.text("hero.free")}</span>
					<SourceChip label={page.text("hero.open")} />
					<span class="s-chip">{gradesText(words, grades)}</span>
				</p>
			</div>
			<figure class="s-panel s-contrast">
				<div class="s-contrast-card">
					<p class="s-contrast-label">{page.text("hero.school.label")}</p>
					<p class="s-contrast-text">{page.text("hero.school.text")}</p>
					<p class="s-contrast-sum">{page.plain("hero.school.sum")}</p>
				</div>
				<div class="s-contrast-card s-contrast-raised">
					<p class="s-contrast-label">{page.text("hero.olympiad.label")}</p>
					<p class="s-contrast-text">{page.plain("task.question")}</p>
					<Diagram
						picture={card.picture}
						label={said.text(`picture.${card.picture.kind}`)}
						locale={said.locale}
					/>
					<p class="s-chips">
						<span class="s-chip s-chip-right">
							{page.text("hero.olympiad.answer", {
								answer: card.options[card.correct] ?? "",
							})}
						</span>
						<span class="s-chip s-chip-trap">
							{page.text("hero.olympiad.trap")}
						</span>
					</p>
					<p class="s-contrast-note">{page.text("hero.olympiad.text")}</p>
				</div>
				<figcaption>{page.text("hero.caption")}</figcaption>
			</figure>
		</section>
	);
}

// Thinking is how MathTrail goes about it, point by point, beside the
// widget's own card of a wrong answer, drawn from the site's data with the
// page's words, and the chat under it.
function Thinking({ page, why }: { page: PageReader; why: Why }) {
	const card = cardWords(page.locale, undefined);
	const numbers = new Intl.NumberFormat(page.locale, {
		minimumIntegerDigits: 2,
	});
	const topics = Object.fromEntries(
		Object.entries(why.topics).map(([slot, id]) => [slot, topicName(card, id)]),
	);
	const shown = answerOf(why.card, {
		language: page.locale,
		child: page.plain("thinking.child"),
		question: page.plain("task.question"),
		trap: page.plain("task.trap"),
		solution: page.plain("task.solution"),
	});
	return (
		<section class="s-wrap s-section">
			<div class="s-split">
				<div class="s-split-part">
					<div class="s-intro">
						<h2>{page.text("thinking.title")}</h2>
						<p class="s-intro-line">{page.text("thinking.lead")}</p>
					</div>
					<ol class="s-points">
						{page.list("thinking.points").map((key, at) => (
							<li key={key} class="s-point">
								<span class="s-tip-number" aria-hidden="true">
									{numbers.format(at + 1)}
								</span>
								<div>
									<h3>{page.text(`${key}.title`)}</h3>
									<p class="s-tile-text">{page.text(`${key}.text`, topics)}</p>
								</div>
							</li>
						))}
					</ol>
					<p>
						<a href={address(page.locale, "topics")}>
							{page.text("thinking.topics")}
						</a>
					</p>
				</div>
				<div class="s-panel s-lesson">
					<figure class="s-hero-card">
						<StaticAnswer
							handed={shown.handed}
							result={shown.result}
							locale={page.locale}
						/>
						<figcaption>{page.text("thinking.caption")}</figcaption>
					</figure>
					<AdultAsks page={page} at="thinking.chat" />
				</div>
			</div>
		</section>
	);
}
