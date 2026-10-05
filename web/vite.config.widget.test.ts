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

let out: string | undefined;
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

// A build that could not even make its directory has nothing to take away,
// and its own failure is the one to tell.
afterAll(async () => {
	if (out !== undefined) {
		await rm(out, { recursive: true, force: true });
	}
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

// The card draws the logo once for each theme, and the theme's tokens show
// one: a page that lost the rule binding a drawing to its token would show
// both, side by side.
test("the widget page shows one drawing of the logo in each theme", () => {
	expect(page).toMatch(/\.mt-mark-light\s*\{\s*display:\s*var\(--mark-light\)/);
	expect(page).toMatch(/\.mt-mark-dark\s*\{\s*display:\s*var\(--mark-dark\)/);
});

// solversIn are the names of the solvers a part of the site's data names, at
// any depth: the solvers that prove its examples' answers.
function solversIn(part: unknown): string[] {
	if (Array.isArray(part)) {
		return part.flatMap(solversIn);
	}
	if (typeof part !== "object" || part === null) {
		return [];
	}
	return Object.entries(part).flatMap(([key, value]) =>
		key === "solver" && typeof value === "string" ? [value] : solversIn(value),
	);
}

// The card offers the topics in the groups of the site's page of topics, read
// from the file the site is built from. The rest of that file is the site's —
// its examples, the solvers that prove them, its pages' cards — and a page a
// host loads for every task has no use for it.
test("the widget page carries the site's groups of topics and nothing else of the site's data", async () => {
	const data: unknown = JSON.parse(
		await readFile(
			join(import.meta.dirname, "..", "site", "data.json"),
			"utf8",
		),
	);
	// A solver named by one word — a cross, a ferry — is a word a page may hold
	// for any reason; one of several words is the example's alone.
	const solvers = solversIn(data).filter((solver) => solver.includes("-"));

	expect(solvers.length).toBeGreaterThan(10);
	for (const solver of solvers) {
		expect(page, solver).not.toContain(solver);
	}
});

// The icons the card was first drawn with came from another app's design, and
// the card draws its own. Every part of their paths, from one move to the
// next, is named here by its SHA-256, so that none of them is kept in the
// repository, not even as a test's words, and an old part is found wherever it
// is drawn: alone, inside a longer path, or put together with others.
const borrowed = new Map([
	[
		"8d363870a61d2b688676f03646bed7a4a2157db375ed2dc97dd1f1393d47f5f8",
		"chevron-right",
	],
	[
		"a77c8a5184899b72e2a50aa706824dba812e5ab48844e40878a84de845a8449a",
		"chevron-left",
	],
	["a03f5ab156762e7447dbec86b5e516649c5c435382f3018cd334fe53ca024db7", "check"],
	["9517c51141a3598e2ab41b6398aa8074466cf2b638050d138b1902fd0d59ab35", "cross"],
	["cf085fb7d5cf054f8e0cecb70f1df319c1711a73f676f045a326fdfe225ff322", "cross"],
	["5c4e3e41d8f78077e4a1d78a7921dc9a5bd15ce05b49cc4b2c12a7ee6fa5cdfc", "hint"],
	["6aac2e08b22158c9b4328b4609053682a40be644f0923f4adcd9693d6d6df8b6", "hint"],
	["b5be2f030588ff17d00bdc93886e6e3fee746d4bceaa8f317e9c690529437960", "hint"],
	["81e7b4d3336a0e1ee91c9b5062757c93207bf5b5578c98a32fb15cce917a0cea", "trap"],
	["a754ad7fc37cef7fc9c8a690e0d646ce480d28940e789bbf553fed22ab0ff68f", "trap"],
	["292881d2bb9c531b4bc47f83cfc53abae4913ca2e6d7ac1599036d571616acdd", "trap"],
	[
		"09131f29dfca5206a2517bf655bb4ab08b672255cf5a782750ac27ccadc06c6b",
		"the tick of a verdict and of a step done",
	],
	[
		"1214a651c16133caba02417326b146d7beb4c6e98557223477272a0041fdb833",
		"the cross of a wrong verdict",
	],
	[
		"617d1ea1c20f6505f0a65a4410800995361394319cd5b1b8cd620f0571001ba4",
		"the cross of a wrong verdict",
	],
	[
		"5b3a860e3b1261ed44a2db8e974880d0f3d9f16e0f03786e75a9084def496884",
		"the spinner's arc",
	],
	[
		"8092a7d0da337e475c4a0ee42802af970fe007a8c43656bbc3a10c94750e1694",
		"the card's mark",
	],
	[
		"8e610311a4122bfd986ea5a2863fd44d638e10fd6090dad35f7f014d6f0f1499",
		"the avatar",
	],
]);

test("the widget page draws none of the icons the first design took from another app", () => {
	const drawn = new Set(partsIn(page));
	// The search finds what the card draws, so that finding none of the old
	// parts means something: every part of the tick, a path of its own, and of
	// a wrong verdict, a path the card puts together from a ring and a cross.
	for (const name of ["check", "verdict-wrong"] as const) {
		expect(
			partsOf(pathOf(renderToString(h(Icon, { name })))).filter(
				(part) => !drawn.has(part),
			),
		).toEqual([]);
	}
	expect(
		[...drawn]
			.map((part) => [borrowed.get(sha256(part)), part])
			.filter(([name]) => name !== undefined)
			.map(([name, part]) => `${name}: ${part}`),
	).toEqual([]);
});

// partsIn is every part of a path the page could draw: each run of path
// commands and numbers that starts with a move, quoted or written into a
// template between the parts it is put together from, cut before each move.
function partsIn(text: string): string[] {
	return [...text.matchAll(/[Mm][\d\s.,+\-MmLlHhVvCcSsQqTtAaZz]*/g)]
		.flatMap(([run]) => partsOf(run))
		.filter((part) => /\d/.test(part));
}

// partsOf is a path cut before each of its moves, each part without the
// spaces or commas that part it from the next.
function partsOf(d: string): string[] {
	return d.split(/(?=[Mm])/).map((part) => part.replace(/[\s,]+$/, ""));
}

// pathOf is the path a drawing's markup draws.
function pathOf(markup: string): string {
	return markup.match(/\bd="([^"]+)"/)?.[1] ?? "";
}

function sha256(text: string): string {
	return createHash("sha256").update(text).digest("hex");
}
