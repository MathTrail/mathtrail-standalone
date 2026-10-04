import type { McpUiHostContext } from "@modelcontextprotocol/ext-apps";
import { act } from "preact/test-utils";
import { afterEach, beforeEach, describe, expect, test, vi } from "vitest";
import { openCard, takeDown } from "./testing/card";
import { toolInfoOf } from "./testing/host";
import { coming, fence, limited, writing } from "./testing/lesson";

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

// asking is the context of a host that drew the card for a task asked for.
const asking: McpUiHostContext = {
	locale: "en-US",
	toolInfo: toolInfoOf("next_task"),
};

// openOn starts the widget on a host with context, tells it the call's
// arguments — a task asked for in the language given, or nothing of the call
// when none is — and lets the card run its effects — Preact runs those of a
// card drawn outside a test's act on a timer of its own — so that the card's
// own clock starts now. The card that the answer draws is told its task is
// being written whenever it asks.
async function openOn(
	context: McpUiHostContext = asking,
	language: string | null = "en",
) {
	const opened = await openCard({ context, tools: () => writing() });
	root = opened.root;
	if (language !== null) {
		await opened.host.sendToolInput({ arguments: { language } });
	}
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

describe("a card a task is asked for on", () => {
	test("shows nothing for a moment, then the topic being picked", async () => {
		await openOn();

		wait(300);
		expect(root.querySelector(".mt-widget")).toBeNull();
		wait(1000);

		expect(text(".mt-gen-title")).toBe("Preparing the next task…");
		expect(labels()).toEqual([
			"In progress: Picked topic and difficulty",
			"Waiting: Writing the task",
			"Waiting: Checking every answer",
			"Waiting: Ready",
		]);
		// Nothing to press, and nothing of a child it does not know yet.
		expect(root.querySelector("button")).toBeNull();
		expect(root.querySelector(".mt-badge")).toBeNull();
	});

	test("turns into the wait for the task when the request is answered", async () => {
		const host = await openOn();
		wait(1000);

		await host.sendToolResult({ content: [], structuredContent: coming });

		await vi.waitFor(() =>
			expect(labels()).toEqual([
				"Done: Picked topic and difficulty",
				"In progress: Writing the task",
				"Waiting: Checking every answer",
				"Waiting: Ready",
			]),
		);
		expect(text(".mt-badge")).toBe("Olympiad coach · Grade 3");
	});

	test("does not flash a wait past when the answer comes with the call", async () => {
		const host = await openOn();
		const waited = everSeen(".mt-gen");

		wait(200);
		await host.sendToolResult({ content: [], structuredContent: limited });
		await vi.waitFor(() =>
			expect(text(".mt-verdict-line")).toMatch(
				/^There are no more new tasks today/,
			),
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

	test("says the task is taking long when no answer has come two minutes after the ask", async () => {
		await openOn();
		wait(1000);

		wait(110_000);
		expect(root.querySelector(".mt-verdict")).toBeNull();
		wait(10_000);

		expect(text(".mt-verdict-line")).toBe(
			"The task is taking longer than usual.",
		);
		expect(text(".mt-verdict-detail")).toBe(
			"If it doesn't appear here, ask for a task in the chat.",
		);
		expect(root.querySelector(".mt-gen")).not.toBeNull();
	});

	// The host's language may not be the lesson's: a parent reads Claude in
	// English and holds the lesson in Russian. The wait speaks the language
	// the task is asked in from its first word, and goes on in it once the
	// service has named the same lesson's.
	test("speaks the language the task is asked in, not the host's, and goes on in it", async () => {
		const host = await openOn(asking, "ru");
		wait(1000);

		expect(text(".mt-gen-title")).toBe("Готовим следующую задачу…");

		await host.sendToolResult({
			content: [],
			structuredContent: { ...coming, language: "ru" },
		});

		await vi.waitFor(() =>
			expect(labels()[0]).toBe("Готово: Выбрали тему и сложность"),
		);
		expect(text(".mt-gen-title")).toBe("Готовим следующую задачу…");
	});

	test("speaks the language the task is asked in when the ask is refused", async () => {
		const host = await openOn(asking, "ru");
		wait(1000);

		await host.sendToolResult({ content: [], structuredContent: limited });

		await vi.waitFor(() =>
			expect(text(".mt-verdict-line")).toMatch(
				/^На сегодня новых задач больше нет/,
			),
		);
	});

	test("says nothing before the answer when the host does not say what the task is asked in", async () => {
		const host = await openOn(asking, null);

		wait(1000);
		expect(root.querySelector(".mt-widget")).toBeNull();
		wait(120_000);
		expect(root.querySelector(".mt-widget")).toBeNull();

		await host.sendToolResult({ content: [], structuredContent: coming });
		await vi.waitFor(() =>
			expect(text(".mt-gen-title")).toBe("Preparing the next task…"),
		);
	});

	test("says in the host's language that the task was not finished, when it does not say what the task is asked in", async () => {
		const host = await openOn(asking, null);
		wait(1000);

		await host.sendToolCancelled({ reason: "user action" });

		await vi.waitFor(() =>
			expect(text(".mt-verdict-line")).toBe("This task wasn't finished."),
		);
	});

	test.each(["MathTrail:next_task", "mcp__MathTrail__next_task"])(
		"knows the ask by its name under a host's prefix, as in %s",
		async (name) => {
			await openOn({ toolInfo: toolInfoOf(name) });
			wait(1000);

			expect(root.querySelector(".mt-gen")).not.toBeNull();
		},
	);
});

describe("a card drawn for anything else", () => {
	test.each([
		[
			"a task handed in, from an earlier list of the tools",
			{ toolInfo: toolInfoOf("submit_task") },
		],
		["another tool", { toolInfo: toolInfoOf("get_progress") }],
		["a host that does not say", {}],
	])("draws nothing before its result, for %s", async (_, context) => {
		const host = await openOn(context, null);

		await host.sendToolInput({
			arguments: { language: "ru", task: { answer: "the-sealed-answer-C" } },
		});
		wait(1000);

		expect(root.querySelector(".mt-widget")).toBeNull();
	});

	test("draws the task it is handed, for a task handed in from an earlier list of the tools", async () => {
		const host = await openOn({ toolInfo: toolInfoOf("submit_task") }, null);

		await host.sendToolResult({ content: [], structuredContent: fence });

		await vi.waitFor(() =>
			expect(text(".mt-task-text")).toBe(fence.task.question),
		);
	});
});
