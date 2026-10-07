import { describe, expect, test } from "vitest";
import topics from "../../../content/catalogs/topics.json";
import traps from "../../../content/catalogs/traps.json";
import file from "../../../site/data.json";
import {
	connectAddress,
	type HomeCard,
	type HomeWords,
	homeAnswerOf,
	homeFile,
	homeResultsOf,
	homeTaskOf,
	readHome,
} from "./home";

// catalog is the service's catalog of topics and traps.
const catalog = { topics, traps };

// own is what the site's own data gives the home page, as its reader takes it.
const own = homeFile.parse(file.home);

// withCard is the page's data with these fields of its card changed.
const withCard = (fields: Partial<HomeCard>) => ({
	...own,
	card: { ...own.card, ...fields },
});

// withTraps is the page's data with these traps behind its card's options.
const withTraps = (changed: Record<string, string | undefined>) =>
	withCard({
		traps: Object.fromEntries(
			Object.entries({ ...own.card.traps, ...changed }).filter(
				(entry): entry is [string, string] => entry[1] !== undefined,
			),
		),
	});

// withExamples is the page's data with these traps named as examples.
const withExamples = (named: [string, string, string]) => ({
	...own,
	traps: named,
});

describe("what the home page takes from the site's own data", () => {
	const home = readHome(catalog, own);

	test("is a card with a trap of the catalog behind every wrong option, and none behind the right one", () => {
		const known = new Set(traps.map((trap) => trap.id));
		const wrong = Object.keys(home.card.options).filter(
			(letter) => letter !== home.card.correct,
		);

		expect(Object.keys(home.card.traps).sort()).toEqual(wrong.sort());
		for (const trap of Object.values(home.card.traps)) {
			expect(known.has(trap)).toBe(true);
		}
	});

	test("names as examples three traps of the catalog, in the data's order", () => {
		expect(home.traps).toEqual(own.traps);
	});
});

describe("the card on the home page", () => {
	const said: HomeWords = {
		language: "ru",
		child: "Комета",
		question: "Сколько столбов?",
		hint: "Нарисуй забор поменьше.",
		traps: {
			A: "Назван шаг.",
			B: "Посчитаны промежутки.",
			D: "Умножено.",
			E: "Взято из условия.",
		},
		solution: "Четыре и один.",
	};

	test("is a task handed out in the page's language, with the page's words, the hint and the data's drawing", () => {
		const handed = homeTaskOf(own.card, said);

		expect(handed.language).toBe("ru");
		expect(handed.child).toEqual({
			pseudonym: "Комета",
			grade: own.card.grade,
			ui_language: null,
		});
		expect(handed.task.question).toBe(said.question);
		expect(handed.task.hint).toBe(said.hint);
		expect(handed.task.drawing).toBe(own.card.drawing);
		expect(handed.task.options).toEqual(own.card.options);
	});

	test("is answered with the option its steps pick, told the trap behind that option", () => {
		const { handed, result } = homeAnswerOf(own.card, said);

		expect(handed.task.hint).toBe("");
		expect(result.correct).toBe(false);
		expect(result.choice).toBe(own.card.choice);
		expect(result.correct_answer).toBe(own.card.correct);
		expect(result.trap).toEqual({
			id: own.card.traps[own.card.choice],
			text: said.traps[own.card.choice],
			repeated: false,
		});
		expect(result.solution).toBe(said.solution);
		expect(result.rating).toEqual({
			before: own.card.rating.before,
			after: own.card.rating.wrong,
		});
	});

	// The card on the first screen takes any option, so the page carries what
	// the service would record of each, as the widget reads it.
	test("has a result for every option: each wrong one told its trap, the right one none", () => {
		const results = homeResultsOf(own.card, said);

		expect(
			Object.entries(results).map(([choice, result]) => [
				choice,
				result.choice,
				result.correct,
				result.trap?.id ?? null,
				result.trap?.text ?? null,
				result.rating?.after,
			]),
		).toEqual([
			["A", "A", false, own.card.traps.A, said.traps.A, own.card.rating.wrong],
			["B", "B", false, own.card.traps.B, said.traps.B, own.card.rating.wrong],
			["C", "C", true, null, null, own.card.rating.right],
			["D", "D", false, own.card.traps.D, said.traps.D, own.card.rating.wrong],
			["E", "E", false, own.card.traps.E, said.traps.E, own.card.rating.wrong],
		]);
		for (const result of Object.values(results)) {
			expect(result.correct_answer).toBe(own.card.correct);
			expect(result.solution).toBe(said.solution);
			expect(result.rating?.before).toBe(own.card.rating.before);
		}
	});

	test("refuses a wrong option the page says nothing of", () => {
		const { E: _, ...silent } = said.traps;

		expect(() => homeResultsOf(own.card, { ...said, traps: silent })).toThrow(
			"the home page says nothing of the trap behind its wrong option E",
		);
	});

	test("refuses a wrong option the card names no trap behind", () => {
		const { E: _, ...untrapped } = own.card.traps;

		expect(() =>
			homeResultsOf({ ...own.card, traps: untrapped }, said),
		).toThrow(
			"the card on the home page leaves its wrong option E without a trap",
		);
	});
});

describe("what the home page takes from the data", () => {
	test.each([
		[
			"a card answered right",
			withCard({ choice: "C" }),
			"the card on the home page answers right, and the page shows a wrong answer",
		],
		[
			"a wrong option with no trap behind it",
			withTraps({ E: undefined }),
			"the card on the home page leaves its wrong option E without a trap",
		],
		[
			"a trap behind the right option",
			withTraps({ C: "off_by_one" }),
			"the card on the home page ties a trap to C, its right option",
		],
		[
			"a trap the catalog does not have behind an option",
			withTraps({ A: "counted_fence" }),
			"the card on the home page ties A to the trap counted_fence, which the catalog does not have",
		],
		[
			"a trap behind no option",
			withTraps({ F: "off_by_one" }),
			"the card on the home page ties a trap to F, which is no option",
		],
		[
			"an example trap the catalog does not have",
			withExamples(["off_by_one", "missed_case", "counted_fence"]),
			"the home page names counted_fence as an example, a trap the catalog does not have",
		],
		[
			"a card of a topic the catalog does not have",
			withCard({ topic: "logic.tables" }),
			"the card on the home page is of logic.tables, a topic the catalog does not have",
		],
		[
			"a card in a grade its topic is not taught in",
			withCard({ topic: "fractions.parts" }),
			"the card on the home page is set in grade 3, which fractions.parts is not taught in",
		],
		[
			"a rating that a right answer does not raise",
			withCard({ rating: { before: 1502, wrong: 1480, right: 1502 } }),
			"the card on the home page moves its rating from 1502 to 1480 on a wrong answer and to 1502 on a right one: a right answer raises it, and a wrong one lowers it",
		],
		[
			"a rating that a wrong answer does not lower",
			withCard({ rating: { before: 1502, wrong: 1510, right: 1524 } }),
			"a right answer raises it, and a wrong one lowers it",
		],
		[
			"a card with an option missing",
			withCard({ options: { A: "3", B: "4", C: "5", D: "36" } }),
			"the card on the home page is no task the widget can draw",
		],
	])("refuses %s", (_, broken, want) => {
		expect(() => readHome(catalog, broken)).toThrow(want);
	});

	test('refuses a card whose steps pick "I don\'t know", or nothing a card shows', () => {
		for (const choice of ["?", "F"]) {
			expect(
				homeFile.safeParse({
					...file.home,
					card: { ...file.home.card, choice },
				}).success,
			).toBe(false);
		}
	});
});

describe("connecting", () => {
	test("is the section on connecting of the home page in the reader's language", () => {
		expect(connectAddress("ru")).toBe("/ru/#connect");
	});

	test("is on the bare domain for a reader of English, whose home page is there", () => {
		expect(connectAddress("en")).toBe("/#connect");
	});
});
