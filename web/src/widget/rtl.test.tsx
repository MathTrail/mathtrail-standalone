import type { CallToolResult } from "@modelcontextprotocol/client";
import { afterEach, beforeEach, describe, expect, test, vi } from "vitest";
import {
	type Drawn,
	drawCard,
	foldIn,
	openCard,
	press,
	takeDown,
	unfold,
} from "./testing/card";
import { type ToolCall, toolInfoOf } from "./testing/host";
import {
	answered,
	coming,
	editRefused,
	fence,
	fenceInArabic,
	fenceSolutionInArabic,
	firstRun,
	profileRead,
	progress,
	refused,
	standing,
	writing,
} from "./testing/lesson";

let drawn: Drawn | undefined;

// The header names the build, and the snapshots name one of their own rather
// than whatever the environment the tests run in was given.
beforeEach(() => {
	vi.stubEnv("VITE_VERSION", "v0.2.1");
});

afterEach(() => {
	takeDown(drawn?.root);
	drawn = undefined;
	vi.unstubAllEnvs();
});

// wrongInArabic is the Arabic fence answered with the wrong B: the trap
// behind it, and the solution, in the task's language.
const wrongInArabic = answered({
	trap: {
		id: "fence_gaps",
		text: "عُدّت المسافات بدلًا من الأعمدة.",
		repeated: false,
	},
	solution: fenceSolutionInArabic,
});

// service answers the card's calls as the service would for the Arabic fence,
// a try at which the checks have turned down once.
function service({ name }: ToolCall): CallToolResult {
	switch (name) {
		case "read_progress":
			return progress;
		case "read_task":
			return writing(1);
		default:
			return wrongInArabic;
	}
}

// draw draws the card a host in locale hands payload to.
async function draw(payload: object, locale = "ar-EG"): Promise<HTMLElement> {
	drawn = await drawCard(payload, { tools: service, context: { locale } });
	return drawn.root;
}

// markupOf is the markup of the card, or of the part of it named by selector
// or given, as its snapshot keeps it: the drawings of its icons left out,
// since they say nothing of the way its words run.
function markupOf(
	root: HTMLElement,
	selector: string | Element | null = ".mt-widget",
): Element {
	const part =
		typeof selector === "string" ? root.querySelector(selector) : selector;
	const card = part?.cloneNode(true);
	if (!(card instanceof Element)) {
		throw new Error("the card is not drawn");
	}
	for (const icon of card.querySelectorAll("svg")) {
		icon.replaceChildren();
	}
	return card;
}

describe("a card in a language written right to left", () => {
	test.each([
		["ar-EG", "ar"],
		["fa-IR", "fa"],
		["ur-PK", "ur"],
	])("in %s lays the page out right to left, in %s", async (locale, tag) => {
		const root = await draw(fence, locale);

		expect(document.documentElement.dir).toBe("rtl");
		expect(document.documentElement.lang).toBe(tag);
		expect(root.querySelector(".mt-widget")?.classList).toContain("mt-rtl");
	});

	test("keeps the task in its own language and direction, and its drawing left to right", async () => {
		const root = await draw(fenceInArabic);

		const question = root.querySelector(".mt-task-text");
		expect(question?.getAttribute("lang")).toBe("ar");
		expect(question?.getAttribute("dir")).toBe("rtl");
		const drawing = root.querySelector("pre");
		expect(drawing?.getAttribute("dir")).toBe("ltr");
		expect(drawing?.textContent).toBe(fenceInArabic.task.drawing.trimEnd());
	});

	test("keeps a task in English running left to right", async () => {
		const root = await draw(fence);

		const question = root.querySelector(".mt-task-text");
		expect(question?.getAttribute("lang")).toBe("en");
		expect(question?.getAttribute("dir")).toBe("ltr");
	});
});

// The markup of eleven screens in Arabic, against the snapshots the review last
// read: which of their parts run right to left and which do not, and what
// they say. A change to them is a change to read in the snapshots' diff.
describe("the screens in Arabic", () => {
	test("the task", async () => {
		const root = await draw(fenceInArabic);

		expect(markupOf(root)).toMatchSnapshot();
	});

	test("the result of a wrong answer", async () => {
		const root = await draw(fenceInArabic);
		const wrong = [...root.querySelectorAll<HTMLElement>(".mt-option")].find(
			(row) => row.querySelector(".mt-option-letter")?.textContent === "B",
		);
		if (wrong === undefined) {
			throw new Error("the card has no option B");
		}
		press(wrong);
		await vi.waitFor(() =>
			expect(root.querySelector(".mt-verdict-line")).not.toBeNull(),
		);

		expect(markupOf(root)).toMatchSnapshot();
	});

	test("the wait for a task asked for", async () => {
		const opened = await openCard({
			context: { locale: "ar-EG", toolInfo: toolInfoOf("next_task") },
		});
		drawn = { root: opened.root, heard: opened.heard };
		await vi.waitFor(() =>
			expect(opened.root.querySelector(".mt-gen")).not.toBeNull(),
		);

		expect(markupOf(opened.root)).toMatchSnapshot();
	});

	test("the wait for a task being written, a try turned down", async () => {
		const root = await draw({
			...coming,
			child: fenceInArabic.child,
			language: "ar",
		});
		await vi.waitFor(
			() =>
				expect(root.querySelector(".mt-news .mt-verdict-line")).not.toBeNull(),
			{ timeout: 3000 },
		);

		expect(markupOf(root)).toMatchSnapshot();
	});

	test("a try that did not pass", async () => {
		const root = await draw({ ...refused, child: fenceInArabic.child });

		expect(markupOf(root)).toMatchSnapshot();
	});

	test("the progress", async () => {
		expect(markupOf(await draw(standing))).toMatchSnapshot();
	});

	test("the review of the progress, opened", async () => {
		const root = await draw(standing);
		unfold(root, "المراجعة");

		expect(
			markupOf(root, foldIn(root, "المراجعة").closest(".mt-fold")),
		).toMatchSnapshot();
	});

	test("the review of the progress, a page the chat did not open", async () => {
		drawn = await drawCard(standing, {
			tools: service,
			context: { locale: "ar-EG" },
			links: "refuse",
		});
		const { root } = drawn;
		unfold(root, "المراجعة");
		const review = foldIn(root, "المراجعة").closest(".mt-fold");
		press(review?.querySelector<HTMLElement>("a.mt-link") ?? root);
		await vi.waitFor(() =>
			expect(
				review?.querySelector(".mt-link-refused:not(:empty)"),
			).not.toBeNull(),
		);

		expect(markupOf(root, review)).toMatchSnapshot();
	});

	test("the profile, its cards chosen to be in Arabic", async () => {
		const inArabic = {
			...profileRead,
			profile: { ...profileRead.profile, ui_language: "ar" },
		};

		expect(markupOf(await draw(inArabic))).toMatchSnapshot();
	});

	test("the first sign-in", async () => {
		expect(markupOf(await draw(firstRun))).toMatchSnapshot();
	});

	test("the form of the profile, refused field by field", async () => {
		drawn = await drawCard(standing, {
			tools: () => editRefused,
			context: { locale: "ar-EG" },
		});
		const { root } = drawn;
		const edit = root.querySelector<HTMLElement>(".mt-fields-head .mt-btn");
		press(
			edit
				?.closest(".mt-fold")
				?.querySelector<HTMLElement>(".mt-fold-button") ?? root,
		);
		press(edit ?? root);
		const box = root.querySelector<HTMLInputElement>(".mt-form .mt-input");
		if (box === null) {
			throw new Error("the card shows no form");
		}
		box.value = "نجم";
		box.dispatchEvent(new Event("input", { bubbles: true }));
		await vi.waitFor(() => expect(box.value).toBe("نجم"));
		press(root.querySelector<HTMLElement>(".mt-form .mt-btn-primary") ?? root);
		await vi.waitFor(() =>
			expect(root.querySelectorAll(".mt-field-problem")).toHaveLength(5),
		);

		expect(markupOf(root, ".mt-form")).toMatchSnapshot();
	});
});
