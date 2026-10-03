import { afterEach, describe, expect, test, vi } from "vitest";
import { buildVersion } from "./version";

afterEach(() => {
	vi.unstubAllEnvs();
});

describe("the build's version", () => {
	test("is the one the build was given", () => {
		vi.stubEnv("VITE_VERSION", "v0.2.1-3-g38087a4");

		expect(buildVersion()).toBe("v0.2.1-3-g38087a4");
	});

	test.each([
		["was given none", undefined],
		["was given an empty one", ""],
	])("is dev when the build %s", (_, given) => {
		vi.stubEnv("VITE_VERSION", given);

		expect(buildVersion()).toBe("dev");
	});
});
