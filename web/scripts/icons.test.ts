// @vitest-environment node
import { readFile } from "node:fs/promises";
import { join } from "node:path";
import { describe, expect, test } from "vitest";
import { markPictureSide } from "../src/site/brand.ts";
import {
	iconSide,
	icons,
	keptIconOf,
	keptSiteIcon,
	onDarkGround,
	pageOf,
	rim,
	rimColour,
	tile,
} from "./icons.ts";

const repository = join(import.meta.dirname, "..", "..");

/** chatgptDark is the ground ChatGPT's dark theme lays a listing's icon on. */
const chatgptDark = "#212121";

/**
 * luminance is a colour's relative luminance, as WCAG defines it, with the
 * sRGB threshold the design's own test of the mark's inks uses.
 */
function luminance(hex: string): number {
	const channels = [1, 3, 5].map((at) => {
		const value = Number.parseInt(hex.slice(at, at + 2), 16) / 255;
		return value <= 0.04045 ? value / 12.92 : ((value + 0.055) / 1.055) ** 2.4;
	});
	const [red = 0, green = 0, blue = 0] = channels;
	return 0.2126 * red + 0.7152 * green + 0.0722 * blue;
}

/** contrast is the ratio WCAG gives two colours, the lighter one first. */
function contrast(one: string, other: string): number {
	const [lighter, darker] = [luminance(one), luminance(other)].sort(
		(a, b) => b - a,
	);
	return ((lighter ?? 0) + 0.05) / ((darker ?? 0) + 0.05);
}

/** attribute is a number the rim's drawing gives an attribute. */
function attribute(name: string): number {
	const found = rim.match(new RegExp(` ${name}="([^"]+)"`));
	if (found === null) {
		throw new Error(`the rim gives no ${name}`);
	}
	return Number(found[1]);
}

async function siteMark(): Promise<string> {
	return readFile(join(repository, "site", "assets", "favicon.svg"), "utf8");
}

describe("the directories' icons", () => {
	test("draw the dark one as the site's mark with a rim over its tile, and nothing else", async () => {
		const mark = await siteMark();
		const dark = onDarkGround(mark);

		expect(dark.split(rim).length - 1).toBe(1);
		expect(dark.replace(`\n  ${rim}`, "")).toBe(mark);
		expect(dark.indexOf(rim)).toBeGreaterThan(dark.indexOf(tile));
	});

	test("keep the rim inside the square, its corners following the tile's", () => {
		const stroke = attribute("stroke-width");

		expect(attribute("x") - stroke / 2).toBe(0);
		expect(attribute("y") - stroke / 2).toBe(0);
		expect(attribute("x") + attribute("width") + stroke / 2).toBe(48);
		expect(attribute("y") + attribute("height") + stroke / 2).toBe(48);
		expect(attribute("rx") + stroke / 2).toBe(6);
	});

	// The navy tile alone sinks into ChatGPT's dark page; a graphic is told
	// apart from what surrounds it at 3 to 1.
	test("set the rim apart from ChatGPT's dark page and from the tile", () => {
		const tileColour = tile.match(/fill="(#[0-9a-f]{6})"/)?.[1] ?? "";

		expect(contrast(tileColour, chatgptDark)).toBeLessThan(1.5);
		expect(contrast(rimColour, chatgptDark)).toBeGreaterThanOrEqual(3);
		expect(contrast(rimColour, tileColour)).toBeGreaterThanOrEqual(3);
	});

	test("refuse a mark whose tile is not there exactly once", async () => {
		const mark = await siteMark();

		expect(() => onDarkGround(mark.replace(tile, ""))).toThrow(/0 times/);
		expect(() => onDarkGround(mark + tile)).toThrow(/2 times/);
	});

	// The package's own test holds the kept pictures to what the directories
	// take; this one holds the pictures drawn to the ones the manifest names.
	test("are drawn under the very names the package's manifest gives its icons", async () => {
		const manifest = JSON.parse(
			await readFile(join(repository, "plugin", "plugin.json"), "utf8"),
		);
		const face = manifest.extensions["com.openai"].interface;
		const named = new Set(
			[face.logo, face.logoDark, face.composerIcon, face.composerIconDark].map(
				(path: string) => join(repository, "plugin", path),
			),
		);

		expect(new Set(Object.keys(icons).map(keptIconOf))).toEqual(named);
	});
});

describe("the site's icon as a picture", () => {
	test("is kept where the site serves it from", () => {
		expect(keptSiteIcon).toBe(
			join(repository, "site", "assets", "favicon.png"),
		);
	});

	// A search engine shows a square picture larger than 48 pixels beside a
	// result; every page states the side the picture has.
	test("is a square PNG larger than 48 pixels, of the side every page states", async () => {
		const picture = await readFile(keptSiteIcon);

		expect(picture.subarray(1, 4).toString("ascii")).toBe("PNG");
		expect([picture.readUInt32BE(16), picture.readUInt32BE(20)]).toEqual([
			markPictureSide,
			markPictureSide,
		]);
		expect(markPictureSide).toBeGreaterThan(48);
	});

	test("is drawn from a page that lays the mark out at its side, as the directories' icons are at theirs", () => {
		expect(pageOf("<svg></svg>", markPictureSide)).toContain(
			`width: ${markPictureSide}px; height: ${markPictureSide}px;`,
		);
		expect(pageOf("<svg></svg>")).toContain(
			`width: ${iconSide}px; height: ${iconSide}px;`,
		);
	});
});
