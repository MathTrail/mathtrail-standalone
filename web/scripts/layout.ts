// The widget's layout, measured in real browsers. Every scene of the preview —
// every state of a lesson on a card — is drawn in every language the widget
// speaks and in the pseudo-language, at the widths a card has to fit, in
// Chromium, the engine of Chrome and of Android's WebView, and in WebKit, the
// engine of Safari and of every browser on an iPhone. A card fails when its
// page scrolls sideways, when a part of it sticks out of the card, when text
// runs out of the box it is set in, or when text a card is written to show
// whole is cut short. The scenes are looked at again once a wait would have
// been given up on. Every card in a language written right to left is
// photographed, to be looked at.
//
//	node scripts/layout.ts [--engine chromium] [--language ar] [--width 320]
//
// It runs where the browsers are, inside the image of the pinned Playwright
// release, against the widget's sources as they are: the preview is served
// here, and serves the widget's own page to its frames.

import { mkdir, rm, writeFile } from "node:fs/promises";
import { join } from "node:path";
import { parseArgs } from "node:util";
import { type BrowserType, chromium, type Page, webkit } from "playwright-core";
import { type Card, letGo, served, settled, stillClock } from "./drive.ts";

/** Finding is one way the card of one frame does not fit. */
export type Finding = {
	/** what went wrong, in words. */
	what: string;
	/** where names the element: its tag, its classes and its first words. */
	where: string;
	/** by is how many pixels too far. */
	by: number;
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
// hidden from sight but read out, and not the drawing, which scrolls sideways
// inside its own frame when it is wider than the card.
function shownIn(card: Element): HTMLElement[] {
	return [...card.querySelectorAll<HTMLElement>("*")].filter((element) => {
		const box = element.getBoundingClientRect();
		return (
			element.closest(".mt-vh, .mt-diagram, [hidden]") === null &&
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
	if (element.clientWidth === 0) {
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

// measuring is the expression a card's page evaluates to its findings.
const measuring = `(() => {
${[describe, shownIn, outOfTheCard, outOfItsBox, findingsIn].join("\n")}
return findingsIn();
})()`;

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
type Shown = { engine: string; language: string; width: number };

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
			await page.clock.fastForward(moment.after);
		}
		const cards = await settled(page);
		for (const [at, card] of cards.entries()) {
			const findings = (await card.frame.evaluate<Finding[]>(measuring)).filter(
				(finding) => !allowed(finding),
			);
			const scene = card.scene;
			measured.push({ ...shown, scene, moment: moment.name, findings });
			// Taking a picture hides the caret of a field, and leaves an empty
			// style behind on it, which is no change of the card's own.
			const markup = card.markup.replaceAll(' style=""', "");
			if (
				photographs &&
				pictured.get(scene) !== markup &&
				(await card.frame.evaluate(() => document.documentElement.dir)) ===
					"rtl"
			) {
				pictured.set(scene, markup);
				await photograph(card, at, join(out, engine, language), moment.name);
			}
		}
		await letGo(cards);
	}
	return measured;
}

// measuredIn measures every page of the preview in one engine: each language
// it offers that is asked for, at each width asked for.
async function measuredIn(
	engine: string,
	base: string,
	asked: { languages?: string[]; widths: number[] },
): Promise<Measured[]> {
	const type = engines[engine];
	if (type === undefined) {
		throw new Error(`layout: no engine ${engine}`);
	}
	const browser = await type.launch();
	try {
		// The spinner of a waiting card turns, and a turned square stands out
		// of its box at every angle but a right one: the cards are measured as
		// the widget draws them for a reader who asks for no motion.
		const context = await browser.newContext({
			viewport: { width: 1280, height: 900 },
			reducedMotion: "reduce",
		});
		await stillClock(context);
		const page = await context.newPage();
		await page.goto(`${base}preview.html`);
		const offered = await page
			.locator('select[aria-label="Language"] option')
			.evaluateAll((options) =>
				options.map((option) => option.getAttribute("value") ?? ""),
			);
		const offeredWidths = await page
			.locator('input[type="checkbox"][value]')
			.evaluateAll((boxes) =>
				boxes.map((box) => Number(box.getAttribute("value"))),
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
		const measured: Measured[] = [];
		for (const language of offered) {
			if (
				asked.languages !== undefined &&
				!asked.languages.includes(language)
			) {
				continue;
			}
			for (const width of asked.widths) {
				measured.push(
					...(await measuredOn(page, base, { engine, language, width })),
				);
			}
		}
		return measured;
	} finally {
		await browser.close();
	}
}

// cardOf names the card a measurement was made on.
function cardOf({ engine, language, width, scene }: Measured): string {
	return `${engine} ${language} ${width}px, ${scene}`;
}

/**
 * unoffered are those of the things asked for that are not among the ones
 * offered: a language or a width the preview does not show, which a run would
 * otherwise wait for in vain, or pass over and call a pass.
 */
export function unoffered<T>(asked: readonly T[], offered: readonly T[]): T[] {
	return asked.filter((thing) => !offered.includes(thing));
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
		},
	});
	const asked = {
		languages: values.language,
		widths: values.width?.map(widthOf) ?? widths,
	};

	await rm(out, { recursive: true, force: true });
	await mkdir(out, { recursive: true });
	// The preview reads the version from the environment it is started in, as
	// the widget's own build does.
	process.env.VITE_VERSION = longestVersion;
	const preview = await served();
	let measured: Measured[];
	try {
		const byEngine = await Promise.all(
			(values.engine ?? Object.keys(engines)).map((engine) =>
				measuredIn(engine, preview.base, asked),
			),
		);
		measured = byEngine.flat();
	} finally {
		await preview.close();
	}

	await writeFile(
		join(out, "report.json"),
		`${JSON.stringify(measured, null, "\t")}\n`,
	);
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

if (import.meta.main) {
	await main();
}
