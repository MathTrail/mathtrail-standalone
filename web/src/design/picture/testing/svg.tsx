import { render } from "preact";
import { act } from "preact/test-utils";
import { Diagram } from "../diagram";
import type { Picture } from "../model";
import { widthOf } from "../text";

/** Shape is one element a picture draws: its tag, its attributes and its text. */
export type Shape = {
	tag: string;
	attributes: Record<string, string>;
	text: string;
};

/** Drawing is a picture as the card draws it: its own element, and every one inside it. */
export type Drawing = { root: Record<string, string>; shapes: Shape[] };

/** Box is the room something takes: its left, top, right and bottom. */
export type Box = { left: number; top: number; right: number; bottom: number };

/** Point is a point of a picture. */
export type Point = { x: number; y: number };

// attributesOf are an element's attributes by their names.
function attributesOf(element: Element): Record<string, string> {
	return Object.fromEntries(
		[...element.attributes].map((one) => [one.name, one.value]),
	);
}

/**
 * drawingOf is how the card draws a picture, in a language, read from the page
 * it is drawn on, or undefined where it draws none.
 */
export function drawingOf(
	picture: Picture,
	locale = "en",
): Drawing | undefined {
	const container = document.createElement("div");
	act(() =>
		render(
			<Diagram picture={picture} label="A picture" locale={locale} />,
			container,
		),
	);
	try {
		const svg = container.querySelector("svg");
		if (svg === null) {
			return undefined;
		}
		return {
			root: attributesOf(svg),
			shapes: [...svg.querySelectorAll("*")].map((element) => ({
				tag: element.tagName.toLowerCase(),
				attributes: attributesOf(element),
				text: element.textContent ?? "",
			})),
		};
	} finally {
		act(() => render(null, container));
	}
}

/** drawn is how the card draws a picture it can draw, in a language. */
export function drawn(picture: Picture, locale = "en"): Drawing {
	const drawing = drawingOf(picture, locale);
	if (drawing === undefined) {
		throw new Error(`the picture was not drawn: ${JSON.stringify(picture)}`);
	}
	return drawing;
}

/** shapesOf are the shapes of a drawing of a tag, and of a class where one is named. */
export function shapesOf(
	drawing: Drawing,
	tag: string,
	className?: string,
): Shape[] {
	return drawing.shapes.filter(
		(shape) =>
			shape.tag === tag &&
			(className === undefined || shape.attributes.class === className),
	);
}

/** numberOf is an attribute of a shape read as a number, NaN where it has none. */
export function numberOf(shape: Shape, name: string): number {
	const value = shape.attributes[name];
	return value === undefined ? Number.NaN : Number(value);
}

/** textOf is the first text a drawing writes these words in. */
export function textOf(drawing: Drawing, words: string): Shape {
	const found = textsOf(drawing).find((text) => text.text === words);
	if (found === undefined) {
		throw new Error(`the drawing does not write ${words}`);
	}
	return found;
}

/** textsOf are the texts a drawing writes. */
export function textsOf(drawing: Drawing): Shape[] {
	return drawing.shapes.filter((shape) => shape.tag === "text");
}

/**
 * textBox is the room a text takes, as the drawing lays it out: as wide as
 * the estimate of its width at its size, and the height of its capitals and
 * digits, about its middle.
 */
export function textBox(shape: Shape): Box {
	const size = numberOf(shape, "font-size");
	const x = numberOf(shape, "x");
	const y = numberOf(shape, "y");
	const width = widthOf(shape.text, size);
	const anchor = shape.attributes["text-anchor"];
	const left =
		anchor === "start" ? x : anchor === "end" ? x - width : x - width / 2;
	return { left, top: y - size / 2, right: left + width, bottom: y + size / 2 };
}

/** overlap says whether two boxes share more than a sliver of room. */
export function overlap(one: Box, other: Box, sliver = 0.5): boolean {
	return (
		Math.min(one.right, other.right) - Math.max(one.left, other.left) >
			sliver &&
		Math.min(one.bottom, other.bottom) - Math.max(one.top, other.top) > sliver
	);
}

/**
 * pointsOfPath are the points a path's data names — the ends of its lines and
 * the points of its curves — written in the absolute commands a picture draws
 * with: M, L, H, V, Q and Z. A command of another kind is an error.
 */
export function pointsOfPath(data: string): Point[] {
	const tokens = data.match(/[A-Za-z]|-?\d*\.?\d+(?:e[-+]?\d+)?/g) ?? [];
	const points: Point[] = [];
	let at: Point = { x: 0, y: 0 };
	let command = "";
	for (let index = 0; index < tokens.length; ) {
		const token = tokens[index] as string;
		if (/[A-Za-z]/.test(token)) {
			command = token;
			index++;
			if (command === "Z") {
				continue;
			}
			if (!"MLHVQ".includes(command)) {
				throw new Error(
					`a path is drawn with ${command}, which a test cannot read`,
				);
			}
			continue;
		}
		const next = (count: number) =>
			tokens.slice(index, index + count).map(Number);
		if (command === "M" || command === "L") {
			const [x = Number.NaN, y = Number.NaN] = next(2);
			at = { x, y };
			points.push(at);
			index += 2;
		} else if (command === "H") {
			at = { x: Number(token), y: at.y };
			points.push(at);
			index += 1;
		} else if (command === "V") {
			at = { x: at.x, y: Number(token) };
			points.push(at);
			index += 1;
		} else if (command === "Q") {
			const [cx = Number.NaN, cy = Number.NaN, x = Number.NaN, y = Number.NaN] =
				next(4);
			points.push({ x: cx, y: cy });
			at = { x, y };
			points.push(at);
			index += 4;
		} else {
			throw new Error(`a path names a number before any command: ${data}`);
		}
	}
	return points;
}

/**
 * boxOfShape is the room a shape's geometry takes, its strokes left out: a
 * path's points, a circle's square and a rectangle; a text's estimated box.
 */
export function boxOfShape(shape: Shape): Box | undefined {
	switch (shape.tag) {
		case "path": {
			const points = pointsOfPath(shape.attributes.d ?? "");
			const xs = points.map((point) => point.x);
			const ys = points.map((point) => point.y);
			return {
				left: Math.min(...xs),
				top: Math.min(...ys),
				right: Math.max(...xs),
				bottom: Math.max(...ys),
			};
		}
		case "circle": {
			const cx = numberOf(shape, "cx");
			const cy = numberOf(shape, "cy");
			const r = numberOf(shape, "r");
			return { left: cx - r, top: cy - r, right: cx + r, bottom: cy + r };
		}
		case "rect": {
			const x = numberOf(shape, "x");
			const y = numberOf(shape, "y");
			return {
				left: x,
				top: y,
				right: x + numberOf(shape, "width"),
				bottom: y + numberOf(shape, "height"),
			};
		}
		case "text":
			return textBox(shape);
		default:
			return undefined;
	}
}
