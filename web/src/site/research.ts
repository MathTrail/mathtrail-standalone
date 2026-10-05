import { z } from "zod";
import { letters } from "../widget/choices";
import type { Catalog } from "./data";
import { gradesOf } from "./topics";

// The page "Research" shows numbers it does not compute: the student model's,
// as the learners' bench computes them from the commit being built, the
// product's counts and constants, the paper's facts, and the live numbers of
// real children in the latest month counted whole, all in one file made at
// build time. The file is read here whole, and every number the page draws
// is held to what it says of the others, so that a file of another shape, or
// one whose numbers do not add up, stops the build rather than drawing a page
// that says what the data does not.

// The shapes of the hashes and the date the file names.
const commit = z.string().regex(/^[0-9a-f]{40}(?:[0-9a-f]{24})?$/);
const paperCommit = z.string().regex(/^[0-9a-f]{7,64}$/);
const buildKey = z.string().regex(/^[0-9a-f]{64}$/);
const utcDate = z.string().regex(/^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}Z$/);
const count = z.number().int().positive();

// interval is a number with the ends of its interval.
const interval = { value: z.number(), low: z.number(), high: z.number() };

// The marks a rule's number of a goal is given: the bench's own, and the
// baseline's, which only the service carries, against a bound read off its
// own number.
const benchMark = z.enum(["reached", "on_the_edge", "not_reached"]);
const serviceMark = z.enum([
	"reached",
	"on_the_edge",
	"not_reached",
	"baseline",
]);

/** Mark is the mark a rule's number of a goal is given. */
export type Mark = z.infer<typeof serviceMark>;

/** marks are every mark the page names, in the order its legend gives them. */
export const marks: readonly Mark[] = serviceMark.options;

const ceiling = z
	.strictObject({ of: z.enum(["oracle", "perfect"]), ...interval })
	.nullable();

/** ceilingKinds are the ceilings a row is bounded by, in the legend's order. */
export const ceilingKinds = ["oracle", "perfect"] as const;

// What every row says of the measure it is.
const measure = {
	id: z.string().regex(/^[a-z][a-z0-9_]*$/),
	generator: z.string(),
	metric: z.string(),
	unit: z.enum(["logit", "share", "answers", "points", "per_100_answers"]),
	read_as: z.enum(["value", "size"]),
	better: z.enum(["lower", "higher"]),
};

const goalRow = z.strictObject({
	...measure,
	kind: z.literal("goal"),
	criterion: z.enum(["step", "mastery", "screen"]),
	bound: z.strictObject({ value: z.number(), own: z.boolean() }),
	values: z.strictObject({
		service: z.strictObject({ ...interval, mark: serviceMark }),
		earlier: z.strictObject({ ...interval, mark: benchMark }),
		ceiling,
	}),
});

const contextRow = z.strictObject({
	...measure,
	kind: z.literal("context"),
	criterion: z.null(),
	bound: z.null(),
	values: z.strictObject({
		service: z.strictObject({ ...interval, mark: z.null() }),
		earlier: z.strictObject({ ...interval, mark: z.null() }),
		ceiling,
	}),
});

const span = z.strictObject({ first: count, last: count });

const paperFile = z.strictObject({
	lang: z.string().regex(/^[a-z]{2,3}(?:-[A-Za-z0-9]{2,8})*$/),
	path: z.string().regex(/^\/assets\/[a-z0-9.-]+\.pdf$/),
	pages: count,
	bytes: count,
	sha256: z.string().regex(/^[0-9a-f]{64}$/),
});

// The live numbers: a month, as a year and a month; a count of a cell; and a
// share or a chance, from nothing to all.
const month = z.string().regex(/^\d{4}-(?:0[1-9]|1[0-2])$/);
const shown = z.number().int().nonnegative();
const share = z.number().min(0).max(1);

// What a cell of the live numbers stands on, and how its answers came out
// against the chance promised.
const counts = { learners: shown, answers: shown };
const weighed = { promised_mean: share, correct_share: share };

const liveRule = z.strictObject({
	learners: count,
	answers: count,
	rounded_to: count,
});

const chanceCell = z.strictObject({
	from: share,
	to: share,
	...counts,
	...weighed,
});

const answersCell = z.strictObject({
	first: count,
	last: count.nullable(),
	...counts,
	came_true_less_promised: z.number().min(-1).max(1),
	standard_error: z.number().nonnegative(),
});

// nothing is what live numbers that show none have of them.
const nothing = {
	total: z.null(),
	chances: z.array(chanceCell).max(0),
	kept_up: z.array(answersCell).max(0),
};

// live is the live numbers in each of their states: still to come, before a
// month is counted whole; too few, when its children were too few for any of
// its numbers to be shown; and shown.
const live = z.discriminatedUnion("state", [
	z.strictObject({
		state: z.literal("coming"),
		rule: liveRule,
		month: z.null(),
		...nothing,
	}),
	z.strictObject({
		state: z.literal("too_few"),
		rule: liveRule,
		month,
		...nothing,
	}),
	z.strictObject({
		state: z.literal("ready"),
		rule: liveRule,
		month,
		total: z.strictObject({ ...counts, ...weighed }),
		chances: z.array(chanceCell),
		kept_up: z.array(answersCell),
	}),
]);

// researchFile is the shape of the file, which the bench writes: a key it does
// not write, or one it leaves out, is refused, as the bench refuses on its side
// a file of a shape it does not know.
const researchFile = z.strictObject({
	schema: z.literal(1),
	built_from: z.strictObject({ commit, date: utcDate }),
	bench: z.strictObject({
		producer: z.string(),
		inputs: buildKey,
		seed: z.number().int().nonnegative(),
		experiment: z.string().min(1),
		children: count,
		answers: count,
		interval: z.number().gt(0).lt(1),
		resamples: count,
		error_after: count,
		screen_windows: z.strictObject({ early: span, late: span }),
		goal_parameters: z.strictObject({
			lag_share: z.number(),
			corridor_share: z.number(),
			unsettled_most: z.number(),
			false_most: z.number(),
			late_times: z.number(),
		}),
		rules: z.array(
			z.strictObject({
				id: z.string(),
				role: z.enum(["service", "earlier", "ceiling"]),
			}),
		),
		rows: z.array(z.discriminatedUnion("kind", [goalRow, contextRow])).min(1),
	}),
	product: z.strictObject({
		topics: count,
		traps: count,
		checks: count,
		grades: span,
		reference_tasks: z.strictObject({
			total: count,
			by_level: z.record(z.string().regex(/^\d+-\d+$/), count),
		}),
		options: count,
		relabel: count,
		relabelled: z.array(z.string().regex(/^[A-Z]$/)),
		guess: z.number(),
		corridor: z.strictObject({
			low: z.number(),
			high: z.number(),
			middle: z.number(),
		}),
		trial_answers: count,
	}),
	paper: z.strictObject({ commit: paperCommit, files: z.array(paperFile) }),
	live,
});

/** ResearchFile is the file as the bench writes it. */
export type ResearchFile = z.infer<typeof researchFile>;

/** Row is a row of the page's table of goals. */
export type Row = ResearchFile["bench"]["rows"][number];

/** PaperFile is a file of the paper the site ships. */
export type PaperFile = z.infer<typeof paperFile>;

/** Live is the live numbers, in one of their states. */
export type Live = ResearchFile["live"];

/** LiveShown is the live numbers of a month whose numbers are shown. */
export type LiveShown = Extract<Live, { state: "ready" }>;

/** ChanceCell is a range of the chance promised, with how it came out. */
export type ChanceCell = LiveShown["chances"][number];

/** AnswersCell is a range of the child's answers, with how it came out. */
export type AnswersCell = LiveShown["kept_up"][number];

/**
 * researchSources is the shape of the part of the site's data the page
 * "Research" takes from it rather than from the bench: the authors whose
 * problem books the reference tasks take their ideas from, in the order the
 * page names them, each with the year the author died, from which most of the
 * world counts the term of the books' copyright, and the books by an id of the
 * site's own. The
 * names and the titles are words of the page, under the same ids.
 */
export const researchSources = z.object({
	sources: z
		.array(
			z.object({
				id: z.string().regex(/^[a-z]+$/),
				died: z.number().int(),
				books: z.array(z.string().regex(/^[a-z]+$/)).min(1),
			}),
		)
		.min(1),
});

/** ResearchSources is what the site's data gives the page "Research". */
export type ResearchSources = z.infer<typeof researchSources>;

/** Author is an author of the problem books the reference tasks come from. */
export type Author = ResearchSources["sources"][number];

/**
 * Level is a level of the reference tasks as the page shows it: the grades it
 * spans and how many reference tasks are set at it.
 */
export type Level = {
	readonly key: string;
	readonly first: number;
	readonly last: number;
	readonly count: number;
};

/**
 * Research is what the page "Research" draws: the file as the bench writes
 * it, the levels of the reference tasks from the lowest, and the authors of
 * the problem books, from the site's own data.
 */
export type Research = ResearchFile & {
	readonly levels: readonly Level[];
	readonly authors: readonly Author[];
};

/**
 * readResearch reads the data of the page "Research" beside the catalog the
 * site reads and the authors the site's data names. A file of another shape
 * is refused, and so is one whose marks do not follow from its numbers, whose
 * counts disagree with the catalog, whose constants contradict what the page
 * says of them (so many options, the letters shifted so far, a guess of one
 * option in all of them, a corridor around its middle), or whose live numbers
 * show a cell their own rule hides.
 */
export function readResearch(
	catalog: Catalog,
	sources: ResearchSources | undefined,
	data: unknown,
): Research {
	const read = researchFile.safeParse(data);
	if (!read.success) {
		throw new Error(`research.json: ${z.prettifyError(read.error)}`);
	}
	const file = read.data;
	try {
		checkRules(file.bench.rules);
		checkRows(file.bench.rows);
		checkCounts(catalog, file.product);
		checkConstants(file.product);
		checkPaper(file.paper.files);
		checkLive(file.live);
	} catch (error) {
		const said = error instanceof Error ? error.message : "it cannot be read";
		throw new Error(`research.json: ${said}`, { cause: error });
	}
	if (sources === undefined) {
		throw new Error(
			"site/data.json names no authors for the page Research, which its data needs",
		);
	}
	return {
		...file,
		levels: levelsOf(file.product.reference_tasks.by_level),
		authors: sources.sources,
	};
}

/**
 * markFrom is the mark a number of a goal is given, by the bench's own rule:
 * reached when its whole interval lies on the better side of the bound, the
 * bound itself included, not reached when it lies wholly on the other side,
 * and on the edge when it holds the bound.
 */
export function markFrom(
	better: "lower" | "higher",
	bound: number,
	{ low, high }: { readonly low: number; readonly high: number },
): z.infer<typeof benchMark> {
	if (better === "lower") {
		if (high <= bound) {
			return "reached";
		}
		return low > bound ? "not_reached" : "on_the_edge";
	}
	if (low >= bound) {
		return "reached";
	}
	return high < bound ? "not_reached" : "on_the_edge";
}

// checkRules refuses rules that do not give each part once: the service, the
// rule before it, and the ceiling.
function checkRules(rules: ResearchFile["bench"]["rules"]): void {
	const roles = rules.map(({ role }) => role);
	if (roles.length !== 3 || new Set(roles).size !== 3) {
		throw new Error(
			`the rules play ${roles.join(", ")}, and want the service, the rule before it and the ceiling, each once`,
		);
	}
}

// checkRows refuses a row named twice, a row whose numbers its drawing cannot
// show, and a mark that does not follow from its number: the service's is the
// baseline exactly where the bound is read off its own number.
function checkRows(rows: readonly Row[]): void {
	const seen = new Set<string>();
	for (const row of rows) {
		if (seen.has(row.id)) {
			throw new Error(`the row ${row.id} is there twice`);
		}
		seen.add(row.id);
		checkNumbers(row);
		if (row.kind === "goal") {
			checkMarks(row);
		}
	}
}

// checkNumbers refuses an interval of a row whose ends are the wrong way
// round, and a number of the row under zero, where its drawing starts.
function checkNumbers(row: Row): void {
	for (const ends of [
		row.values.service,
		row.values.earlier,
		row.values.ceiling,
	]) {
		if (ends === null) {
			continue;
		}
		if (ends.low > ends.high) {
			throw new Error(
				`the row ${row.id} has an interval from ${ends.low} down to ${ends.high}`,
			);
		}
		if (Math.min(ends.value, ends.low) < 0) {
			throw new Error(
				`the row ${row.id} has a number under zero, where its drawing starts`,
			);
		}
	}
	if (row.bound !== null && row.bound.value < 0) {
		throw new Error(
			`the row ${row.id} has a goal under zero, where its drawing starts`,
		);
	}
}

// checkMarks refuses marks of a goal's row that its numbers do not give.
function checkMarks(row: Extract<Row, { kind: "goal" }>): void {
	const { service, earlier } = row.values;
	const serviceWant = row.bound.own
		? "baseline"
		: markFrom(row.better, row.bound.value, service);
	if (service.mark !== serviceWant) {
		throw new Error(
			`the service is marked ${service.mark} on ${row.id}, and its numbers give ${serviceWant}`,
		);
	}
	const earlierWant = markFrom(row.better, row.bound.value, earlier);
	if (earlier.mark !== earlierWant) {
		throw new Error(
			`the rule before is marked ${earlier.mark} on ${row.id}, and its numbers give ${earlierWant}`,
		);
	}
}

// checkCounts refuses counts of the product that disagree with the catalog
// the site reads: the page and the rest of the site would say two things.
function checkCounts(catalog: Catalog, product: ResearchFile["product"]): void {
	const byLevel = new Map<string, number>();
	for (const task of catalog.tasks) {
		byLevel.set(task.grade_level, (byLevel.get(task.grade_level) ?? 0) + 1);
	}
	const [first, last] = gradesOf(
		catalog.topics.flatMap((topic) => topic.grade_levels),
	);
	const disagree = (
		what: string,
		says: number,
		catalogSays: number | undefined,
	) => {
		if (says !== catalogSays) {
			throw new Error(
				`it counts ${says} ${what}, and the catalog ${catalogSays}`,
			);
		}
	};
	disagree("topics", product.topics, catalog.topics.length);
	disagree("traps", product.traps, catalog.traps.length);
	disagree(
		"reference tasks",
		product.reference_tasks.total,
		catalog.tasks.length,
	);
	const levels = Object.entries(product.reference_tasks.by_level);
	disagree("levels", levels.length, byLevel.size);
	for (const [level, tasks] of levels) {
		disagree(`reference tasks at ${level}`, tasks, byLevel.get(level));
	}
	disagree("as the first grade", product.grades.first, first);
	disagree("as the last grade", product.grades.last, last);
}

// checkConstants refuses constants that contradict what the page says of
// them: the options are the card's letters, the second run of the solver puts
// each letter so many places on, the guess is one option in all of them, and
// the corridor lies around its middle, within the chances there are.
function checkConstants(product: ResearchFile["product"]): void {
	const { options, relabel, relabelled, guess, corridor } = product;
	if (options !== letters.length) {
		throw new Error(
			`it offers ${options} options, and a card shows ${letters.length}`,
		);
	}
	if (relabel < 1 || relabel >= options) {
		throw new Error(
			`it moves the letters ${relabel} places on, and among ${options} options a shift that moves every letter is 1 to ${options - 1}`,
		);
	}
	const shifted = letters.map((_, at) => letters[(at + relabel) % options]);
	if (relabelled.join() !== shifted.join()) {
		throw new Error(
			`it relabels the letters as ${relabelled.join("")}, which is no shift by ${relabel}: that is ${shifted.join("")}`,
		);
	}
	if (guess !== 1 / options) {
		throw new Error(
			`its guess is ${guess}, and one option in ${options} is ${1 / options}`,
		);
	}
	const { low, middle, high } = corridor;
	if (!(0 < low && low < middle && middle < high && high < 1)) {
		throw new Error(
			`its corridor runs from ${low} to ${high} around ${middle}`,
		);
	}
}

// checkPaper refuses files of the paper the page could not offer: two at one
// address, or none in English, which every language links until there is one
// in its own.
function checkPaper(files: readonly PaperFile[]): void {
	if (new Set(files.map(({ path }) => path)).size !== files.length) {
		throw new Error("two files of the paper are served at one address");
	}
	if (files.length > 0 && !files.some(({ lang }) => lang === "en")) {
		throw new Error(
			"the paper's files hold none in English, which every language links",
		);
	}
}

/**
 * privacyFloor is the weakest rule the page shows live numbers by: the rule of
 * the public views they are read from, as their SQL writes it. A file that
 * carries a weaker one is refused, so that no file can lower what the page
 * shows below what the views would.
 */
export const privacyFloor = {
	learners: 10,
	answers: 30,
	rounded_to: 5,
} as const;

// checkLive refuses live numbers whose rule is weaker than the views', that
// show what their rule hides, or whose ranges cannot be so: out of order or
// overlapping, or a promise outside its range of chance.
function checkLive(live: Live): void {
	const { rule } = live;
	if (
		rule.learners < privacyFloor.learners ||
		rule.answers < privacyFloor.answers ||
		rule.rounded_to % privacyFloor.rounded_to !== 0
	) {
		throw new Error(
			`the live numbers are shown on ${rule.learners} children and ${rule.answers} answers, counted to ${rule.rounded_to}, and the public views show none under ${privacyFloor.learners} and ${privacyFloor.answers}, counted to ${privacyFloor.rounded_to}`,
		);
	}
	if (live.state !== "ready") {
		return;
	}
	const shownByTheRule = shownBy(live);
	shownByTheRule("the month", live.total);
	live.chances.forEach((cell, at) => {
		const name = `the range of chance ${cell.from}–${cell.to}`;
		shownByTheRule(name, cell);
		const before = live.chances[at - 1];
		if (
			cell.to < cell.from ||
			(before !== undefined && cell.from <= before.to)
		) {
			throw new Error(`${name} is out of order or overlaps the one before it`);
		}
		if (cell.promised_mean < cell.from || cell.promised_mean > cell.to) {
			throw new Error(
				`${name} promises ${cell.promised_mean} on average, outside the chances it holds`,
			);
		}
	});
	live.kept_up.forEach((cell, at) => {
		const name = `the range of answers from ${cell.first}`;
		shownByTheRule(name, cell);
		const before = live.kept_up[at - 1];
		if (
			(cell.last !== null && cell.last < cell.first) ||
			(before !== undefined &&
				(before.last === null || cell.first <= before.last))
		) {
			throw new Error(`${name} is out of order or overlaps the one before it`);
		}
	});
}

// shownBy is the check that a cell of live numbers is one their rule shows: on
// the rule's children and answers or more, rounded to its multiple, and
// counting no more than its month.
function shownBy({
	rule,
	total,
}: LiveShown): (
	name: string,
	cell: { readonly learners: number; readonly answers: number },
) => void {
	return (name, { learners, answers }) => {
		if (learners < rule.learners || answers < rule.answers) {
			throw new Error(
				`${name} stands on ${learners} children and ${answers} answers, and the rule shows nothing under ${rule.learners} and ${rule.answers}`,
			);
		}
		if (learners % rule.rounded_to !== 0 || answers % rule.rounded_to !== 0) {
			throw new Error(
				`${name} counts ${learners} children and ${answers} answers, which are not rounded to ${rule.rounded_to}`,
			);
		}
		if (learners > total.learners || answers > total.answers) {
			throw new Error(
				`${name} counts ${learners} children and ${answers} answers, more than its month's ${total.learners} and ${total.answers}`,
			);
		}
	};
}

// levelsOf are the levels of the reference tasks, from the lowest, each with
// the grades it spans.
function levelsOf(byLevel: Readonly<Record<string, number>>): Level[] {
	return Object.entries(byLevel)
		.map(([key, tasks]) => {
			const [first, last] = gradesOf([key]);
			return { key, first, last, count: tasks };
		})
		.sort((a, b) => a.first - b.first);
}
