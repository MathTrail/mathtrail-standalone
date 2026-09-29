import { render } from "preact";
import { act } from "preact/test-utils";
import { afterEach, describe, expect, test } from "vitest";
import {
	cardWords,
	dictionaries,
	languageChosenIn,
	ratingText,
	useWords,
	WordsContext,
} from "./words";

describe("a card's words", () => {
	test.each([
		["ru", "en-US", "ru"],
		["en", "ru-RU", "en"],
		[undefined, "ru-RU", "ru"],
		["kk", "ru-RU", "ru"],
		[undefined, "es-MX", "en"],
		[undefined, undefined, "en"],
	])(
		"with the language chosen %s and the host's %s are in %s",
		(chosen, host, want) => {
			expect(cardWords(chosen, host).locale).toBe(want);
		},
	);

	test("say what their dictionary says, with the grade in its slot", () => {
		const russian = cardWords("ru", undefined);

		expect(russian.text("task.hint")).toBe("Подсказка");
		expect(russian.text("child.grade", { grade: 3 })).toBe("3 класс");
		expect(
			cardWords(undefined, "en-GB").text("child.grade", { grade: 3 }),
		).toBe("Grade 3");
	});

	test("are in English and in Russian", () => {
		expect([...dictionaries.keys()]).toEqual(
			expect.arrayContaining(["en", "ru"]),
		);
	});
});

describe("a rating", () => {
	test.each([
		["en", "1573"],
		["ru", "1573"],
	])("in %s is written as chess writes it, with no separator", (tag, want) => {
		expect(ratingText(cardWords(tag, undefined), 1573)).toBe(want);
	});

	test("is written in the digits of the words' language", () => {
		const bengali = { locale: "bn", dir: "ltr" as const, text: () => "" };

		expect(ratingText(bengali, 1573)).toBe("১৫৭৩");
	});
});

describe("the language chosen for the cards", () => {
	test.each([
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
		expect(languageChosenIn(payload)).toBe(want);
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
		expect(languageChosenIn(payload)).toBeUndefined();
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

	test("speaks English outside any card", () => {
		act(() => render(<Hint />, root));

		expect(root.textContent).toBe("Hint");
	});
});
