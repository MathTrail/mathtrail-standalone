import { resolve } from "node:path";
import { defineConfig } from "vite";

// The preview is a page that plays a chat host for the widget: it is served
// while the widget is worked on and never built. It frames the widget's own
// page, which imports the design tokens from beside the page the build writes,
// outside this folder; the server hands out that one file and nothing else
// from there.
export default defineConfig({
	root: import.meta.dirname,
	publicDir: false,
	server: {
		port: 5173,
		strictPort: true,
		fs: {
			allow: [
				import.meta.dirname,
				resolve(import.meta.dirname, "../internal/widget/tokens.css"),
			],
		},
	},
});
