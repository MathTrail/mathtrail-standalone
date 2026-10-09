import { describe, expect, test } from "vitest";
import { hands } from "./clock";
import {
	drawn as clock,
	numberOf,
	overlap,
	pointsOfPath,
	textBox,
	textsOf,
} from "./testing/svg";

// face is the middle of a clock face, in the picture's units.
const face = { x: 100, y: 100 };

// directionOf is the direction of a point from the middle of the face, in
// degrees clockwise from twelve.
function directionOf(point: { x: number; y: number }): number {
	const degrees =
		(Math.atan2(point.x - face.x, face.y - point.y) * 180) / Math.PI;
	return (degrees + 360) % 360;
}

describe("a clock's hands", () => {
	test.each([
		["4:30", 135, 180],
		["16:30", 135, 180],
		["0:05", 2.5, 30],
		["12:00", 0, 0],
		["23:59", 359.5, 354],
		["9:45", 292.5, 270],
	])("at %s point %d° and %d° round from twelve", (time, hour, minute) => {
		expect(hands(time)).toEqual({ hour, minute });
	});

	test("are drawn where the time puts them, the hour hand the shorter and wider", () => {
		const drawn = clock({ kind: "clock", time: "9:45" }).shapes.filter(
			(shape) =>
				shape.tag === "path" && shape.attributes["stroke-linecap"] === "round",
		);
		const [hour, minute] = drawn.map(
			(shape) => pointsOfPath(shape.attributes.d ?? "")[1] ?? face,
		);

		expect(drawn.map((shape) => numberOf(shape, "stroke-width"))).toEqual([
			6, 3.5,
		]);
		expect(directionOf(hour ?? face)).toBeCloseTo(292.5, 0);
		expect(directionOf(minute ?? face)).toBeCloseTo(270, 0);
		expect(
			Math.hypot((hour?.x ?? 0) - face.x, (hour?.y ?? 0) - face.y),
		).toBeLessThan(
			Math.hypot((minute?.x ?? 0) - face.x, (minute?.y ?? 0) - face.y),
		);
	});
});

describe("a clock face", () => {
	test("is numbered from 1 to 12, each number thirty degrees round from the last", () => {
		const numerals = textsOf(clock({ kind: "clock", time: "4:30" }));

		expect(numerals.map((text) => text.text)).toEqual(
			Array.from({ length: 12 }, (_, before) => String(before + 1)),
		);
		for (const numeral of numerals) {
			const direction = directionOf({
				x: numberOf(numeral, "x"),
				y: numberOf(numeral, "y"),
			});
			expect(direction).toBeCloseTo((Number(numeral.text) * 30) % 360, 0);
		}
	});

	test("marks its sixty minutes, the hours longer", () => {
		const marks = clock({ kind: "clock", time: "4:30" }).shapes.filter(
			(shape) =>
				shape.tag === "path" &&
				(shape.attributes.d?.match(/M/g)?.length ?? 0) > 1,
		);

		expect(
			marks.map((shape) => shape.attributes.d?.match(/M/g)?.length),
		).toEqual([48, 12]);
	});

	test("labels its hands only where the description does, beside them", () => {
		const bare = textsOf(clock({ kind: "clock", time: "4:30" }));
		const labelled = textsOf(
			clock({
				kind: "clock",
				time: "4:30",
				hour_label: "H",
				minute_label: "M",
			}),
		);

		expect(bare).toHaveLength(12);
		expect(labelled.map((text) => text.text).slice(12)).toEqual(["M", "H"]);
	});

	test("labels its hands clear of its numbers and of each other at every minute of the day", () => {
		for (let minutes = 0; minutes < 24 * 60; minutes++) {
			const time = `${Math.floor(minutes / 60)}:${String(minutes % 60).padStart(2, "0")}`;
			const texts = textsOf(
				clock({ kind: "clock", time, hour_label: "H", minute_label: "M" }),
			);
			for (const [at, one] of texts.entries()) {
				for (const other of texts.slice(at + 1)) {
					expect(
						overlap(textBox(one), textBox(other)),
						`${time}: ${one.text} and ${other.text}`,
					).toBe(false);
				}
			}
		}
	});

	test("writes a hand's label below zero with the sign of mathematics, as every label is", () => {
		const labels = textsOf(
			clock({
				kind: "clock",
				time: "4:30",
				hour_label: "-3",
				minute_label: "-12",
			}),
		)
			.slice(12)
			.map((text) => text.text);

		expect(labels).toEqual(["−12", "−3"]);
	});
});
