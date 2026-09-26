import { render } from "preact";
import { act } from "preact/test-utils";
import { afterEach, beforeEach, describe, expect, test, vi } from "vitest";
import type { Bridge, ToolResult } from "./bridge";
import { WidgetApp } from "./WidgetApp";

// heldBridge is a bridge whose results arrive when the test says, so the card
// can be caught before, between and after them.
function heldBridge() {
	let latest: ToolResult | undefined;
	const listeners = new Set<() => void>();
	const bridge: Bridge = {
		result: () => latest,
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
		for (const listener of listeners) {
			listener();
		}
	};
	return { bridge, deliver };
}

let root: HTMLElement;

beforeEach(() => {
	root = document.createElement("div");
});

// A card drawn by one test is taken down after it, with its subscription, so
// that nothing of it is still listening while the next one runs.
afterEach(() => {
	act(() => render(null, root));
});

describe("the card", () => {
	test("draws nothing before the first result", () => {
		const { bridge } = heldBridge();

		act(() => render(<WidgetApp bridge={bridge} />, root));

		expect(root.innerHTML).toBe("");
	});

	test("shows the payload of the latest result", () => {
		const { bridge, deliver } = heldBridge();
		act(() => render(<WidgetApp bridge={bridge} />, root));

		act(() => deliver({ screen: "task", attempt: 1 }));
		expect(root.textContent).toContain('"attempt": 1');

		act(() => deliver({ screen: "task", attempt: 2 }));
		expect(root.textContent).toContain('"attempt": 2');
		expect(root.textContent).not.toContain('"attempt": 1');
	});

	test("shows a result that arrived before it was drawn", () => {
		const { bridge, deliver } = heldBridge();
		deliver({ screen: "progress" });

		act(() => render(<WidgetApp bridge={bridge} />, root));

		expect(root.textContent).toContain('"screen": "progress"');
	});

	test("shows a result that arrived between drawing and listening", async () => {
		const { bridge, deliver } = heldBridge();

		// Drawn outside act, the card has not subscribed yet when the result
		// comes: its effects run after the frame is painted.
		render(<WidgetApp bridge={bridge} />, root);
		deliver({ screen: "result" });

		await vi.waitFor(() =>
			expect(root.textContent).toContain('"screen": "result"'),
		);
	});

	test("keeps markup inside a payload as text", () => {
		const { bridge, deliver } = heldBridge();
		act(() => render(<WidgetApp bridge={bridge} />, root));

		act(() => deliver({ text: "<img src=x onerror=alert(1)>" }));

		expect(root.querySelector("img")).toBeNull();
		expect(root.textContent).toContain("<img src=x onerror=alert(1)>");
	});
});
