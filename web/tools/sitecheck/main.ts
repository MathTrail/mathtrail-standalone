// The site's checker on the command line: it judges one built directory and
// refuses a site that should not be published, naming every problem at once.
//
//	node tools/sitecheck/main.ts --base https://mathtrail.app --dir ../site/dist

import { parseArgs } from "node:util";
import { check, lineOf, type Options } from "./check.ts";
import { published } from "./published.ts";

const usage =
	"usage: node tools/sitecheck/main.ts --base <origin> --dir <directory> [--reference-locale <locale>] [--max-page-bytes <bytes>] [--max-frame-bytes <bytes>]";

/** defaultMaxPageBytes is what a page may weigh with everything it loads. */
export const defaultMaxPageBytes = 300 * 1024;

/**
 * defaultMaxFrameBytes is what a document a page frames may weigh with
 * everything it loads.
 */
export const defaultMaxFrameBytes = 2 * 1024 * 1024;

/**
 * photoDirectory is where the site keeps its photographs, which a page is
 * weighed without.
 */
export const photoDirectory = "/assets/photos/";

/**
 * main judges the site a command line names and returns the command's exit
 * code: 0 for a publishable site, 1 for one with problems or one that cannot be
 * read, 2 for a command line it does not understand. Every problem is said
 * before it fails, so that one run tells a person everything there is to fix.
 */
export async function main(
	args: string[],
	out: (line: string) => void,
	err: (line: string) => void,
): Promise<number> {
	const options = optionsOf(args);
	if (options === undefined) {
		err(usage);
		return 2;
	}
	let findings: Awaited<ReturnType<typeof check>>;
	try {
		findings = await check(options.dir, options);
	} catch (error) {
		err(
			`sitecheck: ${error instanceof Error ? error.message : "the site cannot be read"}`,
		);
		return 1;
	}
	for (const finding of findings) {
		out(lineOf(finding));
	}
	if (findings.length > 0) {
		err(`sitecheck: ${findings.length} problems in ${options.dir}`);
		return 1;
	}
	out(`${options.dir} is publishable`);
	return 0;
}

// optionsOf reads a command line, or returns nothing when it is not one main
// understands.
function optionsOf(args: string[]): (Options & { dir: string }) | undefined {
	let values: {
		base?: string;
		dir?: string;
		"reference-locale"?: string;
		"max-page-bytes"?: string;
		"max-frame-bytes"?: string;
	};
	try {
		({ values } = parseArgs({
			args,
			options: {
				base: { type: "string" },
				dir: { type: "string" },
				"reference-locale": { type: "string" },
				"max-page-bytes": { type: "string" },
				"max-frame-bytes": { type: "string" },
			},
		}));
	} catch {
		return undefined;
	}
	const maxPageBytes = bytesOf(values["max-page-bytes"], defaultMaxPageBytes);
	const maxFrameBytes = bytesOf(
		values["max-frame-bytes"],
		defaultMaxFrameBytes,
	);
	if (
		values.base === undefined ||
		values.dir === undefined ||
		!Number.isSafeInteger(maxPageBytes) ||
		!Number.isSafeInteger(maxFrameBytes)
	) {
		return undefined;
	}
	return {
		base: values.base,
		dir: values.dir,
		referenceLocale: values["reference-locale"] ?? "en",
		maxPageBytes,
		maxFrameBytes,
		photos: photoDirectory,
		published,
	};
}

// bytesOf reads a budget from the command line, or takes the default when the
// line names none. A budget is a count of bytes in decimal digits: an empty
// value would read as zero, which turns its rule off rather than failing, so
// anything else reads as no number at all.
function bytesOf(value: string | undefined, fallback: number): number {
	const budget = value ?? String(fallback);
	return /^\d+$/.test(budget) ? Number(budget) : Number.NaN;
}

if (import.meta.main) {
	process.exitCode = await main(
		process.argv.slice(2),
		(line) => console.log(line),
		(line) => console.error(line),
	);
}
