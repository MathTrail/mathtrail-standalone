import { describe, expect, test } from "vitest";
import topics from "../../../content/catalogs/topics.json";
import traps from "../../../content/catalogs/traps.json";
import file from "../../../site/data.json";
import fixture from "../../../site/research/testdata/research.json";
import { progressOf, readSiteData, siteData } from "./data";
import { siteDictionaries } from "./words";

// catalog is the service's catalog of topics and traps, with no reference task,
// in which the page of knights and liars alone is published.
const catalog = {
	topics: topics.map((topic) => ({
		...topic,
		site_page: topic.id === "logic.knights_liars",
	})),
	traps,
	tasks: [],
};

describe("the site's own data", () => {
	const data = siteData();

	test("shows every topic of the catalog, each in one of its groups", () => {
		expect(data.topics.all.map((topic) => topic.id)).toEqual(
			topics.map((topic) => topic.id),
		);
		expect(data.topics.groups.flatMap((group) => group.topics).sort()).toEqual(
			topics.map((topic) => topic.id).sort(),
		);
	});

	test.each([...siteDictionaries])(
		"names every group in the site's words in %s",
		(_, dictionary) => {
			for (const group of data.topics.groups) {
				expect(Object.hasOwn(dictionary, `group.${group.id}`)).toBe(true);
			}
		},
	);

	test("holds a progress the widget draws, for the child a page names", () => {
		const report = progressOf(data, "Комета");

		expect(report.profile.pseudonym).toBe("Комета");
		expect(report.topics.length).toBeGreaterThan(0);
	});

	test("gives examples to the pages of the topics the catalog publishes, and to no other", () => {
		expect([...data.examples.keys()].sort()).toEqual(
			topics
				.filter((topic) => topic.site_page)
				.map((topic) => topic.id)
				.sort(),
		);
	});

	test("draws the card of every topic of the catalog, and the first screen of every page it publishes", () => {
		expect([...data.drawings.keys()].sort()).toEqual(
			topics.map((topic) => topic.id).sort(),
		);
		expect(
			[...data.drawings]
				.filter(([, drawings]) => drawings.hero !== undefined)
				.map(([id]) => id)
				.sort(),
		).toEqual(
			topics
				.filter((topic) => topic.site_page)
				.map((topic) => topic.id)
				.sort(),
		);
	});

	test("ranks the traps of every topic of the catalog", () => {
		expect([...data.traps.keys()].sort()).toEqual(
			topics.map((topic) => topic.id).sort(),
		);
	});

	test("ranks the traps of knights and liars a missed condition, a statement trusted and a solution stopped halfway, the commonest first", () => {
		expect(data.traps.get("logic.knights_liars")?.slice(0, 3)).toEqual([
			"ignored_condition",
			"trusted_statement",
			"stopped_early",
		]);
	});

	test("holds no numbers of the page Research when the build gives it none", () => {
		expect(data.research).toBeUndefined();
	});

	test("holds the numbers of the page Research the build gives it, with the authors of the books the reference tasks come from", () => {
		const research = siteData(fixture).research;

		expect(research?.bench.rows.length).toBe(fixture.bench.rows.length);
		expect(research?.authors.map(({ id }) => id)).toEqual([
			"perelman",
			"ignatyev",
			"dudeney",
			"loyd",
			"carroll",
		]);
		for (const author of research?.authors ?? []) {
			expect(author.books.length).toBeGreaterThan(0);
		}
	});
});

describe("the site's data", () => {
	// knights is the site's data with the drawings and the examples of that one
	// page.
	const knights = {
		...file,
		drawings: { "logic.knights_liars": file.drawings["logic.knights_liars"] },
		examples: { "logic.knights_liars": file.examples["logic.knights_liars"] },
	};
	// examples is the site's data with these examples of the topics' pages.
	const examples = (given: unknown) => ({ ...knights, examples: given });
	// drawings is the site's data with these drawings of the topics.
	const drawings = (given: unknown) => ({ ...knights, drawings: given });
	// card is the site's data with this drawing on the card of knights and liars.
	const card = (given: unknown) =>
		drawings({ "logic.knights_liars": { card: given } });
	// clock is a picture of a clock, which every drawing may be.
	const clock = { kind: "clock", time: "4:50" };

	test("carries the drawings of each topic, and of each example, as the file writes them", () => {
		const data = readSiteData(catalog, {
			...knights,
			drawings: {
				"logic.knights_liars": {
					card: { markup: "islanders" },
					hero: { picture: clock },
				},
				"logic.ordering": { card: { markup: "daisy" } },
			},
			examples: {
				"logic.knights_liars": [
					{
						level: "5-6",
						solver: "round-table",
						answer: "5",
						drawing: { markup: "round-table" },
					},
					{ level: "3-4", solver: "two-on-the-road", answer: "A liar" },
				],
			},
		});

		expect(data.drawings.get("logic.knights_liars")).toEqual({
			card: { markup: "islanders" },
			hero: { picture: clock },
		});
		expect(data.drawings.get("logic.ordering")).toEqual({
			card: { markup: "daisy" },
		});
		expect(data.examples.get("logic.knights_liars")).toEqual([
			{ grades: [5, 6], answer: "5", drawing: { markup: "round-table" } },
			{ grades: [3, 4], answer: "A liar" },
		]);
	});

	test("sets each example at the grades of its level, with its answer", () => {
		const data = readSiteData(
			catalog,
			examples({
				"logic.knights_liars": [
					{ level: "5-6", solver: "round-table", answer: "5" },
					{ level: "3-4", solver: "two-on-the-road", answer: "A liar" },
				],
			}),
		);

		expect(data.examples.get("logic.knights_liars")).toEqual([
			{ grades: [5, 6], answer: "5" },
			{ grades: [3, 4], answer: "A liar" },
		]);
	});

	test.each([
		[
			"a file of another shape",
			{ progress: file.progress },
			"site/data.json: ✖ Invalid input: expected array, received undefined\n  → at groups",
		],
		[
			"a progress the widget cannot draw",
			{
				...knights,
				progress: { ...file.progress, overall: { rank: "third" } },
			},
			"site/data.json: the progress is no progress the widget can draw",
		],
		[
			"a group that leaves a topic out",
			{ ...knights, groups: file.groups.slice(1) },
			"site/data.json: the topic logic.ordering is in no group",
		],
		[
			"an example of a page the catalog does not publish",
			examples({
				"logic.ordering": [{ level: "1-2", solver: "one", answer: "1" }],
			}),
			"site/data.json: the examples of logic.ordering are for a page the catalog does not publish",
		],
		[
			"an example of a topic the catalog does not have",
			examples({
				"logic.tables": [{ level: "1-2", solver: "one", answer: "1" }],
			}),
			"site/data.json: the examples of logic.tables are of a topic the catalog does not have",
		],
		[
			"an example at a level its topic is not taught at",
			examples({
				"logic.knights_liars": [{ level: "1-2", solver: "one", answer: "1" }],
			}),
			"site/data.json: an example of logic.knights_liars is set at 1-2, a level the topic is not taught at",
		],
		[
			"an example whose solver is named otherwise than a file of its own",
			examples({
				"logic.knights_liars": [
					{ level: "3-4", solver: "../round-table", answer: "5" },
				],
			}),
			'→ at examples["logic.knights_liars"][0].solver',
		],
		[
			"an example with no answer to prove",
			examples({
				"logic.knights_liars": [{ level: "3-4", solver: "one", answer: "" }],
			}),
			'→ at examples["logic.knights_liars"][0].answer',
		],
		[
			"a drawing of a topic the catalog does not have",
			drawings({ "logic.tables": { card: { markup: "islanders" } } }),
			"site/data.json: the drawings of logic.tables are of a topic the catalog does not have",
		],
		[
			"a drawing of the first screen of a page the catalog does not publish",
			drawings({
				"logic.ordering": {
					card: { markup: "islanders" },
					hero: { markup: "islanders" },
				},
			}),
			"site/data.json: the first screen of logic.ordering has a drawing, and the catalog does not publish its page",
		],
		[
			"a topic's drawings with no card",
			drawings({ "logic.knights_liars": { hero: { picture: clock } } }),
			'→ at drawings["logic.knights_liars"].card',
		],
		[
			"a picture the card's reader refuses",
			card({ picture: { kind: "clock", time: "25:00" } }),
			'→ at drawings["logic.knights_liars"].card.picture.time',
		],
		[
			"a drawing of the site's own it does not draw",
			card({ markup: "spiral" }),
			'→ at drawings["logic.knights_liars"].card.markup',
		],
		[
			"a drawing that is neither a picture nor one of the site's own",
			card({}),
			'a drawing is a picture or a drawing of the site\'s own, one of the two\n  → at drawings["logic.knights_liars"].card',
		],
		[
			"a drawing that is a picture and one of the site's own at once",
			card({ picture: clock, markup: "islanders" }),
			'a drawing is a picture or a drawing of the site\'s own, one of the two\n  → at drawings["logic.knights_liars"].card',
		],
		[
			"a drawing with a member no drawing has",
			card({ markup: "islanders", words: "A" }),
			'Unrecognized key: "words"\n  → at drawings["logic.knights_liars"].card',
		],
		[
			"an example's drawing under a misspelt name",
			examples({
				"logic.knights_liars": [
					{
						level: "3-4",
						solver: "one",
						answer: "1",
						drawng: { picture: clock },
					},
				],
			}),
			'Unrecognized key: "drawng"\n  → at examples["logic.knights_liars"][0]',
		],
		[
			"an example's drawing the card's reader refuses",
			examples({
				"logic.knights_liars": [
					{
						level: "3-4",
						solver: "one",
						answer: "1",
						drawing: { picture: { kind: "clock" } },
					},
				],
			}),
			'→ at examples["logic.knights_liars"][0].drawing.picture.time',
		],
		[
			"a work the page Why cites whose DOI is no DOI",
			{
				...knights,
				why: {
					...file.why,
					sources: {
						...file.why.sources,
						agarwal2020: {
							...file.why.sources.agarwal2020,
							doi: "aeaweb.org/articles?id=10.1257/aeri.20190457",
						},
					},
				},
			},
			"→ at why.sources.agarwal2020.doi",
		],
		[
			"a card on the page Why that the catalog would not set",
			{
				...knights,
				why: { ...file.why, card: { ...file.why.card, grade: 9 } },
			},
			"site/data.json: the card on the page Why is set in grade 9, which counting.gaps is not taught in",
		],
		[
			"a technique that leads to a topic the catalog does not have",
			{
				...knights,
				techniques: {
					groups: [
						{
							id: "see",
							techniques: [
								{
									id: "draw",
									topics: ["logic.tables"],
									example: { level: "1-2", solver: "one", answer: "1" },
								},
							],
						},
					],
					cues: [["draw"]],
				},
			},
			"site/data.json: the technique draw leads to logic.tables, a topic the catalog does not have",
		],
		[
			"a technique with no example to work through",
			{
				...knights,
				techniques: {
					groups: [{ id: "see", techniques: [{ id: "draw", topics: [] }] }],
					cues: [["draw"]],
				},
			},
			"→ at techniques.groups[0].techniques[0].example",
		],
	])("is refused for %s", (_, broken, want) => {
		expect(() => readSiteData(catalog, broken)).toThrow(want);
	});
});
