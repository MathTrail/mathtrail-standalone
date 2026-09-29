import { render } from "preact";
import { act } from "preact/test-utils";
import { afterEach, describe, expect, test } from "vitest";
import { Avatar, Icon, type IconName, Mark } from "./icons";

const root = document.createElement("div");

afterEach(() => {
	act(() => render(null, root));
});

function draw(element: preact.JSX.Element): void {
	act(() => render(element, root));
}

describe("an icon", () => {
	test.each<IconName>([
		"chevron-right",
		"chevron-left",
		"send",
		"check",
		"cross",
		"hint",
		"trap",
		"spinner",
		"verdict-correct",
		"verdict-wrong",
		"step-done",
	])(
		"%s is hidden from a screen reader and drawn with SVG's own attributes",
		(name) => {
			draw(<Icon name={name} />);

			const icon = root.querySelector("svg");
			expect(icon?.getAttribute("aria-hidden")).toBe("true");
			const stroked = icon?.querySelector("path");
			// SVG names its attributes with hyphens; a camel-cased one would be an
			// attribute nothing reads, and the icon a hairline.
			expect(stroked?.getAttribute("stroke-width")).not.toBeNull();
			expect(stroked?.getAttribute("stroke-linecap")).toBe("round");
			expect(stroked?.hasAttribute("strokeWidth")).toBe(false);
		},
	);

	test("spinner's track takes the colour it is given", () => {
		draw(<Icon name="spinner" track="accent-tint" />);

		expect(root.querySelector("circle")?.getAttribute("stroke")).toBe(
			"var(--accent-tint)",
		);
		expect(root.querySelector("svg")?.getAttribute("class")).toBe(
			"mt-icon mt-spin",
		);
	});

	test("step-waiting is an empty ring, drawn with SVG's own attributes", () => {
		draw(<Icon name="step-waiting" />);

		const ring = root.querySelector("svg circle");
		expect(root.querySelector("svg")?.getAttribute("aria-hidden")).toBe("true");
		expect(ring?.getAttribute("stroke-width")).toBe("1.5");
		expect(ring?.hasAttribute("strokeWidth")).toBe(false);
		expect(root.querySelector("svg path")).toBeNull();
	});

	test("step-done is ticked in the colour of the strongest ink", () => {
		draw(<Icon name="step-done" />);

		expect(root.querySelector("svg circle")?.getAttribute("fill")).toBe(
			"var(--ink-strong)",
		);
		expect(root.querySelector("svg path")?.getAttribute("stroke")).toBe(
			"var(--surface)",
		);
	});
});

test("the mark is decoration beside the name it stands with", () => {
	draw(<Mark />);

	const mark = root.querySelector("svg");
	expect(mark?.getAttribute("aria-hidden")).toBe("true");
	expect(mark?.hasAttribute("role")).toBe(false);
});

test("the avatar is decoration", () => {
	draw(<Avatar size={32} />);

	const avatar = root.querySelector("svg");
	expect(avatar?.getAttribute("aria-hidden")).toBe("true");
	expect(avatar?.getAttribute("width")).toBe("32");
});
