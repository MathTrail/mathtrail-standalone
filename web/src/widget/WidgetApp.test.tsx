import { render } from "preact";
import { act } from "preact/test-utils";
import { afterEach, beforeEach, describe, expect, test, vi } from "vitest";
import type { Bridge, Host, ToolResult } from "./bridge";
import { fence } from "./testing/lesson";
import { WidgetApp } from "./WidgetApp";

// idleHost is a host the card asks nothing of in these tests.
const idleHost: Host = {
	callTool: () => Promise.reject(new Error("no tool is called here")),
	sendMessage: () => Promise.reject(new Error("no message is sent here")),
	tellModel: () => Promise.reject(new Error("the model is told nothing here")),
};

// heldBridge is a bridge whose results arrive, and whose host names its locale,
// when the test says, so the card can be caught before, between and after them.
function heldBridge(hostLocale?: string) {
	let latest: ToolResult | undefined;
	let locale = hostLocale;
	const listeners = new Set<() => void>();
	const tell = () => {
		for (const listener of listeners) {
			listener();
		}
	};
	const bridge: Bridge = {
		result: () => latest,
		locale: () => locale,
		subscribe(listener) {
			listeners.add(listener);
			return () => {
				listeners.delete(listener);
			};
		},
		connect: async () => {},
	};
	const deliver = (structuredContent: Record<string, unknown>) => {
		latest = { content: [], structuredContent };
		tell();
	};
	const changeLocale = (next: string) => {
		locale = next;
		tell();
	};
	return { bridge, deliver, changeLocale };
}

let root: HTMLElement;

beforeEach(() => {
	root = document.createElement("div");
});

// A card drawn by one test is taken down after it, with its subscription, so
// that nothing of it is still listening while the next one runs, and the page
// forgets the language it was told.
afterEach(() => {
	act(() => render(null, root));
	document.documentElement.removeAttribute("lang");
	document.documentElement.removeAttribute("dir");
});

describe("the card", () => {
	test("draws nothing before the first result", () => {
		const { bridge } = heldBridge();

		act(() => render(<WidgetApp bridge={bridge} host={idleHost} />, root));

		expect(root.innerHTML).toBe("");
	});

	test("shows the payload of the latest result", () => {
		const { bridge, deliver } = heldBridge();
		act(() => render(<WidgetApp bridge={bridge} host={idleHost} />, root));

		act(() => deliver({ screen: "waiting", attempt: 1 }));
		expect(root.textContent).toContain('"attempt": 1');

		act(() => deliver({ screen: "waiting", attempt: 2 }));
		expect(root.textContent).toContain('"attempt": 2');
		expect(root.textContent).not.toContain('"attempt": 1');
	});

	test("shows a result that arrived before it was drawn", () => {
		const { bridge, deliver } = heldBridge();
		deliver({ screen: "progress" });

		act(() => render(<WidgetApp bridge={bridge} host={idleHost} />, root));

		expect(root.textContent).toContain('"screen": "progress"');
	});

	test("shows a result that arrived between drawing and listening", async () => {
		const { bridge, deliver } = heldBridge();

		// Drawn outside act, the card has not subscribed yet when the result
		// comes: its effects run after the frame is painted.
		render(<WidgetApp bridge={bridge} host={idleHost} />, root);
		deliver({ screen: "result" });

		await vi.waitFor(() =>
			expect(root.textContent).toContain('"screen": "result"'),
		);
	});

	test("draws a task handed out as the task's card", () => {
		const { bridge, deliver } = heldBridge();
		act(() => render(<WidgetApp bridge={bridge} host={idleHost} />, root));

		act(() => deliver(fence));

		expect(root.querySelector(".mt-task-text")?.textContent).toBe(
			fence.task.question,
		);
		expect(root.querySelector(".stub")).toBeNull();
	});

	test("shows a task's payload as it arrived when it does not read as a task", () => {
		const { bridge, deliver } = heldBridge();
		act(() => render(<WidgetApp bridge={bridge} host={idleHost} />, root));

		act(() =>
			deliver({ ...fence, task: { ...fence.task, options: { A: "3" } } }),
		);

		expect(root.querySelector(".mt-option")).toBeNull();
		expect(root.textContent).toContain('"options"');
	});

	test("keeps markup inside a payload as text", () => {
		const { bridge, deliver } = heldBridge();
		act(() => render(<WidgetApp bridge={bridge} host={idleHost} />, root));

		act(() => deliver({ text: "<img src=x onerror=alert(1)>" }));

		expect(root.querySelector("img")).toBeNull();
		expect(root.textContent).toContain("<img src=x onerror=alert(1)>");
	});
});

describe("the card's language", () => {
	// The page as a card finds it: in a language, running a way, that the card
	// has to set rather than assume.
	beforeEach(() => {
		document.documentElement.lang = "ar";
		document.documentElement.dir = "rtl";
	});

	// languageOfPage is the language the page says its words are in, and the
	// way they run.
	const languageOfPage = () => [
		document.documentElement.lang,
		document.documentElement.dir,
	];

	test("is the host's", () => {
		const { bridge, deliver } = heldBridge("ru-RU");
		act(() => render(<WidgetApp bridge={bridge} host={idleHost} />, root));

		act(() =>
			deliver({
				screen: "task",
				child: { pseudonym: "Otter", grade: 2, ui_language: null },
			}),
		);

		expect(languageOfPage()).toEqual(["ru", "ltr"]);
	});

	test("is the one the parent chose for the cards, over the host's", () => {
		const { bridge, deliver } = heldBridge("en-US");
		act(() => render(<WidgetApp bridge={bridge} host={idleHost} />, root));

		act(() =>
			deliver({
				screen: "task",
				child: { pseudonym: "Otter", grade: 2, ui_language: "ru" },
			}),
		);
		expect(languageOfPage()).toEqual(["ru", "ltr"]);

		act(() =>
			deliver({ screen: "progress", profile: { ui_language: "ru-RU" } }),
		);
		expect(languageOfPage()).toEqual(["ru", "ltr"]);
	});

	test("is the host's when the widget has no words in the one the parent chose", () => {
		const { bridge, deliver } = heldBridge("ru-RU");
		act(() => render(<WidgetApp bridge={bridge} host={idleHost} />, root));

		act(() => deliver({ screen: "profile", profile: { ui_language: "kk" } }));

		expect(languageOfPage()).toEqual(["ru", "ltr"]);
	});

	test("is English when the widget has words for neither", () => {
		const { bridge, deliver } = heldBridge("es-MX");
		act(() => render(<WidgetApp bridge={bridge} host={idleHost} />, root));

		act(() => deliver({ screen: "first_run", profile: null }));

		expect(languageOfPage()).toEqual(["en", "ltr"]);
	});

	test("is named on the page before the card is painted", () => {
		const { bridge, deliver } = heldBridge("ru-RU");
		deliver({ screen: "task" });

		// Drawn outside act, the card has run none of the effects that wait for
		// a paint.
		render(<WidgetApp bridge={bridge} host={idleHost} />, root);

		expect(languageOfPage()).toEqual(["ru", "ltr"]);
	});

	test("follows the host when it names another", () => {
		const { bridge, deliver, changeLocale } = heldBridge("en-US");
		act(() => render(<WidgetApp bridge={bridge} host={idleHost} />, root));
		act(() => deliver({ screen: "task" }));

		act(() => changeLocale("ru-RU"));

		expect(languageOfPage()).toEqual(["ru", "ltr"]);
	});
});
