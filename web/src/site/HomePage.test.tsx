import { type Element, Window } from "happy-dom";
import { afterAll, afterEach, describe, expect, test, vi } from "vitest";
import topics from "../../../content/catalogs/topics.json";
import traps from "../../../content/catalogs/traps.json";
import file from "../../../site/data.json";
import { cardWords } from "../widget/dictionaries";
import { trapName } from "../widget/names";
import { address, outputPath } from "./addresses";
import {
	chatGPTDeveloperModeURL,
	claudeConnectorsURL,
	connectorURL,
	sourceURL,
} from "./brand";
import { frontPage } from "./content";
import { readSiteData } from "./data";
import type { Frame } from "./frame";
import { connectSection, lessonSection } from "./home";
import { type Page, sitePages } from "./pages";
import { renderSite } from "./render";
import { gradesOf } from "./topics";

const browser = new Window({
	settings: {
		disableCSSFileLoading: true,
		disableJavaScriptFileLoading: true,
		handleDisabledFileLoadingAsSuccess: true,
	},
});

afterAll(async () => {
	await browser.happyDOM.close();
});

// own is the site's data with nothing changed in it, but no page of a topic
// published: the home page draws its lesson's card, the traps it names and
// the progress.
const own = {
	groups: file.groups,
	examples: {},
	progress: file.progress,
	home: file.home,
};

// dataOf is the site's data over the service's catalog, read from file.
const dataOf = (data: unknown) =>
	readSiteData(
		{
			topics: topics.map((topic) => ({ ...topic, site_page: false })),
			traps,
			tasks: [],
		},
		data,
	);

// words are the page's words, the same in each language: every block the page
// draws, with the slots it fills.
const words = [
	"title: Home",
	"description: The home page.",
	"hero:",
	"  free: Free",
	"  open: Open",
	"  title: Olympiad maths",
	"  lead: For grades {range}.",
	"  see: See a lesson",
	"  note: Works in Claude.",
	"  caption: The card.",
	"chat:",
	"  title: Your chat",
	"  ask: A task, please",
	"  progress: How is it going?",
	"pillars:",
	"  title: Why not the chat",
	"  lead: Because.",
	"  items:",
	"    - title: Endless",
	"      text: Tasks.",
	"    - title: Checked",
	"      text: By a program.",
	"    - title: A diagnosis",
	"      text: Traps {first}; {second}; {third}.",
	"lesson:",
	"  eyebrow: A lesson",
	"  title: One at a time",
	"  lead: Step by step.",
	"  steps:",
	"    generating:",
	"      title: Ask",
	"      text: Type.",
	"    selected:",
	"      title: Answer",
	"      text: Tap.",
	"    hint:",
	"      title: Stuck",
	"      text: A way in.",
	"    wrong:",
	"      title: Wrong",
	"      text: A trap.",
	"    question:",
	"      title: Why",
	"      text: Ask in the chat.",
	"    progress:",
	"      title: Progress",
	"      text: Ranks.",
	"  child: Comet",
	"  chat:",
	"    label: An illustration",
	"    adult: The adult",
	"    model: The model",
	"    question: Why one more?",
	"    reply: Count the posts.",
	"task:",
	"  question: How many posts?",
	"  hint: Draw a smaller fence.",
	"  traps:",
	"    a: Gave the gap.",
	"    b: Counted the gaps.",
	"    d: Multiplied.",
	"    e: Took the length.",
	"  solution: Four gaps. One more post. Five posts.",
	"connect:",
	"  eyebrow: Connect",
	"  title: A few steps",
	"  claude:",
	"    pill: Works now",
	"    steps:",
	"      open: Open {connectors}.",
	"      add: Press **+**.",
	'      paste: "Paste:"',
	"      sign: Sign in.",
	"      chat: Switch it on.",
	"    connectors: The connectors",
	"    note: One connector.",
	"  chatgpt:",
	"    pill: Not yet",
	"    text: Not listed.",
	"    dev: As {openai}.",
	"    openai: OpenAI describes",
	"  need: Two accounts.",
	"ask:",
	"  title: Start",
	"  lead: Add it.",
	"  code: The code",
].join("\n");

// document is a document's text under its title.
const document = (title: string) =>
	`---\ntitle: ${title}\ndescription: D\n---\n\n# ${title}\n`;

// sources are the site's texts: the home page's words in English and in
// Russian, and the documents.
const sources = new Map(
	["en", "ru"].map((locale) => [
		locale,
		new Map([
			["index.yaml", words],
			["privacy.md", document("Privacy")],
			["terms.md", document("Terms")],
		]),
	]),
);

// A frame shaped like the site's own over the one page this site has: the
// menu names pages, none of which is built here, and the header's button
// leads to connecting.
const frame: Frame = {
	menu: [],
	action: { page: frontPage, anchor: connectSection, label: "nav.add" },
	footer: ["privacy", "terms"],
};

// render builds the site over data, its home page the site's own.
const render = (data = dataOf(own)) => {
	const home = sitePages(data).get(frontPage);
	if (home === undefined) {
		throw new Error("the site's pages have no home page");
	}
	const pages = new Map<string, Page>([[frontPage, home]]);
	return renderSite({
		base: "https://example.test",
		sources,
		pages,
		frame,
		data,
	});
};

const files = render();

// homeFile is the file the home page in locale is served from.
const homeFile = (locale: string) => outputPath(address(locale, frontPage));

// pageIn is the home page in locale, read as a document.
const pageIn = (locale: string) =>
	new browser.DOMParser().parseFromString(
		files.find(({ path }) => path === homeFile(locale))?.data ?? "",
		"text/html",
	);

const home = pageIn("en");

// all are the values of name on every element selector finds.
const all = (selector: string, name: string) =>
	[...home.querySelectorAll(selector)].map(
		(element) => element.getAttribute(name) ?? "",
	);

// texts are the words of every element selector finds.
const texts = (selector: string) =>
	[...home.querySelectorAll(selector)].map((element) => element.textContent);

// the is the one element selector finds on the page.
const the = (selector: string): Element => {
	const found = home.querySelector(selector);
	if (found === null) {
		throw new Error(`the page has no ${selector}`);
	}
	return found;
};

// step is the step of the lesson at, counted from one, in the order the
// page shows them.
const step = (at: number): Element => {
	const found = home.querySelectorAll(".s-walk-step")[at - 1];
	if (found === undefined) {
		throw new Error(`the lesson has no step ${at}`);
	}
	return found;
};

// statesIn are the options of the card in root, each by its letter and how
// the card marks it.
const statesIn = (root: Element) =>
	[...root.querySelectorAll(".mt-option")].map(
		(row) =>
			`${row.querySelector(".mt-option-letter")?.textContent} ${row.getAttribute("data-state")}`,
	);

describe("the home page", () => {
	// The page reads in full as it is built: its one script brings it alive,
	// and the data it carries for the script is JSON no browser runs.
	test("reads in full as it is built, its cards inert, and loads the cards' stylesheet", () => {
		expect(all("script", "type")).toEqual(["application/json", "module"]);
		expect(all("script[src]", "src")).toEqual(["/assets/demo.js"]);
		expect(all('link[rel="stylesheet"]', "href")).toContain("/assets/card.css");
		expect(home.querySelectorAll(".s-card")).toHaveLength(7);
		expect(home.querySelectorAll(".s-card:not([inert])")).toHaveLength(0);
	});

	// A page keeps no profile: a card on it says whose it is, and leads to no
	// progress.
	test("draws no card that offers the progress", () => {
		expect(home.querySelectorAll(".s-card .mt-bar")).toHaveLength(6);
		expect(home.querySelector(".s-card button.mt-bar")).toBeNull();
		expect(home.querySelector(".s-card .mt-bar-action")).toBeNull();
	});

	describe("built", () => {
		afterEach(() => {
			vi.unstubAllEnvs();
		});

		// homeBuiltWith is the home page in English of a site built told the
		// version given, as the release's build is told it.
		const homeBuiltWith = (version: string) => {
			vi.stubEnv("VITE_VERSION", version);
			return new browser.DOMParser().parseFromString(
				render().find(({ path }) => path === homeFile("en"))?.data ?? "",
				"text/html",
			);
		};

		test("from a release, names it on every card MathTrail heads, as a chat does", () => {
			const page = homeBuiltWith("v0.2.1");
			const heads = [...page.querySelectorAll(".mt-head")].filter(
				(head) => head.querySelector(".mt-name")?.textContent === "MathTrail",
			);

			expect(heads.length).toBeGreaterThan(0);
			expect(
				heads.map((head) => head.querySelector(".mt-version")?.textContent),
			).toEqual(heads.map(() => "version 0.2.1"));
			expect(page.querySelectorAll(".mt-version")).toHaveLength(heads.length);
		});

		test("from no release, names no build on any card, rather than dev", () => {
			const page = homeBuiltWith("");

			expect(page.querySelector(".mt-head")).not.toBeNull();
			expect(page.querySelector(".mt-version")).toBeNull();
		});
	});

	test("names every part of itself once, though it draws seven cards", () => {
		const ids = all("[id]", "id");

		expect(new Set(ids).size).toBe(ids.length);
	});

	test("has the sections on a lesson and on connecting the page and the header's button lead to", () => {
		expect(home.getElementById(lessonSection)?.tagName).toBe("SECTION");
		expect(home.getElementById(connectSection)?.tagName).toBe("SECTION");
		expect(all(".s-nav-action", "href")).toEqual(["/#connect"]);
		expect(texts(".s-nav-action")).toEqual(["Add to Claude"]);
	});

	test("opens with the grades of the catalog, the way to add it and the way to see a lesson", () => {
		const grades = topics.flatMap((topic) => gradesOf(topic.grade_levels));
		const range = `${Math.min(...grades)}–${Math.max(...grades)}`;

		expect(texts(".s-hero .s-chip")).toEqual([
			"Free",
			"Open",
			`Grades ${range}`,
		]);
		expect(texts(".s-hero .s-lead")).toEqual([`For grades ${range}.`]);
		expect(all(".s-hero-actions a", "href")).toEqual(["/#connect", "/#lesson"]);
		expect(texts(".s-hero-actions a")).toEqual([
			"Add to Claude",
			"See a lesson",
		]);
	});

	test("shows beside its first screen the task as it arrives, in the chat under the parent's message", () => {
		const hero = the(".s-hero-card");

		expect(hero.querySelector(".s-frame .s-message-adult")?.textContent).toBe(
			"A task, please",
		);
		expect(statesIn(hero)).toEqual([
			"A default",
			"B default",
			"C default",
			"D default",
			"E default",
		]);
		expect(hero.querySelector(".mt-note-hint")).toBeNull();
		expect(
			hero.querySelector("svg.mt-picture")?.getAttribute("aria-label"),
		).toBe("Things in a row");
	});

	test("holds the chat of its first screen in a phone, all but the chat a drawing a screen reader skips, in the track the demo holds it in", () => {
		const phone = the(".s-hero-track > .s-hero > .s-hero-card > .s-phone");

		expect(
			phone.querySelector(".s-phone-screen > .s-frame > .s-frame-body .s-card"),
		).not.toBeNull();
		expect(
			[...phone.querySelectorAll(".s-phone-screen > *")].map((part) => [
				part.className,
				part.getAttribute("aria-hidden"),
			]),
		).toEqual([
			["s-phone-status", "true"],
			["s-frame", null],
			["s-phone-composer", "true"],
			["s-phone-home", "true"],
		]);
		expect(home.querySelectorAll(".s-hero-track")).toHaveLength(1);
	});

	test("draws the button of the topic on every card of its task, the coach choosing, locked: the page lets no topic be chosen", () => {
		const cards = [...home.querySelectorAll(".mt-widget")].filter(
			(widget) => widget.querySelector(".mt-option") !== null,
		);
		const buttons = cards.map((widget) =>
			widget.querySelector(".mt-topic-button")?.getAttribute("aria-disabled"),
		);

		expect(cards.length).toBeGreaterThan(1);
		expect(buttons).toEqual(cards.map(() => "true"));
	});

	test("names the catalog's traps a wrong option is tied to as the card names them", () => {
		const card = cardWords("en", undefined);

		expect(texts(".s-tiles .s-tile-text")[2]).toBe(
			`Traps ${file.home.traps.map((id) => trapName(card, id)).join("; ")}.`,
		);
	});

	test("goes through a lesson step by step, numbering each but those within the one before", () => {
		expect(texts(".s-walk-copy h3")).toEqual([
			"Ask",
			"Answer",
			"Stuck",
			"Wrong",
			"Why",
			"Progress",
		]);
		expect(texts(".s-walk-number")).toEqual(["1", "2", "", "3", "", "4"]);
		expect(home.querySelectorAll(".s-walk-within")).toHaveLength(2);
	});

	test("shows the task being written at the first step", () => {
		expect(
			step(1).querySelector('article[aria-label="Next task"]'),
		).not.toBeNull();
		expect(step(1).querySelectorAll(".mt-option")).toHaveLength(0);
	});

	test("shows the option the steps pick being checked at the second", () => {
		expect(statesIn(step(2))).toEqual([
			"A default",
			`${file.home.card.choice} selected`,
			"C default",
			"D default",
			"E default",
		]);
	});

	test("shows the hint open, the answer still to give, within the second", () => {
		expect(step(3).querySelector(".mt-note-hint")?.textContent).toContain(
			"Draw a smaller fence.",
		);
		expect(statesIn(step(3))).not.toContain("C correct");
	});

	test("shows the wrong answer gone over at the third: the card's ask in the chat, and the card of how it went, with its trap, the picture of the solution and the solution step by step", () => {
		expect(step(4).querySelector(".s-message-adult")?.textContent).toBe(
			"Go over the answer",
		);
		expect(step(4).querySelector(".mt-verdict-line")?.textContent).toBe(
			"Not quite — it's 5, not 4.",
		);
		expect(step(4).querySelector(".mt-note-trap")?.textContent).toContain(
			"Counted the gaps.",
		);
		expect(
			step(4).querySelector(".mt-solution-picture svg.mt-picture"),
		).not.toBeNull();
		expect(step(4).querySelector(".mt-total")?.textContent).toBe(
			"12 ÷ 3 + 1 = 5, not 4",
		);
		expect(step(4).querySelectorAll(".mt-steps li")).toHaveLength(3);
		expect(statesIn(step(4))).toEqual([]);
	});

	test("puts the question the child asks once the answer is in under the card, as messages of the chat labelled an illustration", () => {
		const frame = step(5).querySelector(".s-frame-body");
		const parts = [...(frame?.children ?? [])].map(
			(part) => part.getAttribute("class") ?? "",
		);

		expect(parts).toEqual(["s-message s-message-adult", "s-card", "s-chat"]);
		expect(step(5).querySelector(".mt-note-trap")).not.toBeNull();
		expect(step(5).querySelector(".s-chat figcaption")?.textContent).toBe(
			"An illustration",
		);
		expect(
			[...step(5).querySelectorAll(".s-chat .s-bubble")].map(
				(bubble) => bubble.textContent,
			),
		).toEqual(["Why one more?", "Count the posts."]);
	});

	test("shows the progress at the last step, asked for by the parent, its topics and repeating mistakes open", () => {
		expect(step(6).querySelector(".s-message-adult")?.textContent).toBe(
			"How is it going?",
		);
		expect(
			[
				...step(6).querySelectorAll('.mt-fold-button[aria-expanded="true"]'),
			].map((button) => button.querySelector(".mt-fold-title")?.textContent),
		).toEqual(["Topics", "Mistakes that repeat"]);
	});

	test("connects to Claude step by step, the connector's address in full after the third", () => {
		const steps = [...home.querySelectorAll(".s-connect-steps li")];

		expect(
			steps.map((li) => li.querySelector(".s-connect-text")?.textContent),
		).toEqual([
			"Open The connectors.",
			"Press +.",
			"Paste:",
			"Sign in.",
			"Switch it on.",
		]);
		expect(
			steps.map((li) => li.querySelector("code")?.textContent ?? ""),
		).toEqual(["", "", connectorURL, "", ""]);
		expect(all(".s-address", "dir")).toEqual(["ltr"]);
		expect(steps[0]?.querySelector("a")?.getAttribute("href")).toBe(
			claudeConnectorsURL,
		);
	});

	// A button that copies needs a script: the page keeps it in a template
	// beside the address, which shows nothing until the demo puts it there.
	test("keeps the button that copies the address beside it, unshown until the demo puts it", () => {
		const kept = (
			home.querySelector(".s-address + template[data-copy]") as unknown as {
				content: Pick<Element, "querySelector">;
			} | null
		)?.content;

		expect(home.querySelector(".s-copy")).toBeNull();
		expect(kept?.querySelector(".s-copy")?.textContent).toBe("Copy");
		expect(kept?.querySelector(".s-copy")?.getAttribute("data-done")).toBe(
			"Copied",
		);
		expect(kept?.querySelector("[aria-live]")?.getAttribute("aria-live")).toBe(
			"polite",
		);
	});

	test("tells how ChatGPT adds it, as OpenAI describes", () => {
		expect(all(".s-connect-muted a", "href")).toEqual([
			chatGPTDeveloperModeURL,
		]);
	});

	test("closes with the way to add it and the way to its code", () => {
		expect(all(".s-ask a", "href")).toEqual(["/#connect", sourceURL]);
		expect(texts(".s-ask a")).toEqual(["Add to Claude", "The code"]);
	});

	test("draws its cards in the page's language", () => {
		expect(home.querySelector(".s-card .mt-badge")?.textContent).toBe(
			"Olympiad coach · Grade 3",
		);
		expect(pageIn("ru").querySelector(".s-card .mt-badge")?.textContent).toBe(
			"Олимпиадный тренер · 3 класс",
		);
	});

	test("is refused when the site's data has nothing for it", () => {
		const { home: _, ...without } = own;

		expect(() => render(dataOf(without))).toThrow(
			"site/data.json has nothing for the home page",
		);
	});
});
