import { render } from "preact";
import { act } from "preact/test-utils";
import { afterEach, describe, expect, test } from "vitest";
import { Icon } from "../design/icons";
import { drawingAlone, drawingOf } from "../design/testing/drawing";
import { cardWords } from "./dictionaries";
import { LessonButtons } from "./LessonFoot";
import { WordsContext } from "./words";

const root = document.createElement("div");

afterEach(() => {
	act(() => render(null, root));
});

// draw draws the buttons under a task not yet answered, in English, with the
// hint open or shut.
function draw(hintOpen: boolean): void {
	act(() =>
		render(
			<WordsContext.Provider value={cardWords("en", undefined)}>
				<LessonButtons
					locked={false}
					anotherSending={false}
					hintOpen={hintOpen}
					onHint={() => {}}
					onAnother={() => {}}
				/>
			</WordsContext.Provider>,
			root,
		),
	);
}

// shown is each button as a person sees it and a screen reader names it: its
// tone, its icon and its words.
const shown = () =>
	[...root.querySelectorAll("button")].map((button) => ({
		className: button.className,
		icon: drawingOf(button.querySelector("svg")),
		words: button.textContent,
	}));

describe("the buttons under a task", () => {
	test("are the hint, a bulb in its amber, and another task, two arrows in a circle in their blue", () => {
		draw(false);

		expect(shown()).toEqual([
			{
				className: "mt-btn mt-btn-icon mt-btn-hint",
				icon: drawingAlone(<Icon name="hint" />),
				words: "Hint",
			},
			{
				className: "mt-btn mt-btn-icon mt-btn-another",
				icon: drawingAlone(<Icon name="renew" />),
				words: "Another task",
			},
		]);
	});

	test("name the hint by what a press does: shows it, or hides it once open", () => {
		draw(true);

		expect(shown()[0]?.words).toBe("Hide hint");
		expect(root.querySelector("button")?.getAttribute("title")).toBe(
			"Hide hint",
		);
	});
});
