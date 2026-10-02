import { render } from "preact";
import { act } from "preact/test-utils";
import { afterEach, beforeEach, describe, expect, test, vi } from "vitest";
import type { Bridge, Host, ToolResult } from "./bridge";
import {
	exhausted,
	fence,
	firstRun,
	firstRunRefused,
	limited,
	profileRead,
	refused,
	standing,
} from "./testing/lesson";
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

// unreadable is what a card says of a payload it cannot draw.
const unreadable =
	"This card can't be shown here. The chat says the same in words.";

// nameShown is the name at the head of the child's card.
const nameShown = () => root.querySelector(".mt-head .mt-name")?.textContent;

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

	test("draws the latest result", () => {
		const { bridge, deliver } = heldBridge();
		act(() => render(<WidgetApp bridge={bridge} host={idleHost} />, root));

		act(() => deliver(standing));
		expect(nameShown()).toBe("Comet");

		act(() =>
			deliver({
				...standing,
				profile: { ...standing.profile, pseudonym: "Otter" },
			}),
		);
		expect(nameShown()).toBe("Otter");
		expect(root.textContent).not.toContain("Comet");
	});

	test("draws a result that arrived before it was drawn", () => {
		const { bridge, deliver } = heldBridge();
		deliver(standing);

		act(() => render(<WidgetApp bridge={bridge} host={idleHost} />, root));

		expect(nameShown()).toBe("Comet");
	});

	test("draws a result that arrived between drawing and listening", async () => {
		const { bridge, deliver } = heldBridge();

		// Drawn outside act, the card has not subscribed yet when the result
		// comes: its effects run after the frame is painted.
		render(<WidgetApp bridge={bridge} host={idleHost} />, root);
		deliver(standing);

		await vi.waitFor(() => expect(nameShown()).toBe("Comet"));
	});

	test.each([
		["the progress", standing, ".mt-rating-num"],
		["the profile", profileRead, ".mt-fields"],
		["the first sign-in", firstRun, ".mt-check"],
	])("draws %s as its card", (_, payload, drawnPart) => {
		const { bridge, deliver } = heldBridge();
		act(() => render(<WidgetApp bridge={bridge} host={idleHost} />, root));

		act(() => deliver(payload));

		expect(root.querySelector(drawnPart)).not.toBeNull();
		expect(root.textContent).not.toContain(unreadable);
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

	test.each([
		["a task refused", refused, "Preparing the next task…"],
		[
			"a day refused",
			limited,
			"There are no more new tasks today\u00a0— there will be more tomorrow.",
		],
	])("draws the wait after %s as the waiting card", (_, payload, says) => {
		const { bridge, deliver } = heldBridge();
		act(() => render(<WidgetApp bridge={bridge} host={idleHost} />, root));

		act(() => deliver(payload));

		expect(root.querySelector(".mt-widget")?.textContent).toContain(says);
		expect(root.querySelector(".stub")).toBeNull();
	});

	test("starts a new wait for each payload that asks for one, and only then", () => {
		vi.useFakeTimers();
		try {
			const { bridge, deliver, changeLocale } = heldBridge();
			act(() => render(<WidgetApp bridge={bridge} host={idleHost} />, root));
			act(() => deliver(refused));
			act(() => {
				vi.advanceTimersByTime(120_000);
			});
			expect(root.querySelector(".mt-gen")).toBeNull();

			// The host telling more, or telling the same payload again, is not a
			// new wait.
			act(() => changeLocale("en-GB"));
			act(() => deliver({ ...refused }));
			expect(root.querySelector(".mt-gen")).toBeNull();

			act(() => deliver(exhausted));
			expect(root.querySelector(".mt-gen")).not.toBeNull();
			expect(root.textContent).toContain("This task didn't work out.");
		} finally {
			vi.useRealTimers();
		}
	});

	test("starts a card afresh for each payload, and keeps it for the same one told again", () => {
		const { bridge, deliver, changeLocale } = heldBridge();
		// A label passes a press on to its box only on the page.
		document.body.append(root);
		act(() => render(<WidgetApp bridge={bridge} host={idleHost} />, root));
		act(() => deliver(firstRun));
		// ticked says whether the card took the adult's tick in: its button is
		// switched on by it.
		const ticked = () =>
			!root.querySelector<HTMLButtonElement>(".mt-btn-primary")?.disabled;
		act(() => {
			root.querySelector<HTMLElement>(".mt-check span")?.click();
		});
		expect(ticked()).toBe(true);

		act(() => changeLocale("en-GB"));
		act(() => deliver({ ...firstRun }));
		expect(ticked()).toBe(true);

		act(() => deliver(firstRunRefused));
		expect(ticked()).toBe(false);
		root.remove();
	});

	test.each([
		["the progress", standing],
		["the profile", profileRead],
	])("starts %s afresh for each payload", async (_, payload) => {
		const { bridge, deliver } = heldBridge();
		vi.spyOn(console, "error").mockImplementation(() => {});
		act(() => render(<WidgetApp bridge={bridge} host={idleHost} />, root));
		act(() => deliver(payload));
		const note = () => root.querySelector(".mt-action-note")?.textContent;
		act(() => {
			root.querySelector<HTMLElement>(".mt-fields-actions button")?.click();
		});
		// The card says the ask did not reach the chat, in the card's language.
		await vi.waitFor(() => expect(note()).not.toBe(""));

		act(() => deliver({ ...payload, last_answer: null, told: "again" }));

		expect(note()).toBe("");
		vi.restoreAllMocks();
	});

	test.each([
		["a wait that does not say whose card it is", { ...refused, child: null }],
		[
			"a task that does not read as one",
			{ ...fence, task: { ...fence.task, options: { A: "3" } } },
		],
		["a result, which no tool draws a card for", { screen: "result" }],
		["a screen nobody names", { text: "<img src=x onerror=alert(1)>" }],
	])("says it cannot show %s, and shows nothing of it", (_, payload) => {
		const { bridge, deliver } = heldBridge();
		act(() => render(<WidgetApp bridge={bridge} host={idleHost} />, root));

		act(() => deliver(payload));

		expect(root.querySelector(".mt-verdict-line")?.textContent).toBe(
			unreadable,
		);
		expect(root.querySelector(".mt-option, .mt-gen, img")).toBeNull();
		expect(root.textContent).not.toContain("attempts_left");
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

	test("is the one the parent chose for the lessons, over the host's", () => {
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

	test("is the lesson's on a task card, over the host's, when the parent chose none", () => {
		const { bridge, deliver } = heldBridge("en-US");
		act(() => render(<WidgetApp bridge={bridge} host={idleHost} />, root));

		act(() =>
			deliver({
				screen: "task",
				child: { pseudonym: "Otter", grade: 2, ui_language: null },
				language: "ru",
			}),
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
		const { bridge, deliver } = heldBridge("sw-KE");
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
