// @vitest-environment node
import { mkdtemp, readdir, readFile, rm } from "node:fs/promises";
import { tmpdir } from "node:os";
import { join } from "node:path";
import { build } from "vite";
import { expect, test } from "vitest";
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
	} finally {
		await rm(out, { recursive: true, force: true });
	}
}, 60_000);
