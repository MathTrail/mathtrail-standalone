import { describe, expect, test } from "vitest";
import topics from "../../../content/catalogs/topics.json";
import traps from "../../../content/catalogs/traps.json";
import file from "../../../site/data.json";
import { progressOf, readSiteData, siteData } from "./data";
import { siteDictionaries } from "./words";

// catalog is the service's catalog of topics and traps, with no reference task.
const catalog = { topics, traps, tasks: [] };

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
});

describe("the site's data", () => {
	// examples is the site's data with these examples of the topics' pages.
	const examples = (given: unknown) => ({ ...file, examples: given });

	test("sets each example at the grades of its level", () => {
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
			{ grades: [5, 6] },
			{ grades: [3, 4] },
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
				...file,
				progress: { ...file.progress, overall: { rank: "third" } },
			},
			"site/data.json: the progress is no progress the widget can draw",
		],
		[
			"a group that leaves a topic out",
			{ ...file, groups: file.groups.slice(1) },
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
			"site/data.json: the examples of logic.tables are for a page the catalog does not publish",
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
	])("is refused for %s", (_, broken, want) => {
		expect(() => readSiteData(catalog, broken)).toThrow(want);
	});
});
