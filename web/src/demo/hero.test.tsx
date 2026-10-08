import { render } from "preact";
import { act } from "preact/test-utils";
import { afterEach, beforeEach, describe, expect, test, vi } from "vitest";
import { buttonIn, press } from "../widget/testing/card";
import { type DemoData, readDemoData } from "./data";
import { bringHeroAlive, liveScope } from "./hero";
import { checkingTakes } from "./service";
import { openHome } from "./testing/home";

beforeEach(() => {
	vi.useFakeTimers({ toFake: ["setTimeout", "clearTimeout", "Date"] });
	openHome("en");
});

afterEach(() => {
	const live = card();
	if (live !== null) {
		act(() => render(null, live));
	}
	document.body.innerHTML = "";
	vi.restoreAllMocks();
	vi.unstubAllGlobals();
	vi.useRealTimers();
});

// card is the card on the first screen, live or still.
const card = () => document.querySelector<HTMLElement>(".s-hero-card .s-card");

// frame is the body of the chat's frame the card on the first screen stands in.
const frame = () =>
	document.querySelector<HTMLElement>(".s-hero-card .s-frame-body");

// dataOf is the demo's data the page carries.
function dataOf(): DemoData {
	const data = readDemoData(document);
	if (data === undefined) {
		throw new Error("the home page carries no data for its demo");
	}
	return data;
}

// alive brings the card on the first screen alive, and returns the data it
// answers with.
function alive(): DemoData {
	const data = dataOf();
	let brought = false;
	act(() => {
		brought = bringHeroAlive(document, data);
	});
	expect(brought).toBe(true);
	return data;
}

// passing lets the card's clock run on by ms, a step at a time: a card sets
// its next question going once it has drawn the answer to the one before.
async function passing(ms: number): Promise<void> {
	for (let left = ms; left > 0; left -= step) {
		await act(async () => {
			await vi.advanceTimersByTimeAsync(Math.min(step, left));
		});
	}
}

// step is how far the clock runs at a time.
const step = 500;

// option is the option of the live card marked letter.
function option(letter: string): HTMLButtonElement {
	const found = [
		...(card()?.querySelectorAll<HTMLButtonElement>(".mt-option") ?? []),
	].find(
		(row) => row.querySelector(".mt-option-letter")?.textContent === letter,
	);
	if (found === undefined) {
		throw new Error(`the card has no option ${letter}`);
	}
	return found;
}

const states = () =>
	["A", "B", "C", "D", "E"].map((letter) => option(letter).dataset.state);
const text = (selector: string) => card()?.querySelector(selector)?.textContent;
const button = (label: string) => buttonIn(card() as HTMLElement, label);

// shape is every element under a card, in order, as its tag and its classes.
function shape(root: Element | null | undefined): string[] {
	return [...(root?.querySelectorAll("*") ?? [])].map(
		(element) => `${element.tagName.toLowerCase()}.${element.className}`,
	);
}

// ids are the ids every element of the page holds.
const ids = () =>
	[...document.querySelectorAll("[id]")].map((element) => element.id);

describe("the card on the first screen, come alive", () => {
	test("is drawn first as the page was built: the still card's words, made of the same elements, and pressable", () => {
		const still = card()?.querySelector(".mt-widget");
		const words = still?.textContent;
		const made = shape(still);
		expect(card()?.hasAttribute("inert")).toBe(true);

		alive();

		expect(card()?.hasAttribute("inert")).toBe(false);
		expect(card()?.querySelector(".mt-widget")?.textContent).toBe(words);
		expect(shape(card()?.querySelector(".mt-widget"))).toEqual(made);
	});

	// A card follows the room it is given, measured where it stands: a
	// browser gives an element out of the document none of the page's tokens,
	// and a card drawn apart from the page would find no width to measure by.
	test("is drawn in the page, and measures the room it has there, as a card in a chat does", () => {
		const observed: Element[] = [];
		vi.stubGlobal(
			"ResizeObserver",
			class {
				observe(target: Element) {
					observed.push(target);
				}
				disconnect() {}
			},
		);
		vi.spyOn(globalThis, "getComputedStyle").mockImplementation(
			(element) =>
				({
					getPropertyValue: (name: string) =>
						name === "--widget-wide" && element.isConnected ? "640px" : "",
				}) as CSSStyleDeclaration,
		);

		alive();

		expect(observed).toEqual([card()?.querySelector(".mt-widget")]);
	});

	test("checks a wrong option a moment, then tells it with the trap behind it and the solution", async () => {
		const data = alive();

		press(option("B"));
		expect(option("B").dataset.state).toBe("selected");
		await passing(checkingTakes);

		expect(states()).toEqual(["muted", "wrong", "correct", "muted", "muted"]);
		expect(text(".mt-verdict-line")).toBe("Not quite — it's 5, not 4.");
		expect(text(".mt-note-trap p")).toBe(data.results.B.trap?.text);
		expect(card()?.querySelectorAll(".mt-steps li").length).toBeGreaterThan(0);
	});

	test("answers the right option with no trap, and the solution", async () => {
		alive();

		press(option("C"));
		await passing(checkingTakes);

		expect(option("C").dataset.state).toBe("correct");
		expect(card()?.querySelector(".mt-note-trap")).toBeNull();
		expect(card()?.querySelectorAll(".mt-steps li").length).toBeGreaterThan(0);
	});

	test("opens the hint", () => {
		const data = alive();

		press(button("Hint"));

		expect(text(".mt-note-hint p")).toBe(data.handed.task.hint);
	});

	test("asks for another task as a chat does: the parent's message, the card the task is written on, then the lesson's task again", async () => {
		alive();
		press(option("B"));
		await passing(checkingTakes);

		for (const ask of [1, 2]) {
			press(button("Another task"));

			expect(
				document.querySelector(".s-hero-card .s-bubble")?.textContent,
			).toBe("Another task");
			expect(card()?.querySelector(".mt-option")).toBeNull();
			expect(card()?.querySelector(".mt-gen")).not.toBeNull();
			await passing(6000);

			expect(states(), `ask ${ask}`).toEqual([
				"default",
				"default",
				"default",
				"default",
				"default",
			]);
			press(option("D"));
			await passing(checkingTakes);
			expect(option("D").dataset.state).toBe("wrong");
		}
	});

	test("keeps the focus in the chat's frame when the card that asked gives way to the next", async () => {
		alive();
		press(option("B"));
		await passing(checkingTakes);

		press(button("Another task"));

		expect(document.activeElement).toBe(frame());
	});

	test("holds ids of its own, apart from the cards the page was built with, through answers and another task", async () => {
		alive();
		press(option("B"));
		await passing(checkingTakes);
		press(button("Another task"));
		await passing(6000);
		press(option("C"));
		await passing(checkingTakes);

		const held = ids();
		expect(new Set(held).size).toBe(held.length);
		expect(
			[...(card()?.querySelectorAll("[id]") ?? [])].every((element) =>
				element.id.startsWith(liveScope),
			),
		).toBe(true);
	});

	// The page promises to store nothing and send nothing: the card answers
	// from the page's own data.
	test("sends nothing and keeps nothing as it answers", async () => {
		const fetching = vi.spyOn(globalThis, "fetch");
		const storing = vi.spyOn(Storage.prototype, "setItem");
		alive();

		press(button("Hint"));
		press(option("A"));
		await passing(checkingTakes);
		press(button("Another task"));
		await passing(6000);

		expect(fetching).not.toHaveBeenCalled();
		expect(storing).not.toHaveBeenCalled();
	});

	// A page and a script from two builds, one of them still in a cache, may
	// meet: data the card cannot draw leaves the first screen as it was built.
	test("gives the still card its place back when the live one cannot be drawn", () => {
		const data = dataOf();
		const broken = {
			...data,
			handed: {
				...data.handed,
				task: { ...data.handed.task, options: undefined },
			},
		} as unknown as DemoData;
		const still = card();
		vi.spyOn(console, "error").mockImplementation(() => {});

		expect(() =>
			act(() => {
				bringHeroAlive(document, broken);
			}),
		).toThrow();

		expect(card()).toBe(still);
		expect(card()?.hasAttribute("inert")).toBe(true);
	});

	test("leaves a page whose first screen draws no card as it is", () => {
		const data = dataOf();
		document.querySelector(".s-hero-card")?.remove();

		expect(bringHeroAlive(document, data)).toBe(false);
	});
});
