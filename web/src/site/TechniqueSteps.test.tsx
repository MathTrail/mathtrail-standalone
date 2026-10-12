import { Window } from "happy-dom";
import { renderToString } from "preact-render-to-string";
import { afterAll, describe, expect, test } from "vitest";
import { parsePageWords } from "./pagewords";
import { openReader } from "./reader";
import { StepDrawing } from "./TechniqueSteps";

const browser = new Window();

afterAll(async () => {
	await browser.happyDOM.close();
});

// table is the words of the table's pictures, in Chinese, whose children's
// names all begin with the same character.
const table = [
	"techniques:",
	"  table:",
	"    pictures:",
	"      children:",
	"        anya: 小雅",
	"        borya: 小斌",
	"        vera: 小薇",
	"      faces:",
	"        anya: 雅",
	"        borya: 斌",
	"        vera: 薇",
	"      pets:",
	"        cat: 猫",
	"        dog: 狗",
	"        parrot: 鹦鹉",
].join("\n");

describe("the table of a technique", () => {
	test("draws on each child's face the letter its words give, not the first of the name", () => {
		const { page } = openReader(
			parsePageWords(table),
			"zh-Hans/techniques.yaml",
			"zh-Hans",
		);
		const drawn = new browser.DOMParser().parseFromString(
			renderToString(<StepDrawing page={page} technique="table" step={0} />),
			"text/html",
		);
		expect(
			[...drawn.querySelectorAll(".s-step-pets-face")].map(
				(face) => face.textContent,
			),
		).toEqual(["雅", "斌", "薇"]);
	});
});
