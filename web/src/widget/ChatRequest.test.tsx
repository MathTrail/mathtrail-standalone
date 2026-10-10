import { render } from "preact";
import { act } from "preact/test-utils";
import { afterEach, describe, expect, test } from "vitest";
import type { Host } from "./bridge";
import { useModelLines } from "./ChatRequest";

let root: HTMLElement | undefined;

afterEach(() => {
	if (root !== undefined) {
		act(() => render(null, root as HTMLElement));
		root.remove();
	}
	root = undefined;
});

// told draws a card that tells the model what it is given, and is what the
// host was handed each time, the card's telling function, and a way to draw
// the card again.
function told() {
	const handed: string[] = [];
	const host = {
		tellModel: (line: string) => {
			handed.push(line);
			return Promise.resolve();
		},
	} as unknown as Host;
	let tell: (line: string) => Promise<void> = () => Promise.resolve();
	function Card() {
		tell = useModelLines(host);
		return null;
	}
	root = document.createElement("div");
	document.body.append(root);
	const drawn = root;
	act(() => render(<Card />, drawn));
	return {
		handed,
		tell: (line: string) => tell(line),
		redraw: () => act(() => render(<Card />, drawn)),
	};
}

describe("the lines a card tells the model", () => {
	test("each carry the ones before it, the host keeping one line", async () => {
		const card = told();

		await card.tell("Task A has its answer recorded.");
		card.redraw();
		await card.tell("The topic of the lessons was chosen.");

		expect(card.handed).toEqual([
			"Task A has its answer recorded.",
			"Task A has its answer recorded.\n\nThe topic of the lessons was chosen.",
		]);
	});
});
