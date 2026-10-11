import { renderToString } from "preact-render-to-string";
import { describe, expect, test } from "vitest";
import { parsePageWords } from "./pagewords";
import { openReader } from "./reader";
import { provedAnswer, provedBadge, statesTheNumbersOf } from "./WorkedSteps";

describe("an answer's words", () => {
	test.each([
		["a whole number", "12 apples.", "12"],
		["a time, its hours and its minutes", "A las 4:25.", "4:25"],
		[
			"a span in words of their own",
			"4 horas 30 minutos.",
			"4 hours 30 minutes",
		],
		[
			"a number with the ending of an ordinal",
			"In the middle, the 5th cell.",
			"5",
		],
		[
			"an answer proved in words, which has no number",
			"The green boat, 1st.",
			"green",
		],
	])("state %s", (_, words, answer) => {
		expect(statesTheNumbersOf(words, answer)).toBe(true);
	});

	test.each([
		["a number written as a word", "In the middle, the fifth cell.", "5"],
		["another number", "At 4:52.", "4:25"],
		[
			"the numbers in another order",
			"30 minutes and 4 hours.",
			"4 hours 30 minutes",
		],
		["a number inside a longer one", "112 apples.", "12"],
		["another number beside the right one", "3 knights and 2 liars.", "2"],
		["digits of another script", "١٢ تفاحة.", "12"],
	])("do not state %s", (_, words, answer) => {
		expect(statesTheNumbersOf(words, answer)).toBe(false);
	});
});

// readerOf is a reader of words that give an answer and its badge in Spanish.
const readerOf = (answer: string, badge = "") =>
	openReader(
		parsePageWords(`answer: "${answer}"\nbadge: "${badge}"\n`),
		"es/topics/clocks.yaml",
		"es",
	).page;

describe("a worked example's answer", () => {
	test("is the page's words, drawn, where they state the answer its solver proves", () => {
		expect(
			renderToString(
				<p>{provedAnswer(readerOf("A las **4:25**."), "answer", "4:25")}</p>,
			),
		).toBe("<p>A las <strong>4:25</strong>.</p>");
	});

	test("stops the build where its words state another number, naming the file and the key", () => {
		expect(() =>
			provedAnswer(readerOf("A las 4:52."), "answer", "4:25"),
		).toThrow(
			'es/topics/clocks.yaml: answer says "A las 4:52.", and the example\'s solver proves "4:25"',
		);
	});
});

describe("a worked example's badge", () => {
	test.each([
		["the answer's numbers", "4:30", "4 hours 30 minutes"],
		["a place, for an answer in words", "1", "green"],
		["a mark, for an answer in numbers", "✓", "12"],
	])("may show %s", (_, badge, answer) => {
		expect(provedBadge(readerOf("A.", badge), "badge", answer)).toBe(badge);
	});

	test("stops the build where it writes other numbers than the answer, naming the file and the key", () => {
		expect(() =>
			provedBadge(readerOf("A.", "4:03"), "badge", "4 hours 30 minutes"),
		).toThrow(
			'es/topics/clocks.yaml: badge shows "4:03", and the example\'s solver proves "4 hours 30 minutes"',
		);
	});
});
