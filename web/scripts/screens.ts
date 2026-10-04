// The pictures of the widget the README shows: a task, the result of a wrong
// answer and the progress, each in the light and the dark theme, as a phone 428
// px wide shows them at twice its density. They are the preview's own scenes —
// the widget's page, driven as a chat host drives it — photographed in
// Chromium, so that the README shows the cards as they are rather than as they
// were designed. The corners round the card are left clear, for a page of
// either theme to show through.
//
//	node scripts/screens.ts
//
// It runs where the browsers are, inside the image of the pinned Playwright
// release, and writes over the pictures in docs/screens/.

import { join } from "node:path";
import { chromium } from "playwright-core";
import { letGo, served, settled, stillClock } from "./drive.ts";

/** Shot is a scene of the preview, and the file its pictures are named by. */
export type Shot = { scene: string; file: string };

/**
 * shots are the scenes the README shows, in its order: the cards of a child
 * past the trial series, which offer the choice of the topic.
 */
export const shots: readonly Shot[] = [
	{ scene: "task after the trial series", file: "task" },
	{ scene: "wrong after the trial series", file: "wrong" },
	{ scene: "progress, the model's card", file: "progress" },
];

/** themes are the themes each scene is photographed in. */
export const themes = ["light", "dark"] as const;

/** width is how wide a card is photographed, in the pixels of a page. */
export const width = 428;

const screens = join(import.meta.dirname, "..", "..", "docs", "screens");

/**
 * addressOf is the preview's address that shows one scene alone, in a theme,
 * in English, at the width photographed.
 */
export function addressOf(
	base: string,
	shot: Shot,
	theme: (typeof themes)[number],
): string {
	const query = new URLSearchParams({
		scene: shot.scene,
		theme,
		lang: "en",
		widths: `${width}`,
	});
	return `${base}preview.html?${query}`;
}

/** pictureOf is the file a scene's picture in a theme is written to. */
export function pictureOf(shot: Shot, theme: (typeof themes)[number]): string {
	return join(screens, `${shot.file}-${theme}.png`);
}

async function main(): Promise<void> {
	const preview = await served();
	const browser = await chromium.launch();
	try {
		// The spinner of a card that checks an answer turns; a card shown to a
		// reader who asks for no motion holds it still for the picture.
		const context = await browser.newContext({
			viewport: { width: 1280, height: 900 },
			deviceScaleFactor: 2,
			reducedMotion: "reduce",
		});
		await stillClock(context);
		const page = await context.newPage();
		for (const theme of themes) {
			for (const shot of shots) {
				await page.goto(addressOf(preview.base, shot, theme));
				const cards = await settled(page);
				const card = cards.find((found) => found.scene === shot.scene);
				if (card === undefined) {
					throw new Error(`screens: the preview shows no scene ${shot.scene}`);
				}
				await card.element.screenshot({
					path: pictureOf(shot, theme),
					omitBackground: true,
					style:
						".preview-bar { visibility: hidden; } .preview { background: transparent !important; }",
				});
				console.log(`screens: ${shot.file}-${theme}.png`);
				await letGo(cards);
			}
		}
	} finally {
		await browser.close();
		await preview.close();
	}
}

if (import.meta.main) {
	await main();
}
