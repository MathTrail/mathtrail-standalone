import { defineConfig } from "vite";
import { viteSingleFile } from "vite-plugin-singlefile";

// The widget is one HTML page with its script and styles inside it. A chat host
// runs it in a sandbox that lets it load nothing from anywhere, so whatever the
// page needs has to arrive in the page itself.
export default defineConfig({
	root: import.meta.dirname,
	publicDir: false,
	plugins: [viteSingleFile()],
	build: {
		// The Go package that embeds the page. Its own files live beside the
		// page, so nothing there is emptied before a build.
		outDir: "../internal/widget",
		emptyOutDir: false,
		target: "es2022",
		// Everything is in the one page, so there is nothing to preload and no
		// use for the code that would polyfill preloading.
		modulePreload: { polyfill: false },
		rolldownOptions: {
			input: "widget.html",
		},
	},
});
