import { render } from "preact";
import { act } from "preact/test-utils";
import { afterEach, describe, expect, test } from "vitest";
import {
	ProfileFields,
	RankList,
	RankSummary,
	Segments,
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

// fills are how far each step of the course drawn is filled.
function fills(): (string | null | undefined)[] {
	return [...root.querySelectorAll("svg.mt-segment")].map((segment) =>
		segment.querySelector(".mt-segment-fill")?.getAttribute("width"),
	);
}

describe("a course of steps", () => {
	test("fills the steps behind, the one under way as far as it has come, and leaves the rest", () => {
		draw(<Segments of={5} filled={2} part={0.43} />);

		expect(fills()).toEqual(["100%", "100%", "43%", "0%", "0%"]);
		expect(root.querySelector(".mt-segments")?.getAttribute("data-tone")).toBe(
			"ink",
		);
		// Unnamed, it is a picture of what the words beside it say.
		expect(root.querySelector(".mt-vh")).toBeNull();
		expect(
			[...root.querySelectorAll("svg.mt-segment")].every(
				(segment) => segment.getAttribute("aria-hidden") === "true",
			),
		).toBe(true);
		// A card's page promises nothing for a style set on an element.
		expect(root.querySelector("[style]")).toBeNull();
	});

	test("says in words, for a screen reader alone, what it shows", () => {
		draw(
			<Segments
				of={11}
				filled={2}
				part={0.43}
				label="43% of the way to the next rank"
			/>,
		);

		expect(root.querySelector(".mt-segments .mt-vh")?.textContent).toBe(
			"43% of the way to the next rank",
		);
		expect(
			root.querySelectorAll("svg.mt-segment[aria-hidden='true']"),
		).toHaveLength(11);
	});

	test("is drawn open where nothing is measured yet", () => {
		draw(<Segments of={4} filled={0} open />);

		expect(root.querySelectorAll(".mt-segment-open")).toHaveLength(4);
		expect(root.querySelector("svg.mt-segment")).toBeNull();
	});

	test("is filled in the step of the ramp it is given", () => {
		draw(<Segments of={11} filled={6} tone={3} />);

		expect(root.querySelector(".mt-segments")?.getAttribute("data-tone")).toBe(
			"3",
		);
	});

	test.each([
		[-0.5, "0%"],
		[1.5, "100%"],
	])("keeps a part of %s within its step", (part, width) => {
		draw(<Segments of={3} filled={1} part={part} />);

		expect(fills()[1]).toBe(width);
	});
});

describe("a rank summary", () => {
	test("puts the rank's name over the line that places it, its course and what comes next, named for a screen reader", () => {
		draw(
			<RankSummary
				label="Overall rating"
				name="River crossing"
				meta="rank 3 of 11 · rating 1573"
				line="Next rank — Hill."
			>
				<Segments of={11} filled={2} part={0.43} />
			</RankSummary>,
		);

		const summary = root.querySelector("section.mt-rank");
		expect(summary?.getAttribute("aria-label")).toBe("Overall rating");
		expect(summary?.querySelector(".mt-rank-name")?.textContent).toBe(
			"River crossing",
		);
		expect(summary?.querySelector(".mt-rank-head .mt-meta")?.textContent).toBe(
			"rank 3 of 11 · rating 1573",
		);
		expect(summary?.querySelectorAll("svg.mt-segment")).toHaveLength(11);
		expect(summary?.querySelector(".mt-rank-line")?.textContent).toBe(
			"Next rank — Hill.",
		);
	});

	test("named by its name alone, is not named twice", () => {
		draw(
			<RankSummary name="Trial series" meta="3 of 5" line="Ranks come later.">
				<Segments of={5} filled={3} />
			</RankSummary>,
		);

		expect(
			root.querySelector("section.mt-rank")?.hasAttribute("aria-label"),
		).toBe(false);
	});
});

describe("a list of ranks", () => {
	test("gives each topic its name, its mark, how it stands and its rank, over its own course", () => {
		draw(
			<RankList
				label="Topics"
				note="The rank in each topic"
				rows={[
					{
						id: "ordering",
						label: "Ordering",
						mark: { tone: "correct", label: "Mastered" },
						word: "ahead",
						name: "Hill",
						segments: { of: 11, filled: 3, part: 0.27, tone: 2 },
					},
					{
						id: "pigeonhole",
						label: "Pigeonhole principle",
						word: "no answers yet",
						segments: { of: 11, filled: 0, open: true },
					},
				]}
			/>,
		);

		expect(root.querySelector(".mt-section-label")?.textContent).toBe("Topics");
		expect(root.querySelector(".mt-list-lead")?.textContent).toBe(
			"The rank in each topic",
		);
		const rows = [...root.querySelectorAll(".mt-rank-row")];
		expect(
			rows.map((row) => row.querySelector(".mt-rank-row-start")?.textContent),
		).toEqual(["OrderingMastered", "Pigeonhole principle"]);
		expect(
			rows.map((row) => row.querySelector(".mt-rank-row-end")?.textContent),
		).toEqual(["aheadHill", "no answers yet"]);
		expect(
			rows.map((row) =>
				row.querySelector(".mt-segments")?.getAttribute("data-tone"),
			),
		).toEqual(["2", "ink"]);
		expect(rows[1]?.querySelectorAll(".mt-segment-open")).toHaveLength(11);
	});
});

describe("a list", () => {
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

	test("counts a line with a dot for each time, up to a handful, in a frame of its own", () => {
		draw(
			<StatList
				label="Mistakes that repeat"
				rows={[
					{
						id: "a",
						label: "Missed a case",
						count: { times: 3, label: "3 times" },
					},
					{
						id: "b",
						label: "Counted twice",
						count: { times: 8, label: "8 times" },
					},
				]}
				framed
			/>,
		);

		expect(root.querySelector("section.mt-list-framed")).not.toBeNull();
		expect(
			[...root.querySelectorAll(".mt-row")].map((row) => row.textContent),
		).toEqual(["Missed a case3 times", "Counted twice8 times"]);
		expect(
			[...root.querySelectorAll(".mt-dots")].map((dots) => [
				dots.getAttribute("aria-hidden"),
				dots.querySelectorAll(".mt-dot").length,
			]),
		).toEqual([
			["true", 3],
			["true", 5],
		]);
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
		expect(root.querySelector(".mt-list-framed")).toBeNull();
		expect(root.querySelector(".mt-list-note")?.textContent).toBe(
			"In all, 1 task was left without an answer.",
		);
	});
});

describe("a profile's fields", () => {
	test("name each field, say what it holds — a list as its names, each apart — and what explains it", () => {
		draw(
			<ProfileFields
				label="Profile · for the parent"
				fields={[
					{ term: "Grade", value: "3", note: "Only a label." },
					{ term: "Interests", value: ["Space", "Animals"] },
				]}
			/>,
		);

		expect(
			[...root.querySelectorAll(".mt-fields dl > div")].map((field) => [
				field.querySelector("dt")?.textContent,
				field.querySelector("dd")?.textContent,
			]),
		).toEqual([
			["Grade", "3Only a label."],
			["Interests", "SpaceAnimals"],
		]);
		expect(root.querySelector("dd .mt-field-note")?.textContent).toBe(
			"Only a label.",
		);
		expect(
			[...root.querySelectorAll(".mt-chips .mt-chip")].map(
				(chip) => chip.textContent,
			),
		).toEqual(["Space", "Animals"]);
	});

	test("put what can be done at their head, and what became of it under the head", () => {
		draw(
			<ProfileFields
				label="Profile · for the parent"
				fields={[{ term: "Grade", value: "3" }]}
				action={<button type="button">Edit profile</button>}
				status={<p class="mt-action-note">Sent to the chat</p>}
			/>,
		);

		expect(root.querySelector(".mt-fields-head button")?.textContent).toBe(
			"Edit profile",
		);
		expect(
			root.querySelector(".mt-fields-head + .mt-action-note")?.textContent,
		).toBe("Sent to the chat");
	});

	// A file edited by hand may name an interest twice; a list of names says
	// each once.
	test("draw a name given twice once", () => {
		draw(
			<ProfileFields
				label="Profile · for the parent"
				fields={[{ term: "Interests", value: ["space", "chess", "space"] }]}
			/>,
		);

		expect(
			[...root.querySelectorAll(".mt-chip")].map((chip) => chip.textContent),
		).toEqual(["space", "chess"]);
	});

	test("have nothing at their head but their label when nothing can be done", () => {
		draw(
			<ProfileFields label="Your data" fields={[{ term: "a", value: "b" }]} />,
		);

		expect(
			[...(root.querySelector(".mt-fields-head")?.children ?? [])].map(
				(part) => part.textContent,
			),
		).toEqual(["Your data"]);
	});
});
