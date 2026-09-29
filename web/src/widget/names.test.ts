import { afterEach, describe, expect, test, vi } from "vitest";
import skills from "../../../content/catalogs/skills.json";
import topics from "../../../content/catalogs/topics.json";
import english from "../../locales/en.json";
import { languageName, listed, rankName, skillName, topicName } from "./names";
import { cardWords } from "./words";

const inEnglish = cardWords("en", undefined);
const inRussian = cardWords("ru", undefined);

describe("the catalogs' names", () => {
	// A topic or a skill the service can name that the card has no words for
	// would reach the child as its id.
	test.each(topics.map((topic) => topic.id))("the topic %s has words", (id) => {
		expect(Object.hasOwn(english, `topic.${id}`)).toBe(true);
	});

	test.each(skills.map((skill) => skill.id))("the skill %s has words", (id) => {
		expect(Object.hasOwn(english, `skill.${id}`)).toBe(true);
	});

	test("are said in the card's language", () => {
		expect(topicName(inEnglish, "combinatorics.enumeration")).toBe(
			"Enumeration",
		);
		expect(topicName(inRussian, "combinatorics.enumeration")).toBe("Перебор");
		expect(skillName(inRussian, "division_with_remainder")).toBe(
			"Деление с остатком",
		);
	});

	test("the card has no words for are called by their ids", () => {
		expect(topicName(inEnglish, "logic.unknown")).toBe("logic.unknown");
		expect(skillName(inEnglish, "juggling")).toBe("juggling");
	});
});

describe("the ranks", () => {
	test("are eleven, each named", () => {
		for (let rank = 1; rank <= 11; rank++) {
			expect(Object.hasOwn(english, `rank.${rank}`)).toBe(true);
		}
		expect(Object.hasOwn(english, "rank.12")).toBe(false);
		expect(rankName(inEnglish, 3)).toBe("River crossing");
		expect(rankName(inRussian, 11)).toBe("Над облаками");
	});

	test("a card has no name for are called by their numbers", () => {
		expect(rankName(inEnglish, 12)).toBe("12");
	});
});

describe("a list of names", () => {
	test("is written with the marks of the card's language", () => {
		expect(listed(inEnglish, ["space", "animals", "football"])).toBe(
			"space, animals, football",
		);
		expect(listed(inRussian, ["космос", "футбол"])).toBe("космос, футбол");
	});
});

describe("a list of names on a platform with no list formats", () => {
	afterEach(() => {
		vi.unstubAllGlobals();
	});

	test("is set apart by commas, and the card is still drawn", () => {
		vi.stubGlobal("Intl", { ...Intl, ListFormat: undefined });

		expect(listed(inEnglish, ["space", "animals"])).toBe("space, animals");
	});
});

describe("a language", () => {
	test("is named in the card's language, written to begin a line", () => {
		expect(languageName(inEnglish, "pt-BR")).toBe("Brazilian Portuguese");
		expect(languageName(inRussian, "ru")).toBe("Русский");
	});

	test("that no tag names is shown as it is", () => {
		expect(languageName(inEnglish, "not a tag")).toBe("not a tag");
	});
});
