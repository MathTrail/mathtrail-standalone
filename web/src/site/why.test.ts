import { describe, expect, test } from "vitest";
import topics from "../../../content/catalogs/topics.json";
import traps from "../../../content/catalogs/traps.json";
import file from "../../../site/data.json";
import { answerOf, doiAddress, readWhy, type WhyCard, whyFile } from "./why";

// catalog is the service's catalog of topics and traps.
const catalog = { topics, traps };

// own is what the site's own data gives the page "Why", as its reader takes it.
const own = whyFile.parse(file.why);

// Finding is a finding as the site's data writes it: the work, and how it is drawn.
type Finding = (typeof own.findings)[number];

// findingOf is a finding of the work id, drawn as the data's first finding is.
const findingOf = (id: string): Finding => ({
	...(own.findings[0] ?? { hue: "blue", mark: "bars" }),
	id,
});

// withCard is the page's data with these fields of its card changed.
const withCard = (fields: Partial<WhyCard>) => ({
	...own,
	card: { ...own.card, ...fields },
});

describe("what the page Why takes from the site's own data", () => {
	const why = readWhy(catalog, own);

	test("shows its findings in the data's order, each of a work it lists", () => {
		expect(
			why.findings.map(({ id, hue, mark }) => ({ id, hue, mark })),
		).toEqual(own.findings);
		expect(why.apps.id).toBe(own.apps);
	});

	test("names as examples only topics of the catalog", () => {
		const known = new Set(topics.map((topic) => topic.id));

		for (const id of Object.values(why.topics)) {
			expect(known.has(id)).toBe(true);
		}
	});
});

describe("the card on the page Why", () => {
	const said = {
		language: "ru",
		child: "Комета",
		question: "Сколько столбов?",
		trap: "Посчитаны промежутки.",
		solution: "Четыре и один.",
	};

	test("is a task handed out in the page's language, with the page's words and no hint", () => {
		const { handed } = answerOf(own.card, said);

		expect(handed.language).toBe("ru");
		expect(handed.task.language).toBe("ru");
		expect(handed.child).toEqual({
			pseudonym: "Комета",
			grade: own.card.grade,
			ui_language: null,
		});
		expect(handed.task.question).toBe(said.question);
		expect(handed.task.options).toEqual(own.card.options);
		expect(handed.task.hint).toBe("");
	});

	test("is answered wrong, with the trap behind the option picked and the solution", () => {
		const { result } = answerOf(own.card, said);

		expect(result.correct).toBe(false);
		expect(result.choice).toBe(own.card.choice);
		expect(result.correct_answer).toBe(own.card.correct);
		expect(result.trap).toEqual({
			id: own.card.trap,
			text: said.trap,
			repeated: false,
		});
		expect(result.solution).toBe(said.solution);
		expect(result.rating).toEqual(own.card.rating);
	});
});

describe("what the page Why takes from the data", () => {
	test.each([
		[
			"a card of a topic the catalog does not have",
			withCard({ topic: "logic.tables" }),
			"the card on the page Why is of logic.tables, a topic the catalog does not have",
		],
		[
			"a card in a grade its topic is not taught in",
			withCard({ topic: "fractions.parts" }),
			"the card on the page Why is set in grade 3, which fractions.parts is not taught in",
		],
		[
			"a card that names a trap the catalog does not have",
			withCard({ trap: "counted_fence" }),
			"the card on the page Why names the trap counted_fence, which the catalog does not have",
		],
		[
			"a card answered right",
			withCard({ choice: "C" }),
			"the card on the page Why answers right, and the page shows a wrong answer",
		],
		[
			'a card answered "I don\'t know"',
			withCard({ choice: "?" }),
			"the card on the page Why picks ?, and the page shows an option picked",
		],
		[
			"a card with an option missing",
			withCard({ options: { A: "3", B: "4", C: "5", D: "6" } }),
			"the card on the page Why is no task the widget can draw",
		],
		[
			"a card whose right option is no letter",
			withCard({ correct: "F" }),
			"the card on the page Why holds no answer the widget can draw",
		],
		[
			"an example topic the catalog does not have",
			{ ...own, topics: { ...own.topics, tables: "logic.tables" } },
			"the page Why names logic.tables as an example, a topic the catalog does not have",
		],
		[
			"a finding of a work the data does not list",
			{ ...own, findings: [...own.findings, findingOf("smith2024")] },
			"the page Why cites smith2024, which the data's sources do not have",
		],
		[
			"a finding shown twice",
			{
				...own,
				findings: [...own.findings, findingOf(own.findings[0]?.id ?? "")],
			},
			"the page Why shows a finding twice",
		],
		[
			"a work cited nowhere",
			{ ...own, findings: own.findings.slice(1) },
			`the source ${own.findings[0]?.id} is cited nowhere on the page Why`,
		],
	])("refuses %s", (_, broken, want) => {
		expect(() => readWhy(catalog, broken)).toThrow(want);
	});
});

describe("the address of a work", () => {
	test.each([
		["10.1257/aeri.20190457", "https://doi.org/10.1257/aeri.20190457"],
		[
			"10.1002/(SICI)1097-4571(199806)49:8<693::AID-ASI4>3.0.CO;2-0",
			"https://doi.org/10.1002/(SICI)1097-4571(199806)49%3A8%3C693%3A%3AAID-ASI4%3E3.0.CO%3B2-0",
		],
		["10.1000/a#b?c", "https://doi.org/10.1000/a%23b%3Fc"],
	])("of DOI %s is %s: every part escaped, the slashes kept", (doi, want) => {
		expect(doiAddress({ doi })).toBe(want);
	});
});
