import { render } from "preact";
import { act } from "preact/test-utils";
import { afterEach, describe, expect, test } from "vitest";
import { type Service, useService } from "./service";

const root = document.createElement("div");

afterEach(() => {
	act(() => render(null, root));
});

describe("the service of a card drawn where nothing answers for it", () => {
	// A card drawn still on a page asks nobody: every question it could ask
	// ends as one the service never answered.
	test("records no answer, knows no task's standing, takes no task, reads no progress and saves no change", async () => {
		let asked: Service | undefined;
		function Card() {
			asked = useService();
			return null;
		}
		act(() => render(<Card />, root));
		if (asked === undefined) {
			throw new Error("the card was given no service");
		}

		expect(await asked.recordAnswer("task_fence", "B", false)).toEqual({
			kind: "failed",
		});
		expect(await asked.taskStatus("req_1")).toEqual({ kind: "unknown" });
		expect(await asked.takeTask("task_fence")).toEqual({ kind: "failed" });
		expect(await asked.readProgress()).toEqual({ kind: "failed" });
		expect(await asked.saveEdit({ lesson_topic: "" })).toEqual({
			kind: "failed",
		});
	});
});
