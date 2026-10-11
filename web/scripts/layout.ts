// The widget's layout, measured in real browsers. Every scene of the preview —
// every state of a lesson on a card — is drawn in every language the widget
// speaks and in the pseudo-language, at the widths a card has to fit, in
// Chromium, the engine of Chrome and of Android's WebView, and in WebKit, the
// engine of Safari and of every browser on an iPhone. A card fails when its
// page scrolls sideways, when a part of it sticks out of the card, when text
// runs out of the box it is set in, or when text a card is written to show
// whole is cut short; and when a task's picture is wider than its room, or
// writes a word past its edges, smaller than it was written or than 11 px, or
// over another. The scenes are looked at again once a wait would have been
// given up on. Every card in a language written right to left is
// photographed, to be looked at.
//
//	node scripts/layout.ts [--engine chromium] [--language ar] [--width 320] [--shard 2/6] [--sample <commit>]
//
// A run can be shared out between machines that measure at once: --shard 2/6
// measures the second of six parts of it, and the six parts between them
// measure every card once. --sample <commit> measures three languages rather
// than every one: the pseudo-language, English and one the commit picks.
//
// It runs where the browsers are, inside the image of the pinned Playwright
// release, against the widget's sources as they are: the preview is served
// here, and serves the widget's own page to its frames.

import { mkdir, rm, writeFile } from "node:fs/promises";
import { join } from "node:path";
import { parseArgs } from "node:util";
import {
	type Browser,
	type BrowserType,
	chromium,
	type Page,
	webkit,
} from "playwright-core";
import {
	againIfLate,
	type Card,
	letGo,
	readWait,
	served,
	settled,
	stillClock,
	within,
} from "./drive.ts";

/** Finding is one way the card of one frame does not fit. */
export type Finding = {
	/** what went wrong, in words. */
	what: string;
	/** where names the element: its tag, its classes and its first words. */
	where: string;
	/** by is how many pixels too far. */
	by: number;
};

/** Box is the room something takes on the screen: its edges. */
export type Box = { left: number; top: number; right: number; bottom: number };

/**
 * PictureWord is one word of a task's picture as a browser draws it: what it
 * says, the room it takes, the size the picture writes it at, and the size it
 * comes out at on the screen.
 */
export type PictureWord = Box & { text: string; written: number; size: number };

/**
 * DrawnPicture is a task's picture as a browser draws it: what it is, the
 * room it takes, the width it is given, and its words.
 */
export type DrawnPicture = {
	where: string;
	box: Box;
	room: number;
	words: PictureWord[];
};

/** Measured is what the card of one scene showed at one moment. */
export type Measured = {
	engine: string;
	language: string;
	width: number;
	scene: string;
	moment: string;
	findings: Finding[];
};

// The widths a card has to fit: the narrowest, a phone less a host's margins,
// and the design's narrow and wide cards.
const widths = [320, 360, 640];

// The engines a card is measured in.
const engines: Readonly<Record<string, BrowserType>> = { chromium, webkit };

// The moments a card is looked at, each as far after the one before as the
// clock is moved on: a card waiting for a task being handed in gives the wait
// up after two minutes.
const moments: readonly { name: string; after: string }[] = [
	{ name: "at once", after: "" },
	{ name: "after 125 s", after: "02:05" },
];

// Where the cards written right to left are photographed: in Chromium, at a
// phone's width.
const photographed = { engine: "chromium", width: 360 };

// The longest version a build names itself by: one between releases, with
// changes of its own. A card's header shows it, and has to hold it at every
// width.
const longestVersion = "v0.12.345-678-g1a2b3c4d-dirty";

const web = join(import.meta.dirname, "..");
const out = join(web, "layout");

// The functions from here to findingsIn run inside the card's page: their text
// is handed to it as it is written, so they use nothing from around them.

// describe names an element as a finding does.
function describe(element: Element): string {
	const classes = [...element.classList].map((name) => `.${name}`).join("");
	const words = (element.textContent ?? "").trim().slice(0, 40);
	return `${element.tagName.toLowerCase()}${classes} "${words}"`;
}

// shownIn are the elements of the card a reader sees laid out: not what is
// hidden from sight but read out, nor the parts of a drawing that fills its
// box and is cut by it, whose box is all a reader sees of them and is measured
// itself.
function shownIn(card: Element): HTMLElement[] {
	return [...card.querySelectorAll<HTMLElement>("*")].filter((element) => {
		const box = element.getBoundingClientRect();
		const cut =
			element instanceof SVGElement &&
			element.ownerSVGElement
				?.getAttribute("preserveAspectRatio")
				?.endsWith("slice") === true;
		return (
			!cut &&
			element.closest(".mt-vh, [hidden]") === null &&
			getComputedStyle(element).display !== "none" &&
			(box.width > 0 || box.height > 0)
		);
	});
}

// outOfTheCard is how far element stands out of the card's edges, or nothing.
function outOfTheCard(element: HTMLElement, edges: DOMRect): Finding[] {
	const box = element.getBoundingClientRect();
	const beyond = Math.round(
		Math.max(edges.left - box.left, box.right - edges.right),
	);
	return beyond > 1
		? [{ what: "sticks out of the card", where: describe(element), by: beyond }]
		: [];
}

// outOfItsBox is how far what element holds runs out of its box, sideways and
// downwards, where the box does not scroll: cut short where the box ends it
// with an ellipsis, spilling over what is around it otherwise.
function outOfItsBox(element: HTMLElement): Finding[] {
	// A part of a picture has no box its words could run out of: a word's own
	// box is its font's, which a mark above a Thai letter runs past, and where
	// the words of a picture stand is what the picture's findings measure.
	if (element.clientWidth === 0 || element instanceof SVGElement) {
		return [];
	}
	const style = getComputedStyle(element);
	const scrolls = /auto|scroll/;
	const found: Finding[] = [];
	const wider = element.scrollWidth - element.clientWidth;
	if (wider > 1 && !scrolls.test(style.overflowX)) {
		const cut = style.textOverflow === "ellipsis";
		const what = cut ? "is cut short" : "runs out of its box sideways";
		found.push({ what, where: describe(element), by: wider });
	}
	const taller = element.scrollHeight - element.clientHeight;
	if (taller > 1 && !scrolls.test(style.overflowY)) {
		const what = "runs out of its box downwards";
		found.push({ what, where: describe(element), by: taller });
	}
	return found;
}

// findingsIn are the ways the card of the page does not fit.
function findingsIn(): Finding[] {
	const page = document.documentElement;
	const sideways = page.scrollWidth - page.clientWidth;
	const found: Finding[] =
		sideways > 0
			? [{ what: "the page scrolls sideways", where: "html", by: sideways }]
			: [];
	const card = document.querySelector(".mt-widget");
	if (card === null) {
		return found;
	}
	const edges = card.getBoundingClientRect();
	for (const element of shownIn(card)) {
		found.push(...outOfTheCard(element, edges), ...outOfItsBox(element));
	}
	return found;
}

// picturesIn are the pictures of the card of the page a reader sees, as they
// are drawn: the room each takes, the width of the box it stands in, and each
// of its words, with the size it is written at and the size the screen shows
// it at.
function picturesIn(): DrawnPicture[] {
	const edgesOf = (rect: DOMRect) => ({
		left: rect.left,
		top: rect.top,
		right: rect.right,
		bottom: rect.bottom,
	});
	return [...document.querySelectorAll<SVGSVGElement>("svg.mt-picture")]
		.filter(
			(svg) =>
				svg.closest(".mt-vh, [hidden]") === null &&
				svg.getBoundingClientRect().width > 0,
		)
		.map((svg) => {
			const scale = svg.getScreenCTM()?.a ?? 1;
			const holder = svg.parentElement;
			const style = holder === null ? undefined : getComputedStyle(holder);
			const room =
				holder === null || style === undefined
					? svg.getBoundingClientRect().width
					: holder.clientWidth -
						Number.parseFloat(style.paddingLeft) -
						Number.parseFloat(style.paddingRight);
			return {
				where: describe(svg),
				box: edgesOf(svg.getBoundingClientRect()),
				room,
				words: [...svg.querySelectorAll("text")].map((text) => {
					const written = Number(text.getAttribute("font-size"));
					return {
						text: text.textContent ?? "",
						...edgesOf(text.getBoundingClientRect()),
						written,
						size: written * scale,
					};
				}),
			};
		});
}

// measuring is the expression a card's page evaluates to its findings.
const measuring = `(() => {
${[describe, shownIn, outOfTheCard, outOfItsBox, findingsIn].join("\n")}
return findingsIn();
})()`;

// picturing is the expression a card's page evaluates to its pictures.
const picturing = `(() => {
${[describe, picturesIn].join("\n")}
return picturesIn();
})()`;

// smallest is the size below which a picture writes no word, but where its
// drawing means to: the widest labels in a table or a grid of many columns.
const smallest = 11;

// near is how far, in pixels, a measure may be off before it counts.
const near = 1;

// wordOf names a word of a picture as a finding does.
function wordOf(word: PictureWord): string {
	return `text "${word.text}"`;
}

// overlapping says whether two boxes share more than a sliver of room.
function overlapping(one: Box, other: Box): boolean {
	return (
		Math.min(one.right, other.right) - Math.max(one.left, other.left) > near &&
		Math.min(one.bottom, other.bottom) - Math.max(one.top, other.top) > near
	);
}

/**
 * pictureFindings are the ways the pictures of a card do not fit: a picture
 * wider than the box it stands in, a word that runs out of its picture, a
 * word smaller on the screen than 11 px, or than the picture writes it where
 * the picture means it smaller, and a word over another word.
 */
export function pictureFindings(pictures: readonly DrawnPicture[]): Finding[] {
	return pictures.flatMap((picture) => {
		const found: Finding[] = [];
		const wider = picture.box.right - picture.box.left - picture.room;
		if (wider > near) {
			found.push({
				what: "is wider than its box",
				where: picture.where,
				by: Math.round(wider),
			});
		}
		for (const [at, word] of picture.words.entries()) {
			const beyond = Math.max(
				picture.box.left - word.left,
				picture.box.top - word.top,
				word.right - picture.box.right,
				word.bottom - picture.box.bottom,
			);
			if (beyond > near) {
				found.push({
					what: "runs out of its picture",
					where: wordOf(word),
					by: Math.round(beyond),
				});
			}
			const least = Math.min(smallest, word.written);
			if (word.size < least - 0.05) {
				found.push({
					what: `is drawn smaller than ${least} px`,
					where: wordOf(word),
					by: Math.round((least - word.size) * 10) / 10,
				});
			}
			for (const other of picture.words.slice(at + 1)) {
				if (overlapping(word, other)) {
					found.push({
						what: `overlaps ${wordOf(other)}`,
						where: wordOf(word),
						by: near,
					});
				}
			}
		}
		return found;
	});
}

/**
 * allowed says whether a finding is how the card is meant to behave. A
 * pseudonym is the parent's own and may be as long as a profile allows, so the
 * line at the top that shows it gives way to what stands beside it and ends
 * in an ellipsis; and a box a person types into scrolls its own text, so a
 * pseudonym as long as that is wider than the box on a phone, which shows the
 * part being typed.
 */
export function allowed(finding: Finding): boolean {
	return (
		(finding.what === "is cut short" &&
			finding.where.startsWith("span.mt-bar-name ")) ||
		(finding.what === "runs out of its box sideways" &&
			finding.where.startsWith("input.mt-input "))
	);
}

// photograph keeps a picture of a card, named by its place among the scenes,
// its scene and the moment. The bar at the top of the preview stays where it
// is while the page scrolls, and is hidden from the picture it would cover.
async function photograph(
	card: Card,
	at: number,
	folder: string,
	moment: string,
): Promise<void> {
	const when = moment === moments[0]?.name ? "" : `, ${moment}`;
	const name = `${String(at + 1).padStart(2, "0")} ${card.scene}${when}`;
	await card.element.screenshot({
		path: join(folder, `${name.replace(/[^\p{L}\p{N} ,().-]/gu, "")}.png`),
		style: ".preview-bar { visibility: hidden; }",
	});
}

/** Shown is one page of the preview: its cards in a language, at a width, in an engine. */
export type Shown = { engine: string; language: string; width: number };

// measuredCard is what one card is found to have: the ways it and its pictures
// do not fit, but for those that are how the card is meant to behave. Each
// measure is waited on for a time, under where, which names the card.
async function measuredCard(card: Card, where: string): Promise<Finding[]> {
	const parts = await within(
		readWait,
		`measuring ${where}`,
		card.frame.evaluate<Finding[]>(measuring),
	);
	const pictures = await within(
		readWait,
		`measuring the pictures of ${where}`,
		card.frame.evaluate<DrawnPicture[]>(picturing),
	);
	return [...parts, ...pictureFindings(pictures)].filter(
		(finding) => !allowed(finding),
	);
}

// measuredOn measures the cards of one page of the preview at every moment,
// and photographs those written right to left where photographs are taken:
// each as it is first drawn, and again at a later moment only if it changed.
async function measuredOn(page: Page, base: string, shown: Shown) {
	const { engine, language, width } = shown;
	const query = new URLSearchParams({ lang: language, widths: `${width}` });
	await page.goto(`${base}preview.html?${query}`);
	const photographs =
		engine === photographed.engine && width === photographed.width;
	const pictured = new Map<string, string>();
	const measured: Measured[] = [];
	for (const moment of moments) {
		if (moment.after !== "") {
			await within(
				readWait,
				`moving the clock of ${pageOf(shown)} on`,
				page.clock.fastForward(moment.after),
			);
		}
		const cards = await settled(page).catch((error: unknown) => {
			const said =
				error instanceof Error ? error.message : "the cards never settled";
			throw new Error(`layout: ${pageOf(shown)}, ${moment.name}: ${said}`, {
				cause: error,
			});
		});
		for (const [at, card] of cards.entries()) {
			const scene = card.scene;
			const where = `${pageOf(shown)}, ${moment.name}, ${scene}`;
			const findings = await measuredCard(card, where);
			measured.push({ ...shown, scene, moment: moment.name, findings });
			// Taking a picture hides the caret of a field, and leaves an empty
			// style behind on it, which is no change of the card's own.
			const markup = card.markup.replaceAll(' style=""', "");
			if (
				photographs &&
				pictured.get(scene) !== markup &&
				(await within(
					readWait,
					`reading the direction of ${where}`,
					card.frame.evaluate(() => document.documentElement.dir),
				)) === "rtl"
			) {
				pictured.set(scene, markup);
				await photograph(card, at, join(out, engine, language), moment.name);
			}
		}
		await letGo(cards);
	}
	return measured;
}

// pseudo is the tag of the pseudo-language: English stretched as a longer
// language stretches it, longer than any language the widget speaks.
const pseudo = "en-XA";

/**
 * sampled are the languages a pull request is measured in: the
 * pseudo-language, which stands in for every longer language; English, the
 * language the widget's words are written in; and one other of those offered,
 * picked by the commit, so that every run of a commit measures the same one
 * and the next commit most likely another.
 */
export function sampled(offered: readonly string[], commit: string): string[] {
	commitOf(commit);
	const always = [pseudo, "en"];
	const missing = unoffered(always, offered);
	if (missing.length > 0) {
		throw new Error(`layout: the preview offers no ${missing.join(", ")}`);
	}
	const others = offered.filter((language) => !always.includes(language));
	const picked =
		others[Number.parseInt(commit.slice(0, 8), 16) % others.length];
	return picked === undefined ? always : [...always, picked];
}

/**
 * commitOf is a commit asked for to pick a language by: seven to forty
 * lowercase hexadecimal digits.
 */
export function commitOf(asked: string): string {
	if (!/^[0-9a-f]{7,40}$/.test(asked)) {
		throw new Error(`layout: ${asked} is no commit to pick a language by`);
	}
	return asked;
}

/**
 * Asked is what a run measures: the languages asked for, or else those a
 * commit samples, or else every one; the widths; and the shard.
 */
export type Asked = {
	languages?: string[];
	sample?: string;
	widths: number[];
	shard: Shard;
};

// pageIn is the page a browser draws the preview's cards in, its clock held
// still. The spinner of a waiting card turns, and a turned square stands out
// of its box at every angle but a right one: the cards are measured as the
// widget draws them for a reader who asks for no motion.
async function pageIn(browser: Browser): Promise<Page> {
	const context = await browser.newContext({
		viewport: { width: 1280, height: 900 },
		reducedMotion: "reduce",
	});
	await stillClock(context);
	return context.newPage();
}

// measuredIn measures the pages of the preview in one engine that are its
// shard's share: at each width asked for, the page of each language asked
// for. What each page shows joins measured as soon as the page is done, and a
// line says how long it took, so that a run that stops shows how far it came.
// A page whose browser stopped answering is measured again in a new browser,
// once, and a line says so.
async function measuredIn(
	engine: string,
	base: string,
	asked: Asked,
	measured: Measured[],
): Promise<void> {
	const type = engines[engine];
	if (type === undefined) {
		throw new Error(`layout: no engine ${engine}`);
	}
	let browser = await type.launch();
	try {
		let page = await pageIn(browser);
		await page.goto(`${base}preview.html`);
		const offered = await within(
			readWait,
			"reading the languages offered",
			page
				.locator('select[aria-label="Language"] option')
				.evaluateAll(valuesOf),
		);
		const offeredWidths = await within(
			readWait,
			"reading the widths offered",
			page
				.locator('input[type="checkbox"][value]')
				.evaluateAll((boxes) =>
					boxes.map((box) => Number(box.getAttribute("value"))),
				),
		);
		for (const [what, some, all] of [
			["language", asked.languages ?? [], offered],
			["width", asked.widths, offeredWidths],
		] as const) {
			const missing = unoffered<string | number>(some, all);
			if (missing.length > 0) {
				throw new Error(
					`layout: the preview offers no ${what} ${missing.join(", ")}; it offers ${all.join(", ")}`,
				);
			}
		}
		const languages =
			asked.sample === undefined
				? offered.filter(
						(language) =>
							asked.languages === undefined ||
							asked.languages.includes(language),
					)
				: sampled(offered, asked.sample);
		if (asked.sample !== undefined) {
			const said = `measured in ${languages.join(", ")}, the last picked by commit ${asked.sample.slice(0, 7)}`;
			console.log(`layout: ${engine} ${said}`);
			await writeFile(join(out, "languages.txt"), `Widget layout ${said}.\n`);
		}
		const pages = pagesOf(engine, languages, asked.widths);
		for (const shown of shareOf(pages, asked.shard)) {
			const started = Date.now();
			const shows = await againIfLate(
				() => measuredOn(page, base, shown),
				async (late) => {
					console.log(
						`layout: ${pageOf(shown)}: ${late.message}; measuring it again in a new browser`,
					);
					await within(readWait, "closing the browser", browser.close()).catch(
						() => {},
					);
					browser = await type.launch();
					page = await pageIn(browser);
				},
			);
			measured.push(...shows);
			const cards = new Set(shows.map(({ scene }) => scene)).size;
			const seconds = Math.round((Date.now() - started) / 1000);
			console.log(`layout: ${pageOf(shown)}: ${cards} cards in ${seconds} s`);
		}
	} finally {
		await within(readWait, "closing the browser", browser.close()).catch(
			() => {},
		);
	}
}

// cardOf names the card a measurement was made on.
function cardOf(measured: Measured): string {
	return `${pageOf(measured)}, ${measured.scene}`;
}

// pageOf names a page of the preview: its engine, its language and its width.
function pageOf({ engine, language, width }: Shown): string {
	return `${engine} ${language} ${width}px`;
}

/**
 * valuesOf are the values a list's options offer, read from each option
 * itself: where an option's value is its text, as each of the preview's
 * languages is, Preact writes no value attribute, and the option reads its
 * value from the text. It runs in the page, so it uses nothing from outside.
 */
export function valuesOf(options: HTMLOptionElement[]): string[] {
	return options.map((option) => option.value);
}

/**
 * unoffered are those of the things asked for that are not among the ones
 * offered: a language or a width the preview does not show, which a run would
 * otherwise wait for in vain, or pass over and call a pass.
 */
export function unoffered<T>(asked: readonly T[], offered: readonly T[]): T[] {
	return asked.filter((thing) => !offered.includes(thing));
}

/** Shard is one of the equal parts a run is shared out in: the part-th of parts. */
export type Shard = { part: number; parts: number };

// whole is a run measured as one part.
const whole: Shard = { part: 1, parts: 1 };

/**
 * shareOf is the share of the pages one shard measures: every parts-th of
 * them, from the part-th on. The shards of a run between them measure every
 * page once, and no shard measures more than one page more than another.
 */
export function shareOf<T>(pages: readonly T[], { part, parts }: Shard): T[] {
	return pages.filter((_, at) => at % parts === part - 1);
}

/**
 * pagesOf are the pages of one engine, width by width: at each width, the page
 * of each language. Shared out in this order, every shard measures at every
 * width while there are at least as many languages as shards, so that none is
 * left with all of one width — the one that costs the most to draw, or the
 * one photographed.
 */
export function pagesOf(
	engine: string,
	languages: readonly string[],
	widths: readonly number[],
): Shown[] {
	return widths.flatMap((width) =>
		languages.map((language) => ({ engine, language, width })),
	);
}

/**
 * told are the findings of what was measured, each once, where it was first
 * seen: a finding seen at every moment of a scene is one finding.
 */
export function told(measured: Measured[]): string[] {
	const lines = new Map<string, string>();
	for (const card of measured) {
		for (const { what, where, by } of card.findings) {
			const line = `${cardOf(card)}: ${where} ${what} by ${by}px`;
			if (!lines.has(line)) {
				lines.set(line, `${line} (${card.moment})`);
			}
		}
	}
	return [...lines.values()];
}

async function main(): Promise<void> {
	const { values } = parseArgs({
		options: {
			engine: { type: "string", multiple: true },
			language: { type: "string", multiple: true },
			width: { type: "string", multiple: true },
			shard: { type: "string" },
			sample: { type: "string" },
		},
	});
	if (values.sample !== undefined && values.language !== undefined) {
		throw new Error("layout: --sample picks the languages; ask for none");
	}
	const asked: Asked = {
		languages: values.language,
		sample: values.sample === undefined ? undefined : commitOf(values.sample),
		widths: values.width?.map(widthOf) ?? widths,
		shard: values.shard === undefined ? whole : shardOf(values.shard),
	};

	await rm(out, { recursive: true, force: true });
	await mkdir(out, { recursive: true });
	// The preview reads the version from the environment it is started in, as
	// the widget's own build does.
	process.env.VITE_VERSION = longestVersion;
	const preview = await served();
	// What was measured is written down however the run ends, so that a run
	// stopped halfway leaves what it saw.
	// Each engine runs to its end, so that the cards one fails on hide nothing
	// the other finds; the first refusal stands once both are done.
	const measured: Measured[] = [];
	let ended: PromiseSettledResult<void>[];
	try {
		ended = await Promise.allSettled(
			(values.engine ?? Object.keys(engines)).map((engine) =>
				measuredIn(engine, preview.base, asked, measured),
			),
		);
	} finally {
		await writeFile(
			join(out, "report.json"),
			`${JSON.stringify(measured, null, "\t")}\n`,
		);
		await preview.close();
	}
	for (const end of ended) {
		if (end.status === "rejected") {
			throw end.reason;
		}
	}
	const findings = told(measured);
	for (const line of findings) {
		console.log(line);
	}
	const cards = new Set(measured.map(cardOf)).size;
	console.log(`layout: ${cards} cards, ${findings.length} findings`);
	// A run that measured nothing has shown nothing fits.
	if (findings.length > 0 || cards === 0) {
		process.exitCode = 1;
	}
}

/** widthOf is a width asked for, in pixels: a whole number above zero. */
export function widthOf(asked: string): number {
	const width = Number(asked);
	if (!Number.isInteger(width) || width <= 0) {
		throw new Error(`layout: ${asked} is no width in pixels`);
	}
	return width;
}

/** shardOf is a shard asked for as part/parts: 2/6 is the second of six. */
export function shardOf(asked: string): Shard {
	const numbers = /^(\d+)\/(\d+)$/.exec(asked);
	const part = Number(numbers?.[1]);
	const parts = Number(numbers?.[2]);
	if (numbers === null || part < 1 || part > parts) {
		throw new Error(`layout: ${asked} is no shard, such as 2/6`);
	}
	return { part, parts };
}

if (import.meta.main) {
	await main().catch((error: unknown) => {
		console.error(error);
		process.exitCode = 1;
	});
	// A browser that stopped answering may still hold its pipe open, which
	// would keep the run from ending: it ends here, with its own exit code.
	process.exit();
}
