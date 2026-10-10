import { render } from "preact";
import { act } from "preact/test-utils";
import { afterEach, describe, expect, test, vi } from "vitest";
import english from "../../locales/en.json";
import { onAMachineSpeaking } from "../i18n/testing/machine";
import { openWords } from "../i18n/words";
import { cardWords } from "./dictionaries";
import {
	countText,
	type Key,
	languageIn,
	percentText,
	ratingText,
	useWords,
	WordsContext,
} from "./words";

describe("a rating", () => {
	test.each([
		["en", "1573"],
		["ru", "1573"],
	])("in %s is written as chess writes it, with no separator", (tag, want) => {
		expect(ratingText(cardWords(tag, undefined), 1573)).toBe(want);
	});

	test("is written in the digits of the words' language", () => {
		const bengali = {
			locale: "bn",
			dir: "ltr" as const,
			text: () => "",
			has: (_: string): _ is never => false,
		};

		expect(ratingText(bengali, 1573)).toBe("১৫৭৩");
	});
});

// Numbers in Klingon, which no platform has data for.
describe("a number in a language the platform has no data for", () => {
	afterEach(() => {
		vi.restoreAllMocks();
	});

	test("is written in English digits, whatever the machine speaks", () => {
		onAMachineSpeaking("fa");
		const inKlingon = openWords<Key>("tlh", new Map([["en", english]]));

		expect(ratingText(inKlingon, 1573)).toBe("1573");
		expect(countText(inKlingon, 21)).toBe("21");
		expect(percentText(inKlingon, 40)).toBe("40%");
	});
});

describe("the language of a card", () => {
	test.each([
		[
			"a task card, from its lesson over the language chosen",
			{
				screen: "task",
				child: { pseudonym: "Otter", grade: 2, ui_language: "ru" },
				language: "en",
			},
			"en",
		],
		[
			"a task card with no language chosen, from its lesson",
			{
				screen: "task",
				child: { pseudonym: "Otter", grade: 2, ui_language: null },
				language: "ru",
			},
			"ru",
		],
		[
			"a task card from before its lesson's language travelled with it, from the language chosen",
			{
				screen: "task",
				child: { pseudonym: "Otter", grade: 2, ui_language: "ru" },
				task: { language: "en" },
			},
			"ru",
		],
		[
			"a waiting card, from the request it waits for",
			{ screen: "waiting", child: { ui_language: null }, language: "ru" },
			"ru",
		],
		[
			"a task card, with the child",
			{
				screen: "task",
				child: { pseudonym: "Otter", grade: 2, ui_language: "ru" },
			},
			"ru",
		],
		[
			"a waiting card, with the child",
			{ screen: "waiting", child: { ui_language: "pt-BR" } },
			"pt-BR",
		],
		[
			"the progress, with the profile",
			{ screen: "progress", profile: { ui_language: "ru" } },
			"ru",
		],
		[
			"the profile, with the profile",
			{ screen: "profile", profile: { pseudonym: "Otter", ui_language: "en" } },
			"en",
		],
	])("is read from %s", (_, payload, want) => {
		expect(languageIn(payload)).toBe(want);
	});

	test.each([
		[
			"none chosen",
			{
				screen: "task",
				child: { pseudonym: "Otter", grade: 2, ui_language: null },
			},
		],
		["an empty one", { screen: "profile", profile: { ui_language: "" } }],
		["no profile yet", { screen: "first_run", profile: null }],
		[
			"a result, which names no child",
			{ screen: "result", result: { correct: true } },
		],
		[
			"a language that is no text",
			{ screen: "task", child: { ui_language: 7 } },
		],
		["no payload", undefined],
	])("is none for %s", (_, payload) => {
		expect(languageIn(payload)).toBeUndefined();
	});
});

describe("a component inside a card", () => {
	const root = document.createElement("div");

	afterEach(() => {
		act(() => render(null, root));
	});

	function Hint() {
		return <span>{useWords().text("task.hint")}</span>;
	}

	test("speaks the card's words", () => {
		act(() =>
			render(
				<WordsContext.Provider value={cardWords("ru", undefined)}>
					<Hint />
				</WordsContext.Provider>,
				root,
			),
		);

		expect(root.textContent).toBe("Подсказка");
	});

	// A card always speaks the words the place it is drawn in gives it: a
	// component drawn outside any card is a mistake, and says so, rather than
	// speak an English nobody chose.
	test("is refused outside any card", () => {
		expect(() => act(() => render(<Hint />, root))).toThrow(/outside any card/);
	});
});
