import { render } from "preact";
import { useState } from "preact/hooks";
import { act } from "preact/test-utils";
import { afterEach, describe, expect, test } from "vitest";
import {
	Fold,
	GeneratingSteps,
	Note,
	type NoteTone,
	SolutionSteps,
	Statement,
	Verdict,
} from "./blocks";
import { Icon } from "./icons";
import { drawnAlone } from "./testing/drawing";

const root = document.createElement("div");

afterEach(() => {
	act(() => render(null, root));
});

function draw(element: preact.JSX.Element): void {
	act(() => render(element, root));
}

// Folding is a part folded away under its title, opened and folded by its
// title as whoever draws a fold does: remembering which it is.
function Folding({ summary }: { summary?: string }) {
	const [open, setOpen] = useState(false);
	return (
		<Fold
			title="Topics"
			summary={summary}
			open={open}
			onToggle={() => setOpen(!open)}
		>
			<p>What it holds</p>
		</Fold>
	);
}

// title and body are the title and the body of the fold drawn.
const title = () => root.querySelector<HTMLButtonElement>("h2 button");
const body = () => root.querySelector(".mt-fold-body");

describe("a fold", () => {
	test("is a button under a title, saying that the part it leads to is folded", () => {
		draw(<Folding summary="3 up" />);

		expect(title()?.getAttribute("type")).toBe("button");
		expect(title()?.getAttribute("aria-expanded")).toBe("false");
		expect(title()?.getAttribute("aria-controls")).toBe(body()?.id);
		expect(body()?.id).not.toBe("");
	});

	test("keeps the part on the page while it is folded, hidden", () => {
		draw(<Folding />);

		expect(body()?.hasAttribute("hidden")).toBe(true);
		expect(body()?.textContent).toBe("What it holds");
	});

	test("opens the part when pressed, and folds it again", () => {
		draw(<Folding />);

		act(() => title()?.click());

		expect(title()?.getAttribute("aria-expanded")).toBe("true");
		expect(body()?.hasAttribute("hidden")).toBe(false);

		act(() => title()?.click());

		expect(title()?.getAttribute("aria-expanded")).toBe("false");
		expect(body()?.hasAttribute("hidden")).toBe(true);
	});

	test("says its title and its summary as two things, with the arrow beside them", () => {
		draw(<Folding summary="3 up" />);

		expect(title()?.textContent).toBe("Topics 3 up");
		expect(title()?.querySelector(".mt-fold-summary")?.textContent).toBe(
			"3 up",
		);
		expect(title()?.querySelector(".mt-chevron")).not.toBeNull();
	});

	test("with no summary says its title alone", () => {
		draw(<Folding />);

		expect(title()?.textContent).toBe("Topics");
		expect(title()?.querySelector(".mt-fold-summary")).toBeNull();
	});
});

// Stating is a task's wording under its label, folded and opened by its
// button as whoever draws it does: remembering which it is, open at first.
function Stating() {
	const [open, setOpen] = useState(true);
	return (
		<Statement
			label="Task"
			fold="Collapse"
			unfold="Expand"
			open={open}
			onToggle={() => setOpen(!open)}
		>
			<p>How many posts are there?</p>
		</Statement>
	);
}

// label, toggle and wording are the label, the button and the wording of the
// statement drawn.
const label = () => root.querySelector("h3.mt-statement-label");
const toggle = () =>
	root.querySelector<HTMLButtonElement>("button.mt-statement-toggle");
const wording = () => root.querySelector(".mt-statement-body");

describe("a statement", () => {
	test("is the wording under its label, with a button that says it folds it away", () => {
		draw(<Stating />);

		expect(label()?.textContent).toBe("Task");
		expect(label()?.querySelector("svg")).not.toBeNull();
		expect(wording()?.textContent).toBe("How many posts are there?");
		expect(wording()?.hasAttribute("hidden")).toBe(false);
		expect(toggle()?.getAttribute("type")).toBe("button");
		expect(toggle()?.textContent).toBe("Collapse");
		expect(toggle()?.getAttribute("aria-expanded")).toBe("true");
		expect(toggle()?.getAttribute("aria-controls")).toBe(wording()?.id);
		expect(wording()?.id).not.toBe("");
	});

	test("folds the wording away when pressed, keeping it on the page, and opens it again", () => {
		draw(<Stating />);

		act(() => toggle()?.click());

		expect(toggle()?.textContent).toBe("Expand");
		expect(toggle()?.getAttribute("aria-expanded")).toBe("false");
		expect(wording()?.hasAttribute("hidden")).toBe(true);
		expect(wording()?.textContent).toBe("How many posts are there?");
		expect(label()?.textContent).toBe("Task");

		act(() => toggle()?.click());

		expect(toggle()?.textContent).toBe("Collapse");
		expect(toggle()?.getAttribute("aria-expanded")).toBe("true");
		expect(wording()?.hasAttribute("hidden")).toBe(false);
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
	test.each(["correct", "wrong"] as const)(
		"that is %s carries its mark",
		(tone) => {
			draw(<Verdict tone={tone}>Words</Verdict>);

			expect(root.querySelector(".mt-verdict-line svg")?.outerHTML).toBe(
				drawnAlone(<Icon name={`verdict-${tone}`} size={20} />),
			);
			expect(root.querySelector(".mt-verdict-line")?.textContent).toBe("Words");
		},
	);

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
		expect(steps[0]?.querySelector(".mt-gen-icon svg")?.outerHTML).toBe(
			drawnAlone(<Icon name="step-done" size={20} />),
		);
		expect(
			steps[1]
				?.querySelector(".mt-gen-turning svg")
				?.classList.contains("mt-spin"),
		).toBe(true);
		expect(steps[2]?.querySelector(".mt-gen-icon svg")?.outerHTML).toBe(
			drawnAlone(<Icon name="step-waiting" size={20} />),
		);
	});

	test("is heard as it moves", () => {
		draw(<GeneratingSteps title="Title" steps={[]} statusLabels={labels} />);

		expect(root.querySelector(".mt-gen ol")?.getAttribute("aria-live")).toBe(
			"polite",
		);
	});
});
