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
});

describe("the mark", () => {
	test("beside a name is decoration", () => {
		draw(<Mark />);

		const mark = root.querySelector("svg");
		expect(mark?.getAttribute("aria-hidden")).toBe("true");
		expect(mark?.hasAttribute("role")).toBe(false);
	});

	test("alone is an image with a name", () => {
		draw(<Mark label="MathTrail" />);

		const mark = root.querySelector("svg");
		expect(mark?.getAttribute("role")).toBe("img");
		expect(mark?.getAttribute("aria-label")).toBe("MathTrail");
		expect(mark?.hasAttribute("aria-hidden")).toBe(false);
	});
});

test("the avatar is decoration", () => {
	draw(<Avatar size={32} />);

	const avatar = root.querySelector("svg");
	expect(avatar?.getAttribute("aria-hidden")).toBe("true");
	expect(avatar?.getAttribute("width")).toBe("32");
});
