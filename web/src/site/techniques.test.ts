import { describe, expect, test } from "vitest";
import type { TechniquesFile } from "./data";
import { readTechniques } from "./techniques";
import type { CatalogTopic } from "./topics";

// topic is a topic of a catalog, taught at levels.
function topic(id: string, grade_levels: string[]): CatalogTopic {
	return {
		id,
		slug: id.slice(id.indexOf(".") + 1),
		grade_levels,
		builds_on: [],
		site_page: false,
	};
}

// A catalog of two topics: a taught from the first grade, b only in the last
// two.
const catalog = [topic("x.a", ["1-2", "3-4"]), topic("x.b", ["5-6"])];

// example is an example set at level.
function example(level: string, solver = "a-solver") {
	return { level, solver, answer: "1" };
}

// file is a page of three techniques in two groups: one leading to a, one to
// b, and one to no topic.
const file: TechniquesFile = {
	groups: [
		{
			id: "see",
			techniques: [
				{ id: "draw", topics: ["x.a"], example: example("1-2") },
				{ id: "mirror", topics: ["x.b"], example: example("5-6") },
			],
		},
		{
			id: "find",
			techniques: [{ id: "all-the-same", topics: [], example: example("3-4") }],
		},
	],
	cues: [["draw"], ["mirror", "all-the-same"]],
};

// changed is file with its one technique called id replaced by technique.
function changed(
	id: string,
	technique: TechniquesFile["groups"][number]["techniques"][number],
): TechniquesFile {
	return {
		...file,
		groups: file.groups.map((group) => ({
			...group,
			techniques: group.techniques.map((each) =>
				each.id === id ? technique : each,
			),
		})),
	};
}

describe("the techniques as the page shows them", () => {
	const techniques = readTechniques(catalog, file);

	test("keep their groups and their order", () => {
		expect(
			techniques.groups.map((group) => [
				group.id,
				group.techniques.map((technique) => technique.id),
			]),
		).toEqual([
			["see", ["draw", "mirror"]],
			["find", ["all-the-same"]],
		]);
	});

	test("say the grades of the level each example is set at", () => {
		expect(
			techniques.groups.flatMap((group) =>
				group.techniques.map((technique) => technique.example.grades),
			),
		).toEqual([
			[1, 2],
			[5, 6],
			[3, 4],
		]);
	});

	test("lead to their topics, and keep the hint's rows", () => {
		expect(techniques.groups[0]?.techniques[0]?.topics).toEqual(["x.a"]);
		expect(techniques.cues).toEqual(file.cues);
	});

	test.each([
		[
			"a technique no anchor can name",
			changed("draw", {
				id: "Draw it",
				topics: ["x.a"],
				example: example("1-2"),
			}),
			'the technique "Draw it" is no anchor',
		],
		[
			"two techniques of one name",
			changed("mirror", {
				id: "draw",
				topics: ["x.b"],
				example: example("5-6"),
			}),
			"two techniques are called draw",
		],
		[
			"a topic the catalog does not have",
			changed("draw", { id: "draw", topics: ["x.z"], example: example("1-2") }),
			"the technique draw leads to x.z, a topic the catalog does not have",
		],
		[
			"an example at a level its topics are not taught at",
			changed("mirror", {
				id: "mirror",
				topics: ["x.b"],
				example: example("3-4"),
			}),
			"the example of mirror is set at 3-4, a level none of its topics is taught at",
		],
		[
			"an example of no topic at a level the catalog does not teach",
			changed("all-the-same", {
				id: "all-the-same",
				topics: [],
				example: example("7-8"),
			}),
			"the example of all-the-same is set at 7-8, a level none of its topics is taught at",
		],
		[
			"a hint suggesting a technique the page does not have",
			{ ...file, cues: [["draw"], ["guess"]] },
			"the hint suggests guess, a technique the page does not have",
		],
		[
			"two groups of one name",
			{
				...file,
				groups: file.groups.map((group) => ({ ...group, id: "see" })),
			},
			"the page's groups names see twice",
		],
		[
			"a technique that leads to one topic twice",
			changed("draw", {
				id: "draw",
				topics: ["x.a", "x.a"],
				example: example("1-2"),
			}),
			"the technique draw names x.a twice",
		],
		[
			"a row of the hint that suggests one technique twice",
			{ ...file, cues: [["draw", "draw"]] },
			"a row of the hint names draw twice",
		],
	])("are refused for %s", (_, broken, want) => {
		expect(() => readTechniques(catalog, broken)).toThrow(want);
	});
});
