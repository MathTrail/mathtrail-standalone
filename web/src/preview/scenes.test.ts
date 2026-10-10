import { act } from "preact/test-utils";
import { afterEach, describe, expect, test, vi } from "vitest";
import { kinds } from "../design/picture/model";
import { extremes } from "../design/picture/testing/extremes";
import { cardWords } from "../widget/dictionaries";
import { readHandedTask } from "../widget/payload";
import { type Drawn, drawCard, takeDown } from "../widget/testing/card";
import type { Key } from "../widget/words";
import { scenesIn } from "./scenes";

let drawn: Drawn | undefined;

afterEach(() => {
	takeDown(drawn?.root);
	drawn = undefined;
});

describe("the preview's progress with one section open", () => {
	// The progress opens with its topics. Each of these scenes leaves one
	// section open, a card short enough to take in at once: the review in their
	// place, or the topics, as the card first opens.
	test.each<[string, Key]>([
		["progress, the review open", "review.title"],
		["progress, the topics open", "progress.topics"],
	])("%s leaves one section open, %s", async (name, section) => {
		const scene = scenesIn("en").find((found) => found.name === name);
		if (scene?.payload === undefined) {
			throw new Error(`the preview has no scene ${name}`);
		}
		drawn = await drawCard(scene.payload);
		act(() => scene.play?.(document));
		const open = [
			...drawn.root.querySelectorAll('.mt-fold-button[aria-expanded="true"]'),
		].map((title) => title.querySelector(".mt-fold-title")?.textContent);
		expect(open).toEqual([cardWords("en", undefined).text(section)]);
	});
});

describe("the preview's card answered", () => {
	// The scene presses an option: the card turns into how the answer went, in
	// the task's place, and asks the chat for nothing. The answer comes back
	// slowly here, as on a busy machine, and the scene waits for it.
	test("ends on the card turned into how the answer went, even when the answer comes slowly", async () => {
		const scene = scenesIn("en").find(
			(found) => found.name === "answered, the card turned into how it went",
		);
		if (scene?.payload === undefined || scene.answers === undefined) {
			throw new Error("the preview has no card answered");
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
			() => expect(root.querySelector(".mt-verdict-line")).not.toBeNull(),
			{ timeout: 3000 },
		);
		expect(root.querySelector(".mt-option")).toBeNull();
		expect(heard.messages).toEqual([]);
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

// The preview is where the layout of every picture is measured: a kind with no
// scene of its own would be drawn somewhere no measure reaches.
describe("the preview's pictures", () => {
	const scenes = scenesIn("en");
	const pictureOf = (name: string) =>
		readHandedTask(scenes.find((scene) => scene.name === name)?.payload)?.task
			.picture;

	test("show every kind as the content shows the model, in the order of the kinds", () => {
		const shown = scenes
			.filter((scene) => scene.name.startsWith("picture: "))
			.map((scene) => scene.name);

		expect(shown).toEqual(kinds.map((kind) => `picture: ${kind}`));
		for (const kind of kinds) {
			expect(pictureOf(`picture: ${kind}`)?.kind).toBe(kind);
		}
	});

	test("show every kind at its limits, as the card reads it", () => {
		for (const extreme of extremes) {
			expect(pictureOf(`picture at its limits: ${extreme.name}`)).toEqual(
				extreme.picture,
			);
		}
	});
});
