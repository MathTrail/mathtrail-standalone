import type { McpUiHostContext } from "@modelcontextprotocol/ext-apps";
import { act } from "preact/test-utils";
import { afterEach, beforeEach, describe, expect, test, vi } from "vitest";
import { openCard, takeDown } from "./testing/card";
import { toolInfoOf } from "./testing/host";
import { fence, refused } from "./testing/lesson";

let root: HTMLElement;

beforeEach(() => {
	// The card's clock is the test's; what the protocol does in between is
	// delivered at once, and needs none.
	vi.useFakeTimers({ toFake: ["setTimeout", "clearTimeout", "Date"] });
});

afterEach(() => {
	takeDown(root);
	vi.useRealTimers();
});

// handingIn is the context of a host that drew the card for a task being
// handed in.
const handingIn: McpUiHostContext = {
	locale: "en-US",
	toolInfo: toolInfoOf("submit_task"),
};

// openOn starts the widget on a host with context, tells it nothing of the call
// yet, and lets the card run its effects — Preact runs those of a card drawn
// outside a test's act on a timer of its own — so that the card's own clock
// starts now.
async function openOn(context: McpUiHostContext = handingIn) {
	const opened = await openCard({ context });
	root = opened.root;
	wait(100);
	return opened.host;
}

// wait lets milliseconds of the card's own clock pass.
function wait(milliseconds: number): void {
	act(() => {
		vi.advanceTimersByTime(milliseconds);
	});
}

const text = (selector: string) => root.querySelector(selector)?.textContent;
const statuses = () =>
	[...root.querySelectorAll(".mt-gen li")].map((step) =>
		step.getAttribute("data-status"),
	);
const labels = () =>
	[...root.querySelectorAll(".mt-gen li")].map((step) => step.textContent);

// everSeen watches the card for selector, from now on, and says whether it
// ever appeared, however briefly. It reads what each change added rather than
// the card as it is when the changes are told: by then the card may have
// dropped it again.
function everSeen(selector: string): () => boolean {
	const shows = (changes: MutationRecord[]) =>
		changes.some((change) =>
			[...change.addedNodes].some(
				(node) =>
					node instanceof Element &&
					(node.matches(selector) || node.querySelector(selector) !== null),
			),
		);
	let seen = root.querySelector(selector) !== null;
	const watching = new MutationObserver((changes) => {
		seen ||= shows(changes);
	});
	watching.observe(root, { childList: true, subtree: true });
	return () => {
		seen ||= shows(watching.takeRecords());
		watching.disconnect();
		return seen;
	};
}

describe("a card a task is being handed in to", () => {
	test("shows nothing for a moment, then the task being written", async () => {
		await openOn();

		wait(300);
		expect(root.querySelector(".mt-widget")).toBeNull();
		wait(1000);

		expect(text(".mt-gen-title")).toBe("Preparing the next task…");
		expect(labels()).toEqual([
			"Done: Picked topic and difficulty",
			"In progress: Writing the task",
			"Waiting: Checking every answer",
			"Waiting: Ready",
		]);
		// Nothing to press, and nothing of a child it does not know yet.
		expect(root.querySelector("button")).toBeNull();
		expect(root.querySelector(".mt-badge")).toBeNull();
	});

	test("takes up the checks once the task has come whole", async () => {
		const host = await openOn();
		wait(1000);

		await host.sendToolInput({ arguments: {} });

		await vi.waitFor(() =>
			expect(statuses()).toEqual(["done", "done", "active", "waiting"]),
		);
	});

	test("turns into the task when its result comes", async () => {
		const host = await openOn();
		wait(1000);
		await host.sendToolInput({ arguments: {} });

		await host.sendToolResult({ content: [], structuredContent: fence });

		await vi.waitFor(() =>
			expect(text(".mt-task-text")).toBe(fence.task.question),
		);
		expect(root.querySelector(".mt-gen")).toBeNull();
	});

	test("turns into the word that the try failed when a refusal comes", async () => {
		const host = await openOn();
		wait(1000);

		await host.sendToolResult({ content: [], structuredContent: refused });

		await vi.waitFor(() =>
			expect(text(".mt-verdict-line")).toBe("This try didn't pass the checks."),
		);
		expect(root.querySelector(".mt-gen")).toBeNull();
	});

	test("does not flash a wait past when the result comes with the call", async () => {
		const host = await openOn();
		const waited = everSeen(".mt-gen");

		await host.sendToolInput({ arguments: {} });
		wait(200);
		await host.sendToolResult({ content: [], structuredContent: fence });
		await vi.waitFor(() =>
			expect(text(".mt-task-text")).toBe(fence.task.question),
		);
		wait(1000);

		expect(waited()).toBe(false);
	});

	test("says the task was not finished when the call is cancelled", async () => {
		const host = await openOn();
		wait(1000);

		await host.sendToolCancelled({ reason: "user action" });

		await vi.waitFor(() =>
			expect(text(".mt-verdict-line")).toBe("This task wasn't finished."),
		);
		expect(text(".mt-verdict-detail")).toBe("Ask for a task in the chat.");
		expect(root.querySelector(".mt-gen")).toBeNull();
		expect(root.querySelector(".mt-verdict")?.closest("[aria-live]")).toBe(
			root.querySelector(".mt-news"),
		);
	});

	test("gives up two minutes after the call last moved on, counted afresh once the task has come whole", async () => {
		const host = await openOn();
		wait(1000);
		wait(100_000);
		await host.sendToolInput({ arguments: {} });
		await vi.waitFor(() => expect(statuses()[2]).toBe("active"));

		wait(100_000);
		expect(root.querySelector(".mt-gen")).not.toBeNull();
		wait(20_000);

		expect(text(".mt-verdict-line")).toBe(
			"If no new task has appeared below, it isn't being prepared.",
		);
		expect(text(".mt-verdict-detail")).toBe("Ask for a task in the chat.");
		expect(root.querySelector(".mt-gen")).toBeNull();
	});

	test.each(["MathTrail:submit_task", "mcp__MathTrail__submit_task"])(
		"knows the task by its name under a host's prefix, as in %s",
		async (name) => {
			await openOn({ toolInfo: toolInfoOf(name) });
			wait(1000);

			expect(root.querySelector(".mt-gen")).not.toBeNull();
		},
	);

	test("never shows the arguments the task was handed in with", async () => {
		const secret = "the-sealed-answer-C";
		const host = await openOn();

		await host.sendToolInputPartial({
			arguments: { task: { answer: secret } },
		});
		await host.sendToolInput({ arguments: { task: { answer: secret } } });
		wait(1000);
		await vi.waitFor(() => expect(statuses()[2]).toBe("active"));
		expect(root.innerHTML).not.toContain(secret);

		await host.sendToolResult({ content: [], structuredContent: fence });
		await vi.waitFor(() =>
			expect(root.querySelector(".mt-option")).not.toBeNull(),
		);
		expect(root.innerHTML).not.toContain(secret);
	});
});

describe("a card drawn for anything else", () => {
	test.each([
		["another tool", { toolInfo: toolInfoOf("get_progress") }],
		["a host that does not say", {}],
	])("draws nothing before its result, for %s", async (_, context) => {
		await openOn(context);

		wait(1000);

		expect(root.querySelector(".mt-widget")).toBeNull();
	});
});
