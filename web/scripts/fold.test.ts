// @vitest-environment node
import { describe, expect, test } from "vitest";
import { narrowestFit } from "./fold.ts";

// tried are the widths a row is tried at.
const tried = { widest: 1400, narrowest: 720 };

// rowUnder is a row that runs out of room under width.
const rowUnder = (width: number) => async (at: number) => at < width;

describe("the narrowest width a row stays whole at", () => {
	test("is the last width tried before the row runs out of room", async () => {
		expect(await narrowestFit(rowUnder(1039), tried)).toBe(1039);
	});

	test("is found from the widest width down, whatever the row does under the first width it runs out of room at", async () => {
		const row = async (at: number) => at === 1000 || at < 800;

		expect(await narrowestFit(row, tried)).toBe(1001);
	});

	test("is refused when the row is out of room at the widest width tried", async () => {
		await expect(narrowestFit(rowUnder(1401), tried)).rejects.toThrow(
			"the row is out of room at 1400 px, the widest width tried",
		);
	});

	test("is refused when the row is still whole at the narrowest width tried", async () => {
		await expect(narrowestFit(rowUnder(720), tried)).rejects.toThrow(
			"the row stays whole at 720 px, the narrowest width tried",
		);
	});
});
