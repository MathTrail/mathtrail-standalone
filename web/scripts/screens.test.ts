// @vitest-environment node
import { readFileSync } from "node:fs";
import { join, relative } from "node:path";
import { describe, expect, test } from "vitest";
import { scenesIn } from "../src/preview/scenes.ts";
import { previewWidths } from "../src/preview/widths.ts";
import { addressOf, pictureOf, shots, width } from "./screens.ts";

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

	test("are written to the files the README shows, one for each scene", () => {
		const files = shots.map((shot) => relative(root, pictureOf(shot)));
		expect(files.sort()).toEqual([
			"docs/screens/progress-dark.png",
			"docs/screens/task-dark.png",
			"docs/screens/wrong-dark.png",
		]);
	});
});

describe("the README", () => {
	// A picture written that the README does not show is a picture nobody sees,
	// and one it shows that is not written is a picture gone stale. Showing the
	// pictures written and no other is also what shows a reader of a light page
	// the dark theme: no lighter picture is left to choose.
	test("shows every picture written and no other, at the width photographed", () => {
		const readme = readFileSync(join(root, "README.md"), "utf8");
		const shown = readme.match(/docs\/screens\/[^\s"'`),]+/g) ?? [];
		const written = shots.map((shot) => relative(root, pictureOf(shot)));
		expect(shown.sort()).toEqual(written.sort());
		for (const shot of shots) {
			expect(readme).toContain(
				`<img src="${relative(root, pictureOf(shot))}" width="${width}"`,
			);
		}
	});
});
