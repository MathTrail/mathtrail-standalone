// The preview of the widget, driven in a real browser: served from the widget's
// sources as they are, its cards found frame by frame, and each waited for until
// it has become what its scene makes of it, on a clock that moves only when it
// is moved. What measures the cards and what photographs them share it.

import { basename, join } from "node:path";
import {
	type BrowserContext,
	chromium,
	type ElementHandle,
	type Frame,
	type Page,
} from "playwright-core";
import { createServer } from "vite";

const web = join(import.meta.dirname, "..");

/**
 * clockStarts is the moment the pages' clock is set to: any will do, and a
 * fixed one makes every run see the same dates.
 */
export const clockStarts = Date.UTC(2026, 0, 1);

/**
 * stillClock starts the clock of every page of context at clockStarts and has
 * it stand still from an hour after it: every card is drawn on it, moved on
 * only by what moves it. The hour is room to pause in, however slow the
 * machine; a pause at a moment already past is refused.
 */
export async function stillClock(context: BrowserContext): Promise<void> {
	await context.clock.install({ time: clockStarts });
	await context.clock.pauseAt(clockStarts + 3_600_000);
}

/**
 * served starts the preview, and says where it is. It watches no file: a
 * source changed while the cards are driven would reload them halfway. And
 * since nothing it serves changes while it runs, the browser may keep every
 * file it was sent, as Vite lets it keep the packages it bundles: every card
 * of a page loads the same hundred modules, which Chromium would otherwise ask
 * the server after again for each card, some seven thousand requests a page.
 */
export async function served(): Promise<{
	base: string;
	close: () => Promise<void>;
}> {
	const server = await createServer({
		configFile: join(web, "vite.config.preview.ts"),
		logLevel: "warn",
		server: {
			host: "127.0.0.1",
			port: 5173,
			strictPort: false,
			hmr: false,
			watch: null,
			headers: { "Cache-Control": "max-age=31536000, immutable" },
		},
	});
	await server.listen();
	const base = server.resolvedUrls?.local[0];
	if (base === undefined) {
		throw new Error("drive: the preview names no address");
	}
	return { base, close: () => server.close() };
}

/** Card is the frame of one scene in the preview's page. */
export type Card = { scene: string; element: ElementHandle; frame: Frame };

// cardsOf are the cards of the preview's page, in the order of its scenes. A
// frame the page takes down while it is read is not among them, and the page
// is read again until it holds still. Each card holds on to its frame's
// element until it is let go of.
async function cardsOf(page: Page): Promise<Card[]> {
	const cards: Card[] = [];
	for (const element of await page.locator("iframe").elementHandles()) {
		const frame = await element.contentFrame().catch(() => null);
		const scene = await element.getAttribute("title").catch(() => null);
		if (frame !== null && scene !== null) {
			cards.push({ scene, element, frame });
		}
	}
	return cards;
}

// markupOf is what a card's page holds, or nothing while it has no card: a
// card whose markup stays the same has done what its scene does to it.
async function markupOf({ frame }: Card): Promise<string> {
	try {
		return await frame.evaluate(async () => {
			await document.fonts.ready;
			return document.querySelector(".mt-widget") === null
				? ""
				: document.body.innerHTML;
		});
	} catch {
		return "";
	}
}

/** Settled is a card that has become what its scene makes of it. */
export type Settled = Card & { markup: string };

// A card is given at most this many ticks of the page's clock, a tenth of a
// second each, to settle: 28 seconds, short of the two minutes after which a
// waiting card gives up.
const mostTicks = 280;

/**
 * settled waits until every card of the page has been drawn and has become
 * what its scene makes of it, and says what the cards are. The page's clock
 * stands still but for what this moves it on by, a tick at a time, so that the
 * timers a card is drawn with fire, and a card is caught at the moment it is
 * meant to be, however long the machine takes to draw it.
 */
export async function settled(page: Page): Promise<Settled[]> {
	let before: string[] = [];
	let why = "";
	for (let tick = 0; tick < mostTicks; tick++) {
		await page.clock.runFor(100);
		const cards = await cardsOf(page);
		const markups = await Promise.all(cards.map(markupOf));
		why = unsettled(
			cards.map(({ scene }) => scene),
			markups,
			before,
		);
		if (why === "") {
			return cards.map((card, at) => ({ ...card, markup: markups[at] ?? "" }));
		}
		await letGo(cards);
		before = markups;
		await page.waitForTimeout(100);
	}
	throw new Error(`drive: the cards never settled: ${why}`);
}

/**
 * unsettled says what keeps the cards of a page from settling at a tick, given
 * their scenes, what each holds and what each held a tick before: no card at
 * all, cards that came or went in between, or by scene the cards with nothing
 * drawn and those still changing. It says nothing once every card is drawn and
 * holds what it held.
 */
export function unsettled(
	scenes: readonly string[],
	markups: readonly string[],
	before: readonly string[],
): string {
	if (scenes.length === 0) {
		return "no card on the page";
	}
	if (markups.length !== before.length) {
		return `cards came or went: ${before.length} a tick before, ${markups.length} now`;
	}
	const blank = scenes.filter((_, at) => markups[at] === "");
	const changing = scenes.filter(
		(_, at) => markups[at] !== "" && markups[at] !== before[at],
	);
	return [
		blank.length > 0 ? `nothing drawn in ${blank.join(", ")}` : "",
		changing.length > 0 ? `still changing: ${changing.join(", ")}` : "",
	]
		.filter((part) => part !== "")
		.join("; ");
}

/** letGo lets go of the elements the cards held on to. */
export async function letGo(cards: Card[]): Promise<void> {
	await Promise.all(cards.map(({ element }) => element.dispose()));
}

/** density is how many pixels of a picture stand for one of the page. */
export const density = 2;

/**
 * Picture is a card of the preview to photograph: its scene, in a theme and
 * at a width, and the file the picture is written to.
 */
export type Picture = {
	scene: string;
	theme: "light" | "dark";
	width: number;
	path: string;
};

/**
 * addressOf is the address of the preview served at base that shows the scene
 * of picture alone, in its theme, in English, at its width.
 */
export function addressOf(base: string, picture: Picture): string {
	const query = new URLSearchParams({
		scene: picture.scene,
		theme: picture.theme,
		lang: "en",
		widths: `${picture.width}`,
	});
	return `${base}preview.html?${query}`;
}

/**
 * photograph serves the preview and photographs, in Chromium and at density,
 * the card of each picture's scene: the card alone, its corners left clear,
 * with the preview's own bar and ground left out and whatever of the card
 * hidden names. Each picture is written to its file and logged under who.
 */
export async function photograph(
	who: string,
	pictures: readonly Picture[],
	hidden = "",
): Promise<void> {
	const preview = await served();
	const browser = await chromium.launch();
	try {
		// The spinner of a card that checks an answer turns; a card shown to a
		// reader who asks for no motion holds it still for the picture.
		const context = await browser.newContext({
			viewport: { width: 1280, height: 900 },
			deviceScaleFactor: density,
			reducedMotion: "reduce",
		});
		await stillClock(context);
		const page = await context.newPage();
		for (const picture of pictures) {
			await page.goto(addressOf(preview.base, picture));
			const cards = await settled(page);
			const card = cards.find((found) => found.scene === picture.scene);
			if (card === undefined) {
				throw new Error(`${who}: the preview shows no scene ${picture.scene}`);
			}
			await card.element.screenshot({
				path: picture.path,
				omitBackground: true,
				style: `.preview-bar { visibility: hidden; } .preview { background: transparent !important; } ${hidden}`,
			});
			console.log(`${who}: ${basename(picture.path)}`);
			await letGo(cards);
		}
	} finally {
		await browser.close();
		await preview.close();
	}
}
