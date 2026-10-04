import { describe, expect, test } from "vitest";
import { type CatalogTopic, readTopics } from "./topics";

// topic is a topic of a catalog, its page not published unless said.
function topic(
	id: string,
	builds_on: string[] = [],
	grade_levels = ["1-2", "3-4"],
): CatalogTopic {
	return {
		id,
		slug: id.slice(id.indexOf(".") + 1),
		grade_levels,
		builds_on,
		site_page: false,
	};
}

// A catalog of three layers: a and b are foundations, c builds on a, d on c
// and b.
const catalog = [
	topic("x.a"),
	topic("x.b", [], ["5-6"]),
	topic("x.c", ["x.a"], ["3-4", "5-6"]),
	topic("x.d", ["x.b", "x.c"]),
];
const groups = [
	{ id: "first", topics: ["x.a", "x.b"] },
	{ id: "second", topics: ["x.c", "x.d"] },
];

describe("the topics as the site shows them", () => {
	const topics = readTopics(catalog, groups);
	const of = (id: string) => topics.byId.get(id);

	test("stand in the layer of their longest chain of bases", () => {
		expect(topics.all.map((t) => [t.id, t.layer])).toEqual([
			["x.a", 0],
			["x.b", 0],
			["x.c", 1],
			["x.d", 2],
		]);
	});

	test("name the topics they build on and that build on them, directly", () => {
		expect(of("x.c")?.bases).toEqual(["x.a"]);
		expect(of("x.c")?.opens).toEqual(["x.d"]);
		expect(of("x.a")?.opens).toEqual(["x.c"]);
		expect(of("x.d")?.opens).toEqual([]);
	});

	test("name every topic below and above them, in the catalog's order", () => {
		expect(of("x.d")?.before).toEqual(["x.a", "x.b", "x.c"]);
		expect(of("x.a")?.after).toEqual(["x.c", "x.d"]);
		expect(of("x.b")?.before).toEqual([]);
	});

	test("are taught from the first grade of their levels to the last", () => {
		expect(of("x.a")?.grades).toEqual([1, 4]);
		expect(of("x.b")?.grades).toEqual([5, 6]);
		expect(of("x.c")?.grades).toEqual([3, 6]);
	});

	test("are shown in the groups as given", () => {
		expect(topics.groups).toEqual(groups);
	});

	test.each([
		[
			"a topic left out of every group",
			[{ id: "first", topics: ["x.a", "x.b", "x.c"] }],
			"the topic x.d is in no group",
		],
		[
			"a topic in two groups",
			[
				{ id: "first", topics: ["x.a", "x.b", "x.c"] },
				{ id: "second", topics: ["x.c", "x.d"] },
			],
			"the topic x.c is in the group first and in the group second",
		],
		[
			"a topic the catalog does not have",
			[...groups, { id: "third", topics: ["x.z"] }],
			"the group third holds x.z, which the catalog does not have",
		],
		[
			"a group with no topic",
			[...groups, { id: "third", topics: [] }],
			"the group third holds no topic",
		],
		[
			"a group named as another",
			[
				{ id: "first", topics: ["x.a", "x.b"] },
				{ id: "first", topics: ["x.c", "x.d"] },
			],
			"the group first shares its name with another anchor",
		],
		[
			"a group named as a topic's card",
			[
				{ id: "a", topics: ["x.a", "x.b"] },
				{ id: "second", topics: ["x.c", "x.d"] },
			],
			"the group a shares its name with another anchor",
		],
		[
			"a group no anchor can name",
			[
				{ id: "First group", topics: ["x.a", "x.b"] },
				{ id: "second", topics: ["x.c", "x.d"] },
			],
			'the group "First group" is no anchor',
		],
	])("are refused for %s", (_, broken, want) => {
		expect(() => readTopics(catalog, broken)).toThrow(want);
	});
});
