import { render } from "preact";
import { act } from "preact/test-utils";
import { afterEach, describe, expect, test } from "vitest";
import { DayOver } from "./night";

const root = document.createElement("div");

afterEach(() => {
	act(() => render(null, root));
});

function draw(element: preact.JSX.Element): void {
	act(() => render(element, root));
}

const said = {
	title: "Make a wish for tomorrow",
	detail: "New ones come tomorrow.",
};

describe("DayOver", () => {
	test("says the wish, then what happened, under its night", () => {
		draw(<DayOver {...said} />);

		const parts = [...root.querySelectorAll(".mt-dayover > *")];
		expect(parts.map((part) => part.className)).toEqual([
			"mt-night",
			"mt-dayover-words",
		]);
		const title = root.querySelector(".mt-dayover-title");
		expect(title?.tagName).toBe("H2");
		expect(title?.textContent).toBe(said.title);
		expect(root.querySelector(".mt-dayover-line")?.textContent).toBe(
			said.detail,
		);
	});

	test("draws its night for the eye alone", () => {
		draw(<DayOver {...said} />);

		const night = root.querySelector(".mt-night svg");
		expect(night?.getAttribute("aria-hidden")).toBe("true");
		expect(night?.getAttribute("viewBox")).toBe("0 0 592 132");
		expect(root.querySelector(".mt-night")?.textContent).toBe("");
	});

	test("names the glow of each night apart from another's on the page", () => {
		draw(
			<>
				<DayOver {...said} />
				<DayOver {...said} />
			</>,
		);

		const nights = [...root.querySelectorAll(".mt-night svg")];
		const glows = nights.map(
			(night) => night.querySelector("radialGradient")?.id ?? "",
		);
		expect(new Set(glows).size).toBe(2);
		nights.forEach((night, i) => {
			const lit = [...night.querySelectorAll("circle")].filter(
				(circle) => circle.getAttribute("fill") === `url(#${glows[i]})`,
			);
			expect(lit).toHaveLength(1);
		});
	});
});
