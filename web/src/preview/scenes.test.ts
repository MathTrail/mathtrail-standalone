import { act } from "preact/test-utils";
import { afterEach, describe, expect, test, vi } from "vitest";
import { cardWords } from "../widget/dictionaries";
import { type Drawn, drawCard, takeDown } from "../widget/testing/card";
import { scenesIn } from "./scenes";

let drawn: Drawn | undefined;

afterEach(() => {
	takeDown(drawn?.root);
	drawn = undefined;
});

describe("the preview's progress with its review open", () => {
	// The progress opens with its topics; this scene shows its review in their
	// place, a card short enough to take in at once.
	test("leaves the review the one section open", async () => {
		const scene = scenesIn("en").find(
			(found) => found.name === "progress, the review open",
		);
		if (scene?.payload === undefined) {
			throw new Error("the preview has no progress with its review open");
		}
		drawn = await drawCard(scene.payload);
		act(() => scene.play?.(document));
		const open = [
			...drawn.root.querySelectorAll('.mt-fold-button[aria-expanded="true"]'),
		].map((title) => title.querySelector(".mt-fold-title")?.textContent);
		expect(open).toEqual([cardWords("en", undefined).text("review.title")]);
	});
});

describe("the preview's card asked for another task once the answer is in", () => {
	// The scene presses an option and then the next task's button, which is
	// there only once the answer is in: it shows the card done with, its
	// result kept and its buttons gone. The answer comes back slowly here, as
	// on a busy machine, and the scene waits for it.
	test("ends on the card done with, its result above what it says, even when the answer comes slowly", async () => {
		const scene = scenesIn("en").find(
			(found) =>
				found.name ===
				"another task asked once the answer is in, the card done with",
		);
		if (scene?.payload === undefined || scene.answers === undefined) {
			throw new Error("the preview has no card asked for another task");
		}
		const answers = scene.answers;
		drawn = await drawCard(scene.payload, {
			tools: async (call) => {
				await new Promise((later) => setTimeout(later, 400));
				return answers(call.name, 0);
			},
		});
		const { root, heard } = drawn;
		await vi.waitFor(() =>
			expect(root.querySelector(".mt-option")).not.toBeNull(),
		);
		act(() => scene.play?.(document));

		await vi.waitFor(
			() =>
				expect(root.querySelector(".mt-action-note")?.textContent).toBe(
					cardWords("en", undefined).text("task.another_coming"),
				),
			{ timeout: 3000 },
		);
		expect(heard.messages).toEqual([
			cardWords("en", undefined).text("task.another"),
		]);
		expect(root.querySelector(".mt-verdict-line")).not.toBeNull();
		expect(root.querySelector(".mt-btns")).toBeNull();
	});
});

describe("the preview's topic chosen with the choice shut", () => {
	// The button names the topic chosen; the scene chooses the topic whose name
	// is the longest in the card's language, which is not the same topic in
	// every language, so that the measure of the layout sees each language's
	// longest.
	test.each([
		["en", "games.strategy"],
		["de", "parity.alternation"],
		["fr", "logic.sets"],
	])("in %s chooses %s", (language, topic) => {
		const scene = scenesIn(language).find(
			(found) => found.name === "topic chosen, the choice shut",
		);

		expect(scene?.payload).toMatchObject({ topic_choice: { chosen: topic } });
	});
});
