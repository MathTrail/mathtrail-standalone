import type { Topic } from "./topics";
import { gradesText, useSiteWords } from "./words";

/**
 * TopicFoot is the foot of a topic's card on the page of the topics: the
 * topic's grades, and the sign that the card leads to the topic's page once
 * the page is published, or word that the page is coming. The sign is no link
 * of its own, and a reader of links skips it: the card's link is the topic's
 * name, whose box covers the card.
 */
export function TopicFoot({ topic }: { topic: Topic }) {
	const words = useSiteWords();
	return (
		<p class="s-topic-foot">
			<span>{gradesText(words, topic.grades)}</span>
			{topic.sitePage ? (
				<span class="s-topic-more" aria-hidden="true">
					{words.text("topics.more")}
				</span>
			) : (
				<span class="s-soon">{words.text("topics.soon")}</span>
			)}
		</p>
	);
}
