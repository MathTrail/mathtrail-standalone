// @vitest-environment node
import { describe, expect, test } from "vitest";
import {
	allowed,
	type Finding,
	type Measured,
	told,
	unoffered,
	widthOf,
} from "./layout.ts";

// cut is a finding of text cut short in the element described.
const cut = (where: string): Finding => ({
	what: "is cut short",
	where,
	by: 12,
});

describe("a finding of the layout", () => {
	test("is allowed for the pseudonym at the top, which gives way", () => {
		expect(allowed(cut('span.mt-bar-name "SuperCometTheGreatExplorer"'))).toBe(
			true,
		);
	});

	test("is allowed for the text of a box a person types into, which scrolls", () => {
		expect(
			allowed({
				what: "runs out of its box sideways",
				where: 'input.mt-input ""',
				by: 28,
			}),
		).toBe(true);
	});

	test.each<[string, Finding]>([
		["text cut short anywhere else", cut('span.mt-badge "Grade 3"')],
		[
			"the pseudonym spilling over rather than cut",
			{
				what: "runs out of its box sideways",
				where: 'span.mt-bar-name "Comet"',
				by: 4,
			},
		],
		[
			"an element whose name only begins like the pseudonym's",
			cut('span.mt-bar-name-other "Comet"'),
		],
		[
			"a box a person types into that sticks out of the card",
			{ what: "sticks out of the card", where: 'input.mt-input ""', by: 9 },
		],
		[
			"a field around the box that runs out of its own box",
			{
				what: "runs out of its box sideways",
				where: 'div.mt-form-field "Language"',
				by: 10,
			},
		],
		[
			"a page that scrolls sideways",
			{ what: "the page scrolls sideways", where: "html", by: 7 },
		],
	])("is not allowed for %s", (_, finding) => {
		expect(allowed(finding)).toBe(false);
	});
});

describe("the findings told", () => {
	// measured is a card measured at a moment, with what it showed.
	const measured = (moment: string, ...findings: Finding[]): Measured => ({
		engine: "chromium",
		language: "de",
		width: 320,
		scene: "task",
		moment,
		findings,
	});
	const spill: Finding = {
		what: "runs out of its box sideways",
		where: 'div.mt-head-text "Du"',
		by: 17,
	};

	test("are each told once, with the moment they were first seen", () => {
		expect(
			told([
				measured("at once", spill),
				measured("after 35 s", spill),
				measured("after 125 s", spill, cut('span.mt-badge "Klasse 3"')),
			]),
		).toEqual([
			'chromium de 320px, task: div.mt-head-text "Du" runs out of its box sideways by 17px (at once)',
			'chromium de 320px, task: span.mt-badge "Klasse 3" is cut short by 12px (after 125 s)',
		]);
	});

	test("are none for cards that fit", () => {
		expect(told([measured("at once"), measured("after 35 s")])).toEqual([]);
	});
});

describe("what is asked for", () => {
	test("and not offered is named", () => {
		expect(unoffered(["ar", "zh-hans", "xx"], ["ar", "zh-Hans"])).toEqual([
			"zh-hans",
			"xx",
		]);
		expect(unoffered([320, 400], [320, 360, 640])).toEqual([400]);
		expect(unoffered([], ["en"])).toEqual([]);
	});

	test.each(["320", "640"])("reads the width %s", (asked) => {
		expect(widthOf(asked)).toBe(Number(asked));
	});

	test.each(["320px", "0", "-360", "1.5", ""])(
		"refuses %j as a width",
		(asked) => {
			expect(() => widthOf(asked)).toThrow(`${asked} is no width in pixels`);
		},
	);
});
