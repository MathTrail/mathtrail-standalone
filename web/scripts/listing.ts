// The pictures of the card a listing in Claude's directory shows: a task, its
// hint, a wrong answer explained and the progress's review, each the card
// alone with its corners clear. A card 640 px wide is photographed at twice
// its density, so that each picture is 1280 px wide, more than the 1000 the
// directory asks for. They are the preview's own scenes, photographed in
// Chromium as the README's pictures are. The version a card's header shows is
// left out: a picture uploaded to a directory is not taken again with the
// next release, and a version on it would soon be an old one.
//
//	node scripts/listing.ts
//
// It runs where the browsers are, inside the image of the pinned Playwright
// release, and writes the pictures anew into web/listing/, to be uploaded by
// hand with the prompts docs/listing.md gives them.

import { mkdir, rm } from "node:fs/promises";
import { join } from "node:path";
import { type Picture, photograph } from "./drive.ts";

/**
 * Shot is a scene of the preview, the theme it is photographed in, and the
 * file its picture is written to.
 */
export type Shot = { scene: string; theme: "light" | "dark"; file: string };

/**
 * shots are the listing's pictures, in the order docs/listing.md gives their
 * prompts: a card past the trial series, whose row of posts is drawn cut
 * short, so that no picture before the answer gives the count away.
 */
export const shots: readonly Shot[] = [
	{
		scene: "task after the trial series",
		theme: "light",
		file: "1-task-light",
	},
	{
		scene: "hint after the trial series",
		theme: "dark",
		file: "2-hint-dark",
	},
	{
		scene: "result card, the picture of the solution, wrong",
		theme: "light",
		file: "3-wrong-light",
	},
	{
		scene: "progress, the review open",
		theme: "dark",
		file: "4-progress-dark",
	},
];

/** width is how wide a card is photographed, in the pixels of a page. */
export const width = 640;

const listing = join(import.meta.dirname, "..", "listing");

/**
 * hidden is what a picture leaves out of the card: the version in its header,
 * whose room stays as the header lays it out.
 */
const hidden = ".mt-version { visibility: hidden; }";

/**
 * pictureOf is how a scene is photographed: in its theme and at the width
 * photographed, into the file its picture is written to.
 */
export function pictureOf(shot: Shot): Picture {
	return {
		scene: shot.scene,
		theme: shot.theme,
		width,
		path: join(listing, `${shot.file}.png`),
	};
}

async function main(): Promise<void> {
	await rm(listing, { recursive: true, force: true });
	await mkdir(listing, { recursive: true });
	await photograph("listing", shots.map(pictureOf), hidden);
}

if (import.meta.main) {
	await main();
}
