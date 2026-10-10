import type { ComponentChildren } from "preact";
import type { Words } from "../i18n/words";
import { cardWords } from "../widget/dictionaries";
import { topicName, trapName } from "../widget/names";
import type { Key } from "../widget/words";
import { address } from "./addresses";
import type { Art, CellWords, ExampleArt, TopicArt } from "./art/art";
import { topicArt } from "./art";
import type { SiteData, TopicExample } from "./data";
import type { Drawing } from "./drawings";
import { connectAddress } from "./home";
import type { Page, PageProps } from "./pages";
import type { PageReader } from "./reader";
import { Solution } from "./Solution";
import { TableFrame } from "./TableFrame";
import { TopicDrawing } from "./TopicDrawings";
import { TopicThumb } from "./TopicThumbs";
import type { Group, Topic } from "./topics";
import { gradesText, groupKey, type SiteKey, useSiteWords } from "./words";
import { Answer, StepFrame, WorkedSteps } from "./WorkedSteps";

// shownTraps is how many of a topic's traps its page explains: the most
// frequent in its reference tasks, which between them take most of the wrong
// options there are.
const shownTraps = 3;

/**
 * topicPage is the page of topic, drawn by the template every topic's page
 * shares, which loads the card's styles when it draws a picture.
 */
export function topicPage(topic: Topic, data: SiteData): Page {
	return {
		draw: (props) => <TopicPage {...props} topic={topic} />,
		card: drawsPicture(data, topic),
	};
}

// drawsPicture says whether the topic's page may draw a picture: wherever
// the topic has drawings of its own, which draw pictures of the card's kinds
// among their parts, and else on its first screen or beside an example.
function drawsPicture(data: SiteData, topic: Topic): boolean {
	return (
		topicArt.has(topic.slug) ||
		[
			data.drawings.get(topic.id)?.hero,
			...examplesOf(data, topic).map((example) => example.drawing),
		].some((drawing) => drawing !== undefined && "picture" in drawing)
	);
}

// TopicPage is a topic's page for a parent. It opens with the path to it, the
// topic's group and grades, what the topic is, its main move and what it
// teaches, beside the topic's drawing; then the ground the topic stands on,
// and its key idea; how such tasks are solved, with examples worked through to
// their answers; the traps children fall into most often in the topic, under
// the names the card gives them, with what to say; how to help at home; the
// topics to learn first and those it opens; and how to ask for a task on it.
// The words are the topic's own, the labels the site's, and the grades and the
// traps the catalog's.
function TopicPage({
	page,
	data,
	topic,
}: PageProps & { readonly topic: Topic }) {
	const card = cardWords(page.locale, undefined);
	const name = topicName(card, topic.id);
	const group = groupOf(data, topic);
	const art = topicArt.get(topic.slug);
	return (
		<>
			<section class="s-wrap s-subject">
				<Path page={page} group={group} name={name} />
				<Hero
					page={page}
					data={data}
					topic={topic}
					group={group}
					name={name}
					art={art}
				/>
				<Contents page={page} />
			</section>
			<Basis page={page} art={art?.basis} />
			<Idea
				page={page}
				art={art?.idea}
				answers={art?.ideaAnswers}
				legend={art?.ideaLegend}
			/>
			<Solving
				page={page}
				topic={topic}
				examples={examplesOf(data, topic)}
				art={art}
			/>
			<Traps
				page={page}
				traps={trapsOf(data, topic)}
				card={card}
				art={art?.traps}
			/>
			<Home page={page} />
			<Related
				heading="topic.before"
				ids={topic.bases}
				data={data}
				card={card}
			/>
			<Related
				heading="topic.after"
				ids={topic.opens}
				data={data}
				card={card}
			/>
			<Ask page={page} />
		</>
	);
}

// groupOf is the group the page of the topics shows topic in: every topic is
// in exactly one.
function groupOf(data: SiteData, topic: Topic): Group {
	const group = data.topics.groups.find(({ topics }) =>
		topics.includes(topic.id),
	);
	if (group === undefined) {
		throw new Error(`the topic ${topic.id} is in no group`);
	}
	return group;
}

// examplesOf are the examples of the topic's page, as the site's data has them.
function examplesOf(data: SiteData, topic: Topic): readonly TopicExample[] {
	return data.examples.get(topic.id) ?? [];
}

// heroDrawingOf is the drawing the first screen of the topic's page shows:
// every published page has one, or it does not build.
function heroDrawingOf(data: SiteData, topic: Topic): Drawing {
	const hero = data.drawings.get(topic.id)?.hero;
	if (hero === undefined) {
		throw new Error(
			`site/data.json gives the first screen of ${topic.id} no drawing`,
		);
	}
	return hero;
}

// trapsOf are the traps the topic's page explains: the most frequent in its
// reference tasks, the most frequent first.
function trapsOf(data: SiteData, topic: Topic): readonly string[] {
	const traps = data.traps.get(topic.id)?.slice(0, shownTraps) ?? [];
	if (traps.length === 0) {
		throw new Error(
			`no reference task of ${topic.id} names a trap, and its page explains the most frequent`,
		);
	}
	return traps;
}

// Path is the way from the page of the topics to this one, through the group
// the topic is shown in there.
function Path({
	page,
	group,
	name,
}: {
	page: PageReader;
	group: Group;
	name: string;
}) {
	const words = useSiteWords();
	const topics = address(page.locale, "topics");
	return (
		<nav class="s-path" aria-label={words.text("topic.path")}>
			<ol>
				<li>
					<a href={topics}>{words.text("nav.topics")}</a>
				</li>
				<li>
					<a href={`${topics}#${group.id}`}>{words.text(groupKey(group))}</a>
				</li>
				<li aria-current="page">{name}</li>
			</ol>
		</nav>
	);
}

// Hero is the first screen: the topic's group and grades, its name, what it is,
// its main move and what it teaches, beside its drawing: the topic's own on a
// white box, with the line over it where it has one, or the site's data's.
function Hero({
	page,
	data,
	topic,
	group,
	name,
	art,
}: {
	page: PageReader;
	data: SiteData;
	topic: Topic;
	group: Group;
	name: string;
	art: TopicArt | undefined;
}) {
	const words = useSiteWords();
	return (
		<div class="s-hero s-subject-hero">
			<div class="s-hero-copy">
				<p class="s-chips">
					<span class="s-chip s-chip-group">{words.text(groupKey(group))}</span>
					<span class="s-chip">{gradesText(words, topic.grades)}</span>
				</p>
				<h1>{name}</h1>
				<p class="s-lead">{page.text("hero.lead")}</p>
				<dl class="s-facts">
					<div>
						<dt>{words.text("topic.method")}</dt>
						<dd>{page.text("hero.method")}</dd>
					</div>
					<div>
						<dt>{words.text("topic.teaches")}</dt>
						<dd>{page.text("hero.teaches")}</dd>
					</div>
				</dl>
			</div>
			{art === undefined ? (
				<div class="s-panel s-subject-panel">
					<TopicDrawing
						drawing={heroDrawingOf(data, topic)}
						page={page}
						at="hero.drawing"
						where={`the first screen of ${topic.id}`}
					/>
				</div>
			) : (
				<div
					class="s-panel s-subject-panel s-subject-sketch"
					aria-hidden="true"
					dir="ltr"
				>
					{art.heroLine !== undefined && (
						<p class="s-subject-line">
							{art.heroLine({ page, at: "hero.art" })}
						</p>
					)}
					<div class="s-sketch s-topic-art">
						<div class="s-sketch-drawn">
							{art.hero({ page, at: "hero.art" })}
						</div>
					</div>
				</div>
			)}
		</div>
	);
}

// Contents leads to the page's parts below the first screen.
function Contents({ page }: { page: PageReader }) {
	const words = useSiteWords();
	return (
		<nav class="s-contents" aria-labelledby="contents">
			<span id="contents">{words.text("topic.contents")}</span>
			<a href="#basics">{page.text("basis.title")}</a>
			<a href="#idea">{page.text("idea.title")}</a>
			<a href="#solving">{words.text("topic.solving")}</a>
			<a href="#traps">{words.text("topic.traps")}</a>
			<a href="#home">{words.text("topic.home")}</a>
		</nav>
	);
}

// Basis is the ground the topic stands on, card by card: on the island of the
// knights and the liars, its rules. A card shows its drawing on a white box
// where the topic has one, and its mark otherwise.
function Basis({ page, art }: { page: PageReader; art?: readonly Art[] }) {
	return (
		<section id="basics" class="s-wrap s-section">
			<h2 class="s-title">{page.text("basis.title")}</h2>
			<ul class="s-tiles">
				{page.list("basis.cards").map((key, at) => (
					<li key={key} class="s-tile">
						<CardMark page={page} at={key} art={art?.[at]} />
						<h3>{page.text(`${key}.title`)}</h3>
						<p class="s-tile-text">{page.text(`${key}.text`)}</p>
					</li>
				))}
			</ul>
		</section>
	);
}

// CardMark is what a card of the ground shows over its title: its drawing on
// a white box, or its mark.
function CardMark({
	page,
	at,
	art,
}: {
	page: PageReader;
	at: string;
	art: Art | undefined;
}) {
	if (art === undefined) {
		return (
			<span class="s-mark" aria-hidden="true">
				{page.plain(`${at}.mark`)}
			</span>
		);
	}
	return (
		<div class="s-tile-sketch s-topic-art" aria-hidden="true" dir="ltr">
			{art({ page, at: `${at}.art` })}
		</div>
	);
}

// Idea is the topic's key idea: a table, and a note on why it matters. A
// narrow screen scrolls the table sideways inside its frame, which the
// keyboard can reach for that. Where the topic draws a column, a cell of it
// shows its drawing and keeps its words for a screen reader, and the key of
// the drawings, where the topic draws one, stands over the table. The column
// of the answers is in bold: the one the topic names, or else the last.
function Idea({
	page,
	art,
	answers,
	legend,
}: {
	page: PageReader;
	art?: TopicArt["idea"];
	answers?: number;
	legend?: Art;
}) {
	const head = page.list("idea.table.head");
	return (
		<section id="idea" class="s-wrap s-section">
			<div class="s-intro">
				<h2 id="idea-title">{page.text("idea.title")}</h2>
				<p class="s-intro-line">{page.text("idea.lead")}</p>
			</div>
			{legend !== undefined && (
				<p class="s-legend s-idea-legend s-topic-art" aria-hidden="true" dir="ltr">
					{legend({ page, at: "idea.art" })}
				</p>
			)}
			<TableFrame labelledBy="idea-title">
				<table class="s-table s-idea-table" data-answers={answers}>
					<thead>
						<tr>
							{head.map((key) => (
								<th key={key} scope="col">
									{page.text(key)}
								</th>
							))}
						</tr>
					</thead>
					<tbody>
						{page.list("idea.table.rows").map((row, line) => (
							<tr key={row}>
								{cellsOf(page, row, head.length).map((key, at) => {
									const drawn = art?.find((one) => one.column === at);
									const cell = (
										<Cell
											page={page}
											at={key}
											art={drawn?.rows[line]}
											words={drawn?.words ?? "hidden"}
										/>
									);
									return at === 0 ? (
										<th key={key} scope="row">
											{cell}
										</th>
									) : (
										<td key={key}>{cell}</td>
									);
								})}
							</tr>
						))}
					</tbody>
				</table>
			</TableFrame>
			<Note page={page} at="idea.note" />
		</section>
	);
}

// Cell is what a cell of the key idea's table shows: its words, or its
// drawing, the words of the table's drawings under idea.art, with the cell's
// own words under it, beside it or for a screen reader alone, as the topic
// draws the column.
function Cell({
	page,
	at,
	art,
	words,
}: {
	page: PageReader;
	at: string;
	art: Art | undefined;
	words: CellWords;
}) {
	if (art === undefined) {
		return <>{page.text(at)}</>;
	}
	return (
		<span class="s-cell-drawn" data-words={words}>
			<span class="s-cell-sketch s-topic-art" aria-hidden="true" dir="ltr">
				{art({ page, at: "idea.art" })}
			</span>
			<span class={words === "hidden" ? "s-hidden" : "s-cell-words"}>
				{page.text(at)}
			</span>
		</span>
	);
}

// cellsOf are the cells of a row of the table, which has as many as its head:
// a row with more or fewer would leave a column with no heading or a cell with
// none.
function cellsOf(page: PageReader, row: string, columns: number): string[] {
	const cells = page.list(row);
	if (cells.length !== columns) {
		throw new Error(
			`${row} has ${cells.length} cells, and the table's head ${columns}`,
		);
	}
	return cells;
}

// Note is a note on what precedes it: a title, and a line or two, beside its
// picture where it has one.
function Note({
	page,
	at,
	picture,
}: {
	page: PageReader;
	at: string;
	picture?: ComponentChildren;
}) {
	return (
		<aside class="s-note">
			<div class="s-note-words">
				<p class="s-note-title">{page.text(`${at}.title`)}</p>
				<p class="s-note-text">{page.text(`${at}.text`)}</p>
			</div>
			{picture !== undefined && <StepFrame>{picture}</StepFrame>}
		</aside>
	);
}

// Solving is how such tasks are solved: the steps of the move, then examples
// worked through to their answers, from the simplest to an olympiad's. Each
// example's level and drawing are data, the level saying its grades, and the
// words give as many examples as the data does.
function Solving({
	page,
	topic,
	examples,
	art,
}: {
	page: PageReader;
	topic: Topic;
	examples: readonly TopicExample[];
	art: TopicArt | undefined;
}) {
	const words = useSiteWords();
	const numbers = new Intl.NumberFormat(page.locale);
	const worked = page.list("examples").length;
	if (worked !== examples.length) {
		throw new Error(
			`the words work through ${worked} examples, and site/data.json gives ${examples.length}`,
		);
	}
	return (
		<section id="solving" class="s-wrap s-section">
			<div class="s-intro">
				<h2>{page.text("method.title")}</h2>
				<p class="s-intro-line">{page.text("method.lead")}</p>
			</div>
			<ol class="s-moves">
				{page.list("method.steps").map((key, at) => (
					<li key={key}>
						<span class="s-move">
							<span class="s-number" aria-hidden="true">
								{numbers.format(at + 1)}
							</span>
							<span>{page.text(key)}</span>
						</span>
					</li>
				))}
			</ol>
			{art?.legend !== undefined && (
				<p class="s-legend s-topic-art" aria-hidden="true" dir="ltr">
					{art.legend({ page, at: "method.legend" })}
				</p>
			)}
			{examples.map((example, at) => {
				const key = `examples.${at + 1}`;
				return (
					<WorkedExample
						key={key}
						page={page}
						at={key}
						label={words.text("topic.example", {
							number: at + 1,
							grades: gradesText(words, example.grades),
						})}
						answer={example.answer}
						drawing={example.drawing}
						art={art?.examples?.[at]}
						where={`example ${at + 1} of ${topic.id}`}
					/>
				);
			})}
		</section>
	);
}

// WorkedExample is one example: its number and grades, its title, its task
// with its drawing under it where it has one, the steps of its solution,
// each beside its picture where the topic draws them and beside the
// example's one drawing otherwise, a note where it has one, and its answer
// behind its badge. A picture that cannot be drawn stops the build under the
// name where gives the example.
function WorkedExample({
	page,
	at,
	label,
	answer,
	drawing,
	art,
	where,
}: {
	page: PageReader;
	at: string;
	label: string;
	answer: string;
	drawing: Drawing | undefined;
	art: ExampleArt | undefined;
	where: string;
}) {
	const words = useSiteWords();
	const own = `${at}.art`;
	return (
		<article class="s-example">
			<div class="s-example-head">
				<p class="s-example-label">{label}</p>
				<h3>{page.text(`${at}.title`)}</h3>
			</div>
			<div class="s-example-task">
				<p class="s-example-question">{page.text(`${at}.question`)}</p>
				{art?.task !== undefined && (
					<div class="s-task-sketch s-topic-art" aria-hidden="true" dir="ltr">
						{art.task({ page, at: own })}
					</div>
				)}
			</div>
			{art?.steps !== undefined ? (
				<WorkedSteps
					page={page}
					at={`${at}.steps`}
					pictureOf={(step) => art.steps?.[step]?.({ page, at: own })}
				/>
			) : (
				<Solution
					page={page}
					at={at}
					drawing={
						drawing !== undefined && (
							<TopicDrawing
								drawing={drawing}
								page={page}
								at={`${at}.drawing`}
								where={where}
							/>
						)
					}
				/>
			)}
			{page.has(`${at}.note`) && (
				<Note
					page={page}
					at={`${at}.note`}
					picture={art?.note?.({ page, at: own })}
				/>
			)}
			{art?.steps !== undefined && (
				<Answer
					label={words.text("topic.answer")}
					badge={badgeOf(page, at, answer)}
				>
					{page.text(`${at}.answer`)}
				</Answer>
			)}
		</article>
	);
}

// badgeOf is the badge of an example's answer: the one its words give, or
// the answer its solver proves where that is a number or a time, which reads
// the same in every language; an answer in words with no badge in the words
// has none.
function badgeOf(
	page: PageReader,
	at: string,
	answer: string,
): string | undefined {
	if (page.has(`${at}.badge`)) {
		return page.plain(`${at}.badge`);
	}
	return /^\d+(:\d\d)?$/.test(answer) ? answer : undefined;
}

// Traps are the traps the topic's page explains, each under the name the card
// gives it: how it shows in the topic, and what a parent can say.
function Traps({
	page,
	traps,
	card,
	art,
}: {
	page: PageReader;
	traps: readonly string[];
	card: Words<Key>;
	art?: Readonly<Record<string, Art>>;
}) {
	const words = useSiteWords();
	return (
		<section id="traps" class="s-wrap s-section">
			<div class="s-intro">
				<h2>{words.text("topic.traps")}</h2>
				<p class="s-intro-line">{words.text("topic.traps_lead")}</p>
			</div>
			<ul class="s-tiles">
				{traps.map((id) => (
					<li key={id} class="s-trap">
						<span class="s-trap-badge">{words.text("topic.trap")}</span>
						<h3>{trapName(card, id)}</h3>
						<p class="s-tile-text">{page.text(`traps.${id}.shows`)}</p>
						{art?.[id] !== undefined && (
							<div class="s-trap-sketch s-topic-art" aria-hidden="true" dir="ltr">
								{art[id]({ page, at: `traps.${id}.art` })}
							</div>
						)}
						<div class="s-say">
							<p class="s-note-title">{words.text("topic.say")}</p>
							<p class="s-tile-text">{page.text(`traps.${id}.say`)}</p>
						</div>
					</li>
				))}
			</ul>
		</section>
	);
}

// Home is how to help at home, tip by tip.
function Home({ page }: { page: PageReader }) {
	const words = useSiteWords();
	const numbers = new Intl.NumberFormat(page.locale, {
		minimumIntegerDigits: 2,
	});
	return (
		<section id="home" class="s-wrap s-section">
			<h2 class="s-title">{words.text("topic.home")}</h2>
			<ol class="s-tiles">
				{page.list("home").map((key, at) => (
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

// Related are the topics of the catalog's links under heading: those this one
// builds on, to learn first, or those it opens. Each leads to its page when the
// page is published, and says it is coming until then. With no such topic
// there is nothing to say.
function Related({
	heading,
	ids,
	data,
	card,
}: {
	heading: SiteKey;
	ids: readonly string[];
	data: SiteData;
	card: Words<Key>;
}) {
	const words = useSiteWords();
	const related = ids.flatMap((id) => {
		const topic = data.topics.byId.get(id);
		return topic === undefined ? [] : [topic];
	});
	if (related.length === 0) {
		return null;
	}
	const numbers = new Intl.NumberFormat(words.locale);
	return (
		<section class="s-wrap s-section">
			<h2 class="s-title">{words.text(heading)}</h2>
			<ul class="s-related">
				{related.map((topic) => (
					<li key={topic.id}>
						<RelatedCard
							topic={topic}
							name={topicName(card, topic.id)}
							thumb={
								<TopicThumb
									drawing={data.drawings.get(topic.id)?.card}
									numbers={numbers}
								/>
							}
						/>
					</li>
				))}
			</ul>
		</section>
	);
}

// RelatedCard is a related topic: its small drawing, its name and its
// grades, the whole card a link to its page once the page is published, and
// word that it is coming until then.
function RelatedCard({
	topic,
	name,
	thumb,
}: {
	topic: Topic;
	name: string;
	thumb: ComponentChildren;
}) {
	const words = useSiteWords();
	const body = (
		<>
			{thumb}
			<span class="s-related-foot">
				<span class="s-related-words">
					<span class="s-related-name">{name}</span>
					<span class="s-related-grades">
						{gradesText(words, topic.grades)}
					</span>
				</span>
				{topic.sitePage ? (
					<span class="s-related-arrow" aria-hidden="true">
						→
					</span>
				) : (
					<span class="s-soon">{words.text("topics.soon")}</span>
				)}
			</span>
		</>
	);
	return topic.sitePage ? (
		<a
			class="s-related-card"
			href={address(words.locale, `topics/${topic.slug}`)}
		>
			{body}
		</a>
	) : (
		<span class="s-related-card">{body}</span>
	);
}

// Ask is how to ask for a task on the topic: the words to write in the chat,
// the way to add MathTrail to a chat first, and the way back to every topic,
// in one low row.
function Ask({ page }: { page: PageReader }) {
	const words = useSiteWords();
	return (
		<section class="s-section s-ask-wrap s-topic-ask">
			<div class="s-ask s-ask-row">
				<div class="s-ask-copy">
					<h2 class="s-ask-title">{words.text("topic.ask")}</h2>
					<p class="s-ask-phrase">{page.text("ask")}</p>
				</div>
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
