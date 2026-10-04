import { describe, expect, test } from "vitest";
import { pageWordsDisagreements, parsePageWords } from "./pagewords";

describe("a page's words", () => {
	test("are laid out into keys in dot notation, in the order the file writes them", () => {
		const words = parsePageWords(
			[
				"# A note for whoever translates.",
				"title: Why",
				"description: Why olympiad maths.",
				"hero:",
				"  lead: >",
				"    One line,",
				"    folded.",
				"  tips:",
				"    - First",
				"    - question: Second",
				"      answer: 12",
				"drawing: |",
				"  |--3--|",
				"  * *",
			].join("\n"),
		);

		expect([...words]).toEqual([
			["title", "Why"],
			["description", "Why olympiad maths."],
			["hero.lead", "One line, folded."],
			["hero.tips.1", "First"],
			["hero.tips.2.question", "Second"],
			["hero.tips.2.answer", "12"],
			["drawing", "|--3--|\n* *"],
		]);
	});

	test("are text every one, whatever a value looks like", () => {
		expect([
			...parsePageWords("count: 12\nfree: yes\nnone: null\nwhen: 2026-10-04\n"),
		]).toEqual([
			["count", "12"],
			["free", "yes"],
			["none", "null"],
			["when", "2026-10-04"],
		]);
	});

	test("keep a key written with nothing after it, as an empty text", () => {
		expect([...parsePageWords("title: T\nlead:\n")]).toEqual([
			["title", "T"],
			["lead", ""],
		]);
	});

	test.each([
		["an empty file", "", "holds 0 documents"],
		["a file of comments", "# nothing yet\n", "holds 0 documents"],
		["two documents", "title: A\n---\ntitle: B\n", "holds 2 documents"],
		["a list at the top", "- a\n- b\n", "no set of sections and texts"],
		["a text at the top", "just words\n", "no set of sections and texts"],
		["a key written twice", "title: A\ntitle: B\n", "Map keys must be unique"],
		["an anchor and its alias", "a: &x hi\nb: *x\n", "names &x"],
		["a tag", "count: !!int 3\n", "tags a value"],
		["a tag of text", "count: !!str 3\n", "tags a value"],
		["a merge key", "base: &b\n  x: 1\nc:\n  <<: *b\n", "names &b"],
		[
			"a text that begins with a star and no quote",
			"title: **Knights**\n",
			"reads **Knights** as an alias",
		],
		["a key in capitals", "Title: A\n", '"Title" is no key'],
		["a key with a dot", "hero.lead: A\n", '"hero.lead" is no key'],
		["a key of digits", "2: A\n", '"2" is no key'],
		["an empty section", "hero: {}\n", "hero is an empty section"],
		["an empty list", "tips: []\n", "tips is an empty list"],
		["a broken file", "title: [A\n", "Flow sequence in block collection"],
	])("are refused for %s", (_, source, want) => {
		expect(() => parsePageWords(source)).toThrow(want);
	});
});

describe("a translation of a page's words", () => {
	const english = parsePageWords(
		"title: Why\nhero:\n  lead: Grades {grade}.\n  note: Free.\n",
	);

	test("agrees when it has the English keys, slots and order", () => {
		expect(
			pageWordsDisagreements(
				english,
				parsePageWords(
					"title: Зачем\nhero:\n  lead: Классы {grade}.\n  note: Бесплатно.\n",
				),
				"ru",
			),
		).toEqual([]);
	});

	test.each([
		[
			"a key missing",
			"title: Зачем\nhero:\n  lead: Классы {grade}.\n",
			"hero.note is missing",
		],
		[
			"a key English does not have",
			"title: Зачем\nhero:\n  lead: Классы {grade}.\n  note: Да.\n  extra: Ещё.\n",
			"hero.extra is not among the English words",
		],
		[
			"another slot",
			"title: Зачем\nhero:\n  lead: Классы {grades}.\n  note: Да.\n",
			"hero.lead names the slots [grades], the English [grade]",
		],
		[
			"an empty text",
			"title: Зачем\nhero:\n  lead: Классы {grade}.\n  note:\n",
			"hero.note has a text that is empty or no text",
		],
		[
			"a mark a reader cannot see",
			"title: Зачем\nhero:\n  lead: Классы {grade}.\n  note: Бес\u00adплатно.\n",
			"hero.note has a mark a reader cannot see",
		],
		[
			"the English keys in another order",
			"title: Зачем\nhero:\n  note: Да.\n  lead: Классы {grade}.\n",
			"hero.note comes where the English has hero.lead",
		],
	])("disagrees for %s", (_, source, want) => {
		expect(
			pageWordsDisagreements(english, parsePageWords(source), "ru"),
		).toEqual([want]);
	});

	test("holds the English to itself, for an empty text of its own", () => {
		const empty = parsePageWords("title: Why\nlead:\n");

		expect(pageWordsDisagreements(empty, empty, "en")).toEqual([
			"lead has a text that is empty or no text",
		]);
	});
});
