import type { CallToolResult } from "@modelcontextprotocol/client";
import { afterEach, describe, expect, test, vi } from "vitest";
import {
	buttonIn,
	drawCard as drawOnHost,
	press,
	takeDown,
} from "./testing/card";
import type { ToolCall } from "./testing/host";
import {
	exhausted,
	fence,
	limited,
	progress,
	refused,
	staleWait,
} from "./testing/lesson";

let root: HTMLElement;

afterEach(() => {
	takeDown(root);
	vi.restoreAllMocks();
});

// service answers the widget's calls as the service would: the progress read.
function service(_: ToolCall): CallToolResult {
	return progress;
}

// drawCard draws the card a host hands payload to, and returns what the host
// hears from it.
async function drawCard(payload: object) {
	const drawn = await drawOnHost(payload, { tools: service });
	root = drawn.root;
	return drawn.heard;
}

const button = (label: string) => buttonIn(root, label);
const text = (selector: string) => root.querySelector(selector)?.textContent;

describe("a card a task did not come to", () => {
	test.each([
		["a try that did not pass", refused, "This try didn't pass the checks."],
		["a request that is over", staleWait, "No task here."],
		["the last try", exhausted, "This task didn't work out."],
	])(
		"after %s says so, and where the next one comes, with nothing that moves or asks",
		async (_, payload, says) => {
			await drawCard(payload);

			expect(text(".mt-bar-name")).toBe("Comet");
			expect(text(".mt-badge")).toBe("Olympiad coach · Grade 3");
			expect(text(".mt-verdict-line")).toBe(says);
			expect(text(".mt-verdict-detail")).toBe(
				"The next one comes below; if none does, ask for a task in the chat.",
			);
			// It cannot see the next try being written, which comes in a card of its
			// own, and a host may hold a card's message: no course, no button.
			expect(root.querySelector(".mt-gen")).toBeNull();
			expect(root.querySelector(".mt-btns")).toBeNull();
		},
	);

	test("shows nothing of why the model's task was refused", async () => {
		await drawCard(refused);

		expect(root.textContent).not.toContain("solver");
		expect(root.textContent).not.toContain(refused.reasons[0].messages[0]);
	});

	test("speaks the language the parent chose for the lessons", async () => {
		await drawCard({
			...refused,
			child: { ...refused.child, ui_language: "ru" },
		});

		expect(text(".mt-verdict-line")).toBe("Эта попытка не прошла проверку.");
	});

	test("opens the progress, and comes back to the card as it was", async () => {
		const heard = await drawCard(refused);

		press(button("Comet Profile & progress"));
		await vi.waitFor(() =>
			expect(text(".mt-rank-name")).toBe("River crossing"),
		);
		// A card with no task leads back to what it shows.
		press(button("Back"));

		expect(text(".mt-verdict-line")).toBe("This try didn't pass the checks.");
		expect(heard.calls).toEqual([{ name: "read_progress", arguments: {} }]);
	});

	test("leaves the focus where the child put it", async () => {
		vi.spyOn(document, "hasFocus").mockReturnValue(false);
		await drawCard(refused);

		expect(document.activeElement).toBe(document.body);
	});
});

describe("a card refused for the day", () => {
	test("says when there will be more, with nothing to wait for or press", async () => {
		await drawCard(limited);

		expect(text(".mt-verdict-line")).toBe(
			"There are no more new tasks today — there will be more tomorrow.",
		);
		expect(text(".mt-verdict-detail")).toBe(
			"You can still look at your progress, or go back over the last task.",
		);
		expect(root.querySelector(".mt-gen")).toBeNull();
		expect(root.querySelector("button")).toBeNull();
		expect(root.querySelector(".mt-bar")).toBeNull();
		expect(root.querySelector(".mt-badge")).toBeNull();
	});

	test("that knows whose it is leads to the progress", async () => {
		await drawCard({ ...limited, child: fence.child });

		expect(text(".mt-bar-name")).toBe("Comet");
		expect(text(".mt-badge")).toBe("Olympiad coach · Grade 3");
	});
});
