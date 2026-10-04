import { describe, expect, test } from "vitest";
import catalog from "../../../content/catalogs/topics.json";
import { groups } from "../../../site/data.json";
import { groupsOf, topicGroups } from "./topicGroups";

describe("the groups of topics a card offers", () => {
	test("are the site's groups, in the order of its page of topics", () => {
		expect(topicGroups.map((group) => group.id)).toEqual(
			groups.map((group) => group.id),
		);
	});

	test("hold every topic of the catalog once", () => {
		const offered = topicGroups.flatMap((group) =>
			group.topics.map((topic) => topic.id),
		);
		expect([...offered].sort()).toEqual(
			catalog.map((topic) => topic.id).sort(),
		);
		expect(new Set(offered).size).toBe(offered.length);
	});

	test("name the first grade each topic is taught from", () => {
		const from = new Map(
			topicGroups.flatMap((group) =>
				group.topics.map((topic) => [topic.id, topic.fromGrade] as const),
			),
		);
		expect(from.get("time.clocks")).toBe(1);
		expect(from.get("logic.knights_liars")).toBe(3);
		expect(from.get("percent.basic")).toBe(5);
	});

	test("leave out a topic the catalog has no level for", () => {
		expect(
			groupsOf(
				[
					{
						id: "logic",
						topics: ["logic.knights_liars", "logic.unknown", "logic.unset"],
					},
				],
				[
					{ id: "logic.knights_liars", grade_levels: ["5-6", "3-4"] },
					{ id: "logic.unset", grade_levels: [] },
				],
			),
		).toEqual([
			{ id: "logic", topics: [{ id: "logic.knights_liars", fromGrade: 3 }] },
		]);
	});
});
