import { render } from "preact";
import { act } from "preact/test-utils";
import { afterEach, beforeEach, describe, expect, test, vi } from "vitest";
import type { Bridge, Call, Host, ToolResult } from "./bridge";
import { foldIn } from "./testing/card";
import {
	coming,
	editSaved,
	exhausted,
	fence,
	firstRun,
	limited,
	profileRead,
	progress,
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

// notDrawnForATask is the call of a host that names no tool: the card waits for
// its result with nothing to show.
const notDrawnForATask: Call = { tool: undefined, stage: "started" };

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
		call: () => notDrawnForATask,
		locale: () => locale,
		subscribe(listener) {
			listeners.add(listener);
			return () => {
				listeners.delete(listener);
			};
		},
		connect: async () => {},
	};
	const deliverResult = (result: ToolResult) => {
		latest = result;
		tell();
	};
	const deliver = (structuredContent: Record<string, unknown>) =>
		deliverResult({ content: [], structuredContent });
	const changeLocale = (next: string) => {
		locale = next;
		tell();
	};
	return { bridge, deliver, deliverResult, changeLocale };
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
		["the progress", standing, ".mt-rank-name"],
		["the profile", profileRead, ".mt-fields"],
		["the first sign-in", firstRun, ".mt-lead"],
	])("draws %s as its card", (_, payload, drawnPart) => {
		const { bridge, deliver } = heldBridge();
		act(() => render(<WidgetApp bridge={bridge} host={idleHost} />, root));

		act(() => deliver(payload));

		expect(root.querySelector(drawnPart)).not.toBeNull();
		expect(root.textContent).not.toContain(unreadable);
	});

	test("keeps a task's card for the task told again, and shows the child as the payload now says", () => {
		const { bridge, deliver } = heldBridge();
		act(() => render(<WidgetApp bridge={bridge} host={idleHost} />, root));
		act(() => deliver(fence));

		act(() =>
			deliver({ ...fence, child: { ...fence.child, pseudonym: "Star" } }),
		);

		expect(root.querySelector(".mt-bar-name")?.textContent).toBe("Star");
	});

	test("keeps a change saved over a task's card until a later payload says otherwise", async () => {
		const { bridge, deliver } = heldBridge();
		const host: Host = {
			callTool: async (name) =>
				name === "read_progress" ? progress : editSaved({ pseudonym: "Star" }),
			sendMessage: async () => {},
			tellModel: async () => {},
		};
		document.body.append(root);
		act(() => render(<WidgetApp bridge={bridge} host={host} />, root));
		act(() => deliver(fence));
		const press = (selector: string) =>
			act(() => root.querySelector<HTMLElement>(selector)?.click());
		const barName = () => root.querySelector(".mt-bar-name")?.textContent;

		press(".mt-bar");
		await vi.waitFor(() =>
			expect(root.querySelector(".mt-fields-head .mt-btn")).not.toBeNull(),
		);
		act(() => foldIn(root, "Profile").click());
		press(".mt-fields-head .mt-btn");
		const box = root.querySelector<HTMLInputElement>(".mt-form .mt-input");
		act(() => {
			if (box !== null) {
				box.value = "Star";
				box.dispatchEvent(new Event("input", { bubbles: true }));
			}
		});
		press(".mt-form .mt-btn-primary");
		await vi.waitFor(() => expect(root.querySelector(".mt-form")).toBeNull());
		press(".mt-bar-back");
		expect(barName()).toBe("Star");

		act(() =>
			deliver({ ...fence, child: { ...fence.child, pseudonym: "Nova" } }),
		);

		expect(barName()).toBe("Nova");
		root.remove();
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
		["a task refused", refused, "This try didn't pass the checks."],
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

	test("draws a task on its way as the card that waits for it, whose it is at its top", () => {
		const { bridge, deliver } = heldBridge();
		act(() => render(<WidgetApp bridge={bridge} host={idleHost} />, root));

		act(() => deliver(coming));

		expect(root.querySelector(".mt-gen")).not.toBeNull();
		expect(root.querySelector(".mt-bar-name")?.textContent).toBe("Comet");
		expect(root.textContent).not.toContain(unreadable);
	});

	test("draws nothing for a result with nothing for a card, as a tool that answered in words alone sent", () => {
		const { bridge, deliverResult } = heldBridge();
		act(() => render(<WidgetApp bridge={bridge} host={idleHost} />, root));

		act(() =>
			deliverResult({
				content: [
					{ type: "text", text: "Request req_old is open. Package: …" },
				],
			}),
		);

		expect(root.innerHTML).toBe("");
	});

	test("says it cannot show a failure that came with nothing for a card", () => {
		const { bridge, deliverResult } = heldBridge();
		act(() => render(<WidgetApp bridge={bridge} host={idleHost} />, root));

		act(() =>
			deliverResult({
				content: [{ type: "text", text: "Something went wrong." }],
				isError: true,
			}),
		);

		expect(root.querySelector(".mt-verdict-line")?.textContent).toBe(
			unreadable,
		);
	});

	test("turns a card waiting after a refused try to the next payload it is told", () => {
		const { bridge, deliver } = heldBridge();
		act(() => render(<WidgetApp bridge={bridge} host={idleHost} />, root));
		act(() => deliver(refused));
		expect(root.textContent).toContain("This try didn't pass the checks.");

		// The last attempt refused turns it to a card no task is coming to.
		act(() => deliver(exhausted));
		expect(root.textContent).toContain("This task didn't work out.");
		expect(root.querySelector(".mt-gen")).toBeNull();
	});

	test.each([
		["the progress", standing],
		["the profile", profileRead],
	])(
		"keeps %s for the same payload told again, and starts it afresh for another",
		(_, payload) => {
			const { bridge, deliver, changeLocale } = heldBridge();
			act(() => render(<WidgetApp bridge={bridge} host={idleHost} />, root));
			act(() => deliver(payload));
			// open says whether the card's form is open, which a card starts without.
			const open = () => root.querySelector(".mt-form") !== null;
			act(() => {
				root.querySelector<HTMLElement>(".mt-fields-head button")?.click();
			});
			expect(open()).toBe(true);

			act(() => changeLocale("en-GB"));
			act(() => deliver({ ...payload }));
			expect(open()).toBe(true);

			act(() => deliver({ ...payload, last_answer: null, told: "again" }));
			expect(open()).toBe(false);
		},
	);

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
