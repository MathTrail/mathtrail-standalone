// @vitest-environment node
import { mkdtemp, readdir, readFile, rm } from "node:fs/promises";
import { tmpdir } from "node:os";
import { join } from "node:path";
import { build } from "vite";
import { expect, test } from "vitest";
import { pseudoLocale } from "./src/i18n/pseudo.ts";
import config from "./vite.config.widget.ts";

// The host runs the widget in a sandbox that lets it load nothing, so a build
// that left a script or a stylesheet beside the page would ship a card that
// never draws.
test("the widget builds into one page that loads nothing from outside it", async () => {
	const out = await mkdtemp(join(tmpdir(), "widget-"));
	try {
		await build({
			...config,
			configFile: false,
			logLevel: "silent",
			build: { ...config.build, outDir: out, emptyOutDir: true },
		});

		expect(await readdir(out, { recursive: true })).toEqual(["widget.html"]);
		const page = await readFile(join(out, "widget.html"), "utf8");
		expect(page).toMatch(/<script type="module"[^>]*>\S/);
		expect(page).not.toMatch(/<script\b[^>]*\bsrc=/);
		expect(page).not.toMatch(/<link\b[^>]*\bhref=/);
		expect(page).toContain("ui/initialize");
		// The design's tokens and components are inside the page too, and no
		// stylesheet is left to be fetched from anywhere.
		expect(page).toContain("--surface:");
		expect(page).toContain(".mt-option");
		expect(page).not.toContain("@import");
		// The pseudo-language is the preview's and the tests': a host could
		// name its tag, and a card that spoke it would show a child stretched
		// English. The test runner's own environment does not change that.
		expect(page).not.toContain(pseudoLocale);
	} finally {
		await rm(out, { recursive: true, force: true });
	}
}, 60_000);
