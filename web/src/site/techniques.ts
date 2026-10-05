import type { Example, TechniquesFile } from "./data";
import { anchorName, type CatalogTopic, gradesOf } from "./topics";

/**
 * Technique is a technique of problem solving as the page of the techniques
 * shows it, as far as no language changes it: its name, which is its anchor
 * on the page, the topics of the catalog it leads to, and the example it is
 * worked through on.
 */
export type Technique = {
	readonly id: string;
	readonly topics: readonly string[];
	readonly example: Example;
};

/** TechniqueGroup is a group of techniques, in the order the page shows them. */
export type TechniqueGroup = {
	readonly id: string;
	readonly techniques: readonly Technique[];
};

/**
 * Techniques is what the page of the techniques draws from the site's data:
 * its groups, each with its techniques in order, and the rows of its hint on
 * which technique to try, each naming the techniques it suggests.
 */
export type Techniques = {
	readonly groups: readonly TechniqueGroup[];
	readonly cues: readonly (readonly string[])[];
};

/**
 * allTechniques are every technique of the page, group by group, in the order
 * the page shows them: the order the page numbers them in, and how many there
 * are wherever the site counts them.
 */
export function allTechniques(techniques: Techniques): readonly Technique[] {
	return techniques.groups.flatMap((group) => group.techniques);
}

/**
 * readTechniques reads the techniques of the site's data beside the catalog's
 * topics. A technique is refused whose name is no anchor or another's, that
 * leads to a topic the catalog lacks, or whose example is set at a level none
 * of its topics is taught at — or, for a technique of no topic, no topic at
 * all; and so is a row of the hint that names a technique the page lacks.
 * Nothing may be named twice where the page would draw it twice: a group, a
 * topic of one technique, a technique in one row of the hint.
 */
export function readTechniques(
	catalog: readonly CatalogTopic[],
	file: TechniquesFile,
): Techniques {
	const byId = new Map(catalog.map((topic) => [topic.id, topic]));
	const named = new Set<string>();
	checkOnce(
		"the page's groups",
		file.groups.map((group) => group.id),
	);
	const groups = file.groups.map((group) => ({
		id: group.id,
		techniques: group.techniques.map((technique) => {
			checkName(technique.id, named);
			named.add(technique.id);
			checkOnce(`the technique ${technique.id}`, technique.topics);
			const topics = technique.topics.map((id) => {
				const topic = byId.get(id);
				if (topic === undefined) {
					throw new Error(
						`the technique ${technique.id} leads to ${id}, a topic the catalog does not have`,
					);
				}
				return topic;
			});
			return {
				id: technique.id,
				topics: technique.topics,
				example: exampleOf(
					technique.id,
					topics.length === 0 ? catalog : topics,
					technique.example.level,
				),
			};
		}),
	}));
	for (const row of file.cues) {
		checkOnce("a row of the hint", row);
		for (const id of row) {
			if (!named.has(id)) {
				throw new Error(
					`the hint suggests ${id}, a technique the page does not have`,
				);
			}
		}
	}
	return { groups, cues: file.cues };
}

// checkOnce refuses a list that names one thing twice, which the page would
// draw twice.
function checkOnce(where: string, ids: readonly string[]): void {
	const seen = new Set<string>();
	for (const id of ids) {
		if (seen.has(id)) {
			throw new Error(`${where} names ${id} twice`);
		}
		seen.add(id);
	}
}

// checkName refuses a technique's name that is no anchor, and one another
// technique of the page already has.
function checkName(id: string, named: ReadonlySet<string>): void {
	if (!anchorName.test(id)) {
		throw new Error(
			`the technique ${JSON.stringify(id)} is no anchor: a technique is called by lowercase words joined by dashes`,
		);
	}
	if (named.has(id)) {
		throw new Error(`two techniques are called ${id}`);
	}
}

// exampleOf is the example of the technique called id set at level, which
// has to be a level one of topics is taught at: the technique's own, or the
// whole catalog's for a technique that leads to no topic.
function exampleOf(
	id: string,
	topics: readonly CatalogTopic[],
	level: string,
): Example {
	if (!topics.some((topic) => topic.grade_levels.includes(level))) {
		throw new Error(
			`the example of ${id} is set at ${level}, a level none of its topics is taught at`,
		);
	}
	return { grades: gradesOf([level]) };
}
