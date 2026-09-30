import { defineConfig } from "vitest/config";

export default defineConfig({
	test: {
		environment: "happy-dom",
		include: ["src/**/*.test.{ts,tsx}", "scripts/**/*.test.ts", "*.test.ts"],
		coverage: {
			provider: "v8",
			// The code, not the configurations that the build test and every run
			// of the tests load. The entry point only finds the page's root
			// element and starts the widget, and start is what the tests cover;
			// what lives under testing/ serves the tests and is not the widget;
			// the preview is a page for looking at the widget, which ships
			// nowhere and is judged by eye; and the measure of the layout is a
			// check of its own, run in real browsers against the preview.
			include: ["src/**/*.{ts,tsx}", "scripts/**/*.ts"],
			exclude: [
				"**/*.test.{ts,tsx}",
				"**/*.d.ts",
				"src/**/testing/**",
				"src/preview/**",
				"src/widget/main.ts",
				"scripts/layout.ts",
			],
			reportsDirectory: "coverage",
			// Paths are written from the repository's root, where the readers of
			// the report look for the files.
			reporter: [["lcovonly", { projectRoot: ".." }]],
		},
	},
});
