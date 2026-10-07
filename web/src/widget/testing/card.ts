import type { CallToolResult } from "@modelcontextprotocol/client";
import type { McpUiHostContext } from "@modelcontextprotocol/ext-apps";
import type { AppBridge } from "@modelcontextprotocol/ext-apps/app-bridge";
import { render } from "preact";
import { act } from "preact/test-utils";
import { expect, vi } from "vitest";
import { start } from "../start";
import {
	deliver,
	listenAsHost,
	type Opening,
	openTestHost,
	type ToolCall,
} from "./host";

/**
 * CardOptions are what a test sets of the host a card is drawn on. A host
 * given links tells the card it opens pages, and answers each as links says;
 * one given none says nothing of pages. A host that refuses messages refuses
 * every one, or, given a number, that many of the first.
 */
type CardOptions = {
	tools?: (call: ToolCall) => CallToolResult | Promise<CallToolResult>;
	refuseMessages?: boolean | number;
	refuseModelLines?: boolean;
	context?: McpUiHostContext;
	links?: Opening;
};

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
		args = {},
		...options
	}: CardOptions & { args?: Record<string, unknown> } = {},
): Promise<Drawn> {
	const { host, root, heard } = await openCard(options);
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

/**
 * foldIn is the title of root's section named title: the button that opens
 * and folds it. The button's words are the title and the summary beside it,
 * so it is found by the title alone.
 */
export function foldIn(root: HTMLElement, title: string): HTMLButtonElement {
	const found = [
		...root.querySelectorAll<HTMLButtonElement>("button.mt-fold-button"),
	].find(
		(candidate) =>
			!candidate.closest("[hidden]") &&
			candidate.querySelector(".mt-fold-title")?.textContent === title,
	);
	if (found === undefined) {
		throw new Error(`the card has no section ${title}`);
	}
	return found;
}

/** unfold opens root's section named title as a person does, if it is folded. */
export function unfold(root: HTMLElement, title: string): void {
	const section = foldIn(root, title);
	if (section.getAttribute("aria-expanded") !== "true") {
		press(section);
	}
}

/**
 * openCard starts the widget, in an element of its own on the page, on a host
 * that has told it its context and nothing of the call yet, and returns the
 * host for the test to tell it the rest — the call's arguments, its result or
 * its cancelling — as a chat host would, when it would.
 */
export async function openCard({
	tools = () => {
		throw new Error("no tool is answered here");
	},
	refuseMessages = false,
	refuseModelLines = false,
	context = {},
	links,
}: CardOptions = {}): Promise<Drawn & { host: AppBridge }> {
	const { host, widgetSide } = await openTestHost(
		context,
		links === undefined ? {} : { openLinks: {} },
	);
	const heard = listenAsHost(host, tools, {
		refuseMessages,
		refuseModelLines,
		opening: links,
	});
	const root = document.createElement("div");
	document.body.append(root);
	await start(root, widgetSide);
	return { root, heard, host };
}
