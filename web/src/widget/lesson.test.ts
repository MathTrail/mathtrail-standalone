import type { CallToolResult } from "@modelcontextprotocol/client";
import { describe, expect, test } from "vitest";
import { letters } from "./choices";
import {
	type Answer,
	canAnswer,
	isSettled,
	type Lesson,
	type LessonEvent,
	lessonStart,
	modelLineOf,
	next,
	optionStateOf,
} from "./lesson";
import type { AnswerResult, ResultShown } from "./payload";
import {
	answered,
	dontKnowAnswer,
	repeatedAnswer,
	rightAnswer,
} from "./testing/lesson";

// resultOf is the answer's result a tool's result carries.
function resultOf(tool: CallToolResult): AnswerResult {
	return (tool.structuredContent as { result: AnswerResult }).result;
}

const wrong = resultOf(answered());
// wrongShown is how the wrong answer went, as the service tells it whole.
const wrongShown = answered().structuredContent as ResultShown;

// lessonAfter is the lesson after events, from a card as a task arrives on it.
function lessonAfter(...events: LessonEvent[]): Lesson {
	return events.reduce(next, lessonStart);
}

describe("an answer", () => {
	test("is checked once picked, and settles as the service says", () => {
		const checking = lessonAfter({ type: "picked", choice: "B" });
		expect(checking.answer).toEqual({ state: "checking", choice: "B" });

		expect(
			next(checking, {
				type: "told",
				outcome: { kind: "answered", result: wrong },
			}).answer,
		).toEqual({ state: "answered", result: wrong });
		expect(
			next(checking, {
				type: "told",
				outcome: { kind: "answered", result: wrong, shown: wrongShown },
			}).answer,
		).toEqual({ state: "answered", result: wrong, shown: wrongShown });
		expect(
			next(checking, { type: "told", outcome: { kind: "closed" } }).answer,
		).toEqual({
			state: "closed",
		});
		expect(
			next(checking, { type: "told", outcome: { kind: "failed" } }).answer,
		).toEqual({
			state: "failed",
		});
	});

	test("is not picked twice while the first is checked", () => {
		const checking = lessonAfter({ type: "picked", choice: "B" });

		expect(next(checking, { type: "picked", choice: "C" })).toBe(checking);
	});

	test("that failed can be given again, and one recorded or closed cannot", () => {
		const failed = lessonAfter(
			{ type: "picked", choice: "B" },
			{ type: "told", outcome: { kind: "failed" } },
		);
		expect(canAnswer(failed)).toBe(true);
		expect(next(failed, { type: "picked", choice: "C" }).answer).toEqual({
			state: "checking",
			choice: "C",
		});

		for (const outcome of [
			{ kind: "answered", result: wrong },
			{ kind: "closed" },
		] as const) {
			const settled = lessonAfter(
				{ type: "picked", choice: "B" },
				{ type: "told", outcome },
			);
			expect(canAnswer(settled)).toBe(false);
			expect(next(settled, { type: "picked", choice: "C" })).toBe(settled);
		}
	});

	test("is not settled by a reply to nothing being checked", () => {
		const open = lessonAfter();

		expect(next(open, { type: "told", outcome: { kind: "closed" } })).toBe(
			open,
		);
	});
});

describe("the hint", () => {
	test("opens and closes, and stays used once opened", () => {
		const opened = lessonAfter({ type: "hint toggled" });
		expect(opened.hint).toEqual({ open: true, used: true });

		expect(next(opened, { type: "hint toggled" }).hint).toEqual({
			open: false,
			used: true,
		});
	});

	test("does not open while the answer is checked", () => {
		const checking = lessonAfter({ type: "picked", choice: "B" });

		expect(next(checking, { type: "hint toggled" })).toBe(checking);
	});
});

describe("an answer's state", () => {
	test.each<[Answer, boolean]>([
		[{ state: "open" }, false],
		[{ state: "checking", choice: "B" }, false],
		[{ state: "failed" }, false],
		[{ state: "closed" }, true],
		[{ state: "answered", result: wrong }, true],
	])("%o is settled %s", (answer, settled) => {
		expect(isSettled(answer)).toBe(settled);
	});
});

describe("an option", () => {
	const stateOf = (answer: Answer) =>
		letters.map((letter) => optionStateOf(answer, letter));

	test.each<[string, Answer, string[]]>([
		[
			"before an answer",
			{ state: "open" },
			["default", "default", "default", "default", "default"],
		],
		[
			"while B is checked",
			{ state: "checking", choice: "B" },
			["default", "selected", "default", "default", "default"],
		],
		[
			"of a closed task",
			{ state: "closed" },
			["default", "default", "default", "default", "default"],
		],
	])("stands %s", (_, answer, want) => {
		expect(stateOf(answer)).toEqual(want);
	});
});

describe("the line for the model", () => {
	// onTheCard is what the line says the card does with an answer it shows,
	// and what it leaves the model.
	const onTheCard =
		"The card has turned into how the answer went — the verdict, the trap, the solution step by step and the rating — with the buttons for another task and for the topic, and it sends nothing to the chat: never call show_result for it, and say nothing of the answer unless the adult asks; asked, add to what the card shows rather than retell it.";
	const aboutTheStep =
		"Word it about the step, addressing nobody, in short sentences that fit the child's grade, so the adult can read it out as it is, and so it does not show whether the child is a boy or a girl: speak of the child by the pseudonym, never as he or she, praise the step, not the child, and keep to the present tense.";

	test.each([
		[
			"a right answer",
			resultOf(rightAnswer),
			`Task task_fence has its answer recorded: C, which is right. ${onTheCard} ${aboutTheStep}`,
		],
		[
			"a wrong answer",
			wrong,
			`Task task_fence has its answer recorded: B, which is wrong; the right option is C. ${onTheCard} ${aboutTheStep}`,
		],
		[
			'"I don\'t know", recorded before',
			resultOf(dontKnowAnswer),
			`Task task_fence has its answer recorded: "I don't know", which counts as a wrong answer; the right option is C. ${onTheCard} ${aboutTheStep}`,
		],
		[
			"a wrong answer the service told no trap for",
			{ ...wrong, trap: null },
			`Task task_fence has its answer recorded: B, which is wrong; the right option is C. ${onTheCard} ${aboutTheStep}`,
		],
		[
			"a mistake the child has made before",
			resultOf(repeatedAnswer),
			`Task task_fence has its answer recorded: B, which is wrong; the right option is C. The child has made this mistake before among the latest answers, and the card says so. ${onTheCard} ${aboutTheStep}`,
		],
	])(
		"after %s says what is recorded, and that the card shows the rest",
		(_, result, want) => {
			expect(modelLineOf(result)).toBe(want);
		},
	);
});
