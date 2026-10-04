import catalog from "../../../content/catalogs/topics.json";
import { groups } from "../../../site/data.json";

/**
 * TopicGroup is a group of topics as the site's page of topics shows it: its
 * id, which is also the anchor of its part of that page, and its topics in the
 * site's order, each with the first grade it is taught from.
 */
export type TopicGroup = {
	readonly id: string;
	readonly topics: readonly {
		readonly id: string;
		readonly fromGrade: number;
	}[];
};

// firstGrade is the first grade of the lowest level of levels: 3 for a topic
// taught from 3-4 on.
function firstGrade(levels: readonly string[]): number {
	return Math.min(...levels.map((level) => Number(level.split("-")[0])));
}

/**
 * groupsOf are the groups of topics a card offers, made from the site's groups
 * of its page of topics and the catalog's topics: each group with its topics
 * in the site's order, and the first grade each is taught from, by the levels
 * the catalog writes — 1-2, 3-4, 5-6. A topic the catalog has no level for is
 * left out: the card could not say where it begins.
 */
export function groupsOf(
	siteGroups: readonly { id: string; topics: readonly string[] }[],
	catalogTopics: readonly { id: string; grade_levels: readonly string[] }[],
): TopicGroup[] {
	const levelsOf = new Map(
		catalogTopics.map((topic) => [topic.id, topic.grade_levels]),
	);
	return siteGroups.map((group) => ({
		id: group.id,
		topics: group.topics.flatMap((id) => {
			const levels = levelsOf.get(id) ?? [];
			return levels.length === 0 ? [] : [{ id, fromGrade: firstGrade(levels) }];
		}),
	}));
}

/**
 * topicGroups are the topics a card offers to keep the lessons to, grouped and
 * ordered as the site's page of topics shows them, from the two files the site
 * and the service are built from: the site's groups and the catalog's levels.
 */
export const topicGroups: readonly TopicGroup[] = groupsOf(groups, catalog);
