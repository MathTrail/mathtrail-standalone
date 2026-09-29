import type { CallToolResult } from "@modelcontextprotocol/client";
import { act } from "preact/test-utils";
import { afterEach, beforeEach, describe, expect, test, vi } from "vitest";
import english from "../../locales/en.json";
import {
	buttonIn,
	drawCard as drawOnHost,
	press,
	takeDown,
} from "./testing/card";
import type { ToolCall } from "./testing/host";
import { exhausted, fence, limited, progress, refused } from "./testing/lesson";
import { warmUps } from "./waiting";

let root: HTMLElement;

beforeEach(() => {
	// The clock of the wait is the test's; what the protocol does in between is
	// delivered at once, and needs none.
	vi.useFakeTimers({ toFake: ["setTimeout", "clearTimeout", "Date"] });
	vi.setSystemTime(0);
});

afterEach(() => {
	takeDown(root);
	vi.useRealTimers();
	vi.restoreAllMocks();
});

// service answers the widget's calls as the service would: the progress read.
function service(_: ToolCall): CallToolResult {
	return progress;
}

// drawCard draws the card a host hands payload to, and returns what the host
// hears from it.
async function drawCard(
	payload: object,
	options: { refuseMessages?: boolean } = {},
) {
	const drawn = await drawOnHost(payload, { tools: service, ...options });
	root = drawn.root;
	return drawn.heard;
}

// wait lets seconds of the wait pass.
function wait(seconds: number): void {
	act(() => {
		vi.advanceTimersByTime(seconds * 1000);
	});
}

const button = (label: string) => buttonIn(root, label);

const text = (selector: string) => root.querySelector(selector)?.textContent;
const statuses = () =>
	[...root.querySelectorAll(".mt-gen li")].map((step) =>
		step.getAttribute("data-status"),
	);
const shownButtons = () =>
	[...root.querySelectorAll<HTMLButtonElement>(".mt-btns .mt-btn")].map(
		(shown) => shown.textContent,
	);

describe("a card waiting for the next task", () => {
	test("shows whose it is, and the course a task takes", async () => {
		await drawCard(refused);

		expect(text(".mt-bar-name")).toBe("Comet");
		expect(text(".mt-badge")).toBe("Olympiad coach · Grade 3");
		expect(text(".mt-gen-title")).toBe("Preparing the next task…");
		expect(
			[...root.querySelectorAll(".mt-gen li")].map((step) => step.textContent),
		).toEqual([
			"In progress: Picked topic and difficulty",
			"Waiting: Writing the task",
			"Waiting: Checking every answer",
			"Waiting: Checking readability for grade 3",
			"Waiting: Ready",
		]);
		expect(root.querySelector(".mt-gen ol")?.getAttribute("aria-live")).toBe(
			"polite",
		);
		expect(
			root.querySelector<HTMLInputElement>(".mt-field input")?.disabled,
		).toBe(true);
		expect(shownButtons()).toEqual(["I don't know", "Hint", "Another task"]);
		for (const shown of root.querySelectorAll<HTMLButtonElement>(".mt-btn")) {
			expect(shown.disabled).toBe(true);
		}
	});

	test("shows nothing of why the model's task was refused", async () => {
		await drawCard(refused);

		expect(root.textContent).not.toContain("solver");
		expect(root.textContent).not.toContain(refused.reasons[0].messages[0]);
	});

	test("moves along the course by the clock, and never says the task is ready", async () => {
		await drawCard(refused);

		wait(5);
		expect(statuses()).toEqual([
			"done",
			"active",
			"waiting",
			"waiting",
			"waiting",
		]);
		wait(30);
		expect(statuses()).toEqual([
			"done",
			"done",
			"active",
			"waiting",
			"waiting",
		]);
		wait(15);
		expect(statuses()).toEqual(["done", "done", "done", "active", "waiting"]);
		wait(69);
		expect(statuses()).toEqual(["done", "done", "done", "active", "waiting"]);
	});

	test("claims no numbers of the work it cannot see", async () => {
		await drawCard(refused);

		for (const seconds of [0, 5, 30, 15]) {
			wait(seconds);
			for (const step of root.querySelectorAll(".mt-gen li")) {
				expect(step.textContent?.replace("grade 3", "")).not.toMatch(/\d/);
			}
		}
	});

	test("offers a warm-up once the wait drags on, and another when asked", async () => {
		await drawCard(refused);

		wait(29);
		expect(root.querySelector(".mt-note")).toBeNull();
		wait(1);
		expect(text(".mt-note-label")).toBe("Warm-up");
		const first = warmUps.findIndex(
			(key) => english[key] === text(".mt-note p"),
		);
		expect(first).not.toBe(-1);
		expect(root.querySelector(".mt-note")?.closest("[aria-live]")).toBe(
			root.querySelector(".mt-news"),
		);

		press(button("Another warm-up"));

		expect(text(".mt-note p")).toBe(
			english[warmUps[(first + 1) % warmUps.length] ?? warmUps[0]],
		);
		expect(document.activeElement).toBe(button("Another warm-up"));
	});

	test("after two minutes says plainly that no task is coming, and asks again", async () => {
		const heard = await drawCard(refused);

		wait(119);
		expect(root.querySelector(".mt-gen")).not.toBeNull();
		wait(1);

		expect(root.querySelector(".mt-gen")).toBeNull();
		expect(root.querySelector(".mt-note")).toBeNull();
		expect(text(".mt-wait .mt-verdict-line")).toBe(
			"If no new task has appeared below, it isn't being prepared.",
		);
		expect(text(".mt-wait .mt-verdict-detail")).toBe(
			"Ask for one in the chat, or press the button to ask again.",
		);
		expect(shownButtons()).toEqual(["Ask again"]);
		expect(root.querySelector(".mt-btn-primary")?.textContent).toBe(
			"Ask again",
		);

		press(button("Ask again"));

		await vi.waitFor(() => expect(heard.messages).toEqual(["Another task"]));
		expect(statuses()).toEqual([
			"active",
			"waiting",
			"waiting",
			"waiting",
			"waiting",
		]);
		wait(119);
		expect(root.querySelector(".mt-gen")).not.toBeNull();
		wait(1);
		expect(shownButtons()).toEqual(["Ask again"]);
	});

	test("asks again once however often the button is pressed", async () => {
		const heard = await drawCard(refused);
		wait(120);

		act(() => {
			button("Ask again").click();
			button("Ask again").click();
		});

		await vi.waitFor(() => expect(heard.messages).toHaveLength(1));
		expect(root.querySelector(".mt-gen")).not.toBeNull();
	});

	test("whose ask does not reach the chat says so at once", async () => {
		vi.spyOn(console, "error").mockImplementation(() => {});
		const heard = await drawCard(refused, { refuseMessages: true });
		wait(120);

		press(button("Ask again"));

		await vi.waitFor(() => expect(heard.messages).toHaveLength(1));
		await vi.waitFor(() => expect(shownButtons()).toEqual(["Ask again"]));
		expect(text(".mt-wait .mt-verdict-line")).toBe(
			"If no new task has appeared below, it isn't being prepared.",
		);
		expect(document.activeElement).toBe(button("Ask again"));
	});

	test("after the model's last attempt says the task did not work out, and waits for the next", async () => {
		await drawCard(exhausted);

		expect(text(".mt-wait > .mt-verdict-line")).toBe(
			"This task didn't work out.",
		);
		expect(text(".mt-gen-title")).toBe("Preparing the next task…");
		expect(statuses()[0]).toBe("active");
	});

	test("asked again after the last attempt, waits for the new ask alone", async () => {
		await drawCard(exhausted);
		wait(120);

		press(button("Ask again"));

		expect(root.querySelector(".mt-gen")).not.toBeNull();
		expect(root.textContent).not.toContain("This task didn't work out.");
	});

	test("has the list of steps and the place for news heard apart, each once", async () => {
		await drawCard(refused);

		// The place for news is on the page before anything comes into it.
		expect(
			root.querySelector(".mt-news[aria-live='polite']")?.children,
		).toHaveLength(0);
		for (const seconds of [0, 30, 90]) {
			wait(seconds);
			const regions = [...root.querySelectorAll("[aria-live]")];
			expect(regions.length).toBeGreaterThan(0);
			for (const region of regions) {
				expect(region.parentElement?.closest("[aria-live]")).toBeNull();
			}
		}
	});

	test("speaks the language the parent chose for the cards", async () => {
		await drawCard({
			...refused,
			child: { ...refused.child, ui_language: "ru" },
		});

		expect(text(".mt-gen-title")).toBe("Готовим следующую задачу…");
		expect(root.querySelectorAll(".mt-gen li")[3]?.textContent).toBe(
			"Ждёт: Проверяем, понятно ли для 3 класса",
		);
	});

	test("opens the progress, and comes back to the wait as it went on", async () => {
		const heard = await drawCard(refused);

		press(button("Comet Profile & progress"));
		await vi.waitFor(() => expect(text(".mt-rating-num")).toBe("1573"));
		wait(35);
		// A card with no task leads back to what it shows.
		press(button("Back"));

		expect(statuses()[2]).toBe("active");
		expect(heard.calls).toEqual([{ name: "read_progress", arguments: {} }]);
	});
});

describe("a card drawn while the chat has the focus", () => {
	test("leaves the focus where the child put it", async () => {
		vi.spyOn(document, "hasFocus").mockReturnValue(false);
		await drawCard(refused);

		expect(document.activeElement).toBe(document.body);
		wait(120);
		expect(document.activeElement).toBe(document.body);
	});
});

describe("a card refused for the day", () => {
	test("says when there will be more, with nothing to wait for or press", async () => {
		await drawCard(limited);

		expect(text(".mt-verdict-line")).toBe(
			"There are no more new tasks today\u00a0— there will be more tomorrow.",
		);
		expect(text(".mt-verdict-detail")).toBe(
			"You can still look at your progress, or go back over the last task.",
		);
		expect(root.querySelector(".mt-gen")).toBeNull();
		expect(root.querySelector("button")).toBeNull();
		expect(root.querySelector("input")).toBeNull();
		expect(root.querySelector(".mt-bar")).toBeNull();
		expect(root.querySelector(".mt-badge")).toBeNull();
		wait(120);
		expect(root.querySelector(".mt-wait")).toBeNull();
	});

	test("that knows whose it is leads to the progress", async () => {
		await drawCard({ ...limited, child: fence.child });

		expect(text(".mt-bar-name")).toBe("Comet");
		expect(text(".mt-badge")).toBe("Olympiad coach · Grade 3");
	});
});
