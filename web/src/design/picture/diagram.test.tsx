import { readFileSync } from "node:fs";
import { join } from "node:path";
import { render } from "preact";
import { act } from "preact/test-utils";
import { afterEach, describe, expect, test, vi } from "vitest";
import { Diagram } from "./diagram";
import { kinds, type Picture } from "./model";
import { extremes } from "./testing/extremes";
import {
	boxOfShape,
	drawn,
	numberOf,
	overlap,
	pointsOfPath,
	textBox,
	textsOf,
} from "./testing/svg";
import { clearance, room, smallest } from "./text";

afterEach(() => {
	vi.restoreAllMocks();
});

// examples are the example of each kind the content shows the model.
const examples = Object.values(
	import.meta.glob<{ picture: Picture }>(
		"../../../../content/pictures/*.json",
		{
			eager: true,
			import: "default",
		},
	),
).map(
	(example) =>
		[`the example of ${example.picture.kind}`, example.picture] as const,
);

// references are the pictures the reference tasks carry.
const references = Object.values(
	import.meta.glob<readonly { id: string; picture?: Picture }[]>(
		"../../../../content/examples/*.json",
		{
			eager: true,
			import: "default",
		},
	),
).flatMap((tasks) =>
	tasks.flatMap((task) =>
		task.picture === undefined
			? []
			: [[`the picture of ${task.id}`, task.picture] as const],
	),
);

// every is every picture a test of drawing reads: the examples, the reference
// pictures, and each kind at its limits.
const every = [
	...examples,
	...references,
	...extremes.map((one) => [one.name, one.picture] as const),
];

// smaller are the pictures whose labels may be drawn smaller than the smallest
// size: the widest labels in a table of many columns and in a grid of many
// columns, which fit the card at no larger size.
const smaller = new Set([
	"table of six columns of the widest labels",
	"grid of eight rows and columns with wide names, filled and marked",
]);

describe("a picture", () => {
	test("is drawn for every kind the format has", () => {
		const drawnKinds = new Set(examples.map(([, picture]) => picture.kind));

		expect([...drawnKinds].sort()).toEqual([...kinds].sort());
	});

	test.each(every)(
		"%s is an image a screen reader is told of, laid out left to right",
		(_, picture) => {
			const { root } = drawn(picture);

			expect(root.role).toBe("img");
			expect(root["aria-label"]).toBe("A picture");
			expect(root.direction).toBe("ltr");
			expect(root.class).toBe("mt-picture");
			expect(root.viewBox).toBe(
				`${-clearance} ${-clearance} ${root.width} ${root.height}`,
			);
			expect(Number(root.width)).toBeLessThanOrEqual(room + 2 * clearance);
		},
	);

	test.each(every)(
		"%s names no colour, no style and nothing a page could collide with",
		(_, picture) => {
			const { root, shapes } = drawn(picture);

			for (const attributes of [
				root,
				...shapes.map((shape) => shape.attributes),
			]) {
				expect(Object.keys(attributes)).not.toContain("style");
				expect(Object.keys(attributes)).not.toContain("id");
				expect(Object.keys(attributes)).not.toContain("dir");
				expect(Object.keys(attributes)).not.toContain("tabindex");
				expect(Object.keys(attributes)).not.toContain("fill");
				expect(Object.keys(attributes)).not.toContain("stroke");
				for (const value of Object.values(attributes)) {
					expect(value).not.toMatch(/#[0-9a-f]{3}|rgb\(|hsl\(|url\(/i);
				}
			}
			expect(shapes.map((shape) => shape.tag)).not.toContain("defs");
		},
	);

	test.each(every)(
		"%s stands within its frame, every number of it a number",
		(_, picture) => {
			const { root, shapes } = drawn(picture);
			const width = Number(root.width) - 2 * clearance;
			const height = Number(root.height) - 2 * clearance;
			const near = 0.05;

			for (const shape of shapes) {
				const box = boxOfShape(shape);
				if (box === undefined) {
					continue;
				}
				const where = `${shape.tag} ${JSON.stringify(shape.attributes)} ${shape.text}`;
				for (const value of Object.values(box)) {
					expect(Number.isFinite(value), where).toBe(true);
				}
				const reach = shape.tag === "text" ? clearance : 0;
				expect(box.left, where).toBeGreaterThanOrEqual(-reach - near);
				expect(box.top, where).toBeGreaterThanOrEqual(-reach - near);
				expect(box.right, where).toBeLessThanOrEqual(width + reach + near);
				expect(box.bottom, where).toBeLessThanOrEqual(height + reach + near);
			}
		},
	);

	test.each(every)(
		"%s writes its words apart from each other",
		(_, picture) => {
			const texts = textsOf(drawn(picture));

			for (const [at, one] of texts.entries()) {
				for (const other of texts.slice(at + 1)) {
					expect(
						overlap(textBox(one), textBox(other)),
						`${one.text} and ${other.text}`,
					).toBe(false);
				}
			}
		},
	);

	test.each(every)(
		"%s writes no word smaller than the smallest size",
		(name, picture) => {
			const sizes = textsOf(drawn(picture)).map((text) =>
				numberOf(text, "font-size"),
			);

			if (smaller.has(name)) {
				expect(Math.min(...sizes)).toBeLessThan(smallest);
			} else {
				expect(Math.min(...sizes)).toBeGreaterThanOrEqual(smallest);
			}
		},
	);

	test("that cannot be drawn is left out, and why is told once", () => {
		const told = vi.spyOn(console, "error").mockImplementation(() => {});
		const container = document.createElement("div");

		act(() =>
			render(
				<Diagram
					picture={{ kind: "clock" } as unknown as Picture}
					label="A picture"
					locale="en"
				/>,
				container,
			),
		);

		expect(container.querySelector("svg")).toBeNull();
		expect(told).toHaveBeenCalledOnce();
		expect(told.mock.calls[0]?.[0]).toBe(
			"widget: the picture could not be drawn",
		);
		act(() => render(null, container));
	});

	test("is drawn in no shade the themes do not each give", () => {
		const read = (path: string) =>
			readFileSync(join(import.meta.dirname, path), "utf8");
		const styles = read("../mathtrail.css");
		const tokens = read("../../../../internal/widget/tokens.css");
		const used = new Set(
			[
				...styles.matchAll(/\.mt-pic(?:ture|-[\w-]+)[^{}]*\{([^}]*)\}/g),
			].flatMap((rule) =>
				[
					...(rule[1] ?? "").matchAll(
						/(?:fill|stroke|color)\s*:\s*var\((--[\w-]+)\)/g,
					),
				].map((one) => one[1] ?? ""),
			),
		);
		const blocks = tokens
			.split(/\n(?=[:[@])/)
			.filter((block) => block.includes("--surface:"));

		expect(used.size).toBeGreaterThan(0);
		expect(blocks).toHaveLength(3);
		for (const name of used) {
			for (const block of blocks) {
				expect(block, name).toContain(`${name}:`);
			}
		}
	});
});

describe("the points of a path", () => {
	test("are read from every command a picture draws with", () => {
		expect(pointsOfPath("M1 2L3 4H5V6Q7 8 9 10Z")).toEqual([
			{ x: 1, y: 2 },
			{ x: 3, y: 4 },
			{ x: 5, y: 4 },
			{ x: 5, y: 6 },
			{ x: 7, y: 8 },
			{ x: 9, y: 10 },
		]);
	});

	test("are not read from a command a picture does not draw with", () => {
		expect(() => pointsOfPath("M1 2h3")).toThrow(/h/);
	});
});
