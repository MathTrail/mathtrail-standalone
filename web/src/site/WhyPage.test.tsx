import { Window } from "happy-dom";
import { afterAll, describe, expect, test } from "vitest";
import topics from "../../../content/catalogs/topics.json";
import traps from "../../../content/catalogs/traps.json";
import file from "../../../site/data.json";
import { cardWords } from "../widget/dictionaries";
import { topicName } from "../widget/names";
import { readSiteData } from "./data";
import type { Frame } from "./frame";
import { type Page, sitePages } from "./pages";
import { renderSite } from "./render";
import { gradesOf } from "./topics";
import { doiAddress } from "./why";

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

// dataWith is the site's data over the service's catalog, with no example of
// any topic's page, and why as what the page Why draws: its card, its topics'
// names and its works.
const dataWith = (why: typeof file.why | undefined) =>
	readSiteData(
		{
			topics: topics.map((topic) => ({ ...topic, site_page: false })),
			traps,
			tasks: [],
		},
		{
			groups: file.groups,
			examples: {},
			progress: file.progress,
			why,
		},
	);

const data = dataWith(file.why);

// words are the page's words, the same in each language: every block the page
// draws, with the slots it fills, and a finding for each work the data names.
const words = [
	"title: Why",
	"description: Why olympiad maths.",
	"honest: Honestly",
	"task:",
	"  question: How many posts?",
	"  trap: Counted the gaps.",
	"  solution: Four gaps. One more post. Five posts.",
	"hero:",
	"  title: Why olympiad maths",
	"  lead: Up to grade {last}.",
	"  free: Free",
	"  open: Open",
	"  school:",
	"    label: A sum",
	"    text: Divide.",
	"    sum: 12 ÷ 3 = 4",
	"  olympiad:",
	"    label: A problem",
	'    answer: "Answer: {answer}"',
	"    trap: A trap",
	"    text: Understand it.",
	"  caption: Same numbers.",
	"research:",
	"  title: What research found",
	"  lead: Few studies.",
	"  findings:",
	...file.why.findings.flatMap((id) => [
		`    ${id}:`,
		`      title: Found by ${id}`,
		"      text: What it found.",
	]),
	"  note: Mostly observation.",
	"school:",
	"  title: School and olympiads",
	"  lead: Different jobs.",
	"  lesson: A lesson",
	"  olympiad: A problem",
	"  rows:",
	"    - label: What counts",
	"      lesson: Speed",
	"      olympiad: Understanding",
	"time:",
	"  title: No time",
	"  cards:",
	"    - title: A syllabus",
	"      text: For everyone.",
	"parent:",
	"  title: A parent can",
	"  note: Up to grade {last}.",
	"  unneeded:",
	"    title: Not needed",
	"    items:",
	"      - A degree",
	"    note: MathTrail's part.",
	"  needed:",
	"    title: Needed",
	"    items:",
	"      - title: Curiosity",
	"        text: Be curious.",
	"  explain:",
	"    title: How to explain",
	"    quotes:",
	"      - A riddle.",
	"thinking:",
	"  title: How it teaches",
	"  lead: Around each problem.",
	"  points:",
	"    - title: One idea",
	"      text: Topics such as {enumeration}, {parity} or {knights}.",
	"  topics: See the topics",
	"  child: Comet",
	"  caption: The card.",
	"  chat:",
	"    label: An illustration",
	"    adult: The adult",
	"    model: The model",
	"    question: Why not 6?",
	"    reply: One post twice.",
	"app:",
	"  title: No ordinary app",
	"  lead: Apps keep children playing ({source}).",
	"  cards:",
	"    - title: Checked",
	"      text: By a program.",
	"  code: The code is open.",
	"  github: GitHub",
	"ask:",
	"  title: Start",
	"  lead: Together.",
	"  topics: See the topics",
].join("\n");

// document is a document's text under its title.
const document = (title: string) =>
	`---\ntitle: ${title}\ndescription: D\n---\n\n# ${title}\n`;

// sourcesWith are the site's texts, with the page's words in Russian and in
// English, which are these unless others are given.
const sourcesWith = (russian = words, english = words) =>
	new Map(
		Object.entries({ en: english, ru: russian }).map(([locale, words]) => [
			locale,
			new Map([
				["index.md", document("Home")],
				["privacy.md", document("Privacy")],
				["terms.md", document("Terms")],
				["why.yaml", words],
			]),
		]),
	);

const pages = new Map<string, Page>([
	["why", sitePages(data).get("why") ?? { draw: () => <p /> }],
]);

// A menu of the one page this site has.
const frame: Frame = {
	menu: [{ page: "why", label: "nav.why" }],
	footer: ["privacy", "terms"],
};

const render = (sources = sourcesWith(), given = data) =>
	renderSite({
		base: "https://example.test",
		sources,
		pages,
		frame,
		data: given,
	});

const files = render();

// pageIn is the page Why in locale, read as a document, of the site built as
// given, or else of the one built here.
const pageIn = (locale: string, built = files) =>
	new browser.DOMParser().parseFromString(
		built.find(({ path }) => path === `${locale}/why/index.html`)?.data ?? "",
		"text/html",
	);

const why = pageIn("en");

// all are the values of name on every element selector finds.
const all = (selector: string, name: string) =>
	[...why.querySelectorAll(selector)].map(
		(element) => element.getAttribute(name) ?? "",
	);

// doi is where a work of the site's data is linked.
const doi = (id: string) =>
	doiAddress(file.why.sources[id as keyof typeof file.why.sources]);

describe("the page Why", () => {
	test("reads in full with no script", () => {
		expect(why.querySelector("script")).toBeNull();
	});

	// A page keeps no profile: its card says whose it is, and leads to no
	// progress.
	test("draws a card that offers no progress", () => {
		expect(why.querySelector(".s-card .mt-bar")?.tagName.toLowerCase()).toBe(
			"div",
		);
		expect(why.querySelector(".s-card .mt-bar-action")).toBeNull();
	});

	test("draws the widget's own card of how a wrong answer went, inert, and loads its stylesheet", () => {
		expect(why.querySelector(".s-card[inert] .mt-widget")).not.toBeNull();
		expect(all('link[rel="stylesheet"]', "href")).toContain("/assets/card.css");
		expect(why.querySelector(".s-card .mt-verdict-line svg")).not.toBeNull();
		expect(why.querySelector(".s-card .mt-note-trap")).not.toBeNull();
		expect(why.querySelector(".s-card .mt-total")?.textContent).toBe(
			"12 ÷ 3 + 1 = 5, not 4",
		);
		expect(why.querySelectorAll(".s-card .mt-steps li")).toHaveLength(3);
		expect(why.querySelector(".s-card .mt-rating-move")).not.toBeNull();
	});

	test("draws its card in the page's language", () => {
		expect(why.querySelector(".s-card .mt-badge")?.textContent).toBe(
			"Olympiad coach · Grade 3",
		);
		expect(pageIn("ru").querySelector(".s-card .mt-badge")?.textContent).toBe(
			"Олимпиадный тренер · 3 класс",
		);
	});

	test("puts the chat's question and reply under the card, labelled as an illustration", () => {
		const parts = [
			...why.querySelectorAll(".s-lesson .s-card, .s-lesson .s-chat"),
		];
		const chat = why.querySelector(".s-lesson .s-chat");

		expect(parts.map((part) => part.getAttribute("class"))).toEqual([
			"s-card",
			"s-chat",
		]);
		expect(why.querySelector(".s-card .s-chat")).toBeNull();
		expect(chat?.querySelector("figcaption")?.textContent).toBe(
			"An illustration",
		);
		expect(
			[...(chat?.querySelectorAll(".s-message") ?? [])].map((message) =>
				message.getAttribute("class"),
			),
		).toEqual(["s-message s-message-adult", "s-message s-message-model"]);
	});

	test("shows a finding for each work the data names, in its order, each linking its work", () => {
		expect(all(".s-finding .s-cite a", "href")).toEqual(
			file.why.findings.map(doi),
		);
	});

	test("cites the work behind what it says of children's apps", () => {
		expect(
			[...why.querySelectorAll(".s-intro-line a")].map((link) =>
				link.getAttribute("href"),
			),
		).toEqual([doi(file.why.apps)]);
	});

	test("cites a work by its one or two authors, or the first of more, and its year", () => {
		const citedIn = (page: ReturnType<typeof pageIn>) =>
			[...page.querySelectorAll(".s-cite a")].map((link) => link.textContent);
		// The site's works have two authors or more, so one of them is given
		// one author alone.
		const { nunes2007 } = file.why.sources;
		const alone = render(
			sourcesWith(),
			dataWith({
				...file.why,
				sources: {
					...file.why.sources,
					nunes2007: { ...nunes2007, authors: ["Nunes, T."] },
				},
			}),
		);

		expect(citedIn(why)).toContain("Agarwal and Gaule, 2020");
		expect(citedIn(why)).toContain("Rebholz et al., 2022");
		expect(citedIn(pageIn("ru"))).toContain("Rebholz и др., 2022");
		expect(citedIn(pageIn("en", alone))).toContain("Nunes, 2007");
		expect(citedIn(pageIn("ru", alone))).toContain("Nunes, 2007");
	});

	test("names the grades MathTrail is for from the catalog", () => {
		const grades = topics.flatMap((topic) => gradesOf(topic.grade_levels));

		expect(
			[...why.querySelectorAll(".s-hero .s-lead + .s-chips .s-chip")].map(
				(chip) => chip.textContent,
			),
		).toEqual([
			"Free",
			"Open",
			`Grades ${Math.min(...grades)}–${Math.max(...grades)}`,
		]);
	});

	test("names its example topics as the card names them", () => {
		const said = why.querySelector(".s-points")?.textContent ?? "";
		const card = cardWords("en", undefined);

		for (const id of Object.values(file.why.topics)) {
			expect(said).toContain(topicName(card, id));
		}
	});

	test("leads to connecting on the home page, and to every topic", () => {
		expect(all(".s-ask a", "href")).toEqual(["/#connect", "/en/topics/"]);
		expect(why.querySelector(".s-ask .s-btn-filled")?.textContent).toBe(
			"Add to Claude",
		);
	});

	test("is the menu's entry, marked as the page being read", () => {
		expect(all(".s-navlinks a", "href")).toEqual(["/en/why/"]);
		expect(all(".s-navlinks a", "aria-current")).toEqual(["page"]);
	});

	test("is refused when its Russian words lack one the English have", () => {
		const lacking = words.replace("  free: Free\n", "");

		expect(lacking).not.toBe(words);
		expect(() => render(sourcesWith(lacking))).toThrow("hero.free");
	});

	// The data may leave out the page's part, for a site built without the
	// page; a site that has the page and not its part names the file to fix,
	// rather than drawing an empty page.
	test("is refused when the site's data has nothing for it", () => {
		expect(() => render(sourcesWith(), dataWith(undefined))).toThrow(
			"site/data.json has nothing for the page Why",
		);
	});

	test("is refused when its words give a finding no work of the data backs", () => {
		const extra = words.replace(
			"  note: Mostly observation.",
			[
				"    smith2024:",
				"      title: Found by nobody",
				"      text: Nothing.",
				"  note: Mostly observation.",
			].join("\n"),
		);

		expect(() => render(sourcesWith(extra, extra))).toThrow(
			"the page never shows research.findings.smith2024",
		);
	});
});
