import { describe, expect, test } from "vitest";
import { byCodeUnits } from "./order";

describe("the order of texts", () => {
	test("is by code units, wherever it is worked out", () => {
		expect(["ru", "en", "zh-Hans", "en-GB", "Z"].sort(byCodeUnits)).toEqual([
			"Z",
			"en",
			"en-GB",
			"ru",
			"zh-Hans",
		]);
		expect(byCodeUnits("en", "en")).toBe(0);
	});
});
