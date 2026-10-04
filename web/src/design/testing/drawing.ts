import { render } from "preact";
import { act } from "preact/test-utils";

/**
 * drawnAlone is the markup element draws on its own, in an element of its own
 * that is taken down again once it is read.
 */
export function drawnAlone(element: preact.JSX.Element): string {
	return alone(element, (container) => container.innerHTML);
}

/**
 * drawingAlone is what the SVG element draws on its own, as drawingOf reads
 * it.
 */
export function drawingAlone(element: preact.JSX.Element): string[] {
	return alone(element, (container) =>
		drawingOf(container.querySelector("svg")),
	);
}

// alone draws element in an element of its own, reads it, and takes it down.
function alone<T>(
	element: preact.JSX.Element,
	read: (container: HTMLElement) => T,
): T {
	const container = document.createElement("div");
	act(() => render(element, container));
	try {
		return read(container);
	} finally {
		act(() => render(null, container));
	}
}

// placeDecides are the attributes of a drawing's own element that the place it
// is shown in decides: its size, its class and its words for a screen reader.
const placeDecides = new Set([
	"width",
	"height",
	"class",
	"role",
	"aria-hidden",
	"aria-label",
	"xmlns",
]);

/**
 * drawingOf is what an SVG draws, element by element: each element's name and
 * its attributes in order of name, with the names of its gradients and clips
 * replaced by their place among them, since a page names them as it needs.
 */
export function drawingOf(svg: Element | null): string[] {
	if (svg === null) {
		return [];
	}
	const elements = [svg, ...svg.querySelectorAll("*")];
	const places = new Map(
		elements
			.filter((element) => element.id !== "")
			.map((element, place) => [element.id, `#${place}`]),
	);
	const placed = (value: string) =>
		value.replace(
			/url\(#([^)]+)\)/g,
			(_, id: string) => `url(${places.get(id) ?? id})`,
		);
	return elements.map((element) => {
		const attributes = [...element.attributes]
			.filter(({ name }) => element !== svg || !placeDecides.has(name))
			.map(
				({ name, value }) =>
					`${name}=${name === "id" ? places.get(value) : placed(value)}`,
			)
			.sort();
		return `${element.tagName.toLowerCase()} ${attributes.join(" ")}`;
	});
}
