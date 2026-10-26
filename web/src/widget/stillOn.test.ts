import { afterEach, describe, expect, test, vi } from "vitest";
import type { TaskStatus } from "./payload";
import type { Service } from "./service";
import {
	askAgainAfter,
	questionLongest,
	requestIdOf,
	stillOnTheCard,
} from "./stillOn";
import { fence } from "./testing/lesson";

afterEach(() => {
	vi.useRealTimers();
});

// asking is a service that tells how a request stands with the answers given,
// one a question and the last again after them, and keeps the requests it was
// asked about.
function asking(...answers: TaskStatus[]) {
	const asked: string[] = [];
	const service: Service = {
		recordAnswer: () => Promise.resolve({ kind: "failed" }),
		taskStatus: (requestId) => {
			asked.push(requestId);
			return Promise.resolve(
				answers[Math.min(asked.length - 1, answers.length - 1)] ?? {
					kind: "unknown",
				},
			);
		},
		readProgress: () => Promise.resolve({ kind: "failed" }),
		saveEdit: () => Promise.resolve({ kind: "failed" }),
	};
	return { service, asked };
}

const setAside: TaskStatus = { kind: "over", anotherWaits: true };
const over: TaskStatus = { kind: "over" };
const unknown: TaskStatus = { kind: "unknown" };
const writing: TaskStatus = { kind: "writing", refused: 0 };
const onTheCard: TaskStatus = { kind: "task", handed: fence };

// told is what the check says of a card whose task is tsk_0b5f, the clock
// moved on as far as the check can wait.
async function told(service: Service): Promise<boolean | undefined> {
	vi.useFakeTimers();
	const said = stillOnTheCard(service, "tsk_0b5f");
	await vi.advanceTimersByTimeAsync(2 * questionLongest + askAgainAfter);
	return said;
}

// answering is a service whose answers come only when the case gives them,
// and which keeps the requests it was asked about.
function answering() {
	const asked: string[] = [];
	const answers: ((status: TaskStatus) => void)[] = [];
	const service: Service = {
		...asking().service,
		taskStatus: (requestId) => {
			asked.push(requestId);
			return new Promise((resolve) => {
				answers.push(resolve);
			});
		},
	};
	return { service, asked, answers };
}

describe("the request of a task", () => {
	test("is named after the task, as the service names a task after its request", () => {
		expect(requestIdOf("tsk_0b5f")).toBe("req_0b5f");
	});

	test.each(["task_fence", "tsk_", "req_0b5f", ""])(
		"is not named for a task whose id is %j",
		(taskId) => {
			expect(requestIdOf(taskId)).toBeUndefined();
		},
	);
});

describe("whether a card may still ask", () => {
	test("is no once the service says twice that another task waits on the child's card in its task's place", async () => {
		const { service, asked } = asking(setAside);

		expect(await told(service)).toBe(false);
		expect(asked).toEqual(["req_0b5f", "req_0b5f"]);
	});

	test("asks the second time only once askAgainAfter has passed", async () => {
		vi.useFakeTimers();
		const { service, asked } = asking(setAside);

		const said = stillOnTheCard(service, "tsk_0b5f");
		await vi.advanceTimersByTimeAsync(askAgainAfter - 1);
		expect(asked).toHaveLength(1);
		await vi.advanceTimersByTimeAsync(1);

		expect(asked).toHaveLength(2);
		expect(await said).toBe(false);
	});

	test.each<[string, TaskStatus[], number]>([
		["the task is on the card", [onTheCard], 1],
		["no task waits for an answer, which an ask skips nothing of", [over], 1],
		["the word was a read too early", [setAside, onTheCard], 2],
		["the task is being written, which no card's task is", [writing], 1],
		["no answer came", [unknown], 1],
	])("is yes when %s", async (_, answers, questions) => {
		const { service, asked } = asking(...answers);

		expect(await told(service)).toBe(true);
		expect(asked).toHaveLength(questions);
	});

	test("is yes when the question fails outright", async () => {
		const service: Service = {
			...asking().service,
			taskStatus: () => Promise.reject(new Error("the host went away")),
		};

		expect(await told(service)).toBe(true);
	});

	test("is asked of nobody for a task whose request the card cannot name, so that the card goes on at once", () => {
		const { service, asked } = asking(setAside);

		expect(stillOnTheCard(service, "task_fence")).toBeUndefined();
		expect(asked).toEqual([]);
	});

	test("is yes once an answer has taken longer than the card waits, and asks nothing more", async () => {
		vi.useFakeTimers();
		const { service, asked, answers } = answering();

		const said = stillOnTheCard(service, "tsk_0b5f");
		await vi.advanceTimersByTimeAsync(questionLongest - 1);
		const before = await Promise.race([said, Promise.resolve("waiting")]);
		await vi.advanceTimersByTimeAsync(1);

		expect(before).toBe("waiting");
		expect(await said).toBe(true);
		answers[0]?.(setAside);
		await vi.advanceTimersByTimeAsync(askAgainAfter + questionLongest);
		expect(asked).toHaveLength(1);
	});

	test("is yes when the second answer takes longer than the card waits", async () => {
		vi.useFakeTimers();
		const { service, asked, answers } = answering();

		const said = stillOnTheCard(service, "tsk_0b5f");
		await vi.advanceTimersByTimeAsync(0);
		answers[0]?.(setAside);
		await vi.advanceTimersByTimeAsync(askAgainAfter);
		expect(asked).toHaveLength(2);
		await vi.advanceTimersByTimeAsync(questionLongest);

		expect(await said).toBe(true);
	});

	test("leaves no clock running once it is told", async () => {
		const { service } = asking(setAside);

		await told(service);

		expect(vi.getTimerCount()).toBe(0);
	});
});
