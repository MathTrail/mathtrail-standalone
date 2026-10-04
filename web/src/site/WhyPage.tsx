import type { Words } from "../i18n/words";
import { topicName } from "../widget/names";
import { cardWords } from "../widget/words";
import { address } from "./addresses";
import { sourceURL } from "./brand";
import { frontPage } from "./content";
import type { PageProps } from "./pages";
import type { Fill, PageReader } from "./reader";
import { StaticAnswer } from "./StaticCard";
import { TableFrame } from "./TableFrame";
import type { Topics } from "./topics";
import { answerOf, doiAddress, doiShown, type Source, type Why } from "./why";
import { gradesText, type SiteKey, useSiteWords } from "./words";

/**
 * WhyPage is the page that tells a parent why olympiad maths is worth a
 * child's time. It opens with a school sum beside an olympiad problem; then
 * what research has found, each finding under its work and with what it does
 * not prove said plainly; how a school lesson and an olympiad problem differ,
 * and why school has no time for the second; what a parent needs and does not
 * need to prepare a child; how MathTrail goes about it, beside the widget's
 * own card of a wrong answer and the chat under it; why it is no ordinary app;
 * how to start; and the works it cites. The grades are the catalog's, the
 * topics' names and the card's words the widget's, and the works the site's
 * data.
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
			<Research page={page} why={why} />
			<School page={page} />
			<NoTime page={page} />
			<Parent page={page} last={grades[1]} />
			<Thinking page={page} why={why} />
			<App page={page} why={why} />
			<Ask page={page} />
			<Sources page={page} why={why} />
		</>
	);
}

// gradesOfAll are the first and the last grade any topic of the catalog is
// taught at: the grades MathTrail is for.
function gradesOfAll(topics: Topics): [number, number] {
	const grades = topics.all.flatMap((topic) => topic.grades);
	return [Math.min(...grades), Math.max(...grades)];
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
	const { card } = why;
	return (
		<section class="s-wrap s-hero">
			<div class="s-hero-copy">
				<h1>{page.text("hero.title")}</h1>
				<p class="s-lead">{page.text("hero.lead", { last: grades[1] })}</p>
				<p class="s-chips">
					<span class="s-chip">{gradesText(words, grades)}</span>
					<span class="s-chip">{page.text("hero.free")}</span>
					<span class="s-chip">{page.text("hero.open")}</span>
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
					<pre class="s-drawing" dir="ltr">
						{card.drawing}
					</pre>
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

// Research is what studies have found about olympiad maths, a finding for each
// work the site's data names, in its order, each linked to its work; then a
// note on what they do not prove.
function Research({ page, why }: { page: PageReader; why: Why }) {
	return (
		<section class="s-wrap s-section">
			<div class="s-intro">
				<h2>{page.text("research.title")}</h2>
				<p class="s-intro-line">{page.text("research.lead")}</p>
			</div>
			<ul class="s-tiles s-tiles-two">
				{why.findings.map((source) => {
					const at = `research.findings.${source.id}`;
					return (
						<li key={source.id} class="s-tile s-finding">
							<h3>{page.text(`${at}.title`)}</h3>
							<p class="s-tile-text">{page.text(`${at}.text`)}</p>
							<p class="s-cite">
								<Cite source={source} />
							</p>
						</li>
					);
				})}
			</ul>
			<Honest page={page} at="research.note" />
		</section>
	);
}

// Honest is a note that says plainly what the page does not promise.
function Honest({
	page,
	at,
	slots,
}: {
	page: PageReader;
	at: string;
	slots?: Readonly<Record<string, Fill>>;
}) {
	return (
		<aside class="s-note">
			<p class="s-note-title">{page.text("honest")}</p>
			<p class="s-note-text">{page.text(at, slots)}</p>
		</aside>
	);
}

// School is how a school lesson and an olympiad problem differ, as a table
// with the olympiad's column marked.
function School({ page }: { page: PageReader }) {
	return (
		<section class="s-wrap s-section">
			<div class="s-intro">
				<h2 id="school-title">{page.text("school.title")}</h2>
				<p class="s-intro-line">{page.text("school.lead")}</p>
			</div>
			<TableFrame labelledBy="school-title">
				<table class="s-table s-versus">
					<thead>
						<tr>
							<td />
							<th scope="col">{page.text("school.lesson")}</th>
							<th scope="col" class="s-versus-olympiad">
								{page.text("school.olympiad")}
							</th>
						</tr>
					</thead>
					<tbody>
						{page.list("school.rows").map((row) => (
							<tr key={row}>
								<th scope="row">{page.text(`${row}.label`)}</th>
								<td>{page.text(`${row}.lesson`)}</td>
								<td class="s-versus-olympiad">
									{page.text(`${row}.olympiad`)}
								</td>
							</tr>
						))}
					</tbody>
				</table>
			</TableFrame>
		</section>
	);
}

// NoTime is why school has no time for olympiad problems, card by card.
function NoTime({ page }: { page: PageReader }) {
	const numbers = new Intl.NumberFormat(page.locale, {
		minimumIntegerDigits: 2,
	});
	return (
		<section class="s-wrap s-section">
			<h2 class="s-title">{page.text("time.title")}</h2>
			<ol class="s-tiles">
				{page.list("time.cards").map((key, at) => (
					<li key={key} class="s-tile">
						<span class="s-tip-number" aria-hidden="true">
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

// Parent is what a parent needs to prepare a child, up to the last grade
// MathTrail is for, and what they do not need; then how to explain to the
// child what it is for.
function Parent({ page, last }: { page: PageReader; last: number }) {
	const numbers = new Intl.NumberFormat(page.locale);
	return (
		<section class="s-wrap s-section">
			<div class="s-split">
				<div class="s-split-part">
					<h2 class="s-title">{page.text("parent.title")}</h2>
					<Honest page={page} at="parent.note" slots={{ last }} />
					<h3 class="s-label">{page.text("parent.unneeded.title")}</h3>
					<ul class="s-chips s-unneeded">
						{page.list("parent.unneeded.items").map((key) => (
							<li key={key} class="s-chip">
								{page.text(key)}
							</li>
						))}
					</ul>
					<p class="s-tile-text">{page.text("parent.unneeded.note")}</p>
				</div>
				<div class="s-split-part">
					<h3 class="s-label">{page.text("parent.needed.title")}</h3>
					<ol class="s-needs">
						{page.list("parent.needed.items").map((key, at) => (
							<li key={key} class="s-need">
								<span class="s-number" aria-hidden="true">
									{numbers.format(at + 1)}
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
			<div class="s-explain">
				<h3>{page.text("parent.explain.title")}</h3>
				<ul class="s-quotes">
					{page.list("parent.explain.quotes").map((key) => (
						<li key={key} class="s-quote">
							{page.text(key)}
						</li>
					))}
				</ul>
			</div>
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
					<Chat page={page} />
				</div>
			</div>
		</section>
	);
}

// Chat is the child's question and the model's reply as messages of the chat
// under the card, labelled as an illustration: a card never holds the model's
// words.
function Chat({ page }: { page: PageReader }) {
	return (
		<figure class="s-chat">
			<figcaption>{page.text("thinking.chat.label")}</figcaption>
			<div class="s-message s-message-child">
				<span class="s-hidden">{page.text("thinking.chat.child")}</span>
				<p class="s-bubble">{page.text("thinking.chat.question")}</p>
			</div>
			<div class="s-message s-message-model">
				<span class="s-hidden">{page.text("thinking.chat.model")}</span>
				<p class="s-bubble">{page.text("thinking.chat.reply")}</p>
			</div>
		</figure>
	);
}

// App is why MathTrail is no ordinary app: what research found of the apps
// children play, then card by card how MathTrail differs, and the code where
// that can be checked.
function App({ page, why }: { page: PageReader; why: Why }) {
	return (
		<section class="s-wrap s-section">
			<div class="s-intro">
				<h2>{page.text("app.title")}</h2>
				<p class="s-intro-line">
					{page.text("app.lead", { source: <Cite source={why.apps} /> })}
				</p>
			</div>
			<ul class="s-tiles s-tiles-two">
				{page.list("app.cards").map((key) => (
					<li key={key} class="s-tile">
						<h3>{page.text(`${key}.title`)}</h3>
						<p class="s-tile-text">{page.text(`${key}.text`)}</p>
					</li>
				))}
			</ul>
			<p class="s-code">
				<span>{page.text("app.code")}</span>
				<a class="s-btn" href={sourceURL}>
					{page.text("app.github")}
				</a>
			</p>
		</section>
	);
}

// Ask is how to start: connecting MathTrail, which the home page tells, and
// every topic.
function Ask({ page }: { page: PageReader }) {
	return (
		<section class="s-section s-ask-wrap">
			<div class="s-ask">
				<h2 class="s-ask-title">{page.text("ask.title")}</h2>
				<p class="s-ask-lead">{page.text("ask.lead")}</p>
				<p class="s-choices">
					<a class="s-btn s-btn-filled" href={address(page.locale, frontPage)}>
						{page.text("ask.connect")}
					</a>
					<a class="s-btn" href={address(page.locale, "topics")}>
						{page.text("ask.topics")}
					</a>
				</p>
			</div>
		</section>
	);
}

// Sources are the works the page cites, in the order it first cites them,
// each as its journal prints it and linked by its DOI. A screen reader is told
// the entry is English, as every work is, and told the language of a journal
// whose name is not.
function Sources({ page, why }: { page: PageReader; why: Why }) {
	return (
		<section class="s-wrap s-section s-sources">
			<h2 class="s-sources-title">{page.text("sources.title")}</h2>
			<ol>
				{why.sources.map((source) => (
					<li key={source.id} lang="en">
						{`${source.authors.join(", ")} (${source.year}). `}
						{sentence(source.title)}{" "}
						<em lang={source.journal_language}>{source.journal}</em>
						{`, ${source.volume}(${source.issue}), ${source.pages}. `}
						<a href={doiAddress(source)}>{doiShown(source)}</a>
					</li>
				))}
			</ol>
		</section>
	);
}

// sentence is a title closed as a sentence: with a full stop, unless it ends
// in a mark of its own, as a question does.
function sentence(title: string): string {
	return /[.?!]$/.test(title) ? title : `${title}.`;
}

// Cite is a work as a sentence cites it, its first author and its year, linked
// to the work itself.
function Cite({ source }: { source: Source }) {
	const words = useSiteWords();
	return <a href={doiAddress(source)}>{citeText(words, source)}</a>;
}

// citeText is how words cite source: by its one or two authors, or by the
// first of more, and its year, as the language writes a citation.
function citeText(words: Words<SiteKey>, source: Source): string {
	const [first = "", second = ""] = source.authors.map(familyOf);
	const year = String(source.year);
	switch (source.authors.length) {
		case 1:
			return words.text("cite.one", { first, year });
		case 2:
			return words.text("cite.two", { first, second, year });
		default:
			return words.text("cite.many", { first, year });
	}
}

// familyOf is the family name of an author written "Family, I.".
function familyOf(author: string): string {
	return author.split(",", 1)[0] ?? author;
}
