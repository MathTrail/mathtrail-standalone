// The icons a listing in the chats' directories shows: the site's logo,
// square, as a picture of a side the directories all take, with its corners
// left clear — once as it is, for a light page, and once with a rim around its
// tile, for a dark one. Each is drawn by Chromium from the logo the site and
// the card draw, so that a listing shows the very same mark.
//
//	node scripts/icons.ts
//
// It runs where the browsers are, inside the image of the pinned Playwright
// release, and writes over the icons in plugin/assets/, to be looked at
// before they are kept.

import { readFile } from "node:fs/promises";
import { join } from "node:path";
import { chromium } from "playwright-core";

const repository = join(import.meta.dirname, "..", "..");

/**
 * iconSide is the side of every icon, in pixels: within what both directories
 * take, a square of 512 to 2048 pixels for one and of 48 to 4096 for the
 * other.
 */
export const iconSide = 1024;

/** tile is the logo's dark ground, as the site's mark draws it. */
export const tile = '<rect width="48" height="48" rx="6" fill="#0c1a33"/>';

/**
 * rimColour is the line drawn around the tile for a dark page. Against
 * ChatGPT's dark page, #212121, the navy tile stands at about 1.08 to 1 and
 * its edge is lost; the rim stands at 3 to 1 or more both against that page
 * and against the tile, the contrast a graphic needs to be told apart.
 */
export const rimColour = "#5774a6";

/**
 * rim is drawn inside the square, as the light card's hairline is: its
 * stroke runs from the square's edge inwards, and its corners follow the
 * tile's.
 */
export const rim = `<rect x="0.75" y="0.75" width="46.5" height="46.5" rx="5.25" fill="none" stroke="${rimColour}" stroke-width="1.5"/>`;

/**
 * onDarkGround is the mark for a dark page: the same drawing with the rim
 * laid over its tile. It refuses a mark whose tile it does not find exactly
 * once, which a redrawn logo would be, rather than draw an icon without one.
 */
export function onDarkGround(mark: string): string {
	const found = mark.split(tile).length - 1;
	if (found !== 1) {
		throw new Error(`icons: the mark holds its tile ${found} times, not once`);
	}
	return mark.replace(tile, `${tile}\n  ${rim}`);
}

/**
 * icons are the pictures a listing takes, by the name each is kept under,
 * with how each draws the mark.
 */
export const icons: Readonly<Record<string, (mark: string) => string>> = {
	"logo.png": (mark) => mark,
	"logo-dark.png": onDarkGround,
};

/** keptIconOf is where an icon is kept, in the ChatGPT package's assets. */
export function keptIconOf(name: string): string {
	return join(repository, "plugin", "assets", name);
}

/**
 * pageOf is a page that shows the mark alone, filling a square of the icon's
 * side, on no ground at all, so that the corners the tile rounds off stay
 * clear in the picture.
 */
export function pageOf(mark: string): string {
	return `<!doctype html><html><head><style>
html, body { margin: 0; background: transparent; }
svg { display: block; width: ${iconSide}px; height: ${iconSide}px; }
</style></head><body>${mark}</body></html>`;
}

async function main(): Promise<void> {
	const mark = await readFile(
		join(repository, "site", "assets", "favicon.svg"),
		"utf8",
	);
	const browser = await chromium.launch();
	try {
		const page = await browser.newPage({
			viewport: { width: iconSide, height: iconSide },
		});
		for (const [name, draw] of Object.entries(icons)) {
			await page.setContent(pageOf(draw(mark)));
			await page.screenshot({ path: keptIconOf(name), omitBackground: true });
			console.log(`icons: plugin/assets/${name}`);
		}
	} finally {
		await browser.close();
	}
}

if (import.meta.main) {
	await main();
}
