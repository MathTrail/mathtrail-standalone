// Whether a change touches the widget's cards: whether any of the files it
// changes is one the preview's build reads or stands beside one, or one the
// measuring of the cards is made of. A change that touches none of them
// leaves every card as it was, and its layout needs no measuring again.
//
//	git diff --name-only HEAD^1 HEAD | node scripts/scope.ts [--to <file>]
//
// The files changed come one a line, as paths from the repository's root. It
// writes card=true or card=false, as a step of a workflow sets an output, to
// the end of the file named, or else to its standard output, and says on its
// error stream which file decided it.

import { appendFile } from "node:fs/promises";
import { dirname, join, relative } from "node:path";
import { text } from "node:stream/consumers";
import { parseArgs } from "node:util";
import type { Plugin } from "vite";
import { builtPreview } from "./drive.ts";

const repository = join(import.meta.dirname, "..", "..");

/**
 * measuring are the files the measuring of the cards is made of, which no
 * build reads: the scripts that drive the browsers and say what is measured,
 * the build's configuration and packages, and what picks the browsers'
 * image, the Node that runs it and the parts it is shared out in.
 */
export const measuring: readonly string[] = [
	"web/scripts/layout.ts",
	"web/scripts/drive.ts",
	"web/scripts/scope.ts",
	"web/vite.config.preview.ts",
	"web/package.json",
	"web/package-lock.json",
	"web/tsconfig.json",
	"justfile",
	".github/workflows/ci.yml",
	".github/actions/node/action.yml",
	".devcontainer/Dockerfile",
];

// reading is a plugin that keeps, once the build is over, the id of every
// module it read.
function reading(read: Set<string>): Plugin {
	return {
		name: "mathtrail:reading",
		buildEnd() {
			for (const id of this.getModuleIds()) {
				read.add(id);
			}
		},
	};
}

/**
 * drawnFrom are the files of the repository the cards are drawn from, as
 * paths from its root: every one the preview's build reads, the widget's own
 * sources and the data and words outside them alike, but for the packages it
 * installs, which its lockfile stands for.
 */
export async function drawnFrom(): Promise<Set<string>> {
	const read = new Set<string>();
	await builtPreview({ plugins: [reading(read)] });
	return new Set(
		[...read]
			.filter((id) => !id.startsWith("\0"))
			.map((id) => relative(repository, id.split("?")[0] ?? id))
			.filter(
				(path) => !path.startsWith("..") && !path.includes("node_modules/"),
			),
	);
}

/**
 * touching is the first of the files changed that the cards are drawn from,
 * that stands in a folder one of them stands in, or that the cards are
 * measured by; or nothing where none is. A folder counts, since the build
 * reads some of its files by a pattern: a dictionary or a picture added or
 * taken away is no file the build read.
 */
export function touching(
	changed: readonly string[],
	drawn: ReadonlySet<string>,
): string | undefined {
	const folders = new Set([...drawn].map(dirname));
	return changed.find(
		(path) =>
			drawn.has(path) || folders.has(dirname(path)) || measuring.includes(path),
	);
}

async function main(): Promise<void> {
	const { values } = parseArgs({ options: { to: { type: "string" } } });
	const changed = (await text(process.stdin))
		.split("\n")
		.map((line) => line.trim())
		.filter((line) => line !== "");
	const touched = touching(changed, await drawnFrom());
	console.error(
		touched === undefined
			? `scope: none of the ${changed.length} files changed is one the cards are drawn from`
			: `scope: the cards are drawn from ${touched}, which changed`,
	);
	const answer = `card=${touched !== undefined}\n`;
	if (values.to === undefined) {
		process.stdout.write(answer);
	} else {
		await appendFile(values.to, answer);
	}
}

if (import.meta.main) {
	await main();
}
