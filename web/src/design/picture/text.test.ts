import { describe, expect, test } from "vitest";
import { fitted, smallest } from "./text";

describe("the size a picture writes a text at", () => {
	test("is the largest of the sizes at which it fits", () => {
		expect(fitted([16, 14, 12], (size) => size <= 14)).toBe(14);
	});

	test("is the last of them where it fits at none", () => {
		expect(fitted([16, 14, 12], () => false)).toBe(12);
	});

	test("is the smallest a picture writes at where there are no sizes to try", () => {
		expect(fitted([], () => true)).toBe(smallest);
	});
});
