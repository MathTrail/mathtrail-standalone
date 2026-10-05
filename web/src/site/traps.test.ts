import { describe, expect, test } from "vitest";
import { type ReferenceTask, rankTraps } from "./traps";

// task is a reference task of topic whose wrong options name traps, one each.
function task(topic: string, ...named: string[]): ReferenceTask {
	return {
		id: `${topic}-${named.join("-")}`,
		topic,
		grade_level: "3-4",
		distractors: Object.fromEntries(
			named.map((trap, at) => ["BCDE"[at] ?? "", { trap }]),
		),
	};
}

// A catalog of four traps, in this order.
const catalog = [
	{ id: "first" },
	{ id: "second" },
	{ id: "third" },
	{ id: "fourth" },
];

describe("the traps of a topic", () => {
	test("come the most frequent first, counted over every wrong option of its reference tasks", () => {
		const ranked = rankTraps(catalog, [
			task("x", "fourth", "second", "fourth"),
			task("x", "fourth", "third"),
			task("y", "first"),
		]);

		expect(ranked.get("x")).toEqual(["fourth", "second", "third"]);
		expect(ranked.get("y")).toEqual(["first"]);
	});

	test("named as often as each other stand in the catalog's order", () => {
		const ranked = rankTraps(catalog, [
			task("x", "fourth", "second"),
			task("x", "third", "first"),
		]);

		expect(ranked.get("x")).toEqual(["first", "second", "third", "fourth"]);
	});

	test("are refused when a reference task names a trap the catalog does not have", () => {
		expect(() => rankTraps(catalog, [task("x", "first", "fifth")])).toThrow(
			"the reference task x-first-fifth names the trap fifth, which the catalog does not have",
		);
	});
});
