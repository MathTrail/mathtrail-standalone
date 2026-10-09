import { z } from "zod";
import topics from "../../../content/catalogs/topics.json";
import traps from "../../../content/catalogs/traps.json";
import file from "../../../site/data.json";
import { type ProgressReport, readScreen } from "../widget/payload";
import {
	type Drawing,
	drawingFile,
	readDrawings,
	type TopicDrawings,
	topicDrawingsFile,
} from "./drawings";
import { type Home, homeFile, readHome } from "./home";
import { type Research, readResearch, researchSources } from "./research";
import { readTechniques, type Techniques } from "./techniques";
import { type CatalogTopic, gradesOf, readTopics, type Topics } from "./topics";
import { type CatalogTrap, type ReferenceTask, rankTraps } from "./traps";
import { readWhy, type Why, whyFile } from "./why";

/**
 * Catalog is what the site reads of the service's content: its topics, its
 * traps, and its reference tasks, whose wrong options name the traps.
 */
export type Catalog = {
	readonly topics: readonly CatalogTopic[];
	readonly traps: readonly CatalogTrap[];
	readonly tasks: readonly ReferenceTask[];
};

/**
 * Example is an example a page works through, as far as no language changes
 * it: the first and the last grade of the level it is set at.
 */
export type Example = { readonly grades: readonly [number, number] };

/**
 * TopicExample is an example a topic's page works through: an example, and
 * the drawing beside its steps where it has one.
 */
export type TopicExample = Example & { readonly drawing?: Drawing };

/**
 * SiteData is what the site's pages draw that no language changes: the
 * catalog's topics with the groups they are shown in, each topic's traps, the
 * drawings of its card and its page, the examples its page works through, the
 * progress of the child the site's cards are drawn for, an invented one, what
 * the page "Why" shows — its card and the works it cites — the techniques of
 * the page of the techniques, and the numbers of the page "Research".
 */
export type SiteData = {
	readonly topics: Topics;
	/** traps are each topic's traps, the most frequent in its reference tasks first. */
	readonly traps: ReadonlyMap<string, readonly string[]>;
	/** drawings are the drawings of each topic's card and page, by the topic's id. */
	readonly drawings: ReadonlyMap<string, TopicDrawings>;
	/** examples are the examples of each topic's page, by the topic's id. */
	readonly examples: ReadonlyMap<string, readonly TopicExample[]>;
	/** progress is a progress as the service sends it, but for the child's name. */
	readonly progress: Sample;
	/** why is what the page "Why" draws, when the data has it. */
	readonly why?: Why;
	/** home is what the home page draws, when the data has it. */
	readonly home?: Home;
	/** techniques are what the page of the techniques draws, when the data has them. */
	readonly techniques?: Techniques;
	/**
	 * research is what the page "Research" draws, when the build was given the
	 * file of its numbers.
	 */
	readonly research?: Research;
};

// sample is a progress as the service sends it, whose profile is completed
// with the name a page gives the child.
const sample = z.looseObject({ profile: z.record(z.string(), z.unknown()) });

// Sample is a progress the site's data holds, before a page names the child.
type Sample = z.infer<typeof sample>;

// solverName is what an example's solver may be called: lowercase words and
// numbers joined by dashes, the name of its file.
const solverName = /^[a-z0-9]+(?:-[a-z0-9]+)*$/;

// exampleFile is an example as the site's data writes it: the level it is set
// at, and the solver and the answer that prove it.
const exampleFile = z.object({
	level: z.string(),
	solver: z.string().regex(solverName),
	answer: z.string().min(1),
});

// topicExampleFile is an example of a topic's page as the site's data writes
// it: an example, and its drawing where it has one. Nothing else may stand
// beside them, so that a drawing under a misspelt name stops the build rather
// than leaving the example drawn without it.
const topicExampleFile = z.strictObject({
	...exampleFile.shape,
	drawing: drawingFile.optional(),
});

// techniquesFile is what the page of the techniques takes from the site's
// data: its groups, each with its techniques in order — a technique's name,
// the topics it leads to and its example — and the rows of its hint, each
// with the techniques it suggests.
const techniquesFile = z.object({
	groups: z
		.array(
			z.object({
				id: z.string(),
				techniques: z
					.array(
						z.object({
							id: z.string(),
							topics: z.array(z.string()),
							example: exampleFile,
						}),
					)
					.min(1),
			}),
		)
		.min(1),
	cues: z.array(z.array(z.string()).min(1)).min(1),
});

/** TechniquesFile is what the page of the techniques takes from the site's data. */
export type TechniquesFile = z.infer<typeof techniquesFile>;

// dataFile is the shape of the site's data file: its groups, the drawings and
// the examples of the topics' cards and pages, the progress its cards are
// drawn from, what the page "Why" draws, the techniques of the page of the
// techniques, and the authors the page "Research" names.
const dataFile = z.object({
	groups: z.array(z.object({ id: z.string(), topics: z.array(z.string()) })),
	drawings: z.record(z.string(), topicDrawingsFile).optional(),
	examples: z.record(z.string(), z.array(topicExampleFile)),
	progress: sample,
	why: whyFile.optional(),
	home: homeFile.optional(),
	techniques: techniquesFile.optional(),
	research: researchSources.optional(),
});

/**
 * readSiteData reads the site's data beside the catalog it speaks of. A file
 * of another shape is refused, and so is a progress the widget could not draw
 * whatever the child's name, a drawing of a topic the catalog does not have,
 * a first screen's drawing or an example of a topic whose page is not
 * published, and an example at a level the topic is not taught at, so that a
 * mistake in the file stops the build rather than drawing a page with a part
 * missing.
 * What the home page and the page "Why" draw, and the techniques, are held
 * to the catalog as well, and so are the numbers of the page "Research", when
 * the build is given them.
 */
export function readSiteData(
	catalog: Catalog,
	data: unknown,
	research?: unknown,
): SiteData {
	const read = dataFile.safeParse(data);
	if (!read.success) {
		throw new Error(`site/data.json: ${z.prettifyError(read.error)}`);
	}
	const site = readParts(catalog, read.data);
	return research === undefined
		? site
		: {
				...site,
				research: readResearch(catalog, read.data.research, research),
			};
}

// readParts reads every part of the site's data the pages draw, a mistake in
// any of them named as the data file's.
function readParts(catalog: Catalog, data: z.infer<typeof dataFile>): SiteData {
	const ranked = rankTraps(catalog.traps, catalog.tasks);
	try {
		const site = {
			topics: readTopics(catalog.topics, data.groups),
			traps: ranked,
			drawings: readDrawings(catalog.topics, data.drawings ?? {}),
			examples: readExamples(catalog.topics, data.examples),
			progress: data.progress,
			why: data.why === undefined ? undefined : readWhy(catalog, data.why),
			home: data.home === undefined ? undefined : readHome(catalog, data.home),
			techniques:
				data.techniques === undefined
					? undefined
					: readTechniques(catalog.topics, data.techniques),
		};
		progressOf(site, "Comet");
		return site;
	} catch (error) {
		const said = error instanceof Error ? error.message : "it cannot be read";
		throw new Error(`site/data.json: ${said}`, { cause: error });
	}
}

// referenceTasks are the service's reference tasks, of every topic.
const referenceTasks: readonly ReferenceTask[] = Object.values(
	import.meta.glob<readonly ReferenceTask[]>(
		"../../../content/examples/*.json",
		{ eager: true, import: "default" },
	),
).flat();

/** siteCatalog is the service's catalogs and reference tasks, as the site reads them. */
export const siteCatalog: Catalog = { topics, traps, tasks: referenceTasks };

/**
 * siteData reads the data of this site: its own file, the service's catalogs
 * and reference tasks, and, when the build is given it, the file of the
 * numbers of the page "Research".
 */
export function siteData(research?: unknown): SiteData {
	return readSiteData(siteCatalog, file, research);
}

// readExamples reads the examples of the topics' pages: each of a topic of the
// catalog whose page is published, set at a level the topic is taught at, with
// its drawing where it has one.
function readExamples(
	catalog: readonly CatalogTopic[],
	examples: Readonly<
		Record<string, readonly { level: string; drawing?: Drawing }[]>
	>,
): ReadonlyMap<string, readonly TopicExample[]> {
	const byId = new Map(catalog.map((topic) => [topic.id, topic]));
	return new Map(
		Object.entries(examples).map(([id, listed]) => {
			const topic = byId.get(id);
			if (topic === undefined) {
				throw new Error(
					`the examples of ${id} are of a topic the catalog does not have`,
				);
			}
			if (!topic.site_page) {
				throw new Error(
					`the examples of ${id} are for a page the catalog does not publish`,
				);
			}
			return [
				id,
				listed.map(({ level, drawing }): TopicExample => {
					const example = exampleAt(topic, level);
					return drawing === undefined ? example : { ...example, drawing };
				}),
			];
		}),
	);
}

// exampleAt is an example of topic set at level, which has to be one the
// topic is taught at.
function exampleAt(topic: CatalogTopic, level: string): Example {
	if (!topic.grade_levels.includes(level)) {
		throw new Error(
			`an example of ${topic.id} is set at ${level}, a level the topic is not taught at`,
		);
	}
	return { grades: gradesOf([level]) };
}

/**
 * progressOf is the progress of the site's data as the widget reads one,
 * the child called child: the language a page draws it in names the child.
 */
export function progressOf(data: SiteData, child: string): ProgressReport {
	const screen = readScreen({
		...data.progress,
		profile: { ...data.progress.profile, pseudonym: child },
	});
	if (screen?.screen !== "progress") {
		throw new Error("the progress is no progress the widget can draw");
	}
	return screen.report;
}
