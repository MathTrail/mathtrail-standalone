import { defineConfig } from "vite";

// The site's stylesheet, built beside the pages the site's own script draws.
// The pages are plain HTML and load what they need from the site's own origin,
// so this is an ordinary build rather than one page with everything inside it.
export default defineConfig({
	root: import.meta.dirname,
	publicDir: false,
	// The preview serves a built page from its directory, and answers 404 for
	// an address the site does not have, as the host the site is published on
	// does. Unlike the host, it does not lead an address written without its
	// trailing slash to the one with it: /en/privacy answers 404 here.
	appType: "mpa",
	// The directory the site is built into is named by whoever builds or
	// serves it, so that it is named in one place.
	build: {
		// Nothing is written into a page or a stylesheet as data: a check of the
		// site counts such an address as a load from somewhere else.
		assetsInlineLimit: 0,
		rolldownOptions: {
			// Stylesheets alone: Vite leaves out the empty script such an entry
			// would otherwise bring. The design's tokens are not imported into
			// them — the site ships them as a file of their own, the same file
			// the widget and the sign-in pages read. The cards' styles are the
			// very ones a chat draws a card with, for the pages that draw one.
			input: {
				style: "src/site/style.css",
				card: "src/design/mathtrail.css",
			},
			output: {
				// The stylesheet's address was published with the first site and
				// stays what it was; the font files it names keep the names their
				// package gives them.
				assetFileNames: "assets/[name][extname]",
			},
		},
	},
});
