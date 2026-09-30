import type { CallToolResult } from "@modelcontextprotocol/client";
import { describe, expect, test } from "vitest";
import {
	type AnswerResult,
	readAnswer,
	readHandedTask,
	readScreen,
	readWaiting,
} from "./payload";
import {
	answered,
	exhausted,
	failure,
	fence,
	firstRun,
	firstRunRefused,
	inTrial,
	limited,
	profileRead,
	profileRefused,
	refused,
	staleAnswer,
	standing,
	toldAgain,
} from "./testing/lesson";

// staleRequest is a task handed in for a request no longer open, with no task
// on the card: the card waits for the task the model is told to ask for.
const staleRequest = {
	screen: "waiting",
	status: "stale",
	code: "stale_request",
	last_answer: null,
	child: fence.child,
	task: null,
};

describe("a task handed to the card", () => {
	test("is read with what the child may see and whose card it is", () => {
		expect(readHandedTask(fence)).toEqual({
			screen: "task",
			child: fence.child,
			task: fence.task,
		});
	});

	test("is read when it is handed out again to the card it is on", () => {
		const again = { ...fence, status: "stale", code: "stale_request" };

		expect(readHandedTask(again)?.task.id).toBe(fence.task.id);
	});

	test.each([
		["a waiting screen", { ...fence, screen: "waiting", task: null }],
		["the progress", { screen: "progress", profile: null }],
		[
			"a grade that is no number",
			{ ...fence, child: { ...fence.child, grade: "3" } },
		],
		[
			"an option missing",
			{
				...fence,
				task: {
					...fence.task,
					options: { A: "3", B: "4", C: "5", D: "6" },
				},
			},
		],
		["no child", { ...fence, child: null }],
		["no task", { ...fence, task: null }],
		["nothing", undefined],
	])("is not read from %s", (_, payload) => {
		expect(readHandedTask(payload)).toBeUndefined();
	});
});

describe("an answer sent from the card", () => {
	test("is recorded with the result the service recorded", () => {
		const outcome = readAnswer(toldAgain, fence.task.id);

		expect(outcome.kind).toBe("answered");
		expect(outcome.kind === "answered" && outcome.result.choice).toBe("D");
	});

	test("from before mistakes were marked as repeating is recorded with its trap as one that does not", () => {
		const outcome = readAnswer(
			answered({
				trap: {
					id: "fence_gaps",
					text: "Counted the gaps instead of the posts.",
				} as AnswerResult["trap"],
			}),
			fence.task.id,
		);

		expect(outcome.kind === "answered" && outcome.result.trap).toEqual({
			id: "fence_gaps",
			text: "Counted the gaps instead of the posts.",
			repeated: false,
		});
	});

	test("is closed when the task is no longer the one being solved", () => {
		expect(readAnswer(staleAnswer, fence.task.id)).toEqual({ kind: "closed" });
		expect(
			readAnswer(
				{
					content: [],
					structuredContent: {
						screen: "first_run",
						status: "stale",
						code: "stale_task",
						result: null,
					},
				},
				fence.task.id,
			),
		).toEqual({ kind: "closed" });
	});

	test.each<[string, CallToolResult]>([
		["the service failed", failure],
		["the result is another task's", answered({ task_id: "task_other" })],
		[
			"the answer itself was refused",
			{
				content: [],
				structuredContent: {
					screen: "task",
					status: "rejected",
					code: "invalid_arguments",
					result: null,
				},
			},
		],
		[
			"the result does not read",
			{
				content: [],
				structuredContent: { screen: "result", result: { choice: "F" } },
			},
		],
		["there is no payload", { content: [] }],
	])("is not recorded when %s", (_, result) => {
		expect(readAnswer(result, fence.task.id)).toEqual({ kind: "failed" });
	});
});

describe("a wait for the next task", () => {
	test.each([
		["a task refused, with attempts left", refused, "working"],
		["a task handed in for no open request", staleRequest, "working"],
		["the model's last attempt refused", exhausted, "exhausted"],
	])("after %s is read with whose card it is", (_, payload, kind) => {
		expect(readWaiting(payload)).toEqual({ kind, child: fence.child });
	});

	test("refused for the day is read whether or not it says whose card it is", () => {
		expect(readWaiting(limited)).toEqual({ kind: "limit", child: undefined });
		expect(readWaiting({ ...limited, child: fence.child })).toEqual({
			kind: "limit",
			child: fence.child,
		});
	});

	test.each([
		["a task", fence],
		["a wait that does not say whose card it is", { ...refused, child: null }],
		[
			"a grade that is no number",
			{ ...refused, child: { ...fence.child, grade: "3" } },
		],
		["the progress", { screen: "progress", profile: null }],
		["nothing", undefined],
	])("is not read from %s", (_, payload) => {
		expect(readWaiting(payload)).toBeUndefined();
	});
});

describe("the screen a payload draws", () => {
	test.each([
		["a task", fence, "task"],
		["a wait", refused, "waiting"],
		["the progress", standing, "progress"],
		["the progress in the trial series", inTrial, "progress"],
		["the profile", profileRead, "profile"],
		["a refused change to the profile", profileRefused, "profile"],
		["the first sign-in", firstRun, "first_run"],
		["a refused first profile", firstRunRefused, "first_run"],
	])("is read from %s", (_, payload, screen) => {
		expect(readScreen(payload)?.screen).toBe(screen);
	});

	test.each([
		["a result, which no tool draws a card for", { screen: "result" }],
		["a screen nobody names", { screen: "settings" }],
		["the progress with no profile", { ...standing, profile: null }],
		[
			"the progress with a topic's rating as text",
			{
				...standing,
				topics: [{ ...standing.topics[0], rating: "1712" }],
			},
		],
		[
			"the progress with a mistake made no times",
			{ ...standing, mistakes: [{ trap: "missed_case", times: 0 }] },
		],
		[
			"the progress with a mistake made half a time",
			{ ...standing, mistakes: [{ trap: "missed_case", times: 1.5 }] },
		],
		["the profile with no details", { ...profileRead, profile: null }],
		["nothing", undefined],
	])("is none for %s", (_, payload) => {
		expect(readScreen(payload)).toBeUndefined();
	});

	test("of the profile keeps the details and where the file is, and drops the parent's notes", () => {
		const shown = readScreen(profileRead);

		expect(shown).toEqual({
			screen: "profile",
			profile: {
				details: {
					pseudonym: "Comet",
					grade: 3,
					interests: ["space", "animals", "football"],
					excluded_skills: ["division_with_remainder"],
					ui_language: "ru",
				},
				location: {
					folder: "MathTrail",
					file: "mathtrail-profile.json",
					others: [{ file: "mathtrail-profile (1).json" }],
				},
				refused: false,
			},
		});
		expect(JSON.stringify(shown)).not.toContain("Loses heart");
	});

	test("of the progress from before the map of mistakes is read with none", () => {
		const before = Object.fromEntries(
			Object.entries(standing).filter(([field]) => field !== "mistakes"),
		);

		const shown = readScreen(before);

		expect(shown?.screen === "progress" && shown.report.mistakes).toEqual([]);
	});

	test("of the progress keeps of each mistake its name and how many times, and nothing of a task", () => {
		const shown = readScreen({
			...standing,
			mistakes: [
				{
					trap: "missed_case",
					times: 3,
					task_id: "task_fence",
					text: "Counted the gaps instead of the posts.",
					answer: "C",
				},
			],
		});

		expect(shown?.screen === "progress" && shown.report.mistakes).toEqual([
			{ trap: "missed_case", times: 3 },
		]);
	});

	test("says whether a change asked for was refused", () => {
		const refusedChange = readScreen(profileRefused);
		const refusedFirst = readScreen(firstRunRefused);

		expect(
			refusedChange?.screen === "profile" && refusedChange.profile.refused,
		).toBe(true);
		expect(refusedFirst).toEqual({
			screen: "first_run",
			firstRun: { refused: true },
		});
		expect(readScreen(firstRun)).toEqual({
			screen: "first_run",
			firstRun: { refused: false },
		});
	});
});
