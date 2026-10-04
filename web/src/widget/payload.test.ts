import type { CallToolResult } from "@modelcontextprotocol/client";
import { describe, expect, test } from "vitest";
import {
	type AnswerResult,
	readAnswer,
	readEdited,
	readHandedTask,
	readScreen,
	readTaskStatus,
	readWaiting,
} from "./payload";
import {
	answered,
	askRefused,
	coming,
	editGone,
	editRefused,
	editSaved,
	exhausted,
	failure,
	fence,
	firstRun,
	firstRunRefused,
	inTrial,
	limited,
	moving,
	notComing,
	onTheCard,
	profileRead,
	profileRefused,
	refused,
	savedWords,
	staleAnswer,
	staleWait,
	standing,
	standingBefore,
	toldAgain,
	writing,
} from "./testing/lesson";

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
		["a task refused, with attempts left", refused, "refused"],
		["a task handed in for no open request", staleWait, "stale"],
		["the model's last attempt refused", exhausted, "exhausted"],
		["an ask no request could be opened from", askRefused, "stale"],
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
	test("of a task on its way keeps the request and whose card it is", () => {
		expect(readScreen(coming)).toEqual({
			screen: "coming",
			coming: { requestId: "req_fence", child: fence.child },
		});
	});

	test.each([
		["a task", fence, "task"],
		["a task on its way", coming, "coming"],
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
		[
			"the progress with a share past the whole way",
			{ ...standing, overall: { ...standing.overall, share: 101 } },
		],
		["the profile with no details", { ...profileRead, profile: null }],
		["a task on its way for no request", { ...coming, request_id: "" }],
		["a task on its way with no child", { ...coming, child: null }],
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

	test("of the progress keeps how far through its rank the rating has come, each topic's rank, the total of the skips and where the file is", () => {
		const shown = readScreen(standing);
		if (shown?.screen !== "progress") {
			throw new Error(`the progress reads as ${shown?.screen}`);
		}

		expect(shown.report.overall?.share).toBe(43);
		expect(shown.report.topics[0]).toMatchObject({
			topic: "logic.ordering",
			rank: 4,
			share: 27,
			compared: "ahead",
		});
		expect(shown.report.topics.at(-1)).toMatchObject({
			topic: "pigeonhole.basic",
			rank: null,
			answers: 0,
		});
		expect(shown.report.skipped).toBe(1);
		expect(shown.report.location?.file).toBe("mathtrail-profile.json");
	});

	test("of the progress needs no overall rating's number, which the card does not draw", () => {
		const { rating: _, ...overall } = standing.overall;

		const shown = readScreen({ ...standing, overall });

		expect(shown?.screen).toBe("progress");
	});

	test("of the progress from before the topics had ranks is read with none of what came after", () => {
		const shown = readScreen(standingBefore);
		if (shown?.screen !== "progress") {
			throw new Error(`the progress reads as ${shown?.screen}`);
		}

		expect(shown.report.overall?.share).toBeUndefined();
		expect(shown.report.topics[0]?.rank).toBeUndefined();
		expect(shown.report.skipped).toBeUndefined();
		expect(shown.report.location).toBeUndefined();
	});

	test("of the progress is read whatever a later release says of how a rank compares", () => {
		const later = {
			...standing,
			topics: [{ ...standing.topics[0], compared: "far_ahead" }],
		};

		expect(readScreen(later)?.screen).toBe("progress");
	});

	test("of the progress from before the map of mistakes is read with none", () => {
		const before = Object.fromEntries(
			Object.entries(standing).filter(([field]) => field !== "mistakes"),
		);

		const shown = readScreen(before);

		expect(shown?.screen === "progress" && shown.report.mistakes).toEqual([]);
	});

	test("of the progress reads how the ranks moved over each while, and nothing of a rating before it the card does not draw", () => {
		const shown = readScreen(moving);
		if (shown?.screen !== "progress") {
			throw new Error(`the progress reads as ${shown?.screen}`);
		}

		expect(shown.report.overall?.change).toEqual({
			last_task: { rank: 3, share: 47, moved: "back" },
			week: { rank: 2, share: 80, moved: "rank_up" },
		});
		expect(shown.report.topics[0]?.change).toEqual({
			last_task: { rank: 4, share: 27, moved: "same" },
			week: { rank: 3, share: 90, moved: "rank_up" },
		});
		expect(shown.report.topics[4]?.change).toBeUndefined();
	});

	test.each<[string, unknown, unknown]>([
		[
			"a while that does not read, as one that cannot be told",
			{
				last_task: { rank: 3, share: 47, moved: "back" },
				week: { rank: "two", moved: "rank_up" },
			},
			{ last_task: { rank: 3, share: 47, moved: "back" }, week: null },
		],
		[
			"a while left out, as one that cannot be told",
			{ week: { moved: "new" } },
			{ last_task: null, week: { moved: "new" } },
		],
		["moves that do not read, as none", "moved", undefined],
		[
			"a word a later release adds, as it is",
			{ last_task: null, week: { rank: 3, share: 40, moved: "leap" } },
			{ last_task: null, week: { rank: 3, share: 40, moved: "leap" } },
		],
	])(
		"of the progress reads %s, and the progress with it",
		(_, change, read) => {
			const shown = readScreen({
				...moving,
				overall: { ...moving.overall, change },
			});
			if (shown?.screen !== "progress") {
				throw new Error(`the progress reads as ${shown?.screen}`);
			}

			expect(shown.report.overall?.change).toEqual(read);
		},
	);

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

describe("a change sent from the form", () => {
	test("saved is read with the details as they now stand and the words for the model", () => {
		expect(readEdited(editSaved({ pseudonym: "Nova" }))).toEqual({
			kind: "saved",
			details: { ...standing.profile, pseudonym: "Nova" },
			told: savedWords,
		});
	});

	test("that changed nothing has no words for the model", () => {
		const unchanged = editSaved();
		expect(
			readEdited({
				...unchanged,
				structuredContent: {
					...(unchanged.structuredContent ?? {}),
					changed: false,
				},
			}),
		).toEqual({ kind: "saved", details: standing.profile, told: undefined });
	});

	test("refused is read field by field, by the codes of the rules broken", () => {
		expect(readEdited(editRefused)).toEqual({
			kind: "refused",
			problems: [
				{ field: "excluded_skills", code: "not_in_catalog" },
				{ field: "grade", code: "out_of_range" },
				{ field: "interests", code: "entry_length" },
				{ field: "pseudonym", code: "required" },
				{ field: "ui_language", code: "not_a_language" },
			],
		});
	});

	test("refused by a service from before the codes is read with codes it has no words for", () => {
		expect(
			readEdited({
				content: [],
				structuredContent: {
					screen: "profile",
					status: "rejected",
					code: "invalid_profile",
					problems: [{ field: "pseudonym", rule: "is required" }],
					profile: standing.profile,
				},
			}),
		).toEqual({
			kind: "refused",
			problems: [{ field: "pseudonym", code: "" }],
		});
	});

	test("for a profile no longer there is read as gone", () => {
		expect(readEdited(editGone)).toEqual({ kind: "gone" });
	});

	test.each<[string, CallToolResult]>([
		["a failure of the service", failure],
		[
			"a payload that does not read",
			{ content: [], structuredContent: { profile: 3 } },
		],
		["no payload at all", { content: [] }],
		[
			"a refusal that names no field",
			{
				content: [],
				structuredContent: { status: "rejected", profile: standing.profile },
			},
		],
		[
			"a status the card does not know",
			{
				content: [],
				structuredContent: { status: "limited", profile: standing.profile },
			},
		],
		[
			"a save that hands back no details",
			{ content: [], structuredContent: { screen: "profile", changed: true } },
		],
	])("is not saved after %s", (_, result) => {
		expect(readEdited(result)).toEqual({ kind: "failed" });
	});
});

describe("how the task a card waits for stands", () => {
	test.each<[string, CallToolResult, ReturnType<typeof readTaskStatus>]>([
		["being written", writing(), { kind: "writing", refused: 0 }],
		[
			"being written after two tries turned down",
			writing(2),
			{ kind: "writing", refused: 2 },
		],
		[
			"on the card",
			onTheCard,
			{
				kind: "task",
				handed: readHandedTask(onTheCard.structuredContent) ?? fence,
			},
		],
		["over", notComing, { kind: "over" }],
		[
			"gone with the profile",
			{ content: [], structuredContent: firstRun },
			{ kind: "over" },
		],
	])("is read %s", (_, result, status) => {
		expect(readTaskStatus(result)).toEqual(status);
	});

	test.each<[string, CallToolResult]>([
		["the service failed", failure],
		["nothing came", { content: [] }],
		[
			"a task that does not read as one",
			{ content: [], structuredContent: { ...fence, task: null } },
		],
		[
			"tries turned down that are no count",
			{ content: [], structuredContent: { screen: "coming", refused: -1 } },
		],
		[
			"a screen nobody names",
			{ content: [], structuredContent: { screen: "settings" } },
		],
	])("is unknown when %s", (_, result) => {
		expect(readTaskStatus(result)).toEqual({ kind: "unknown" });
	});

	test("never carries what the card was not handed: an answer in the payload stays out", () => {
		const status = readTaskStatus({
			content: [],
			structuredContent: {
				...fence,
				task: {
					...fence.task,
					correct_answer: "C",
					solution: "Count the posts.",
				},
			},
		});

		expect(JSON.stringify(status)).not.toContain("Count the posts.");
		expect(JSON.stringify(status)).not.toContain("correct_answer");
	});
});
