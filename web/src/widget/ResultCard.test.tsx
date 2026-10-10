import type { CallToolResult } from "@modelcontextprotocol/client";
import { act } from "preact/test-utils";
import { afterEach, describe, expect, test, vi } from "vitest";
import {
	buttonIn,
	drawCard as drawOnHost,
	press,
	takeDown,
} from "./testing/card";
import {
	fence,
	fenceInArabic,
	fenceSolution,
	fenceSolutionInArabic,
	fenceSolutionRow,
	fenceTotal,
	flagsSolution,
	flagsSolutionPicture,
	flagsTask,
	flagsTotal,
	nothingShown,
	shown,
	topicSaved,
	topicWords,
} from "./testing/lesson";

let root: HTMLElement;

afterEach(() => {
	takeDown(root);
	vi.restoreAllMocks();
});

// drawCard draws the card of how an answer went that a host hands payload to,
// and returns what the host hears.
async function drawCard(
	payload: object,
	options: Parameters<typeof drawOnHost>[1] = {},
) {
	const drawn = await drawOnHost(payload, options);
	root = drawn.root;
	await vi.waitFor(() => expect(root.querySelector("article")).not.toBeNull());
	return drawn.heard;
}

const text = (selector: string) => root.querySelector(selector)?.textContent;
const button = (label: string) => buttonIn(root, label);
const shownButtons = () =>
	[...root.querySelectorAll<HTMLButtonElement>(".mt-btns .mt-btn")].map(
		(found) => found.textContent,
	);
const steps = () =>
	[...root.querySelectorAll(".mt-steps li")].map((step) => step.textContent);
// facts are the panel's terms, each with what it says.
const facts = () =>
	[...root.querySelectorAll(".mt-fact")].map((fact) => [
		fact.querySelector("dt")?.textContent,
		fact.querySelector("dd")?.textContent,
	]);
// piecesOf is what the element under selector holds, piece by piece: an
// element by its class, and words as they read.
const piecesOf = (selector: string) =>
	[...(root.querySelector(selector)?.childNodes ?? [])].map((piece) =>
		piece instanceof HTMLElement ? piece.className : piece.textContent,
	);

describe("the card of how an answer went", () => {
	// It is drawn for an answer given elsewhere — in the chat — and nothing was
	// pressed on it: what has the focus keeps it.
	test("takes no focus as it is drawn", async () => {
		await drawCard(shown(fence));

		expect(document.hasFocus()).toBe(true);
		expect(document.activeElement).toBe(document.body);
	});

	test("tells a wrong answer from its trap, step by step, with the topic and how the rating moved", async () => {
		await drawCard(shown(fence, {}, {}), { links: "open" });

		expect(text(".mt-verdict-line")).toBe("Not quite — it's 5, not 4.");
		expect(root.querySelector(".mt-verdict-line svg")).not.toBeNull();
		expect(text(".mt-note-trap p")).toBe(
			"Counted the gaps instead of the posts.",
		);
		expect(root.querySelector(".mt-note-detail")).toBeNull();
		expect(steps()).toEqual([
			"1 12 ÷ 3 = 4 gaps.",
			"2 A straight fence with posts at both ends has one more post than gaps.",
			"3 4 + 1 = 5 posts.",
		]);
		expect(facts()).toEqual([
			["Type of task", "Gaps and boundaries"],
			["Rating in this topic", "1502 → 1480"],
		]);
		expect(root.querySelector(".mt-fact a.mt-link")?.getAttribute("href")).toBe(
			"https://mathtrail.app/en/topics/gaps-and-boundaries/",
		);
		expect(root.querySelector(".mt-rating-after")?.className).toBe(
			"mt-rating-after mt-rating-loss",
		);
		expect(shownButtons()).toEqual(["Another task", "Topic: Coach"]);
		// The task and its options are the task's card's, above.
		expect(root.querySelector(".mt-task-text")).toBeNull();
		expect(root.querySelector(".mt-option")).toBeNull();
	});

	test("says under the trap that the mistake has come up before", async () => {
		await drawCard(
			shown(fence, {
				trap: {
					id: "fence_gaps",
					text: "Counted the gaps instead of the posts.",
					repeated: true,
				},
			}),
		);

		expect(text(".mt-note-detail")).toBe("This mistake has come up before.");
	});

	test("praises a right answer, with no trap, the rating up in the colour of a gain", async () => {
		await drawCard(
			shown(fence, {
				choice: "C",
				correct: true,
				trap: null,
				rating: { before: 1502, after: 1519 },
			}),
		);

		expect(text(".mt-verdict-line")).toBe("Correct! It's 5.");
		expect(root.querySelector(".mt-note-trap")).toBeNull();
		expect(root.querySelector(".mt-rating-after")?.className).toBe(
			"mt-rating-after mt-rating-gain",
		);
	});

	test('tells "I don\'t know" the solution, with no verdict and the rating it cost', async () => {
		await drawCard(
			shown(fence, {
				choice: "?",
				trap: null,
				rating: { before: 1502, after: 1488 },
			}),
		);

		expect(text(".mt-verdict-line")).toBe("Here's how to solve it.");
		expect(root.querySelector(".mt-verdict-line svg")).toBeNull();
		expect(steps()).toHaveLength(3);
		expect(text(".mt-rating-move")).toBe("1502 → 1488");
	});

	test("shows how far the trial series has got instead of a rating, and offers no topic during it", async () => {
		await drawCard(
			shown(fence, { rating: null, trial: { answered: 3, of: 5 } }),
		);

		expect(facts()).toEqual([
			["Type of task", "Gaps and boundaries"],
			["Trial series", "3 of 5"],
		]);
		expect(root.querySelector(".mt-topic-button")).toBeNull();
		expect(shownButtons()).toEqual(["Another task"]);
	});

	test("names the topic with no link where the chat opens none", async () => {
		await drawCard(shown(fence));

		expect(facts()[0]).toEqual(["Type of task", "Gaps and boundaries"]);
		expect(root.querySelector(".mt-fact a")).toBeNull();
	});

	test("writes the rating as the card's language reads it, the arrow its way", async () => {
		await drawCard(shown(fenceInArabic));

		expect(piecesOf(".mt-rating-move")).toEqual([
			"mt-rating-before",
			" ← ",
			"mt-rating-after mt-rating-loss",
		]);
	});

	test("keeps markup in the texts it shows as text", async () => {
		const trick = "<img src=x onerror=alert(1)>";
		await drawCard(
			shown(
				{
					...fence,
					child: { ...fence.child, pseudonym: trick },
					task: { ...fence.task, options: { ...fence.task.options, B: trick } },
				},
				{
					trap: { id: "fence_gaps", text: trick, repeated: false },
					solution: trick,
				},
			),
		);

		expect(root.querySelector("img")).toBeNull();
		expect(text(".mt-note-trap p")).toBe(trick);
		expect(text(".mt-verdict-line")).toContain(trick);
	});

	test("marks the texts of the task as written in its language, which runs its way", async () => {
		await drawCard(
			shown({
				...fence,
				task: { ...fence.task, language: "yi" },
			}),
		);

		const trap = root.querySelector(".mt-note-trap p");
		expect(trap?.getAttribute("lang")).toBe("yi");
		expect(trap?.getAttribute("dir")).toBe("rtl");
		expect(root.querySelector(".mt-steps ol")?.getAttribute("dir")).toBe("rtl");
		// The card's own words stay in the card's language.
		expect(root.querySelector(".mt-section-label")?.hasAttribute("lang")).toBe(
			false,
		);
	});

	test("cuts the solution into its steps by the rules of the task's language", async () => {
		await drawCard(shown(fence, { solution: fenceSolution }));

		expect(steps()).toHaveLength(3);
	});
});

describe("the picture of the solution, on the card of how an answer went", () => {
	// pictured is the fence answered with fields, its solution drawn and its
	// total under the picture.
	const pictured = (fields: Parameters<typeof shown>[1] = {}) =>
		shown(fence, {
			solution_picture: fenceSolutionRow,
			solution_total: fenceTotal,
			...fields,
		});
	// order is where the trap, the picture of the solution and the steps stand
	// in the card, top to bottom.
	const order = () =>
		[...(root.querySelector(".mt-body")?.children ?? [])]
			.map((part) => part.className)
			.filter((name) => /trap|solution-picture|steps/.test(name));

	test("stands between the trap and the solution, the equality under it, and after a wrong option the option it was not", async () => {
		await drawCard(pictured());

		expect(order()).toEqual([
			"mt-note mt-note-trap",
			"mt-solution-picture",
			"mt-steps",
		]);
		const drawn = root.querySelector(".mt-solution-picture svg");
		expect(drawn?.getAttribute("role")).toBe("img");
		expect(drawn?.getAttribute("aria-label")).toBe("Things in a row");
		expect(text(".mt-total")).toBe("12 ÷ 3 + 1 = 5, not 4");
		expect(text(".mt-total-sum")).toBe(fenceTotal);
		expect(piecesOf(".mt-total")).toEqual(["mt-total-sum", ", not 4"]);
		expect(steps()).toHaveLength(3);
	});

	test('writes the equality alone under a right answer and under "I don\'t know"', async () => {
		await drawCard(pictured({ choice: "C", correct: true, trap: null }));
		expect(text(".mt-total")).toBe(fenceTotal);
		takeDown(root);

		await drawCard(pictured({ choice: "?", trap: null }));
		expect(text(".mt-total")).toBe(fenceTotal);
		expect(piecesOf(".mt-total")).toEqual(["mt-total-sum"]);
	});

	test("stands alone with no total, and is not drawn with no picture, or with one the card cannot draw", async () => {
		await drawCard(pictured({ solution_total: undefined }));
		expect(root.querySelector(".mt-solution-picture svg")).not.toBeNull();
		expect(root.querySelector(".mt-total")).toBeNull();
		takeDown(root);

		await drawCard(pictured({ solution_picture: undefined }));
		expect(root.querySelector(".mt-solution-picture")).toBeNull();
		takeDown(root);

		await drawCard(
			shown(fence, {
				solution_total: fenceTotal,
				// A kind the card does not know, as a later release may send.
				solution_picture: { kind: "pie" } as never,
			}),
		);
		expect(root.querySelector(".mt-solution-picture")).toBeNull();
		expect(steps()).toHaveLength(3);
	});

	test("writes under a picture of the solution that paints the key of its colours, in the task's language", async () => {
		await drawCard(
			shown(flagsTask, {
				choice: "C",
				correct_answer: "D",
				solution: flagsSolution,
				solution_picture: flagsSolutionPicture,
				solution_total: flagsTotal,
			}),
		);

		const key = root.querySelector(".mt-solution-picture ul.mt-pic-key");
		expect(key?.getAttribute("lang")).toBe("en");
		expect(key?.getAttribute("dir")).toBe("ltr");
		expect(
			[...(key?.querySelectorAll("li") ?? [])].map(
				(entry) => entry.textContent,
			),
		).toEqual(["Rred", "Yyellow", "Bblue"]);
		expect(text(".mt-total")).toBe("2 + 2 + 2 = 6, not 5");
	});

	test("reads the equality left to right in a card written right to left, and the words around it the card's way", async () => {
		await drawCard(
			shown(fenceInArabic, {
				solution: fenceSolutionInArabic,
				solution_picture: fenceSolutionRow,
				solution_total: fenceTotal,
			}),
		);

		expect(root.querySelector(".mt-total-sum")?.getAttribute("dir")).toBe(
			"ltr",
		);
		expect(text(".mt-total")).toBe("12 ÷ 3 + 1 = 5، وليس 4");
	});
});

describe("another task, from the card of how an answer went", () => {
	test("is asked of the chat, and the card, its ask taken, is done with", async () => {
		const heard = await drawCard(shown(fence, {}, {}));

		press(button("Another task"));

		await vi.waitFor(() => expect(heard.messages).toEqual(["Another task"]));
		await vi.waitFor(() =>
			expect(text(".mt-action-note")).toBe(
				"Once the ask reaches the chat, the new task will come below, in a new card.",
			),
		);
		expect(root.querySelector(".mt-btns")).toBeNull();
		expect(root.querySelector(".mt-topic-button")).toBeNull();
		// How the answer went stays on the card.
		expect(text(".mt-verdict-line")).toBe("Not quite — it's 5, not 4.");
		await vi.waitFor(() =>
			expect(document.activeElement).toBe(
				root.querySelector(".mt-action-note"),
			),
		);
		expect(heard.calls).toEqual([]);
	});

	test("that the chat did not take is said, and the card given back to ask again", async () => {
		vi.spyOn(console, "error").mockImplementation(() => {});
		const heard = await drawCard(shown(fence), { refuseMessages: 1 });

		press(button("Another task"));

		await vi.waitFor(() =>
			expect(text(".mt-action-note")).toBe("Not sent — try again"),
		);
		press(button("Another task"));
		await vi.waitFor(() =>
			expect(heard.messages).toEqual(["Another task", "Another task"]),
		);
	});
});

describe("the topic, chosen on the card of how an answer went", () => {
	// choice is the choice of the open panel named label.
	function choice(label: string): HTMLButtonElement {
		const found = [
			...root.querySelectorAll<HTMLButtonElement>(".mt-topic-option"),
		].find(
			(option) =>
				option.querySelector(".mt-topic-option-name")?.textContent === label,
		);
		if (found === undefined) {
			throw new Error(`the panel has no choice ${label}`);
		}
		return found;
	}

	// pending is the saving of a choice, which arrives when the test says.
	function pending() {
		let arrive: (result: CallToolResult) => void = () => {};
		const result = new Promise<CallToolResult>((resolve) => {
			arrive = resolve;
		});
		return { result, arrive };
	}

	test("is saved, told to the model, and then asked for, the card done with", async () => {
		const heard = await drawCard(shown(fence, {}, {}), {
			tools: () => topicSaved("percent.basic"),
		});

		press(button("Topic: Coach"));
		press(choice("Percentages"));

		await vi.waitFor(() => expect(heard.messages).toHaveLength(1));
		expect(heard.calls).toEqual([
			{ name: "edit_profile", arguments: { lesson_topic: "percent.basic" } },
		]);
		expect(heard.modelLines).toEqual([topicWords]);
		expect(heard.messages[0]).toContain("Percentages");
		await vi.waitFor(() => expect(root.querySelector(".mt-btns")).toBeNull());
	});

	test("locks the next task while it is saved", async () => {
		const saved = pending();
		const heard = await drawCard(shown(fence, {}, {}), {
			tools: () => saved.result,
		});

		press(button("Topic: Coach"));
		press(choice("Percentages"));
		press(button("Another task"));

		expect(button("Another task").getAttribute("aria-disabled")).toBe("true");
		saved.arrive(topicSaved("percent.basic"));
		await vi.waitFor(() => expect(heard.messages).toHaveLength(1));
		expect(heard.messages[0]).toContain("Percentages");
	});

	test("turns away the next task pressed in the moment a topic is chosen, before the card is drawn again", async () => {
		const saved = pending();
		const heard = await drawCard(shown(fence, {}, {}), {
			tools: () => saved.result,
		});

		press(button("Topic: Coach"));
		const percentages = choice("Percentages");
		const another = button("Another task");
		act(() => {
			percentages.click();
			another.click();
		});

		saved.arrive(topicSaved("percent.basic"));
		await vi.waitFor(() => expect(heard.messages).toHaveLength(1));
		expect(heard.messages[0]).toContain("Percentages");
	});
});

describe("a card of how an answer went with nothing to show", () => {
	test.each([
		["not_answered", "This task has no answer yet."],
		[
			"stale_task",
			"This task is no longer on the card, so how its answer went can't be shown.",
		],
		[
			"told_no_more",
			"This answer is recorded and counts, but how it went can no longer be shown.",
		],
	] as const)("for %s says why, and offers nothing", async (code, said) => {
		await drawCard(nothingShown(fence, code));

		expect(text(".mt-verdict-line")).toBe(said);
		expect(root.querySelector(".mt-btns")).toBeNull();
		expect(root.querySelector(".mt-steps")).toBeNull();
	});
});
