import { address } from "./addresses";
import type { Topic } from "./topics";
import { gradesText, useSiteWords } from "./words";

/**
 * TopicFoot is the foot of a topic's card wherever the site shows one: the
 * topic's grades, and a link to its page once the page is published, or word
 * that the page is coming.
 */
export function TopicFoot({ topic }: { topic: Topic }) {
	const words = useSiteWords();
	return (
		<p class="s-topic-foot">
			<span>{gradesText(words, topic.grades)}</span>
			{topic.sitePage ? (
				<a href={address(words.locale, `topics/${topic.slug}`)}>
					{words.text("topics.more")}
				</a>
			) : (
				<span class="s-soon">{words.text("topics.soon")}</span>
			)}
		</p>
	);
}
