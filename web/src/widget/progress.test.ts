import { describe, expect, test } from "vitest";
import { type Progress, type ProgressEvent, progressAfter } from "./progress";

// progressFrom is the progress after events, from a card with the progress
// closed.
function progressFrom(...events: ProgressEvent[]): Progress | undefined {
	return events.reduce<Progress | undefined>(progressAfter, undefined);
}

describe("the progress", () => {
	test("is read once opened, and closed by the way back", () => {
		const opened = progressFrom({ type: "opened", opening: 1 });
		expect(opened).toEqual({ state: "reading", opening: 1 });

		const read = progressAfter(opened, {
			type: "read",
			opening: 1,
			payload: { screen: "progress" },
		});
		expect(read).toEqual({
			state: "read",
			opening: 1,
			payload: { screen: "progress" },
		});

		expect(progressAfter(read, { type: "closed" })).toBeUndefined();
	});

	test("that fails says so", () => {
		expect(
			progressFrom(
				{ type: "opened", opening: 1 },
				{ type: "failed", opening: 1 },
			),
		).toEqual({ state: "failed", opening: 1 });
	});

	test("opened while open stays the opening it was", () => {
		const opened = progressFrom({ type: "opened", opening: 1 });

		expect(progressAfter(opened, { type: "opened", opening: 2 })).toBe(opened);
	});

	test("read after it was closed is not opened again", () => {
		const closed = progressFrom(
			{ type: "opened", opening: 1 },
			{ type: "closed" },
		);

		expect(
			progressAfter(closed, { type: "read", opening: 1, payload: {} }),
		).toBeUndefined();
		expect(
			progressAfter(closed, { type: "failed", opening: 1 }),
		).toBeUndefined();
	});

	test("opened again is settled by its own reply, not the one before", () => {
		const reopened = progressFrom(
			{ type: "opened", opening: 1 },
			{ type: "closed" },
			{ type: "opened", opening: 2 },
		);

		expect(progressAfter(reopened, { type: "failed", opening: 1 })).toBe(
			reopened,
		);
		expect(
			progressAfter(reopened, { type: "read", opening: 1, payload: "old" }),
		).toBe(reopened);
		expect(
			progressAfter(reopened, { type: "read", opening: 2, payload: "new" }),
		).toEqual({ state: "read", opening: 2, payload: "new" });
	});

	test("read is not read again by a second reply", () => {
		const read = progressFrom(
			{ type: "opened", opening: 1 },
			{ type: "read", opening: 1, payload: "first" },
		);

		expect(
			progressAfter(read, { type: "read", opening: 1, payload: "second" }),
		).toBe(read);
		expect(progressAfter(read, { type: "failed", opening: 1 })).toBe(read);
	});
});
