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
	test("is decoration beside the name it stands with", () => {
		draw(<Mark />);

		const mark = root.querySelector("svg");
		expect(mark?.getAttribute("aria-hidden")).toBe("true");
		expect(mark?.hasAttribute("role")).toBe(false);
		expect(mark?.getAttribute("width")).toBe("32");
	});

	test("draws the logo the site serves as its icon", () => {
		const served = new DOMParser().parseFromString(
			readFileSync(
				join(import.meta.dirname, "../../../site/assets/favicon.svg"),
				"utf8",
			),
			"image/svg+xml",
		).documentElement;
		draw(<Mark size={48} />);

		expect(drawingOf(root.querySelector("svg"))).toEqual(drawingOf(served));
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

// referencesIn is every name a drawing points at with url(#…).
function referencesIn(svg: Element): string[] {
	return [...svg.querySelectorAll("*")].flatMap((element) =>
		[...element.attributes].flatMap(({ value }) =>
			[...value.matchAll(/url\(#([^)]+)\)/g)].map((match) => match[1] ?? ""),
		),
	);
}
