import { describe, expect, test } from "vitest";
import { commonMark, emphasisOf, inlineTokens } from "./inline";

describe("a text's emphasis", () => {
	test.each([
		[
			"emphasis and strong emphasis the page draws",
			"It is **free**, and *always*.",
			"emphasis 1, strong emphasis 1, emphasis signs shown as written 0",
		],
		[
			"a star that stands for a digit, shown as it is written",
			"The star in 38*6",
			"emphasis 0, strong emphasis 0, emphasis signs shown as written 1",
		],
		[
			"a star written as an escape",
			"A \\* stays a star",
			"emphasis 0, strong emphasis 0, emphasis signs shown as written 1",
		],
		[
			"an underscore of a tool's name, shown as it is written",
			"submit_task — up to 3 attempts",
			"emphasis 0, strong emphasis 0, emphasis signs shown as written 1",
		],
		[
			"an emphasis of underscores the page cannot read",
			"__無料。__いつでも",
			"emphasis 0, strong emphasis 0, emphasis signs shown as written 4",
		],
		[
			"an emphasis closed after a full stop and before a letter, which the page cannot read",
			"**矛盾。**次に",
			"emphasis 0, strong emphasis 0, emphasis signs shown as written 4",
		],
	])("counts %s", (_, text, want) => {
		expect(emphasisOf(inlineTokens(text))).toBe(want);
	});

	test("reads a list's item in its own words, so its emphasis is no asterisk", () => {
		expect(emphasisOf(commonMark.lexer("- The **English** text.\n"))).toBe(
			"emphasis 0, strong emphasis 1, emphasis signs shown as written 0",
		);
	});
});
