import catalog from "../../../content/catalogs/topics.json";
import { groups } from "../../../site/data.json";

/**
 * TopicGroup is a group of topics as the site's page of topics shows it: its
 * id, which is also the anchor of its part of that page, and its topics in the
 * site's order, each with the first grade it is taught from.
 */
export type TopicGroup = {
	readonly id: string;
	readonly topics: readonly { readonly id: string; readonly fromGrade: number }[];
};

// levelsOf are the levels each topic of the catalog is taught at, as the
// catalog writes them: 1-2, 3-4, 5-6.
const levelsOf: ReadonlyMap<string, readonly string[]> = new Map(
	catalog.map((topic) => [topic.id, topic.grade_levels]),
);

// firstGrade is the first grade of the lowest level of levels: 3 for a topic
// taught from 3-4 on.
function firstGrade(levels: readonly string[]): number {
	return Math.min(...levels.map((level) => Number(level.split("-")[0])));
}

/**
 * topicGroups are the topics a card offers to keep the lessons to, grouped and
 * ordered as the site's page of topics shows them, from the two files the site
 * and the service are built from: the site's groups and the catalog's levels.
 * A topic the catalog does not have is left out.
 */
export const topicGroups: readonly TopicGroup[] = groups.map((group) => ({
	id: group.id,
	topics: group.topics.flatMap((id) => {
		const levels = levelsOf.get(id);
		return levels === undefined || levels.length === 0
			? []
			: [{ id, fromGrade: firstGrade(levels) }];
	}),
}));
