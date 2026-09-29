import type { CallToolResult } from "@modelcontextprotocol/client";
import { describe, expect, test } from "vitest";
import { readAnswer, readHandedTask } from "./payload";
import {
	answered,
	failure,
	fence,
	staleAnswer,
	toldAgain,
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
