import { describe, expect, test, vi } from "vitest";
import { disagreements } from "./dictionaries";
import { onAMachineSpeaking } from "./testing/machine";
import type { Dictionary } from "./words";

// The comparison passes for words that agree only if it can fail for words
// that do not: each way to disagree is caught, and named.
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

	test.each([
		["a mark that turns text right to left", 0x200f],
		["the mark of Arabic letters", 0x061c],
		["an embedding that turns it left to right", 0x202a],
		["a space nobody sees", 0x200b],
		["a joiner of words that keeps a line from breaking", 0x2060],
		["a soft hyphen", 0x00ad],
	])("are caught for %s", (_, mark) => {
		const hidden: Dictionary = {
			...english,
			"task.hint": `Hi${String.fromCodePoint(mark)}nt`,
		};

		expect(disagreements(english, hidden, "en")).toContain(
			"task.hint has a mark a reader cannot see",
		);
	});

	test.each([
		["the non-joiner Persian writes its words with", 0x200c],
		["the joiner the scripts of India write theirs with", 0x200d],
	])("let %s through", (_, joiner) => {
		const joined: Dictionary = {
			...english,
			"task.hint": `می${String.fromCodePoint(joiner)}شود`,
		};

		expect(disagreements(english, joined, "en")).toEqual([]);
	});

	// Klingon, which no platform has data for, counts as English does, whatever
	// language the machine that checks it speaks.
	test("are held to English's categories in a language the platform has no data for", () => {
		onAMachineSpeaking("ru");
		try {
			expect(disagreements(english, english, "tlh")).toEqual([]);
		} finally {
			vi.restoreAllMocks();
		}
	});

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
