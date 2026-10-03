// @vitest-environment node
import { readFileSync } from "node:fs";
import { join } from "node:path";
import { describe, expect, test } from "vitest";
import { scenesIn } from "../src/preview/scenes.ts";
import { previewWidths } from "../src/preview/widths.ts";
import { addressOf, pictureOf, shots, themes, width } from "./screens.ts";

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

	test("ask the preview for one scene alone, in its theme, in English, at the width photographed", () => {
		const address = new URL(
			addressOf(
				"http://127.0.0.1:5173/",
				{ scene: "wrong", file: "wrong" },
				"dark",
			),
		);
		expect(address.pathname).toBe("/preview.html");
		expect(Object.fromEntries(address.searchParams)).toEqual({
			scene: "wrong",
			theme: "dark",
			lang: "en",
			widths: "428",
		});
	});

	test("are written to the files the README shows, one for each scene and theme", () => {
		const files = themes.flatMap((theme) =>
			shots.map((shot) =>
				pictureOf(shot, theme).split("/").slice(-3).join("/"),
			),
		);
		expect(files.sort()).toEqual([
			"docs/screens/progress-dark.png",
			"docs/screens/progress-light.png",
			"docs/screens/task-dark.png",
			"docs/screens/task-light.png",
			"docs/screens/wrong-dark.png",
			"docs/screens/wrong-light.png",
		]);
	});
});

describe("the README", () => {
	// A picture written that the README does not show is a picture nobody sees,
	// and one it shows that is not written is a picture gone stale.
	test("shows every picture written, each in the theme of its page", () => {
		const readme = readFileSync(
			join(import.meta.dirname, "..", "..", "README.md"),
			"utf8",
		);
		for (const shot of shots) {
			expect(readme).toContain(
				`<source media="(prefers-color-scheme: dark)" srcset="docs/screens/${shot.file}-dark.png">`,
			);
			expect(readme).toContain(
				`<img src="docs/screens/${shot.file}-light.png" width="${width}"`,
			);
		}
	});
});
