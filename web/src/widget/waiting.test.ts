import { describe, expect, test } from "vitest";
import {
	isLate,
	moments,
	stepsAt,
	type Wait,
	waitAfter,
	waitStart,
} from "./waiting";

describe("a wait", () => {
	test("starts its round afresh with each ask", () => {
		const lost: Wait = { round: 1, lost: true };

		expect(waitAfter(lost, { type: "asked", round: 2 })).toEqual({
			round: 2,
			lost: false,
		});
	});

	test("is lost with the ask of its round", () => {
		const asked = waitAfter(waitStart, { type: "asked", round: 1 });

		expect(waitAfter(asked, { type: "lost", round: 1 })).toEqual({
			round: 1,
			lost: true,
		});
	});

	test("is not lost with the ask of a round before", () => {
		const again = waitAfter(waitAfter(waitStart, { type: "asked", round: 1 }), {
			type: "asked",
			round: 2,
		});

		expect(waitAfter(again, { type: "lost", round: 1 })).toBe(again);
	});

	test("lost once is not lost again", () => {
		const lost: Wait = { round: 1, lost: true };

		expect(waitAfter(lost, { type: "lost", round: 1 })).toBe(lost);
	});

	test("is late once its ask is lost, or once it lasts past the deadline", () => {
		const asked: Wait = { round: 1, lost: false };

		expect(isLate(asked, 0)).toBe(false);
		expect(isLate(asked, moments.deadline - 1)).toBe(false);
		expect(isLate(asked, moments.deadline)).toBe(true);
		expect(isLate({ round: 1, lost: true }, 0)).toBe(true);
	});
});

describe("the course of a task", () => {
	const statuses = (seconds: number) =>
		stepsAt(seconds).map(({ status }) => status);

	test("names its steps in their order", () => {
		expect(stepsAt(0).map(({ key }) => key)).toEqual([
			"waiting.step.topic",
			"waiting.step.writing",
			"waiting.step.answers",
			"waiting.step.readability",
			"waiting.step.ready",
		]);
	});

	test.each([
		[0, ["active", "waiting", "waiting", "waiting", "waiting"]],
		[
			moments.writing - 1,
			["active", "waiting", "waiting", "waiting", "waiting"],
		],
		[moments.writing, ["done", "active", "waiting", "waiting", "waiting"]],
		[moments.answers, ["done", "done", "active", "waiting", "waiting"]],
		[moments.readability, ["done", "done", "done", "active", "waiting"]],
		[moments.deadline, ["done", "done", "done", "active", "waiting"]],
	])("stands after %i seconds as %o", (seconds, want) => {
		expect(statuses(seconds)).toEqual(want);
	});

	test("never claims the task is ready, and has one step under way at a time", () => {
		for (let seconds = 0; seconds <= 2 * moments.deadline; seconds++) {
			const now = statuses(seconds);
			expect(now.at(-1)).toBe("waiting");
			expect(now.filter((status) => status === "active")).toHaveLength(1);
			// Done, then under way, then to come: the course never goes back.
			expect(now.join(" ")).toMatch(/^(done )*active( waiting)*$/);
		}
	});
});
