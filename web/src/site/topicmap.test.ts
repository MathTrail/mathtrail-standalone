import { describe, expect, test } from "vitest";
import { board, lightRules, lineId, litBy, mapOf, nodeId } from "./topicmap";
import { type CatalogTopic, readTopics } from "./topics";

// topic is a topic of a catalog by its id's last word.
function topic(name: string, ...bases: string[]): CatalogTopic {
	return {
		id: `x.${name}`,
		slug: name,
		grade_levels: ["1-2"],
		builds_on: bases.map((base) => `x.${base}`),
		site_page: false,
	};
}

// Three layers: a, b and c are foundations; d builds on a, e and h on c; f
// on d, and g on b and e, the link from b skipping a column. h shares a base
// with e and has nothing to do with g.
const topics = readTopics(
	[
		topic("a"),
		topic("b"),
		topic("c"),
		topic("d", "a"),
		topic("e", "c"),
		topic("f", "d"),
		topic("g", "b", "e"),
		topic("h", "c"),
	],
	[
		{
			id: "all",
			topics: ["x.a", "x.b", "x.c", "x.d", "x.e", "x.f", "x.g", "x.h"],
		},
	],
);
const map = mapOf(topics);
// segmentsOf are the straight stretches between the points a path passes
// through or bends towards, in order: a curve is taken by its control points,
// which hold it.
function segmentsOf(path: string): [[number, number], [number, number]][] {
	const segments: [[number, number], [number, number]][] = [];
	let at: [number, number] = [0, 0];
	for (const [, command = "", written = ""] of path.matchAll(
		/([MHVQC])([^MHVQC]*)/g,
	)) {
		const numbers = written.trim().split(/\s+/).map(Number);
		const points: [number, number][] = [];
		if (command === "H") {
			points.push([numbers[0] ?? 0, at[1]]);
		} else if (command === "V") {
			points.push([at[0], numbers[0] ?? 0]);
		} else {
			for (let i = 0; i + 1 < numbers.length; i += 2) {
				points.push([numbers[i] ?? 0, numbers[i + 1] ?? 0]);
			}
		}
		for (const point of points) {
			if (command !== "M") {
				segments.push([at, point]);
			}
			at = point;
		}
	}
	return segments;
}

const spot = (name: string) =>
	map.spots.find((at) => at.topic.id === `x.${name}`);

describe("the map of the topics", () => {
	test("puts each topic in the column of its layer, foundations in the catalog's order", () => {
		expect(map.columns.map((column) => column.map((t) => t.slug))).toEqual([
			["a", "b", "c"],
			["d", "e", "h"],
			["f", "g"],
		]);
	});

	test("orders a column by where its topics' bases stand", () => {
		const ordered = mapOf(
			readTopics(
				[topic("a"), topic("b"), topic("c"), topic("z", "c"), topic("y", "a")],
				[{ id: "all", topics: ["x.a", "x.b", "x.c", "x.z", "x.y"] }],
			),
		);

		expect(ordered.columns[1]?.map((t) => t.slug)).toEqual(["y", "z"]);
	});

	test("draws every link of the catalog once, from the base to the topic", () => {
		expect(map.lines.map(lineId)).toEqual([
			"line-a-d",
			"line-c-e",
			"line-d-f",
			"line-b-g",
			"line-e-g",
			"line-c-h",
		]);
	});

	test("starts a line at its base's side and ends it at its topic's", () => {
		for (const line of map.lines) {
			const from = spot(line.from.slug);
			const to = spot(line.to.slug);
			const middle = board.height / 2;

			expect(
				line.path.startsWith(
					`M${(from?.x ?? 0) + board.width} ${(from?.y ?? 0) + middle}`,
				),
			).toBe(true);
			expect(line.end).toEqual([to?.x, (to?.y ?? 0) + middle]);
			expect(
				line.path.endsWith(`${to?.x} ${(to?.y ?? 0) + middle}`) ||
					line.path.endsWith(`H${to?.x}`),
			).toBe(true);
		}
	});

	test("leads a line that skips a column under every box of that column", () => {
		const skipping = map.lines.find((line) => lineId(line) === "line-b-g");
		const middle = map.columns[1] ?? [];
		const left = Math.min(...middle.map((t) => spot(t.slug)?.x ?? 0));
		const bottom = Math.max(...map.spots.map((at) => at.y + board.height));
		// A stretch crosses the column when its ends lie on either side of the
		// column's boxes or within them.
		const across = segmentsOf(skipping?.path ?? "").filter(
			([[x1], [x2]]) =>
				Math.max(x1, x2) >= left && Math.min(x1, x2) <= left + board.width,
		);

		expect(across).not.toEqual([]);
		expect(across.every(([[, y1], [, y2]]) => y1 > bottom && y2 > bottom)).toBe(
			true,
		);
		expect(map.height).toBeGreaterThan(Math.max(...across.map(([[, y]]) => y)));
	});
});

describe("the light of a topic pointed at", () => {
	const g = topics.byId.get("x.g");
	if (g === undefined) {
		throw new Error("the example has no topic g");
	}

	test("falls on the lines among the topics below it and into it", () => {
		expect(litBy(g, map.lines).map(lineId)).toEqual([
			"line-c-e",
			"line-b-g",
			"line-e-g",
		]);
	});

	test("falls on the lines out of it and among the topics above it", () => {
		const a = topics.byId.get("x.a");

		expect(a && litBy(a, map.lines).map(lineId)).toEqual([
			"line-a-d",
			"line-d-f",
		]);
	});

	test("is written as rules naming exactly the topics below and above it", () => {
		const rules = lightRules(topics, map);
		const pointedAt = `.s-map:has(#${nodeId(g)}:is(:hover,:focus-visible))`;
		const named = rules
			.split("\n")
			.map((line) => line.replace(/,$/, ""))
			.filter((line) => line.startsWith(pointedAt))
			.map((line) => line.slice(pointedAt.length).trim().replace(/\{.*$/, ""));

		expect(named).toEqual([
			"#map-b",
			"#map-c",
			"#map-e",
			"#line-c-e",
			"#line-b-g",
			"#line-e-g",
			"#say-g",
		]);
	});
});
