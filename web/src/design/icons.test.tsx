import { readFileSync } from "node:fs";
import { join } from "node:path";
import { render } from "preact";
import { act } from "preact/test-utils";
import { afterEach, describe, expect, test } from "vitest";
import { Icon, type IconName, Mark } from "./icons";
import { drawingOf } from "./testing/drawing";

const root = document.createElement("div");

afterEach(() => {
	act(() => render(null, root));
});

function draw(element: preact.JSX.Element): void {
	act(() => render(element, root));
}

// colours is the colour each icon is drawn in: the text's, or the token of
// what the icon marks. A new icon cannot be added without a line here.
const colours: Record<IconName, string> = {
	"chevron-right": "currentColor",
	"chevron-left": "currentColor",
	check: "currentColor",
	cross: "currentColor",
	dash: "currentColor",
	hint: "currentColor",
	trap: "currentColor",
	menu: "currentColor",
	chat: "currentColor",
	renew: "currentColor",
	sparkle: "currentColor",
	tag: "currentColor",
	spinner: "currentColor",
	"verdict-correct": "var(--correct)",
	"verdict-wrong": "var(--wrong)",
	"step-done": "var(--ink-strong)",
	"step-waiting": "var(--track)",
	avatar: "var(--text-muted)",
};

describe("an icon", () => {
	test.each(Object.keys(colours) as IconName[])(
		"%s is a line of the set's width on a grid of 24, in one colour, hidden from a screen reader",
		(name) => {
			draw(<Icon name={name} size={20} />);

			const icon = root.querySelector("svg");
			expect(icon?.getAttribute("aria-hidden")).toBe("true");
			expect(icon?.getAttribute("width")).toBe("20");
			expect(icon?.getAttribute("viewBox")).toBe("0 0 24 24");
			expect(icon?.getAttribute("fill")).toBe("none");
			const lines = [...(icon?.children ?? [])];
			expect(lines.length).toBeGreaterThan(0);
			for (const line of lines) {
				expect(line.tagName.toLowerCase()).toBe("path");
				// SVG names its attributes with hyphens; a camel-cased one would
				// be an attribute nothing reads, and the icon a hairline.
				expect(line.getAttribute("stroke-width")).toBe("2.5");
				expect(line.getAttribute("stroke-linecap")).toBe("round");
				expect(line.getAttribute("stroke-linejoin")).toBe("round");
				expect(line.hasAttribute("strokeWidth")).toBe(false);
				expect(line.hasAttribute("fill")).toBe(false);
				expect(line.hasAttribute("style")).toBe(false);
			}
			expect(lines.at(-1)?.getAttribute("stroke")).toBe(colours[name]);
		},
	);

	test("spinner turns its arc on a track of the colour it is given", () => {
		draw(<Icon name="spinner" track="accent-tint" />);

		const [track, arc] = root.querySelectorAll("svg path");
		expect(track?.getAttribute("stroke")).toBe("var(--accent-tint)");
		expect(arc?.getAttribute("stroke")).toBe("currentColor");
		expect(root.querySelector("svg")?.getAttribute("class")).toBe(
			"mt-icon mt-spin",
		);
	});

	test("takes the classes it is given beside its own", () => {
		draw(<Icon name="chevron-right" className="mt-chevron" />);

		expect(root.querySelector("svg")?.getAttribute("class")).toBe(
			"mt-icon mt-chevron",
		);
	});
});

describe("the mark", () => {
	test("is decoration beside the name it stands with, drawn once for each theme", () => {
		draw(<Mark />);

		const marks = [...root.querySelectorAll("svg")];
		expect(marks.map((mark) => mark.getAttribute("class"))).toEqual([
			"mt-icon mt-mark-light",
			"mt-icon mt-mark-dark",
		]);
		for (const mark of marks) {
			expect(mark.getAttribute("aria-hidden")).toBe("true");
			expect(mark.hasAttribute("role")).toBe(false);
			expect(mark.getAttribute("width")).toBe("32");
		}
	});

	test("draws for the dark theme the logo the site serves as its icon", () => {
		const served = new DOMParser().parseFromString(
			readFileSync(
				join(import.meta.dirname, "../../../site/assets/favicon.svg"),
				"utf8",
			),
			"image/svg+xml",
		).documentElement;
		draw(<Mark size={48} />);

		expect(drawingOf(root.querySelector(".mt-mark-dark"))).toEqual(
			drawingOf(served),
		);
	});

	test("draws for the light theme the same ribbon and star on a white tile, its hairline all within the square", () => {
		draw(<Mark size={48} />);
		const light = root.querySelector(".mt-mark-light");
		const tile = light?.querySelector(":scope > rect");
		const at = (name: string) => Number(tile?.getAttribute(name));

		expect(shapesOf(light)).toEqual(
			shapesOf(root.querySelector(".mt-mark-dark")),
		);
		expect(tile?.getAttribute("fill")).toBe("#ffffff");
		expect(tile?.getAttribute("stroke")).toMatch(/^#[0-9a-f]{6}$/);
		// A line is drawn on its shape's edge, half of it outside: the tile
		// stands that far in from the square, so all of the hairline shows.
		const half = at("stroke-width") / 2;
		expect(half).toBeGreaterThan(0);
		expect([at("x"), at("y")]).toEqual([half, half]);
		expect([
			at("x") + at("width") + half,
			at("y") + at("height") + half,
		]).toEqual([48, 48]);
	});

	test("lightens no ink of the site's icon for the light theme, and deepens its palest, so that they hold on a white tile", () => {
		draw(<Mark size={48} />);
		const light = inksOf(root.querySelector(".mt-mark-light"));
		const dark = inksOf(root.querySelector(".mt-mark-dark"));

		expect(light).toHaveLength(dark.length);
		light.forEach((ink, at) => {
			expect(luminance(ink)).toBeLessThanOrEqual(luminance(dark[at] ?? ""));
		});
		expect(Math.max(...light.map(luminance))).toBeLessThan(
			Math.max(...dark.map(luminance)),
		);
	});

	test("names its gradients and its clip apart from another mark's", () => {
		draw(
			<>
				<Mark />
				<Mark />
			</>,
		);

		const ids = [...root.querySelectorAll("[id]")].map((node) => node.id);
		expect(ids.length).toBeGreaterThan(0);
		expect(new Set(ids).size).toBe(ids.length);
		for (const node of root.querySelectorAll("svg")) {
			for (const reference of referencesIn(node)) {
				expect(node.querySelector(`[id="${reference}"]`)).not.toBeNull();
			}
		}
	});
});

// shapesOf are the lines a drawing of the mark draws its ribbon and its star
// with, in their order.
function shapesOf(svg: Element | null): (string | null)[] {
	return [...(svg?.querySelectorAll("path") ?? [])].map((path) =>
		path.getAttribute("d"),
	);
}

// inksOf are the colours a drawing of the mark runs its parts through, from
// the top and the bottom of each, in their order.
function inksOf(svg: Element | null): string[] {
	return [...(svg?.querySelectorAll("stop") ?? [])].map(
		(stop) => stop.getAttribute("stop-color") ?? "",
	);
}

// luminance is how light a colour written #rrggbb is to the eye, from 0 for
// black to 1 for white.
function luminance(colour: string): number {
	const [red = 0, green = 0, blue = 0] = [1, 3, 5].map((at) => {
		const channel = Number.parseInt(colour.slice(at, at + 2), 16) / 255;
		return channel <= 0.04045
			? channel / 12.92
			: ((channel + 0.055) / 1.055) ** 2.4;
	});
	return 0.2126 * red + 0.7152 * green + 0.0722 * blue;
}

// referencesIn is every name a drawing points at with url(#…).
function referencesIn(svg: Element): string[] {
	return [...svg.querySelectorAll("*")].flatMap((element) =>
		[...element.attributes].flatMap(({ value }) =>
			[...value.matchAll(/url\(#([^)]+)\)/g)].map((match) => match[1] ?? ""),
		),
	);
}
