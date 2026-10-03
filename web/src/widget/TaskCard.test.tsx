import type { CallToolResult } from "@modelcontextprotocol/client";
import { act } from "preact/test-utils";
import { afterEach, describe, expect, test, vi } from "vitest";
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
	type Handed,
	progress,
	repeatedAnswer,
	rightAnswer,
	staleAnswer,
	toldAgain,
	trialAnswer,
} from "./testing/lesson";

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
		refuseMessages?: boolean;
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
const replies = () => [...root.querySelectorAll(".mt-replies .mt-reply")];
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
		const drawing = root.querySelector("pre.mt-diagram");
		expect(drawing?.getAttribute("dir")).toBe("ltr");
		expect(drawing?.textContent).toBe("|--3--|--3--|--3--|--3--|");
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
		expect(shownButtons()).toEqual(["I don't know", "Hint", "Another task"]);
	});

	test("names in its header the build the widget came with", async () => {
		vi.stubEnv("VITE_VERSION", "v0.2.1");
		await drawCard();

		expect(text(".mt-head .mt-version")).toBe("v0.2.1");
	});

	test("draws no drawing for a task that has none", async () => {
		await drawCard({ ...fence, task: { ...fence.task, drawing: "" } });

		expect(root.querySelector(".mt-diagram")).toBeNull();
	});

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
					drawing: trick,
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

		await vi.waitFor(() => expect(replies()).toHaveLength(1));
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

	test('of "I don\'t know" is recorded as ?', async () => {
		const heard = await drawCard(fence, () => dontKnowAnswer);

		press(button("I don't know"));

		await vi.waitFor(() => expect(heard.calls).toHaveLength(1));
		expect(heard.calls[0]?.arguments).toEqual({
			task_id: "task_fence",
			answer: "?",
			hint_used: false,
		});
	});

	test("is shown being checked, and is sent once however often it is pressed", async () => {
		const reply = pending();
		const heard = await drawCard(fence, () => reply.result);

		press(option("B"));
		press(option("B"));
		press(option("C"));
		press(button("I don't know"));

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
			button("I don't know").click();
		});

		await vi.waitFor(() => expect(replies()).toHaveLength(1));
		expect(heard.calls).toHaveLength(1);
		expect(heard.calls[0]?.arguments.answer).toBe("B");
	});

	test("that is wrong is told from its trap, step by step, with the rating", async () => {
		const heard = await drawCard();

		press(option("B"));

		await vi.waitFor(() => expect(replies()).toHaveLength(1));
		expect(states()).toEqual(["muted", "wrong", "correct", "muted", "muted"]);
		expect(option("B").textContent).toBe("B 4 Your answer");
		expect(option("C").textContent).toBe("C 5 Correct answer");
		expect(text(".mt-options legend")).toBe("Answers");
		expect(text(".mt-verdict-line")).toBe("Not quite — it's 5, not 4.");
		expect(root.querySelector(".mt-verdict-line svg")).not.toBeNull();
		expect(text(".mt-note-trap p")).toBe(
			"Counted the gaps instead of the posts.",
		);
		expect(root.querySelector(".mt-note-detail")).toBeNull();
		expect(
			[...root.querySelectorAll(".mt-steps li")].map(
				(step) => step.textContent,
			),
		).toEqual([
			"1 12 ÷ 3 = 4 gaps.",
			"2 A straight fence with posts at both ends has one more post than gaps.",
			"3 4 + 1 = 5 posts.",
		]);
		expect(text(".mt-note-plain")).toBe("Rating in this topic1502 → 1480");
		expect(shownButtons()).toEqual(["Another task"]);
		expect(root.querySelector(".mt-btn-primary")?.textContent).toBe(
			"Another task",
		);
		await vi.waitFor(() => expect(heard.modelLines).toHaveLength(1));
		expect(heard.modelLines[0]).toBe(
			"Task task_fence has its answer recorded: B, which is wrong; the right option is C. The card shows the trap and the solution. In a language with grammatical gender, word it so it does not show whether the child is a boy or a girl: praise the step, not the child, and keep to the present tense.",
		);
	});

	test("that repeats a mistake made before says so under its trap, and asks the model for a reminder", async () => {
		const heard = await drawCard(fence, () => repeatedAnswer);

		press(option("B"));

		await vi.waitFor(() => expect(replies()).toHaveLength(1));
		expect(
			[...root.querySelectorAll(".mt-note-trap p")].map(
				(said) => said.textContent,
			),
		).toEqual([
			"Counted the gaps instead of the posts.",
			"This mistake has come up before.",
		]);
		await vi.waitFor(() => expect(heard.modelLines).toHaveLength(1));
		expect(heard.modelLines[0]).toContain(
			"The child has made this mistake before among the latest answers: end your explanation with one short reminder of it",
		);
	});

	test("that is right is praised, with no trap", async () => {
		await drawCard(fence, () => rightAnswer);

		press(option("C"));

		await vi.waitFor(() => expect(replies()).toHaveLength(1));
		expect(states()).toEqual(["muted", "muted", "correct", "muted", "muted"]);
		expect(text(".mt-verdict-line")).toBe("Correct! It's 5.");
		expect(root.querySelector(".mt-note-trap")).toBeNull();
		expect(text(".mt-note-plain p")).toBe("1502 → 1519");
	});

	test('of "I don\'t know" shows the solution with no verdict, and the rating it cost', async () => {
		await drawCard(fence, () => dontKnowAnswer);
		const idk = button("I don't know");

		press(idk);

		await vi.waitFor(() => expect(replies()).toHaveLength(1));
		expect(states()).toEqual(["muted", "muted", "correct", "muted", "muted"]);
		expect(text(".mt-verdict-line")).toBe("Here's how to solve it.");
		expect(root.querySelector(".mt-verdict-line svg")).toBeNull();
		expect(root.querySelectorAll(".mt-steps li")).toHaveLength(3);
		expect(text(".mt-note-plain p")).toBe("1502 → 1488");
		// The button pressed is gone: the focus goes to the one thing left to
		// do, while the replies read the result out.
		expect(idk.isConnected).toBe(false);
		expect(document.activeElement).toBe(button("Another task"));
	});

	test("keeps the focus on the option pressed", async () => {
		await drawCard();

		press(option("B"));

		await vi.waitFor(() => expect(replies()).toHaveLength(1));
		expect(document.activeElement).toBe(option("B"));
	});

	test("in the trial series shows how far it has got instead of a rating", async () => {
		await drawCard(fence, () => trialAnswer);

		press(option("B"));

		await vi.waitFor(() => expect(replies()).toHaveLength(1));
		expect(text(".mt-note-plain")).toBe("Trial series3 of 5");
	});

	test("recorded before is shown, and told to the model, as it was recorded", async () => {
		const heard = await drawCard(fence, () => toldAgain);

		press(option("B"));

		await vi.waitFor(() => expect(replies()).toHaveLength(1));
		expect(states()).toEqual(["muted", "muted", "correct", "wrong", "muted"]);
		expect(text(".mt-reply-body > p")).toBe(
			"This task was answered before — here is that answer.",
		);
		expect(text(".mt-verdict-line")).toBe("Not quite — it's 5, not 6.");
		await vi.waitFor(() => expect(heard.modelLines).toHaveLength(1));
		expect(heard.modelLines[0]).toBe(
			"Task task_fence has its answer recorded: D, which is wrong; the right option is C. The card shows the trap and the solution. In a language with grammatical gender, word it so it does not show whether the child is a boy or a girl: praise the step, not the child, and keep to the present tense.",
		);
	});

	test("to a task no longer being solved closes the task, and asks for no other", async () => {
		await drawCard(fence, () => staleAnswer);
		const idk = button("I don't know");

		press(idk);

		await vi.waitFor(() => expect(replies()).toHaveLength(1));
		expect(text(".mt-verdict-line")).toBe(
			"This task is closed. The newest task is further down the chat.",
		);
		expect(option("A").getAttribute("aria-disabled")).toBe("true");
		// Another task asked for here would skip the one on the newer card.
		expect(shownButtons()).toEqual([]);
		expect(document.activeElement).toBe(
			root.querySelector(".mt-replies > [tabindex='-1']"),
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
			expect(text(".mt-verdict-line")).toBe(
				"The answer couldn't be checked. Try again.",
			),
		);
		expect(states()).toEqual([
			"default",
			"default",
			"default",
			"default",
			"default",
		]);
		press(option("C"));
		await vi.waitFor(() => expect(heard.calls).toHaveLength(2));
		await vi.waitFor(() => expect(option("C").dataset.state).toBe("correct"));
	});
});

describe("a card whose host lets it down", () => {
	afterEach(() => {
		vi.restoreAllMocks();
	});

	test("still shows the result when the model cannot be told of it", async () => {
		const logged = vi.spyOn(console, "error").mockImplementation(() => {});
		const heard = await drawCard(fence, service, { refuseModelLines: true });

		press(option("B"));

		await vi.waitFor(() => expect(heard.modelLines).toHaveLength(1));
		await vi.waitFor(() => expect(logged).toHaveBeenCalled());
		expect(text(".mt-verdict-line")).toBe("Not quite — it's 5, not 4.");
	});

	test("says the ask for another task did not reach the chat, and takes it again", async () => {
		const logged = vi.spyOn(console, "error").mockImplementation(() => {});
		const heard = await drawCard(fence, service, { refuseMessages: true });

		press(button("Another task"));

		await vi.waitFor(() => expect(heard.messages).toEqual(["Another task"]));
		await vi.waitFor(() =>
			expect(text(".mt-action-note")).toBe("Not sent — try again"),
		);
		expect(logged).toHaveBeenCalled();
		expect(button("Another task").getAttribute("aria-disabled")).toBeNull();
		expect(root.querySelector(".mt-option")).not.toBeNull();

		press(button("Another task"));

		await vi.waitFor(() =>
			expect(heard.messages).toEqual(["Another task", "Another task"]),
		);
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

		press(option("B"));

		await vi.waitFor(() => expect(replies()).toHaveLength(1));
		expect(root.querySelector(".mt-note-trap p")?.getAttribute("lang")).toBe(
			"he",
		);
		expect(root.querySelector(".mt-steps ol")?.getAttribute("dir")).toBe("rtl");
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

		await vi.waitFor(() => expect(replies()).toHaveLength(1));
		expect(root.querySelector(".mt-note-hint")).toBeNull();
	});
});

describe("another task", () => {
	test("is asked of the chat, while the card keeps its task and says where the new one will come", async () => {
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
		expect(root.querySelector(".mt-option")).not.toBeNull();
		expect(root.textContent).not.toContain("Sent to the chat");
		// The button pressed keeps the focus, and once the chat has the ask it
		// may be pressed again: a host may hold the message for the person to
		// send, and the card cannot see whether it went.
		expect(document.activeElement).toBe(button("Another task"));
		expect(button("Another task").getAttribute("aria-disabled")).toBeNull();
		expect(heard.calls).toEqual([]);
	});

	test("is asked for once while the ask is on its way, and again once the chat has it", async () => {
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

		press(button("Another task"));

		await vi.waitFor(() =>
			expect(heard.messages).toEqual(["Another task", "Another task"]),
		);
	});

	test("is not asked for while an answer pressed in the same moment is on its way", async () => {
		const heard = await drawCard();

		act(() => {
			option("B").click();
			button("Another task").click();
		});

		await vi.waitFor(() => expect(replies()).toHaveLength(1));
		expect(heard.messages).toEqual([]);
		expect(root.querySelector(".mt-gen")).toBeNull();
		expect(text(".mt-verdict-line")).toBe("Not quite — it's 5, not 4.");
	});

	test("keeps the top line to the progress, and the card as it was", async () => {
		const heard = await drawCard();
		press(button("Another task"));
		await vi.waitFor(() => expect(heard.messages).toEqual(["Another task"]));

		press(topLine());

		await vi.waitFor(() =>
			expect(text(".mt-rank-name")).toBe("River crossing"),
		);
		press(button("Back to task"));
		expect(text(".mt-task-text")).toBe(fence.task.question);
		expect(text(".mt-action-note")).toBe(
			"Once the ask reaches the chat, the new task will come below, in a new card.",
		);
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

	test("can be asked for once the answer is in", async () => {
		const heard = await drawCard();
		press(option("B"));
		await vi.waitFor(() => expect(replies()).toHaveLength(1));

		press(button("Another task"));

		await vi.waitFor(() => expect(heard.messages).toEqual(["Another task"]));
		await vi.waitFor(() =>
			expect(text(".mt-action-note")).toBe(
				"Once the ask reaches the chat, the new task will come below, in a new card.",
			),
		);
		expect(text(".mt-verdict-line")).toBe("Not quite — it's 5, not 4.");
	});
});

describe("the progress", () => {
	test("is read in the card, and the task comes back as it was left", async () => {
		const heard = await drawCard();
		press(button("Hint"));
		press(option("B"));
		await vi.waitFor(() => expect(replies()).toHaveLength(1));
		const task = root.querySelector(".mt-widget > div")?.innerHTML;

		press(topLine());

		await vi.waitFor(() =>
			expect(text(".mt-rank-name")).toBe("River crossing"),
		);
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

	test("is read afresh each time it is opened, its sections open as they were left", async () => {
		const heard = await drawCard();
		press(topLine());
		await vi.waitFor(() =>
			expect(text(".mt-rank-name")).toBe("River crossing"),
		);
		expect(foldIn(root, "Topics").getAttribute("aria-expanded")).toBe("false");
		press(foldIn(root, "Topics"));
		press(button("Back to task"));

		press(topLine());
		await vi.waitFor(() =>
			expect(text(".mt-rank-name")).toBe("River crossing"),
		);

		expect(foldIn(root, "Topics").getAttribute("aria-expanded")).toBe("true");
		expect(foldIn(root, "Profile").getAttribute("aria-expanded")).toBe("false");
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
		await vi.waitFor(() =>
			expect(text(".mt-rank-name")).toBe("River crossing"),
		);
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

	test("that does not arrive says so", async () => {
		await drawCard(fence, () => failure);

		press(topLine());

		await vi.waitFor(() =>
			expect(text(".mt-progress")).toBe(
				"The progress didn't load. Go back and try again.",
			),
		);
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

describe("the replies", () => {
	test("are listened to before the first one arrives", async () => {
		await drawCard();

		const list = root.querySelector(".mt-replies");
		expect(list?.getAttribute("aria-live")).toBe("polite");
		expect(list?.children).toHaveLength(0);
	});
});
