import { act } from "preact/test-utils";
import { afterEach, describe, expect, test } from "vitest";
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
