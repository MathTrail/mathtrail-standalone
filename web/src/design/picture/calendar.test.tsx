import { describe, expect, test } from "vitest";
import { weekdayNames } from "./calendar";
import type { Calendar } from "./model";
import { drawn, numberOf, shapesOf, textOf, textsOf } from "./testing/svg";

// column is how wide a day's column is drawn.
const column = 40;

// draw is a month's page as the card draws it, in a language.
const draw = (calendar: Omit<Calendar, "kind">, locale = "en") =>
	drawn({ kind: "calendar", ...calendar }, locale);

describe("a month's page", () => {
	test.each([
		["Monday", 1, undefined, 0],
		["Wednesday", 3, undefined, 2],
		["Sunday", 7, undefined, 6],
		["Wednesday, its weeks on Sundays", 3, "sunday" as const, 3],
		["Sunday, its weeks on Sundays", 7, "sunday" as const, 0],
	])(
		"starting on a %s puts the 1st under its weekday",
		(_, first, weekStarts, place) => {
			const drawing = draw({ first, days: 30, week_starts: weekStarts });

			expect(numberOf(textOf(drawing, "1"), "x")).toBeCloseTo(
				(place + 0.5) * column,
				1,
			);
		},
	);

	test("runs its weeks under one another, as many as its days take", () => {
		const weeksOf = (first: number, days: number) =>
			new Set(
				textsOf(draw({ first, days }))
					.filter((text) => /^\d+$/.test(text.text))
					.map((text) => numberOf(text, "y")),
			).size;

		expect(weeksOf(1, 28)).toBe(4);
		expect(weeksOf(7, 31)).toBe(6);
	});

	test("writes its days in Latin digits in every language", () => {
		const days = textsOf(draw({ first: 1, days: 31 }, "ar"))
			.map((text) => text.text)
			.slice(7);

		expect(days).toEqual(
			Array.from({ length: 31 }, (_, before) => String(before + 1)),
		);
	});

	test("sets a marked day on a square and writes its label under its week", () => {
		const drawing = draw({ first: 1, days: 30, marks: { "12": "A" } });
		const [square] = shapesOf(drawing, "rect", "mt-pic-fill");
		const day = textOf(drawing, "12");

		expect(
			numberOf(square ?? day, "x") + numberOf(square ?? day, "width") / 2,
		).toBeCloseTo(numberOf(day, "x"), 1);
		expect(numberOf(textOf(drawing, "A"), "x")).toBeCloseTo(
			numberOf(day, "x"),
			1,
		);
		expect(numberOf(textOf(drawing, "A"), "y")).toBeGreaterThan(
			numberOf(day, "y"),
		);
	});
});

describe("the names of the weekdays", () => {
	test.each([
		["en", ["Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"]],
		["ru", ["пн", "вт", "ср", "чт", "пт", "сб", "вс"]],
		["ar", ["ن", "ث", "ر", "خ", "ج", "س", "ح"]],
		["th", ["จ", "อ", "พ", "พฤ", "ศ", "ส", "อา"]],
	])(
		"in %s are the short ones where they fit, else the narrow ones",
		(locale, names) => {
			expect(weekdayNames(locale, false, column).names).toEqual(names);
		},
	);

	test("start on Sunday where the weeks do", () => {
		expect(weekdayNames("en", true, column).names).toEqual([
			"Sun",
			"Mon",
			"Tue",
			"Wed",
			"Thu",
			"Fri",
			"Sat",
		]);
	});

	test("are the short ones at the smaller size where only that fits them", () => {
		expect(weekdayNames("pl", false, column)).toEqual({
			names: ["pon.", "wt.", "śr.", "czw.", "pt.", "sob.", "niedz."],
			size: 11,
		});
	});

	test("are never the narrow ones in Latin letters for a language written in another script", () => {
		const { names } = weekdayNames("ur", false, 20);

		expect(names).toEqual([
			"پیر",
			"منگل",
			"بدھ",
			"جمعرات",
			"جمعہ",
			"ہفتہ",
			"اتوار",
		]);
	});

	test("in a tag no browser reads are those the browser writes", () => {
		expect(weekdayNames("not a tag", false, column).names).toHaveLength(7);
	});

	test("in the pseudo-language are those of the language it stretches", () => {
		expect(weekdayNames("en-XA", false, column).names[0]).toBe("Mon");
	});
});
