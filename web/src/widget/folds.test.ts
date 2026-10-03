import { describe, expect, test } from "vitest";
import { allFolded, foldsAfter, type Open, type Section } from "./folds";

// pressed is which sections are open once each of sections is pressed in
// turn, from a progress with every section folded.
function pressed(...sections: Section[]): Open {
	return sections.reduce(foldsAfter, allFolded);
}

describe("the sections of the progress", () => {
	test("are all folded as it first opens", () => {
		expect([...allFolded]).toEqual([]);
	});

	test("open when pressed, and fold when pressed again", () => {
		expect([...pressed("topics")]).toEqual(["topics"]);
		expect([...pressed("topics", "topics")]).toEqual([]);
	});

	test("open and fold each on its own", () => {
		expect([...pressed("topics", "profile", "topics")]).toEqual(["profile"]);
	});

	test("count every press that comes together", () => {
		expect(new Set(pressed("topics", "mistakes", "recent", "profile"))).toEqual(
			new Set(["topics", "mistakes", "recent", "profile"]),
		);
	});

	test("leave the sections a press was given as they were", () => {
		const open = pressed("recent");

		foldsAfter(open, "profile");

		expect([...open]).toEqual(["recent"]);
	});
});
