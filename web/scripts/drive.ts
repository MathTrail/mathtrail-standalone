// The preview of the widget, driven in a real browser: built from the widget's
// sources as they are and served, its cards found frame by frame, and each
// waited for until it has become what its scene makes of it, on a clock that
// moves only when it is moved. What measures the cards and what photographs
// them share it.

import { mkdtemp, rm } from "node:fs/promises";
import { tmpdir } from "node:os";
import { basename, join } from "node:path";
import {
	type BrowserContext,
	chromium,
	type ElementHandle,
	errors,
	type Frame,
	type Page,
} from "playwright-core";
import { build, type Plugin, preview } from "vite";

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

// previewConfig is the preview's own configuration, which builds it as it
// serves it.
const previewConfig = join(web, "vite.config.preview.ts");

/** previewFolder begins the name of the folder a built preview is served from. */
export const previewFolder = "mathtrail-preview-";

/**
 * builtPreview builds the preview and the widget's page it frames into
 * outDir, or writes nothing where outDir is not given, and lets plugins watch
 * the build. The build is one for development, as the preview's server is: it
 * speaks the pseudo-language, and a mistake in the words stops it.
 */
export async function builtPreview({
	outDir,
	plugins = [],
}: {
	outDir?: string;
	plugins?: Plugin[];
}): Promise<void> {
	// The build reads whether it is one for development from the environment,
	// as a build run from the command line would; the environment is given
	// back as it was once the build is over.
	const environment = process.env.NODE_ENV;
	process.env.NODE_ENV = "development";
	try {
		await build({
			configFile: previewConfig,
			mode: "development",
			logLevel: "warn",
			plugins,
			build: {
				outDir,
				write: outDir !== undefined,
				emptyOutDir: true,
				// The preview's build ships nowhere, and carries every dictionary.
				chunkSizeWarningLimit: 4096,
				rolldownOptions: {
					input: {
						preview: join(web, "preview.html"),
						widget: join(web, "widget.html"),
					},
				},
			},
		});
	} finally {
		if (environment === undefined) {
			delete process.env.NODE_ENV;
		} else {
			process.env.NODE_ENV = environment;
		}
	}
}

/**
 * served builds the preview, with plugins watching the build, and serves what
 * was built, and says where it is. A card then loads the widget as a few built
 * files rather than as the hundred modules of its sources, which takes a
 * browser a fraction of the time, every card of a page over, and lays it out
 * to the pixel as the preview's server does. Since nothing served changes
 * while it runs, the browser may keep every file it was sent. A build that
 * fails leaves no folder behind.
 */
export async function served({
	plugins = [],
}: {
	plugins?: Plugin[];
} = {}): Promise<{
	base: string;
	close: () => Promise<void>;
}> {
	const folder = await mkdtemp(join(tmpdir(), previewFolder));
	const removed = () => rm(folder, { recursive: true, force: true });
	try {
		await builtPreview({ outDir: folder, plugins });
		const server = await preview({
			configFile: previewConfig,
			mode: "development",
			logLevel: "warn",
			build: { outDir: folder },
			preview: {
				host: "127.0.0.1",
				port: 5173,
				strictPort: false,
				headers: { "Cache-Control": "max-age=31536000, immutable" },
			},
		});
		const base = server.resolvedUrls?.local[0];
		if (base === undefined) {
			await server.close();
			throw new Error("drive: the preview names no address");
		}
		return {
			base,
			close: async () => {
				await server.close();
				await removed();
			},
		};
	} catch (error) {
		await removed();
		throw error;
	}
}

/**
 * readWait is how long a page or a card is waited on for one answer, in
 * milliseconds. A page's first answers wait on its cards loading, which keeps
 * a browser busy for some ten seconds on four cores, and on a slower machine
 * for twice as long and more; a browser that takes longer than this to answer
 * has stopped answering.
 */
export const readWait = 60_000;

/** Late is the refusal of a wait on the browser that took longer than it was given. */
export class Late extends Error {}

/**
 * within is what promise comes to, unless it takes longer than ms: then it is
 * refused as Late, naming what was waited for. The browser's own waits on a
 * page or a frame have no end, so a browser that stops answering would hold a
 * run for as long as its machine is lent; every such wait is given one here.
 * What was waited for may go on in the browser: only the waiting for it ends.
 */
export async function within<T>(
	ms: number,
	what: string,
	promise: Promise<T>,
): Promise<T> {
	let timer: ReturnType<typeof setTimeout> | undefined;
	const late = new Promise<never>((_, refuse) => {
		timer = setTimeout(
			() => refuse(new Late(`drive: ${what} took longer than ${ms / 1000} s`)),
			ms,
		);
	});
	try {
		return await Promise.race([promise, late]);
	} finally {
		clearTimeout(timer);
	}
}

/**
 * lateIn is the refusal error is, or was caused by, that says a browser did
 * not answer in time: a Late, or Playwright's own refusal of a wait it gives a
 * time to, such as opening a page.
 */
export function lateIn(error: unknown): Error | undefined {
	for (let cause = error; cause instanceof Error; cause = cause.cause) {
		if (cause instanceof Late || cause instanceof errors.TimeoutError) {
			return cause;
		}
	}
	return undefined;
}

/**
 * againIfLate is what attempt comes to. An attempt refused because the
 * browser stopped answering is made once more, after renew has given it a
 * browser that answers; the second refusal stands. A browser under load may
 * stop answering of itself, and a second browser does what the first did not;
 * a card that stops every browser that draws it stops the second one too,
 * and is not hidden. Any other refusal stands at once.
 */
export async function againIfLate<T>(
	attempt: () => Promise<T>,
	renew: (late: Error) => Promise<void>,
): Promise<T> {
	try {
		return await attempt();
	} catch (error) {
		const late = lateIn(error);
		if (late === undefined) {
			throw error;
		}
		await renew(late);
		return attempt();
	}
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

/**
 * markupOf is what a card's page holds, or nothing while it has no card or
 * its frame is between pages: a card whose markup stays the same has done
 * what its scene does to it. A card that does not answer within readWait is
 * refused by its scene's name.
 */
export function markupOf({ frame, scene }: Pick<Card, "frame" | "scene">) {
	const reading = frame
		.evaluate(async () => {
			await document.fonts.ready;
			return document.querySelector(".mt-widget") === null
				? ""
				: document.body.innerHTML;
		})
		.catch(() => "");
	return within(readWait, `reading the card ${scene}`, reading);
}

/** Settled is a card that has become what its scene makes of it. */
export type Settled = Card & { markup: string };

// A card is given at most this many ticks of the page's clock, a tenth of a
// second each, to settle: 28 seconds, short of the two minutes after which a
// waiting card gives up.
const mostTicks = 280;

/**
 * settleWait is how long the cards of a page are waited on to settle, in
 * milliseconds, however few ticks that is: a page's cards settle in well under
 * a minute, and on a slow machine 280 ticks would take longer than a page has.
 */
export const settleWait = 4 * 60_000;

/**
 * settled waits until every card of the page has been drawn and has become
 * what its scene makes of it, and says what the cards are. The page's clock
 * stands still but for what this moves it on by, a tick at a time, so that the
 * timers a card is drawn with fire, and a card is caught at the moment it is
 * meant to be, however long the machine takes to draw it. Cards that have not
 * settled in their ticks, or in settleWait, are named in its refusal.
 */
export async function settled(page: Page): Promise<Settled[]> {
	const ends = Date.now() + settleWait;
	let before: string[] = [];
	let why = "";
	for (let tick = 0; tick < mostTicks && Date.now() < ends; tick++) {
		await within(readWait, "moving the clock on", page.clock.runFor(100));
		const cards = await within(readWait, "finding the cards", cardsOf(page));
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
	await within(
		readWait,
		"letting go of the cards",
		Promise.all(cards.map(({ element }) => element.dispose())),
	);
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
	try {
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
					throw new Error(
						`${who}: the preview shows no scene ${picture.scene}`,
					);
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
		}
	} finally {
		await preview.close();
	}
}
