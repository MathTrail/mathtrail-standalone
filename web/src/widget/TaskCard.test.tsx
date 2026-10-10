import type { CallToolResult } from "@modelcontextprotocol/client";
import { render } from "preact";
import { act } from "preact/test-utils";
import { afterEach, describe, expect, test, vi } from "vitest";
import type { Host } from "./bridge";
import { OpensProgress } from "./CardFrame";
import { cardWords } from "./dictionaries";
import { ProgressOverCard } from "./ProgressOverCard";
import { readAnswer } from "./payload";
import { type Service, ServiceContext } from "./service";
import { TaskCard } from "./TaskCard";
import {
	buttonIn,
	drawCard as drawOnHost,
	foldIn,
	press,
	takeDown,
	unfold,
} from "./testing/card";
import type { ToolCall } from "./testing/host";
import {
	answered,
	dontKnowAnswer,
	editSaved,
	failure,
	fence,
	fenceInRussian,
	firstRun,
	flagsTask,
	type Handed,
	progress,
	progressMoving,
	repeatedAnswer,
	rightAnswer,
	staleAnswer,
	toldAgain,
} from "./testing/lesson";
import { WordsContext } from "./words";

let root: HTMLElement;

afterEach(() => {
	takeDown(root);
	vi.unstubAllEnvs();
});

// service answers the widget's calls as the service would for the fence: a
// wrong B recorded, and the progress read.
function service({ name }: ToolCall): CallToolResult {
	return name === "read_progress" ? progress : answered();
}

// drawCard draws the card a host hands payload to, the host answering the
// widget's tool calls with tools, and returns what the host hears.
async function drawCard(
	payload: Handed = fence,
	tools: (call: ToolCall) => CallToolResult | Promise<CallToolResult> = service,
	options: {
		refuseMessages?: boolean | number;
		refuseModelLines?: boolean;
		args?: Record<string, unknown>;
	} = {},
) {
	const drawn = await drawOnHost(payload, { tools, ...options });
	root = drawn.root;
	await vi.waitFor(() =>
		expect(root.querySelector(".mt-option")).not.toBeNull(),
	);
	return drawn.heard;
}

// pending is a tool's result that arrives when the test says.
function pending() {
	let arrive: (result: CallToolResult) => void = () => {};
	const result = new Promise<CallToolResult>((resolve) => {
		arrive = resolve;
	});
	return { result, arrive };
}

function option(letter: string): HTMLButtonElement {
	const found = [
		...root.querySelectorAll<HTMLButtonElement>(".mt-option"),
	].find(
		(row) => row.querySelector(".mt-option-letter")?.textContent === letter,
	);
	if (found === undefined) {
		throw new Error(`the card has no option ${letter}`);
	}
	return found;
}

const button = (label: string) => buttonIn(root, label);

// topLine is the line at the top of the task: the child's pseudonym and the
// way to the progress.
function topLine(): HTMLButtonElement {
	const found = root.querySelector<HTMLButtonElement>(
		".mt-bar:not(.mt-bar-back)",
	);
	if (found === null) {
		throw new Error("the card has no line at its top");
	}
	return found;
}

const text = (selector: string) => root.querySelector(selector)?.textContent;
// runSpans are how many steps each bracket under the progress's course spans.
const runSpans = () =>
	[...root.querySelectorAll(".mt-grades-runs td")].map((run) =>
		run.getAttribute("colspan"),
	);
// answerNote is what the card says of its answer, under the task.
const answerNote = () => text(".mt-answer-note");
const reviewComing =
	"Once the ask reaches the chat, how the answer went will come below, in a new card.";
const answeredBefore = "This task already has its answer.";
const shownButtons = () =>
	[...root.querySelectorAll<HTMLButtonElement>(".mt-btns .mt-btn")].map(
		(shown) => shown.textContent,
	);
const states = () =>
	["A", "B", "C", "D", "E"].map((letter) => option(letter).dataset.state);

describe("a task card", () => {
	test("shows the task as it was handed out, without its answer", async () => {
		await drawCard();

		expect(text(".mt-bar-name")).toBe("Comet");
		expect(text(".mt-bar-action")).toBe("Profile & progress");
		expect(text(".mt-badge")).toBe("Olympiad coach · Grade 3");
		expect(text(".mt-task-text")).toBe(fence.task.question);
		const picture = root.querySelector("svg.mt-picture");
		expect(picture?.getAttribute("role")).toBe("img");
		expect(picture?.getAttribute("aria-label")).toBe("Things in a row");
		expect(picture?.getAttribute("direction")).toBe("ltr");
		expect(
			[...root.querySelectorAll(".mt-option")].map((row) => row.textContent),
		).toEqual(["A 3", "B 4", "C 5", "D 6", "E 12"]);
		expect(states()).toEqual([
			"default",
			"default",
			"default",
			"default",
			"default",
		]);
		expect(text(".mt-options legend")).toBe("Pick one answer");
		expect(root.querySelector(".mt-note-hint")).toBeNull();
		expect(shownButtons()).toEqual(["Hint", "Another task"]);
	});

	test("names in its header the build the widget came with, by its number", async () => {
		vi.stubEnv("VITE_VERSION", "v0.2.1");
		await drawCard();

		expect(text(".mt-head .mt-version-label")).toBe("version");
		expect(text(".mt-head .mt-version-number")).toBe("0.2.1");
	});

	test("says the word for the version in the card's language", async () => {
		vi.stubEnv("VITE_VERSION", "v0.2.1");
		await drawCard({
			...fenceInRussian,
			child: { ...fenceInRussian.child, ui_language: "ru" },
		});

		expect(text(".mt-head .mt-version-label")).toBe("версия");
		expect(text(".mt-head .mt-version-number")).toBe("0.2.1");
	});

	test("names a month's weekdays in the language the task is written in", async () => {
		await drawCard({
			...fence,
			task: {
				...fence.task,
				language: "es",
				picture: { kind: "calendar", first: 1, days: 30 },
			},
		});

		const header = [...root.querySelectorAll("svg.mt-picture text")]
			.slice(0, 7)
			.map((name) => name.textContent);
		expect(header).toEqual(["lun", "mar", "mié", "jue", "vie", "sáb", "dom"]);
		expect(
			root.querySelector("svg.mt-picture")?.getAttribute("aria-label"),
		).toBe("A month's page");
	});

	test("writes under a picture that paints the key of its colours, in the language the task is written in", async () => {
		await drawCard({
			...flagsTask,
			task: {
				...flagsTask.task,
				language: "ru",
				picture: {
					kind: "flags",
					colors: { red: "красная", blue: "синяя" },
					groups: [{ flags: [["red", "blue"]] }],
				},
			},
		});

		const key = root.querySelector("ul.mt-pic-key");
		expect(key?.getAttribute("aria-label")).toBe("Colours");
		expect(key?.getAttribute("lang")).toBe("ru");
		expect(
			[...(key?.querySelectorAll("li") ?? [])].map(
				(entry) => entry.textContent,
			),
		).toEqual(["Ккрасная", "Ссиняя"]);
		expect(
			root.querySelector("svg.mt-picture")?.getAttribute("aria-label"),
		).toBe("Flags of stripes");
	});

	// What a service of another release, or a broken one, may hand out is no
	// task the card's types describe: it is handed as it would arrive.
	test.each<[string, Handed]>([
		["has none", { ...fence, task: { ...fence.task, picture: undefined } }],
		[
			"was written for a card of before pictures, as a text drawing",
			{
				...fence,
				task: { ...fence.task, picture: undefined, drawing: "|--3--|--3--|" },
			} as unknown as Handed,
		],
		[
			"the card cannot draw",
			{
				...fence,
				task: { ...fence.task, picture: { kind: "pie" } },
			} as unknown as Handed,
		],
	])(
		"draws no picture for a task that %s, and shows the task",
		async (_, handed) => {
			await drawCard(handed);

			expect(root.querySelector(".mt-picture")).toBeNull();
			expect(root.querySelector("pre")).toBeNull();
			expect(text(".mt-task-text")).toBe(fence.task.question);
		},
	);

	test("has no button that opens nothing", async () => {
		await drawCard();

		expect(root.querySelector("[aria-haspopup]")).toBeNull();
	});

	test("keeps markup in the task's words as text", async () => {
		const trick = "<img src=x onerror=alert(1)>";
		await drawCard(
			{
				...fence,
				child: { ...fence.child, pseudonym: trick },
				task: {
					...fence.task,
					question: trick,
					hint: trick,
					options: { ...fence.task.options, A: trick },
				},
			},
			() =>
				answered({
					trap: { id: "fence_gaps", text: trick, repeated: false },
					solution: trick,
				}),
		);
		press(button("Hint"));
		press(option("B"));

		await vi.waitFor(() => expect(answerNote()).toBe(reviewComing));
		expect(root.querySelector("img")).toBeNull();
		expect(root.textContent).toContain(trick);
	});

	test("never shows what the tool was called with", async () => {
		await drawCard(fence, service, {
			args: {
				task: { correct_answer: "C", solution: "the sealed solution" },
			},
		});

		expect(root.textContent).not.toContain("the sealed solution");
	});
});

describe("an answer", () => {
	test("is recorded with the letter pressed", async () => {
		const heard = await drawCard();

		press(option("B"));

		await vi.waitFor(() => expect(heard.calls).toHaveLength(1));
		expect(heard.calls[0]).toEqual({
			name: "submit_answer",
			arguments: { task_id: "task_fence", answer: "B", hint_used: false },
		});
	});

	test("is shown being checked, and is sent once however often it is pressed", async () => {
		const reply = pending();
		const heard = await drawCard(fence, () => reply.result);

		press(option("B"));
		press(option("B"));
		press(option("C"));

		await vi.waitFor(() => expect(heard.calls).toHaveLength(1));
		expect(option("B").dataset.state).toBe("selected");
		expect(option("B").textContent).toBe("B 4 Checking…");
		expect(option("C").getAttribute("aria-disabled")).toBe("true");
		expect(button("Hint").getAttribute("aria-disabled")).toBe("true");
		reply.arrive(answered());
		await vi.waitFor(() => expect(option("B").dataset.state).toBe("wrong"));
		expect(heard.calls).toHaveLength(1);
	});

	test("is sent once when two presses come before the card redraws", async () => {
		const heard = await drawCard();

		act(() => {
			option("B").click();
			option("C").click();
		});

		await vi.waitFor(() => expect(option("B").dataset.state).toBe("wrong"));
		expect(heard.calls).toHaveLength(1);
		expect(heard.calls[0]?.arguments.answer).toBe("B");
	});

	test("that is wrong is marked, told to the model, and gone over in the chat, below", async () => {
		const heard = await drawCard();

		press(option("B"));

		await vi.waitFor(() => expect(answerNote()).toBe(reviewComing));
		expect(states()).toEqual(["muted", "wrong", "correct", "muted", "muted"]);
		expect(option("B").textContent).toBe("B 4 Your answer");
		expect(option("C").textContent).toBe("C 5 Correct answer");
		expect(text(".mt-options legend")).toBe("Answers");
		// How the answer went is the card's of its own, below: this one shows
		// the task and the options marked, and nothing more to press.
		expect(root.querySelector(".mt-verdict-line")).toBeNull();
		expect(root.querySelector(".mt-steps")).toBeNull();
		expect(shownButtons()).toEqual([]);
		for (const letter of ["A", "B", "C", "D", "E"]) {
			expect(option(letter).getAttribute("aria-disabled")).toBe("true");
		}
		expect(heard.messages).toEqual(["Go over the answer"]);
		// The model reads the line with the message, so the line goes first.
		expect(heard.order).toEqual(["call", "model line", "message"]);
		expect(heard.modelLines).toEqual([
			"Task task_fence has its answer recorded: B, which is wrong; the right option is C. The card asks the chat, as the adult's message, to go over the answer: then call show_result with task_id task_fence, which draws below the card of how the answer went, with the trap and the solution step by step, and explain in two or three short sentences beside it. If show_result is not among your tools, this chat has an earlier list of MathTrail's tools: explain in words, and offer another task. Word it about the step, addressing nobody, in short sentences that fit the child's grade, so the adult can read it out as it is, and so it does not show whether the child is a boy or a girl: speak of the child by the pseudonym, never as he or she, praise the step, not the child, and keep to the present tense.",
		]);
	});

	test("that repeats a mistake made before asks the model for a reminder", async () => {
		const heard = await drawCard(fence, () => repeatedAnswer);

		press(option("B"));

		await vi.waitFor(() => expect(heard.modelLines).toHaveLength(1));
		expect(heard.modelLines[0]).toContain(
			"The child has made this mistake before among the latest answers: end your explanation with one short reminder of it",
		);
	});

	test("that is right is marked, and gone over in the chat as any other", async () => {
		const heard = await drawCard(fence, () => rightAnswer);

		press(option("C"));

		await vi.waitFor(() => expect(answerNote()).toBe(reviewComing));
		expect(states()).toEqual(["muted", "muted", "correct", "muted", "muted"]);
		expect(heard.messages).toEqual(["Go over the answer"]);
		expect(heard.modelLines[0]).toContain(
			"Task task_fence has its answer recorded: C, which is right.",
		);
	});

	test("keeps the focus on the option pressed", async () => {
		await drawCard();

		press(option("B"));

		await vi.waitFor(() => expect(answerNote()).toBe(reviewComing));
		expect(document.activeElement).toBe(option("B"));
	});

	test("that takes away the button holding the focus hands the focus to what the card says of the review", async () => {
		await drawCard();
		const hint = button("Hint");

		// A browser that gives a pressed option no focus leaves it where it
		// was, on a button that goes once the answer is in.
		act(() => {
			hint.focus();
			option("B").click();
		});

		await vi.waitFor(() => expect(answerNote()).toBe(reviewComing));
		expect(hint.isConnected).toBe(false);
		await vi.waitFor(() =>
			expect(document.activeElement).toBe(
				root.querySelector(".mt-answer-note"),
			),
		);
	});

	test.each<[string, CallToolResult, string[]]>([
		[
			'of "I don\'t know", said in the chat, to an option pressed after it',
			dontKnowAnswer,
			["muted", "muted", "correct", "muted", "muted"],
		],
		[
			"with D, given before, to an option pressed after it",
			toldAgain,
			["muted", "muted", "correct", "wrong", "muted"],
		],
	])(
		"recorded before is marked as it was recorded, %s, and offers to go over it rather than asking",
		async (_, recorded, marked) => {
			const heard = await drawCard(fence, () => recorded);

			press(option("B"));

			await vi.waitFor(() => expect(answerNote()).toBe(answeredBefore));
			expect(states()).toEqual(marked);
			await vi.waitFor(() => expect(heard.modelLines).toHaveLength(1));
			expect(heard.modelLines[0]).toContain(
				"The card asks nothing more of it: if the adult asks to go over the answer, call show_result with task_id task_fence",
			);
			expect(heard.messages).toEqual([]);

			press(button("Go over the answer"));

			await vi.waitFor(() => expect(answerNote()).toBe(reviewComing));
			expect(heard.messages).toEqual(["Go over the answer"]);
			expect(() => button("Go over the answer")).toThrow();
		},
	);

	test("to a task no longer being solved closes the task, and asks for no other", async () => {
		const heard = await drawCard(fence, () => staleAnswer);
		const hint = button("Hint");

		// A browser that gives a pressed option no focus leaves it on a button,
		// and the buttons go once the task is closed.
		act(() => {
			hint.focus();
			option("B").click();
		});

		await vi.waitFor(() =>
			expect(answerNote()).toBe(
				"This task is closed. The newest task is further down the chat.",
			),
		);
		expect(option("A").getAttribute("aria-disabled")).toBe("true");
		// Another task asked for here would skip the one on the newer card.
		expect(shownButtons()).toEqual([]);
		expect(hint.isConnected).toBe(false);
		expect(heard.messages).toEqual([]);
		await vi.waitFor(() =>
			expect(document.activeElement).toBe(
				root.querySelector(".mt-answer-note"),
			),
		);
	});

	test.each([
		["the service fails", () => failure],
		[
			"the call is lost",
			(): CallToolResult => {
				throw new Error("the host went away");
			},
		],
	])("that is not recorded because %s can be given again", async (_, fails) => {
		let calls = 0;
		const heard = await drawCard(fence, () =>
			calls++ === 0 ? fails() : answered({ choice: "C", correct: true }),
		);

		press(option("B"));

		await vi.waitFor(() =>
			expect(answerNote()).toBe("The answer couldn't be checked. Try again."),
		);
		expect(states()).toEqual([
			"default",
			"default",
			"default",
			"default",
			"default",
		]);
		expect(heard.messages).toEqual([]);
		press(option("C"));
		await vi.waitFor(() => expect(heard.calls).toHaveLength(2));
		await vi.waitFor(() => expect(option("C").dataset.state).toBe("correct"));
	});
});

describe("a card whose host lets it down", () => {
	afterEach(() => {
		vi.restoreAllMocks();
	});

	test("still marks the answer, and asks to go over it, when the model cannot be told of it", async () => {
		const logged = vi.spyOn(console, "error").mockImplementation(() => {});
		const heard = await drawCard(fence, service, { refuseModelLines: true });

		press(option("B"));

		await vi.waitFor(() => expect(heard.modelLines).toHaveLength(1));
		await vi.waitFor(() => expect(logged).toHaveBeenCalled());
		await vi.waitFor(() => expect(answerNote()).toBe(reviewComing));
		expect(option("B").dataset.state).toBe("wrong");
		expect(heard.messages).toEqual(["Go over the answer"]);
	});

	test("says the ask to go over the answer did not reach the chat, and offers it again", async () => {
		const logged = vi.spyOn(console, "error").mockImplementation(() => {});
		const heard = await drawCard(fence, service, { refuseMessages: 1 });

		press(option("B"));

		await vi.waitFor(() => expect(answerNote()).toBe("Not sent — try again"));
		expect(logged).toHaveBeenCalled();
		expect(heard.messages).toEqual(["Go over the answer"]);
		// The answer stays recorded and marked: only the ask is made again.
		expect(option("B").dataset.state).toBe("wrong");
		expect(shownButtons()).toEqual(["Go over the answer"]);

		press(button("Go over the answer"));

		await vi.waitFor(() => expect(answerNote()).toBe(reviewComing));
		expect(heard.messages).toEqual([
			"Go over the answer",
			"Go over the answer",
		]);
		expect(heard.calls).toHaveLength(1);
		expect(shownButtons()).toEqual([]);
	});

	test("says the ask for another task did not reach the chat, gives the card back, and takes it again", async () => {
		const logged = vi.spyOn(console, "error").mockImplementation(() => {});
		const heard = await drawCard(fence, service, { refuseMessages: 1 });

		press(button("Another task"));

		await vi.waitFor(() => expect(heard.messages).toEqual(["Another task"]));
		await vi.waitFor(() =>
			expect(text(".mt-action-note")).toBe("Not sent — try again"),
		);
		expect(logged).toHaveBeenCalled();
		// The chat did not take the ask, so the card is not done with: its
		// buttons are back, and its options can be pressed again.
		expect(button("Another task").getAttribute("aria-disabled")).toBeNull();
		expect(button("Hint").getAttribute("aria-disabled")).toBeNull();
		for (const letter of ["A", "B", "C", "D", "E"]) {
			expect(option(letter).getAttribute("aria-disabled")).toBeNull();
		}

		press(button("Another task"));

		await vi.waitFor(() =>
			expect(heard.messages).toEqual(["Another task", "Another task"]),
		);
		await vi.waitFor(() =>
			expect(text(".mt-action-note")).toBe(
				"Once the ask reaches the chat, the new task will come below, in a new card.",
			),
		);
		expect(root.querySelector(".mt-btns")).toBeNull();
		expect(option("A").getAttribute("aria-disabled")).toBe("true");
	});

	test("says the progress did not load when the call is lost", async () => {
		vi.spyOn(console, "error").mockImplementation(() => {});
		await drawCard(fence, ({ name }) => {
			if (name === "read_progress") {
				throw new Error("the host went away");
			}
			return answered();
		});

		press(topLine());

		await vi.waitFor(() =>
			expect(text(".mt-progress")).toBe(
				"The progress didn't load. Go back and try again.",
			),
		);
	});
});

describe("a task in a language the card has no words for", () => {
	// The card speaks its lesson's language when it has words for it, and the
	// host's otherwise: Hebrew is one it has none for, and is written right to
	// left.
	test("is marked as written in it, and runs its way", async () => {
		await drawCard(
			{
				...fence,
				task: { ...fence.task, language: "he", question: "כמה עמודים?" },
				language: "he",
			},
			() => answered(),
		);
		press(button("Hint"));

		for (const said of [
			".mt-task-text",
			".mt-option-value",
			".mt-note-hint p",
		]) {
			const words = root.querySelector(said);
			expect(words?.getAttribute("lang")).toBe("he");
			expect(words?.getAttribute("dir")).toBe("rtl");
		}
		// The card's own words stay in the card's language.
		expect(root.querySelector(".mt-options legend")?.hasAttribute("lang")).toBe(
			false,
		);
	});
});

describe("the hint", () => {
	test("is shown and hidden, and its use goes with the answer", async () => {
		const heard = await drawCard();

		press(button("Hint"));

		expect(text(".mt-note-hint p")).toBe(fence.task.hint);
		expect(button("Hide hint").getAttribute("aria-expanded")).toBe("true");
		press(button("Hide hint"));
		expect(root.querySelector(".mt-note-hint")).toBeNull();
		expect(button("Hint").getAttribute("aria-expanded")).toBe("false");

		press(option("B"));

		await vi.waitFor(() => expect(heard.calls).toHaveLength(1));
		expect(heard.calls[0]?.arguments.hint_used).toBe(true);
	});

	test("goes once the answer is in", async () => {
		await drawCard();
		press(button("Hint"));

		press(option("B"));

		await vi.waitFor(() => expect(option("B").dataset.state).toBe("wrong"));
		expect(root.querySelector(".mt-note-hint")).toBeNull();
	});
});

describe("another task", () => {
	test("is asked of the chat, and the card, its ask taken, keeps its task, locks it and asks nothing more", async () => {
		const heard = await drawCard();

		press(button("Another task"));

		await vi.waitFor(() => expect(heard.messages).toEqual(["Another task"]));
		await vi.waitFor(() =>
			expect(text(".mt-action-note")).toBe(
				"Once the ask reaches the chat, the new task will come below, in a new card.",
			),
		);
		// The card no longer turns to a wait it cannot see the end of: the
		// task stays, and nothing claims to be under way.
		expect(root.querySelector(".mt-gen")).toBeNull();
		expect(text(".mt-task-text")).toBe(fence.task.question);
		expect(root.textContent).not.toContain("Sent to the chat");
		// The next task comes below, on a card of its own: this one is done
		// with, its options locked and its buttons gone, and what it says of
		// the task to come holds the focus the pressed button had.
		expect(root.querySelector(".mt-btns")).toBeNull();
		for (const letter of ["A", "B", "C", "D", "E"]) {
			expect(option(letter).getAttribute("aria-disabled")).toBe("true");
		}
		expect(text(".mt-options legend")).toBe("Answers");
		const note = root.querySelector(".mt-action-note");
		expect(note?.getAttribute("tabindex")).toBe("-1");
		await vi.waitFor(() => expect(document.activeElement).toBe(note));
		expect(heard.calls).toEqual([]);
	});

	test("is asked for once while the ask is on its way, and never again once the chat has it", async () => {
		const heard = await drawCard();
		const another = button("Another task");

		act(() => {
			another.click();
			another.click();
		});
		await vi.waitFor(() =>
			expect(text(".mt-action-note")).toBe(
				"Once the ask reaches the chat, the new task will come below, in a new card.",
			),
		);
		expect(heard.messages).toEqual(["Another task"]);
		expect(() => button("Another task")).toThrow();
	});

	test("turns away an answer pressed in the same moment after it, the options locked while the ask is on its way", async () => {
		const heard = await drawCard();

		act(() => {
			button("Another task").click();
			option("B").click();
		});
		expect(option("B").getAttribute("aria-disabled")).toBe("true");

		await vi.waitFor(() => expect(heard.messages).toEqual(["Another task"]));
		expect(heard.calls).toEqual([]);
		expect(states()).toEqual([
			"default",
			"default",
			"default",
			"default",
			"default",
		]);
	});

	test("is not asked for while an answer pressed in the same moment is on its way", async () => {
		const heard = await drawCard();

		act(() => {
			option("B").click();
			button("Another task").click();
		});

		await vi.waitFor(() => expect(answerNote()).toBe(reviewComing));
		expect(heard.messages).toEqual(["Go over the answer"]);
		expect(root.querySelector(".mt-gen")).toBeNull();
		expect(option("B").dataset.state).toBe("wrong");
	});

	test("keeps the top line to the progress, and the card as it was", async () => {
		const heard = await drawCard();
		press(button("Another task"));
		await vi.waitFor(() => expect(heard.messages).toEqual(["Another task"]));

		press(topLine());

		await vi.waitFor(() => expect(text(".mt-rank-name")).toBe("Rank 3"));
		press(button("Back to task"));
		expect(text(".mt-task-text")).toBe(fence.task.question);
		expect(text(".mt-action-note")).toBe(
			"Once the ask reaches the chat, the new task will come below, in a new card.",
		);
		expect(root.querySelector(".mt-btns")).toBeNull();
		expect(heard.calls.map((call) => call.name)).toEqual(["read_progress"]);
	});

	test("is asked for in the card's language", async () => {
		const heard = await drawCard({
			...fenceInRussian,
			child: { ...fenceInRussian.child, ui_language: "ru" },
		});

		press(button("Другая задача"));

		await vi.waitFor(() => expect(heard.messages).toEqual(["Другая задача"]));
	});

	test("is not offered once the answer is in: the card of how it went offers it", async () => {
		await drawCard();

		press(option("B"));

		await vi.waitFor(() => expect(answerNote()).toBe(reviewComing));
		expect(() => button("Another task")).toThrow();
		expect(root.querySelector(".mt-topic-button")).toBeNull();
	});
});

describe("the progress", () => {
	test("is read in the card, and the task comes back as it was left", async () => {
		const heard = await drawCard();
		press(button("Hint"));
		press(option("B"));
		await vi.waitFor(() => expect(answerNote()).toBe(reviewComing));
		const task = root.querySelector(".mt-widget > div")?.innerHTML;

		press(topLine());

		await vi.waitFor(() => expect(text(".mt-rank-name")).toBe("Rank 3"));
		expect(document.activeElement).toBe(button("Back to task"));
		expect(root.querySelector(".mt-widget > div")?.hasAttribute("hidden")).toBe(
			true,
		);
		press(button("Back to task"));

		expect(root.querySelector(".mt-widget > div")?.innerHTML).toBe(task);
		expect(document.activeElement).toBe(topLine());
		expect(heard.calls.map((call) => call.name)).toEqual([
			"submit_answer",
			"read_progress",
		]);
		expect(heard.calls[1]?.arguments).toEqual({});
	});

	test("shows the outline of the screen under the child's line while it is read, and the screen in its place once it is", async () => {
		const read = pending();
		await drawCard(fence, ({ name }) =>
			name === "read_progress" ? read.result : answered(),
		);

		press(topLine());

		const over = root.querySelector(
			'article[aria-label="Profile and progress"]',
		);
		expect(
			[...(over?.children ?? [])].map((part) => part.classList[0]),
		).toEqual(["mt-loading-bar", "mt-head", "mt-skeleton"]);
		expect(over?.querySelector(".mt-head .mt-name")?.textContent).toBe(
			fence.child.pseudonym,
		);
		expect(over?.querySelector(".mt-skeleton > output")?.textContent).toBe(
			"Loading the progress…",
		);
		expect(root.querySelector(".mt-rank-name")).toBeNull();
		const outlined = runSpans();
		expect(outlined).toHaveLength(3);

		read.arrive(progress);

		await vi.waitFor(() => expect(text(".mt-rank-name")).toBe("Rank 3"));
		// The brackets under the course stand where their outlines stood.
		expect(runSpans()).toEqual(outlined);
		expect(root.querySelector(".mt-loading-bar")).toBeNull();
		expect(root.querySelector(".mt-skeleton")).toBeNull();
	});

	test("keeps the while chosen for the moves when it is opened again", async () => {
		await drawCard(fence, ({ name }) =>
			name === "read_progress" ? progressMoving : answered(),
		);
		press(topLine());
		await vi.waitFor(() => expect(period("last_task")).not.toBeNull());
		press(period("last_task") as HTMLElement);
		press(button("Back to task"));

		press(topLine());
		await vi.waitFor(() => expect(period("last_task")).not.toBeNull());

		expect(period("last_task")?.checked).toBe(true);
		expect(period("week")?.checked).toBe(false);
	});

	test("is read afresh each time it is opened, its sections open as they were left", async () => {
		const heard = await drawCard();
		press(topLine());
		await vi.waitFor(() => expect(text(".mt-rank-name")).toBe("Rank 3"));
		expect(foldIn(root, "Topics").getAttribute("aria-expanded")).toBe("true");
		expect(foldIn(root, "Profile").getAttribute("aria-expanded")).toBe("false");
		press(foldIn(root, "Topics"));
		press(foldIn(root, "Profile"));
		press(button("Back to task"));

		press(topLine());
		await vi.waitFor(() => expect(text(".mt-rank-name")).toBe("Rank 3"));

		expect(foldIn(root, "Topics").getAttribute("aria-expanded")).toBe("false");
		expect(foldIn(root, "Profile").getAttribute("aria-expanded")).toBe("true");
		expect(heard.calls.map((call) => call.name)).toEqual([
			"read_progress",
			"read_progress",
		]);
	});

	test("keeps a change saved on its form, and the card shows the child as the profile now says", async () => {
		const heard = await drawCard(fence, ({ name }) =>
			name === "read_progress"
				? progress
				: editSaved({ pseudonym: "Star", grade: 4 }),
		);
		press(topLine());
		await vi.waitFor(() => expect(text(".mt-rank-name")).toBe("Rank 3"));
		unfold(root, "Profile");
		press(button("Edit"));
		const box = root.querySelector<HTMLInputElement>(".mt-form .mt-input");
		act(() => {
			if (box !== null) {
				box.value = "Star";
				box.dispatchEvent(new Event("input", { bubbles: true }));
			}
		});
		press(button("Save"));
		await vi.waitFor(() =>
			expect(heard.calls.map((call) => call.name)).toContain("edit_profile"),
		);
		await vi.waitFor(() => expect(root.querySelector(".mt-form")).toBeNull());

		press(button("Back to task"));

		expect(text(".mt-bar-name")).toBe("Star");
		expect(text(".mt-head .mt-badge")).toContain("4");
		expect(root.querySelector(".mt-task-text")?.textContent).toBe(
			fence.task.question,
		);
	});

	test("that does not arrive says so, in place of its outline", async () => {
		await drawCard(fence, () => failure);

		press(topLine());

		await vi.waitFor(() =>
			expect(text(".mt-progress")).toBe(
				"The progress didn't load. Go back and try again.",
			),
		);
		expect(root.querySelector(".mt-loading-bar")).toBeNull();
		expect(root.querySelector(".mt-skeleton")).toBeNull();
	});

	test("that does not read says it did not load", async () => {
		await drawCard(fence, () => ({
			content: [],
			structuredContent: { screen: "progress", profile: null },
		}));

		press(topLine());

		await vi.waitFor(() =>
			expect(text(".mt-progress")).toBe(
				"The progress didn't load. Go back and try again.",
			),
		);
	});

	test("that finds the profile gone shows the first sign-in, and the way back", async () => {
		await drawCard(fence, () => ({ content: [], structuredContent: firstRun }));

		press(topLine());

		await vi.waitFor(() =>
			expect(root.querySelector(".mt-lead")).not.toBeNull(),
		);
		expect(button("Back to task")).toBeDefined();
		press(button("Back to task"));
		expect(root.querySelector(".mt-lead")).toBeNull();
		expect(root.querySelector(".mt-option")).not.toBeNull();
	});
});

describe("what the card says of its answer", () => {
	test("is listened to before it says anything", async () => {
		await drawCard();

		const note = root.querySelector(".mt-answer-note");
		expect(note?.getAttribute("aria-live")).toBe("polite");
		expect(note?.textContent).toBe("");
	});
});

describe("a lesson begun with its answer in", () => {
	test("marks the answer as one recorded on the card is marked, and asks the host nothing", () => {
		const recorded = readAnswer(answered(), fence.task.id);
		if (recorded.kind !== "answered") {
			throw new Error("the example is no answer recorded");
		}
		const asked: string[] = [];
		const refuse = (what: string) => {
			asked.push(what);
			return Promise.reject(new Error(`${what} is not asked here`));
		};
		const host: Host = {
			callTool: (name) => refuse(name),
			sendMessage: () => refuse("a message"),
			tellModel: () => refuse("a line for the model"),
			canOpenLinks: () => false,
			openLink: () => refuse("a page"),
		};
		root = document.createElement("div");
		document.body.append(root);

		act(() =>
			render(
				<WordsContext.Provider value={cardWords("en", undefined)}>
					<TaskCard
						handed={fence}
						host={host}
						start={{
							hint: { open: false, used: false },
							answer: { state: "answered", result: recorded.result },
						}}
					/>
				</WordsContext.Provider>,
				root,
			),
		);

		expect(states()).toEqual(["muted", "wrong", "correct", "muted", "muted"]);
		expect(text(".mt-options legend")).toBe("Answers");
		expect(shownButtons()).toEqual([]);
		expect(answerNote()).toBe("");
		expect(asked).toEqual([]);
	});
});

describe("a card drawn where a page answers for the service", () => {
	// A page that shows a card live answers for the service itself: the card
	// asks the service it is given, and never calls a tool through a host.
	test("records an answer through that service, and calls no tool", async () => {
		const recorded = readAnswer(answered(), fence.task.id);
		if (recorded.kind !== "answered") {
			throw new Error("the example is no answer recorded");
		}
		const tools: string[] = [];
		const host: Host = {
			callTool: (name) => {
				tools.push(name);
				return Promise.reject(new Error("no tool is called on this page"));
			},
			sendMessage: () => Promise.resolve(),
			tellModel: () => Promise.resolve(),
			canOpenLinks: () => false,
			openLink: () => Promise.resolve(false),
		};
		const answers: unknown[][] = [];
		const page: Service = {
			recordAnswer: (...given) => {
				answers.push(given);
				return Promise.resolve(recorded);
			},
			taskStatus: () => Promise.resolve({ kind: "unknown" }),
			readProgress: () => Promise.resolve({ kind: "failed" }),
			saveEdit: () => Promise.resolve({ kind: "failed" }),
		};
		root = document.createElement("div");
		document.body.append(root);
		act(() =>
			render(
				<WordsContext.Provider value={cardWords("en", undefined)}>
					<ServiceContext.Provider value={page}>
						<TaskCard handed={fence} host={host} />
					</ServiceContext.Provider>
				</WordsContext.Provider>,
				root,
			),
		);

		press(option("B"));

		await vi.waitFor(() => expect(option("B").dataset.state).toBe("wrong"));
		expect(answers).toEqual([["task_fence", "B", false]]);
		expect(tools).toEqual([]);
	});

	test("reads the progress it opens through that service, and calls no tool", async () => {
		const tools: string[] = [];
		const host: Host = {
			callTool: (name) => {
				tools.push(name);
				return Promise.reject(new Error("no tool is called on this page"));
			},
			sendMessage: () => Promise.resolve(),
			tellModel: () => Promise.resolve(),
			canOpenLinks: () => false,
			openLink: () => Promise.resolve(false),
		};
		const page: Service = {
			recordAnswer: () => Promise.resolve({ kind: "failed" }),
			taskStatus: () => Promise.resolve({ kind: "unknown" }),
			readProgress: () =>
				Promise.resolve({ kind: "read", payload: progress.structuredContent }),
			saveEdit: () => Promise.resolve({ kind: "failed" }),
		};
		root = document.createElement("div");
		document.body.append(root);
		act(() =>
			render(
				<WordsContext.Provider value={cardWords("en", undefined)}>
					<ServiceContext.Provider value={page}>
						<OpensProgress.Provider value={ProgressOverCard}>
							<TaskCard handed={fence} host={host} />
						</OpensProgress.Provider>
					</ServiceContext.Provider>
				</WordsContext.Provider>,
				root,
			),
		);

		press(topLine());

		await vi.waitFor(() =>
			expect(root.querySelector(".mt-bar-back")).not.toBeNull(),
		);
		await vi.waitFor(() => expect(text(".mt-rank-name")).toBe("Rank 3"));
		expect(tools).toEqual([]);
	});
});

// period is the while named on the switch of the progress open over the card.
function period(value: string): HTMLInputElement | null {
	return root.querySelector<HTMLInputElement>(
		`.mt-switch input[value="${value}"]`,
	);
}
