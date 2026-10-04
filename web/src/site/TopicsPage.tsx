import type { Section } from "../widget/folds";
import { listed, topicName } from "../widget/names";
import { cardWords } from "../widget/words";
import { address } from "./addresses";
import { progressOf, type SiteData } from "./data";
import type { PageProps } from "./pages";
import type { PageReader } from "./reader";
import { StaticProgress } from "./StaticCard";
import { TopicMap } from "./TopicMap";
import { lightRules, mapOf } from "./topicmap";
import type { Group, Topic } from "./topics";
import { type SiteKey, useSiteWords } from "./words";

// topicsOpen is the section of the progress the page shows open: the topics,
// each with its rank, which is what the page is about.
const topicsOpen: ReadonlySet<Section> = new Set(["topics"]);

/**
 * TopicsPage is the page of the topics. It opens with what a child's tasks
 * are about, beside the progress a chat shows; then the map of how the topics
 * build on one another; then every topic of the catalog by its group, each
 * with a drawing, a line for the parent, the topics it stands on and opens,
 * its grades, and a link to its own page once the page is published. A
 * topic's name is the card's own, so that the page and the card agree.
 */
export function TopicsPage({ page, data }: PageProps) {
	const words = useSiteWords();
	const card = cardWords(page.locale, undefined);
	const name = (topic: Topic) => topicName(card, topic.id);
	const names = (ids: readonly string[]) =>
		listed(
			card,
			ids.map((id) => topicName(card, id)),
		);
	return (
		<>
			<section class="s-wrap s-hero">
				<div class="s-hero-copy">
					<h1>{page.text("hero.title")}</h1>
					<p class="s-lead">{page.text("hero.lead")}</p>
				</div>
				<figure class="s-panel s-hero-card">
					<StaticProgress
						report={progressOf(data, page.plain("hero.child"))}
						locale={page.locale}
						open={topicsOpen}
					/>
					<figcaption>{page.text("hero.caption")}</figcaption>
				</figure>
			</section>
			<section class="s-wrap s-section">
				<div class="s-intro">
					<h2>{page.text("map.title")}</h2>
					<p class="s-intro-line">{page.text("map.lead")}</p>
				</div>
				<TopicMap page={page} topics={data.topics} name={name} names={names} />
			</section>
			<section class="s-wrap s-section">
				<nav class="s-pills" aria-label={page.plain("groups.label")}>
					{data.topics.groups.map((group) => (
						<a key={group.id} href={`#${group.id}`}>
							{words.text(groupKey(group))}
						</a>
					))}
				</nav>
				{data.topics.groups.map((group) => (
					<section key={group.id} id={group.id} class="s-group">
						<div class="s-intro">
							<h2>{words.text(groupKey(group))}</h2>
							<p class="s-intro-line">{page.text(`groups.${group.id}`)}</p>
						</div>
						<div class="s-topics">
							{group.topics.flatMap((id) => {
								const topic = data.topics.byId.get(id);
								return topic === undefined
									? []
									: [
											<TopicCard
												key={id}
												page={page}
												topic={topic}
												name={name(topic)}
												names={names}
											/>,
										];
							})}
						</div>
					</section>
				))}
			</section>
		</>
	);
}

/**
 * topicsStyle are the rules the page's head carries: those that light the map
 * of the topics while a topic is pointed at, written from the catalog's links.
 */
export function topicsStyle(data: SiteData): string {
	return lightRules(data.topics, mapOf(data.topics));
}

// groupKey is the site's word for a group's name, which a topic's own page
// says too.
function groupKey(group: Group): SiteKey {
	return `group.${group.id}` as SiteKey;
}

// TopicCard is a topic as the page of the topics shows it, under the anchor of
// its slug, which the map leads to.
function TopicCard({
	page,
	topic,
	name,
	names,
}: {
	page: PageReader;
	topic: Topic;
	name: string;
	names: (ids: readonly string[]) => string;
}) {
	const words = useSiteWords();
	const key = `topics.${topic.slug}`;
	const grades = new Intl.NumberFormat(page.locale);
	return (
		<article id={topic.slug} class="s-topic">
			<pre class="s-topic-drawing" dir="ltr">
				{page.plain(`${key}.drawing`)}
			</pre>
			<h3>{name}</h3>
			<p class="s-topic-phrase">{page.text(`${key}.phrase`)}</p>
			{topic.bases.length > 0 && (
				<p class="s-topic-links">
					{words.text("topics.builds_on", { topics: names(topic.bases) })}
				</p>
			)}
			{topic.opens.length > 0 && (
				<p class="s-topic-links">
					{words.text("topics.opens", { topics: names(topic.opens) })}
				</p>
			)}
			<p class="s-topic-foot">
				<span>
					{words.text("topics.grades", {
						range: `${grades.format(topic.grades[0])}–${grades.format(topic.grades[1])}`,
					})}
				</span>
				{topic.sitePage ? (
					<a href={address(page.locale, `topics/${topic.slug}`)}>
						{words.text("topics.more")}
					</a>
				) : (
					<span class="s-soon">{words.text("topics.soon")}</span>
				)}
			</p>
		</article>
	);
}
