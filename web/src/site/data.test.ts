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

	test("draws the card of every topic of the catalog", () => {
		expect([...data.drawings.keys()].sort()).toEqual(
			topics.map((topic) => topic.id).sort(),
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
	// home is the site's data with these members of the home page's.
	const home = (given: object) => ({
		...knights,
		home: { ...file.home, ...given },
	});
	// pick is the site's data with these members of how the home page's card
	// was picked.
	const pick = (given: object) =>
		home({ pick: { ...file.home.pick, ...given } });

	test("carries the drawing of each topic's card as the file writes it", () => {
		const data = readSiteData(catalog, {
			...knights,
			drawings: {
				"logic.knights_liars": { card: { markup: "knight-and-liar" } },
				"logic.ordering": { card: { markup: "lined-up" } },
			},
		});

		expect(data.drawings.get("logic.knights_liars")).toEqual({
			card: { markup: "knight-and-liar" },
		});
		expect(data.drawings.get("logic.ordering")).toEqual({
			card: { markup: "lined-up" },
		});
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
			drawings({ "logic.tables": { card: { markup: "fence" } } }),
			"site/data.json: the drawings of logic.tables are of a topic the catalog does not have",
		],
		[
			"a drawing of a first screen, which a topic's page draws itself",
			drawings({
				"logic.knights_liars": {
					card: { markup: "knight-and-liar" },
					hero: { markup: "fence" },
				},
			}),
			'Unrecognized key: "hero"\n  → at drawings["logic.knights_liars"]',
		],
		[
			"a topic's drawings with no card",
			drawings({ "logic.knights_liars": {} }),
			'→ at drawings["logic.knights_liars"].card',
		],
		[
			"a drawing of the site's own it does not draw",
			card({ markup: "spiral" }),
			'→ at drawings["logic.knights_liars"].card.markup',
		],
		[
			"a card that names no drawing",
			card({}),
			'→ at drawings["logic.knights_liars"].card.markup',
		],
		[
			"a picture on a card, which draws the site's own drawings alone",
			card({ markup: "fence", picture: { kind: "clock", time: "4:50" } }),
			'Unrecognized key: "picture"\n  → at drawings["logic.knights_liars"].card',
		],
		[
			"a drawing with a member no drawing has",
			card({ markup: "fence", words: "A" }),
			'Unrecognized key: "words"\n  → at drawings["logic.knights_liars"].card',
		],
		[
			"an example with a member no example has",
			examples({
				"logic.knights_liars": [
					{
						level: "3-4",
						solver: "one",
						answer: "1",
						drawng: { markup: "fence" },
					},
				],
			}),
			'Unrecognized key: "drawng"\n  → at examples["logic.knights_liars"][0]',
		],
		[
			"an example's drawing, which a topic's page draws itself",
			examples({
				"logic.knights_liars": [
					{
						level: "3-4",
						solver: "one",
						answer: "1",
						drawing: { markup: "fence" },
					},
				],
			}),
			'Unrecognized key: "drawing"\n  → at examples["logic.knights_liars"][0]',
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
		[
			"a skill of the home page's child in a topic the catalog does not have",
			pick({
				skills: [
					...file.home.pick.skills,
					{ topic: "logic.tables", share: 0.5 },
				],
			}),
			"site/data.json: the home page shows the child's skill in logic.tables, a topic the catalog does not have",
		],
		[
			"a skill's bar filled past its end",
			pick({ skills: [{ topic: "counting.gaps", share: 1.2 }] }),
			"site/data.json: the home page fills 1.2 of the bar of counting.gaps, which is no share of it",
		],
		[
			"a pick that leaves the card's topic out of the child's skills",
			pick({ skills: [{ topic: "parity.alternation", share: 0.58 }] }),
			"site/data.json: the home page picks counting.gaps, the topic of its card, and leaves it out of the child's skills",
		],
		[
			"a corridor that runs the wrong way",
			pick({ corridor: { low: 0.85, high: 0.7 } }),
			"site/data.json: the home page's corridor runs from 0.85 to 0.7: two shares, the lower first",
		],
		[
			"a corridor that ends past a share",
			pick({ corridor: { low: 0.7, high: 1.5 } }),
			"site/data.json: the home page's corridor runs from 0.7 to 1.5: two shares, the lower first",
		],
		[
			"a chance below its corridor",
			pick({ chance: 0.5 }),
			"site/data.json: the home page picks its task at a chance of 0.5, outside the corridor from 0.7 to 0.85",
		],
		[
			"a chance above its corridor",
			pick({ chance: 0.9 }),
			"site/data.json: the home page picks its task at a chance of 0.9, outside the corridor from 0.7 to 0.85",
		],
		[
			"a task of the chat alone with one right option",
			home({ alone: { ...file.home.alone, right: ["A"] } }),
			"site/data.json: the task of the chat alone has 1 right options, and the home page shows it for having more than one",
		],
		[
			"a task of the chat alone right in an option it does not have",
			home({ alone: { ...file.home.alone, right: ["A", "D"] } }),
			"site/data.json: the task of the chat alone has D right, which is no option of it",
		],
		[
			"a map of a topic the catalog does not have",
			home({ map: "logic.tables" }),
			"site/data.json: the home page shows what logic.tables builds on, a topic the catalog does not have",
		],
		[
			"a map of a topic that builds on other than two",
			home({ map: "counting.gaps" }),
			"site/data.json: the home page shows counting.gaps between two topics it builds on, and it builds on 0",
		],
	])("is refused for %s", (_, broken, want) => {
		expect(() => readSiteData(catalog, broken)).toThrow(want);
	});

	test("is refused for a map of a topic that builds on more than two", () => {
		const three = {
			...catalog,
			topics: catalog.topics.map((topic) =>
				topic.id === file.home.map
					? { ...topic, builds_on: [...topic.builds_on, "counting.gaps"] }
					: topic,
			),
		};

		expect(() => readSiteData(three, knights)).toThrow(
			`site/data.json: the home page shows ${file.home.map} between two topics it builds on, and it builds on 3`,
		);
	});
});
