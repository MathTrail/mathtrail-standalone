// @vitest-environment node
import { readFile } from "node:fs/promises";
import { join } from "node:path";
import { describe, expect, test } from "vitest";
import fixture from "../../site/research/testdata/research.json";
import { outputPath } from "../src/site/addresses.ts";
import { sharingPicturePath, sharingPictureSize } from "../src/site/brand.ts";
import { siteData } from "../src/site/data.ts";
import { renderSite } from "../src/site/render.tsx";
import { siteDictionaries } from "../src/site/words.ts";
import { languagesAsked, pictureStyle } from "./og.ts";
import { homeOf, keptPictureOf, readSources } from "./prerender-site.ts";

const repository = join(import.meta.dirname, "..", "..");

// languages are the site's languages, those of its dictionaries.
const languages = [...siteDictionaries.keys()];

describe("the sharing pictures", () => {
	test("are kept in the site's assets, each under the name its language's pages serve it by", () => {
		for (const locale of languages) {
			expect(keptPictureOf(locale)).toBe(
				join(repository, "site", sharingPicturePath(locale)),
			);
		}
	});

	// A page declares the size of its picture, and a host that trusts the
	// declaration crops or scales a picture of another size.
	test("are PNG pictures of the size every page declares", async () => {
		for (const locale of languages) {
			const picture = await readFile(keptPictureOf(locale));

			expect(picture.subarray(1, 4).toString("latin1"), locale).toBe("PNG");
			expect(
				{ width: picture.readUInt32BE(16), height: picture.readUInt32BE(20) },
				locale,
			).toEqual(sharingPictureSize);
		}
	});

	test.each([
		[[], ["en", "ru", "zh-Hans"]],
		[["--language", "ru"], ["ru"]],
		[
			["--language", "zh-Hans", "--language", "en"],
			["en", "zh-Hans"],
		],
	])("taken with the arguments %j are of %j", (args, want) => {
		expect(languagesAsked(args, ["en", "ru", "zh-Hans"])).toEqual(want);
	});

	test.each([
		[["--language", "xx"], "the site has no language xx"],
		[["--language", "zh-hans"], "the site has no language zh-hans"],
		[["ru"], "Unexpected argument"],
		[["--lang", "ru"], "Unknown option"],
	])("are refused with the arguments %j", (args, said) => {
		expect(() => languagesAsked(args, ["en", "ru", "zh-Hans"])).toThrow(said);
	});

	// A part of the page the picture's layout names and the page no longer
	// draws would be laid out by no rule at all: a renamed button would come
	// back into the picture, and nothing would say so.
	test("lay out only parts the home page draws, in every language", async () => {
		const sources = await readSources(join(repository, "site", "content"));
		const files = renderSite({
			base: "https://mathtrail.app",
			sources,
			data: siteData(fixture),
		});
		const named = [
			...new Set(
				[...pictureStyle.matchAll(/\.(s-[a-z-]+)/g)].map(([, name]) => name),
			),
		];
		expect(named).not.toEqual([]);
		for (const locale of sources.keys()) {
			const home =
				files.find(({ path }) => path === outputPath(homeOf(locale)))?.data ??
				"";
			const classes = new Set(
				[...home.matchAll(/class="([^"]*)"/g)].flatMap(([, value]) =>
					(value ?? "").split(" "),
				),
			);

			expect(
				named.filter((name) => !classes.has(name ?? "")),
				locale,
			).toEqual([]);
		}
	}, 60_000);
});
