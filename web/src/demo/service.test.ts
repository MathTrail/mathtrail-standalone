import { afterEach, beforeEach, describe, expect, test, vi } from "vitest";
import { type Choice, dontKnow, letters } from "../widget/choices";
import { readAnswer, readHandedTask } from "../widget/payload";
import { answered, fence, rightAnswer } from "../widget/testing/lesson";
import type { DemoData } from "./data";
import { checkingTakes, demoHost, demoService } from "./service";

beforeEach(() => {
	vi.useFakeTimers({ toFake: ["setTimeout", "clearTimeout"] });
});

afterEach(() => {
	vi.useRealTimers();
});

// fenceData is the demo's data for the fence: B told wrong, C told right, and
// every other answer told as B is.
function fenceData(): DemoData {
	const handed = readHandedTask(fence);
	const wrong = readAnswer(answered(), fence.task.id);
	const right = readAnswer(rightAnswer, fence.task.id);
	if (
		handed === undefined ||
		wrong.kind !== "answered" ||
		right.kind !== "answered"
	) {
		throw new Error("the fence is no task answered");
	}
	const choices: Choice[] = [...letters, dontKnow];
	return {
		locale: "en",
		words: {},
		handed,
		results: Object.fromEntries(
			choices.map((choice) => [
				choice,
				choice === "C" ? right.result : { ...wrong.result, choice },
			]),
		) as DemoData["results"],
	};
}

describe("the service the demo answers for", () => {
	test("records an answer after a moment, as the page's data says the service would, with the hint as the card says it was used", async () => {
		const data = fenceData();
		const recorded = vi.fn();
		void demoService(data).recordAnswer("task_9", "C", true).then(recorded);

		await vi.advanceTimersByTimeAsync(checkingTakes - 1);
		expect(recorded).not.toHaveBeenCalled();
		await vi.advanceTimersByTimeAsync(1);
		expect(recorded).toHaveBeenCalledWith({
			kind: "answered",
			result: { ...data.results.C, task_id: "task_9", hint_used: true },
		});
	});

	// A child who answers again is rated from where the answer before left
	// the rating, as the service rates a child.
	test("moves the rating on from where the answer before left it", async () => {
		const data = fenceData();
		const service = demoService(data);
		const recorded: unknown[] = [];
		void service.recordAnswer("task_1", "B", false).then((told) => {
			recorded.push(told.kind === "answered" ? told.result.rating : told);
		});
		await vi.advanceTimersByTimeAsync(checkingTakes);
		void service.recordAnswer("task_2", "C", false).then((told) => {
			recorded.push(told.kind === "answered" ? told.result.rating : told);
		});
		await vi.advanceTimersByTimeAsync(checkingTakes);

		const wrong = data.results.B.rating;
		const right = data.results.C.rating;
		if (wrong === null || right === null) {
			throw new Error("the fence's answers move no rating");
		}
		const afterWrong = wrong.after;
		expect(recorded).toEqual([
			{ before: wrong.before, after: afterWrong },
			{ before: afterWrong, after: afterWrong + right.after - right.before },
		]);
	});

	test("says a task asked for is being written, then hands it out: the lesson's task, under an id of its own", async () => {
		const data = fenceData();
		const service = demoService(data);

		expect(await service.taskStatus("ask_1")).toEqual({
			kind: "writing",
			refused: 0,
		});
		const handed = await service.taskStatus("ask_1");
		expect(handed).toEqual({
			kind: "task",
			handed: {
				...data.handed,
				task: { ...data.handed.task, id: "ask_1_task" },
			},
		});
		expect(await service.taskStatus("ask_2")).toEqual({
			kind: "writing",
			refused: 0,
		});
	});

	test("saves no change to a profile, since the page keeps none", async () => {
		expect(
			await demoService(fenceData()).saveEdit({ lesson_topic: "" }),
		).toEqual({ kind: "failed" });
	});
});

describe("the chat around the card on the page", () => {
	test("hands a message the card sends to whoever shows it, and calls no tool and opens no page", async () => {
		const sent: string[] = [];
		const host = demoHost((text) => sent.push(text));

		await host.sendMessage("Another task");
		await host.tellModel("a line nobody reads");

		expect(sent).toEqual(["Another task"]);
		await expect(host.callTool("submit_answer", {})).rejects.toThrow(
			"the page calls no tool",
		);
		expect(host.canOpenLinks()).toBe(false);
		expect(await host.openLink("https://example.test/")).toBe(false);
	});
});
