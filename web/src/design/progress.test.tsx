import { render } from "preact";
import { act } from "preact/test-utils";
import { afterEach, describe, expect, test } from "vitest";
import {
	ProfileFields,
	RatingSummary,
	StatList,
	type StatusTone,
} from "./progress";

const root = document.createElement("div");

afterEach(() => {
	act(() => render(null, root));
});

function draw(element: preact.JSX.Element): void {
	act(() => render(element, root));
}

describe("a rating summary", () => {
	test("puts the number under the line that places it, named for a screen reader", () => {
		draw(
			<RatingSummary
				label="Overall rating"
				rankLabel="Rank 3 of 11 · River crossing"
				rating="1573"
				ratingLabel="overall rating"
			/>,
		);

		const summary = root.querySelector("section.mt-rating");
		expect(summary?.getAttribute("aria-label")).toBe("Overall rating");
		expect(summary?.querySelector(".mt-rating-rank")?.textContent).toBe(
			"Rank 3 of 11 · River crossing",
		);
		expect(summary?.querySelector(".mt-rating-num")?.textContent).toBe("1573");
		expect(summary?.querySelector(".mt-meta")?.textContent).toBe(
			"overall rating",
		);
		expect(summary?.querySelector(".mt-pips")).toBeNull();
	});

	test("shows how many steps are behind, hidden from a screen reader", () => {
		draw(
			<RatingSummary
				rankLabel="Trial series"
				rating="3 of 5"
				ratingLabel="the rating comes after 5 tasks"
				pips={{ on: 3, of: 5 }}
			/>,
		);

		// Named by its line above alone, it is not named twice.
		expect(
			root.querySelector("section.mt-rating")?.hasAttribute("aria-label"),
		).toBe(false);
		const pips = root.querySelector(".mt-pips");
		expect(pips?.getAttribute("aria-hidden")).toBe("true");
		expect(
			[...(pips?.children ?? [])].map((pip) => pip.getAttribute("data-on")),
		).toEqual(["true", "true", "true", "false", "false"]);
	});
});

describe("a list", () => {
	test("gives each line what it is about, its mark and its number, in order", () => {
		draw(
			<StatList
				label="Topics"
				rows={[
					{ id: "a", label: "Enumeration", value: "1627" },
					{
						id: "b",
						label: "Ordering",
						status: { tone: "correct", label: "Mastered" },
						value: "1712",
					},
				]}
			/>,
		);

		expect(root.querySelector(".mt-section-label")?.textContent).toBe("Topics");
		expect(
			[...root.querySelectorAll(".mt-row")].map((row) => row.textContent),
		).toEqual(["Enumeration1627", "OrderingMastered1712"]);
		expect(root.querySelector(".mt-list-note")).toBeNull();
	});

	test.each<[StatusTone, string]>([
		["correct", "M3.5 8.5l3 3 6-7"],
		["wrong", "M4.5 4.5l7 7M11.5 4.5l-7 7"],
		["skipped", "M4.5 8h7"],
	])("marks a line that went %s with its own mark and colour", (tone, mark) => {
		draw(
			<StatList
				label="Recent answers"
				rows={[{ id: "a", label: "Clocks", status: { tone, label: "Words" } }]}
			/>,
		);

		const status = root.querySelector(".mt-status");
		expect(status?.classList.contains(`mt-status-${tone}`)).toBe(true);
		expect(status?.querySelector("path")?.getAttribute("d")).toBe(mark);
		expect(status?.textContent).toBe("Words");
	});

	test("gives a line that counts how often, and a bar as long as its share", () => {
		draw(
			<StatList
				label="Mistakes that repeat"
				rows={[
					{
						id: "a",
						label: "Missed a case",
						bar: { share: 1, count: "3 times" },
					},
					{
						id: "b",
						label: "Counted twice",
						bar: { share: 2 / 3, count: "2 times" },
					},
				]}
			/>,
		);

		expect(
			[...root.querySelectorAll(".mt-row-bar")].map((row) => row.textContent),
		).toEqual(["Missed a case3 times", "Counted twice2 times"]);
		const charts = [...root.querySelectorAll("svg.mt-bar-chart")];
		expect(charts.map((chart) => chart.getAttribute("aria-hidden"))).toEqual([
			"true",
			"true",
		]);
		expect(
			charts.map((chart) =>
				chart.querySelector(".mt-bar-fill")?.getAttribute("width"),
			),
		).toEqual(["100%", "67%"]);
		// A card's page promises nothing for a style set on an element.
		expect(root.querySelector("[style]")).toBeNull();
	});

	test.each([
		[-0.5, "0%"],
		[1.5, "100%"],
	])("keeps a share of %s within its track", (share, width) => {
		draw(
			<StatList
				label="Mistakes that repeat"
				rows={[{ id: "a", label: "Missed a case", bar: { share, count: "" } }]}
			/>,
		);

		expect(root.querySelector(".mt-bar-fill")?.getAttribute("width")).toBe(
			width,
		);
	});

	test("tells apart two lines that say the same, and says more under them", () => {
		draw(
			<StatList
				label="Recent answers"
				rows={[
					{ id: "0", label: "Clocks" },
					{ id: "1", label: "Clocks" },
				]}
				note="In all, 1 task was left without an answer."
			/>,
		);

		expect(root.querySelectorAll(".mt-row")).toHaveLength(2);
		expect(root.querySelector(".mt-list-note")?.textContent).toBe(
			"In all, 1 task was left without an answer.",
		);
	});
});

describe("a profile's fields", () => {
	test("name each field, say what it holds and what explains it, and what can be done", () => {
		draw(
			<ProfileFields
				label="Profile · for the parent"
				fields={[
					{ term: "Grade", value: "3", note: "Only a label." },
					{ term: "Interests", value: "Space" },
				]}
				action={<button type="button">Edit profile</button>}
			/>,
		);

		expect(
			[...root.querySelectorAll(".mt-fields dl > div")].map((field) =>
				[...field.children].map((part) => part.textContent),
			),
		).toEqual([
			["Grade", "3", "Only a label."],
			["Interests", "Space"],
		]);
		expect(root.querySelector("dd.mt-field-note")?.textContent).toBe(
			"Only a label.",
		);
		expect(root.querySelector(".mt-fields-actions button")?.textContent).toBe(
			"Edit profile",
		);
	});

	test("have no place for actions when nothing can be done", () => {
		draw(
			<ProfileFields label="Your data" fields={[{ term: "a", value: "b" }]} />,
		);

		expect(root.querySelector(".mt-fields-actions")).toBeNull();
	});
});
