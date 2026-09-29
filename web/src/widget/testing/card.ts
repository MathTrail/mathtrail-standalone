import type { CallToolResult } from "@modelcontextprotocol/client";
import type { McpUiHostContext } from "@modelcontextprotocol/ext-apps";
import { render } from "preact";
import { act } from "preact/test-utils";
import { expect, vi } from "vitest";
import { start } from "../start";
import { deliver, listenAsHost, openTestHost, type ToolCall } from "./host";

/**
 * Drawn is a card a test drew: the element it is drawn in, and what the host
 * heard from it.
 */
export type Drawn = {
	root: HTMLElement;
	heard: ReturnType<typeof listenAsHost>;
};

/**
 * drawCard draws, in an element of its own on the page, the card a host hands
 * payload to, as a chat host does: the widget started over the protocol, the
 * host's context told, the call's arguments and then its result delivered,
 * and the widget's calls answered with tools. It resolves once the card is on
 * the page.
 */
export async function drawCard(
	payload: object,
	{
		tools = () => {
			throw new Error("no tool is answered here");
		},
		refuseMessages = false,
		refuseModelLines = false,
		context = {},
		args = {},
	}: {
		tools?: (call: ToolCall) => CallToolResult | Promise<CallToolResult>;
		refuseMessages?: boolean;
		refuseModelLines?: boolean;
		context?: McpUiHostContext;
		args?: Record<string, unknown>;
	} = {},
): Promise<Drawn> {
	const { host, widgetSide } = await openTestHost(context);
	const heard = listenAsHost(host, tools, { refuseMessages, refuseModelLines });
	const root = document.createElement("div");
	document.body.append(root);
	await start(root, widgetSide);
	await deliver(host, payload, args);
	await vi.waitFor(() =>
		expect(root.querySelector(".mt-widget")).not.toBeNull(),
	);
	return { root, heard };
}

/**
 * takeDown takes a drawn card off the page, and the language it gave the page
 * with it, so that nothing of it is left for the next test.
 */
export function takeDown(root: HTMLElement | undefined): void {
	if (root === undefined) {
		return;
	}
	act(() => render(null, root));
	root.remove();
	document.documentElement.removeAttribute("lang");
	document.documentElement.removeAttribute("dir");
}

/** buttonIn is the button of root a person would find by its words. */
export function buttonIn(root: HTMLElement, label: string): HTMLButtonElement {
	const found = [...root.querySelectorAll<HTMLButtonElement>("button")].find(
		(candidate) =>
			!candidate.closest("[hidden]") &&
			(candidate.textContent === label ||
				candidate.getAttribute("aria-label") === label),
	);
	if (found === undefined) {
		throw new Error(`the card has no button ${label}`);
	}
	return found;
}

/** press gives element the focus and presses it, as a person does. */
export function press(element: HTMLElement): void {
	act(() => {
		element.focus();
		element.click();
	});
}
