// @vitest-environment node
import { describe, expect, test } from "vitest";
import { addressOf, unsettled } from "./drive.ts";

describe("what kept the cards from settling", () => {
	test("names by scene the cards with nothing drawn and those still changing", () => {
		expect(
			unsettled(
				["task", "wrong answer", "progress"],
				["", "<p>2</p>", "<p>1</p>"],
				["", "<p>1</p>", "<p>1</p>"],
			),
		).toBe("nothing drawn in task; still changing: wrong answer");
	});

	test("names only the cards still changing when every card is drawn", () => {
		expect(
			unsettled(
				["task", "progress"],
				["<p>2</p>", "<p>1</p>"],
				["<p>1</p>", "<p>1</p>"],
			),
		).toBe("still changing: task");
	});

	test("counts the cards when some came or went in between", () => {
		expect(
			unsettled(["task", "progress"], ["<p>1</p>", "<p>1</p>"], ["<p>1</p>"]),
		).toBe("cards came or went: 1 a tick before, 2 now");
	});

	test("says so when the page holds no card", () => {
		expect(unsettled([], [], [])).toBe("no card on the page");
	});

	test("is nothing once every card is drawn and holds what it held", () => {
		expect(
			unsettled(
				["task", "progress"],
				["<p>1</p>", "<p>2</p>"],
				["<p>1</p>", "<p>2</p>"],
			),
		).toBe("");
	});

	test("keeps a card drawn for the first time from counting as settled", () => {
		expect(unsettled(["task"], ["<p>1</p>"], [""])).toBe(
			"still changing: task",
		);
	});
});

describe("the address of a picture", () => {
	test("asks the preview for its scene alone, in its theme, in English, at its width", () => {
		const address = new URL(
			addressOf("http://127.0.0.1:5173/", {
				scene: "wrong after the trial series",
				theme: "light",
				width: 640,
				path: "wrong.png",
			}),
		);
		expect(address.pathname).toBe("/preview.html");
		expect(Object.fromEntries(address.searchParams)).toEqual({
			scene: "wrong after the trial series",
			theme: "light",
			lang: "en",
			widths: "640",
		});
	});
});
