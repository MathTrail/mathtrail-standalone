import { Window } from "happy-dom";
import type { VNode } from "preact";
import { renderToString } from "preact-render-to-string";
import { afterAll, describe, expect, test } from "vitest";
import { Chain, DigitTree } from "./Art";

const browser = new Window();

afterAll(async () => {
	await browser.happyDOM.close();
});

// drawn is markup as a page draws it, read as a document.
function drawn(markup: VNode): Document {
	return new browser.DOMParser().parseFromString(
		renderToString(markup),
		"text/html",
	) as unknown as Document;
}

// textsOf are the texts of the elements of a document a selector picks.
function textsOf(doc: Document, selector: string): string[] {
	return [...doc.querySelectorAll(selector)].map(
		(element) => element.textContent ?? "",
	);
}

describe("a chain", () => {
	test("told forward points on from each value, a step's words before its arrow", () => {
		const doc = drawn(
			<Chain start="12" steps={[{ by: "a half", to: "6" }, { to: "2" }]} />,
		);

		expect(textsOf(doc, ".s-chain > *")).toEqual([
			"12",
			"a half →",
			"6",
			"→",
			"2",
		]);
		expect(doc.querySelectorAll(".s-art-dot-accent")).toHaveLength(0);
	});

	test("told back points to the value before each, a step's words after its arrow, every value marked", () => {
		const doc = drawn(
			<Chain
				start="12"
				steps={[{ by: "× 2", to: "6" }, { to: "2" }]}
				back
				accent
			/>,
		);

		expect(textsOf(doc, ".s-chain > *")).toEqual([
			"12",
			"← × 2",
			"6",
			"←",
			"2",
		]);
		expect(
			doc.querySelectorAll(".s-chain > .s-art-dot.s-art-dot-accent"),
		).toHaveLength(3);
	});
});

describe("a tree of digits", () => {
	test("grows from each digit the numbers it begins with every other digit, in the page's numerals", () => {
		const doc = drawn(
			<DigitTree digits={[1, 2, 3]} numbers={new Intl.NumberFormat("ar-EG")} />,
		);

		expect(textsOf(doc, ".s-tree-branch > .s-art-dot")).toEqual([
			"١",
			"٢",
			"٣",
		]);
		expect(textsOf(doc, ".s-tree-leaves > .s-art-chip")).toEqual([
			"١٢",
			"١٣",
			"٢١",
			"٢٣",
			"٣١",
			"٣٢",
		]);
	});
});
