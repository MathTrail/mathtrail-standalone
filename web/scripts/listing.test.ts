// @vitest-environment node
import { readFileSync } from "node:fs";
import { join, relative } from "node:path";
import { describe, expect, test } from "vitest";
import { scenesIn } from "../src/preview/scenes.ts";
import { previewWidths } from "../src/preview/widths.ts";
import { density } from "./drive.ts";
import { pictureOf, shots, width } from "./listing.ts";

const root = join(import.meta.dirname, "..", "..");

/**
 * promptsOf are the numbered prompts of the listing's section on its
 * screenshots, in docs/listing.md, in their order: the section runs to the
 * next heading, of whatever level.
 */
function promptsOf(): string[] {
	const listing = readFileSync(join(root, "docs", "listing.md"), "utf8");
	const section = listing
		.split("### Screenshots (CL-11)")[1]
		?.split(/\n#+ /)[0];
	if (section === undefined) {
		throw new Error("docs/listing.md has no section on its screenshots");
	}
	return [...section.matchAll(/^\d+\. (.+)$/gm)].map(
		([, prompt]) => prompt ?? "",
	);
}

describe("the pictures of the listing in Claude's directory", () => {
	// A scene the preview does not have, or a width it does not offer, would
	// leave the page with no card to photograph.
	test("are of scenes the preview has, at a width it offers", () => {
		const scenes = scenesIn("en").map((scene) => scene.name);
		for (const shot of shots) {
			expect(scenes).toContain(shot.scene);
		}
		expect(previewWidths).toContain(width);
	});

	// The directory takes three to five pictures, at least 1000 px wide.
	test("are three to five, at least 1000 px wide, in both themes", () => {
		expect(shots.length).toBeGreaterThanOrEqual(3);
		expect(shots.length).toBeLessThanOrEqual(5);
		expect(width * density).toBeGreaterThanOrEqual(1000);
		expect(new Set(shots.map((shot) => shot.theme))).toEqual(
			new Set(["light", "dark"]),
		);
	});

	test("are each of its scene, in the theme it names, at the width photographed", () => {
		for (const shot of shots) {
			expect(pictureOf(shot)).toMatchObject({
				scene: shot.scene,
				theme: shot.theme,
				width,
			});
		}
	});

	test("are written into web/listing/, each under a name of its own", () => {
		const files = shots.map((shot) => relative(root, pictureOf(shot).path));
		expect(new Set(files).size).toBe(shots.length);
		for (const file of files) {
			expect(file).toMatch(/^web\/listing\/[a-z0-9-]+\.png$/);
		}
	});

	// Each picture is uploaded beside a prompt of its own: the words typed in
	// the chat, quoted, and the theme the picture is in, which names no other.
	test("match the prompts docs/listing.md gives them, one by one, in theme", () => {
		const prompts = promptsOf();
		expect(prompts).toHaveLength(shots.length);
		shots.forEach((shot, at) => {
			const other = shot.theme === "light" ? "dark" : "light";
			expect(prompts[at]).toMatch(/^"[^"]+"/);
			expect(prompts[at]).toContain(`${shot.theme} theme`);
			expect(prompts[at]).not.toContain(`${other} theme`);
		});
	});

	// The pictures show one task, drawn alike in each, and those taken before
	// the child answers must not give the answer away: the row of posts drawn
	// in full would show the count, which is the answer.
	test("draw the task's row cut short in every picture of the task", () => {
		const scenes = scenesIn("en");
		const drawings = shots.flatMap((shot) => {
			const scene = scenes.find((found) => found.name === shot.scene);
			const handed = scene?.payload as { task?: { drawing?: string } };
			return handed?.task === undefined
				? []
				: [{ scene: shot.scene, drawing: handed.task.drawing }];
		});
		expect(drawings.length).toBeGreaterThan(0);
		for (const { scene, drawing } of drawings) {
			expect(drawing, scene).toContain(" ... ");
		}
	});
});
