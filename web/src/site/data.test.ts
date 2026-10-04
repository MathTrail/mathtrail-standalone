import { describe, expect, test } from "vitest";
import catalog from "../../../content/catalogs/topics.json";
import file from "../../../site/data.json";
import { progressOf, readSiteData, siteData } from "./data";
import { siteDictionaries } from "./words";

describe("the site's own data", () => {
	const data = siteData();

	test("shows every topic of the catalog, each in one of its groups", () => {
		expect(data.topics.all.map((topic) => topic.id)).toEqual(
			catalog.map((topic) => topic.id),
		);
		expect(data.topics.groups.flatMap((group) => group.topics).sort()).toEqual(
			catalog.map((topic) => topic.id).sort(),
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
});

describe("the site's data", () => {
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
	])("is refused for %s", (_, broken, want) => {
		expect(() => readSiteData(catalog, broken)).toThrow(want);
	});
});
