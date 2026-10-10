import { readFileSync } from "node:fs";
import { join } from "node:path";
import { render } from "preact";
import { act } from "preact/test-utils";
import { describe, expect, test } from "vitest";
import { Diagram } from "./diagram";
import { type Flags, paints } from "./model";
import { drawn, type Shape, shapesOf, textsOf } from "./testing/svg";
import { room } from "./text";

// threeColours are the flags of two stripes in red, blue and yellow, gathered
// by their top stripe, each group named by that colour.
const threeColours: Flags = {
	kind: "flags",
	colors: { red: "red", blue: "blue", yellow: "yellow" },
	groups: (["red", "blue", "yellow"] as const).map((top) => ({
		color: top,
		flags: (["red", "blue", "yellow"] as const)
			.filter((bottom) => bottom !== top)
			.map((bottom) => [top, bottom]),
	})),
};

// paintedIn are the paints a drawing's stripes and swatches are painted in,
// read off the class of each, in the order they are drawn.
const paintedIn = (flags: Flags) =>
	drawn(flags)
		.shapes.map((shape) => shape.attributes.class ?? "")
		.filter((name) => name.startsWith("mt-paint mt-paint-"))
		.map((name) => name.replace("mt-paint mt-paint-", ""));

describe("flags", () => {
	test("paint each stripe from the top, each group's colour under it, and write a paint's letters on it", () => {
		const drawing = drawn(threeColours);

		expect(paintedIn(threeColours)).toEqual([
			"red",
			"blue",
			"red",
			"yellow",
			"red",
			"blue",
			"red",
			"blue",
			"yellow",
			"blue",
			"yellow",
			"red",
			"yellow",
			"blue",
			"yellow",
		]);
		expect(textsOf(drawing).map((text) => text.text)).toEqual([
			"R",
			"B",
			"R",
			"Y",
			"R",
			"B",
			"R",
			"B",
			"Y",
			"B",
			"Y",
			"R",
			"Y",
			"B",
			"Y",
		]);
	});

	test("leave a stripe whose colour is the unknown in the card's colour, with a ?", () => {
		const drawing = drawn({
			kind: "flags",
			colors: { red: "red" },
			groups: [{ label: "A", flags: [["red", "?"]] }],
		});

		expect(shapesOf(drawing, "rect", "mt-pic-plate")).toHaveLength(1);
		expect(textsOf(drawing).map((text) => text.text)).toEqual(["R", "?", "A"]);
	});

	test("outline a colour that names a group, so that a paint of the card's own colour shows", () => {
		const drawing = drawn({
			kind: "flags",
			colors: { white: "white", black: "black" },
			groups: [
				{ color: "white", flags: [["white", "black"]] },
				{ color: "black", flags: [["black", "white"]] },
			],
		});
		const rects = shapesOf(drawing, "rect");
		const boxOf = ({ attributes: { x, y, width, height, rx } }: Shape) =>
			[x, y, width, height, rx].join(" ");
		const swatches = rects.filter(
			(shape) =>
				shape.attributes.rx !== undefined &&
				shape.attributes.class?.startsWith("mt-paint ") === true,
		);
		const outlineOf = (swatch: Shape) =>
			rects
				.filter(
					(shape) =>
						shape.attributes.class === "mt-pic-line" &&
						boxOf(shape) === boxOf(swatch),
				)
				.map((shape) => shape.attributes["stroke-width"]);

		expect(swatches.map((swatch) => swatch.attributes.class)).toEqual([
			"mt-paint mt-paint-white",
			"mt-paint mt-paint-black",
		]);
		for (const swatch of swatches) {
			expect(outlineOf(swatch)).toEqual(["1"]);
		}
	});

	test("set as many groups side by side as fit the card, and the rest under them", () => {
		const six = (["red", "blue"] as const).map((color) => ({
			color,
			flags: Array.from(
				{ length: 6 },
				() => ["red", "blue"] as ("red" | "blue")[],
			),
		}));
		const two = drawn({
			kind: "flags",
			colors: { red: "red", blue: "blue" },
			groups: six,
		});
		const one = drawn({
			kind: "flags",
			colors: { red: "red", blue: "blue" },
			groups: six.slice(0, 1),
		});

		expect(Number(two.root.width)).toBeLessThanOrEqual(room + 4);
		expect(Number(two.root.height)).toBeGreaterThan(
			2 * Number(one.root.height) - 4,
		);
	});

	test("draw a stripe taller where the letters on it reach higher and lower", () => {
		const stripeOf = (colors: Flags["colors"]) =>
			Number(
				shapesOf(
					drawn({
						kind: "flags",
						colors,
						groups: [{ flags: [["red", "blue"]] }],
					}),
					"rect",
					"mt-paint mt-paint-red",
				)[0]?.attributes.height,
			);

		expect(stripeOf({ red: "red", blue: "blue" })).toBe(15);
		expect(stripeOf({ red: "красная", blue: "синяя" })).toBe(15);
		expect(stripeOf({ red: "赤", blue: "青" })).toBe(15);
		expect(stripeOf({ red: "أحمر", blue: "أزرق" })).toBe(19);
		expect(stripeOf({ red: "लाल", blue: "blue" })).toBe(19);
	});

	test("are painted in the palette's colours, which the tokens give once for every theme", () => {
		const read = (path: string) =>
			readFileSync(join(import.meta.dirname, path), "utf8");
		const styles = read("../mathtrail.css");
		const tokens = read("../../../../internal/widget/tokens.css");
		const shared = tokens.slice(tokens.indexOf("\n:root {"));

		for (const paint of paints) {
			expect(styles).toContain(
				`.mt-paint-${paint} {\n\tfill: var(--paint-${paint});`,
			);
			expect(styles).toContain(
				`.mt-paint-ink-${paint} {\n\tfill: var(--paint-${paint}-ink);`,
			);
			expect(shared).toContain(`--paint-${paint}:`);
			expect(shared).toContain(`--paint-${paint}-ink:`);
		}
	});
});

describe("the key under a picture that paints", () => {
	// keyOf is the key a picture is drawn with in a language, each colour as
	// its swatch's letters and its word, and what the key says of itself.
	function keyOf(flags: Flags, locale = "en") {
		const container = document.createElement("div");
		act(() =>
			render(
				<Diagram
					picture={flags}
					label="Flags"
					locale={locale}
					keyLabel="Colours"
					said={{ lang: locale, dir: locale === "ar" ? "rtl" : "ltr" }}
				/>,
				container,
			),
		);
		const list = container.querySelector("ul.mt-pic-key");
		const read = {
			label: list?.getAttribute("aria-label"),
			lang: list?.getAttribute("lang"),
			dir: list?.getAttribute("dir"),
			entries: [...(list?.querySelectorAll("li") ?? [])].map((entry) => [
				entry.querySelector(".mt-pic-swatch")?.textContent,
				entry.lastElementChild?.textContent,
			]),
			hidden: [...(list?.querySelectorAll(".mt-pic-swatch") ?? [])].every(
				(swatch) => swatch.getAttribute("aria-hidden") === "true",
			),
		};
		act(() => render(null, container));
		return read;
	}

	test("lists each colour's paint with its letters beside its word, in the palette's order", () => {
		expect(keyOf(threeColours)).toEqual({
			label: "Colours",
			lang: "en",
			dir: "ltr",
			entries: [
				["R", "red"],
				["Y", "yellow"],
				["B", "blue"],
			],
			hidden: true,
		});
	});

	test("says its words in the task's language, which runs its way", () => {
		const read = keyOf(
			{
				kind: "flags",
				colors: { red: "أحمر", blue: "أزرق" },
				groups: [{ flags: [["red", "blue"]] }],
			},
			"ar",
		);

		expect(read.lang).toBe("ar");
		expect(read.dir).toBe("rtl");
		expect(read.entries).toEqual([
			["أح", "أحمر"],
			["أز", "أزرق"],
		]);
	});

	test("is not drawn for a picture that paints nothing", () => {
		expect(
			keyOf({ kind: "flags", groups: [{ flags: [["?", "?"]] }] }).entries,
		).toEqual([]);
	});
});
