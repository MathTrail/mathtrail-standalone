import { describe, expect, test } from "vitest";
import { moments, stepsOf } from "./waiting";

describe("the course of a task a card waits for", () => {
	test("has the topic and difficulty being picked while the ask is answered", () => {
		expect(stepsOf("choosing")).toEqual([
			{ key: "waiting.step.topic", status: "active" },
			{ key: "waiting.step.writing", status: "waiting" },
			{ key: "waiting.step.answers", status: "waiting" },
			{ key: "waiting.step.ready", status: "waiting" },
		]);
	});

	test("has the topic picked and the task being written, and claims neither the checks nor the task ready while it is awaited", () => {
		expect(stepsOf("writing")).toEqual([
			{ key: "waiting.step.topic", status: "done" },
			{ key: "waiting.step.writing", status: "active" },
			{ key: "waiting.step.answers", status: "waiting" },
			{ key: "waiting.step.ready", status: "waiting" },
		]);
	});

	test("has the task written and checked once it has come, with nothing under way", () => {
		expect(stepsOf("checked")).toEqual([
			{ key: "waiting.step.topic", status: "done" },
			{ key: "waiting.step.writing", status: "done" },
			{ key: "waiting.step.answers", status: "done" },
			{ key: "waiting.step.ready", status: "waiting" },
		]);
	});

	test("has every step done once the task is ready", () => {
		expect(stepsOf("ready")).toEqual([
			{ key: "waiting.step.topic", status: "done" },
			{ key: "waiting.step.writing", status: "done" },
			{ key: "waiting.step.answers", status: "done" },
			{ key: "waiting.step.ready", status: "done" },
		]);
	});
});

describe("the wait's own moments", () => {
	test("show it after a moment, ask often, and call it long well after a task's minute", () => {
		expect(moments.shown).toBeLessThan(1000);
		expect(moments.ask).toBeLessThanOrEqual(5000);
		expect(moments.askSlowly).toBeGreaterThan(moments.ask);
		expect(moments.slow).toBeGreaterThanOrEqual(90_000);
		expect(moments.spread).toBeLessThan(moments.ask);
	});

	test("finish the course in under a second, so that a task that has come is not held back", () => {
		expect(moments.beat).toBeGreaterThan(0);
		expect(moments.held).toBeGreaterThan(0);
		expect(moments.beat + moments.held).toBeLessThan(1000);
	});
});
