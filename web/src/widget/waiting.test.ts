import { describe, expect, test } from "vitest";
import { moments, stepsOf } from "./waiting";

describe("the course of a task being handed in", () => {
	test("has the topic chosen and the task being written while its arguments come", () => {
		expect(stepsOf("started")).toEqual([
			{ key: "waiting.step.topic", status: "done" },
			{ key: "waiting.step.writing", status: "active" },
			{ key: "waiting.step.answers", status: "waiting" },
			{ key: "waiting.step.ready", status: "waiting" },
		]);
	});

	test("takes up the checks once the arguments have come whole, and is never ready", () => {
		expect(stepsOf("running")).toEqual([
			{ key: "waiting.step.topic", status: "done" },
			{ key: "waiting.step.writing", status: "done" },
			{ key: "waiting.step.answers", status: "active" },
			{ key: "waiting.step.ready", status: "waiting" },
		]);
	});
});

describe("the wait's own moments", () => {
	test("show it after a moment, and give it up long after a task's checks would end", () => {
		expect(moments.shown).toBeLessThan(1000);
		expect(moments.givenUp).toBeGreaterThanOrEqual(60_000);
	});
});
