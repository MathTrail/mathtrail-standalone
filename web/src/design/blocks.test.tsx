import { render } from "preact";
import { act } from "preact/test-utils";
import { afterEach, describe, expect, test } from "vitest";
import {
	Diagram,
	GeneratingSteps,
	Note,
	type NoteTone,
	SolutionSteps,
	Verdict,
} from "./blocks";

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
		expect(note?.querySelector(".mt-note-detail")).toBeNull();
	});

	test("says its detail under its words, in the card's language rather than theirs", () => {
		draw(
			<Note
				tone="trap"
				label="The trap"
				said={{ lang: "ru", dir: "ltr" }}
				detail="This mistake has come up before."
			>
				Посчитаны промежутки, а не столбы.
			</Note>,
		);

		const said = [...root.querySelectorAll(".mt-note p")];
		expect(said.map((line) => line.textContent)).toEqual([
			"Посчитаны промежутки, а не столбы.",
			"This mistake has come up before.",
		]);
		expect(said.map((line) => line.getAttribute("lang"))).toEqual(["ru", null]);
		expect(said[1]?.classList.contains("mt-note-detail")).toBe(true);
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
		expect(root.querySelector(".mt-verdict")).toBeNull();
	});

	test("with more to say says it under its line", () => {
		draw(
			<Verdict detail="Ask for one in the chat.">No task is coming.</Verdict>,
		);

		expect(
			[...(root.querySelector(".mt-verdict")?.children ?? [])].map((line) => [
				line.className,
				line.textContent,
			]),
		).toEqual([
			["mt-verdict-line", "No task is coming."],
			["mt-verdict-detail", "Ask for one in the chat."],
		]);
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

describe("the course being followed", () => {
	const labels = { done: "Done:", active: "In progress:", waiting: "Waiting:" };

	test("marks each step as it stands, and tells a screen reader how", () => {
		draw(
			<GeneratingSteps
				title="Preparing the next task…"
				steps={[
					{ label: "Picked topic and difficulty", status: "done" },
					{ label: "Writing the task", status: "active" },
					{ label: "Ready", status: "waiting" },
				]}
				statusLabels={labels}
			/>,
		);

		expect(root.querySelector(".mt-gen-title")?.textContent).toBe(
			"Preparing the next task…",
		);
		const steps = [...root.querySelectorAll(".mt-gen li")];
		expect(
			steps.map((step) => [
				step.getAttribute("data-status"),
				step.querySelector(".mt-vh")?.textContent,
				step.textContent,
			]),
		).toEqual([
			["done", "Done: ", "Done: Picked topic and difficulty"],
			["active", "In progress: ", "In progress: Writing the task"],
			["waiting", "Waiting: ", "Waiting: Ready"],
		]);
		expect(
			steps[0]?.querySelector(".mt-gen-icon circle")?.getAttribute("fill"),
		).toBe("var(--ink-strong)");
		expect(
			steps[1]
				?.querySelector(".mt-gen-turning svg")
				?.classList.contains("mt-spin"),
		).toBe(true);
		expect(
			steps[2]?.querySelector(".mt-gen-icon circle")?.getAttribute("r"),
		).toBe("8.5");
	});

	test("is heard as it moves", () => {
		draw(<GeneratingSteps title="Title" steps={[]} statusLabels={labels} />);

		expect(root.querySelector(".mt-gen ol")?.getAttribute("aria-live")).toBe(
			"polite",
		);
	});
});
