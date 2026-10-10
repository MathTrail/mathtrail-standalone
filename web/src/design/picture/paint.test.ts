import { describe, expect, test } from "vitest";
import { paintedOf } from "./paint";

// lettersOf are the letters each colour's paint carries, by its word.
function lettersOf(colors: Record<string, string>, locale = "en") {
	return Object.fromEntries(
		paintedOf(colors, locale).map(({ word, letters }) => [word, letters]),
	);
}

describe("the letters on a paint", () => {
	test("are the first letter of its word, a capital, where no other word shares it", () => {
		expect(lettersOf({ red: "red", blue: "blue", yellow: "yellow" })).toEqual({
			red: "R",
			blue: "B",
			yellow: "Y",
		});
		expect(
			lettersOf({ red: "красная", blue: "синяя", yellow: "жёлтая" }, "ru"),
		).toEqual({ красная: "К", синяя: "С", жёлтая: "Ж" });
	});

	test("run on until no other word shares them, three at most", () => {
		expect(lettersOf({ black: "black", blue: "blue", red: "red" })).toEqual({
			black: "Bla",
			blue: "Blu",
			red: "R",
		});
	});

	test("past three, are the first letter and the first one where the words part", () => {
		expect(
			lettersOf({ red: "vermelho", green: "verde", blue: "azul" }, "pt"),
		).toEqual({ vermelho: "Vm", verde: "Vd", azul: "A" });
		// A word that is the start of another parts from it nowhere, and keeps
		// its three letters; the longer one parts where the shorter ends.
		expect(lettersOf({ red: "brown", blue: "browny" })).toEqual({
			brown: "Bro",
			browny: "By",
		});
	});

	test("tell a word of two by its first word's start and its second word's first letter", () => {
		expect(
			lettersOf(
				{ green: "xanh lá", blue: "xanh dương", red: "đỏ", black: "đen" },
				"vi",
			),
		).toEqual({ "xanh lá": "Xl", "xanh dương": "Xd", đỏ: "Đỏ", đen: "Đe" });
	});

	test("count a letter as a reader sees it, its marks with it", () => {
		expect(lettersOf({ yellow: "पीला", blue: "नीला" }, "hi")).toEqual({
			पीला: "पी",
			नीला: "नी",
		});
		expect(lettersOf({ white: "白", black: "黒" }, "ja")).toEqual({
			白: "白",
			黒: "黒",
		});
	});

	test("follow the palette's order, and leave out a colour with no word", () => {
		expect(
			paintedOf({ black: "black", red: "red" }, "en").map(({ paint }) => paint),
		).toEqual(["red", "black"]);
		expect(paintedOf(undefined, "en")).toEqual([]);
	});
});
