import { render } from "preact";
import { act } from "preact/test-utils";
import { afterEach, describe, expect, test } from "vitest";
import { AdviceSteps, JudgedList, NamedLine, ReviewPart } from "./review";

const root = document.createElement("div");

afterEach(() => {
	act(() => render(null, root));
});

function draw(element: preact.JSX.Element): void {
	act(() => render(element, root));
}

describe("a part of a review", () => {
	test("is headed in its tone, a heading under the title of the part it stands in", () => {
		draw(
			<ReviewPart tone="accent" label="What to do next">
				<p>the steps</p>
			</ReviewPart>,
		);

		const heading = root.querySelector("section.mt-review-part > h3");
		expect(heading?.textContent).toBe("What to do next");
		expect(heading?.getAttribute("data-tone")).toBe("accent");
		expect(root.querySelector("section.mt-review-part > p")?.textContent).toBe(
			"the steps",
		);
	});
});

describe("the topics a review names", () => {
	test("are each after a dot in their side's colour, hidden from a screen reader, with the name and why under it", () => {
		draw(
			<JudgedList
				tone="wrong"
				rows={[
					{
						id: "a",
						name: "Enumeration",
						line: "A mistake keeps coming back.",
					},
					{ id: "b", name: "Clocks" },
				]}
			/>,
		);

		const rows = [...root.querySelectorAll("ul.mt-judged > li")];
		expect(
			rows.map((row) => [
				row.querySelector(".mt-judged-name")?.textContent,
				row.querySelector(".mt-review-note")?.textContent ?? null,
			]),
		).toEqual([
			["Enumeration", "A mistake keeps coming back."],
			["Clocks", null],
		]);
		expect(
			rows.map((row) => {
				const dot = row.querySelector(".mt-judged-dot");
				return [
					dot?.getAttribute("data-tone"),
					dot?.getAttribute("aria-hidden"),
				];
			}),
		).toEqual([
			["wrong", "true"],
			["wrong", "true"],
		]);
	});
});

describe("a line of names", () => {
	test("says the names, and under them what they have in common", () => {
		draw(<NamedLine names="Clocks, Calendar" note="Too few answers so far." />);

		expect(root.querySelector(".mt-judged-name")?.textContent).toBe(
			"Clocks, Calendar",
		);
		expect(root.querySelector(".mt-review-note")?.textContent).toBe(
			"Too few answers so far.",
		);
	});
});

describe("the steps a review advises", () => {
	test("are numbered in order, each the topic it is for over what to do, or what to do alone", () => {
		draw(
			<AdviceSteps
				steps={[
					{ id: "0", topic: "Enumeration", text: "List the cases." },
					{ id: "1", text: "Reread the question." },
				]}
			/>,
		);

		expect(
			[...root.querySelectorAll("ol.mt-advice > li")].map((step) => [
				step.querySelector(".mt-step-num")?.textContent,
				step.querySelector(".mt-advice-topic")?.textContent ?? null,
				step.querySelector(".mt-review-text > span:last-child")?.textContent,
			]),
		).toEqual([
			["1", "Enumeration", "List the cases."],
			["2", null, "Reread the question."],
		]);
	});
});
