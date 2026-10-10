// The pictures of the widget the README shows: a task, how a wrong answer
// went, with the picture of its solution, and the progress, in the dark theme, as a phone 428 px wide shows them
// at twice its density. They are the preview's own scenes — the widget's page,
// driven as a chat host drives it — photographed in Chromium, so that the
// README shows the cards as they are rather than as they were designed. The
// README shows the dark theme to a reader of either, and the corners round the
// card are left clear, for a page of either theme to show through.
//
//	node scripts/screens.ts
//
// It runs where the browsers are, inside the image of the pinned Playwright
// release, and writes the pictures anew into web/screens/. The README shows
// them from a branch of their own, where they are published under the names
// they are written under.

import { mkdir, rm } from "node:fs/promises";
import { join } from "node:path";
import { type Picture, photograph } from "./drive.ts";

/** Shot is a scene of the preview, and the file its pictures are named by. */
export type Shot = { scene: string; file: string };

/**
 * shots are the scenes the README shows, in its order: the cards of a child
 * past the trial series, which offer the choice of the topic.
 */
export const shots: readonly Shot[] = [
	{ scene: "task after the trial series", file: "task" },
	{ scene: "result card, the picture of the solution, wrong", file: "wrong" },
	{ scene: "progress, the model's card", file: "progress" },
];

/** theme is the theme every scene is photographed in. */
const theme = "dark";

/** width is how wide a card is photographed, in the pixels of a page. */
export const width = 428;

const screens = join(import.meta.dirname, "..", "screens");

/**
 * published is the address of the branch the README shows the pictures from.
 * The branch holds the newest pictures alone. GitHub shows a README a picture
 * from this address within minutes of a change, where a picture from
 * elsewhere comes from a copy it keeps.
 */
const published =
	"https://raw.githubusercontent.com/MathTrail/mathtrail-standalone/screens/";

/** nameOf is the name a scene's picture is written and published under. */
function nameOf(shot: Shot): string {
	return `${shot.file}-${theme}.png`;
}

/**
 * pictureOf is how a scene is photographed: in the theme and at the width
 * photographed, into the file its picture is written to.
 */
export function pictureOf(shot: Shot): Picture {
	return { scene: shot.scene, theme, width, path: join(screens, nameOf(shot)) };
}

/** shownAt is the address the README shows a scene's picture at. */
export function shownAt(shot: Shot): string {
	return `${published}${nameOf(shot)}`;
}

async function main(): Promise<void> {
	// Whatever the folder holds is published, so a picture of a scene no longer
	// photographed must not outlive the run that stopped photographing it.
	await rm(screens, { recursive: true, force: true });
	await mkdir(screens, { recursive: true });
	await photograph("screens", shots.map(pictureOf));
}

if (import.meta.main) {
	await main();
}
