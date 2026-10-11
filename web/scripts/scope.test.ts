// @vitest-environment node
import { spawnSync } from "node:child_process";
import { mkdtemp, readdir, readFile, rm } from "node:fs/promises";
import { tmpdir } from "node:os";
import { join } from "node:path";
import { describe, expect, test, vi } from "vitest";
import { drawnFrom, measuring, scope, touching } from "./scope.ts";

describe("the files the cards are drawn from", () => {
	test("are those the preview's build reads, the data and words outside the widget's sources among them", async () => {
		const drawn = await drawnFrom();

		for (const path of [
			"web/widget.html",
			"web/src/widget/main.ts",
			"web/src/preview/scenes.ts",
			"web/locales/ru.json",
			"internal/widget/tokens.css",
			"site/data.json",
			"content/pictures/flags.json",
		]) {
			expect(drawn.has(path), path).toBe(true);
		}
		for (const path of drawn) {
			expect(path, path).not.toMatch(/^\.\.|node_modules\/|^web\/src\/site\//);
		}
	}, 60_000);
});

describe("a change", () => {
	// drawn are some of the files the cards are drawn from.
	const drawn = new Set(["web/src/widget/main.ts", "site/data.json"]);

	test("touches the cards where it changes a file they are drawn from", () => {
		expect(touching(["docs/decisions.md", "site/data.json"], drawn)).toBe(
			"site/data.json",
		);
	});

	test("touches the cards where it changes how they are measured", () => {
		for (const path of measuring) {
			expect(touching([path], drawn), path).toBe(path);
		}
	});

	test("touches the cards where it changes a workflow that measures them", async () => {
		const workflows = join(
			import.meta.dirname,
			"..",
			"..",
			".github",
			"workflows",
		);
		const measuringThem: string[] = [];
		for (const file of await readdir(workflows)) {
			const text = await readFile(join(workflows, file), "utf8");
			if (text.includes("just web-layout")) {
				measuringThem.push(`.github/workflows/${file}`);
			}
		}

		expect(measuringThem).not.toEqual([]);
		expect(measuring).toEqual(expect.arrayContaining(measuringThem));
	});

	test("touches the cards where it adds or takes away a file beside one they are drawn from", () => {
		expect(touching(["web/src/widget/scenes.json"], drawn)).toBe(
			"web/src/widget/scenes.json",
		);
	});

	test("leaves the cards as they were where it changes neither", () => {
		expect(
			touching(
				[
					"web/src/site/HomePage.tsx",
					"internal/domain/rating/elo.go",
					"README.md",
				],
				drawn,
			),
		).toBeUndefined();
		expect(touching([], drawn)).toBeUndefined();
	});
});

describe("the answer", () => {
	// drawn are some of the files the cards are drawn from.
	const drawn = new Set(["web/src/widget/main.ts", "site/data.json"]);

	test("is written to the end of the file named, and names the file that decided it", async () => {
		const folder = await mkdtemp(join(tmpdir(), "scope-"));
		const told = vi.spyOn(console, "error").mockImplementation(() => {});
		try {
			const to = join(folder, "output");
			await scope("README.md\n  site/data.json  \n\n", drawn, to);

			expect(await readFile(to, "utf8")).toBe("card=true\n");
			expect(told).toHaveBeenCalledWith(
				"scope: the cards are drawn from site/data.json, which changed",
			);
		} finally {
			told.mockRestore();
			await rm(folder, { recursive: true, force: true });
		}
	});

	test("is written to the standard output where no file is named, and says none of the files touched the cards", async () => {
		const told = vi.spyOn(console, "error").mockImplementation(() => {});
		const written = vi
			.spyOn(process.stdout, "write")
			.mockImplementation(() => true);
		try {
			await scope("README.md\ndocs/decisions.md\n", drawn);

			expect(written).toHaveBeenCalledWith("card=false\n");
			expect(told).toHaveBeenCalledWith(
				"scope: none of the 2 files changed is one the cards are drawn from",
			);
		} finally {
			written.mockRestore();
			told.mockRestore();
		}
	});

	test("is written to the end of the file named, whatever the build prints", async () => {
		const folder = await mkdtemp(join(tmpdir(), "scope-"));
		try {
			const to = join(folder, "output");
			const run = spawnSync(
				process.execPath,
				["scripts/scope.ts", "--to", to],
				{
					cwd: join(import.meta.dirname, ".."),
					input: "README.md\nsite/data.json\n",
					encoding: "utf8",
				},
			);

			expect(run.status, run.stderr).toBe(0);
			expect(await readFile(to, "utf8")).toBe("card=true\n");
			expect(run.stdout).not.toContain("card=");
		} finally {
			await rm(folder, { recursive: true, force: true });
		}
	}, 60_000);
});
