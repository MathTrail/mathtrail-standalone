import { act } from "preact/test-utils";
import { afterEach, describe, expect, test, vi } from "vitest";
import {
	buttonIn,
	type Drawn,
	drawCard,
	press,
	takeDown,
} from "./testing/card";
import { firstRun, firstRunRefused } from "./testing/lesson";

let drawn: Drawn | undefined;

afterEach(() => {
	takeDown(drawn?.root);
	drawn = undefined;
	vi.restoreAllMocks();
});

async function draw(
	payload: object,
	options: Parameters<typeof drawCard>[1] = {},
): Promise<Drawn> {
	drawn = await drawCard(payload, options);
	return drawn;
}

// tick ticks the adult's statement, as a person does, on its words.
function tick(root: HTMLElement): void {
	act(() => {
		root.querySelector<HTMLElement>(".mt-check span")?.click();
	});
}

describe("the first sign-in", () => {
	test("tells the adult what MathTrail is, the pseudonym's rule, what is asked and where it is kept", async () => {
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
		]);
		expect(root.querySelector(".mt-verdict-line")).toBeNull();
	});

	test("asks for the profile only once the adult has said who they are", async () => {
		const { root, heard } = await draw(firstRun);
		const create = buttonIn(root, "Create a profile");

		expect(create.disabled).toBe(true);
		press(create);
		expect(heard.messages).toEqual([]);

		tick(root);
		expect(
			root.querySelector<HTMLInputElement>(".mt-check input")?.checked,
		).toBe(true);
		expect(create.disabled).toBe(false);
		press(create);

		await vi.waitFor(() =>
			expect(heard.messages).toEqual(["Create a profile"]),
		);
		await vi.waitFor(() =>
			expect(root.querySelector(".mt-action-note")?.textContent).toBe(
				"Sent to the chat",
			),
		);
		expect(heard.calls).toEqual([]);
		// Once the chat has the ask, the button asks no more.
		expect(create.getAttribute("aria-disabled")).toBe("true");
		press(create);
		await Promise.resolve();
		expect(heard.messages).toEqual(["Create a profile"]);
	});

	test("that the chat does not take says so", async () => {
		vi.spyOn(console, "error").mockImplementation(() => {});
		const { root } = await draw(firstRun, { refuseMessages: true });
		tick(root);

		press(buttonIn(root, "Create a profile"));

		await vi.waitFor(() =>
			expect(root.querySelector(".mt-action-note")?.textContent).toBe(
				"Not sent — try again",
			),
		);
	});

	test("after a first profile refused, says it was not made", async () => {
		const { root } = await draw(firstRunRefused);

		expect(root.querySelector(".mt-verdict-line")?.textContent).toBe(
			"The profile wasn't created.",
		);
		expect(root.textContent).not.toContain("9 given");
	});
});
