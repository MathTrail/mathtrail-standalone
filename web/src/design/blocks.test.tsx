import { render } from "preact";
import { act } from "preact/test-utils";
import { afterEach, describe, expect, test } from "vitest";
import { Diagram, Note, type NoteTone, SolutionSteps, Verdict } from "./blocks";

const root = document.createElement("div");

afterEach(() => {
	act(() => render(null, root));
});

function draw(element: preact.JSX.Element): void {
	act(() => render(element, root));
}

describe("a drawing", () => {
	test("is laid out left to right, named for a screen reader, and reachable to scroll", () => {
		draw(<Diagram drawing={"A---B\n|   |\n"} label="Drawing" />);

		const drawing = root.querySelector("pre.mt-diagram");
		expect(drawing?.getAttribute("dir")).toBe("ltr");
		expect(drawing?.getAttribute("role")).toBe("img");
		expect(drawing?.getAttribute("aria-label")).toBe("Drawing");
		expect(drawing?.getAttribute("tabindex")).toBe("0");
		expect(drawing?.textContent).toBe("A---B\n|   |");
	});

	test("keeps markup in it as text", () => {
		draw(<Diagram drawing="<b>A</b>" label="Drawing" />);

		expect(root.querySelector("b")).toBeNull();
		expect(root.querySelector("pre")?.textContent).toBe("<b>A</b>");
	});
});

describe("a note", () => {
	test.each<[NoteTone, string, boolean]>([
		["plain", "DIV", false],
		["hint", "ASIDE", true],
		["trap", "DIV", true],
	])("in the tone %s is a %s, with an icon: %s", (tone, element, icon) => {
		draw(
			<Note tone={tone} label="Label">
				Words
			</Note>,
		);

		const note = root.querySelector(".mt-note");
		expect(note?.tagName).toBe(element);
		expect(note?.classList.contains(`mt-note-${tone}`)).toBe(true);
		expect(note?.querySelector(".mt-note-label span")?.textContent).toBe(
			"Label",
		);
		expect(note?.querySelector("p")?.textContent).toBe("Words");
		expect(note?.querySelector(".mt-note-label svg") !== null).toBe(icon);
	});
});

test("a note and the steps give their words in the language they are said in", () => {
	draw(
		<>
			<Note tone="hint" label="Hint" said={{ lang: "ru", dir: "ltr" }}>
				Начни с забора поменьше.
			</Note>
			<SolutionSteps
				label="Solution"
				steps={["12 : 3 = 4."]}
				said={{ lang: "ru", dir: "ltr" }}
			/>
		</>,
	);

	expect(root.querySelector(".mt-note p")?.getAttribute("lang")).toBe("ru");
	expect(root.querySelector(".mt-note-label")?.hasAttribute("lang")).toBe(
		false,
	);
	expect(root.querySelector(".mt-steps ol")?.getAttribute("lang")).toBe("ru");
});

describe("a verdict", () => {
	test.each([
		["correct", "var(--correct)"],
		["wrong", "var(--wrong)"],
	] as const)("that is %s carries its mark", (tone, colour) => {
		draw(<Verdict tone={tone}>Words</Verdict>);

		expect(
			root.querySelector(".mt-verdict-line svg circle")?.getAttribute("fill"),
		).toBe(colour);
		expect(root.querySelector(".mt-verdict-line")?.textContent).toBe("Words");
	});

	test("with no verdict carries no mark", () => {
		draw(<Verdict>Here's how to solve it.</Verdict>);

		expect(root.querySelector(".mt-verdict-line svg")).toBeNull();
	});
});

describe("the solution", () => {
	test("numbers its steps from one, under its label", () => {
		draw(
			<SolutionSteps
				label="Solution"
				steps={["12 ÷ 3 = 4 gaps.", "4 + 1 = 5 posts.", "4 + 1 = 5 posts."]}
			/>,
		);

		expect(root.querySelector(".mt-section-label")?.textContent).toBe(
			"Solution",
		);
		expect(
			[...root.querySelectorAll("ol li")].map((step) => [
				step.querySelector(".mt-step-num")?.textContent,
				step.lastElementChild?.textContent,
			]),
		).toEqual([
			["1", "12 ÷ 3 = 4 gaps."],
			["2", "4 + 1 = 5 posts."],
			["3", "4 + 1 = 5 posts."],
		]);
	});
});
