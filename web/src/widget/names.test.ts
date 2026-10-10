import { afterEach, describe, expect, test, vi } from "vitest";
import skills from "../../../content/catalogs/skills.json";
import topics from "../../../content/catalogs/topics.json";
import traps from "../../../content/catalogs/traps.json";
import english from "../../locales/en.json";
import russian from "../../locales/ru.json";
import { onAMachineSpeaking } from "../i18n/testing/machine";
import { openWords } from "../i18n/words";
import siteEnglish from "../site/locales/en.json";
import siteRussian from "../site/locales/ru.json";
import { cardWords, catalogSkills } from "./dictionaries";
import {
	countryName,
	groupName,
	knownTrapName,
	languageName,
	listed,
	skillName,
	topicName,
	trapAdvice,
	trapName,
} from "./names";
import { topicGroups } from "./topicGroups";
import type { Key } from "./words";

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

	// A page of the site gives a card the words of its own language alone,
	// with no English behind them: a name is asked of the words the card was
	// given, never of a dictionary the card did not get.
	test("are asked of the words the card was given, whatever English has", () => {
		const russianAlone = openWords<Key>("ru", new Map([["ru", russian]]));
		const nameless = openWords<Key>("ru", new Map([["ru", {}]]));

		expect(topicName(russianAlone, "combinatorics.enumeration")).toBe(
			"Перебор",
		);
		expect(topicName(nameless, "combinatorics.enumeration")).toBe(
			"combinatorics.enumeration",
		);
	});
});

describe("the groups of topics", () => {
	// The card names a group as the site's page of topics does, in the two
	// languages the site is written in, so that a link from one leads to a
	// part of the page under the same name.
	test.each(topicGroups.map((group) => group.id))(
		"the group %s is named as the site names it",
		(id) => {
			const key = `group.${id}`;
			expect(groupName(inEnglish, id)).toBe(
				siteEnglish[key as keyof typeof siteEnglish],
			);
			expect(groupName(inRussian, id)).toBe(
				siteRussian[key as keyof typeof siteRussian],
			);
		},
	);

	test("the card has no words for are called by their ids", () => {
		expect(groupName(inEnglish, "puzzles")).toBe("puzzles");
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

// Names in Klingon, which no platform has data for.
describe("names in a language the platform has no data for", () => {
	const inKlingon = openWords<Key>("tlh", new Map([["en", english]]));

	afterEach(() => {
		vi.restoreAllMocks();
	});

	test("are listed and named as English does, whatever the machine speaks", () => {
		onAMachineSpeaking("de");

		expect(listed(inKlingon, ["space", "animals", "football"])).toBe(
			"space, animals, football",
		);
		expect(countryName(inKlingon, "FR")).toBe("France");
		expect(languageName(inKlingon, "fr")).toBe("French");
	});
});

describe("the skills a parent can leave out", () => {
	// A skill the form does not offer is one the parent cannot leave out from
	// the card, and the form offers them in the order the catalog lists them.
	test("are every skill of the catalog, in its order", () => {
		expect(catalogSkills).toEqual(skills.map((skill) => skill.id));
	});
});
