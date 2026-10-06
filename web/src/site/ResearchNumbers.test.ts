import { describe, expect, test } from "vitest";
import { monthWritten, written } from "./ResearchNumbers";

// noBreak is the space Russian writes between a number's groups and before
// its percent sign, which keeps the number on one line.
const noBreak = String.fromCodePoint(0xa0);

describe("a number of the page Research", () => {
	test.each([
		["a share, as a percentage to a tenth", 0.036, "share", "3.6%"],
		["a share, rounded to a tenth", 0.5377, "share", "53.8%"],
		["a measure in logits, to two decimals", 0.5377, "logit", "0.54"],
		["a measure in logits, its decimals kept at zero", 0, "logit", "0.00"],
		["answers, to a tenth", 4.6305, "answers", "4.6"],
		["points, whole", 40.6, "points", "41"],
		["changes a hundred answers, to a tenth", 1.6333, "per_100_answers", "1.6"],
		["a chance, to a tenth at the least", 0.2, "chance", "0.2"],
		["a chance, to a thousandth at the most", 0.12345, "chance", "0.123"],
		["an end of a range of chance, to two decimals", 0.6, "hundredths", "0.60"],
		["an end of a range of chance at nothing", 0, "hundredths", "0.00"],
		["a difference above nothing, with its sign", 0.018, "signed", "+0.018"],
		["a difference below nothing, with its sign", -0.0125, "signed", "-0.013"],
		["no difference, with no sign", 0, "signed", "0.0"],
		["a count, in groups", 20261001, "count", "20,261,001"],
		["a year, its digits together", 1942, "year", "1942"],
		["a seed, its digits together", 20261001, "seed", "20261001"],
	] as const)("is written in English as %s", (_, value, form, text) => {
		expect(written("en", value, form)).toBe(text);
	});

	test.each([
		["a share", 0.036, "share", `3,6${noBreak}%`],
		["a measure in logits", 0.5377, "logit", "0,54"],
		["a count", 20261001, "count", `20${noBreak}261${noBreak}001`],
		["an end of a range of chance", 0.6, "hundredths", "0,60"],
		["a difference of chances", -0.021, "signed", "-0,021"],
		["a year", 1942, "year", "1942"],
		["a seed", 20261001, "seed", "20261001"],
	] as const)(
		"is written in Russian as Russian writes %s",
		(_, value, form, text) => {
			expect(written("ru", value, form)).toBe(text);
		},
	);

	test("is written with at least as many digits as it is asked for", () => {
		expect(written("en", 2, "count", 2)).toBe("02");
		expect(written("en", 12, "count", 2)).toBe("12");
	});
});

describe("a month of the page Research", () => {
	test.each([
		["en", "2026-11", "November 2026"],
		["en", "2027-01", "January 2027"],
		["ru", "2026-11", "ноябрь 2026 г."],
	])("is named in %s as its language names %s", (locale, month, named) => {
		expect(monthWritten(locale, month)).toBe(named);
	});
});
