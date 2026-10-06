// @vitest-environment node
import { readFileSync } from "node:fs";
import { join, relative } from "node:path";
import { describe, expect, test } from "vitest";
import { scenesIn } from "../src/preview/scenes.ts";
import { previewWidths } from "../src/preview/widths.ts";
import { addressOf, pictureOf, shots, shownAt, width } from "./screens.ts";

const root = join(import.meta.dirname, "..", "..");

describe("the pictures of the README", () => {
	// A scene the preview does not have, or a width it does not offer, would
	// leave the page with no card to photograph.
	test("are of scenes the preview has, at a width it offers", () => {
		const scenes = scenesIn("en").map((scene) => scene.name);
		for (const shot of shots) {
			expect(scenes).toContain(shot.scene);
		}
		expect(previewWidths).toContain(width);
	});

	test("ask the preview for one scene alone, in the dark theme, in English, at the width photographed", () => {
		const address = new URL(
			addressOf("http://127.0.0.1:5173/", { scene: "wrong", file: "wrong" }),
		);
		expect(address.pathname).toBe("/preview.html");
		expect(Object.fromEntries(address.searchParams)).toEqual({
			scene: "wrong",
			theme: "dark",
			lang: "en",
			widths: "428",
		});
	});

	test("are written into web/screens/, one for each scene", () => {
		const files = shots.map((shot) => relative(root, pictureOf(shot)));
		expect(files.sort()).toEqual([
			"web/screens/progress-dark.png",
			"web/screens/task-dark.png",
			"web/screens/wrong-dark.png",
		]);
	});
});

describe("the README", () => {
	// A picture written that the README does not show is a picture nobody sees,
	// and one it shows that is not written is a picture missing from the page,
	// whether HTML or Markdown asks for it, and at whatever address. Showing the
	// pictures written and no other is also what shows a reader of a light page
	// the dark theme: no lighter picture is left to choose.
	test("shows every picture written and no other, at the width photographed", () => {
		const readme = readFileSync(join(root, "README.md"), "utf8");
		const shown = readme.match(/[^\s"'`()<>[\]]*screens\/[a-z-]+\.png/g) ?? [];
		expect(shown.sort()).toEqual(shots.map(shownAt).sort());
		for (const shot of shots) {
			expect(readme).toContain(`<img src="${shownAt(shot)}" width="${width}"`);
		}
	});
});
