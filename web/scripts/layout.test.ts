// @vitest-environment node
import { Window } from "happy-dom";
import { h, render } from "preact";
import { describe, expect, test } from "vitest";
import {
	allowed,
	type DrawnPicture,
	type Finding,
	type Measured,
	type PictureWord,
	pagesOf,
	pictureFindings,
	type Shard,
	shardOf,
	shareOf,
	told,
	unoffered,
	valuesOf,
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

	test.each<[string, Shard]>([
		["1/1", { part: 1, parts: 1 }],
		["2/6", { part: 2, parts: 6 }],
		["6/6", { part: 6, parts: 6 }],
	])("reads the shard %s", (asked, shard) => {
		expect(shardOf(asked)).toEqual(shard);
	});

	test.each([
		"0/6",
		"7/6",
		"2",
		"2/",
		"/6",
		"2/6/1",
		"-1/6",
		"1.5/6",
		"",
		"2 / 6",
	])("refuses %j as a shard", (asked) => {
		expect(() => shardOf(asked)).toThrow(`${asked} is no shard`);
	});
});

describe("a run shared out in shards", () => {
	test("measures every page once, in shares a page apart at most", () => {
		for (let parts = 1; parts <= 8; parts++) {
			for (let count = 0; count <= 30; count++) {
				const pages = Array.from({ length: count }, (_, at) => at);
				const shares = Array.from({ length: parts }, (_, at) =>
					shareOf(pages, { part: at + 1, parts }),
				);
				const run = `${count} pages in ${parts} shards`;
				expect(
					shares.flat().sort((one, other) => one - other),
					run,
				).toEqual(pages);
				const sizes = shares.map((share) => share.length);
				expect(
					Math.max(...sizes) - Math.min(...sizes),
					run,
				).toBeLessThanOrEqual(1);
			}
		}
	});

	test("measures at every width in every shard", () => {
		const languages = Array.from({ length: 23 }, (_, at) => `language ${at}`);
		const pages = pagesOf("chromium", languages, [320, 360, 640]);
		for (let part = 1; part <= 6; part++) {
			const widths = shareOf(pages, { part, parts: 6 }).map(
				(page) => page.width,
			);
			expect(new Set(widths), `shard ${part}/6`).toEqual(
				new Set([320, 360, 640]),
			);
		}
	});
});

describe("the values a list offers", () => {
	test("are its options' own, where Preact writes no attribute for them", () => {
		// The preview's list of languages, each option named by its value.
		const { document } = new Window();
		const list = document.createElement("select");
		const tags = ["en", "ar", "zh-Hans"];
		render(
			tags.map((tag) => h("option", { key: tag, value: tag }, tag)),
			list as unknown as HTMLSelectElement,
		);
		expect(
			valuesOf([...list.options] as unknown as HTMLOptionElement[]),
		).toEqual(tags);
	});
});

// word is a word of a picture, 14 px and drawn at that size, where given.
const word = (
	text: string,
	left: number,
	top: number,
	fields: Partial<PictureWord> = {},
): PictureWord => ({
	text,
	left,
	top,
	right: left + 20,
	bottom: top + 16,
	written: 14,
	size: 14,
	...fields,
});

// pictureOf is a picture 200 px wide in a box 286 px wide, with words.
const pictureOf = (
	words: PictureWord[],
	fields: Partial<DrawnPicture> = {},
): DrawnPicture => ({
	where: "svg.mt-picture",
	box: { left: 10, top: 10, right: 210, bottom: 110 },
	room: 286,
	words,
	...fields,
});

describe("a picture of a card", () => {
	test("that fits its box, its words inside it and apart at their size, is found to be no trouble", () => {
		expect(
			pictureFindings([
				pictureOf([word("A", 20, 20), word("B", 50, 20), word("C", 20, 40)]),
			]),
		).toEqual([]);
	});

	test("wider than its box is found", () => {
		expect(pictureFindings([pictureOf([], { room: 190 })])).toEqual([
			{ what: "is wider than its box", where: "svg.mt-picture", by: 10 },
		]);
	});

	test.each([
		["left", word("A", 7, 20)],
		["top", word("A", 20, 7)],
		["right", word("A", 193, 20)],
		["bottom", word("A", 20, 97)],
	])("with a word past its %s edge is found", (_, past) => {
		expect(pictureFindings([pictureOf([past])])).toEqual([
			{ what: "runs out of its picture", where: 'text "A"', by: 3 },
		]);
	});

	test("with a word drawn smaller than 11 px is found", () => {
		expect(
			pictureFindings([pictureOf([word("A", 20, 20, { size: 10.5 })])]),
		).toEqual([
			{ what: "is drawn smaller than 11 px", where: 'text "A"', by: 0.5 },
		]);
	});

	test("with a word the picture writes smaller than 11 px is found only where it is drawn smaller still", () => {
		const meant = word("A", 20, 20, { written: 8, size: 8 });
		const shrunk = word("B", 50, 20, { written: 8, size: 7.5 });

		expect(pictureFindings([pictureOf([meant, shrunk])])).toEqual([
			{ what: "is drawn smaller than 8 px", where: 'text "B"', by: 0.5 },
		]);
	});

	test("with a word over another is found", () => {
		expect(
			pictureFindings([pictureOf([word("A", 20, 20), word("B", 30, 25)])]),
		).toEqual([{ what: 'overlaps text "B"', where: 'text "A"', by: 1 }]);
	});
});
