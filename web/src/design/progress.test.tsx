import { render } from "preact";
import { act } from "preact/test-utils";
import { afterEach, describe, expect, test } from "vitest";
import { Icon, type IconName } from "./icons";
import {
	GradeLegend,
	type GradeRun,
	MoveCounts,
	MoveLegend,
	MoveLine,
	ProfileFields,
	RankList,
	RankSummary,
	Segments,
	StatList,
	StatusDots,
	type StatusTone,
} from "./progress";
import { drawnAlone } from "./testing/drawing";

const root = document.createElement("div");

afterEach(() => {
	act(() => render(null, root));
});

function draw(element: preact.JSX.Element): void {
	act(() => render(element, root));
}

// runs are three runs of a course of eleven steps, the course standing in the
// first.
const runs: GradeRun[] = [
	{
		first: 1,
		last: 4,
		label: "Gr. 1–2",
		said: "Ranks 1–4, here.",
		current: true,
	},
	{ first: 5, last: 7, label: "Gr. 3–4", said: "Ranks 5–7.", current: false },
	{ first: 8, last: 11, label: "Gr. 5–6", said: "Ranks 8–11.", current: false },
];

// fills are how far each step of the course drawn is filled.
function fills(): (string | null | undefined)[] {
	return [...root.querySelectorAll("svg.mt-segment")].map((segment) =>
		segment.querySelector(".mt-segment-fill")?.getAttribute("width"),
	);
}

// striped are how each step of the course drawn is drawn: how far it is
// filled, how far the stripes of a move reach under the fill, and the way the
// move's stripes say — the last two null on a step no move is drawn across.
function striped(): [string, string | null, string | null][] {
	return [...root.querySelectorAll("svg.mt-segment")].map((segment) => [
		segment.querySelector(".mt-segment-fill")?.getAttribute("width") ?? "",
		segment.querySelector(".mt-segment-stripes")?.getAttribute("width") ?? null,
		segment.querySelector("pattern")?.getAttribute("class") ?? null,
	]);
}

describe("a course that moved", () => {
	test("is filled as far as where it stood and striped on to where it stands, when it moved up", () => {
		draw(
			<Segments of={4} filled={1} part={0.6} was={{ filled: 1, part: 0.2 }} />,
		);

		expect(striped()).toEqual([
			["100%", null, null],
			["20%", "60%", "mt-stripes mt-stripes-gain"],
			["0%", null, null],
			["0%", null, null],
		]);
	});

	test("is filled as far as where it stands and striped on to where it stood, when it moved back", () => {
		draw(
			<Segments of={4} filled={1} part={0.2} was={{ filled: 1, part: 0.6 }} />,
		);

		expect(striped()).toEqual([
			["100%", null, null],
			["20%", "60%", "mt-stripes mt-stripes-loss"],
			["0%", null, null],
			["0%", null, null],
		]);
	});

	test("is striped across every step a rank's move crossed", () => {
		draw(
			<Segments of={4} filled={2} part={0.3} was={{ filled: 1, part: 0.8 }} />,
		);

		expect(striped()).toEqual([
			["100%", null, null],
			["80%", "100%", "mt-stripes mt-stripes-gain"],
			["0%", "30%", "mt-stripes mt-stripes-gain"],
			["0%", null, null],
		]);
	});

	test.each<[string, number, number, [string, string | null, string | null]]>([
		[
			"a gain of a point ends at the edge",
			0.44,
			0.43,
			["30%", "44%", "mt-stripes mt-stripes-gain"],
		],
		[
			"a gain from the step's start reaches past the edge",
			0.01,
			0,
			["0%", "14%", "mt-stripes mt-stripes-gain"],
		],
		[
			"a step back of a point starts at the edge",
			0.43,
			0.44,
			["43%", "57%", "mt-stripes mt-stripes-loss"],
		],
		[
			"a step back near the step's end keeps within it",
			0.98,
			0.99,
			["86%", "100%", "mt-stripes mt-stripes-loss"],
		],
	])(
		"is striped over a seventh of a step at least, where %s",
		(_, now, then, step) => {
			draw(
				<Segments
					of={3}
					filled={1}
					part={now}
					was={{ filled: 1, part: then }}
				/>,
			);

			expect(striped()[1]).toEqual(step);
		},
	);

	test("draws a move seen on the steps it crossed as it is, and nothing past its edge", () => {
		draw(
			<Segments of={3} filled={1} part={0} was={{ filled: 0, part: 0.8 }} />,
		);

		expect(striped()).toEqual([
			["80%", "100%", "mt-stripes mt-stripes-gain"],
			["0%", null, null],
			["0%", null, null],
		]);
	});

	test("shows a rank just reached, at the start of its step", () => {
		draw(
			<Segments of={3} filled={1} part={0} was={{ filled: 0, part: 0.99 }} />,
		);

		expect(striped()).toEqual([
			["99%", "100%", "mt-stripes mt-stripes-gain"],
			["0%", "14%", "mt-stripes mt-stripes-gain"],
			["0%", null, null],
		]);
	});

	test("draws each step's stripes by a pattern no other drawing on the page is named by, with no style set on an element", () => {
		draw(
			<Segments of={4} filled={2} part={0.3} was={{ filled: 1, part: 0.8 }} />,
		);

		const patterns = [...root.querySelectorAll("pattern")];
		const names = patterns.map((pattern) => pattern.id);
		expect(new Set(names).size).toBe(2);
		expect(
			[...root.querySelectorAll(".mt-segment-stripes")].map((stripes) =>
				stripes.getAttribute("fill"),
			),
		).toEqual(names.map((name) => `url(#${name})`));
		expect(
			patterns.every(
				(pattern) => pattern.getAttribute("patternUnits") === "userSpaceOnUse",
			),
		).toBe(true);
		expect(root.querySelector("[style]")).toBeNull();
	});
});

describe("how a rank moved", () => {
	test("is a line in the colour of its way, after an arrow a screen reader does not read, and read out whenever it changes", () => {
		draw(<MoveLine way="loss" text="The last task moved the bar back" />);

		const line = root.querySelector(".mt-rank-move");
		expect(line?.getAttribute("data-way")).toBe("loss");
		expect(line?.getAttribute("aria-live")).toBe("polite");
		expect(line?.querySelector("[aria-hidden='true']")?.textContent).toBe("↓");
		expect(line?.textContent).toBe("↓The last task moved the bar back");
	});

	test("is a plain line where nothing moved, and an empty one with nothing to say", () => {
		draw(<MoveLine text="The last task didn't move the bar" />);
		expect(root.querySelector(".mt-rank-move")?.hasAttribute("data-way")).toBe(
			false,
		);
		expect(root.querySelector(".mt-rank-move [aria-hidden]")).toBeNull();

		draw(<MoveLine />);
		expect(root.querySelector(".mt-rank-move")?.textContent).toBe("");
	});

	test("is told between the course, with what marks it, and the next rank to reach", () => {
		draw(
			<RankSummary
				name="Rank 3"
				meta="of 11"
				line="Next — rank 4."
				move={<MoveLine text="moved" />}
			>
				<Segments of={11} filled={2} part={0.43} />
				<GradeLegend of={11} runs={runs} note="For parents." />
			</RankSummary>,
		);

		expect(
			[...(root.querySelector(".mt-rank")?.children ?? [])].map(
				(child) => child.className,
			),
		).toEqual([
			"mt-rank-head",
			"mt-segments",
			"mt-grades",
			"mt-rank-move",
			"mt-rank-line",
		]);
	});

	test("is counted by the title of a list, each way that moved with its arrow, and the same in words for a screen reader", () => {
		draw(
			<MoveCounts
				up="2"
				down="1"
				label="2 topics moved forward and 1 topic slipped back"
			/>,
		);

		expect(
			[...root.querySelectorAll(".mt-move-counted [data-way]")].map((count) => [
				count.getAttribute("data-way"),
				count.textContent,
			]),
		).toEqual([
			["gain", "↑ 2"],
			["loss", "↓ 1"],
		]);
		expect(
			root.querySelector(".mt-move-counted")?.getAttribute("aria-hidden"),
		).toBe("true");
		expect(root.querySelector(".mt-move-counts .mt-vh")?.textContent).toBe(
			"2 topics moved forward and 1 topic slipped back",
		);
	});

	test("has a legend: a sample of each way's stripes beside its name, and the while", () => {
		draw(<MoveLegend gain="gain" loss="loss" period="over the past week" />);

		expect(
			[...root.querySelectorAll(".mt-move-sample")].map((sample) => [
				sample.getAttribute("data-way"),
				sample.getAttribute("aria-hidden"),
			]),
		).toEqual([
			["gain", "true"],
			["loss", "true"],
		]);
		expect(root.querySelector(".mt-move-legend")?.textContent).toBe(
			"gainloss· over the past week",
		);
	});

	test("is a topic's word in the colour of its way, after its arrow", () => {
		draw(
			<RankList
				rows={[
					{
						id: "a",
						label: "Ordering",
						word: "new rank",
						way: "gain",
						name: "Hill",
						segments: { of: 3, filled: 1 },
					},
					{
						id: "b",
						label: "Gaps",
						word: "even",
						name: "Ford",
						segments: { of: 3, filled: 1 },
					},
				]}
			/>,
		);

		const words = [...root.querySelectorAll(".mt-rank-word")];
		expect(
			words.map((word) => [word.getAttribute("data-way"), word.textContent]),
		).toEqual([
			["gain", "↑new rank"],
			[null, "even"],
		]);
		expect(words[0]?.querySelector("[aria-hidden='true']")?.textContent).toBe(
			"↑",
		);
	});
});

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
				name="Rank 3"
				meta="of 11"
				line="Next — rank 4."
			>
				<Segments of={11} filled={2} part={0.43} />
			</RankSummary>,
		);

		const summary = root.querySelector("section.mt-rank");
		expect(summary?.getAttribute("aria-label")).toBe("Overall rating");
		expect(summary?.querySelector(".mt-rank-name")?.textContent).toBe("Rank 3");
		expect(summary?.querySelector(".mt-rank-head .mt-meta")?.textContent).toBe(
			"of 11",
		);
		expect(summary?.querySelectorAll("svg.mt-segment")).toHaveLength(11);
		expect(summary?.querySelector(".mt-rank-line")?.textContent).toBe(
			"Next — rank 4.",
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

describe("a legend of grades", () => {
	test("spans each run's steps of the course in a column of its own, marks the run the course stands in, and says each run for a screen reader instead", () => {
		draw(<GradeLegend of={11} runs={runs} note="For parents." />);

		const table = root.querySelector(".mt-grades > table.mt-grades-runs");
		expect(table?.getAttribute("aria-hidden")).toBe("true");
		expect(
			table?.querySelector(":scope > colgroup > col")?.getAttribute("span"),
		).toBe("11");
		const cells = [
			...(table?.querySelectorAll(":scope > tbody > tr > td") ?? []),
		];
		expect(
			cells.map((cell) => [
				cell.getAttribute("colspan"),
				cell.querySelector(".mt-grades-bracket") !== null,
				cell.querySelector(".mt-grades-label")?.textContent,
				cell.hasAttribute("data-current"),
			]),
		).toEqual([
			["4", true, "Gr. 1–2", true],
			["3", true, "Gr. 3–4", false],
			["4", true, "Gr. 5–6", false],
		]);
		expect(
			[...root.querySelectorAll(".mt-grades > ul.mt-vh > li")].map(
				(said) => said.textContent,
			),
		).toEqual(["Ranks 1–4, here.", "Ranks 5–7.", "Ranks 8–11."]);
		expect(
			root.querySelector(".mt-grades > .mt-grades-note")?.textContent,
		).toBe("For parents.");
		expect(root.querySelector("[style]")).toBeNull();
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
						name: "Rank 4",
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
		).toEqual(["aheadRank 4", "no answers yet"]);
		expect(
			rows.map((row) =>
				row.querySelector(".mt-segments")?.getAttribute("data-tone"),
			),
		).toEqual(["2", "ink"]);
		expect(rows[1]?.querySelectorAll(".mt-segment-open")).toHaveLength(11);
	});
});

describe("a list", () => {
	test.each<[StatusTone, IconName]>([
		["correct", "check"],
		["wrong", "cross"],
		["skipped", "dash"],
	])("marks a line that went %s with its own mark and colour", (tone, mark) => {
		draw(
			<StatList
				label="Recent answers"
				rows={[{ id: "a", label: "Clocks", status: { tone, label: "Words" } }]}
			/>,
		);

		const status = root.querySelector(".mt-status");
		expect(status?.classList.contains(`mt-status-${tone}`)).toBe(true);
		expect(status?.querySelector("svg")?.outerHTML).toBe(
			drawnAlone(<Icon name={mark} size={16} />),
		);
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

	test("carries no label of its own where what it stands in names it", () => {
		draw(<StatList rows={[{ id: "a", label: "Clocks" }]} />);

		expect(root.querySelector(".mt-section-label")).toBeNull();
		expect(root.querySelector(".mt-row")?.textContent).toBe("Clocks");
	});
});

describe("the dots of a run of entries", () => {
	test("are a dot for each entry in its tone's colour, the first first, said in words for a screen reader alone", () => {
		draw(
			<StatusDots
				tones={["wrong", "skipped", "correct"]}
				label="Wrong, Skipped, Right"
			/>,
		);

		const dots = root.querySelector(".mt-status-dots .mt-dots");
		expect(dots?.getAttribute("aria-hidden")).toBe("true");
		expect(
			[...(dots?.querySelectorAll(".mt-dot") ?? [])].map((dot) =>
				dot.getAttribute("data-tone"),
			),
		).toEqual(["wrong", "skipped", "correct"]);
		expect(root.querySelector(".mt-status-dots .mt-vh")?.textContent).toBe(
			"Wrong, Skipped, Right",
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

	test("carry no label where what they stand in names them, and keep what can be done at their head", () => {
		draw(
			<ProfileFields
				fields={[{ term: "Grade", value: "3" }]}
				action={<button type="button">Edit</button>}
			/>,
		);

		expect(root.querySelector(".mt-section-label")).toBeNull();
		expect(
			[...(root.querySelector(".mt-fields-head")?.children ?? [])].map(
				(part) => part.textContent,
			),
		).toEqual(["Edit"]);
	});

	test("have no head with neither a label nor anything to be done", () => {
		draw(<ProfileFields fields={[{ term: "Grade", value: "3" }]} />);

		expect(root.querySelector(".mt-fields-head")).toBeNull();
		expect(root.querySelector("dt")?.textContent).toBe("Grade");
	});
});
