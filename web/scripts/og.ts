// The pictures a link to the site shows where it is shared: one for each
// language of the site, at the size a sharing preview is drawn at. Each is
// the home page of its language as the site is built — the mark, the heading,
// its chips and its promise, and the card of a task beside them — laid out as
// a picture and photographed in Chromium, so that a shared link shows the page
// it leads to rather than a picture drawn apart from it.
//
//	node scripts/og.ts
//
// It runs where the browsers are, inside the image of the pinned Playwright
// release, over the site built into site/dist/, and writes over the pictures
// in site/assets/, to be looked at before they are kept.

import { join } from "node:path";
import { chromium } from "playwright-core";
import { preview } from "vite";
import { sharingPicturePath, sharingPictureSize } from "../src/site/brand.ts";
import { homeOf, keptPictureOf, readSources } from "./prerender-site.ts";

const web = join(import.meta.dirname, "..");
const repository = join(web, "..");
const built = join(repository, "site", "dist");

/**
 * pictureStyle lays the first screen of the home page out as a picture: the
 * mark and the name on top, with no menu; the heading, its chips and its
 * promise on the left, and no buttons, which nobody can press in a picture;
 * the card beside them in its phone, running off the picture's lower edge as
 * the page runs on; and nothing under the first screen.
 */
export const pictureStyle = `
	.s-navlinks, .s-nav-tools, .s-footer, .s-hero-actions, .s-hero-note,
	.s-main > :not(:first-child) {
		display: none !important;
	}
	html, body { overflow: hidden; }
	.s-nav { position: static; box-shadow: none; background: none; }
	.s-nav-inner, .s-hero { max-inline-size: none; padding-inline: 56px; }
	.s-nav-inner { min-block-size: 96px; }
	.s-hero {
		gap: 56px;
		padding-block-start: 8px;
	}
	.s-hero-copy { padding-block-start: 8px; }
	.s-hero h1 { font-size: 60px; }
	.s-hero .s-lead { font-size: 22px; }
`;

async function main(): Promise<void> {
	// The languages are the site's texts', as its build reads them.
	const locales = [
		...(await readSources(join(repository, "site", "content"))).keys(),
	].sort((a, b) => a.localeCompare(b, "en"));
	const server = await preview({
		configFile: join(web, "vite.config.site.ts"),
		logLevel: "warn",
		build: { outDir: built },
		preview: { host: "127.0.0.1", port: 4174, strictPort: false },
	});
	try {
		const base = server.resolvedUrls?.local[0];
		if (base === undefined) {
			throw new Error("og: the site's preview names no address");
		}
		const browser = await chromium.launch();
		try {
			// The picture is of the page as it is built: no script of the site is
			// fetched, since one that ran first would decide what the picture shows
			// by how fast it ran. The browser's own scripting stays on, since it is
			// what lays the picture out.
			const page = await browser.newPage({ viewport: sharingPictureSize });
			await page.route("**/*.js", (route) => route.abort());
			for (const locale of locales) {
				const answer = await page.goto(
					new URL(homeOf(locale).slice(1), base).href,
				);
				if (!answer?.ok()) {
					throw new Error(
						`og: ${built} has no home page in ${locale}: build the site first`,
					);
				}
				await page.addStyleTag({ content: pictureStyle });
				await page.evaluate(() => document.fonts.ready);
				await page.screenshot({ path: keptPictureOf(locale) });
				console.log(`og: ${sharingPicturePath(locale)}`);
			}
		} finally {
			await browser.close();
		}
	} finally {
		await server.close();
	}
}

if (import.meta.main) {
	await main();
}
