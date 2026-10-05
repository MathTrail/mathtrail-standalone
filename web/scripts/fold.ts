// The width the site's menu folds at. On a wide screen the menu shows its
// entries in a row beside the mark and the tools, and below a width its
// stylesheet sets it folds behind a button: the width at which the entries of
// the language with the longest words stop fitting in the row. This measures
// that width for every language of the site, in Chromium with a classic scroll
// bar, which takes room of its own, and in WebKit, over the site built into
// site/dist/.
//
//	node scripts/fold.ts
//
// It runs where the browsers are, inside the image of the pinned Playwright
// release, and changes nothing: the fold in the stylesheet is set by hand,
// from the widths it prints.

import { join } from "node:path";
import { type Browser, chromium, webkit } from "playwright-core";
import { preview } from "vite";
import { readSources } from "./prerender-site.ts";

const web = join(import.meta.dirname, "..");
const repository = join(web, "..");
const built = join(repository, "site", "dist");

/**
 * rowStyle shows the menu's entries in a row at every width, folded behind no
 * button, so that the width at which the row stops fitting can be found.
 */
export const rowStyle =
	".s-navlinks { display: flex !important; } .s-menu { display: none !important; }";

// widest and narrowest are the widths the row is tried at, and height the
// window's, under a page taller than it, so that a scroll bar is drawn.
const widest = 1400;
const narrowest = 720;
const height = 800;

/**
 * narrowestFit is the narrowest width, in whole pixels, at which a row stays
 * whole: tried from widest down a pixel at a time, the last width before
 * overflows says the row runs out of room. Tried so, it finds the widest width
 * at which the row runs out of room even where the stylesheet lays the header
 * out otherwise at some narrower width. A row out of room at widest, or still
 * whole at narrowest, folds at a width outside those tried, and is refused.
 */
export async function narrowestFit(
	overflows: (width: number) => Promise<boolean>,
	tried: { readonly widest: number; readonly narrowest: number },
): Promise<number> {
	for (let width = tried.widest; width >= tried.narrowest; width -= 1) {
		if (await overflows(width)) {
			if (width === tried.widest) {
				throw new Error(
					`the row is out of room at ${tried.widest} px, the widest width tried`,
				);
			}
			return width + 1;
		}
	}
	throw new Error(
		`the row stays whole at ${tried.narrowest} px, the narrowest width tried`,
	);
}

/**
 * menuFit is the narrowest width at which the menu of the home page of locale
 * stays in a row in browser.
 */
async function menuFit(
	browser: Browser,
	base: string,
	locale: string,
): Promise<number> {
	const page = await browser.newPage({ viewport: { width: widest, height } });
	try {
		// The menu is drawn by the page as it is built: no script of the site
		// is fetched, and the browser's own scripting stays on, since it is what
		// measures the row.
		await page.route("**/*.js", (route) => route.abort());
		const answer = await page.goto(new URL(`${locale}/`, base).href);
		if (!answer?.ok()) {
			throw new Error(
				`fold: ${built} has no home page in ${locale}: build the site first`,
			);
		}
		await page.addStyleTag({ content: rowStyle });
		await page.evaluate(() => document.fonts.ready);
		const overflows = async (width: number) => {
			await page.setViewportSize({ width, height });
			const overflow = await page.evaluate(() => {
				const row = document.querySelector(".s-nav-inner");
				return row === null ? null : row.scrollWidth > row.clientWidth;
			});
			if (overflow === null) {
				throw new Error("the page has no row of the menu to measure");
			}
			return overflow;
		};
		try {
			return await narrowestFit(overflows, { widest, narrowest });
		} catch (error) {
			throw new Error(
				`fold: ${browser.browserType().name()} ${locale}: ${error instanceof Error ? error.message : "the menu cannot be measured"}`,
				{ cause: error },
			);
		}
	} finally {
		await page.close();
	}
}

async function main(): Promise<void> {
	// The languages are the site's texts', as its build reads them.
	const locales = [
		...(await readSources(join(repository, "site", "content"))).keys(),
	].sort((a, b) => a.localeCompare(b));
	const server = await preview({
		configFile: join(web, "vite.config.site.ts"),
		logLevel: "warn",
		build: { outDir: built },
		preview: { host: "127.0.0.1", port: 4174, strictPort: false },
	});
	try {
		const base = server.resolvedUrls?.local[0];
		if (base === undefined) {
			throw new Error("fold: the site's preview names no address");
		}
		// Headless Chromium hides its scroll bars unless told otherwise, and a
		// reader's classic one takes room from the row.
		for (const [name, browserOf] of [
			[
				"chromium",
				() => chromium.launch({ ignoreDefaultArgs: ["--hide-scrollbars"] }),
			],
			["webkit", () => webkit.launch()],
		] as const) {
			const browser = await browserOf();
			try {
				for (const locale of locales) {
					const fits = await menuFit(browser, base, locale);
					console.log(
						`fold: ${name} ${locale}: the menu stays in a row down to ${fits} px`,
					);
				}
			} finally {
				await browser.close();
			}
		}
	} finally {
		await server.close();
	}
}

if (import.meta.main) {
	await main();
}
