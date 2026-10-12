import { afterEach, describe, expect, test, vi } from "vitest";
import { type Drawn, drawCard, press, takeDown } from "./testing/card";
import { fence, staleAnswer, withTopicChoice } from "./testing/lesson";

let drawn: Drawn | undefined;

afterEach(() => {
	takeDown(drawn?.root);
	drawn = undefined;
	vi.restoreAllMocks();
});

// probing is the fence handed out as a probe: a task the coach set in a topic
// on pause, whose easiest tasks are too hard yet.
const probing = { ...fence, task: { ...fence.task, probe: true } };

// draw draws the card of payload, as it first opens.
async function draw(
	payload: object,
	options: Parameters<typeof drawCard>[1] = {},
): Promise<HTMLElement> {
	drawn = await drawCard(payload, options);
	return drawn.root;
}

// option is the row of the option of this letter.
function option(root: HTMLElement, letter: string): HTMLButtonElement {
	const found = [
		...root.querySelectorAll<HTMLButtonElement>(".mt-option"),
	].find(
		(row) => row.querySelector(".mt-option-letter")?.textContent === letter,
	);
	if (found === undefined) {
		throw new Error(`the card has no option ${letter}`);
	}
	return found;
}

describe("the mark of a probe", () => {
	test("stands above a probe and says it checks the topic", async () => {
		const root = await draw(probing);

		const mark = root.querySelector(".mt-body > .mt-probe-chip");
		expect(mark?.textContent).toBe("Topic check");
		expect(mark?.nextElementSibling?.className).toBe("mt-task-text");
	});

	test.each([
		["any other task", fence],
		[
			"a task whose mark does not read",
			{ ...fence, task: { ...fence.task, probe: "yes" } },
		],
	])("is not above %s", async (_, payload) => {
		const root = await draw(payload);

		expect(root.querySelector(".mt-probe-chip")).toBeNull();
	});

	test("is not above a task that is closed", async () => {
		const root = await draw(probing, { tools: () => staleAnswer });
		expect(root.querySelector(".mt-probe-chip")).not.toBeNull();

		press(option(root, "B"));

		await vi.waitFor(() =>
			expect(root.querySelector(".mt-probe-chip")).toBeNull(),
		);
	});

	// A probe is the coach's choice and never comes on a topic a person keeps
	// the lessons to; a card that is told both shows the person's choice.
	test("gives way to the mark of the topic chosen", async () => {
		const root = await draw(
			withTopicChoice(probing, { chosen: probing.task.topic }),
		);

		expect(root.querySelector(".mt-topic-chip")).not.toBeNull();
		expect(root.querySelector(".mt-probe-chip")).toBeNull();
	});
});
