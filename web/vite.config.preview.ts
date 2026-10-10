import { resolve } from "node:path";
import { defineConfig } from "vite";

// The preview is a page that plays a chat host for the widget: it is served
// while the widget is worked on, and built, with the widget's page, for the
// browsers that measure and photograph the cards, which then load the widget
// as a few files rather than as its hundred modules. It frames the widget's
// own page, which imports the design tokens from beside the page the build
// writes, outside this folder; the server hands out that one file and nothing
// else from there.
export default defineConfig({
	root: import.meta.dirname,
	publicDir: false,
	server: {
		// The address the editor forwards a port of the container from: a
		// server left to name localhost itself listens on IPv6's alone, which a
		// forwarded port does not reach.
		host: "127.0.0.1",
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
