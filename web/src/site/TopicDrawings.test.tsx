import { Window } from "happy-dom";
import { renderToString } from "preact-render-to-string";
import { afterAll, describe, expect, test } from "vitest";
import type { Picture } from "../design/picture/model";
import {
	drawn as drawnPicture,
	numberOf,
	textsOf as textsIn,
} from "../design/picture/testing/svg";
import { smallest } from "../design/picture/text";
import { siteData } from "./data";
import type { Drawing, Markup } from "./drawings";
import { parsePageWords } from "./pagewords";
import { openReader } from "./reader";
import { TopicDrawing } from "./TopicDrawings";
import { siteDictionaries } from "./words";

const browser = new Window();

afterAll(async () => {
	await browser.happyDOM.close();
});

// drawn is a drawing as a page in a language draws it from the words of its
// file, as a document, and the keys of the words it never read.
function drawn(drawing: Drawing, words = "title: A test", locale = "en") {
	const { page, unread } = openReader(
		parsePageWords(words),
		`${locale}/topics/test.yaml`,
		locale,
	);
	const html = renderToString(
		<TopicDrawing
			drawing={drawing}
			page={page}
			at="drawing"
			where="the drawing of the test"
		/>,
	);
	return {
		doc: new browser.DOMParser().parseFromString(
			html,
			"text/html",
		) as unknown as Document,
		unread: unread(),
	};
}

// markup is the drawing of the site's own called name.
function markup(name: Markup): Drawing {
	return { markup: name };
}

// textsOf are the texts of the elements of a document a selector picks.
function textsOf(doc: Document, selector: string): string[] {
	return [...doc.querySelectorAll(selector)].map(
		(element) => element.textContent ?? "",
	);
}

// turnOf is the share of a full turn a place of a ring stands at, in degrees.
function turnOf(element: Element): number {
	const turn = /--turn:\s*(-?[\d.]+)deg/.exec(
		element.getAttribute("style") ?? "",
	);
	if (turn?.[1] === undefined) {
		throw new Error(`no turn on ${element.outerHTML}`);
	}
	return Number(turn[1]);
}

describe("a topic's picture", () => {
	const clock: Picture = { kind: "clock", time: "11:50" };

	test("is drawn as a card draws it, in a box a screen reader passes over", () => {
		const { doc } = drawn({ picture: clock });
		const box = doc.querySelector(".s-picture");
		const svg = box?.querySelector("svg.mt-picture");

		expect(box?.getAttribute("aria-hidden")).toBe("true");
		expect(svg?.getAttribute("aria-hidden")).toBe("true");
		expect(svg?.hasAttribute("role")).toBe(false);
		expect(svg?.childElementCount).toBeGreaterThan(0);
	});

	test("that cannot be drawn stops the build, named by where it stands", () => {
		expect(() =>
			drawn(
				{ picture: { kind: "clock" } as unknown as Picture },
				undefined,
				"ru",
			),
		).toThrow("the drawing of the test: the picture cannot be drawn in ru");
	});
});

describe("a drawing of the site's own", () => {
	test("is drawn left to right, in a box a screen reader passes over", () => {
		const { doc } = drawn(markup("eggs-backwards"));
		const box = doc.querySelector(".s-art");

		expect(box?.getAttribute("aria-hidden")).toBe("true");
		expect(box?.getAttribute("dir")).toBe("ltr");
	});

	test("of islanders shows each with what they say, in the page's words", () => {
		const { doc, unread } = drawn(
			markup("islanders"),
			[
				"drawing:",
				"  first: A",
				"  first-says: “I am a knight”",
				"  second: B",
				"  second-says: “A is lying”",
			].join("\n"),
		);

		expect(textsOf(doc, ".s-speech-line")).toEqual([
			"A“I am a knight”",
			"B“A is lying”",
		]);
		expect(unread).toEqual([]);
	});

	test("of a round table seats knights and liars in turn, all the way round", () => {
		const { doc, unread } = drawn(
			markup("round-table"),
			"drawing:\n  knight: Р\n  liar: Л\n",
			"ru",
		);
		const places = [...doc.querySelectorAll(".s-ring-place")];
		const seated = places.map((place) => place.textContent);

		expect(seated).toEqual(["Р", "Л", "Р", "Л", "Р", "Л", "Р", "Л", "Р", "Л"]);
		for (const [at, seat] of seated.entries()) {
			expect(seat).not.toBe(seated[(at + 1) % seated.length]);
		}
		expect(places.map(turnOf)).toEqual([
			0, 36, 72, 108, 144, 180, 216, 252, 288, 324,
		]);
		expect(unread).toEqual([]);
	});

	test("of a cube joins a black corner and a white one with every edge", () => {
		const { doc } = drawn(markup("cube-corners"));
		const corners = new Map(
			[...doc.querySelectorAll("circle")].map((corner) => [
				`${corner.getAttribute("cx")},${corner.getAttribute("cy")}`,
				corner.classList.contains("s-cube-dark"),
			]),
		);
		const edges = [...doc.querySelectorAll("line")].map((edge) => [
			corners.get(`${edge.getAttribute("x1")},${edge.getAttribute("y1")}`),
			corners.get(`${edge.getAttribute("x2")},${edge.getAttribute("y2")}`),
		]);

		expect(corners.size).toBe(8);
		expect([...corners.values()].filter((dark) => dark)).toHaveLength(4);
		expect(edges).toHaveLength(12);
		for (const [one, other] of edges) {
			expect(one).toBeDefined();
			expect(other).toBe(one === undefined ? undefined : !one);
		}
	});

	test("of a daisy tears off two petals of twelve, opposite each other", () => {
		const { doc } = drawn(markup("daisy"));
		const petals = [...doc.querySelectorAll(".s-ring-place")];
		const torn = petals
			.filter((petal) => petal.querySelector(".s-petal-torn") !== null)
			.map(turnOf);

		expect(petals).toHaveLength(12);
		expect(torn).toHaveLength(2);
		expect(Math.abs((torn[1] ?? 0) - (torn[0] ?? 0))).toBe(180);
	});

	test("of the numbers made of 1, 2 and 3 grows each from its first digit", () => {
		const { doc } = drawn(markup("number-tree"));

		expect(textsOf(doc, ".s-tree-branch > .s-art-dot")).toEqual([
			"1",
			"2",
			"3",
		]);
		expect(textsOf(doc, ".s-tree-leaves > .s-art-chip")).toEqual([
			"12",
			"13",
			"21",
			"23",
			"31",
			"32",
		]);
	});

	test.each([
		["eggs-backwards", ["7", "3", "1", "0"], ["→", "→", "→"]],
		["price-changes", ["400", "500", "400"], ["+25% →", "−20% →"]],
	] as const)("of %s reads the solution's chain", (name, values, steps) => {
		const { doc } = drawn(markup(name));

		expect(textsOf(doc, ".s-chain > .s-art-dot")).toEqual(values);
		expect(textsOf(doc, ".s-chain > .s-chain-step")).toEqual(steps);
	});

	test("of a time through the hour writes its minutes in the page's words", () => {
		const { doc, unread } = drawn(
			markup("through-the-hour"),
			'drawing:\n  step: "+{count} мин"\n',
			"ru",
		);

		expect(textsOf(doc, ".s-chain > .s-art-dot")).toEqual([
			"4:50",
			"5:00",
			"5:15",
		]);
		expect(textsOf(doc, ".s-chain > .s-chain-step")).toEqual([
			"+10 мин →",
			"+15 мин →",
		]);
		expect(unread).toEqual([]);
	});

	test("of a strip cut in two cuts it every way into two pieces of four that hold together", () => {
		const { doc } = drawn(markup("strip-cuts"));
		const cuts = [...doc.querySelectorAll(".s-cut")].map((cut) =>
			[...cut.querySelectorAll(".s-cut-cell")].map((cell) =>
				cell.classList.contains("s-cut-first"),
			),
		);

		expect(cuts).toHaveLength(4);
		const seen = new Set<string>();
		for (const cells of cuts) {
			expect(cells).toHaveLength(8);
			for (const piece of [true, false]) {
				const taken = cells.flatMap((first, at) =>
					first === piece ? [at] : [],
				);
				expect(taken).toHaveLength(4);
				expect(holdsTogether(taken)).toBe(true);
			}
			seen.add(cells.map((first) => (first ? "#" : ".")).join(""));
			seen.add(cells.map((first) => (first ? "." : "#")).join(""));
		}
		expect(seen.size).toBe(2 * cuts.length);
	});

	test("of flags of two stripes draws each pair of different colours once", () => {
		const { doc } = drawn(markup("two-stripe-flags"));
		const flags = [...doc.querySelectorAll(".s-flag-pair")].map((flag) =>
			[...flag.querySelectorAll(".s-flag-stripe")].map(
				(stripe) =>
					[...stripe.classList].find((name) => name !== "s-flag-stripe") ?? "",
			),
		);

		expect(flags).toHaveLength(6);
		for (const [top, bottom] of flags) {
			expect(top).not.toBe(bottom);
		}
		expect(new Set(flags.map((flag) => flag.join("/"))).size).toBe(6);
	});

	test("of sum-regrouped marks the numbers that make a hundred", () => {
		const { doc } = drawn(markup("sum-regrouped"));

		expect(textsOf(doc, ".s-art-chip-accent")).toEqual(["100", "100"]);
	});
});

// holdsTogether says whether cells of a strip of 2 by 4, counted row by row,
// are joined side by side into one piece.
function holdsTogether(cells: readonly number[]): boolean {
	const left = new Set(cells.slice(1));
	const reached = [cells[0]];
	for (let next = reached.pop(); next !== undefined; next = reached.pop()) {
		const row = Math.floor(next / 4);
		const column = next % 4;
		for (const [r, c] of [
			[row - 1, column],
			[row + 1, column],
			[row, column - 1],
			[row, column + 1],
		] as const) {
			const cell = r * 4 + c;
			if (r >= 0 && r < 2 && c >= 0 && c < 4 && left.delete(cell)) {
				reached.push(cell);
			}
		}
	}
	return left.size === 0;
}

// rooms are the widths a topic's picture has where it stands, on the screen
// that gives it the least. A card: the narrowest of four in a row, 270 px,
// less its padding of 20 px on each side. The first screen and an example: a
// phone 320 px wide, less the page's margin of 16 px on each side, the
// picture reaching across the padding of the panel or the example around it.
const rooms = { card: 230, hero: 288, example: 288 } as const;

// Placed is a picture of the site's topics, where it stands and whose it is.
type Placed = readonly [place: keyof typeof rooms, whose: string, Picture];

// placed are the pictures of the site's topics: their cards', their first
// screens' and their examples'.
const placed: readonly Placed[] = (() => {
	const data = siteData();
	const pictureOf = (drawing: Drawing | undefined) =>
		drawing !== undefined && "picture" in drawing ? [drawing.picture] : [];
	return [
		...[...data.drawings].flatMap(([id, { card, hero }]) => [
			...pictureOf(card).map((picture): Placed => ["card", id, picture]),
			...pictureOf(hero).map((picture): Placed => ["hero", id, picture]),
		]),
		...[...data.examples].flatMap(([id, examples]) =>
			examples.flatMap((example, at) =>
				pictureOf(example.drawing).map(
					(picture): Placed => ["example", `${id} ${at + 1}`, picture],
				),
			),
		),
	];
})();

describe("the pictures of the site's topics", () => {
	// No topic's card draws a picture of a kind any more: each card draws a
	// markup of the site's own. A card keeps its room, for a picture one may
	// draw again.
	test("stand on the first screens and in the examples", () => {
		expect(new Set(placed.map(([place]) => place))).toEqual(
			new Set(["hero", "example"]),
		);
	});

	test.each(
		[...siteDictionaries.keys()].flatMap((locale) =>
			placed.map(
				([place, whose, picture]) => [place, whose, locale, picture] as const,
			),
		),
	)(
		"draws the %s of %s in %s in its room, no label smaller than the smallest size",
		(place, _, locale, picture) => {
			const drawing = drawnPicture(picture, locale);
			const width = Number(drawing.root.width);
			const shrunk = place === "card" ? Math.min(1, rooms.card / width) : 1;
			const sizes = textsIn(drawing).map((text) => numberOf(text, "font-size"));

			if (place !== "card") {
				expect(width).toBeLessThanOrEqual(rooms[place]);
			}
			for (const size of sizes) {
				expect(size * shrunk).toBeGreaterThanOrEqual(smallest);
			}
		},
	);
});
