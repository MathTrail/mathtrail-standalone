import { describe, expect, test } from "vitest";
import { movesCounted, periodShown, type Reading, readingOf } from "./moves";
import type { Moves } from "./payload";

const told = { rank: 3, share: 40, moved: "forward" };

describe("the while the progress draws", () => {
	test.each<
		[string, Moves | undefined, Parameters<typeof periodShown>[0], string]
	>([
		[
			"the week, when it can be told",
			{ last_task: told, week: told },
			undefined,
			"week",
		],
		[
			"the last task, when only it can be told",
			{ last_task: told, week: null },
			undefined,
			"last_task",
		],
		[
			"the week, when neither can be told",
			{ last_task: null, week: null },
			undefined,
			"week",
		],
		["the week, when there are no moves at all", undefined, undefined, "week"],
		[
			"the one chosen, whatever can be told",
			{ last_task: told, week: told },
			"last_task",
			"last_task",
		],
		[
			"the one chosen, even one that cannot be told",
			{ last_task: told, week: null },
			"week",
			"week",
		],
	])("is %s", (_, moves, chosen, want) => {
		expect(periodShown(chosen, moves)).toBe(want);
	});
});

describe("a move, as the card draws it", () => {
	test.each<[string, Parameters<typeof readingOf>[0], Reading]>([
		["a while that cannot be told", null, { kind: "untold" }],
		["no move", { rank: 3, share: 40, moved: "same" }, { kind: "same" }],
		["a topic new to the while", { moved: "new" }, { kind: "new" }],
		[
			"a rank up",
			{ rank: 2, share: 80, moved: "rank_up" },
			{ kind: "moved", way: "gain", ranked: true, rank: 2, share: 80 },
		],
		[
			"forward within a rank",
			{ rank: 3, share: 20, moved: "forward" },
			{ kind: "moved", way: "gain", ranked: false, rank: 3, share: 20 },
		],
		[
			"back within a rank",
			{ rank: 3, share: 60, moved: "back" },
			{ kind: "moved", way: "loss", ranked: false, rank: 3, share: 60 },
		],
		[
			"a rank down",
			{ rank: 4, share: 5, moved: "rank_down" },
			{ kind: "moved", way: "loss", ranked: true, rank: 4, share: 5 },
		],
		[
			"a word a later release adds",
			{ rank: 3, share: 40, moved: "leap" },
			{ kind: "unknown" },
		],
		["a move with nowhere it stood", { moved: "forward" }, { kind: "unknown" }],
		[
			"a move with no share it stood at",
			{ rank: 3, moved: "back" },
			{ kind: "unknown" },
		],
	])("reads %s", (_, move, want) => {
		expect(readingOf(move)).toEqual(want);
	});
});

describe("the topics that moved", () => {
	test("count a gain and a topic new to the while as up, a step back as down, and nothing else", () => {
		const readings: Reading[] = [
			{ kind: "moved", way: "gain", ranked: true, rank: 2, share: 80 },
			{ kind: "moved", way: "gain", ranked: false, rank: 3, share: 20 },
			{ kind: "new" },
			{ kind: "moved", way: "loss", ranked: false, rank: 3, share: 60 },
			{ kind: "same" },
			{ kind: "untold" },
			{ kind: "unknown" },
		];

		expect(movesCounted(readings)).toEqual({ up: 3, down: 1 });
		expect(movesCounted([])).toEqual({ up: 0, down: 0 });
	});
});
