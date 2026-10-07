import { readFileSync } from "node:fs";
import { dirname, join } from "node:path";
import { fileURLToPath } from "node:url";
import { Window } from "happy-dom";
import { afterAll, describe, expect, test } from "vitest";
import file from "../../../site/data.json";
import fixture from "../../../site/research/testdata/research.json";
import { letters } from "../widget/choices";
import { readSiteData, type SiteData, siteCatalog } from "./data";
import type { Frame } from "./frame";
import { sitePages } from "./pages";
import { runsOf } from "./ResearchLive";
import { barScale, linePlaces, type Place } from "./ResearchModel";
import { researchSections } from "./ResearchPage";
import { secondRun } from "./ResearchTheses";
import { renderSite } from "./render";
import type { Research, Row } from "./research";
import { handTyped } from "./testing/handtyped";
import { answer, example, fence, fencePosts } from "./worked";

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

// textOf is a file of the repository, by its path from the repository's root:
// the page's own words, and the paper's sources its title is held to.
const textOf = (...path: string[]) =>
	readFileSync(
		join(dirname(fileURLToPath(import.meta.url)), "..", "..", "..", ...path),
		"utf8",
	);
const english = textOf("site", "content", "en", "research.yaml");
const russian = textOf("site", "content", "ru", "research.yaml");
const tex = textOf("research", "paper-a", "main.tex");
const russianDraft = textOf("research", "paper-a", "draft.ru.md");

// names are the names on the page that carry digits of their own.
const names = ["XChaCha20-Poly1305"];

// withPaper is the fixture as the bench writes it, the paper's PDF named;
// withoutPaper is the same with no file of the paper shipped.
const withPaper = dataOf(fixture);
const withoutPaper = dataOf({
	...fixture,
	paper: { ...fixture.paper, files: [] },
});

// dataOf is the site's data with the numbers of the page Research, over the
// service's catalog with no topic's page published: the site of these tests
// has the one page.
function dataOf(research: unknown): SiteData {
	return readSiteData(
		{
			...siteCatalog,
			topics: siteCatalog.topics.map((topic) => ({
				...topic,
				site_page: false,
			})),
		},
		{
			groups: file.groups,
			examples: {},
			progress: file.progress,
			research: file.research,
		},
		research,
	);
}

// document is a document's text under its title.
const document = (title: string) =>
	`---\ntitle: ${title}\ndescription: D\n---\n\n# ${title}\n`;

// sources are the site's texts, the page's own words in both its languages.
const sources = new Map(
	Object.entries({ en: english, ru: russian }).map(([locale, words]) => [
		locale,
		new Map([
			["index.md", document("Home")],
			["privacy.md", document("Privacy")],
			["terms.md", document("Terms")],
			["research.yaml", words],
		]),
	]),
);

// A menu of the one page this site has.
const frame: Frame = {
	menu: [{ page: "research", label: "nav.research" }],
	footer: ["privacy", "terms"],
};

// pagesOf are the pages of the site drawn from data, the page Research in
// every language, read as documents.
function pagesOf(data: SiteData): Map<string, Document> {
	const page = sitePages(data).get("research");
	if (page === undefined) {
		throw new Error("the site draws no page Research");
	}
	const files = renderSite({
		base: "https://example.test",
		sources,
		pages: new Map([["research", page]]),
		frame,
		data,
	});
	return new Map(
		["en", "ru"].map((locale) => [
			locale,
			new browser.DOMParser().parseFromString(
				files.find(({ path }) => path === `${locale}/research/index.html`)
					?.data ?? "",
				"text/html",
			) as unknown as Document,
		]),
	);
}

const pages = pagesOf(withPaper);
const page = pages.get("en") as Document;
const russianPage = pages.get("ru") as Document;
const unshipped = pagesOf(withoutPaper);

// shownNothing is what live numbers that show none have of them.
const shownNothing = { total: null, chances: [], kept_up: [] };

// The page with its live numbers in each of their other states: still to
// come; too few in the latest month; a month shown with none of its ranges;
// and a month shown with the range of every answer past the others.
const coming = pagesOf(
	dataOf({
		...fixture,
		live: { ...fixture.live, state: "coming", month: null, ...shownNothing },
	}),
);
const tooFew = pagesOf(
	dataOf({
		...fixture,
		live: { ...fixture.live, state: "too_few", ...shownNothing },
	}),
);
const noRanges = pagesOf(
	dataOf({ ...fixture, live: { ...fixture.live, chances: [], kept_up: [] } }),
);
const openRange = pagesOf(
	dataOf({
		...fixture,
		live: {
			...fixture.live,
			kept_up: [
				...fixture.live.kept_up,
				{
					first: 201,
					last: null,
					learners: 10,
					answers: 30,
					came_true_less_promised: 0.05,
					standard_error: 0.06,
				},
			],
		},
	}),
);

// notReached is the page in English with a goal of the screen that the
// service misses wholly, so that every mark has a row to stand on.
const notReached = pagesOf(
	dataOf({
		...fixture,
		bench: {
			...fixture.bench,
			rows: fixture.bench.rows.map((row) =>
				row.id === "card_move_static"
					? {
							...row,
							bound: { value: 30, own: false },
							values: {
								...row.values,
								service: { ...row.values.service, mark: "not_reached" },
							},
						}
					: row,
			),
		},
	}),
).get("en") as Document;

// values are the values of every element selector finds on a page.
const values = (on: Document, selector: string) =>
	[...on.querySelectorAll(selector)].map(
		(element) => element.getAttribute("value") ?? "",
	);

// texts are the texts of every element selector finds on a page.
const texts = (on: Document, selector: string) =>
	[...on.querySelectorAll(selector)].map((element) =>
		(element.textContent ?? "").replace(/\s+/g, " ").trim(),
	);

// research is what the page draws, read from the fixture.
const research = withPaper.research as Research;

describe("the page Research", () => {
	test("reads in full with no script", () => {
		for (const each of pages.values()) {
			expect(each.querySelector("script")).toBeNull();
		}
	});

	test("is headed by the paper's title up to its colon, in English as the paper gives it and in Russian as its Russian draft does", () => {
		const title = /\\title\{([^}]*)\}/.exec(tex)?.[1] ?? "";
		const russianTitle = /^# (.+)$/m.exec(russianDraft)?.[1] ?? "";
		const head = (whole: string) =>
			(
				whole
					.replace(/\\\\\s*/g, " ")
					.replace(/\s+/g, " ")
					.split(":")[0] ?? ""
			)
				.trim()
				.toLowerCase();
		const heading = (on: Document) =>
			(texts(on, "h1")[0] ?? "").replace(/\.$/, "").toLowerCase();

		expect(texts(page, "h1")).toHaveLength(1);
		expect(heading(page)).toBe(head(title));
		expect(heading(russianPage)).toBe(head(russianTitle));
	});

	test("says on its first screen that it is written for technical readers, in every language", () => {
		expect(texts(page, ".s-research-hero-copy .s-badge-tech")).toEqual([
			"For technical readers",
		]);
		expect(texts(russianPage, ".s-research-hero-copy .s-badge-tech")).toEqual([
			"Для технических специалистов",
		]);
	});

	test("counts its facts and its lead from the data", () => {
		const { product } = research;
		expect(values(page, ".s-research-facts data")).toEqual(
			[
				product.reference_tasks.total,
				product.checks,
				product.topics,
				product.traps,
				product.grades.first,
				product.grades.last,
			].map(String),
		);
		expect(values(page, ".s-research-lead data")).toEqual([
			String(product.checks),
		]);
	});

	test.each([
		[1, "прошедшие 1 проверку"],
		[2, "прошедшие 2 проверки"],
		[5, "прошедшие 5 проверок"],
		[21, "прошедшие 21 проверку"],
	])(
		"words the checks a task passes as Russian counts them, here %s",
		(checks, said) => {
			const moved = pagesOf(
				dataOf({ ...fixture, product: { ...fixture.product, checks } }),
			).get("ru") as Document;

			expect(texts(moved, ".s-research-lead").join(" ")).toContain(said);
		},
	);

	test("draws a task's way with a cell for every check, and the worked task's options", () => {
		const { product } = research;
		expect(page.querySelectorAll(".s-research-flow-checks li")).toHaveLength(
			product.checks,
		);
		expect(values(page, ".s-research-flow-count data")).toEqual([
			String(product.checks),
		]);
		expect(texts(page, ".s-research-flow-options [data-given]")).toEqual(
			example.map(String),
		);
	});

	test("works the first screen's fence out to the right option of the worked task", () => {
		expect(fencePosts).toBe(example[answer]);
		expect(texts(page, ".s-research-flow-code")).toEqual([
			`${fence.length} ÷ ${fence.gap} + 1 = ${fencePosts} → ${letters[answer]}`,
		]);
	});

	test("offers the paper's PDF with its pages when the data has one, the English file in every language", () => {
		const [paper] = research.paper.files;
		for (const each of pages.values()) {
			const buttons = [...each.querySelectorAll('a[type="application/pdf"]')];
			expect(buttons.map((button) => button.getAttribute("href"))).toEqual([
				paper?.path,
				paper?.path,
			]);
			expect(buttons.map((button) => button.getAttribute("hreflang"))).toEqual([
				"en",
				"en",
			]);
			expect(values(each, 'a[type="application/pdf"] data')).toEqual([
				String(paper?.pages),
			]);
		}
	});

	test("offers no PDF while the site ships none, and says when it comes", () => {
		for (const each of unshipped.values()) {
			expect(each.querySelector('a[type="application/pdf"]')).toBeNull();
		}
		expect(
			texts(
				unshipped.get("en") as Document,
				"#paper .s-research-paper-words p",
			),
		).toEqual([
			"Its methods, tables, protocols, limitations and ethics. The PDF comes once the paper names its authors and its archive.",
		]);
	});

	test("leads to the code from the first screen and from the paper's part", () => {
		const code = [...page.querySelectorAll("main a")].filter(
			(link) =>
				link.getAttribute("href") ===
				"https://github.com/MathTrail/mathtrail-standalone",
		);
		expect(code).toHaveLength(2);
	});

	test("draws the solver's second run from the data's relabelling", () => {
		const { relabelled } = research.product;
		expect(secondRun([3, 4, 5, 6, 12], ["C", "D", "E", "A", "B"])).toEqual([
			6, 12, 3, 4, 5,
		]);
		const runs = [...page.querySelectorAll(".s-research-runs .s-research-run")];
		const second = runs[1];
		expect(
			[
				...(second?.querySelectorAll(".s-research-cell [data-given]") ?? []),
			].map((cell) => cell.textContent),
		).toEqual(secondRun(example, relabelled).map(String));
		expect(
			second?.querySelector(".s-research-right .s-research-letter")
				?.textContent,
		).toBe(relabelled[answer]);
		expect(second?.querySelector(".s-research-picks")?.textContent).toBe(
			`→ ${relabelled[answer]}`,
		);
	});

	test("says the shift, the corridor, the guess and the trial series through the data, in every language", () => {
		const { product } = research;
		for (const each of pages.values()) {
			const theses = [...each.querySelectorAll(".s-research-thesis")];
			expect(values(theses[1] as unknown as Document, "p data")).toContain(
				String(product.relabel),
			);
			expect(values(theses[3] as unknown as Document, "data")).toEqual(
				expect.arrayContaining(
					[
						product.corridor.middle,
						product.corridor.low,
						product.corridor.high,
						product.trial_answers,
						product.guess,
						product.options,
					].map(String),
				),
			);
		}
	});

	test("pins its drawings left to right in every language", () => {
		for (const selector of [
			".s-research-flow-options",
			".s-research-runs",
			".s-research-scale",
			".s-research-formula",
			".s-research-bar",
			".s-research-plot-frame",
			".s-research-margins",
			".s-research-share",
		]) {
			const drawn = [...russianPage.querySelectorAll(selector)];
			expect(drawn.length, selector).toBeGreaterThan(0);
			for (const element of drawn) {
				expect(element.getAttribute("dir")).toBe("ltr");
			}
		}
	});

	test("shows every row of the goals in the data's order, the service's number with its interval and the goal's mark", () => {
		const rows = [...page.querySelectorAll(".s-research-measure")];
		expect(rows).toHaveLength(research.bench.rows.length);
		const order = ["step", "mastery", "screen", null];
		const sorted = [...research.bench.rows].sort(
			(a, b) => order.indexOf(a.criterion) - order.indexOf(b.criterion),
		);
		sorted.forEach((row, at) => {
			const shown = rows[at];
			expect(
				values(shown as unknown as Document, ".s-research-measure-value data"),
			).toEqual(
				[
					row.values.service.value,
					row.values.service.low,
					row.values.service.high,
				].map(String),
			);
			expect(
				[...(shown?.querySelectorAll(".s-research-mark") ?? [])].map(
					(pill) => pill.className,
				),
			).toEqual(
				row.values.service.mark === null
					? []
					: [`s-research-mark s-research-${row.values.service.mark}`],
			);
		});
	});

	test("shows no number of a rule before the service", () => {
		for (const each of pages.values()) {
			expect(
				each.querySelectorAll(
					`#${researchSections.model} th, #${researchSections.model} table`,
				),
			).toHaveLength(0);
		}
	});

	test.each([
		["a share", { unit: "share", high: 0.4, bound: 0.25 }, 1],
		["a size under one", { unit: "logit", high: 0.556, bound: 0.24 }, 1],
		["answers", { unit: "answers", high: 4.79, bound: 7.37 }, 10],
		["points", { unit: "points", high: 41, bound: 73 }, 100],
		["changes", { unit: "per_100_answers", high: 1.9, bound: 3.81 }, 5],
		["a size under a half", { unit: "logit", high: 0.49, bound: 0 }, 1],
		["a scale it reaches", { unit: "answers", high: 2, bound: 1 }, 2],
		["nothing at all", { unit: "logit", high: 0, bound: 0 }, 1],
	] as const)(
		"scales the bar of %s to a round number that holds it",
		(_, { unit, high, bound }, scale) => {
			const row = {
				...research.bench.rows[0],
				unit,
				bound: { value: bound, own: false },
				values: {
					service: { value: high / 2, low: 0, high, mark: "reached" },
					ceiling: null,
				},
			} as Row;
			expect(barScale(row)).toBe(scale);
		},
	);

	test.each([
		[
			"a goal alone, short of the middle",
			0.24,
			undefined,
			{ goal: ["end", false] },
		],
		[
			"a goal alone, past the middle",
			0.69,
			undefined,
			{ goal: ["start", false] },
		],
		[
			"a ceiling alone, near the end",
			undefined,
			0.9,
			{ ceiling: ["start", false] },
		],
		[
			"a goal and a ceiling after it, each with room",
			0.439,
			0.611,
			{ goal: ["start", false], ceiling: ["end", false] },
		],
		[
			"a goal and a ceiling before it, each with room",
			0.6,
			0.4,
			{ goal: ["end", false], ceiling: ["start", false] },
		],
		[
			"a goal and a ceiling near the bar's end",
			0.6,
			0.8,
			{ goal: ["start", false], ceiling: ["start", true] },
		],
		[
			"a ceiling near nothing before its goal",
			0.25,
			0.05,
			{ goal: ["end", false], ceiling: ["end", true] },
		],
		[
			"a goal near nothing before its ceiling",
			0.03,
			0.5,
			{ goal: ["end", false], ceiling: ["end", true] },
		],
	] as const)(
		"places the labels of %s where neither meets the other or leaves the bar",
		(_, goal, ceiling, want) => {
			const placed = linePlaces(goal, ceiling);
			const as = (place: Place | undefined) =>
				place === undefined ? undefined : [place.side, place.low];
			expect({ goal: as(placed.goal), ceiling: as(placed.ceiling) }).toEqual({
				goal: undefined,
				ceiling: undefined,
				...want,
			});
		},
	);

	test("puts the ceiling's label of a bar on a second row where it would meet the goal's", () => {
		const crowded = pagesOf(
			dataOf({
				...fixture,
				bench: {
					...fixture.bench,
					rows: fixture.bench.rows.map((row) =>
						row.id === "corridor_learning"
							? {
									...row,
									values: {
										...row.values,
										ceiling: {
											of: "oracle",
											value: 0.9,
											low: 0.88,
											high: 0.92,
										},
									},
								}
							: row,
					),
				},
			}),
		).get("en") as Document;
		const bars = [...crowded.querySelectorAll(".s-research-bar-two")];
		expect(bars).toHaveLength(1);
		expect(
			bars[0]?.querySelector(".s-research-line-label.s-research-line-ceiling")
				?.className,
		).toContain("s-research-line-low");
		expect(page.querySelectorAll(".s-research-bar-two")).toHaveLength(0);
	});

	test("draws each goal's line where its bound lies on the bar's scale, labelled with the bound", () => {
		const rows = [...page.querySelectorAll(".s-research-measure")];
		const order = ["step", "mastery", "screen", null];
		const sorted = [...research.bench.rows].sort(
			(a, b) => order.indexOf(a.criterion) - order.indexOf(b.criterion),
		);
		sorted.forEach((row, at) => {
			const line = rows[at]?.querySelector(
				".s-research-line.s-research-line-goal",
			);
			if (row.bound === null) {
				expect(line ?? null).toBeNull();
				return;
			}
			const style = (line?.getAttribute("style") ?? "").replace(/\s/g, "");
			const place = Number(/inset-inline-start:([\d.]+)%/.exec(style)?.[1]);
			expect(place).toBeCloseTo((row.bound.value / barScale(row)) * 100);
			expect(
				values(
					rows[at]?.querySelector(
						".s-research-line-label.s-research-line-goal",
					) as unknown as Document,
					"data",
				),
			).toEqual([String(row.bound.value)]);
		});
	});

	test("says on the lag's row the best any rule could do, the oracle's", () => {
		const lag = research.bench.rows.find((row) => row.id === "lag");
		const unit = page.querySelector(
			".s-research-measure .s-research-measure-unit",
		);
		expect(values(unit as unknown as Document, "data")).toEqual([
			String(lag?.values.ceiling?.value),
		]);
	});

	test("says which answers the screen's measures are read over, and which their goals are drawn from", () => {
		const { early, late } = research.bench.screen_windows;
		expect(values(page, ".s-research-group-note data")).toEqual(
			[late.first, late.last, early.first, early.last].map(String),
		);
	});

	test.each([
		["the fixture's", () => page, ["Reached", "On the edge", "Baseline"]],
		[
			"one not reached",
			() => notReached,
			["Reached", "On the edge", "Not reached", "Baseline"],
		],
	])(
		"names in its legend the marks %s rows hold, and no other",
		(_, on, named) => {
			expect(texts(on(), ".s-research-marks dt")).toEqual(named);
		},
	);

	test("names the ceiling in its key while a bar draws one, and not once none does", () => {
		const key = (on: Document) =>
			on.querySelectorAll(".s-research-key .s-research-swatch-ceiling").length;
		const none = pagesOf(
			dataOf({
				...fixture,
				bench: {
					...fixture.bench,
					rows: fixture.bench.rows.map((row) =>
						row.values.ceiling?.of === "oracle"
							? {
									...row,
									values: {
										...row.values,
										ceiling: { of: "oracle", value: 0, low: 0, high: 0 },
									},
								}
							: row,
					),
				},
			}),
		).get("en") as Document;
		expect(key(page)).toBe(1);
		expect(key(none)).toBe(0);
		expect(none.querySelectorAll(".s-research-line-ceiling")).toHaveLength(0);
	});

	test("draws a board whose every goal is better lower, with no word for a goal better higher", () => {
		const lower = pagesOf(
			dataOf({
				...fixture,
				bench: {
					...fixture.bench,
					rows: fixture.bench.rows.map((row) =>
						row.kind === "goal" && row.better === "higher"
							? { ...row, better: "lower" }
							: row,
					),
				},
			}),
		).get("en") as Document;
		expect(
			texts(lower, ".s-research-line-label.s-research-line-goal").every(
				(said) => said.startsWith("goal ≤"),
			),
		).toBe(true);
	});

	test("hides the drawing of its bars from a screen reader, which reads their labels and the numbers", () => {
		const model = page.querySelector(`#${researchSections.model}`);
		const tracks = [...(model?.querySelectorAll(".s-research-track") ?? [])];
		expect(tracks).toHaveLength(research.bench.rows.length);
		for (const part of [
			...tracks,
			...(model?.querySelectorAll(".s-research-line") ?? []),
		]) {
			expect(part.getAttribute("aria-hidden")).toBe("true");
		}
		for (const label of model?.querySelectorAll(".s-research-line-label") ??
			[]) {
			expect(label.getAttribute("aria-hidden")).toBeNull();
		}
	});

	test("draws a row whose every number is nothing as an empty bar, scaled to one rather than to nothing", () => {
		const zero = { value: 0, low: 0, high: 0 };
		const rows = fixture.bench.rows.map((row) =>
			row.kind === "context"
				? {
						...row,
						values: {
							service: { ...zero, mark: null },
							ceiling: { of: "oracle", ...zero },
						},
					}
				: row,
		);
		const empty = pagesOf(
			dataOf({ ...fixture, bench: { ...fixture.bench, rows } }),
		).get("en") as Document;

		const bars = [
			...empty.querySelectorAll(`#${researchSections.model} .s-research-track`),
		];
		const styles = bars.flatMap((bar) =>
			[...bar.querySelectorAll("[style]")].map(
				(part) => part.getAttribute("style") ?? "",
			),
		);
		expect(styles.filter((style) => style.includes("NaN"))).toEqual([]);
		const drawnEmpty = bars.filter((bar) =>
			[...bar.querySelectorAll("[style]")].every((part) =>
				/^[\w-]+:0%(?:;[\w-]+:0%)*;?$/.test(
					(part.getAttribute("style") ?? "").replace(/\s/g, ""),
				),
			),
		);
		expect(drawnEmpty).toHaveLength(
			rows.filter((row) => row.kind === "context").length,
		);
	});

	test("is not drawn from data that holds no numbers of it", () => {
		expect(() => pagesOf(dataOf(undefined))).toThrow(
			"the build was given no numbers for the page Research",
		);
	});

	test("tells the commit, the date and the run its numbers come from, and the commit of the paper's", () => {
		expect(texts(page, ".s-research-run-line code")).toEqual([
			research.built_from.commit.slice(0, 12),
		]);
		expect(texts(page, ".s-research-paper-note code")).toEqual([
			research.paper.commit.slice(0, 12),
		]);
		expect(
			page.querySelector(".s-research-run-line time")?.getAttribute("datetime"),
		).toBe(research.built_from.date);
		expect(values(page, ".s-research-run-line data")).toEqual(
			[
				research.bench.children,
				research.bench.answers,
				research.bench.seed,
			].map(String),
		);
	});

	test("lists the authors in the data's order, each with the year they died", () => {
		expect(values(page, ".s-research-died data")).toEqual(
			file.research.sources.map(({ died }) => String(died)),
		);
		expect(texts(page, ".s-research-died")).toContain("† 1942");
	});

	test("draws the reference tasks by level as a bar of their counts", () => {
		const parts = [...page.querySelectorAll(".s-research-levels li")];
		expect(
			parts.map((part) => part.getAttribute("style")?.replace(/[\s;]/g, "")),
		).toEqual(research.levels.map(({ count }) => `flex-grow:${count}`));
	});

	test("says in its live block that the numbers are still to come, names its three measures, gives the rule they come by, and shows none", () => {
		const live = (coming.get("en") as Document).querySelector(
			`#${researchSections.live}`,
		);
		expect(live?.querySelector(".s-research-soon")?.textContent).toBe("Coming");
		expect(texts(live as unknown as Document, "h3")).toEqual([
			"Promised → came true",
			"Came true − promised, by the number of answers",
			"The share right on the rule's tasks",
		]);
		expect(values(live as unknown as Document, "data")).toEqual(
			[
				research.product.corridor.middle,
				research.live.rule.learners,
				research.live.rule.answers,
				research.live.rule.rounded_to,
			].map(String),
		);
		expect(
			live?.querySelectorAll(
				"time, table, .s-research-plot-point, .s-research-margin-fill",
			).length,
		).toBe(0);
		expect(
			live?.querySelectorAll(
				".s-research-plot, .s-research-margins, .s-research-share",
			),
		).toHaveLength(3);
	});

	test.each([
		["en", "Too few yet", "November 2026"],
		["ru", "Пока мало", "ноябрь 2026 г."],
	])(
		"says in %s that the latest month had too few children, and names the month",
		(locale, badge, month) => {
			const live = (tooFew.get(locale) as Document).querySelector(
				`#${researchSections.live}`,
			);
			expect(live?.querySelector(".s-research-soon")?.textContent).toBe(badge);
			const time = live?.querySelector("time");
			expect(time?.getAttribute("datetime")).toBe(research.live.month);
			expect(time?.textContent).toBe(month);
			expect(
				live?.querySelectorAll("table, .s-research-plot-point").length,
			).toBe(0);
		},
	);

	test("lists every range of a month shown in the data's order, each with what it promised and what came true", () => {
		const live = research.live;
		if (live.state !== "ready") {
			throw new Error("the fixture's live numbers are of no month shown");
		}
		const rowsOf = (title: string) =>
			[
				...page.querySelectorAll(`table[aria-labelledby="${title}"] tbody tr`),
			].map((row) => values(row as unknown as Document, "data"));
		expect(rowsOf("live-chances")).toEqual(
			live.chances.map((cell) =>
				[
					cell.from,
					cell.to,
					cell.promised_mean,
					cell.correct_share,
					cell.answers,
					cell.learners,
				].map(String),
			),
		);
		expect(rowsOf("live-kept-up")).toEqual(
			live.kept_up.map((cell) =>
				[
					cell.first,
					cell.last,
					cell.came_true_less_promised,
					cell.standard_error,
					cell.answers,
					cell.learners,
				].map(String),
			),
		);
		expect(rowsOf("live-share")).toEqual([
			[
				live.total.correct_share,
				live.total.promised_mean,
				live.total.answers,
				live.total.learners,
			].map(String),
		]);
	});

	test("draws each range of chance as a point, across at its promise and up at its share, on axes of one scale from four tenths", () => {
		const live = research.live;
		if (live.state !== "ready") {
			throw new Error("the fixture's live numbers are of no month shown");
		}
		const scaled = (x: number) => ((x - 0.4) / 0.6) * 100;
		const points = [...page.querySelectorAll(".s-research-plot-point")].map(
			(point) => {
				const style = (point.getAttribute("style") ?? "").replace(/\s/g, "");
				return [
					Number(/inset-inline-start:([\d.]+)%/.exec(style)?.[1]),
					Number(/inset-block-end:([\d.]+)%/.exec(style)?.[1]),
				];
			},
		);
		expect(points).toHaveLength(live.chances.length);
		live.chances.forEach((cell, at) => {
			expect(points[at]?.[0]).toBeCloseTo(scaled(cell.promised_mean));
			expect(points[at]?.[1]).toBeCloseTo(scaled(cell.correct_share));
		});
		expect(texts(page, ".s-research-plot .s-research-x-high")).toEqual(["1.0"]);
		expect(texts(page, ".s-research-plot .s-research-y")).toEqual([
			"0.4",
			"1.0",
		]);
	});

	test("joins the points of ranges that meet, and never across a range the month leaves out", () => {
		const live = research.live;
		if (live.state !== "ready") {
			throw new Error("the fixture's live numbers are of no month shown");
		}
		const [first, second, third, fourth] = live.chances;
		if (
			fourth === undefined ||
			third === undefined ||
			second === undefined ||
			first === undefined
		) {
			throw new Error(
				"the fixture's month shows fewer than four ranges of chance",
			);
		}
		expect(runsOf(live.chances)).toEqual([[first, second, third, fourth]]);
		expect(runsOf([first, third, fourth])).toEqual([[first], [third, fourth]]);

		const gap = pagesOf(
			dataOf({
				...fixture,
				live: { ...fixture.live, chances: [first, third, fourth] },
			}),
		).get("en") as Document;
		const lines = [...gap.querySelectorAll(".s-research-plot-path")].map(
			(line) => (line.getAttribute("points") ?? "").split(" ").length,
		);
		expect(lines).toEqual([2]);
	});

	test("counts the answers of each range of the child's answers with their noun", () => {
		const live = research.live;
		if (live.state !== "ready") {
			throw new Error("the fixture's live numbers are of no month shown");
		}
		expect(texts(page, ".s-research-margins-range > span:last-child")).toEqual(
			live.kept_up.map((cell) => `${cell.answers} answers`),
		);
		expect(
			texts(russianPage, ".s-research-margins-range > span:last-child").every(
				(said) => /^\d+ ответ(а|ов)?$/.test(said),
			),
		).toBe(true);
	});

	test("draws a margin below its promise to the left of nothing, and one above it to the right", () => {
		const fills = [
			...page.querySelectorAll(
				`#${researchSections.live} .s-research-margin-fill`,
			),
		].map((fill) => {
			const style = fill.getAttribute("style") ?? "";
			const start = Number(/inset-inline-start:\s*([\d.]+)%/.exec(style)?.[1]);
			const size = Number(/inline-size:\s*([\d.]+)%/.exec(style)?.[1]);
			return [start, start + size];
		});
		const live = research.live;
		if (live.state !== "ready") {
			throw new Error("the fixture's live numbers are of no month shown");
		}
		expect(fills).toHaveLength(live.kept_up.length);
		live.kept_up.forEach((cell, at) => {
			const [start, end] = fills[at] ?? [];
			if (cell.came_true_less_promised < 0) {
				expect(start).toBeLessThan(50);
				expect(end).toBeCloseTo(50);
			} else {
				expect(start).toBeCloseTo(50);
				expect(end).toBeGreaterThan(50);
			}
		});
	});

	test("draws the month's share right as a point over the corridor, with a dashed line at the corridor's middle", () => {
		const live = research.live;
		if (live.state !== "ready") {
			throw new Error("the fixture's live numbers are of no month shown");
		}
		const share = page.querySelector(
			`#${researchSections.live} .s-research-share`,
		);
		const styleOf = (selector: string) =>
			(share?.querySelector(selector)?.getAttribute("style") ?? "").replace(
				/\s/g,
				"",
			);
		expect(styleOf(".s-research-share-point")).toContain(
			`inset-inline-start:${live.total.correct_share * 100}%`,
		);
		expect(styleOf(".s-research-share-middle")).toContain(
			`inset-inline-start:${research.product.corridor.middle * 100}%`,
		);
	});

	test("says of a month shown with no range of either kind that none holds enough children, and draws the share alone", () => {
		const live = (noRanges.get("en") as Document).querySelector(
			`#${researchSections.live}`,
		);
		expect(texts(live as unknown as Document, ".s-research-caption")).toEqual(
			expect.arrayContaining([
				"This month no range of the chance promised holds enough children to be shown.",
				"This month no range of the child's answers holds enough children to be shown.",
			]),
		);
		expect(
			live?.querySelectorAll(".s-research-plot-point, .s-research-margin")
				.length,
		).toBe(0);
		expect(live?.querySelectorAll(".s-research-share").length).toBe(1);
	});

	test.each([
		["en", "201 and on"],
		["ru", "201 и дальше"],
	])(
		"names in %s the range of every answer past the others by its first",
		(locale, named) => {
			const rows = [
				...(openRange.get(locale) as Document).querySelectorAll(
					'table[aria-labelledby="live-kept-up"] tbody th',
				),
			];
			expect(rows.at(-1)?.textContent?.trim()).toBe(named);
		},
	);

	test("hides its live drawings from a screen reader, which reads the tables, and the tables from the eye", () => {
		const live = page.querySelector(`#${researchSections.live}`);
		const drawings = [
			...(live?.querySelectorAll(
				".s-research-plot-frame, .s-research-margins, .s-research-share",
			) ?? []),
		];
		expect(drawings).toHaveLength(3);
		for (const drawing of drawings) {
			expect(drawing.getAttribute("aria-hidden")).toBe("true");
		}
		const tables = [...(live?.querySelectorAll("table") ?? [])];
		expect(tables).toHaveLength(3);
		for (const table of tables) {
			expect(table.parentElement?.getAttribute("class")).toBe("s-hidden");
			expect(table.getAttribute("class")).toBeNull();
		}
	});

	test("has the anchors other pages lead to", () => {
		for (const id of Object.values(researchSections)) {
			expect(page.getElementById(id), id).not.toBeNull();
		}
	});
});

describe("the numbers on the page Research", () => {
	test("are typed into none of its words, in any language, but in the names that carry their own", () => {
		for (const words of [english, russian]) {
			const typed = words
				.split("\n")
				.filter((line) =>
					/\p{Nd}/u.test(
						names.reduce((left, name) => left.split(name).join(""), line),
					),
				);
			expect(typed).toEqual([]);
		}
	});

	test("show no digit that did not come from the data, in any language", () => {
		const commits = new Set([
			research.built_from.commit,
			research.paper.commit,
			research.built_from.date,
			research.live.month ?? "",
		]);
		for (const [locale, each] of [
			...pages,
			...unshipped,
			...coming,
			...tooFew,
			...noRanges,
			...openRange,
		]) {
			expect(handTyped(each, { names, values: commits }), locale).toEqual([]);
		}
	});

	test("are each a number of the data", () => {
		const known = new Set<string>();
		const collect = (value: unknown): void => {
			if (typeof value === "number") {
				known.add(String(value));
			} else if (typeof value === "object" && value !== null) {
				Object.values(value).forEach(collect);
			}
		};
		collect(fixture);
		collect(file.research);
		collect(research.levels);
		for (const each of pages.values()) {
			for (const value of values(each, "data")) {
				expect(known.has(value), value).toBe(true);
			}
		}
	});

	test("all move when the data moves", () => {
		const moved = (value: unknown): unknown => {
			if (typeof value === "number") {
				return value + 1;
			}
			if (Array.isArray(value)) {
				return value.map(moved);
			}
			if (typeof value === "object" && value !== null) {
				return Object.fromEntries(
					Object.entries(value).map(([key, inner]) => [key, moved(inner)]),
				);
			}
			return value;
		};
		// Every ceiling of the oracle's stands off nothing, so that moving the
		// numbers draws no ceiling the page did not draw before.
		const standing = {
			...research,
			bench: {
				...research.bench,
				rows: research.bench.rows.map((row) =>
					row.values.ceiling?.of === "oracle" && row.values.ceiling.value === 0
						? {
								...row,
								values: {
									...row.values,
									ceiling: { of: "oracle", value: 0.5, low: 0.5, high: 0.5 },
								},
							}
						: row,
				),
			},
		} as Research;
		const valuesOf = (drawn: Research) =>
			values(
				pagesOf({ ...withPaper, research: drawn }).get("en") as Document,
				"data",
			);
		const before = valuesOf(standing);
		const after = valuesOf(moved(standing) as Research);
		expect(after).toHaveLength(before.length);
		after.forEach((value, at) => {
			expect(value, `the number ${before[at]}`).not.toBe(before[at]);
		});
	});
});

describe("the check of numbers typed by hand", () => {
	// pageWith is a page whose main holds markup, under a head of its own.
	const pageWith = (
		main: string,
		head = '<title>T</title><meta name="description" content="D">',
	) =>
		new browser.DOMParser().parseFromString(
			`<html><head>${head}</head><body><main>${main}</main></body></html>`,
			"text/html",
		) as unknown as Document;
	const rules = {
		names,
		values: new Set(["0123456789abcdef", "2026-10-05T00:00:00Z"]),
	};

	test.each([
		["a text", "<p>Nine of 9 checks</p>"],
		["an aria-label", '<p aria-label="2 of them">two</p>'],
		["a title", '<abbr title="3 checks">three</abbr>'],
		["an alt", '<img alt="4 options">'],
		["a digit of another script", "<p>٣ ловушки</p>"],
		["a commit that is not the data's", "<code>fedcba9876</code>"],
		["a code too short to name a commit", "<code>01234</code>"],
		[
			"a date that is not the data's",
			'<time datetime="2020-01-01">2020</time>',
		],
	])("finds a digit typed into %s", (_, main) => {
		expect(handTyped(pageWith(main), rules)).not.toEqual([]);
	});

	test.each([
		["its title", "<title>Research 2026</title>"],
		["its description", '<meta name="description" content="17 topics">'],
		[
			"the title a shared link shows",
			'<meta property="og:title" content="Research 2026">',
		],
		[
			"the description a shared link shows",
			'<meta property="og:description" content="17 topics">',
		],
		[
			"the words a shared link gives its picture",
			'<meta property="og:image:alt" content="A card of 5 options">',
		],
	])("finds a digit in %s", (_, head) => {
		expect(handTyped(pageWith("", head), rules)).not.toEqual([]);
	});

	test("lets through a number written as data or marked as given, a name of rules, and a commit and a date of the data", () => {
		const main = [
			'<p><data value="9">9</data> checks, <span data-given="">01</span></p>',
			"<p>XChaCha20-Poly1305</p>",
			"<code>0123456789ab</code>",
			'<time datetime="2026-10-05T00:00:00Z">5 October 2026</time>',
		].join("");
		expect(handTyped(pageWith(main), rules)).toEqual([]);
	});
});

describe("the page Research under the site's stylesheets", () => {
	// screen is a window as wide as a desktop's, under the stylesheets every
	// page of the site loads.
	const screen = new Window({ width: 1280, height: 800 });
	for (const sheet of [
		["internal", "widget", "tokens.css"],
		["web", "src", "site", "style.css"],
	]) {
		const style = screen.document.createElement("style");
		style.textContent = textOf(...sheet);
		screen.document.head.append(style);
	}

	afterAll(async () => {
		await screen.happyDOM.close();
	});

	// drawn is what selector finds on a page, drawn alone on the screen, with
	// the style it is drawn in there.
	const drawn = (selector: string, on: Document = notReached) => {
		screen.document.body.innerHTML =
			on.querySelector(selector)?.outerHTML ?? "";
		const element = screen.document.body.firstElementChild;
		if (element === null) {
			throw new Error(`the page draws nothing ${selector} finds`);
		}
		return element;
	};
	const styleOf = (element: ReturnType<typeof drawn>) =>
		screen.getComputedStyle(element);

	// token is a colour of the site's palette, by its name.
	const token = (name: string) =>
		styleOf(screen.document.documentElement).getPropertyValue(name).trim();

	test.each([
		["reached", "--s-right", "--s-on-right"],
		["not_reached", "--s-trap", "--s-on-trap"],
		["on_the_edge", "--s-wip", "--s-on-wip"],
		["baseline", "--s-track", "--s-muted"],
	])(
		"draws the mark %s in the colours of its kind",
		(mark, background, text) => {
			const style = styleOf(drawn(`.s-research-marks .s-research-${mark}`));

			expect(style.backgroundColor).toBe(token(background));
			expect(style.color).toBe(token(text));
		},
	);
});
