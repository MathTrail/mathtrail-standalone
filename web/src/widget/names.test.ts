import { afterEach, describe, expect, test, vi } from "vitest";
import skills from "../../../content/catalogs/skills.json";
import topics from "../../../content/catalogs/topics.json";
import traps from "../../../content/catalogs/traps.json";
import english from "../../locales/en.json";
import {
	catalogSkills,
	countryName,
	knownTrapName,
	languageName,
	listed,
	rankCount,
	rankName,
	skillName,
	topicName,
	trapAdvice,
	trapName,
} from "./names";
import { cardWords } from "./words";

const inEnglish = cardWords("en", undefined);
const inRussian = cardWords("ru", undefined);

describe("the catalogs' names", () => {
	// A topic, a skill or a mistake the service can name that the card has no
	// words for would reach the child as its id.
	test.each(topics.map((topic) => topic.id))("the topic %s has words", (id) => {
		expect(Object.hasOwn(english, `topic.${id}`)).toBe(true);
	});

	test.each(skills.map((skill) => skill.id))("the skill %s has words", (id) => {
		expect(Object.hasOwn(english, `skill.${id}`)).toBe(true);
	});

	test.each(traps.map((trap) => trap.id))("the mistake %s has words", (id) => {
		expect(Object.hasOwn(english, `trap.${id}`)).toBe(true);
	});

	// The advice is the same English sentence the service tells the model, so
	// that there is one English for every language to be written from.
	test.each(traps.map((trap) => [trap.id, trap.advice]))(
		"the mistake %s has advice, in English the catalog's own",
		(id, advice) => {
			expect((english as Record<string, unknown>)[`advice.${id}`]).toBe(advice);
		},
	);

	test("advise in the card's language, and not at all on a mistake the card has no advice for", () => {
		expect(trapAdvice(inRussian, "missed_case")).toBe(
			"Выписывать случаи в одном порядке, начиная с меньшего, и отмечать каждый, чтобы ни один не потерялся.",
		);
		expect(trapAdvice(inEnglish, "counted_the_cat")).toBeUndefined();
	});

	test("are said in the card's language", () => {
		expect(topicName(inEnglish, "combinatorics.enumeration")).toBe(
			"Enumeration",
		);
		expect(topicName(inRussian, "combinatorics.enumeration")).toBe("Перебор");
		expect(skillName(inRussian, "division_with_remainder")).toBe(
			"Деление с остатком",
		);
		expect(trapName(inEnglish, "missed_case")).toBe(
			"Missed a case while listing",
		);
		expect(trapName(inRussian, "double_count")).toBe(
			"Одно и то же посчитано дважды",
		);
	});

	test("the card has no words for have no name to say inside a sentence", () => {
		expect(knownTrapName(inRussian, "double_count")).toBe(
			"Одно и то же посчитано дважды",
		);
		expect(knownTrapName(inEnglish, "counted_the_cat")).toBeUndefined();
	});

	test("the card has no words for are called by their ids", () => {
		expect(topicName(inEnglish, "logic.unknown")).toBe("logic.unknown");
		expect(skillName(inEnglish, "juggling")).toBe("juggling");
		expect(trapName(inEnglish, "counted_the_cat")).toBe("counted_the_cat");
	});
});

describe("the ranks", () => {
	test("are as many as the card counts, eleven, each named", () => {
		expect(rankCount).toBe(11);
		for (let rank = 1; rank <= rankCount; rank++) {
			expect(Object.hasOwn(english, `rank.${rank}`)).toBe(true);
		}
		expect(Object.hasOwn(english, `rank.${rankCount + 1}`)).toBe(false);
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

describe("a country", () => {
	test("is named in the card's language", () => {
		expect(countryName(inEnglish, "US")).toBe("United States");
		expect(countryName(inRussian, "FR")).toBe("Франция");
	});

	test("that no code names is shown as it is", () => {
		expect(countryName(inEnglish, "not a code")).toBe("not a code");
	});
});

describe("the skills a parent can leave out", () => {
	// A skill the form does not offer is one the parent cannot leave out from
	// the card, and the form offers them in the order the catalog lists them.
	test("are every skill of the catalog, in its order", () => {
		expect(catalogSkills).toEqual(skills.map((skill) => skill.id));
	});
});
