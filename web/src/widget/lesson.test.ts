import type { CallToolResult } from "@modelcontextprotocol/client";
import { describe, expect, test } from "vitest";
import { letters } from "./choices";
import {
	type Answer,
	canAnswer,
	isOver,
	isSettled,
	type Lesson,
	type LessonEvent,
	lessonStart,
	modelLineOf,
	next,
	optionStateOf,
} from "./lesson";
import type { AnswerResult } from "./payload";
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
	test.each<[Answer, boolean, boolean]>([
		[{ state: "open" }, false, false],
		[{ state: "checking", choice: "B" }, false, false],
		[{ state: "failed" }, false, true],
		[{ state: "closed" }, true, true],
		[{ state: "answered", result: wrong }, true, true],
	])(
		"%o is settled %s and has something to say %s",
		(answer, settled, over) => {
			expect(isSettled(answer)).toBe(settled);
			expect(isOver(answer)).toBe(over);
		},
	);
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
			'while "I don\'t know" is checked',
			{ state: "checking", choice: "?" },
			["default", "default", "default", "default", "default"],
		],
		[
			"after a wrong B",
			{ state: "answered", result: wrong },
			["muted", "wrong", "correct", "muted", "muted"],
		],
		[
			"after a right C",
			{ state: "answered", result: resultOf(rightAnswer) },
			["muted", "muted", "correct", "muted", "muted"],
		],
		[
			'after "I don\'t know"',
			{ state: "answered", result: resultOf(dontKnowAnswer) },
			["muted", "muted", "correct", "muted", "muted"],
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
	test.each([
		[
			"a right answer",
			resultOf(rightAnswer),
			"Task task_fence has its answer recorded: C, which is right. The card shows the solution. In a language with grammatical gender, word it so it does not show whether the child is a boy or a girl: praise the step, not the child, and keep to the present tense.",
		],
		[
			"a wrong answer",
			wrong,
			"Task task_fence has its answer recorded: B, which is wrong; the right option is C. The card shows the trap and the solution. In a language with grammatical gender, word it so it does not show whether the child is a boy or a girl: praise the step, not the child, and keep to the present tense.",
		],
		[
			'"I don\'t know"',
			resultOf(dontKnowAnswer),
			'Task task_fence has its answer recorded: "I don\'t know", which counts as a wrong answer; the right option is C. The card shows the solution. In a language with grammatical gender, word it so it does not show whether the child is a boy or a girl: praise the step, not the child, and keep to the present tense.',
		],
		[
			"a wrong answer the service told no trap for",
			{ ...wrong, trap: null },
			"Task task_fence has its answer recorded: B, which is wrong; the right option is C. The card shows the solution. In a language with grammatical gender, word it so it does not show whether the child is a boy or a girl: praise the step, not the child, and keep to the present tense.",
		],
		[
			"a mistake the child has made before",
			resultOf(repeatedAnswer),
			"Task task_fence has its answer recorded: B, which is wrong; the right option is C. The card shows the trap and the solution. The child has made this mistake before among the latest answers: end your explanation with one short reminder of it, in your own words, that the child can keep in mind next time. In a language with grammatical gender, word it so it does not show whether the child is a boy or a girl: praise the step, not the child, and keep to the present tense.",
		],
	])("after %s says what is recorded", (_, result, want) => {
		expect(modelLineOf(result)).toBe(want);
	});
});
