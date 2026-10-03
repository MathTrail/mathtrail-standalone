import { afterEach, describe, expect, test } from "vitest";
import { type Drawn, drawCard, takeDown } from "./testing/card";
import { firstRun, firstRunRefused } from "./testing/lesson";

let drawn: Drawn | undefined;

afterEach(() => {
	takeDown(drawn?.root);
	drawn = undefined;
});

async function draw(payload: object): Promise<Drawn> {
	drawn = await drawCard(payload);
	return drawn;
}

describe("the first sign-in", () => {
	test("tells the adult what MathTrail is, the pseudonym's rule, what is asked, where it is kept and how to start", async () => {
		const { root } = await draw(firstRun);

		expect(root.querySelector("article")?.getAttribute("aria-label")).toBe(
			"Welcome to MathTrail",
		);
		expect(root.querySelector(".mt-badge")).toBeNull();
		expect(root.querySelector(".mt-lead")?.textContent).toContain(
			"An adult sets it up: a parent or a tutor.",
		);
		expect(
			[...root.querySelectorAll(".mt-note-label")].map(
				(label) => label.textContent,
			),
		).toEqual([
			"A pseudonym",
			"What the chat will ask",
			"Where the data is kept",
			"How to start",
		]);
		expect(root.textContent).toContain(
			"Ask in this chat for a MathTrail profile, and answer what it asks.",
		);
		expect(root.textContent).toContain("the language of the lessons");
		expect(root.querySelector(".mt-verdict-line")).toBeNull();
	});

	// The profile is made in the chat, where the adult says they are the child's
	// parent or tutor: the card asks nothing, sends nothing and calls nothing.
	test("only tells: it has nothing to tick and nothing to press", async () => {
		const { root, heard } = await draw(firstRun);

		expect(root.querySelector("input, button, select, form")).toBeNull();
		expect(heard.messages).toEqual([]);
		expect(heard.calls).toEqual([]);
	});

	test("in a card of an earlier chat, after a first profile refused, says it was not made", async () => {
		const { root } = await draw(firstRunRefused);

		expect(root.querySelector(".mt-verdict-line")?.textContent).toBe(
			"The profile wasn't created.",
		);
		expect(root.textContent).not.toContain("9 given");
	});
});
