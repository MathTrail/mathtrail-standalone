import { topicName } from "../widget/names";
import { cardWords } from "../widget/words";
import { address } from "./addresses";
import type { SiteData } from "./data";
import { connectAddress } from "./home";
import type { PageProps } from "./pages";
import type { PageReader } from "./reader";
import { Solution } from "./Solution";
import { TechniqueDrawing } from "./TechniqueDrawings";
import type { Technique, Techniques } from "./techniques";
import { gradesText, useSiteWords } from "./words";

/**
 * TechniquesPage is the page of the techniques of problem solving. It opens
 * with what a technique is for and every technique by its group, each a link
 * to its part of the page; then the four steps that come before any
 * technique; then each group with its techniques, each worked through on a
 * problem whose solution, drawing and answer open on request, so that the
 * problem can be tried first; then a hint on which technique the words of a
 * problem suggest; and last, the way to the topics.
 */
export function TechniquesPage({ page, data }: PageProps) {
	const techniques = data.techniques;
	if (techniques === undefined) {
		throw new Error("site/data.json gives the page of the techniques none");
	}
	const numbers = numbersOf(page, techniques);
	return (
		<>
			<Hero page={page} techniques={techniques} numbers={numbers} />
			<Start page={page} />
			{techniques.groups.map((group) => (
				<section key={group.id} class="s-wrap s-section">
					<div class="s-intro">
						<h2>{page.text(`groups.${group.id}.title`)}</h2>
						<p class="s-intro-line">{page.text(`groups.${group.id}.lead`)}</p>
					</div>
					{group.techniques.map((technique) => (
						<TechniqueCard
							key={technique.id}
							page={page}
							data={data}
							technique={technique}
							number={numbers.get(technique.id) ?? ""}
						/>
					))}
				</section>
			))}
			<Cues page={page} techniques={techniques} numbers={numbers} />
			<End page={page} />
		</>
	);
}

// numbersOf are the numbers of the techniques, by name, in the order the page
// shows them: 01, 02 and on, written the way the page's language writes
// numbers.
function numbersOf(
	page: PageReader,
	techniques: Techniques,
): ReadonlyMap<string, string> {
	const two = new Intl.NumberFormat(page.locale, { minimumIntegerDigits: 2 });
	return new Map(
		techniques.groups
			.flatMap((group) => group.techniques)
			.map((technique, at) => [technique.id, two.format(at + 1)]),
	);
}

// Hero is the first screen: the page's heading, what a technique is for and
// how to read the page, beside every technique by its group.
function Hero({
	page,
	techniques,
	numbers,
}: {
	page: PageReader;
	techniques: Techniques;
	numbers: ReadonlyMap<string, string>;
}) {
	const words = useSiteWords();
	return (
		<section class="s-wrap s-hero">
			<div class="s-hero-copy">
				<h1>{page.text("hero.title")}</h1>
				<p class="s-lead">
					{page.text("hero.lead", {
						count: words.text("techniques.count", { count: numbers.size }),
					})}
				</p>
				<p class="s-intro-line">{page.text("hero.advice")}</p>
			</div>
			<nav class="s-panel s-overview" aria-label={words.text("topic.contents")}>
				{techniques.groups.map((group) => (
					<div key={group.id} class="s-overview-group">
						<p class="s-overview-title">
							{page.text(`groups.${group.id}.title`)}
						</p>
						<ul class="s-overview-list">
							{group.techniques.map((technique) => (
								<li key={technique.id}>
									<TechniqueLink
										page={page}
										id={technique.id}
										number={numbers.get(technique.id) ?? ""}
									/>
								</li>
							))}
						</ul>
					</div>
				))}
			</nav>
		</section>
	);
}

// TechniqueLink leads to a technique's part of the page under its number and
// its name.
function TechniqueLink({
	page,
	id,
	number,
}: {
	page: PageReader;
	id: string;
	number: string;
}) {
	return (
		<a class="s-technique-link" href={`#${id}`}>
			<span class="s-technique-link-number">{number}</span>
			{page.text(`techniques.${id}.name`)}
		</a>
	);
}

// Start is what comes before any technique: the four steps of solving a
// problem, numbered.
function Start({ page }: { page: PageReader }) {
	const numbers = new Intl.NumberFormat(page.locale);
	return (
		<section class="s-wrap s-section">
			<div class="s-intro">
				<h2>{page.text("start.title")}</h2>
				<p class="s-intro-line">{page.text("start.lead")}</p>
			</div>
			<ol class="s-tiles s-tiles-four">
				{page.list("start.steps").map((key, at) => (
					<li key={key} class="s-tile">
						<span class="s-number s-number-large" aria-hidden="true">
							{numbers.format(at + 1)}
						</span>
						<h3>{page.text(`${key}.title`)}</h3>
						<p class="s-tile-text">{page.text(`${key}.text`)}</p>
					</li>
				))}
			</ol>
		</section>
	);
}

// TechniqueCard is one technique: its number and name, what it is and when it
// helps, the problem it is worked through on, set at its example's grades,
// the solution — its steps, its drawing and its answer — folded until asked
// for, and the topics it leads to.
function TechniqueCard({
	page,
	data,
	technique,
	number,
}: {
	page: PageReader;
	data: SiteData;
	technique: Technique;
	number: string;
}) {
	const words = useSiteWords();
	const at = `techniques.${technique.id}`;
	return (
		<article id={technique.id} class="s-technique">
			<div class="s-technique-head">
				<span class="s-technique-number">{number}</span>
				<h3>{page.text(`${at}.name`)}</h3>
				{page.has(`${at}.aka`) && (
					<p class="s-technique-aka">{page.text(`${at}.aka`)}</p>
				)}
				<p class="s-technique-what">{page.text(`${at}.what`)}</p>
				<p class="s-technique-when">
					<strong>{page.text("labels.when")}</strong> {page.text(`${at}.when`)}
				</p>
			</div>
			<div class="s-technique-task">
				<p class="s-example-label">
					{page.text("labels.task", {
						grades: gradesText(words, technique.example.grades),
					})}
				</p>
				<p class="s-technique-question">{page.text(`${at}.task`)}</p>
			</div>
			<details class="s-solution">
				<summary class="s-solution-toggle">{page.text("labels.show")}</summary>
				<div class="s-solution-body">
					<Solution
						page={page}
						at={at}
						drawing={
							<figure class="s-figure">
								<div class="s-art" aria-hidden="true" dir="ltr">
									<TechniqueDrawing
										id={technique.id}
										page={page}
										at={`${at}.drawing`}
									/>
								</div>
								<figcaption>{page.text(`${at}.caption`)}</figcaption>
							</figure>
						}
					/>
				</div>
			</details>
			<TopicChips page={page} data={data} ids={technique.topics} />
		</article>
	);
}

// TopicChips are the topics a technique leads to, under the names the card
// gives them: each a link to its page once the page is published, and word
// that the page is coming until then. A technique of no topic has none.
function TopicChips({
	page,
	data,
	ids,
}: {
	page: PageReader;
	data: SiteData;
	ids: readonly string[];
}) {
	const words = useSiteWords();
	const topics = ids.flatMap((id) => {
		const topic = data.topics.byId.get(id);
		return topic === undefined ? [] : [topic];
	});
	if (topics.length === 0) {
		return null;
	}
	const card = cardWords(page.locale, undefined);
	return (
		<p class="s-technique-topics">
			<span class="s-technique-topics-label">{page.text("labels.topics")}</span>
			{topics.map((topic) =>
				topic.sitePage ? (
					<a
						key={topic.id}
						class="s-chip s-chip-link"
						href={address(page.locale, `topics/${topic.slug}`)}
					>
						{topicName(card, topic.id)}
					</a>
				) : (
					<span key={topic.id} class="s-chip">
						{topicName(card, topic.id)}{" "}
						<span class="s-soon">{words.text("topics.soon")}</span>
					</span>
				),
			)}
		</p>
	);
}

// Cues is the hint on which technique to try: the words a problem says, each
// beside the techniques they suggest. The words give as many rows as the data
// does.
function Cues({
	page,
	techniques,
	numbers,
}: {
	page: PageReader;
	techniques: Techniques;
	numbers: ReadonlyMap<string, string>;
}) {
	const rows = page.list("cues.rows");
	if (rows.length !== techniques.cues.length) {
		throw new Error(
			`the hint has ${rows.length} rows in the words of the page, and ${techniques.cues.length} in site/data.json`,
		);
	}
	return (
		<section class="s-wrap s-section">
			<div class="s-intro">
				<h2>{page.text("cues.title")}</h2>
				<p class="s-intro-line">{page.text("cues.lead")}</p>
			</div>
			<div class="s-cues">
				<p class="s-cues-head">
					<span>{page.text("cues.if")}</span>
					<span>{page.text("cues.try")}</span>
				</p>
				<dl class="s-cues-list">
					{rows.map((key, at) => (
						<div key={key} class="s-cue">
							<dt>{page.text(key)}</dt>
							<dd>
								{(techniques.cues[at] ?? []).map((id) => (
									<TechniqueLink
										key={id}
										page={page}
										id={id}
										number={numbers.get(id) ?? ""}
									/>
								))}
							</dd>
						</div>
					))}
				</dl>
			</div>
		</section>
	);
}

// End closes the page: the techniques are to be tried on problems, in a chat
// MathTrail is added to, and every topic has them.
function End({ page }: { page: PageReader }) {
	const words = useSiteWords();
	return (
		<section class="s-section s-ask-wrap">
			<div class="s-ask">
				<h2 class="s-ask-title">{page.text("end.title")}</h2>
				<p class="s-ask-lead">{page.text("end.lead")}</p>
				<p class="s-choices">
					<a class="s-btn s-btn-filled" href={connectAddress(page.locale)}>
						{words.text("nav.add")}
					</a>
					<a class="s-btn" href={address(page.locale, "topics")}>
						{words.text("topic.all")}
					</a>
				</p>
			</div>
		</section>
	);
}
