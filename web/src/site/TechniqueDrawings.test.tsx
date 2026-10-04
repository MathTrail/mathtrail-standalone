import { Window } from "happy-dom";
import { renderToString } from "preact-render-to-string";
import { afterAll, describe, expect, test } from "vitest";
import { parsePageWords } from "./pagewords";
import { openReader } from "./reader";
import { TechniqueDrawing } from "./TechniqueDrawings";

const browser = new Window();

afterAll(async () => {
	await browser.happyDOM.close();
});

// drawn is the drawing of the technique called id from words, as a document,
// and the keys of the words it never read.
function drawn(id: string, words: string) {
	const { page, unread } = openReader(
		parsePageWords(words),
		"en/techniques.yaml",
		"en",
	);
	const html = renderToString(
		<TechniqueDrawing id={id} page={page} at="drawing" />,
	);
	return {
		doc: new browser.DOMParser().parseFromString(html, "text/html"),
		unread: unread(),
	};
}

describe("the drawing of the table of pets", () => {
	// The table's names, written in another order than its rows and columns.
	const words = [
		"drawing:",
		"  title: Table",
		"  children:",
		"    vera: Vera",
		"    anya: Anya",
		"    borya: Borya",
		"  pets:",
		"    parrot: Parrot",
		"    dog: Dog",
		"    cat: Cat",
	].join("\n");

	test("ticks each child's own pet, whatever order the words name them in", () => {
		const { doc, unread } = drawn("table", words);
		const cells = [...(doc.querySelector(".s-pets")?.children ?? [])];
		const heads = cells.slice(1, 4).map((cell) => cell.textContent);
		const owned = Object.fromEntries(
			[0, 1, 2].map((row) => {
				const marks = cells.slice(5 + row * 4, 8 + row * 4);
				const ticked = marks.findIndex((cell) =>
					cell.classList.contains("s-pets-yes"),
				);
				return [cells[4 + row * 4]?.textContent, heads[ticked]];
			}),
		);

		expect(owned).toEqual({ Anya: "Dog", Borya: "Parrot", Vera: "Cat" });
		expect(unread).toEqual([]);
	});
});

describe("a technique with no drawing", () => {
	test("stops the build", () => {
		expect(() => drawn("guess", "drawing:\n  title: None\n")).toThrow(
			"the page of the techniques has no drawing for guess",
		);
	});
});
