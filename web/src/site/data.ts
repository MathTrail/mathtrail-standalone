import { z } from "zod";
import catalog from "../../../content/catalogs/topics.json";
import file from "../../../site/data.json";
import { type ProgressReport, readScreen } from "../widget/payload";
import { type CatalogTopic, readTopics, type Topics } from "./topics";

/**
 * SiteData is what the site's pages draw that no language changes: the
 * catalog's topics with the groups they are shown in, and the progress of the
 * child the site's cards are drawn for, an invented one.
 */
export type SiteData = {
	readonly topics: Topics;
	/** progress is a progress as the service sends it, but for the child's name. */
	readonly progress: Sample;
};

// sample is a progress as the service sends it, whose profile is completed
// with the name a page gives the child.
const sample = z.looseObject({ profile: z.record(z.string(), z.unknown()) });

// Sample is a progress the site's data holds, before a page names the child.
type Sample = z.infer<typeof sample>;

// dataFile is the shape of the site's data file: its groups, and the progress
// its cards are drawn from.
const dataFile = z.object({
	groups: z.array(z.object({ id: z.string(), topics: z.array(z.string()) })),
	progress: sample,
});

/**
 * readSiteData reads the site's data beside the catalog it speaks of. A file
 * of another shape is refused, and so is a progress the widget could not draw
 * whatever the child's name, so that a mistake in the file stops the build
 * rather than drawing a card with a part missing.
 */
export function readSiteData(
	topics: readonly CatalogTopic[],
	data: unknown,
): SiteData {
	const read = dataFile.safeParse(data);
	if (!read.success) {
		throw new Error(`site/data.json: ${z.prettifyError(read.error)}`);
	}
	try {
		const site = {
			topics: readTopics(topics, read.data.groups),
			progress: read.data.progress,
		};
		progressOf(site, "Comet");
		return site;
	} catch (error) {
		const said = error instanceof Error ? error.message : "it cannot be read";
		throw new Error(`site/data.json: ${said}`, { cause: error });
	}
}

/** siteData reads the data of this site: its own file and the service's catalog. */
export function siteData(): SiteData {
	return readSiteData(catalog, file);
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
