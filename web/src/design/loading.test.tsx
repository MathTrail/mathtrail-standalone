import { render } from "preact";
import { act } from "preact/test-utils";
import { renderToString } from "preact-render-to-string";
import { afterEach, describe, expect, test } from "vitest";
import { LoadingBar, ProgressSkeleton } from "./loading";

const root = document.createElement("div");

// grades are the runs of an outline's course of eleven steps: four, three and
// four steps long.
const grades = [
	{ first: 1, last: 4 },
	{ first: 5, last: 7 },
	{ first: 8, last: 11 },
];

afterEach(() => {
	act(() => render(null, root));
});

function draw(element: preact.JSX.Element): void {
	act(() => render(element, root));
}

describe("the bar over a screen being read", () => {
	test("runs where a screen reader passes it over", () => {
		draw(<LoadingBar />);

		const bar = root.querySelector(".mt-loading-bar");
		expect(bar?.getAttribute("aria-hidden")).toBe("true");
		expect(bar?.querySelector(".mt-loading-run")).not.toBeNull();
	});
});

describe("the outline of a progress being read", () => {
	test("tells a screen reader, as a status, what is being read, and hides every shape from it", () => {
		draw(<ProgressSkeleton label="Loading the progress…" runs={grades} />);

		const status = root.querySelector(".mt-skeleton > output");
		expect(status?.textContent).toBe("Loading the progress…");
		expect(status?.classList.contains("mt-vh")).toBe(true);
		const read = [...root.querySelectorAll(".mt-skeleton > *")].filter(
			(part) => part.getAttribute("aria-hidden") !== "true",
		);
		expect(read).toEqual([status]);
	});

	// A screen reader says what changes in a status already on the page, and
	// passes over what one says as it comes: the outline is first drawn with
	// its status empty.
	test("is first drawn with its status empty, for the words to come into it", () => {
		const first = renderToString(
			<ProgressSkeleton label="Loading the progress…" runs={grades} />,
		);

		expect(first).toContain('<output class="mt-vh"></output>');
		expect(first).not.toContain("Loading the progress…");
	});

	test.each<[{ first: number; last: number }[], number, string[]]>([
		[grades, 11, ["4", "3", "4"]],
		[
			[
				{ first: 1, last: 1 },
				{ first: 2, last: 3 },
				{ first: 4, last: 5 },
			],
			5,
			["1", "2", "2"],
		],
	])(
		"draws the runs %j under a course of %i steps that they take in turn",
		(runs, steps, spans) => {
			draw(<ProgressSkeleton label="Loading the progress…" runs={runs} />);

			expect(root.querySelectorAll(".mt-segments > .mt-segment")).toHaveLength(
				steps,
			);
			const table = root.querySelector(".mt-grades-runs");
			expect(table?.getAttribute("aria-hidden")).toBe("true");
			expect(
				table?.querySelector(":scope > colgroup > col")?.getAttribute("span"),
			).toBe(String(steps));
			expect(
				[...(table?.querySelectorAll(":scope > tbody > tr > td") ?? [])].map(
					(run) => run.getAttribute("colspan"),
				),
			).toEqual(spans);
		},
	);

	test("draws the switch of the while, what comes next and four sections", () => {
		draw(<ProgressSkeleton label="Loading the progress…" runs={grades} />);

		expect(
			root.querySelectorAll(".mt-switch > .mt-switch-option"),
		).toHaveLength(2);
		expect(root.querySelector(".mt-progress > .mt-note")).not.toBeNull();
		expect(root.querySelectorAll(".mt-folds > .mt-fold")).toHaveLength(4);
	});

	test("sets no style on an element", () => {
		draw(
			<>
				<LoadingBar />
				<ProgressSkeleton label="Loading the progress…" runs={grades} />
			</>,
		);

		expect(root.querySelector("[style]")).toBeNull();
	});
});
