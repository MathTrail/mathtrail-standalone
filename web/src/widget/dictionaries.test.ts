import { describe, expect, test } from "vitest";
import { type Dictionary, placeholder, type Wording } from "../i18n/words";
import { dictionaries } from "./words";

// textsOf are the texts a wording says: its one, or one per plural category.
function textsOf(wording: Wording): unknown[] {
	return typeof wording === "string" ? [wording] : Object.values(wording);
}

// slotsOf are the slots a wording names, in any of its texts, sorted.
function slotsOf(wording: Wording): string[] {
	const names = textsOf(wording).flatMap((text) =>
		typeof text === "string"
			? [...text.matchAll(placeholder)].flatMap(([, name]) => name ?? [])
			: [],
	);
	return [...new Set(names)].sort();
}

// strayBraces says whether a text has a brace that is not part of a
// placeholder — {Grade}, {task_id}, an unclosed {grade — which the card would
// show as it is written and the comparison of slots could not see.
function strayBraces(text: string): boolean {
	return /[{}]/.test(text.replace(placeholder, ""));
}

// disagreements are the ways the words of one language, tagged tag, say other
// things than the English ones: a key one of them lacks, a wording with other
// slots, a wording plural in one and not in the other, a text that is empty or
// no text, a brace that is no placeholder, and a plural wording that does not
// name exactly the categories its language counts in — one it lacks is a form
// of the words the card cannot say, and one its language never chooses is a
// form nobody reads.
function disagreements(
	english: Dictionary,
	words: Dictionary,
	tag: string,
): string[] {
	const found = Object.keys(english)
		.filter((key) => !Object.hasOwn(words, key))
		.map((key) => `${key} is missing`);
	const counted = [
		...new Intl.PluralRules(tag).resolvedOptions().pluralCategories,
	].sort();
	for (const [key, wording] of Object.entries(words)) {
		const reference = english[key];
		if (reference === undefined) {
			found.push(`${key} is not among the English words`);
			continue;
		}
		if (slotsOf(wording).join() !== slotsOf(reference).join()) {
			found.push(
				`${key} names the slots [${slotsOf(wording)}], the English [${slotsOf(reference)}]`,
			);
		}
		if ((typeof wording === "string") !== (typeof reference === "string")) {
			found.push(`${key} changes with a number in one language only`);
		}
		if (
			textsOf(wording).some(
				(text) => typeof text !== "string" || text.trim() === "",
			)
		) {
			found.push(`${key} has a text that is empty or no text`);
		}
		if (
			textsOf(wording).some(
				(text) => typeof text === "string" && strayBraces(text),
			)
		) {
			found.push(`${key} has a brace that is no placeholder`);
		}
		if (typeof wording !== "string") {
			const named = Object.keys(wording).sort();
			if (named.join() !== counted.join()) {
				found.push(
					`${key} is written for [${named}], ${tag} counts in [${counted}]`,
				);
			}
		}
	}
	return found;
}

describe("the widget's words", () => {
	const english = dictionaries.get("en") ?? {};

	test("are in English, the language every other is written from", () => {
		expect(dictionaries.has("en")).toBe(true);
		expect(Object.keys(english).length).toBeGreaterThan(0);
	});

	test.each([...dictionaries.keys()])(
		"in %s are named by the tag a browser writes",
		(tag) => {
			expect(new Intl.Locale(tag).baseName).toBe(tag);
		},
	);

	test.each([...dictionaries])(
		"in %s say what the English say",
		(tag, words) => {
			expect(disagreements(english, words, tag)).toEqual([]);
		},
	);
});

// The comparison above passes for words that agree only if it can fail for
// words that do not: each way to disagree is caught, and named.
describe("words that disagree with English", () => {
	const english: Dictionary = {
		"task.hint": "Hint",
		"result.wrong": "Not quite — it's {correct}, not {picked}.",
		"progress.skipped": { one: "{count} skipped", other: "{count} skipped" },
	};
	const russian: Dictionary = {
		"task.hint": "Подсказка",
		"result.wrong": "Не совсем: ответ {correct}, а не {picked}.",
		"progress.skipped": {
			one: "{count} пропущена",
			few: "{count} пропущены",
			many: "{count} пропущено",
			other: "{count} пропущено",
		},
	};

	test("agree when they say the same", () => {
		expect(disagreements(english, russian, "ru")).toEqual([]);
		expect(disagreements(english, english, "en")).toEqual([]);
	});

	test.each<[string, Dictionary, string]>([
		[
			"a key missing",
			Object.fromEntries(
				Object.entries(russian).filter(([key]) => key !== "result.wrong"),
			),
			"result.wrong is missing",
		],
		[
			"a key English lacks",
			{ ...russian, "task.more": "Ещё" },
			"task.more is not among the English words",
		],
		[
			"another slot",
			{
				...russian,
				"result.wrong": "Не совсем: ответ {right}, а не {picked}.",
			},
			"result.wrong names the slots [picked,right], the English [correct,picked]",
		],
		[
			"a number in one language only",
			{ ...russian, "progress.skipped": "пропущено: {count}" },
			"progress.skipped changes with a number in one language only",
		],
		[
			"an empty text",
			{ ...russian, "task.hint": " " },
			"task.hint has a text that is empty or no text",
		],
		[
			"a category its language needs left out",
			{
				...russian,
				"progress.skipped": {
					one: "{count} пропущена",
					other: "{count} пропущено",
				},
			},
			"progress.skipped is written for [one,other], ru counts in [few,many,one,other]",
		],
	])("are caught for %s", (_, words, want) => {
		expect(disagreements(english, words, "ru")).toContain(want);
	});

	test.each(["{Grade} класс", "{grade_label} класс", "{grade класс"])(
		"are caught for a brace that is no placeholder, in %s",
		(wording) => {
			const typo: Dictionary = { ...english, "child.grade": wording };

			expect(disagreements(typo, typo, "ru")).toContain(
				"child.grade has a brace that is no placeholder",
			);
		},
	);

	test("are caught for a category their language never chooses", () => {
		const extra: Dictionary = {
			...english,
			"progress.skipped": {
				one: "{count} skipped",
				few: "{count} skipped",
				other: "{count} skipped",
			},
		};

		expect(disagreements(english, extra, "en")).toContain(
			"progress.skipped is written for [few,one,other], en counts in [one,other]",
		);
	});
});
