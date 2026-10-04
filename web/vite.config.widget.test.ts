// @vitest-environment node
import { createHash } from "node:crypto";
import { mkdtemp, readdir, readFile, rm } from "node:fs/promises";
import { tmpdir } from "node:os";
import { join } from "node:path";
import { h } from "preact";
import { renderToString } from "preact-render-to-string";
import { build } from "vite";
import { afterAll, beforeAll, expect, test } from "vitest";
import { Icon } from "./src/design/icons.tsx";
import { pseudoLocale } from "./src/i18n/pseudo.ts";
import config from "./vite.config.widget.ts";

let out: string;
let files: string[];
let page: string;

// The widget is built once, and each test reads the page the build left.
beforeAll(async () => {
	out = await mkdtemp(join(tmpdir(), "widget-"));
	await build({
		...config,
		configFile: false,
		logLevel: "silent",
		build: { ...config.build, outDir: out, emptyOutDir: true },
	});
	files = await readdir(out, { recursive: true });
	page = await readFile(join(out, "widget.html"), "utf8");
}, 60_000);

afterAll(async () => {
	await rm(out, { recursive: true, force: true });
});

// The host runs the widget in a sandbox that lets it load nothing, so a build
// that left a script or a stylesheet beside the page would ship a card that
// never draws.
test("the widget builds into one page that loads nothing from outside it", () => {
	expect(files).toEqual(["widget.html"]);
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
});

// The icons the card was first drawn with came from another app's design, and
// the card draws its own. They are named here by the SHA-256 of their paths,
// so that none of them is kept in the repository, not even as a test's words.
const borrowed = new Map([
	["8d363870a61d2b688676f03646bed7a4a2157db375ed2dc97dd1f1393d47f5f8", "chevron-right"],
	["a77c8a5184899b72e2a50aa706824dba812e5ab48844e40878a84de845a8449a", "chevron-left"],
	["a03f5ab156762e7447dbec86b5e516649c5c435382f3018cd334fe53ca024db7", "check"],
	["3fe400c78783024a37f5398e3353db4140d24f271a74a645d18041b7cff08f5a", "cross"],
	["010fc08297ce23e950174653d97eb7fe80963b35c6f6738a93879e6cb8258e39", "hint"],
	["41518db561fcc3c6d7f6c33754b1250c8bac205ec24fa041bd2a0b3c2124459f", "trap"],
	["09131f29dfca5206a2517bf655bb4ab08b672255cf5a782750ac27ccadc06c6b", "the tick of a verdict and of a step done"],
	["70e2d4b6b10859ff74f12e925a0203f0fb89755ac20e8f4f89bace77dfe3dd55", "the cross of a wrong verdict"],
	["5b3a860e3b1261ed44a2db8e974880d0f3d9f16e0f03786e75a9084def496884", "the spinner's arc"],
	["8092a7d0da337e475c4a0ee42802af970fe007a8c43656bbc3a10c94750e1694", "the card's mark"],
	["8e610311a4122bfd986ea5a2863fd44d638e10fd6090dad35f7f014d6f0f1499", "the avatar"],
]);

test("the widget page draws none of the icons the first design took from another app", () => {
	const drawn = pathsIn(page);
	// The search finds what the card draws: the tick it draws now is among
	// the paths it finds, so finding none of the old ones means something.
	expect(drawn).toContain(pathOf(renderToString(h(Icon, { name: "check" }))));
	expect(
		drawn
			.filter((d) => borrowed.has(sha256(d)))
			.map((d) => `${borrowed.get(sha256(d))}: ${d}`),
	).toEqual([]);
});

// pathsIn is every string of page that reads as the path of a drawing: a
// quoted run of path commands and numbers, starting with a move.
function pathsIn(text: string): string[] {
	return [
		...text.matchAll(/(["'`])([Mm][\d\s.,+\-MmLlHhVvCcSsQqTtAaZz]*)\1/g),
	]
		.map((match) => match[2] ?? "")
		.filter((d) => /\d/.test(d));
}

// pathOf is the path a drawing's markup draws.
function pathOf(markup: string): string {
	return markup.match(/\bd="([^"]+)"/)?.[1] ?? "";
}

function sha256(text: string): string {
	return createHash("sha256").update(text).digest("hex");
}
