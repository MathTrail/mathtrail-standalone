import { afterEach, describe, expect, test, vi } from "vitest";
import { buildVersion, versionGiven, versionNumber } from "./version";

afterEach(() => {
	vi.unstubAllEnvs();
});

describe("the build's version", () => {
	test("is the one the build was given", () => {
		vi.stubEnv("VITE_VERSION", "v0.2.1-3-g38087a4");

		expect(buildVersion()).toBe("v0.2.1-3-g38087a4");
		expect(versionGiven()).toBe(true);
	});

	test.each([
		["was given none", undefined],
		["was given an empty one", ""],
		["was told it is dev", "dev"],
	])("is dev, and not given, when the build %s", (_, given) => {
		vi.stubEnv("VITE_VERSION", given);

		expect(buildVersion()).toBe("dev");
		expect(versionGiven()).toBe(false);
	});
});

describe("the number a card shows for a version", () => {
	test.each([
		["a release", "v0.2.5", "0.2.5"],
		[
			"a build past a release",
			"v0.2.5-4-gfcd5198-dirty",
			"0.2.5-4-gfcd5198-dirty",
		],
		["a release of two-digit parts", "v10.12.345", "10.12.345"],
		["a build that was told nothing", "dev", "dev"],
		["a bare commit", "8ff72f8", "8ff72f8"],
		["a word that only begins with v", "vnext", "vnext"],
		["a version with a v inside it", "0.2.5-v2", "0.2.5-v2"],
	])("drops only the v of a tag, for %s", (_, version, number) => {
		expect(versionNumber(version)).toBe(number);
	});
});
